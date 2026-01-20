//go:build integration

package asc

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestIntegrationEndpoints(t *testing.T) {
	keyID := os.Getenv("ASC_KEY_ID")
	issuerID := os.Getenv("ASC_ISSUER_ID")
	keyPath := os.Getenv("ASC_PRIVATE_KEY_PATH")
	appID := os.Getenv("ASC_APP_ID")

	if keyID == "" || issuerID == "" || keyPath == "" || appID == "" {
		t.Skip("integration tests require ASC_KEY_ID, ASC_ISSUER_ID, ASC_PRIVATE_KEY_PATH, ASC_APP_ID")
	}

	client, err := NewClient(keyID, issuerID, keyPath)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	feedback, err := client.GetFeedback(ctx, appID, WithFeedbackLimit(1))
	if err != nil {
		t.Fatalf("failed to fetch feedback: %v", err)
	}
	if feedback == nil {
		t.Fatal("expected feedback response")
	}
	if len(feedback.Data) > 0 && feedback.Data[0].Type == "" {
		t.Fatal("expected feedback data type to be set")
	}

	crashes, err := client.GetCrashes(ctx, appID, WithCrashLimit(1))
	if err != nil {
		t.Fatalf("failed to fetch crashes: %v", err)
	}
	if crashes == nil {
		t.Fatal("expected crashes response")
	}
	if len(crashes.Data) > 0 && crashes.Data[0].Type == "" {
		t.Fatal("expected crash data type to be set")
	}

	reviews, err := client.GetReviews(ctx, appID, WithLimit(1))
	if err != nil {
		t.Fatalf("failed to fetch reviews: %v", err)
	}
	if reviews == nil {
		t.Fatal("expected reviews response")
	}
	if len(reviews.Data) > 0 && reviews.Data[0].Type == "" {
		t.Fatal("expected review data type to be set")
	}
}
