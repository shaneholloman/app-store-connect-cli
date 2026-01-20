package asc

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func() error) string {
	t.Helper()

	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe error: %v", err)
	}
	os.Stdout = w

	err = fn()

	if closeErr := w.Close(); closeErr != nil {
		t.Fatalf("close error: %v", closeErr)
	}
	os.Stdout = orig

	var buf bytes.Buffer
	if _, readErr := io.Copy(&buf, r); readErr != nil {
		t.Fatalf("read error: %v", readErr)
	}
	if err != nil {
		t.Fatalf("function error: %v", err)
	}

	return buf.String()
}

func TestPrintTable_Feedback(t *testing.T) {
	resp := &FeedbackResponse{
		Data: []Resource[FeedbackAttributes]{
			{
				ID: "1",
				Attributes: FeedbackAttributes{
					CreatedDate: "2026-01-20T00:00:00Z",
					Email:       "tester@example.com",
					Comment:     "Looks good",
				},
			},
		},
	}

	output := captureStdout(t, func() error {
		return PrintTable(resp)
	})

	if !strings.Contains(output, "Created") || !strings.Contains(output, "Email") {
		t.Fatalf("expected header in output, got: %s", output)
	}
	if !strings.Contains(output, "tester@example.com") {
		t.Fatalf("expected email in output, got: %s", output)
	}
}

func TestPrintMarkdown_Reviews(t *testing.T) {
	resp := &ReviewsResponse{
		Data: []Resource[ReviewAttributes]{
			{
				ID: "1",
				Attributes: ReviewAttributes{
					CreatedDate: "2026-01-20T00:00:00Z",
					Rating:      5,
					Title:       "Great app",
					Territory:   "US",
				},
			},
		},
	}

	output := captureStdout(t, func() error {
		return PrintMarkdown(resp)
	})

	if !strings.Contains(output, "| Created | Rating |") {
		t.Fatalf("expected markdown header, got: %s", output)
	}
	if !strings.Contains(output, "Great app") {
		t.Fatalf("expected title in output, got: %s", output)
	}
}

func TestPrintTable_Apps(t *testing.T) {
	resp := &AppsResponse{
		Data: []Resource[AppAttributes]{
			{
				ID: "123",
				Attributes: AppAttributes{
					Name:     "Demo App",
					BundleID: "com.example.demo",
					SKU:      "SKU-1",
				},
			},
		},
	}

	output := captureStdout(t, func() error {
		return PrintTable(resp)
	})

	if !strings.Contains(output, "Bundle ID") {
		t.Fatalf("expected apps header in output, got: %s", output)
	}
	if !strings.Contains(output, "com.example.demo") {
		t.Fatalf("expected bundle ID in output, got: %s", output)
	}
}

func TestPrintMarkdown_Apps(t *testing.T) {
	resp := &AppsResponse{
		Data: []Resource[AppAttributes]{
			{
				ID: "123",
				Attributes: AppAttributes{
					Name:     "Demo App",
					BundleID: "com.example.demo",
					SKU:      "SKU-1",
				},
			},
		},
	}

	output := captureStdout(t, func() error {
		return PrintMarkdown(resp)
	})

	if !strings.Contains(output, "| ID | Name | Bundle ID | SKU |") {
		t.Fatalf("expected markdown header, got: %s", output)
	}
	if !strings.Contains(output, "Demo App") {
		t.Fatalf("expected app name in output, got: %s", output)
	}
}

func TestPrintTable_Builds(t *testing.T) {
	resp := &BuildsResponse{
		Data: []Resource[BuildAttributes]{
			{
				ID: "1",
				Attributes: BuildAttributes{
					Version:         "1.2.3",
					UploadedDate:    "2026-01-20T00:00:00Z",
					ProcessingState: "PROCESSING",
					Expired:         false,
				},
			},
		},
	}

	output := captureStdout(t, func() error {
		return PrintTable(resp)
	})

	if !strings.Contains(output, "Processing") {
		t.Fatalf("expected builds header in output, got: %s", output)
	}
	if !strings.Contains(output, "1.2.3") {
		t.Fatalf("expected build version in output, got: %s", output)
	}
}

func TestPrintMarkdown_Builds(t *testing.T) {
	resp := &BuildsResponse{
		Data: []Resource[BuildAttributes]{
			{
				ID: "1",
				Attributes: BuildAttributes{
					Version:         "1.2.3",
					UploadedDate:    "2026-01-20T00:00:00Z",
					ProcessingState: "PROCESSING",
					Expired:         false,
				},
			},
		},
	}

	output := captureStdout(t, func() error {
		return PrintMarkdown(resp)
	})

	if !strings.Contains(output, "| Version | Uploaded | Processing | Expired |") {
		t.Fatalf("expected markdown header, got: %s", output)
	}
	if !strings.Contains(output, "1.2.3") {
		t.Fatalf("expected build version in output, got: %s", output)
	}
}

func TestPrintTable_BuildInfo(t *testing.T) {
	resp := &BuildResponse{
		Data: Resource[BuildAttributes]{
			ID: "1",
			Attributes: BuildAttributes{
				Version:         "2.0.0",
				UploadedDate:    "2026-01-20T00:00:00Z",
				ProcessingState: "VALID",
				Expired:         true,
			},
		},
	}

	output := captureStdout(t, func() error {
		return PrintTable(resp)
	})

	if !strings.Contains(output, "Processing") {
		t.Fatalf("expected build info header in output, got: %s", output)
	}
	if !strings.Contains(output, "2.0.0") {
		t.Fatalf("expected build version in output, got: %s", output)
	}
}

func TestPrintMarkdown_BuildInfo(t *testing.T) {
	resp := &BuildResponse{
		Data: Resource[BuildAttributes]{
			ID: "1",
			Attributes: BuildAttributes{
				Version:         "2.0.0",
				UploadedDate:    "2026-01-20T00:00:00Z",
				ProcessingState: "VALID",
				Expired:         true,
			},
		},
	}

	output := captureStdout(t, func() error {
		return PrintMarkdown(resp)
	})

	if !strings.Contains(output, "| Version | Uploaded | Processing | Expired |") {
		t.Fatalf("expected markdown header, got: %s", output)
	}
	if !strings.Contains(output, "2.0.0") {
		t.Fatalf("expected build version in output, got: %s", output)
	}
}

func TestPrintPrettyJSON(t *testing.T) {
	resp := &ReviewsResponse{
		Data: []Resource[ReviewAttributes]{
			{
				ID: "1",
				Attributes: ReviewAttributes{
					CreatedDate: "2026-01-20T00:00:00Z",
					Rating:      5,
					Title:       "Great app",
					Body:        "Nice work",
					Territory:   "US",
				},
			},
		},
	}

	output := captureStdout(t, func() error {
		return PrintPrettyJSON(resp)
	})

	if !strings.Contains(output, "\n  \"data\"") {
		t.Fatalf("expected pretty JSON indentation, got: %s", output)
	}
}
