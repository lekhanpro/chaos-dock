package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadChaosConfig_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "chaos.yaml")

	content := `
experiments:
  - name: db-latency
    targetContainer: postgres
    enabled: true
    fault:
      type: network-latency
      delay: 300ms
    schedule:
      every: 30s
      jitter: 5s
  - name: kill-db
    targetContainer: postgres
    enabled: true
    fault:
      type: kill
      signal: SIGTERM
    schedule:
      every: 60s
`
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	cfg, err := LoadChaosConfig(path)
	if err != nil {
		t.Fatalf("LoadChaosConfig returned error: %v", err)
	}
	if len(cfg.Experiments) != 2 {
		t.Fatalf("expected 2 experiments, got %d", len(cfg.Experiments))
	}
}

func TestLoadChaosConfig_InvalidSignal(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "chaos.yaml")

	content := `
experiments:
  - name: kill-db
    targetContainer: postgres
    enabled: true
    fault:
      type: kill
      signal: SIGBOGUS
    schedule:
      every: 60s
`
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, err := LoadChaosConfig(path)
	if err == nil || !strings.Contains(err.Error(), "fault.signal") {
		t.Fatalf("expected fault.signal validation error, got %v", err)
	}
}
