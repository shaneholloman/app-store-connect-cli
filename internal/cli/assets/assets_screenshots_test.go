package assets

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
)

func TestAssetsScreenshotsSizesCommandFilter(t *testing.T) {
	cmd := AssetsScreenshotsSizesCommand()
	cmd.FlagSet.SetOutput(io.Discard)
	if err := cmd.FlagSet.Parse([]string{"--display-type", "APP_IPHONE_65"}); err != nil {
		t.Fatalf("parse error: %v", err)
	}

	stdout, stderr := captureOutput(t, func() {
		if err := cmd.Exec(context.Background(), cmd.FlagSet.Args()); err != nil {
			t.Fatalf("exec error: %v", err)
		}
	})

	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}

	var result asc.ScreenshotSizesResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if len(result.Sizes) != 1 {
		t.Fatalf("expected 1 size entry, got %d", len(result.Sizes))
	}
	if result.Sizes[0].DisplayType != "APP_IPHONE_65" {
		t.Fatalf("expected APP_IPHONE_65, got %q", result.Sizes[0].DisplayType)
	}
}

func captureOutput(t *testing.T, fn func()) (string, string) {
	t.Helper()

	origStdout := os.Stdout
	origStderr := os.Stderr
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}

	os.Stdout = wOut
	os.Stderr = wErr

	fn()

	_ = wOut.Close()
	_ = wErr.Close()
	os.Stdout = origStdout
	os.Stderr = origStderr

	outBytes, _ := io.ReadAll(rOut)
	errBytes, _ := io.ReadAll(rErr)
	_ = rOut.Close()
	_ = rErr.Close()

	return string(outBytes), string(errBytes)
}
