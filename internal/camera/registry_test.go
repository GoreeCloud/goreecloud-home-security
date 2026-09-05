package camera

import (
	"strings"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-home-security/internal/config"
)

func TestRegistrySnapshotNeverContainsStreamOrSecretReferences(t *testing.T) {
	r := NewRegistry([]config.Camera{{ID: "front-door", Name: "Front Door", StreamURL: "rtsp://camera.local/live", UsernameEnv: "CAMERA_USER", PasswordEnv: "CAMERA_PASS", Enabled: true}})
	snapshot := r.Snapshot()
	if len(snapshot) != 1 || snapshot[0].ID != "front-door" {
		t.Fatalf("snapshot = %#v", snapshot)
	}
}

func TestMediaStatusRejectsRawErrorText(t *testing.T) {
	now := time.Now()
	status := MediaStatus{State: MediaUnavailable, LastProbeAt: &now, Reason: "dial tcp 10.0.0.5:554: refused"}
	if err := status.Validate(); err == nil || !strings.Contains(err.Error(), "categorical") {
		t.Fatalf("expected safe reason validation, got %v", err)
	}
}
