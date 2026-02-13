package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/lekhanhr/chaos-dock/internal/domain/fault"
)

type Runner struct {
	Injector fault.FaultInjector
}

func (r *Runner) ApplyNetworkLatency(ctx context.Context, containerID string, delay time.Duration) error {
	if r.Injector == nil {
		return fmt.Errorf("fault injector is not configured")
	}

	return r.Injector.InjectNetworkLatency(ctx, containerID, delay)
}

