package media

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/GoreeCloud/goreecloud-home-security/internal/config"
)

const (
	ReasonCredentialedProbeBlocked = "credentialed_probe_blocked"
	ReasonCredentialUnavailable    = "credential_unavailable"
	ReasonDependencyUnavailable    = "dependency_unavailable"
	ReasonProbeTimeout             = "probe_timeout"
	ReasonProbeFailed              = "probe_failed"
	ReasonInvalidProbeOutput       = "invalid_probe_output"
	ReasonSessionFailed            = "session_failed"
	ReasonSessionExited            = "session_exited"
	ReasonSessionStalled           = "session_stalled"
	ReasonRTSPConnectFailed        = "rtsp_connect_failed"
	ReasonRTSPAuthRequired         = "rtsp_auth_required"
	ReasonRTSPAuthFailed           = "rtsp_auth_failed"
	ReasonRTSPAuthUnsupported      = "rtsp_auth_unsupported"
	ReasonRTSPBasicInsecure        = "rtsp_basic_insecure"
	ReasonRTSPProtocolFailed       = "rtsp_protocol_failed"
)

var codecNamePattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,32}$`)

type Metadata struct {
	VideoCodec string
	AudioCodec string
	Width      int
	Height     int
	FPS        float64
}
type ProbeError struct{ Code string }

func (e *ProbeError) Error() string { return e.Code }
func ErrorCode(err error) string {
	if err == nil {
		return ""
	}
	var probeErr *ProbeError
	if errors.As(err, &probeErr) {
		return probeErr.Code
	}
	return ReasonProbeFailed
}

type Runner interface {
	Run(ctx context.Context, executable string, args []string) ([]byte, error)
}
type execRunner struct{}

func (execRunner) Run(ctx context.Context, executable string, args []string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Env = SanitizedWorkerEnvironment(nil)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return stdout.Bytes(), nil
}

type Prober interface {
	Probe(ctx context.Context, camera config.Camera) (Metadata, error)
}
type FFProbe struct {
	executable string
	timeout    time.Duration
	runner     Runner
}

func NewFFProbe(executable string, timeout time.Duration) *FFProbe {
	if strings.TrimSpace(executable) == "" {
		executable = "ffprobe"
	}
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	return &FFProbe{executable: executable, timeout: timeout, runner: execRunner{}}
}
func newFFProbeWithRunner(executable string, timeout time.Duration, runner Runner) *FFProbe {
	return &FFProbe{executable: executable, timeout: timeout, runner: runner}
}
func (p *FFProbe) Probe(ctx context.Context, camera config.Camera) (Metadata, error) {
	if err := camera.Validate(); err != nil {
		return Metadata{}, err
	}
	if camera.UsernameEnv != "" || camera.PasswordEnv != "" {
		return Metadata{}, &ProbeError{Code: ReasonCredentialedProbeBlocked}
	}
	probeCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()
	args := []string{"-v", "error", "-rtsp_transport", "tcp", "-show_entries", "stream=codec_type,codec_name,width,height,avg_frame_rate", "-of", "json", camera.StreamURL}
	payload, err := p.runner.Run(probeCtx, p.executable, args)
	if err != nil {
		if errors.Is(probeCtx.Err(), context.DeadlineExceeded) {
			return Metadata{}, &ProbeError{Code: ReasonProbeTimeout}
		}
		var execErr *exec.Error
		if errors.As(err, &execErr) {
			return Metadata{}, &ProbeError{Code: ReasonDependencyUnavailable}
		}
		return Metadata{}, &ProbeError{Code: ReasonProbeFailed}
	}
	metadata, err := parseProbe(payload)
	if err != nil {
		return Metadata{}, &ProbeError{Code: ReasonInvalidProbeOutput}
	}
	return metadata, nil
}

type ffprobeOutput struct {
	Streams []ffprobeStream `json:"streams"`
}
type ffprobeStream struct {
	CodecType    string `json:"codec_type"`
	CodecName    string `json:"codec_name"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	AvgFrameRate string `json:"avg_frame_rate"`
}

func parseProbe(payload []byte) (Metadata, error) {
	var output ffprobeOutput
	dec := json.NewDecoder(bytes.NewReader(payload))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&output); err != nil {
		return Metadata{}, fmt.Errorf("decode probe: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return Metadata{}, errors.New("probe output must contain one JSON value")
	}
	var result Metadata
	foundVideo := false
	for _, stream := range output.Streams {
		if stream.CodecName != "" && !codecNamePattern.MatchString(stream.CodecName) {
			return Metadata{}, errors.New("probe codec name is invalid")
		}
		switch stream.CodecType {
		case "video":
			if foundVideo {
				continue
			}
			if stream.Width < 1 || stream.Width > 16384 || stream.Height < 1 || stream.Height > 16384 {
				return Metadata{}, errors.New("probe video dimensions are invalid")
			}
			fps, err := parseRate(stream.AvgFrameRate)
			if err != nil {
				return Metadata{}, err
			}
			result.VideoCodec, result.Width, result.Height, result.FPS = stream.CodecName, stream.Width, stream.Height, fps
			foundVideo = true
		case "audio":
			if result.AudioCodec == "" {
				result.AudioCodec = stream.CodecName
			}
		}
	}
	if !foundVideo {
		return Metadata{}, errors.New("probe contains no video stream")
	}
	return result, nil
}
func parseRate(value string) (float64, error) {
	if value == "" || value == "0/0" {
		return 0, nil
	}
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		return 0, errors.New("invalid frame rate")
	}
	num, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, errors.New("invalid frame rate")
	}
	den, err := strconv.ParseFloat(parts[1], 64)
	if err != nil || den == 0 {
		return 0, errors.New("invalid frame rate")
	}
	fps := num / den
	if fps < 0 || fps > 1000 {
		return 0, errors.New("frame rate out of range")
	}
	return fps, nil
}
