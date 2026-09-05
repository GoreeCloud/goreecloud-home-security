package camera

import (
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-home-security/internal/config"
)

func TestSnapshotNeverContainsStreamOrSecretReferences(t *testing.T) {
	r := NewRegistry([]config.Camera{{ID: "front-door", Name: "Front Door", StreamURL: "rtsp://camera.local/live", UsernameEnv: "CAMERA_USER", PasswordEnv: "CAMERA_PASS", Enabled: true}})
	got := r.Snapshot()
	if len(got) != 1 || got[0].ID != "front-door" || got[0].Name != "Front Door" || !got[0].Enabled {
		t.Fatalf("unexpected snapshot: %#v", got)
	}
	if got[0].Session.State != SessionIdle {
		t.Fatalf("session = %#v", got[0].Session)
	}
}

func TestSessionStatusIsSanitizedAndSummarized(t *testing.T) {
	r := NewRegistry([]config.Camera{{ID: "garage", Name: "Garage", StreamURL: "rtsp://camera.local/live", Enabled: true}})
	now := time.Unix(100, 0).UTC()
	if err := r.UpdateSession("garage", SessionStatus{State: SessionBackoff, Attempt: 2, LastExitAt: &now, Reason: "session_failed"}); err != nil {
		t.Fatal(err)
	}
	if got := r.Summary(); got.SessionsBackoff != 1 {
		t.Fatalf("summary = %#v", got)
	}
}
