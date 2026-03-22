package appleauth

import (
	"encoding/json"
	"testing"
)

func TestAuthOptionsResponseAuthOptionsCopiesTrustedPhoneNumbers(t *testing.T) {
	var resp AuthOptionsResponse
	if err := json.Unmarshal([]byte(`{
		"noTrustedDevices": true,
		"trustedPhoneNumbers": [
			{"id": 7, "pushMode": "sms", "numberWithDialCode": "+1 (•••) •••-••66"}
		]
	}`), &resp); err != nil {
		t.Fatalf("unmarshal auth options response: %v", err)
	}

	got := resp.AuthOptions()
	if got == nil {
		t.Fatal("expected auth options response conversion result")
	}
	if !got.NoTrustedDevices {
		t.Fatal("expected noTrustedDevices to be preserved")
	}
	if len(got.TrustedPhoneNumbers) != 1 {
		t.Fatalf("expected one trusted phone number, got %d", len(got.TrustedPhoneNumbers))
	}
	if got.TrustedPhoneNumbers[0].ID != 7 {
		t.Fatalf("expected trusted phone id 7, got %d", got.TrustedPhoneNumbers[0].ID)
	}

	got.TrustedPhoneNumbers[0].ID = 99
	if resp.TrustedPhoneNumbers[0].ID != 7 {
		t.Fatalf("expected trusted phone slice to be copied, got %d", resp.TrustedPhoneNumbers[0].ID)
	}
}

func TestNilAuthOptionsResponseReturnsEmptyOptions(t *testing.T) {
	var resp *AuthOptionsResponse
	got := resp.AuthOptions()
	if got == nil {
		t.Fatal("expected empty auth options for nil response")
	}
	if got.NoTrustedDevices {
		t.Fatal("did not expect noTrustedDevices on nil response")
	}
	if len(got.TrustedPhoneNumbers) != 0 {
		t.Fatalf("expected no trusted phone numbers, got %d", len(got.TrustedPhoneNumbers))
	}
}
