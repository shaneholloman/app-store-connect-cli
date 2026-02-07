package cmdtest

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestSubscriptionsPricePointsListPaginateUsesPerPageTimeout(t *testing.T) {
	setupAuth(t)
	t.Setenv("ASC_TIMEOUT", "120ms")

	originalTransport := http.DefaultTransport
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	requests := 0
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		time.Sleep(70 * time.Millisecond)

		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}

		switch req.URL.RawQuery {
		case "limit=200":
			if req.URL.Path != "/v1/subscriptions/sub-1/pricePoints" {
				t.Fatalf("unexpected first page path: %s", req.URL.Path)
			}
			body := `{"data":[{"type":"subscriptionPricePoints","id":"pp-1"}],"links":{"next":"https://api.appstoreconnect.apple.com/v1/subscriptions/sub-1/pricePoints?cursor=AQ&limit=200"}}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		case "cursor=AQ&limit=200":
			if req.URL.Path != "/v1/subscriptions/sub-1/pricePoints" {
				t.Fatalf("unexpected second page path: %s", req.URL.Path)
			}
			body := `{"data":[{"type":"subscriptionPricePoints","id":"pp-2"}],"links":{"next":"https://api.appstoreconnect.apple.com/v1/subscriptions/sub-1/pricePoints?cursor=BQ&limit=200"}}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		case "cursor=BQ&limit=200":
			if req.URL.Path != "/v1/subscriptions/sub-1/pricePoints" {
				t.Fatalf("unexpected third page path: %s", req.URL.Path)
			}
			body := `{"data":[{"type":"subscriptionPricePoints","id":"pp-3"}],"links":{}}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		default:
			t.Fatalf("unexpected request path/query: %s?%s", req.URL.Path, req.URL.RawQuery)
			return nil, nil
		}
	})

	root := RootCommand("1.2.3")
	root.FlagSet.SetOutput(io.Discard)

	stdout, stderr := captureOutput(t, func() {
		if err := root.Parse([]string{
			"subscriptions", "price-points", "list",
			"--subscription-id", "sub-1",
			"--paginate",
			"--output", "json",
		}); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if err := root.Run(context.Background()); err != nil {
			t.Fatalf("run error: %v", err)
		}
	})

	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	if requests != 3 {
		t.Fatalf("expected 3 paginated requests, got %d", requests)
	}
	if !strings.Contains(stdout, `"id":"pp-1"`) || !strings.Contains(stdout, `"id":"pp-2"`) || !strings.Contains(stdout, `"id":"pp-3"`) {
		t.Fatalf("expected aggregated paginated output, got %q", stdout)
	}
}
