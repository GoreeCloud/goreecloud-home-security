package events

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestJournalPersistsAndLimitsEvents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.jsonl")
	j, err := NewJournal(path)
	if err != nil {
		t.Fatal(err)
	}
	for i, id := range []string{"one", "two", "three"} {
		if err := j.Append(Event{ID: id, CameraID: "front-door", Type: "motion", StartedAt: time.Unix(int64(i+1), 0).UTC()}); err != nil {
			t.Fatal(err)
		}
	}
	got, err := j.List(2)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].ID != "two" || got[1].ID != "three" {
		t.Fatalf("unexpected events: %#v", got)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("journal permissions too broad: %o", info.Mode().Perm())
	}
}

func TestJournalFailsClosedOnMalformedData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.jsonl")
	j, err := NewJournal(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("not-json\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := j.List(10); err == nil {
		t.Fatal("expected malformed journal error")
	}
}
