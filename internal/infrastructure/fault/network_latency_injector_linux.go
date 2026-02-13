//go:build linux

package fault

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	domainfault "github.com/lekhanpro/chaos-dock/internal/domain/fault"
)

const (
	defaultInterfaceName = "eth0"
	defaultNsenterBinary = "nsenter"
	defaultCommandTTL    = 10 * time.Second
)

type PIDResolver interface {
	ContainerPID(ctx context.Context, containerID string) (int, error)
}

type NetworkLatencyInjector struct {
	pidResolver    PIDResolver
	interfaceName  string
	nsenterBinary  string
	commandTimeout time.Duration
}

func NewNetworkLatencyInjector(pidResolver PIDResolver) *NetworkLatencyInjector {
	return &NetworkLatencyInjector{
		pidResolver:    pidResolver,
		interfaceName:  defaultInterfaceName,
		nsenterBinary:  defaultNsenterBinary,
		commandTimeout: defaultCommandTTL,
	}
}

func (n *NetworkLatencyInjector) InjectNetworkLatency(ctx context.Context, containerID string, delay time.Duration) error {
	containerID = strings.TrimSpace(containerID)
	if containerID == "" {
		return domainfault.ErrInvalidContainerID
	}
	if delay <= 0 {
		return domainfault.ErrInvalidLatencyDuration
	}
	if n.pidResolver == nil {
		return fmt.Errorf("pid resolver is required")
	}

	pid, err := n.pidResolver.ContainerPID(ctx, containerID)
	if err != nil {
		return fmt.Errorf("resolve pid for container %q: %w", containerID, err)
	}

	args := []string{"qdisc", "replace", "dev", n.interfaceName, "root", "netem", "delay", delay.String()}
	if err := n.runTC(ctx, pid, args); err != nil {
		return fmt.Errorf("inject latency into container %q: %w", containerID, err)
	}

	return nil
}

func (n *NetworkLatencyInjector) RevertNetworkLatency(ctx context.Context, containerID string) error {
	containerID = strings.TrimSpace(containerID)
	if containerID == "" {
		return domainfault.ErrInvalidContainerID
	}
	if n.pidResolver == nil {
		return fmt.Errorf("pid resolver is required")
	}

	pid, err := n.pidResolver.ContainerPID(ctx, containerID)
	if err != nil {
		return fmt.Errorf("resolve pid for container %q: %w", containerID, err)
	}

	args := []string{"qdisc", "del", "dev", n.interfaceName, "root"}
	err = n.runTC(ctx, pid, args)
	if err != nil && !isMissingQDisc(err) {
		return fmt.Errorf("revert latency in container %q: %w", containerID, err)
	}

	return nil
}

func (n *NetworkLatencyInjector) runTC(ctx context.Context, pid int, tcArgs []string) error {
	callCtx, cancel := context.WithTimeout(ctx, n.commandTimeout)
	defer cancel()

	nsenterArgs := []string{
		"--target", strconv.Itoa(pid),
		"--net",
		"--mount",
		"--",
		"tc",
	}
	nsenterArgs = append(nsenterArgs, tcArgs...)

	cmd := exec.CommandContext(callCtx, n.nsenterBinary, nsenterArgs...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		return nil
	}

	if errors.Is(callCtx.Err(), context.DeadlineExceeded) {
		return fmt.Errorf("%w: %s", domainfault.ErrCommandTimeout, strings.TrimSpace(stderr.String()))
	}

	if errors.Is(err, exec.ErrNotFound) {
		return fmt.Errorf("%w: binary %q not found on host", domainfault.ErrNamespaceToolMissing, n.nsenterBinary)
	}

	combined := strings.ToLower(strings.TrimSpace(stderr.String() + " " + stdout.String() + " " + err.Error()))
	switch {
	case strings.Contains(combined, "tc: not found"),
		strings.Contains(combined, "executable file not found"),
		strings.Contains(combined, "command not found"):
		// Sidecar Pattern: if the target image is distroless/scratch and lacks iproute2,
		// a privileged helper container can join the same network namespace and run tc.
		return fmt.Errorf("%w: container namespace does not expose tc/iproute2", domainfault.ErrIPRoute2Missing)
	case strings.Contains(combined, "cannot open network namespace"),
		strings.Contains(combined, "no such file or directory"):
		return fmt.Errorf("%w: %s", domainfault.ErrNetworkNamespaceUnavailable, strings.TrimSpace(stderr.String()))
	case strings.Contains(combined, "operation not permitted"),
		strings.Contains(combined, "permission denied"):
		return fmt.Errorf("%w: %s", domainfault.ErrInsufficientPrivileges, strings.TrimSpace(stderr.String()))
	default:
		return fmt.Errorf("%w: %v: %s", domainfault.ErrTCCommandFailed, err, strings.TrimSpace(stderr.String()))
	}
}

func isMissingQDisc(err error) bool {
	if err == nil {
		return false
	}
	raw := strings.ToLower(err.Error())
	return strings.Contains(raw, "no such file") || strings.Contains(raw, "cannot find qdisc")
}

