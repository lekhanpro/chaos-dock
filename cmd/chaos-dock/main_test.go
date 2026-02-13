package main

import (
	"path/filepath"
	"testing"
)

func TestSplitCSV(t *testing.T) {
	items := splitCSV("api, db ,,redis")
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	if items[0] != "api" || items[1] != "db" || items[2] != "redis" {
		t.Fatalf("unexpected CSV parse output: %#v", items)
	}
}

func TestWriteDefaultConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "chaos.yaml")

	if err := writeDefaultConfig(path, false); err != nil {
		t.Fatalf("writeDefaultConfig returned error: %v", err)
	}

	if err := writeDefaultConfig(path, false); err == nil {
		t.Fatalf("expected error when writing existing config without force")
	}

	if err := writeDefaultConfig(path, true); err != nil {
		t.Fatalf("writeDefaultConfig(force) returned error: %v", err)
	}
}
