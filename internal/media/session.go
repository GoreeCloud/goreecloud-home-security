package media

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/GoreeCloud/goreecloud-home-security/internal/config"
)

type Session interface {
	Run(context.Context, config.Camera, func()) error
}

type WorkerProcessRunner interface {
	Run(ctx context.Context, executable string, descriptor WorkerDescriptor, started func()) error
}

type execWorkerProcessRunner struct{}

func (execWorkerProcessRunner) Run(ctx context.Context, executable string, descriptor WorkerDescriptor, started func()) error {
	descriptorReader, descriptorCleanup, err := NewWorkerDescriptorPipe(descriptor)
	if err != nil {
		return err
	}
	defer descriptorCleanup()

	statusReader, statusWriter, err := os.Pipe()
	if err != nil {
		return err
	}
	defer statusReader.Close()

	cmd := exec.CommandContext(ctx, executable, ProtectedWorkerArgs()...)
	cmd.Env = SanitizedWorkerEnvironment(nil)
	cmd.ExtraFiles = []*os.File{descriptorReader, statusWriter}
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		_ = statusWriter.Close()
		return err
	}
	_ = statusWriter.Close()

	eventCh := make(chan WorkerEvent)
	eventDone := make(chan struct{})
	go func() {
		defer close(eventDone)
		defer close(eventCh)
		scanner := bufio.NewScanner(statusReader)
		scanner.Buffer(make([]byte, 1024), workerEventMaxBytes)
		for scanner.Scan() {
			event, err := DecodeWorkerEvent(scanner.Bytes())
			if err != nil {
				continue
			}
			select {
			case eventCh <- event:
			case <-ctx.Done():
				return
			}
		}
	}()

	waitCh := make(chan error, 1)
	go func() { waitCh <- cmd.Wait() }()

	startedPublished := false
	for {
		select {
		case event, ok := <-eventCh:
			if ok && event.Type == WorkerEventMediaReady && !startedPublished {
				startedPublished = true
				if started != nil {
					started()
				}
			}
			if !ok {
				eventCh = nil
			}
		case err := <-waitCh:
			<-eventDone
			return err
		case <-ctx.Done():
			err := <-waitCh
			<-eventDone
			if err != nil {
				return ctx.Err()
			}
			return ctx.Err()
		}
	}
}

type WorkerSession struct {
	executable string
	lookup     SecretLookup
	rwTimeout  time.Duration
	runner     WorkerProcessRunner
}

func NewWorkerSession(executable string, lookup SecretLookup, rwTimeout time.Duration) *WorkerSession {
	if strings.TrimSpace(executable) == "" {
		executable = config.DefaultMediaWorkerExecutable
	}
	if rwTimeout <= 0 {
		rwTimeout = time.Duration(config.DefaultMediaSessionRWTimeoutSeconds) * time.Second
	}
	return &WorkerSession{executable: executable, lookup: lookup, rwTimeout: rwTimeout, runner: execWorkerProcessRunner{}}
}

func newWorkerSessionWithRunner(executable string, lookup SecretLookup, rwTimeout time.Duration, runner WorkerProcessRunner) *WorkerSession {
	return &WorkerSession{executable: executable, lookup: lookup, rwTimeout: rwTimeout, runner: runner}
}

func (s *WorkerSession) Run(ctx context.Context, camera config.Camera, started func()) error {
	if !camera.Enabled {
		return errors.New("camera must be enabled")
	}
	descriptor, err := ResolveWorkerDescriptor(camera, s.lookup, s.rwTimeout)
	if err != nil {
		return err
	}
	err = s.runner.Run(ctx, s.executable, descriptor, started)
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
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return &ProbeError{Code: WorkerReasonForExitCode(exitErr.ExitCode())}
	}
	return &ProbeError{Code: ReasonSessionFailed}
}
