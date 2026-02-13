package safety

import (
	"context"
	"testing"
	"time"
)

type mockInjector struct {
	reverted []string
}

func (m *mockInjector) InjectNetworkLatency(_ context.Context, _ string, _ time.Duration) error {
	return nil
}

func (m *mockInjector) RevertNetworkLatency(_ context.Context, containerID string) error {
	m.reverted = append(m.reverted, containerID)
	return nil
}

type mockRestarter struct {
	restarted []string
}

func (m *mockRestarter) Restart(_ context.Context, containerID string) error {
	m.restarted = append(m.restarted, containerID)
	return nil
}

func TestPanicButton_TriggerAllUsesRegistry(t *testing.T) {
	registry := NewTargetRegistry()
	registry.Mark("db")
	registry.Mark("api")

	injector := &mockInjector{}
	restarter := &mockRestarter{}

	button := &PanicButton{
		Injector:  injector,
		Restarter: restarter,
		Registry:  registry,
	}

	if err := button.TriggerAll(context.Background()); err != nil {
		t.Fatalf("TriggerAll returned error: %v", err)
	}
	if len(injector.reverted) != 2 {
		t.Fatalf("expected 2 revert calls, got %d", len(injector.reverted))
	}
	if len(restarter.restarted) != 2 {
		t.Fatalf("expected 2 restart calls, got %d", len(restarter.restarted))
	}
	if len(registry.Snapshot()) != 0 {
		t.Fatalf("expected registry reset after panic trigger")
	}
}
