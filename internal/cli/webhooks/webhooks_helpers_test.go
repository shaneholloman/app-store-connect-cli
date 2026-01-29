package webhooks

import "testing"

func TestNormalizeWebhookEvents(t *testing.T) {
	values, err := normalizeWebhookEvents("build_upload_state_updated, build_beta_detail_external_build_state_updated")
	if err != nil {
		t.Fatalf("normalizeWebhookEvents() error: %v", err)
	}
	if len(values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(values))
	}
	if string(values[0]) != "BUILD_UPLOAD_STATE_UPDATED" {
		t.Fatalf("expected normalized event, got %q", values[0])
	}
}
