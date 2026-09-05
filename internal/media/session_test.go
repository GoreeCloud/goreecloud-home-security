package media

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-home-security/internal/config"
)

type fakeProcessRunner struct {
	calls      int
	executable string
	args       []string
	err        error
	started    bool
}

func (r *fakeProcessRunner) Run(_ context.Context, executable string, args []string, started func()) error {
	r.calls++
	r.executable = executable
	r.args = append([]string(nil), args...)
	if started != nil {
		started()
		r.started = true
	}
	return r.err
}

func TestSessionPlanUsesBoundedRWTimeoutAndNullSink(t *testing.T) {
	plan, err := BuildSessionPlan("ffmpeg", config.Camera{ID: "front-door", Name: "Front Door", StreamURL: "rtsp://camera.local/live", Enabled: true}, 15*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(plan.Args, " ")
	if !strings.Contains(joined, "-rw_timeout 15000000") || !strings.Contains(joined, "-f null -") || !strings.Contains(joined, "-c copy") {
		t.Fatalf("args = %#v", plan.Args)
	}
}
func TestSessionBlocksCredentialedCameraBeforeProcessExecution(t *testing.T) {
	runner := &fakeProcessRunner{}
	session := newFFmpegSessionWithRunner("ffmpeg", 15*time.Second, runner)
	err := session.Run(context.Background(), config.Camera{ID: "front-door", Name: "Front Door", StreamURL: "rtsp://camera.local/live", UsernameEnv: "CAMERA_USER", PasswordEnv: "CAMERA_PASS", Enabled: true}, nil)
	if ErrorCode(err) != ReasonCredentialedProbeBlocked {
		t.Fatalf("error = %v", err)
	}
	if runner.calls != 0 {
		t.Fatalf("runner calls = %d", runner.calls)
	}
}
func TestSessionCollapsesProcessFailureToCategoricalReason(t *testing.T) {
	runner := &fakeProcessRunner{err: errors.New("rtsp://private-camera/path connection failed")}
	session := newFFmpegSessionWithRunner("ffmpeg", 15*time.Second, runner)
	err := session.Run(context.Background(), config.Camera{ID: "front-door", Name: "Front Door", StreamURL: "rtsp://camera.local/live", Enabled: true}, func() {})
	if ErrorCode(err) != ReasonSessionFailed {
		t.Fatalf("error = %v", err)
	}
	if !runner.started {
		t.Fatal("expected start callback")
	}
}
