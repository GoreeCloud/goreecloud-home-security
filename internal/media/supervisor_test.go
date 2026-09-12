package media

import (
	"context"
	"github.com/GoreeCloud/goreecloud-home-security/internal/camera"
	"github.com/GoreeCloud/goreecloud-home-security/internal/config"
	"testing"
	"time"
)

type reasonSession struct{ reason string }

func (f reasonSession) Run(context.Context, config.Camera, func()) error {
	return &ProbeError{Code: f.reason}
}
func TestSupervisorDoesNotRetryAuthenticationFailure(t *testing.T) {
	cfg := []config.Camera{{ID: "front-door", Name: "Front Door", StreamURL: "rtsp://camera.local/live", Enabled: true}}
	reg := camera.NewRegistry(cfg)
	NewSupervisor(cfg, reg, reasonSession{ReasonRTSPAuthFailed}, time.Millisecond, 2*time.Millisecond).Run(context.Background())
	s := reg.Snapshot()[0].Session
	if s.State != camera.SessionBlocked || s.Attempt != 1 || s.Reason != ReasonRTSPAuthFailed {
		t.Fatalf("status=%#v", s)
	}
}
func TestSupervisorBacksOffConnectivityFailure(t *testing.T) {
	if isNonRetryableSessionReason(ReasonRTSPConnectFailed) {
		t.Fatal("connect failure should retry")
	}
}
