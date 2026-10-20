package resource

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type fakeSampler struct {
	mu       sync.Mutex
	snapshot Snapshot
	err      error
}

func (f *fakeSampler) Snapshot(context.Context) (Snapshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.snapshot, f.err
}

func TestGovernorPreservesAutoHeadroom(t *testing.T) {
	sampler := &fakeSampler{snapshot:Snapshot{RAM:DeviceMemory{TotalBytes:1000, AvailableBytes:800}}}
	g, err := NewGovernor(GovernorConfig{Policy:Policy{}, Sampler:sampler})
	if err != nil {
		t.Fatal(err)
	}
	first, err := g.TryAdmit(context.Background(), Request{ID:"a", Estimate:Estimate{RAMBytes:600, Heavy:true}})
	if err != nil {
		t.Fatal(err)
	}
	defer first.Release()

	_, err = g.TryAdmit(context.Background(), Request{ID:"b", Estimate:Estimate{RAMBytes:150}})
	if !errors.Is(err, ErrWouldBlock) {
		t.Fatalf("expected headroom admission block, got %v", err)
	}
	status := g.Status()
	if status.ReservedRAMBytes != 600 || status.ActiveTasks != 1 {
		t.Fatalf("unexpected governor status: %+v", status)
	}
}

func TestGovernorRejectsRequestLargerThanConfiguredBudget(t *testing.T) {
	sampler := &fakeSampler{snapshot:Snapshot{RAM:DeviceMemory{TotalBytes:1000, AvailableBytes:1000}}}
	g, err := NewGovernor(GovernorConfig{Policy:Policy{}, Sampler:sampler})
	if err != nil {
		t.Fatal(err)
	}
	_, err = g.TryAdmit(context.Background(), Request{Estimate:Estimate{RAMBytes:901}})
	if !errors.Is(err, ErrImpossible) {
		t.Fatalf("expected impossible request, got %v", err)
	}
}

func TestGovernorHeavyConcurrency(t *testing.T) {
	policy := Policy{HeavyTaskConcurrency:1}
	sampler := &fakeSampler{snapshot:Snapshot{RAM:DeviceMemory{TotalBytes:10_000, AvailableBytes:10_000}}}
	g, err := NewGovernor(GovernorConfig{Policy:policy, Sampler:sampler})
	if err != nil {
		t.Fatal(err)
	}
	lease, err := g.TryAdmit(context.Background(), Request{ID:"a", Estimate:Estimate{RAMBytes:100, Heavy:true}})
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Release()
	if _, err := g.TryAdmit(context.Background(), Request{ID:"b", Estimate:Estimate{RAMBytes:100, Heavy:true}}); !errors.Is(err, ErrWouldBlock) {
		t.Fatalf("expected heavy concurrency block, got %v", err)
	}
}

func TestGovernorWaitsUntilLeaseReleased(t *testing.T) {
	sampler := &fakeSampler{snapshot:Snapshot{RAM:DeviceMemory{TotalBytes:1000, AvailableBytes:800}}}
	g, err := NewGovernor(GovernorConfig{Policy:Policy{}, Sampler:sampler, PollInterval:time.Second})
	if err != nil {
		t.Fatal(err)
	}
	first, err := g.TryAdmit(context.Background(), Request{ID:"first", Estimate:Estimate{RAMBytes:700}})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	result := make(chan error, 1)
	go func() {
		second, err := g.Admit(ctx, Request{ID:"second", Estimate:Estimate{RAMBytes:100}})
		if second != nil {
			second.Release()
		}
		result <- err
	}()

	select {
	case err := <-result:
		t.Fatalf("second admission returned before release: %v", err)
	case <-time.After(30*time.Millisecond):
	}
	first.Release()
	select {
	case err := <-result:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("second admission did not wake after release")
	}
}

func TestGovernorRejectsUnknownVRAMByDefault(t *testing.T) {
	sampler := &fakeSampler{snapshot:Snapshot{RAM:DeviceMemory{TotalBytes:1000, AvailableBytes:1000}}}
	g, err := NewGovernor(GovernorConfig{Policy:Policy{}, Sampler:sampler})
	if err != nil {
		t.Fatal(err)
	}
	_, err = g.TryAdmit(context.Background(), Request{Estimate:Estimate{VRAMBytes:1}})
	if !errors.Is(err, ErrResourceUnknown) {
		t.Fatalf("expected unknown VRAM rejection, got %v", err)
	}
}

func TestGovernorHonorsVRAMHeadroom(t *testing.T) {
	sampler := &fakeSampler{snapshot:Snapshot{
		RAM:DeviceMemory{TotalBytes:10_000, AvailableBytes:10_000},
		VRAM:&DeviceMemory{TotalBytes:1000, AvailableBytes:800},
	}}
	g, err := NewGovernor(GovernorConfig{Policy:Policy{}, Sampler:sampler})
	if err != nil {
		t.Fatal(err)
	}
	lease, err := g.TryAdmit(context.Background(), Request{Estimate:Estimate{VRAMBytes:700}})
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Release()
	if _, err := g.TryAdmit(context.Background(), Request{Estimate:Estimate{VRAMBytes:1}}); !errors.Is(err, ErrWouldBlock) {
		t.Fatalf("expected VRAM reservation block, got %v", err)
	}
}
