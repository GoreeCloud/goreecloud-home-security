package media

import (
	"fmt"

	"github.com/GoreeCloud/goreecloud-home-security/internal/camera"
)

// AvailabilityTransition is a sanitized camera-connectivity transition. These
// transitions are deliberately distinct from tamper/security events: ordinary
// media loss or recovery must never be promoted to a tamper claim.
type AvailabilityTransition string

const (
	AvailabilityOffline   AvailabilityTransition = "camera_offline"
	AvailabilityRecovered AvailabilityTransition = "camera_recovered"
)

// ClassifyAvailabilityTransition returns a durable-event candidate only for a
// known online -> unavailable loss or unavailable -> online recovery. Startup
// unknown/probing observations, disabled cameras, and policy/authentication
// blocking do not fabricate availability transitions.
func ClassifyAvailabilityTransition(
	previous camera.MediaStatus,
	current camera.MediaStatus,
) (AvailabilityTransition, bool, error) {
	if err := previous.Validate(); err != nil {
		return "", false, fmt.Errorf("validate previous media status: %w", err)
	}
	if err := current.Validate(); err != nil {
		return "", false, fmt.Errorf("validate current media status: %w", err)
	}

	switch {
	case previous.State == camera.MediaOnline && current.State == camera.MediaUnavailable:
		return AvailabilityOffline, true, nil
	case previous.State == camera.MediaUnavailable && current.State == camera.MediaOnline:
		return AvailabilityRecovered, true, nil
	default:
		return "", false, nil
	}
}
