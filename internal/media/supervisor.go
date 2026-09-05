package media

import (
	"context"
	"sync"
	"time"

	"github.com/GoreeCloud/goreecloud-home-security/internal/camera"
	"github.com/GoreeCloud/goreecloud-home-security/internal/config"
)

type Supervisor struct {
	cameras                []config.Camera
	registry               *camera.Registry
	session                Session
	minBackoff, maxBackoff time.Duration
	now                    func() time.Time
}

func NewSupervisor(cameras []config.Camera, registry *camera.Registry, session Session, minBackoff, maxBackoff time.Duration) *Supervisor {
	if minBackoff <= 0 {
		minBackoff = time.Second
	}
	if maxBackoff <= 0 {
		maxBackoff = 30 * time.Second
	}
	if maxBackoff < minBackoff {
		maxBackoff = minBackoff
	}
	return &Supervisor{cameras: append([]config.Camera(nil), cameras...), registry: registry, session: session, minBackoff: minBackoff, maxBackoff: maxBackoff, now: time.Now}
}
func (s *Supervisor) Run(ctx context.Context) {
	var wg sync.WaitGroup
	for _, cfg := range s.cameras {
		cfg := cfg
		if !cfg.Enabled {
			_ = s.registry.UpdateSession(cfg.ID, camera.SessionStatus{State: camera.SessionDisabled})
			continue
		}
		wg.Add(1)
		go func() { defer wg.Done(); s.runCamera(ctx, cfg) }()
	}
	wg.Wait()
}
func (s *Supervisor) runCamera(ctx context.Context, cfg config.Camera) {
	attempt := 0
	for {
		if ctx.Err() != nil {
			s.publishStopped(cfg.ID, attempt)
			return
		}
		attempt++
		startAt := s.now().UTC()
		_ = s.registry.UpdateSession(cfg.ID, camera.SessionStatus{State: camera.SessionStarting, Attempt: attempt, LastStartAt: &startAt})
		started := false
		err := s.session.Run(ctx, cfg, func() {
			started = true
			now := s.now().UTC()
			_ = s.registry.UpdateSession(cfg.ID, camera.SessionStatus{State: camera.SessionRunning, Attempt: attempt, LastStartAt: &now})
		})
		if ctx.Err() != nil {
			s.publishStopped(cfg.ID, attempt)
			return
		}
		reason := ErrorCode(err)
		if isNonRetryableSessionReason(reason) {
			exitAt := s.now().UTC()
			_ = s.registry.UpdateSession(cfg.ID, camera.SessionStatus{State: camera.SessionBlocked, Attempt: attempt, LastStartAt: &startAt, LastExitAt: &exitAt, Reason: reason})
			return
		}
		exitAt := s.now().UTC()
		delay := backoffForAttempt(attempt, s.minBackoff, s.maxBackoff)
		next := exitAt.Add(delay)
		status := camera.SessionStatus{State: camera.SessionBackoff, Attempt: attempt, LastExitAt: &exitAt, NextRetryAt: &next, Reason: reason}
		if started {
			status.LastStartAt = &startAt
		}
		_ = s.registry.UpdateSession(cfg.ID, status)
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			s.publishStopped(cfg.ID, attempt)
			return
		case <-timer.C:
		}
	}
}
func isNonRetryableSessionReason(reason string) bool {
	switch reason {
	case ReasonCredentialedProbeBlocked, ReasonCredentialUnavailable, ReasonRTSPAuthRequired, ReasonRTSPAuthFailed, ReasonRTSPAuthUnsupported, ReasonRTSPBasicInsecure:
		return true
	default:
		return false
	}
}
func (s *Supervisor) publishStopped(id string, attempt int) {
	now := s.now().UTC()
	_ = s.registry.UpdateSession(id, camera.SessionStatus{State: camera.SessionStopped, Attempt: attempt, LastExitAt: &now})
}
func backoffForAttempt(attempt int, min, max time.Duration) time.Duration {
	if attempt <= 1 {
		return min
	}
	d := min
	for i := 1; i < attempt; i++ {
		if d >= max/2 {
			return max
		}
		d *= 2
	}
	if d > max {
		return max
	}
	return d
}
