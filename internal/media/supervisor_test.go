package media

import (
	"context"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-home-security/internal/camera"
	"github.com/GoreeCloud/goreecloud-home-security/internal/config"
)

type fakeSession struct {
	block   bool
	started bool
}

func (f *fakeSession) Run(ctx context.Context, cfg config.Camera, started func()) error {
	if cfg.UsernameEnv != "" {
		return &ProbeError{Code: ReasonCredentialedProbeBlocked}
	}
	if started != nil {
		started()
		f.started = true
	}
	if f.block {
		<-ctx.Done()
		return ctx.Err()
	}
	return &ProbeError{Code: ReasonSessionFailed}
}
func TestSupervisorBlocksCredentialedCameraWithoutRetryLoop(t *testing.T) {
	cfg := []config.Camera{{ID: "front-door", Name: "Front Door", StreamURL: "rtsp://camera.local/live", UsernameEnv: "CAMERA_USER", PasswordEnv: "CAMERA_PASS", Enabled: true}}
	registry := camera.NewRegistry(cfg)
	sup := NewSupervisor(cfg, registry, &fakeSession{}, time.Millisecond, 2*time.Millisecond)
	sup.Run(context.Background())
	status := registry.Snapshot()[0].Session
	if status.State != camera.SessionBlocked || status.Reason != ReasonCredentialedProbeBlocked || status.Attempt != 1 {
		t.Fatalf("status = %#v", status)
	}
}
func TestSupervisorPublishesRunningThenStopsOnContext(t *testing.T) {
	cfg := []config.Camera{{ID: "garage", Name: "Garage", StreamURL: "rtsp://camera.local/live", Enabled: true}}
	registry := camera.NewRegistry(cfg)
	session := &fakeSession{block: true}
	sup := NewSupervisor(cfg, registry, session, time.Millisecond, 2*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { sup.Run(ctx); close(done) }()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if registry.Snapshot()[0].Session.State == camera.SessionRunning {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if registry.Snapshot()[0].Session.State != camera.SessionRunning {
		t.Fatalf("status = %#v", registry.Snapshot()[0].Session)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("supervisor did not stop")
	}
	if registry.Snapshot()[0].Session.State != camera.SessionStopped {
		t.Fatalf("status = %#v", registry.Snapshot()[0].Session)
	}
}
func TestBackoffCapsAtMaximum(t *testing.T) {
	if got := backoffForAttempt(10, time.Second, 30*time.Second); got != 30*time.Second {
		t.Fatalf("backoff = %v", got)
	}
}
