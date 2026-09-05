package camera

import (
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/GoreeCloud/goreecloud-home-security/internal/config"
)

type MediaState string

const (
	MediaDisabled    MediaState = "disabled"
	MediaUnknown     MediaState = "unknown"
	MediaProbing     MediaState = "probing"
	MediaOnline      MediaState = "online"
	MediaUnavailable MediaState = "unavailable"
	MediaBlocked     MediaState = "blocked"
)

type SessionState string

const (
	SessionDisabled SessionState = "disabled"
	SessionIdle     SessionState = "idle"
	SessionStarting SessionState = "starting"
	SessionRunning  SessionState = "running"
	SessionBackoff  SessionState = "backoff"
	SessionBlocked  SessionState = "blocked"
	SessionStopped  SessionState = "stopped"
)

var statusTokenPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)
var codecPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,32}$`)

type MediaStatus struct {
	State       MediaState `json:"state"`
	VideoCodec  string     `json:"video_codec,omitempty"`
	AudioCodec  string     `json:"audio_codec,omitempty"`
	Width       int        `json:"width,omitempty"`
	Height      int        `json:"height,omitempty"`
	FPS         float64    `json:"fps,omitempty"`
	LastProbeAt *time.Time `json:"last_probe_at,omitempty"`
	Reason      string     `json:"reason,omitempty"`
}

type SessionStatus struct {
	State       SessionState `json:"state"`
	Attempt     int          `json:"attempt,omitempty"`
	LastStartAt *time.Time   `json:"last_start_at,omitempty"`
	LastExitAt  *time.Time   `json:"last_exit_at,omitempty"`
	NextRetryAt *time.Time   `json:"next_retry_at,omitempty"`
	Reason      string       `json:"reason,omitempty"`
}

type Public struct {
	ID      string        `json:"id"`
	Name    string        `json:"name"`
	Enabled bool          `json:"enabled"`
	Media   MediaStatus   `json:"media"`
	Session SessionStatus `json:"session"`
}

type Summary struct {
	Total           int `json:"total"`
	Enabled         int `json:"enabled"`
	Online          int `json:"online"`
	Unavailable     int `json:"unavailable"`
	Blocked         int `json:"blocked"`
	Unknown         int `json:"unknown"`
	SessionsRunning int `json:"sessions_running"`
	SessionsBackoff int `json:"sessions_backoff"`
	SessionsBlocked int `json:"sessions_blocked"`
}

type Registry struct {
	mu      sync.RWMutex
	cameras []Public
	index   map[string]int
}

func NewRegistry(cameras []config.Camera) *Registry {
	items := make([]Public, 0, len(cameras))
	index := make(map[string]int, len(cameras))
	for _, c := range cameras {
		mediaState := MediaUnknown
		sessionState := SessionIdle
		if !c.Enabled {
			mediaState = MediaDisabled
			sessionState = SessionDisabled
		}
		index[c.ID] = len(items)
		items = append(items, Public{
			ID:      c.ID,
			Name:    c.Name,
			Enabled: c.Enabled,
			Media:   MediaStatus{State: mediaState},
			Session: SessionStatus{State: sessionState},
		})
	}
	return &Registry{cameras: items, index: index}
}

func (r *Registry) Snapshot() []Public {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]Public(nil), r.cameras...)
}

func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.cameras)
}

func (r *Registry) UpdateMedia(cameraID string, status MediaStatus) error {
	if err := status.Validate(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	idx, ok := r.index[cameraID]
	if !ok {
		return fmt.Errorf("unknown camera %q", cameraID)
	}
	if !r.cameras[idx].Enabled && status.State != MediaDisabled {
		return errors.New("disabled camera media state must remain disabled")
	}
	r.cameras[idx].Media = status
	return nil
}

func (r *Registry) UpdateSession(cameraID string, status SessionStatus) error {
	if err := status.Validate(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	idx, ok := r.index[cameraID]
	if !ok {
		return fmt.Errorf("unknown camera %q", cameraID)
	}
	if !r.cameras[idx].Enabled && status.State != SessionDisabled {
		return errors.New("disabled camera session state must remain disabled")
	}
	r.cameras[idx].Session = status
	return nil
}

func (r *Registry) Summary() Summary {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := Summary{Total: len(r.cameras)}
	for _, item := range r.cameras {
		if item.Enabled {
			result.Enabled++
		}
		switch item.Media.State {
		case MediaOnline:
			result.Online++
		case MediaUnavailable:
			result.Unavailable++
		case MediaBlocked:
			result.Blocked++
		case MediaUnknown, MediaProbing:
			result.Unknown++
		}
		switch item.Session.State {
		case SessionRunning:
			result.SessionsRunning++
		case SessionBackoff:
			result.SessionsBackoff++
		case SessionBlocked:
			result.SessionsBlocked++
		}
	}
	return result
}

func (s MediaStatus) Validate() error {
	switch s.State {
	case MediaDisabled, MediaUnknown, MediaProbing, MediaOnline, MediaUnavailable, MediaBlocked:
	default:
		return fmt.Errorf("unsupported media state %q", s.State)
	}
	if s.Reason != "" && !statusTokenPattern.MatchString(s.Reason) {
		return errors.New("media status reason must be a bounded categorical token")
	}
	for field, value := range map[string]string{"video_codec": s.VideoCodec, "audio_codec": s.AudioCodec} {
		if value != "" && !codecPattern.MatchString(value) {
			return fmt.Errorf("%s contains unsupported characters", field)
		}
	}
	if s.Width < 0 || s.Width > 16384 || s.Height < 0 || s.Height > 16384 {
		return errors.New("media dimensions are out of range")
	}
	if s.FPS < 0 || s.FPS > 1000 {
		return errors.New("media fps is out of range")
	}
	return nil
}

func (s SessionStatus) Validate() error {
	switch s.State {
	case SessionDisabled, SessionIdle, SessionStarting, SessionRunning, SessionBackoff, SessionBlocked, SessionStopped:
	default:
		return fmt.Errorf("unsupported session state %q", s.State)
	}
	if s.Attempt < 0 || s.Attempt > 1_000_000 {
		return errors.New("session attempt is out of range")
	}
	if s.Reason != "" && !statusTokenPattern.MatchString(s.Reason) {
		return errors.New("session status reason must be a bounded categorical token")
	}
	return nil
}
