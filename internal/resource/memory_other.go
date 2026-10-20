//go:build !linux

package resource

import (
	"context"
	"fmt"
	"runtime"
)

type hostSampler struct{}

func NewHostSampler() Sampler {
	return hostSampler{}
}

func (hostSampler) Snapshot(ctx context.Context) (Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return Snapshot{}, err
	}
	return Snapshot{}, fmt.Errorf("host RAM sampling is not implemented on %s yet", runtime.GOOS)
}
