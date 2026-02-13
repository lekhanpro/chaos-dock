package safety

import "testing"

func TestTargetRegistry_MarkSnapshotReset(t *testing.T) {
	registry := NewTargetRegistry()
	registry.Mark("api")
	registry.Mark("db")
	registry.Mark("api")

	snapshot := registry.Snapshot()
	if len(snapshot) != 2 {
		t.Fatalf("expected 2 unique targets, got %d", len(snapshot))
	}
	if snapshot[0] != "api" || snapshot[1] != "db" {
		t.Fatalf("unexpected snapshot ordering/content: %#v", snapshot)
	}

	registry.Reset()
	if len(registry.Snapshot()) != 0 {
		t.Fatalf("expected empty registry after reset")
	}
}
