package asc

import (
	"bytes"
	"net/url"
	"strings"
	"testing"
)

func TestBuildReviewQuery(t *testing.T) {
	query := buildReviewQuery([]ReviewOption{
		WithRating(5),
		WithTerritory("us"),
		WithLimit(25),
	})

	values, err := url.ParseQuery(query)
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	if got := values.Get("filter[rating]"); got != "5" {
		t.Fatalf("expected filter[rating]=5, got %q", got)
	}

	if got := values.Get("filter[territory]"); got != "US" {
		t.Fatalf("expected filter[territory]=US, got %q", got)
	}

	if got := values.Get("limit"); got != "25" {
		t.Fatalf("expected limit=25, got %q", got)
	}
}

func TestBuildReviewQuery_InvalidRating(t *testing.T) {
	query := buildReviewQuery([]ReviewOption{
		WithRating(9),
	})

	values, err := url.ParseQuery(query)
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	if got := values.Get("filter[rating]"); got != "" {
		t.Fatalf("expected empty filter[rating], got %q", got)
	}
}

func TestBuildFeedbackQuery(t *testing.T) {
	query := &feedbackQuery{}
	opts := []FeedbackOption{
		WithFeedbackDeviceModels([]string{"iPhone15,3", " iPhone15,2 "}),
		WithFeedbackOSVersions([]string{"17.2", ""}),
		WithFeedbackLimit(10),
	}
	for _, opt := range opts {
		opt(query)
	}

	values, err := url.ParseQuery(buildFeedbackQuery(query))
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	if got := values.Get("filter[deviceModel]"); got != "iPhone15,3,iPhone15,2" {
		t.Fatalf("expected filter[deviceModel] to be CSV, got %q", got)
	}
	if got := values.Get("filter[osVersion]"); got != "17.2" {
		t.Fatalf("expected filter[osVersion]=17.2, got %q", got)
	}
	if got := values.Get("limit"); got != "10" {
		t.Fatalf("expected limit=10, got %q", got)
	}
}

func TestBuildCrashQuery(t *testing.T) {
	query := &crashQuery{}
	opts := []CrashOption{
		WithCrashDeviceModels([]string{"iPhone16,1"}),
		WithCrashOSVersions([]string{"18.0"}),
		WithCrashLimit(5),
	}
	for _, opt := range opts {
		opt(query)
	}

	values, err := url.ParseQuery(buildCrashQuery(query))
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	if got := values.Get("filter[deviceModel]"); got != "iPhone16,1" {
		t.Fatalf("expected filter[deviceModel]=iPhone16,1, got %q", got)
	}
	if got := values.Get("filter[osVersion]"); got != "18.0" {
		t.Fatalf("expected filter[osVersion]=18.0, got %q", got)
	}
	if got := values.Get("limit"); got != "5" {
		t.Fatalf("expected limit=5, got %q", got)
	}
}

func TestBuildRequestBody(t *testing.T) {
	body, err := BuildRequestBody(map[string]string{"hello": "world"})
	if err != nil {
		t.Fatalf("BuildRequestBody() error: %v", err)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(body); err != nil {
		t.Fatalf("read body error: %v", err)
	}

	if !strings.Contains(buf.String(), `"hello":"world"`) {
		t.Fatalf("unexpected body: %s", buf.String())
	}
}

func TestParseError(t *testing.T) {
	payload := []byte(`{"errors":[{"code":"FORBIDDEN","title":"Forbidden","detail":"not allowed"}]}`)
	err := ParseError(payload)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Forbidden") {
		t.Fatalf("unexpected error message: %v", err)
	}
}
