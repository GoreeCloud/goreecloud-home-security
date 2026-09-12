package media

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-home-security/internal/events"
)

func TestPersistAvailabilityEventCandidateIsRestartReplaySafeAndMinimized(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.jsonl")
	journal, err := events.NewJournal(path)
	if err != nil {
		t.Fatal(err)
	}
	candidate := AvailabilityEventCandidate{
		SchemaVersion: AvailabilityEventSchemaVersion,
		EventType:     AvailabilityOffline,
		CameraID:      "front-door",
		OccurredAt:    time.Date(2026, 9, 7, 20, 0, 0, 0, time.UTC),
		Reason:        "session_stalled",
	}

	appended, err := PersistAvailabilityEventCandidate(journal, candidate, DefaultAvailabilityDebounceWindow)
	if err != nil || !appended {
		t.Fatalf("first persist: appended=%v err=%v", appended, err)
	}

	reopened, err := events.NewJournal(path)
	if err != nil {
		t.Fatal(err)
	}
	appended, err = PersistAvailabilityEventCandidate(reopened, candidate, DefaultAvailabilityDebounceWindow)
	if err != nil {
		t.Fatal(err)
	}
	if appended {
		t.Fatal("exact replay must not append a second durable event")
	}

	got, err := reopened.List(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Type != "camera" || got[0].Label != string(AvailabilityOffline) {
		t.Fatalf("unexpected durable events: %#v", got)
	}
	encoded, err := json.Marshal(got[0])
	if err != nil {
		t.Fatal(err)
	}
	payload := string(encoded)
	for _, forbidden := range []string{"session_stalled", "reason", "credential", "stream_url", "codec", "tamper"} {
		if strings.Contains(payload, forbidden) {
			t.Fatalf("durable connectivity event leaks forbidden value %q: %s", forbidden, payload)
		}
	}
}

func TestPersistAvailabilityEventCandidateDebouncesOppositeFlap(t *testing.T) {
	journal, err := events.NewJournal(filepath.Join(t.TempDir(), "events.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	base := time.Date(2026, 9, 7, 20, 0, 0, 0, time.UTC)
	offline := AvailabilityEventCandidate{
		SchemaVersion: AvailabilityEventSchemaVersion,
		EventType:     AvailabilityOffline,
		CameraID:      "garage",
		OccurredAt:    base,
	}
	if appended, err := PersistAvailabilityEventCandidate(journal, offline, 30*time.Second); err != nil || !appended {
		t.Fatalf("offline persist: appended=%v err=%v", appended, err)
	}

	recovered := AvailabilityEventCandidate{
		SchemaVersion: AvailabilityEventSchemaVersion,
		EventType:     AvailabilityRecovered,
		CameraID:      "garage",
		OccurredAt:    base.Add(10 * time.Second),
	}
	if appended, err := PersistAvailabilityEventCandidate(journal, recovered, 30*time.Second); err != nil || appended {
		t.Fatalf("flap should be suppressed: appended=%v err=%v", appended, err)
	}

	recovered.OccurredAt = base.Add(31 * time.Second)
	if appended, err := PersistAvailabilityEventCandidate(journal, recovered, 30*time.Second); err != nil || !appended {
		t.Fatalf("settled recovery persist: appended=%v err=%v", appended, err)
	}
	got, err := journal.List(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[1].Label != string(AvailabilityRecovered) {
		t.Fatalf("unexpected durable events: %#v", got)
	}
}

func TestPersistAvailabilityEventCandidateFailsClosedOnOutOfOrderObservation(t *testing.T) {
	journal, err := events.NewJournal(filepath.Join(t.TempDir(), "events.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	base := time.Date(2026, 9, 7, 20, 0, 0, 0, time.UTC)
	first := AvailabilityEventCandidate{
		SchemaVersion: AvailabilityEventSchemaVersion,
		EventType:     AvailabilityOffline,
		CameraID:      "side-yard",
		OccurredAt:    base,
	}
	if _, err := PersistAvailabilityEventCandidate(journal, first, 0); err != nil {
		t.Fatal(err)
	}
	older := AvailabilityEventCandidate{
		SchemaVersion: AvailabilityEventSchemaVersion,
		EventType:     AvailabilityRecovered,
		CameraID:      "side-yard",
		OccurredAt:    base.Add(-time.Second),
	}
	if _, err := PersistAvailabilityEventCandidate(journal, older, 0); err == nil {
		t.Fatal("expected out-of-order observation rejection")
	}
}
