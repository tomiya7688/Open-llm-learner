//go:build linux

package resource

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type hostSampler struct{}

func NewHostSampler() Sampler {
	return hostSampler{}
}

func (hostSampler) Snapshot(ctx context.Context) (Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return Snapshot{}, err
	}
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return Snapshot{}, fmt.Errorf("read /proc/meminfo: %w", err)
	}
	values := map[string]uint64{}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSuffix(fields[0], ":")
		if key != "MemTotal" && key != "MemAvailable" && key != "MemFree" && key != "Buffers" && key != "Cached" {
			continue
		}
		kib, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return Snapshot{}, fmt.Errorf("parse /proc/meminfo %s: %w", key, err)
		}
		values[key] = kib * 1024
	}
	total := values["MemTotal"]
	if total == 0 {
		return Snapshot{}, errors.New("/proc/meminfo did not contain MemTotal")
	}
	available := values["MemAvailable"]
	if available == 0 {
		available = values["MemFree"] + values["Buffers"] + values["Cached"]
	}
	if available > total {
		available = total
	}
	return Snapshot{RAM:DeviceMemory{TotalBytes:total, AvailableBytes:available}}, nil
}
