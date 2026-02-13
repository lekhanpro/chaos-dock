package safety

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/lekhanpro/chaos-dock/internal/domain/fault"
)

type ContainerRestarter interface {
	Restart(ctx context.Context, containerID string) error
}

type PanicButton struct {
	Injector  fault.FaultInjector
	Restarter ContainerRestarter
	Registry  *TargetRegistry
}

func (p *PanicButton) TriggerAll(ctx context.Context) error {
	return p.Trigger(ctx, nil)
}

// Trigger attempts best-effort rollback of network faults and restarts targets.
func (p *PanicButton) Trigger(ctx context.Context, containerIDs []string) error {
	containerIDs = normalizeTargets(containerIDs)
	if len(containerIDs) == 0 && p.Registry != nil {
		containerIDs = p.Registry.Snapshot()
	}

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

	if p.Registry != nil {
		p.Registry.Reset()
	}

	return errors.Join(errs...)
}

func normalizeTargets(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))

	for _, raw := range in {
		id := strings.TrimSpace(raw)
		if id == "" {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}

	return out
}
