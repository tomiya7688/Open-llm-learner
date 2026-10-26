package resource

import (
	"errors"
	"fmt"
	"math"
)

const DefaultAutoMaxFraction = 0.90

type MemoryLimit struct {
	Mode         string   `json:"mode"`
	MaxFraction  *float64 `json:"max_fraction,omitempty"`
	MinFreeBytes *uint64  `json:"min_free_bytes,omitempty"`
}

type DiskCacheLimit struct {
	MaxBytes *uint64 `json:"max_bytes,omitempty"`
}

type Policy struct {
	RAM                  MemoryLimit    `json:"ram"`
	VRAM                 MemoryLimit    `json:"vram"`
	DiskCache            DiskCacheLimit `json:"disk_cache"`
	HeavyTaskConcurrency int            `json:"heavy_task_concurrency"`
}

type Budget struct {
	TotalBytes      uint64
	ReservedBytes   uint64
	UsableBytes     uint64
	MaxFraction     float64
	ExplicitReserve uint64
}

func (p *Policy) Normalize() {
	normalizeLimit(&p.RAM)
	normalizeLimit(&p.VRAM)
	if p.HeavyTaskConcurrency <= 0 {
		p.HeavyTaskConcurrency = 1
	}
}

func normalizeLimit(l *MemoryLimit) {
	if l.Mode == "" {
		l.Mode = "auto"
	}
	if l.Mode == "auto" && l.MaxFraction == nil {
		v := DefaultAutoMaxFraction
		l.MaxFraction = &v
	}
}

func (p Policy) Validate() error {
	if err := validateLimit("ram", p.RAM); err != nil {
		return err
	}
	if err := validateLimit("vram", p.VRAM); err != nil {
		return err
	}
	if p.HeavyTaskConcurrency < 0 {
		return errors.New("heavy_task_concurrency cannot be negative")
	}
	return nil
}

func validateLimit(name string, l MemoryLimit) error {
	mode := l.Mode
	if mode == "" {
		mode = "auto"
	}
	if mode != "auto" && mode != "manual" && mode != "unlimited" {
		return fmt.Errorf("%s mode %q is invalid", name, mode)
	}
	if l.MaxFraction != nil && (*l.MaxFraction <= 0 || *l.MaxFraction > 1) {
		return fmt.Errorf("%s max_fraction must be > 0 and <= 1", name)
	}
	if mode == "auto" && l.MaxFraction != nil && *l.MaxFraction >= 1 {
		return fmt.Errorf("%s auto mode must preserve headroom; max_fraction must be < 1", name)
	}
	return nil
}

func ComputeBudget(totalBytes uint64, limit MemoryLimit) (Budget, error) {
	if totalBytes == 0 {
		return Budget{}, errors.New("totalBytes must be greater than zero")
	}
	if limit.Mode == "" {
		limit.Mode = "auto"
	}
	if limit.Mode == "auto" && limit.MaxFraction == nil {
		v := DefaultAutoMaxFraction
		limit.MaxFraction = &v
	}
	if err := validateLimit("memory", limit); err != nil {
		return Budget{}, err
	}
	if limit.Mode == "unlimited" {
		return Budget{TotalBytes: totalBytes, UsableBytes: totalBytes, MaxFraction: 1}, nil
	}
	fraction := 1.0
	if limit.MaxFraction != nil {
		fraction = *limit.MaxFraction
	}
	usableByFraction := uint64(math.Floor(float64(totalBytes) * fraction))
	derivedReserve := totalBytes - usableByFraction
	explicitReserve := uint64(0)
	if limit.MinFreeBytes != nil {
		explicitReserve = *limit.MinFreeBytes
	}
	reserve := derivedReserve
	if explicitReserve > reserve {
		reserve = explicitReserve
	}
	if reserve >= totalBytes {
		return Budget{}, errors.New("memory reserve leaves no usable capacity")
	}
	return Budget{
		TotalBytes:      totalBytes,
		ReservedBytes:   reserve,
		UsableBytes:     totalBytes - reserve,
		MaxFraction:     fraction,
		ExplicitReserve: explicitReserve,
	}, nil
}
