package publish

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
)

type stubPublishBetaGroupsClient struct {
	responses []*asc.BetaGroupsResponse
	err       error
	calls     int
}

func (s *stubPublishBetaGroupsClient) GetBetaGroups(ctx context.Context, appID string, opts ...asc.BetaGroupsOption) (*asc.BetaGroupsResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.calls >= len(s.responses) {
		return &asc.BetaGroupsResponse{}, nil
	}
	resp := s.responses[s.calls]
	s.calls++
	return resp, nil
}

func TestResolvePublishBetaGroupIDsFromList_ResolvesByNameAndID(t *testing.T) {
	groups := &asc.BetaGroupsResponse{
		Data: []asc.Resource[asc.BetaGroupAttributes]{
			{ID: "GROUP_A", Attributes: asc.BetaGroupAttributes{Name: "External Testers"}},
			{ID: "GROUP_B", Attributes: asc.BetaGroupAttributes{Name: "Internal Team"}},
		},
	}

	got, err := resolvePublishBetaGroupIDsFromList(
		[]string{" external testers ", "GROUP_B", "GROUP_A", "EXTERNAL TESTERS"},
		groups,
	)
	if err != nil {
		t.Fatalf("resolvePublishBetaGroupIDsFromList() error = %v", err)
	}

	want := []string{"GROUP_A", "GROUP_B"}
	if len(got) != len(want) {
		t.Fatalf("expected %d groups, got %d (%v)", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected group %d to be %q, got %q", i, want[i], got[i])
		}
	}
}

func TestResolvePublishBetaGroupIDsFromList_MissingGroup(t *testing.T) {
	groups := &asc.BetaGroupsResponse{
		Data: []asc.Resource[asc.BetaGroupAttributes]{
			{ID: "GROUP_A", Attributes: asc.BetaGroupAttributes{Name: "External Testers"}},
		},
	}

	_, err := resolvePublishBetaGroupIDsFromList([]string{"does-not-exist"}, groups)
	if err == nil {
		t.Fatal("expected error for missing beta group")
	}
	if !strings.Contains(err.Error(), `beta group "does-not-exist" not found`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolvePublishBetaGroupIDsFromList_AmbiguousName(t *testing.T) {
	groups := &asc.BetaGroupsResponse{
		Data: []asc.Resource[asc.BetaGroupAttributes]{
			{ID: "GROUP_A", Attributes: asc.BetaGroupAttributes{Name: "QA"}},
			{ID: "GROUP_B", Attributes: asc.BetaGroupAttributes{Name: "QA"}},
		},
	}

	_, err := resolvePublishBetaGroupIDsFromList([]string{"qa"}, groups)
	if err == nil {
		t.Fatal("expected error for ambiguous beta group name")
	}
	if !strings.Contains(err.Error(), `multiple beta groups named "qa"; use group ID`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolvePublishBetaGroupIDsFromList_NoGroupsReturned(t *testing.T) {
	_, err := resolvePublishBetaGroupIDsFromList([]string{"group"}, nil)
	if err == nil {
		t.Fatal("expected error when no group list is available")
	}
	if !strings.Contains(err.Error(), "no beta groups returned for app") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListAllPublishBetaGroups_Paginates(t *testing.T) {
	client := &stubPublishBetaGroupsClient{
		responses: []*asc.BetaGroupsResponse{
			{
				Data: []asc.Resource[asc.BetaGroupAttributes]{
					{ID: "GROUP_A", Attributes: asc.BetaGroupAttributes{Name: "Alpha"}},
				},
				Links: asc.Links{
					Next: "https://api.appstoreconnect.apple.com/v1/apps/APP_ID/betaGroups?cursor=NEXT",
				},
			},
			{
				Data: []asc.Resource[asc.BetaGroupAttributes]{
					{ID: "GROUP_B", Attributes: asc.BetaGroupAttributes{Name: "Beta"}},
				},
			},
		},
	}

	got, err := listAllPublishBetaGroups(context.Background(), client, "APP_ID")
	if err != nil {
		t.Fatalf("listAllPublishBetaGroups() error = %v", err)
	}
	if client.calls != 2 {
		t.Fatalf("expected 2 GetBetaGroups calls, got %d", client.calls)
	}
	if len(got.Data) != 2 {
		t.Fatalf("expected 2 beta groups, got %d", len(got.Data))
	}
	if got.Data[0].ID != "GROUP_A" || got.Data[1].ID != "GROUP_B" {
		t.Fatalf("unexpected group IDs: %#v", got.Data)
	}
}

func TestListAllPublishBetaGroups_FirstPageError(t *testing.T) {
	client := &stubPublishBetaGroupsClient{err: errors.New("boom")}
	_, err := listAllPublishBetaGroups(context.Background(), client, "APP_ID")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("unexpected error: %v", err)
	}
}
