package media

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-home-security/internal/camera"
)

func TestBuildAvailabilityEventCandidateOfflineIsSanitized(t *testing.T) {
	observed := time.Date(2026, 9, 6, 14, 30, 0, 0, time.FixedZone("CDT", -5*60*60))
	event, changed, err := BuildAvailabilityEventCandidate(
		"front-door",
		camera.MediaStatus{State: camera.MediaOnline},
		camera.MediaStatus{State: camera.MediaUnavailable, Reason: "session_stalled"},
		observed,
	)
	if err != nil {
		t.Fatalf("build event: %v", err)
	}
	if !changed {
		t.Fatal("expected online to unavailable transition")
	}
	if event.EventType != AvailabilityOffline {
		t.Fatalf("unexpected event type: %q", event.EventType)
	}
	if event.SchemaVersion != AvailabilityEventSchemaVersion {
		t.Fatalf("unexpected schema version: %d", event.SchemaVersion)
	}
	if event.CameraID != "front-door" || event.Reason != "session_stalled" {
		t.Fatalf("unexpected event: %#v", event)
	}
	if !event.OccurredAt.Equal(observed) || event.OccurredAt.Location() != time.UTC {
		t.Fatalf("observation time was not normalized to UTC: %v", event.OccurredAt)
	}

	encoded, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}
	payload := string(encoded)
	for _, forbidden := range []string{"tamper", "credential", "stream_url", "video_codec", "audio_codec"} {
		if strings.Contains(payload, forbidden) {
			t.Fatalf("event payload contains forbidden field %q: %s", forbidden, payload)
		}
	}
}

func TestSanitizedAvailabilityReasonDropsFreeFormDiagnostic(t *testing.T) {
	privateDiagnostic := "dial rtsp://operator:secret@192.0.2.10/live failed"
	if got := sanitizedAvailabilityReason(privateDiagnostic); got != "" {
		t.Fatalf("free-form diagnostic was not dropped: %q", got)
	}
	if got := sanitizedAvailabilityReason("session_stalled"); got != "session_stalled" {
		t.Fatalf("bounded categorical reason was not preserved: %q", got)
	}
}

func TestBuildAvailabilityEventCandidateRejectsFreeFormReasonBeforeEmission(t *testing.T) {
	event, changed, err := BuildAvailabilityEventCandidate(
		"front-door",
		camera.MediaStatus{State: camera.MediaOnline},
		camera.MediaStatus{
			State:  camera.MediaUnavailable,
			Reason: "dial rtsp://operator:secret@192.0.2.10/live failed",
		},
		time.Date(2026, 9, 6, 20, 0, 0, 0, time.UTC),
	)
	if err == nil {
		t.Fatal("expected free-form media diagnostic to fail closed")
	}
	if changed || event != (AvailabilityEventCandidate{}) {
		t.Fatalf("invalid diagnostic produced an event candidate: changed=%v event=%#v", changed, event)
	}
}

func TestBuildAvailabilityEventCandidateRecovery(t *testing.T) {
	event, changed, err := BuildAvailabilityEventCandidate(
		"garage",
		camera.MediaStatus{State: camera.MediaUnavailable, Reason: "network_unreachable"},
		camera.MediaStatus{State: camera.MediaOnline},
		time.Date(2026, 9, 6, 20, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("build event: %v", err)
	}
	if !changed || event.EventType != AvailabilityRecovered {
		t.Fatalf("unexpected recovery result: changed=%v event=%#v", changed, event)
	}
}

func TestBuildAvailabilityEventCandidateDoesNotPromoteBlockedOrStartupState(t *testing.T) {
	cases := []struct {
		name     string
		previous camera.MediaStatus
		current  camera.MediaStatus
	}{
		{"startup", camera.MediaStatus{State: camera.MediaUnknown}, camera.MediaStatus{State: camera.MediaUnavailable}},
		{"policy-block", camera.MediaStatus{State: camera.MediaOnline}, camera.MediaStatus{State: camera.MediaBlocked, Reason: "authentication_failed"}},
		{"probing", camera.MediaStatus{State: camera.MediaProbing}, camera.MediaStatus{State: camera.MediaOnline}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, changed, err := BuildAvailabilityEventCandidate(
				"camera-1",
				tc.previous,
				tc.current,
				time.Now().UTC(),
			)
			if err != nil {
				t.Fatalf("build event: %v", err)
			}
			if changed {
				t.Fatal("unexpected connectivity event candidate")
			}
		})
	}
}

func TestBuildAvailabilityEventCandidateRejectsInvalidIdentityAndTime(t *testing.T) {
	previous := camera.MediaStatus{State: camera.MediaOnline}
	current := camera.MediaStatus{State: camera.MediaUnavailable}
	if _, _, err := BuildAvailabilityEventCandidate("../camera", previous, current, time.Now().UTC()); err == nil {
		t.Fatal("expected invalid camera id rejection")
	}
	if _, _, err := BuildAvailabilityEventCandidate("camera-1", previous, current, time.Time{}); err == nil {
		t.Fatal("expected zero observation time rejection")
	}
}
