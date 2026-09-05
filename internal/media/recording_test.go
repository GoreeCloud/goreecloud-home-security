package media

import (
	"strings"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-home-security/internal/config"
)

func TestRecordingPlanUsesArgumentVectorAndCameraScopedPath(t *testing.T) {
	plan, err := BuildRecordingPlan("ffmpeg", "/srv/home-security", config.Camera{ID: "front-door", Name: "Front Door", StreamURL: "rtsp://camera.local/live?profile=main", Enabled: true}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Executable != "ffmpeg" || !strings.Contains(plan.SegmentPattern, "front-door") {
		t.Fatalf("plan = %#v", plan)
	}
	joined := strings.Join(plan.Args, " ")
	if !strings.Contains(joined, "-nostdin") || !strings.Contains(joined, "-c copy") {
		t.Fatalf("args = %#v", plan.Args)
	}
}

func TestRecordingPlanBlocksCredentialedCamera(t *testing.T) {
	_, err := BuildRecordingPlan("ffmpeg", t.TempDir(), config.Camera{ID: "front-door", Name: "Front Door", StreamURL: "rtsp://camera.local/live", UsernameEnv: "CAMERA_USER", PasswordEnv: "CAMERA_PASS", Enabled: true}, time.Minute)
	if ErrorCode(err) != ReasonCredentialedProbeBlocked {
		t.Fatalf("error = %v", err)
	}
}
