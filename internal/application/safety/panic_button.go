package safety

import (
	"context"
	"errors"
	"fmt"

	"github.com/lekhanhr/chaos-dock/internal/domain/fault"
)

type ContainerRestarter interface {
	Restart(ctx context.Context, containerID string) error
}

type PanicButton struct {
	Injector  fault.FaultInjector
	Restarter ContainerRestarter
}

// Trigger attempts best-effort rollback of network faults and restarts targets.
func (p *PanicButton) Trigger(ctx context.Context, containerIDs []string) error {
	var errs []error

	for _, id := range containerIDs {
		if p.Injector != nil {
			if err := p.Injector.RevertNetworkLatency(ctx, id); err != nil {
				errs = append(errs, fmt.Errorf("revert latency on %s: %w", id, err))
			}
		}

		if p.Restarter != nil {
			if err := p.Restarter.Restart(ctx, id); err != nil {
				errs = append(errs, fmt.Errorf("restart %s: %w", id, err))
			}
		}
	}

	return errors.Join(errs...)
}

