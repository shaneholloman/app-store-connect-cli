package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rudrankriyam/App-Store-Connect-CLI/apps/studio/internal/studio/settings"
)

func TestCheckAuthStatusTreatsEmptyStoredCredentialsAsUnauthenticated(t *testing.T) {
	app := newCheckAuthStatusTestApp(t, `{
		"storageBackend":"System Keychain",
		"storageLocation":"system keychain",
		"profile":"",
		"environmentCredentialsComplete":false,
		"credentials":[]
	}`)

	status, err := app.CheckAuthStatus()
	if err != nil {
		t.Fatalf("CheckAuthStatus() error = %v", err)
	}
	if status.Authenticated {
		t.Fatal("status.Authenticated = true, want false when no stored or environment credentials exist")
	}
	if status.Storage != "System Keychain" {
		t.Fatalf("status.Storage = %q, want System Keychain", status.Storage)
	}
}

func TestCheckAuthStatusTreatsCompleteEnvironmentCredentialsAsAuthenticated(t *testing.T) {
	app := newCheckAuthStatusTestApp(t, `{
		"storageBackend":"Config File",
		"storageLocation":"/tmp/config.json",
		"profile":"",
		"environmentCredentialsComplete":true,
		"credentials":[]
	}`)

	status, err := app.CheckAuthStatus()
	if err != nil {
		t.Fatalf("CheckAuthStatus() error = %v", err)
	}
	if !status.Authenticated {
		t.Fatal("status.Authenticated = false, want true when complete environment credentials are available")
	}
	if status.Storage != "Config File" {
		t.Fatalf("status.Storage = %q, want Config File", status.Storage)
	}
}

func newCheckAuthStatusTestApp(t *testing.T, output string) *App {
	t.Helper()

	rootDir := t.TempDir()
	t.Setenv("HOME", rootDir)

	ascPath := filepath.Join(rootDir, "asc")
	script := "#!/bin/sh\ncat <<'EOF'\n" + output + "\nEOF\n"
	if err := os.WriteFile(ascPath, []byte(script), 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	settingsStore := settings.NewStore(rootDir)
	if err := settingsStore.Save(settings.StudioSettings{
		SystemASCPath:    ascPath,
		PreferBundledASC: false,
	}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	return &App{
		rootDir:  rootDir,
		settings: settingsStore,
	}
}
