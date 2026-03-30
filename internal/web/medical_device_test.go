package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSetMedicalDeviceDeclarationRejectsTrue(t *testing.T) {
	client := testWebClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))

	_, err := client.SetMedicalDeviceDeclaration(context.Background(), "account-123", "app-123", true)
	if err == nil || !strings.Contains(err.Error(), "only false is currently supported") {
		t.Fatalf("expected false-only error, got %v", err)
	}
}

func TestSetMedicalDeviceDeclarationPostsExpectedRequest(t *testing.T) {
	requirementsCalls := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/ppm/complianceform/v1/accounts/account-123/requirements":
			requirementsCalls++
			if got := r.URL.Query().Get("contentId"); got != "app-123" {
				t.Fatalf("expected contentId app-123, got %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			if requirementsCalls == 1 {
				_, _ = w.Write([]byte(`{
					"accountId":"account-123",
					"requirementData":[{
						"contentId":"app-123",
						"requirements":[{
							"id":"req-123",
							"name":"MEDICAL_DEVICE",
							"status":"PENDING_COLLECTION"
						}]
					}]
				}`))
				return
			}
			_, _ = w.Write([]byte(`{
				"accountId":"account-123",
				"requirementData":[{
					"contentId":"app-123",
					"requirements":[{
						"id":"req-123",
						"name":"MEDICAL_DEVICE",
						"status":"COLLECTED",
						"formId":"form-123"
					}]
				}]
			}`))
		case r.Method == http.MethodGet && r.URL.Path == "/ppm/complianceform/v1/accounts/account-123/requirements/req-123/forms":
			if got := r.URL.Query().Get("contentId"); got != "app-123" {
				t.Fatalf("expected contentId app-123, got %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"data":{
					"accountId":"account-123",
					"contentId":"app-123",
					"requirementId":"req-123",
					"requirementName":"MEDICAL_DEVICE",
					"medicalDeviceData":{}
				},
				"constraints":{
					"$[*].countriesOrRegions":{
						"attributeName":"countriesOrRegions",
						"options":[
							{"value":"USA"},
							{"value":"GBR"},
							{"value":"EU"}
						]
					},
					"$[*].medicalDeviceData.contactInformation[0].countriesOrRegions":{
						"attributeName":"countriesOrRegions",
						"options":[
							{"listValues":["USA","GBR","EEA"]}
						]
					}
				}
			}`))
		case r.Method == http.MethodPost && r.URL.Path == "/ppm/complianceform/v1/accounts/account-123/contents/app-123/requirements/req-123/forms":
			var body struct {
				AccountID          string   `json:"accountId"`
				ContentID          string   `json:"contentId"`
				RequirementID      string   `json:"requirementId"`
				RequirementName    string   `json:"requirementName"`
				CountriesOrRegions []string `json:"countriesOrRegions"`
				MedicalDeviceData  struct {
					Declaration string `json:"declaration"`
				} `json:"medicalDeviceData"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.AccountID != "account-123" || body.ContentID != "app-123" || body.RequirementID != "req-123" {
				t.Fatalf("unexpected identifiers in body: %#v", body)
			}
			if body.RequirementName != "MEDICAL_DEVICE" {
				t.Fatalf("expected requirement name MEDICAL_DEVICE, got %q", body.RequirementName)
			}
			if body.MedicalDeviceData.Declaration != "no" {
				t.Fatalf("expected declaration no, got %q", body.MedicalDeviceData.Declaration)
			}
			if got := strings.Join(body.CountriesOrRegions, ","); got != "EEA,GBR,USA" {
				t.Fatalf("expected normalized countries EEA,GBR,USA, got %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	client := testWebClient(server)
	got, err := client.SetMedicalDeviceDeclaration(context.Background(), "account-123", "app-123", false)
	if err != nil {
		t.Fatalf("SetMedicalDeviceDeclaration() error = %v", err)
	}
	if got == nil {
		t.Fatal("expected result")
	}
	if got.AppID != "app-123" {
		t.Fatalf("expected app id app-123, got %q", got.AppID)
	}
	if got.RequirementID != "req-123" || got.RequirementName != "MEDICAL_DEVICE" {
		t.Fatalf("unexpected requirement metadata: %#v", got)
	}
	if got.Status != "COLLECTED" {
		t.Fatalf("expected collected status, got %q", got.Status)
	}
	if got.Declared {
		t.Fatalf("expected declared false, got true")
	}
	if got := strings.Join(got.CountriesOrRegions, ","); got != "EEA,GBR,USA" {
		t.Fatalf("expected countries EEA,GBR,USA, got %q", got)
	}
}

func TestSetMedicalDeviceDeclarationPrefersExactContentIDRequirements(t *testing.T) {
	requirementsCalls := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/ppm/complianceform/v1/accounts/account-123/requirements":
			requirementsCalls++
			w.Header().Set("Content-Type", "application/json")
			if requirementsCalls == 1 {
				_, _ = w.Write([]byte(`{
					"accountId":"account-123",
					"requirementData":[
						{
							"contentId":"",
							"requirements":[{
								"id":"req-generic",
								"name":"OTHER_REQUIREMENT",
								"status":"PENDING_COLLECTION"
							}]
						},
						{
							"contentId":"app-123",
							"requirements":[{
								"id":"req-app",
								"name":"MEDICAL_DEVICE",
								"status":"PENDING_COLLECTION"
							}]
						}
					]
				}`))
				return
			}
			_, _ = w.Write([]byte(`{
				"accountId":"account-123",
				"requirementData":[
					{
						"contentId":"",
						"requirements":[{
							"id":"req-generic",
							"name":"OTHER_REQUIREMENT",
							"status":"PENDING_COLLECTION"
						}]
					},
					{
						"contentId":"app-123",
						"requirements":[{
							"id":"req-app",
							"name":"MEDICAL_DEVICE",
							"status":"COLLECTED",
							"formId":"form-app"
						}]
					}
				]
			}`))
		case r.Method == http.MethodGet && r.URL.Path == "/ppm/complianceform/v1/accounts/account-123/requirements/req-app/forms":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"constraints":{
					"$[*].countriesOrRegions":{
						"attributeName":"countriesOrRegions",
						"options":[
							{"value":"USA"}
						]
					}
				}
			}`))
		case r.Method == http.MethodPost && r.URL.Path == "/ppm/complianceform/v1/accounts/account-123/contents/app-123/requirements/req-app/forms":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	client := testWebClient(server)
	got, err := client.SetMedicalDeviceDeclaration(context.Background(), "account-123", "app-123", false)
	if err != nil {
		t.Fatalf("SetMedicalDeviceDeclaration() error = %v", err)
	}
	if got == nil {
		t.Fatal("expected result")
	}
	if got.RequirementID != "req-app" {
		t.Fatalf("expected exact app requirement id req-app, got %q", got.RequirementID)
	}
	if got.Status != "COLLECTED" {
		t.Fatalf("expected collected status, got %q", got.Status)
	}
}
