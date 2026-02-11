package asc

import (
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
	"testing"
)

func TestValidateScreenshotDimensionsValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "valid.png")
	writePNG(t, path, 640, 960)

	if err := ValidateScreenshotDimensions(path, "APP_IPHONE_35"); err != nil {
		t.Fatalf("expected valid dimensions, got %v", err)
	}
}

func TestValidateScreenshotDimensionsInvalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.png")
	writePNG(t, path, 100, 100)

	err := ValidateScreenshotDimensions(path, "APP_IPHONE_35")
	if err == nil {
		t.Fatal("expected dimension validation error, got nil")
	}
	message := err.Error()
	if !strings.Contains(message, "100x100") {
		t.Fatalf("expected actual size in error, got %q", message)
	}
	if !strings.Contains(message, "640x960") {
		t.Fatalf("expected allowed size in error, got %q", message)
	}
	if !strings.Contains(message, "asc assets screenshots sizes") {
		t.Fatalf("expected hint in error, got %q", message)
	}
}

func TestScreenshotDisplayTypesMatchOpenAPI(t *testing.T) {
	specTypes := openAPIScreenshotDisplayTypes(t)
	codeTypes := ScreenshotDisplayTypes()
	sort.Strings(codeTypes)

	if !slices.Equal(specTypes, codeTypes) {
		t.Fatalf("screenshot display types drifted from OpenAPI: spec=%v code=%v", specTypes, codeTypes)
	}
}

func openAPIScreenshotDisplayTypes(t *testing.T) []string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve test file path")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	path := filepath.Join(root, "docs", "openapi", "latest.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read openapi: %v", err)
	}

	var spec struct {
		Components struct {
			Schemas map[string]struct {
				Enum []string `json:"enum"`
			} `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("parse openapi: %v", err)
	}
	entry, ok := spec.Components.Schemas["ScreenshotDisplayType"]
	if !ok || len(entry.Enum) == 0 {
		t.Fatal("missing ScreenshotDisplayType enum in OpenAPI")
	}
	enum := append([]string(nil), entry.Enum...)
	sort.Strings(enum)
	return enum
}

func writePNG(t *testing.T, path string, width, height int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create image: %v", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
}
