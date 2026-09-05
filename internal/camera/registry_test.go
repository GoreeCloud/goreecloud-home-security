package camera

import (
	"testing"

	"github.com/GoreeCloud/goreecloud-home-security/internal/config"
)

func TestSnapshotNeverContainsStreamOrSecretReferences(t *testing.T) {
	r := NewRegistry([]config.Camera{{
		ID: "front-door", Name: "Front Door", StreamURL: "rtsp://camera.local/live",
		UsernameEnv: "CAMERA_USER", PasswordEnv: "CAMERA_PASS", Enabled: true,
	}})
	got := r.Snapshot()
	if len(got) != 1 || got[0].ID != "front-door" || got[0].Name != "Front Door" || !got[0].Enabled {
		t.Fatalf("unexpected snapshot: %#v", got)
	}
}
