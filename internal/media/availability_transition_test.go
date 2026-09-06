package media

import (
	"testing"

	"github.com/GoreeCloud/goreecloud-home-security/internal/camera"
)

func TestClassifyAvailabilityTransitionOfflineAndRecovery(t *testing.T) {
	offline, emit, err := ClassifyAvailabilityTransition(
		camera.MediaStatus{State: camera.MediaOnline},
		camera.MediaStatus{State: camera.MediaUnavailable, Reason: "connection_failed"},
	)
	if err != nil {
		t.Fatalf("classify offline transition: %v", err)
	}
	if !emit || offline != AvailabilityOffline {
		t.Fatalf("expected offline transition, got %q emit=%v", offline, emit)
	}

	recovered, emit, err := ClassifyAvailabilityTransition(
		camera.MediaStatus{State: camera.MediaUnavailable, Reason: "connection_failed"},
		camera.MediaStatus{State: camera.MediaOnline},
	)
	if err != nil {
		t.Fatalf("classify recovery transition: %v", err)
	}
	if !emit || recovered != AvailabilityRecovered {
		t.Fatalf("expected recovery transition, got %q emit=%v", recovered, emit)
	}
}

func TestClassifyAvailabilityTransitionDoesNotFabricateStartupOrTamperEvents(t *testing.T) {
	cases := []struct {
		name     string
		previous camera.MediaStatus
		current  camera.MediaStatus
	}{
		{
			name:     "startup unavailable",
			previous: camera.MediaStatus{State: camera.MediaUnknown},
			current:  camera.MediaStatus{State: camera.MediaUnavailable, Reason: "connection_failed"},
		},
		{
			name:     "probing unavailable",
			previous: camera.MediaStatus{State: camera.MediaProbing},
			current:  camera.MediaStatus{State: camera.MediaUnavailable, Reason: "connection_failed"},
		},
		{
			name:     "authentication blocked",
			previous: camera.MediaStatus{State: camera.MediaOnline},
			current:  camera.MediaStatus{State: camera.MediaBlocked, Reason: "credentialed_probe_blocked"},
		},
		{
			name:     "disabled",
			previous: camera.MediaStatus{State: camera.MediaDisabled},
			current:  camera.MediaStatus{State: camera.MediaDisabled},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			transition, emit, err := ClassifyAvailabilityTransition(tc.previous, tc.current)
			if err != nil {
				t.Fatalf("classify transition: %v", err)
			}
			if emit || transition != "" {
				t.Fatalf("unexpected availability event %q emit=%v", transition, emit)
			}
		})
	}
}

func TestClassifyAvailabilityTransitionFailsClosedOnInvalidStatus(t *testing.T) {
	_, _, err := ClassifyAvailabilityTransition(
		camera.MediaStatus{State: camera.MediaState("invalid")},
		camera.MediaStatus{State: camera.MediaOnline},
	)
	if err == nil {
		t.Fatal("expected invalid previous status to fail closed")
	}
}
