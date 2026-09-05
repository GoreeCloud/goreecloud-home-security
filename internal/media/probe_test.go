package media

import (
	"context"
	"errors"
	"os/exec"
	"reflect"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-home-security/internal/config"
)

type fakeRunner struct {
	payload    []byte
	err        error
	calls      int
	executable string
	args       []string
}

func (r *fakeRunner) Run(_ context.Context, executable string, args []string) ([]byte, error) {
	r.calls++
	r.executable = executable
	r.args = append([]string(nil), args...)
	return r.payload, r.err
}

func TestFFProbeParsesSanitizedMetadata(t *testing.T) {
	runner := &fakeRunner{payload: []byte(`{"streams":[{"codec_type":"video","codec_name":"h264","width":1920,"height":1080,"avg_frame_rate":"30000/1001"},{"codec_type":"audio","codec_name":"aac"}]}`)}
	p := newFFProbeWithRunner("ffprobe", time.Second, runner)
	metadata, err := p.Probe(context.Background(), config.Camera{ID: "front-door", Name: "Front Door", StreamURL: "rtsp://camera.local/live", Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if metadata.VideoCodec != "h264" || metadata.AudioCodec != "aac" || metadata.Width != 1920 || metadata.Height != 1080 {
		t.Fatalf("metadata = %#v", metadata)
	}
	if metadata.FPS < 29.9 || metadata.FPS > 30.1 {
		t.Fatalf("fps = %f", metadata.FPS)
	}
	wantPrefix := []string{"-v", "error", "-rtsp_transport", "tcp"}
	if !reflect.DeepEqual(runner.args[:4], wantPrefix) {
		t.Fatalf("args = %#v", runner.args)
	}
}

func TestFFProbeBlocksCredentialedCameraBeforeProcessExecution(t *testing.T) {
	runner := &fakeRunner{}
	p := newFFProbeWithRunner("ffprobe", time.Second, runner)
	_, err := p.Probe(context.Background(), config.Camera{ID: "front-door", Name: "Front Door", StreamURL: "rtsp://camera.local/live", UsernameEnv: "CAMERA_USER", PasswordEnv: "CAMERA_PASS", Enabled: true})
	if ErrorCode(err) != ReasonCredentialedProbeBlocked {
		t.Fatalf("error = %v", err)
	}
	if runner.calls != 0 {
		t.Fatalf("runner calls = %d", runner.calls)
	}
}

func TestFFProbeClassifiesMissingExecutable(t *testing.T) {
	runner := &fakeRunner{err: &exec.Error{Name: "missing-ffprobe", Err: errors.New("not found")}}
	p := newFFProbeWithRunner("missing-ffprobe", time.Second, runner)
	_, err := p.Probe(context.Background(), config.Camera{ID: "front-door", Name: "Front Door", StreamURL: "rtsp://camera.local/live", Enabled: true})
	if ErrorCode(err) != ReasonDependencyUnavailable {
		t.Fatalf("error code = %s", ErrorCode(err))
	}
}
