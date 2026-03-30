package threads

import (
	"testing"
	"time"
)

func TestSaveThreadRoundTrip(t *testing.T) {
	store := NewStore(t.TempDir())
	now := time.Now().UTC()
	thread := Thread{
		ID:        "thread-1",
		Title:     "Release Prep",
		CreatedAt: now,
		UpdatedAt: now,
		Messages: []Message{
			{ID: "msg-1", Role: RoleUser, Kind: KindMessage, Content: "Validate 2.3.0", CreatedAt: now},
		},
	}

	if err := store.SaveThread(thread); err != nil {
		t.Fatalf("SaveThread() error = %v", err)
	}

	got, err := store.Get("thread-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Title != thread.Title {
		t.Fatalf("Title = %q, want %q", got.Title, thread.Title)
	}
	if len(got.Messages) != 1 {
		t.Fatalf("len(Messages) = %d, want 1", len(got.Messages))
	}
}
