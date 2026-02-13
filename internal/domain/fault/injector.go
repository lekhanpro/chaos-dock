package fault

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidContainerID          = errors.New("container id is required")
	ErrInvalidLatencyDuration      = errors.New("latency duration must be greater than zero")
	ErrInvalidKillSignal           = errors.New("invalid kill signal")
	ErrNamespaceToolMissing        = errors.New("namespace tooling is unavailable on host")
	ErrNetworkNamespaceUnavailable = errors.New("container network namespace is unavailable")
	ErrIPRoute2Missing             = errors.New("iproute2/tc is not available in target namespace")
	ErrInsufficientPrivileges      = errors.New("insufficient privileges to alter qdisc")
	ErrCommandTimeout              = errors.New("fault injection command timed out")
	ErrTCCommandFailed             = errors.New("tc command execution failed")
	ErrContainerKillFailed         = errors.New("container kill failed")
	ErrUnsupportedPlatform         = errors.New("this injector supports linux hosts only")
)

// FaultInjector defines fault operations used by the application layer.
type FaultInjector interface {
	InjectNetworkLatency(ctx context.Context, containerID string, delay time.Duration) error
	RevertNetworkLatency(ctx context.Context, containerID string) error
}

// ContainerKiller defines a fault that terminates container processes.
type ContainerKiller interface {
	KillContainer(ctx context.Context, containerID string, signal string) error
}
