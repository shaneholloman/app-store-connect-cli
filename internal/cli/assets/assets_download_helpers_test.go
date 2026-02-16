package assets

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type readerThatFailsAfterFirstRead struct {
	readOnce bool
}

func (r *readerThatFailsAfterFirstRead) Read(p []byte) (int, error) {
	if !r.readOnce {
		r.readOnce = true
		return copy(p, "NEW-DATA"), nil
	}
	return 0, errors.New("simulated read failure")
}

func TestWriteDownloadedFile_Overwrite_ErrorPreservesExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.bin")

	if err := os.WriteFile(path, []byte("OLD-DATA"), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	_, err := writeDownloadedFile(path, &readerThatFailsAfterFirstRead{}, true)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	data, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("ReadFile() error: %v", readErr)
	}
	if string(data) != "OLD-DATA" {
		t.Fatalf("expected existing file contents preserved, got %q", string(data))
	}
}

func TestWriteDownloadedFile_Overwrite_ReplacesExistingFileOnSuccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.bin")

	if err := os.WriteFile(path, []byte("OLD-DATA"), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	written, err := writeDownloadedFile(path, strings.NewReader("NEW-DATA"), true)
	if err != nil {
		t.Fatalf("writeDownloadedFile() error: %v", err)
	}
	if written != int64(len("NEW-DATA")) {
		t.Fatalf("expected written=%d, got %d", len("NEW-DATA"), written)
	}

	data, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("ReadFile() error: %v", readErr)
	}
	if string(data) != "NEW-DATA" {
		t.Fatalf("expected new file contents, got %q", string(data))
	}
}
