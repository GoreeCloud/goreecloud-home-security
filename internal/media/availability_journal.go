package media

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/GoreeCloud/goreecloud-home-security/internal/events"
)

const (
	DefaultAvailabilityDebounceWindow = 30 * time.Second
	maxAvailabilityDebounceWindow     = 10 * time.Minute
)

// PersistAvailabilityEventCandidate maps one sanitized connectivity candidate
// into the existing durable JSONL journal. It intentionally persists no media
// reason, endpoint, credential, codec, recording, image, or tamper metadata.
//
// Deterministic event IDs make exact replay idempotent across process restarts.
// Repeated same-state candidates are suppressed, and opposite-state flapping
// inside the bounded debounce window is not persisted. Out-of-order observations
// fail closed rather than rewriting journal history.
func PersistAvailabilityEventCandidate(
	journal *events.Journal,
	candidate AvailabilityEventCandidate,
	debounceWindow time.Duration,
) (bool, error) {
	if journal == nil {
		return false, errors.New("availability event journal is required")
	}
	if debounceWindow < 0 || debounceWindow > maxAvailabilityDebounceWindow {
		return false, errors.New("availability debounce window must be between 0 and 10 minutes")
	}
	if candidate.SchemaVersion != AvailabilityEventSchemaVersion {
		return false, fmt.Errorf("unsupported availability event schema version %d", candidate.SchemaVersion)
	}
	if !availabilityCameraIDPattern.MatchString(candidate.CameraID) {
		return false, errors.New("camera id is invalid")
	}
	if candidate.OccurredAt.IsZero() {
		return false, errors.New("availability event observation time is required")
	}
	if candidate.Reason != "" && !availabilityReasonPattern.MatchString(candidate.Reason) {
		return false, errors.New("availability event reason is not categorical")
	}
	switch candidate.EventType {
	case AvailabilityOffline, AvailabilityRecovered:
	default:
		return false, fmt.Errorf("unsupported availability event type %q", candidate.EventType)
	}

	occurredAt := candidate.OccurredAt.UTC()
	last, err := journal.LastForCamera(candidate.CameraID, "camera")
	if err != nil {
		return false, err
	}
	if last != nil {
		if occurredAt.Before(last.StartedAt) {
			return false, errors.New("availability event observation precedes durable camera event")
		}
		if last.Label == string(candidate.EventType) {
			return false, nil
		}
		if occurredAt.Sub(last.StartedAt) < debounceWindow {
			return false, nil
		}
	}

	event := events.Event{
		ID:        availabilityEventID(candidate.CameraID, candidate.EventType, occurredAt),
		CameraID:  candidate.CameraID,
		Type:      "camera",
		Label:     string(candidate.EventType),
		StartedAt: occurredAt,
	}
	return journal.AppendUnique(event)
}

func availabilityEventID(cameraID string, eventType AvailabilityTransition, occurredAt time.Time) string {
	payload := fmt.Sprintf(
		"availability-v%d\x00%s\x00%s\x00%s",
		AvailabilityEventSchemaVersion,
		cameraID,
		eventType,
		occurredAt.UTC().Format(time.RFC3339Nano),
	)
	digest := sha256.Sum256([]byte(payload))
	return fmt.Sprintf("camera-availability-%x", digest[:])
}
