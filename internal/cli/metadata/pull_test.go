package metadata

import (
	"testing"
)

func TestMetadataPullCommand_AppInfoFlagDefined(t *testing.T) {
	cmd := MetadataPullCommand()
	if cmd.FlagSet.Lookup("app-info") == nil {
		t.Fatal("expected --app-info flag to be defined on metadata pull")
	}
}

func TestMetadataPullCommand_AppInfoFlagDefault(t *testing.T) {
	cmd := MetadataPullCommand()
	f := cmd.FlagSet.Lookup("app-info")
	if f == nil {
		t.Fatal("expected --app-info flag to be defined")
	}
	if f.DefValue != "" {
		t.Fatalf("expected --app-info default to be empty, got %q", f.DefValue)
	}
}

func TestBuildMetadataPullAppInfoExample(t *testing.T) {
	tests := []struct {
		name     string
		appID    string
		version  string
		platform string
		dir      string
		want     string
	}{
		{
			name:    "basic",
			appID:   "123",
			version: "1.0",
			dir:     "./metadata",
			want:    `asc metadata pull --app "123" --version "1.0" --dir "./metadata" --app-info "info-1"`,
		},
		{
			name:     "with platform",
			appID:    "123",
			version:  "1.0",
			platform: "IOS",
			dir:      "./metadata",
			want:     `asc metadata pull --app "123" --version "1.0" --platform IOS --dir "./metadata" --app-info "info-1"`,
		},
		{
			name:    "empty dir uses default",
			appID:   "123",
			version: "1.0",
			dir:     "",
			want:    `asc metadata pull --app "123" --version "1.0" --dir "./metadata" --app-info "info-1"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildMetadataPullAppInfoExample(tc.appID, tc.version, tc.platform, tc.dir, "info-1")
			if got != tc.want {
				t.Fatalf("buildMetadataPullAppInfoExample() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestMetadataPullCommand_ShortUsageMentionsAppInfo(t *testing.T) {
	cmd := MetadataPullCommand()
	if cmd.ShortUsage == "" {
		t.Fatal("expected non-empty ShortUsage")
	}
	if !contains(cmd.ShortUsage, "--app-info") {
		t.Fatalf("expected ShortUsage to mention --app-info, got %q", cmd.ShortUsage)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchContains(s, substr)
}

func searchContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
