package settings

import (
	"path/filepath"
	"testing"
)

func TestLoadDefaultsWhenSettingsMissing(t *testing.T) {
	store := NewStore(t.TempDir())

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.PreferredPreset != "codex" {
		t.Fatalf("PreferredPreset = %q, want codex", cfg.PreferredPreset)
	}
	if !cfg.PreferBundledASC {
		t.Fatalf("PreferBundledASC = false, want true")
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	store := NewStore(t.TempDir())
	input := DefaultSettings()
	input.AgentCommand = "codex-acp"
	input.AgentArgs = []string{"serve"}
	input.SystemASCPath = "/usr/local/bin/asc"
	input.WorkspaceRoot = filepath.Join(t.TempDir(), "workspace")

	if err := store.Save(input); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.AgentCommand != input.AgentCommand {
		t.Fatalf("AgentCommand = %q, want %q", got.AgentCommand, input.AgentCommand)
	}
	if got.SystemASCPath != input.SystemASCPath {
		t.Fatalf("SystemASCPath = %q, want %q", got.SystemASCPath, input.SystemASCPath)
	}
	if got.WorkspaceRoot != input.WorkspaceRoot {
		t.Fatalf("WorkspaceRoot = %q, want %q", got.WorkspaceRoot, input.WorkspaceRoot)
	}
}
