package cmdtest

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	cmd "github.com/rudrankriyam/App-Store-Connect-CLI/cmd"
	webcmd "github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/web"
	webcore "github.com/rudrankriyam/App-Store-Connect-CLI/internal/web"
)

func TestWebSubscriptionsAvailabilityRemoveFromSaleRunWithAppSelector(t *testing.T) {
	restoreSession := webcmd.SetResolveWebSession(func(ctx context.Context, appleID, password, twoFactorCode string) (*webcore.AuthSession, string, error) {
		return &webcore.AuthSession{
			Client: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return webSubscriptionsAvailabilityResponse(t, req)
			})},
		}, "cache", nil
	})
	t.Cleanup(restoreSession)

	stdout, stderr := captureOutput(t, func() {
		code := cmd.Run([]string{
			"--profile", "test-web",
			"web", "subscriptions", "availability", "remove-from-sale",
			"--output", "json",
			"--app", "app-1",
			"--subscription-id", "availability",
			"--confirm",
		}, "1.0.0")
		if code != cmd.ExitSuccess {
			t.Fatalf("exit code = %d, want %d", code, cmd.ExitSuccess)
		}
	})
	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}

	var payload struct {
		SubscriptionID            string   `json:"subscriptionId"`
		PlanAvailabilityID        string   `json:"planAvailabilityId"`
		RemovedFromSale           bool     `json:"removedFromSale"`
		AvailableInNewTerritories bool     `json:"availableInNewTerritories"`
		AvailableTerritories      []string `json:"availableTerritories"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error: %v; stdout=%q", err, stdout)
	}
	if payload.SubscriptionID != "sub-1" || payload.PlanAvailabilityID != "plan-1" || !payload.RemovedFromSale {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if payload.AvailableInNewTerritories || len(payload.AvailableTerritories) != 0 {
		t.Fatalf("expected subscription to be removed from sale, got %+v", payload)
	}
}

func TestWebSubscriptionsAvailabilityRemoveFromSaleRunUsageErrors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name: "missing subscription id",
			args: []string{
				"web", "subscriptions", "availability", "remove-from-sale",
				"--confirm",
			},
			wantErr: "--subscription-id is required",
		},
		{
			name: "missing confirm",
			args: []string{
				"web", "subscriptions", "availability", "remove-from-sale",
				"--subscription-id", "sub-1",
			},
			wantErr: "--confirm is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, stderr := captureOutput(t, func() {
				code := cmd.Run(test.args, "1.0.0")
				if code != cmd.ExitUsage {
					t.Fatalf("exit code = %d, want %d", code, cmd.ExitUsage)
				}
			})
			if !strings.Contains(stderr, test.wantErr) {
				t.Fatalf("expected stderr to contain %q, got %q", test.wantErr, stderr)
			}
		})
	}
}

func webSubscriptionsAvailabilityResponse(t *testing.T, req *http.Request) (*http.Response, error) {
	t.Helper()

	switch {
	case req.Method == http.MethodGet && req.URL.Path == "/iris/v1/apps/app-1/subscriptionGroups":
		if req.URL.Query().Get("include") != "subscriptions" {
			t.Fatalf("expected subscriptions include, got %q", req.URL.RawQuery)
		}
		return webSubscriptionsJSONResponse(`{
			"data": [{
				"id": "group-1",
				"type": "subscriptionGroups",
				"attributes": {"referenceName": "Premium"},
				"relationships": {
					"subscriptions": {
						"data": [{"type": "subscriptions", "id": "sub-1"}]
					}
				}
			}],
			"included": [{
				"id": "sub-1",
				"type": "subscriptions",
				"attributes": {
					"productId": "availability",
					"name": "Monthly",
					"state": "APPROVED"
				}
			}]
		}`), nil
	case req.Method == http.MethodGet && req.URL.Path == "/iris/v1/subscriptions/sub-1/planAvailabilities":
		return webSubscriptionsJSONResponse(`{
			"data": [{
				"id": "plan-1",
				"type": "subscriptionPlanAvailabilities",
				"attributes": {
					"availableInNewTerritories": false,
					"planType": "UPFRONT"
				},
				"relationships": {
					"availableTerritories": {"data": []}
				}
			}]
		}`), nil
	case req.Method == http.MethodPatch && req.URL.Path == "/iris/v1/subscriptionPlanAvailabilities/plan-1":
		rawBody, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		var payload struct {
			Data struct {
				Attributes struct {
					AvailableInNewTerritories bool `json:"availableInNewTerritories"`
				} `json:"attributes"`
				Relationships struct {
					AvailableTerritories struct {
						Data []any `json:"data"`
					} `json:"availableTerritories"`
				} `json:"relationships"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rawBody, &payload); err != nil {
			t.Fatalf("decode request body: %v\nbody=%s", err, string(rawBody))
		}
		if payload.Data.Attributes.AvailableInNewTerritories {
			t.Fatal("expected availableInNewTerritories=false")
		}
		if len(payload.Data.Relationships.AvailableTerritories.Data) != 0 {
			t.Fatalf("expected availableTerritories.data to be empty, got %#v", payload.Data.Relationships.AvailableTerritories.Data)
		}
		return webSubscriptionsJSONResponse(`{
			"data": {
				"id": "plan-1",
				"type": "subscriptionPlanAvailabilities",
				"attributes": {
					"availableInNewTerritories": false,
					"planType": "UPFRONT"
				}
			}
		}`), nil
	default:
		t.Fatalf("unexpected request: %s %s", req.Method, req.URL.String())
		return nil, nil
	}
}

func webSubscriptionsJSONResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
