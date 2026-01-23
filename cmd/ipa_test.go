package cmd

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"howett.net/plist"
)

func TestExtractBundleInfoFromIPA(t *testing.T) {
	plistData := buildInfoPlist(t, "1.2.3", "45")
	ipaPath := writeTestIPA(t, map[string][]byte{
		"Payload/Demo.app/Info.plist": plistData,
	})

	info, err := extractBundleInfoFromIPA(ipaPath)
	if err != nil {
		t.Fatalf("extractBundleInfoFromIPA() error: %v", err)
	}
	if info.Version != "1.2.3" {
		t.Fatalf("expected version 1.2.3, got %q", info.Version)
	}
	if info.BuildNumber != "45" {
		t.Fatalf("expected build number 45, got %q", info.BuildNumber)
	}
}

func TestExtractBundleInfoFromIPA_PrefersTopLevelApp(t *testing.T) {
	plistData := buildInfoPlist(t, "2.0.0", "200")
	extensionPlist := buildInfoPlist(t, "9.9.9", "999")
	ipaPath := writeTestIPA(t, map[string][]byte{
		"Payload/Demo.app/Info.plist":                              plistData,
		"Payload/Demo.app/PlugIns/Widget.appex/Info.plist":         extensionPlist,
		"Payload/Demo.app/PlugIns/Widget.appex/Other.plist":        []byte("ignored"),
		"Payload/Demo.app/PlugIns/Widget.appex/Info.plist.bak":     []byte("ignored"),
		"Payload/Demo.app/Frameworks/Demo.framework/Info.plist":    extensionPlist,
		"Payload/Demo.app/Frameworks/Another.framework/Info.plist": extensionPlist,
	})

	info, err := extractBundleInfoFromIPA(ipaPath)
	if err != nil {
		t.Fatalf("extractBundleInfoFromIPA() error: %v", err)
	}
	if info.Version != "2.0.0" {
		t.Fatalf("expected version 2.0.0, got %q", info.Version)
	}
	if info.BuildNumber != "200" {
		t.Fatalf("expected build number 200, got %q", info.BuildNumber)
	}
}

func TestExtractBundleInfoFromIPA_MissingInfoPlist(t *testing.T) {
	ipaPath := writeTestIPA(t, map[string][]byte{
		"Payload/Demo.app/README.txt": []byte("no plist"),
	})

	_, err := extractBundleInfoFromIPA(ipaPath)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "Info.plist not found") {
		t.Fatalf("expected Info.plist not found error, got %q", err.Error())
	}
}

func writeTestIPA(t *testing.T, files map[string][]byte) string {
	t.Helper()

	ipaPath := filepath.Join(t.TempDir(), "app.ipa")
	file, err := os.Create(ipaPath)
	if err != nil {
		t.Fatalf("create IPA: %v", err)
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	for name, data := range files {
		entry, err := zipWriter.Create(name)
		if err != nil {
			t.Fatalf("create zip entry %q: %v", name, err)
		}
		if _, err := entry.Write(data); err != nil {
			t.Fatalf("write zip entry %q: %v", name, err)
		}
	}
	if err := zipWriter.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}

	return ipaPath
}

func buildInfoPlist(t *testing.T, version string, build string) []byte {
	t.Helper()

	var payload struct {
		ShortVersion string `plist:"CFBundleShortVersionString"`
		BuildVersion string `plist:"CFBundleVersion"`
	}
	payload.ShortVersion = version
	payload.BuildVersion = build

	data, err := plist.Marshal(payload, plist.XMLFormat)
	if err != nil {
		t.Fatalf("marshal plist: %v", err)
	}
	return data
}
