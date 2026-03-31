package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (a *App) CheckAuthStatus() (AuthStatus, error) {
	defer configGuard()()
	ascPath, err := a.resolveASCPath()
	if err != nil {
		return AuthStatus{RawOutput: "Could not find asc binary: " + err.Error()}, nil
	}

	ctx, cancel := context.WithTimeout(a.contextOrBackground(), 10*time.Second)
	defer cancel()

	cmd := a.newASCCommand(ctx, ascPath, "auth", "status", "--output", "json")
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))

	status := AuthStatus{RawOutput: output}

	if err != nil {
		status.Authenticated = false
		return status, nil
	}

	var jsonStatus struct {
		StorageBackend                 string `json:"storageBackend"`
		StorageLocation                string `json:"storageLocation"`
		Profile                        string `json:"profile"`
		EnvironmentCredentialsComplete bool   `json:"environmentCredentialsComplete"`
		Credentials                    []struct {
			Name      string `json:"name"`
			KeyID     string `json:"keyId"`
			IsDefault bool   `json:"isDefault"`
		} `json:"credentials"`
	}
	if json.Unmarshal([]byte(output), &jsonStatus) == nil {
		status.Storage = jsonStatus.StorageBackend
		status.Profile = jsonStatus.Profile
		status.Authenticated = len(jsonStatus.Credentials) > 0 || jsonStatus.EnvironmentCredentialsComplete
		for _, cred := range jsonStatus.Credentials {
			if cred.IsDefault {
				status.Profile = cred.Name
				break
			}
		}
		if status.Profile == "" && len(jsonStatus.Credentials) > 0 {
			status.Profile = jsonStatus.Credentials[0].Name
		}
		return status, nil
	}

	// Older asc builds may still emit non-JSON output for auth status; preserve
	// the prior success-path behavior in that fallback case.
	status.Authenticated = true
	return status, nil
}

// cacheAuthFromConfig reads auth credentials from config once and caches them
// so that subsequent asc commands don't depend on config.json staying intact.
func (a *App) cacheAuthFromConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	data, err := os.ReadFile(filepath.Join(home, ".asc", "config.json"))
	if err != nil {
		return
	}
	var cfg struct {
		KeyID          string `json:"key_id"`
		IssuerID       string `json:"issuer_id"`
		PrivateKeyPath string `json:"private_key_path"`
		DefaultKeyName string `json:"default_key_name"`
		Keys           []struct {
			Name           string `json:"name"`
			KeyID          string `json:"key_id"`
			IssuerID       string `json:"issuer_id"`
			PrivateKeyPath string `json:"private_key_path"`
		} `json:"keys"`
	}
	if json.Unmarshal(data, &cfg) != nil {
		return
	}
	// Prefer named key matching default_key_name
	for _, k := range cfg.Keys {
		if strings.TrimSpace(k.Name) == strings.TrimSpace(cfg.DefaultKeyName) && k.KeyID != "" {
			a.cachedKeyID = k.KeyID
			a.cachedIssuerID = k.IssuerID
			a.cachedPrivateKeyPath = k.PrivateKeyPath
			return
		}
	}
	// Fallback to top-level fields
	if cfg.KeyID != "" {
		a.cachedKeyID = cfg.KeyID
		a.cachedIssuerID = cfg.IssuerID
		a.cachedPrivateKeyPath = cfg.PrivateKeyPath
	}
}

// configGuard saves a snapshot of ~/.asc/config.json before running an asc
// command and restores it afterwards if the command mutated the file.
// This defends against CLI auth codepaths that accidentally wipe credentials
// during read-only operations (a known issue in the auth resolver).
func configGuard() func() {
	home, err := os.UserHomeDir()
	if err != nil {
		return func() {}
	}
	path := filepath.Join(home, ".asc", "config.json")

	ascConfigGuard.mu.Lock()
	if ascConfigGuard.active == 0 || ascConfigGuard.path != path {
		original, err := os.ReadFile(path)
		ascConfigGuard.path = path
		if err != nil {
			ascConfigGuard.original = nil
			ascConfigGuard.valid = false
		} else {
			ascConfigGuard.original = append(ascConfigGuard.original[:0], original...)
			ascConfigGuard.valid = true
		}
	}
	ascConfigGuard.active++
	valid := ascConfigGuard.valid
	original := append([]byte(nil), ascConfigGuard.original...)
	ascConfigGuard.mu.Unlock()

	return func() {
		ascConfigGuard.mu.Lock()
		ascConfigGuard.active--
		shouldRestore := ascConfigGuard.active == 0 && valid
		if ascConfigGuard.active == 0 {
			ascConfigGuard.original = nil
			ascConfigGuard.valid = false
			ascConfigGuard.path = ""
		}
		ascConfigGuard.mu.Unlock()

		if !shouldRestore {
			return
		}

		current, err := os.ReadFile(path)
		if err != nil || bytes.Equal(current, original) {
			return
		}
		_ = os.WriteFile(path, original, 0o600)
	}
}
