package resource

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrWouldBlock      = errors.New("resource admission would block")
	ErrImpossible      = errors.New("resource request cannot fit configured budget")
	ErrResourceUnknown = errors.New("required resource capacity is unavailable")
)

type DeviceMemory struct {
	TotalBytes     uint64
	AvailableBytes uint64
}

type Snapshot struct {
	RAM  DeviceMemory
	VRAM *DeviceMemory
}

type Sampler interface {
	Snapshot(context.Context) (Snapshot, error)
}

type Estimate struct {
	RAMBytes       uint64
	VRAMBytes      uint64
	DiskCacheBytes uint64
	Heavy          bool
}

type Request struct {
	ID       string
	Estimate Estimate
	Priority int
}

type GovernorConfig struct {
	Policy           Policy
	Sampler          Sampler
	PollInterval     time.Duration
	AllowUnknownVRAM bool
}

type Governor struct {
	policy           Policy
	sampler          Sampler
	pollInterval     time.Duration
	allowUnknownVRAM bool

	mu      sync.Mutex
	active  map[string]Estimate
	notify  chan struct{}
	counter atomic.Uint64
}

type Lease struct {
	governor *Governor
	id       string
	once     sync.Once
}

type Status struct {
	ActiveTasks       int
	ActiveHeavyTasks  int
	ReservedRAMBytes  uint64
	ReservedVRAMBytes uint64
	ReservedDiskBytes uint64
}

type AdmissionError struct {
	Kind      error
	Resource  string
	Requested uint64
	Available uint64
	Budget    uint64
	Message   string
}

func (e *AdmissionError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("%s: requested=%d available=%d budget=%d: %v", e.Resource, e.Requested, e.Available, e.Budget, e.Kind)
}

func (e *AdmissionError) Unwrap() error { return e.Kind }

func NewGovernor(cfg GovernorConfig) (*Governor, error) {
	if cfg.Sampler == nil {
		return nil, errors.New("resource sampler is required")
	}
	cfg.Policy.Normalize()
	if err := cfg.Policy.Validate(); err != nil {
		return nil, err
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 500 * time.Millisecond
	}
	return &Governor{
		policy: cfg.Policy,
		sampler: cfg.Sampler,
		pollInterval: cfg.PollInterval,
		allowUnknownVRAM: cfg.AllowUnknownVRAM,
		active: map[string]Estimate{},
		notify: make(chan struct{}),
	}, nil
}

