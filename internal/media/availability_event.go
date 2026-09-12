package media

import (
	"errors"
	"regexp"
	"time"

	"github.com/GoreeCloud/goreecloud-home-security/internal/camera"
)

const AvailabilityEventSchemaVersion = 1

var (
	availabilityCameraIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,127}$`)
	availabilityReasonPattern   = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)
)

// AvailabilityEventCandidate is a privacy-minimized durable-event candidate for
// ordinary camera connectivity. It deliberately has no tamper, intrusion,
// security-outcome, credential, stream URL, codec, or image metadata fields.
type AvailabilityEventCandidate struct {
	SchemaVersion int                    `json:"schema_version"`
	EventType     AvailabilityTransition `json:"event_type"`
	CameraID      string                 `json:"camera_id"`
	OccurredAt    time.Time              `json:"occurred_at"`
	Reason        string                 `json:"reason,omitempty"`
}

// BuildAvailabilityEventCandidate converts a classified online/offline
// transition into a sanitized event candidate. It does not persist the event.
// Startup ambiguity, policy/authentication blocking, disabled state, and other
// non-connectivity transitions return changed=false and cannot become a tamper
// or security event through this boundary.
func BuildAvailabilityEventCandidate(
	cameraID string,
	previous camera.MediaStatus,
	current camera.MediaStatus,
	observedAt time.Time,
) (AvailabilityEventCandidate, bool, error) {
	if !availabilityCameraIDPattern.MatchString(cameraID) {
		return AvailabilityEventCandidate{}, false, errors.New("camera id is invalid")
	}
	if observedAt.IsZero() {
		return AvailabilityEventCandidate{}, false, errors.New("availability event observation time is required")
	}

	transition, changed, err := ClassifyAvailabilityTransition(previous, current)
	if err != nil {
		return AvailabilityEventCandidate{}, false, err
	}
	if !changed {
		return AvailabilityEventCandidate{}, false, nil
	}

	return AvailabilityEventCandidate{
		SchemaVersion: AvailabilityEventSchemaVersion,
		EventType:     transition,
		CameraID:      cameraID,
		OccurredAt:    observedAt.UTC(),
		Reason:        sanitizedAvailabilityReason(current.Reason),
	}, true, nil
}

// sanitizedAvailabilityReason only carries closed-shape categorical diagnostics
// into the future durable-event path. Free-form adapter/runtime diagnostics may
// contain endpoints, paths, credentials, or other private implementation detail
// and are therefore omitted rather than copied into the journal candidate.
func sanitizedAvailabilityReason(reason string) string {
	if availabilityReasonPattern.MatchString(reason) {
		return reason
	}
	return ""
}
