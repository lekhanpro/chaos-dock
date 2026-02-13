package fault

import (
	"context"
	"errors"
	"testing"

	domainfault "github.com/lekhanpro/chaos-dock/internal/domain/fault"
)

type mockKillExecutor struct {
	lastContainerID string
	lastSignal      string
	err             error
}

func (m *mockKillExecutor) Kill(_ context.Context, containerID string, signal string) error {
	m.lastContainerID = containerID
	m.lastSignal = signal
	return m.err
}

func TestContainerKillInjector_DefaultSignal(t *testing.T) {
	exec := &mockKillExecutor{}
	injector := NewContainerKillInjector(exec)

	err := injector.KillContainer(context.Background(), "api", "")
	if err != nil {
		t.Fatalf("KillContainer returned error: %v", err)
	}
	if exec.lastSignal != "SIGKILL" {
		t.Fatalf("expected default SIGKILL, got %q", exec.lastSignal)
	}
}

func TestContainerKillInjector_NormalizesSignalPrefix(t *testing.T) {
	exec := &mockKillExecutor{}
	injector := NewContainerKillInjector(exec)

	err := injector.KillContainer(context.Background(), "api", "term")
	if err != nil {
		t.Fatalf("KillContainer returned error: %v", err)
	}
	if exec.lastSignal != "SIGTERM" {
		t.Fatalf("expected SIGTERM, got %q", exec.lastSignal)
	}
}

func TestContainerKillInjector_InvalidSignal(t *testing.T) {
	exec := &mockKillExecutor{}
	injector := NewContainerKillInjector(exec)

	err := injector.KillContainer(context.Background(), "api", "SIGFAKE")
	if !errors.Is(err, domainfault.ErrInvalidKillSignal) {
		t.Fatalf("expected ErrInvalidKillSignal, got %v", err)
	}
}

func TestContainerKillInjector_WrapsExecutorErrors(t *testing.T) {
	exec := &mockKillExecutor{err: errors.New("docker failure")}
	injector := NewContainerKillInjector(exec)

	err := injector.KillContainer(context.Background(), "api", "SIGTERM")
	if !errors.Is(err, domainfault.ErrContainerKillFailed) {
		t.Fatalf("expected ErrContainerKillFailed, got %v", err)
	}
}
