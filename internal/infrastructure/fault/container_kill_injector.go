package fault

import (
	"context"
	"fmt"
	"strings"

	domainfault "github.com/lekhanpro/chaos-dock/internal/domain/fault"
)

const defaultKillSignal = "SIGKILL"

var supportedSignals = map[string]struct{}{
	"SIGKILL": {},
	"SIGTERM": {},
	"SIGINT":  {},
	"SIGQUIT": {},
	"SIGHUP":  {},
	"SIGUSR1": {},
	"SIGUSR2": {},
}

type killExecutor interface {
	Kill(ctx context.Context, containerID string, signal string) error
}

type ContainerKillInjector struct {
	executor killExecutor
}

func NewContainerKillInjector(executor killExecutor) *ContainerKillInjector {
	return &ContainerKillInjector{executor: executor}
}

func (c *ContainerKillInjector) KillContainer(ctx context.Context, containerID string, signal string) error {
	containerID = strings.TrimSpace(containerID)
	if containerID == "" {
		return domainfault.ErrInvalidContainerID
	}
	if c.executor == nil {
		return fmt.Errorf("kill executor is required")
	}

	normalizedSignal, err := normalizeSignal(signal)
	if err != nil {
		return err
	}

	if err := c.executor.Kill(ctx, containerID, normalizedSignal); err != nil {
		return fmt.Errorf("%w: %v", domainfault.ErrContainerKillFailed, err)
	}

	return nil
}

func normalizeSignal(raw string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(raw))
	if normalized == "" {
		normalized = defaultKillSignal
	}
	if !strings.HasPrefix(normalized, "SIG") {
		normalized = "SIG" + normalized
	}

	if _, ok := supportedSignals[normalized]; !ok {
		return "", fmt.Errorf("%w: %q", domainfault.ErrInvalidKillSignal, raw)
	}

	return normalized, nil
}