func (g *Governor) TryAdmit(ctx context.Context, request Request) (*Lease, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	snapshot, err := g.sampler.Snapshot(ctx)
	if err != nil {
		return nil, fmt.Errorf("sample host resources: %w", err)
	}
	if snapshot.RAM.TotalBytes == 0 {
		return nil, &AdmissionError{Kind:ErrResourceUnknown, Resource:"ram", Requested:request.Estimate.RAMBytes, Message:"RAM total capacity is unavailable"}
	}
	ramBudget, err := ComputeBudget(snapshot.RAM.TotalBytes, g.policy.RAM)
	if err != nil {
		return nil, err
	}
	if request.Estimate.RAMBytes > ramBudget.UsableBytes {
		return nil, &AdmissionError{
			Kind:ErrImpossible, Resource:"ram", Requested:request.Estimate.RAMBytes,
			Available:availableAfterReserve(snapshot.RAM.AvailableBytes, ramBudget.ReservedBytes),
			Budget:ramBudget.UsableBytes,
		}
	}

	var vramBudget Budget
	if request.Estimate.VRAMBytes > 0 {
		if snapshot.VRAM == nil || snapshot.VRAM.TotalBytes == 0 {
			if !g.allowUnknownVRAM {
				return nil, &AdmissionError{Kind:ErrResourceUnknown, Resource:"vram", Requested:request.Estimate.VRAMBytes, Message:"VRAM capacity is unavailable; refusing a VRAM reservation without explicit override"}
			}
		} else {
			vramBudget, err = ComputeBudget(snapshot.VRAM.TotalBytes, g.policy.VRAM)
			if err != nil {
				return nil, err
			}
			if request.Estimate.VRAMBytes > vramBudget.UsableBytes {
				return nil, &AdmissionError{
					Kind:ErrImpossible, Resource:"vram", Requested:request.Estimate.VRAMBytes,
					Available:availableAfterReserve(snapshot.VRAM.AvailableBytes, vramBudget.ReservedBytes),
					Budget:vramBudget.UsableBytes,
				}
			}
		}
	}
	if g.policy.DiskCache.MaxBytes != nil && request.Estimate.DiskCacheBytes > *g.policy.DiskCache.MaxBytes {
		return nil, &AdmissionError{Kind:ErrImpossible, Resource:"disk_cache", Requested:request.Estimate.DiskCacheBytes, Budget:*g.policy.DiskCache.MaxBytes}
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	status := g.statusLocked()
	id := request.ID
	if id == "" {
		id = fmt.Sprintf("resource-task-%d", g.counter.Add(1))
	}
	if _, exists := g.active[id]; exists {
		return nil, fmt.Errorf("resource task id %q is already active", id)
	}

	if request.Estimate.Heavy && status.ActiveHeavyTasks >= g.policy.HeavyTaskConcurrency {
		return nil, &AdmissionError{Kind:ErrWouldBlock, Resource:"heavy_task_concurrency", Requested:1, Available:0, Budget:uint64(g.policy.HeavyTaskConcurrency)}
	}
	ramAvailable := availableAfterReserve(snapshot.RAM.AvailableBytes, ramBudget.ReservedBytes)
	ramAvailable = subtractFloor(ramAvailable, status.ReservedRAMBytes)
	if request.Estimate.RAMBytes > ramAvailable {
		return nil, &AdmissionError{Kind:ErrWouldBlock, Resource:"ram", Requested:request.Estimate.RAMBytes, Available:ramAvailable, Budget:ramBudget.UsableBytes}
	}
	if request.Estimate.VRAMBytes > 0 && snapshot.VRAM != nil && snapshot.VRAM.TotalBytes > 0 {
		vramAvailable := availableAfterReserve(snapshot.VRAM.AvailableBytes, vramBudget.ReservedBytes)
		vramAvailable = subtractFloor(vramAvailable, status.ReservedVRAMBytes)
		if request.Estimate.VRAMBytes > vramAvailable {
			return nil, &AdmissionError{Kind:ErrWouldBlock, Resource:"vram", Requested:request.Estimate.VRAMBytes, Available:vramAvailable, Budget:vramBudget.UsableBytes}
		}
	}
	if g.policy.DiskCache.MaxBytes != nil {
		diskAvailable := subtractFloor(*g.policy.DiskCache.MaxBytes, status.ReservedDiskBytes)
		if request.Estimate.DiskCacheBytes > diskAvailable {
			return nil, &AdmissionError{Kind:ErrWouldBlock, Resource:"disk_cache", Requested:request.Estimate.DiskCacheBytes, Available:diskAvailable, Budget:*g.policy.DiskCache.MaxBytes}
		}
	}

	g.active[id] = request.Estimate
	return &Lease{governor:g, id:id}, nil
}

func (g *Governor) Admit(ctx context.Context, request Request) (*Lease, error) {
	for {
		lease, err := g.TryAdmit(ctx, request)
		if err == nil {
			return lease, nil
		}
		if !errors.Is(err, ErrWouldBlock) {
			return nil, err
		}

		g.mu.Lock()
		notify := g.notify
		g.mu.Unlock()
		timer := time.NewTimer(g.pollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-notify:
			timer.Stop()
		case <-timer.C:
		}
	}
}

func (l *Lease) ID() string {
	if l == nil { return "" }
	return l.id
}

func (l *Lease) Release() {
	if l == nil || l.governor == nil {
		return
	}
	l.once.Do(func() {
		g := l.governor
		g.mu.Lock()
		delete(g.active, l.id)
		close(g.notify)
		g.notify = make(chan struct{})
		g.mu.Unlock()
	})
}

func (g *Governor) Status() Status {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.statusLocked()
}

func (g *Governor) statusLocked() Status {
	var status Status
	status.ActiveTasks = len(g.active)
	for _, estimate := range g.active {
		status.ReservedRAMBytes += estimate.RAMBytes
		status.ReservedVRAMBytes += estimate.VRAMBytes
		status.ReservedDiskBytes += estimate.DiskCacheBytes
		if estimate.Heavy {
			status.ActiveHeavyTasks++
		}
	}
	return status
}

func availableAfterReserve(available, reserve uint64) uint64 {
	return subtractFloor(available, reserve)
}

func subtractFloor(value, amount uint64) uint64 {
	if amount >= value {
		return 0
	}
	return value - amount
}
