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

func TestPruneBeforeRemovesOnlyExpiredEventsAndPreservesPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.jsonl")
	j, err := NewJournal(path)
	if err != nil {
		t.Fatal(err)
	}
	cutoff := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)
	endedAt := cutoff.Add(time.Hour)
	events := []Event{
		{ID: "expired", CameraID: "front-door", Type: "camera", StartedAt: cutoff.Add(-time.Hour)},
		{ID: "boundary", CameraID: "front-door", Type: "camera", StartedAt: cutoff},
		{ID: "ended-later", CameraID: "front-door", Type: "motion", StartedAt: cutoff.Add(-2 * time.Hour), EndedAt: &endedAt},
	}
	for _, event := range events {
		if err := j.Append(event); err != nil {
			t.Fatal(err)
		}
	}
	removed, err := j.PruneBefore(cutoff)
	if err != nil {
		t.Fatal(err)
	}
	if removed != 1 {
		t.Fatalf("removed = %d, want 1", removed)
	}
	got, err := j.List(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].ID != "boundary" || got[1].ID != "ended-later" {
		t.Fatalf("unexpected retained events: %#v", got)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("retained journal mode = %o, want 600", info.Mode().Perm())
	}
}

func TestPruneBeforeFailsClosedWithoutReplacingMalformedJournal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.jsonl")
	j, err := NewJournal(path)
	if err != nil {
		t.Fatal(err)
	}
	original := []byte("not-json\n")
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := j.PruneBefore(time.Now().UTC()); err == nil {
		t.Fatal("expected malformed journal retention failure")
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Fatalf("malformed journal changed during failed retention: %q", got)
	}
}
