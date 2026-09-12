package media

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/GoreeCloud/goreecloud-home-security/internal/config"
)

var recordingCameraIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,62}$`)

type RecordingPlan struct {
	Executable     string
	Args           []string
	SegmentPattern string
}

func BuildRecordingPlan(
	executable string,
	dataDir string,
	camera config.Camera,
	segmentDuration time.Duration,
) (RecordingPlan, error) {
	if !camera.Enabled {
		return RecordingPlan{}, errors.New("camera must be enabled")
	}
	if camera.UsernameEnv != "" || camera.PasswordEnv != "" {
		return RecordingPlan{}, &ProbeError{Code: ReasonCredentialedProbeBlocked}
	}
	if !recordingCameraIDPattern.MatchString(camera.ID) {
		return RecordingPlan{}, errors.New("camera id is not safe for recording paths")
	}
	if strings.TrimSpace(dataDir) == "" {
		return RecordingPlan{}, errors.New("data directory must not be empty")
	}
	if segmentDuration < 10*time.Second || segmentDuration > 10*time.Minute {
		return RecordingPlan{}, errors.New("segment duration must be between 10 seconds and 10 minutes")
	}
	if strings.TrimSpace(executable) == "" {
		executable = "ffmpeg"
	}

	root := filepath.Join(dataDir, "recordings", camera.ID)
	pattern := filepath.Join(root, "%Y", "%m", "%d", "%Y%m%dT%H%M%S.mkv")
	args := []string{
		"-nostdin",
		"-hide_banner",
		"-loglevel", "error",
		"-rtsp_transport", "tcp",
		"-i", camera.StreamURL,
		"-map", "0:v:0",
		"-map", "0:a?",
		"-c", "copy",
		"-f", "segment",
		"-segment_time", fmt.Sprintf("%.0f", segmentDuration.Seconds()),
		"-strftime", "1",
		"-reset_timestamps", "1",
		pattern,
	}
	return RecordingPlan{
		Executable:     executable,
		Args:           args,
		SegmentPattern: pattern,
	}, nil
}
