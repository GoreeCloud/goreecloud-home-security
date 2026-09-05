package media

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/GoreeCloud/goreecloud-home-security/internal/config"
)

type SessionPlan struct {
	Executable string
	Args       []string
}

func BuildSessionPlan(executable string, camera config.Camera, rwTimeout time.Duration) (SessionPlan, error) {
	if err := camera.Validate(); err != nil {
		return SessionPlan{}, err
	}
	if !camera.Enabled {
		return SessionPlan{}, errors.New("camera must be enabled")
	}
	if camera.UsernameEnv != "" || camera.PasswordEnv != "" {
		return SessionPlan{}, &ProbeError{Code: ReasonCredentialedProbeBlocked}
	}
	if rwTimeout < 5*time.Second || rwTimeout > 5*time.Minute {
		return SessionPlan{}, errors.New("session rw timeout must be between 5 seconds and 5 minutes")
	}
	if strings.TrimSpace(executable) == "" {
		executable = "ffmpeg"
	}
	args := []string{"-nostdin", "-hide_banner", "-loglevel", "error", "-rw_timeout", fmt.Sprintf("%d", rwTimeout.Microseconds()), "-rtsp_transport", "tcp", "-i", camera.StreamURL, "-map", "0:v:0", "-c", "copy", "-f", "null", "-"}
	return SessionPlan{Executable: executable, Args: args}, nil
}

type ProcessRunner interface {
	Run(ctx context.Context, executable string, args []string, started func()) error
}

type execProcessRunner struct{}

func (execProcessRunner) Run(ctx context.Context, executable string, args []string, started func()) error {
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Env = SanitizedWorkerEnvironment(nil)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		return err
	}
	if started != nil {
		started()
	}
	return cmd.Wait()
}

type Session interface {
	Run(context.Context, config.Camera, func()) error
}

type FFmpegSession struct {
	executable string
	rwTimeout  time.Duration
	runner     ProcessRunner
}

func NewFFmpegSession(executable string, rwTimeout time.Duration) *FFmpegSession {
	return &FFmpegSession{executable: executable, rwTimeout: rwTimeout, runner: execProcessRunner{}}
}
func newFFmpegSessionWithRunner(executable string, rwTimeout time.Duration, runner ProcessRunner) *FFmpegSession {
	return &FFmpegSession{executable: executable, rwTimeout: rwTimeout, runner: runner}
}

func (s *FFmpegSession) Run(ctx context.Context, camera config.Camera, started func()) error {
	plan, err := BuildSessionPlan(s.executable, camera, s.rwTimeout)
	if err != nil {
		return err
	}
	err = s.runner.Run(ctx, plan.Executable, plan.Args, started)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err == nil {
		return &ProbeError{Code: ReasonSessionExited}
	}
	var execErr *exec.Error
	if errors.As(err, &execErr) {
		return &ProbeError{Code: ReasonDependencyUnavailable}
	}
	return &ProbeError{Code: ReasonSessionFailed}
}
