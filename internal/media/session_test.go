package media

import (
	"context"
	"errors"
	"os/exec"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-home-security/internal/config"
)

type fakeWorkerRunner struct {
	descriptor WorkerDescriptor
	executable string
	started    bool
	err        error
}

func (f *fakeWorkerRunner) Run(_ context.Context, executable string, descriptor WorkerDescriptor, started func()) error {
	f.executable = executable
	f.descriptor = descriptor
	if started != nil {
		started()
		f.started = true
	}
	return f.err
}

func TestWorkerSessionResolvesCredentialsIntoDescriptorOnly(t *testing.T) {
	values := map[string]string{"CAMERA_USER": "alice", "CAMERA_PASS": "top-secret"}
	runner := &fakeWorkerRunner{}
	s := newWorkerSessionWithRunner("worker-bin", func(name string) (string, bool) { v, ok := values[name]; return v, ok }, 15*time.Second, runner)
	err := s.Run(context.Background(), config.Camera{ID: "front-door", Name: "Front Door", StreamURL: "rtsp://camera.local/live", UsernameEnv: "CAMERA_USER", PasswordEnv: "CAMERA_PASS", Enabled: true}, func() {})
	if ErrorCode(err) != ReasonSessionExited {
		t.Fatalf("error=%v", err)
	}
	if runner.executable != "worker-bin" || runner.descriptor.Username != "alice" || runner.descriptor.Password != "top-secret" || !runner.started {
		t.Fatalf("runner=%#v", runner)
	}
}

func TestWorkerSessionMapsSanitizedExitReason(t *testing.T) {
	runner := &fakeWorkerRunner{err: &exec.ExitError{ProcessState: nil}}
	_ = runner
	if WorkerReasonForExitCode(WorkerExitAuthFailed) != ReasonRTSPAuthFailed {
		t.Fatal("exit mapping failed")
	}
	if WorkerReasonForExitCode(255) != ReasonSessionFailed {
		t.Fatal("unknown exit mapping failed")
	}
}

func TestWorkerSessionFailsClosedWhenCredentialMissing(t *testing.T) {
	runner := &fakeWorkerRunner{}
	s := newWorkerSessionWithRunner("worker-bin", func(string) (string, bool) { return "", false }, 15*time.Second, runner)
	err := s.Run(context.Background(), config.Camera{ID: "front-door", Name: "Front Door", StreamURL: "rtsp://camera.local/live", UsernameEnv: "CAMERA_USER", PasswordEnv: "CAMERA_PASS", Enabled: true}, nil)
	if ErrorCode(err) != ReasonCredentialUnavailable {
		t.Fatalf("error=%v", err)
	}
	if runner.executable != "" {
		t.Fatal("runner should not have been called")
	}
}

var _ = errors.New
