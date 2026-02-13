//go:build !linux

package fault

import (
	"context"
	"fmt"
	"time"

	domainfault "github.com/lekhanpro/chaos-dock/internal/domain/fault"
)

type PIDResolver interface {
	ContainerPID(ctx context.Context, containerID string) (int, error)
}

type NetworkLatencyInjector struct{}

func NewNetworkLatencyInjector(_ PIDResolver) *NetworkLatencyInjector {
	return &NetworkLatencyInjector{}
}

func (n *NetworkLatencyInjector) InjectNetworkLatency(ctx context.Context, containerID string, delay time.Duration) error {
	_ = ctx
	_ = containerID
	_ = delay
	return fmt.Errorf("%w", domainfault.ErrUnsupportedPlatform)
}

func (n *NetworkLatencyInjector) RevertNetworkLatency(ctx context.Context, containerID string) error {
	_ = ctx
	_ = containerID
	return fmt.Errorf("%w", domainfault.ErrUnsupportedPlatform)
}

