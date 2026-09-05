package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	DefaultListenAddress                 = "127.0.0.1:8787"
	DefaultMediaProbeIntervalSeconds     = 60
	DefaultMediaProbeTimeoutSeconds      = 8
	DefaultMediaSessionRWTimeoutSeconds  = 15
	DefaultMediaSessionRestartMinSeconds = 1
	DefaultMediaSessionRestartMaxSeconds = 30
	DefaultMediaWorkerExecutable         = "home-security-media-worker"
)

var (
	cameraIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,62}$`)
	envNamePattern  = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)
)

type Config struct {
	ListenAddress                 string   `json:"listen_address"`
	DataDir                       string   `json:"data_dir"`
	MediaProbeIntervalSeconds     int      `json:"media_probe_interval_seconds,omitempty"`
	MediaProbeTimeoutSeconds      int      `json:"media_probe_timeout_seconds,omitempty"`
	MediaSessionsEnabled          bool     `json:"media_sessions_enabled,omitempty"`
	MediaSessionRWTimeoutSeconds  int      `json:"media_session_rw_timeout_seconds,omitempty"`
	MediaSessionRestartMinSeconds int      `json:"media_session_restart_min_seconds,omitempty"`
	MediaSessionRestartMaxSeconds int      `json:"media_session_restart_max_seconds,omitempty"`
	MediaWorkerExecutable         string   `json:"media_worker_executable,omitempty"`
	Cameras                       []Camera `json:"cameras"`
}

type Camera struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	StreamURL   string `json:"stream_url"`
	UsernameEnv string `json:"username_env,omitempty"`
	PasswordEnv string `json:"password_env,omitempty"`
	Enabled     bool   `json:"enabled"`
}

func Load(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	var cfg Config
	if err := dec.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	if err := ensureEOF(dec); err != nil {
		return Config{}, err
	}
	cfg.applyDefaults()
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func ensureEOF(dec *json.Decoder) error {
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("decode config: multiple JSON values are not allowed")
		}
		return fmt.Errorf("decode config: %w", err)
	}
	return nil
}

func (c *Config) applyDefaults() {
	if strings.TrimSpace(c.ListenAddress) == "" {
		c.ListenAddress = DefaultListenAddress
	}
	if strings.TrimSpace(c.DataDir) == "" {
		c.DataDir = "./data"
	}
	if c.MediaProbeIntervalSeconds == 0 {
		c.MediaProbeIntervalSeconds = DefaultMediaProbeIntervalSeconds
	}
	if c.MediaProbeTimeoutSeconds == 0 {
		c.MediaProbeTimeoutSeconds = DefaultMediaProbeTimeoutSeconds
	}
	if c.MediaSessionRWTimeoutSeconds == 0 {
		c.MediaSessionRWTimeoutSeconds = DefaultMediaSessionRWTimeoutSeconds
	}
	if c.MediaSessionRestartMinSeconds == 0 {
		c.MediaSessionRestartMinSeconds = DefaultMediaSessionRestartMinSeconds
	}
	if c.MediaSessionRestartMaxSeconds == 0 {
		c.MediaSessionRestartMaxSeconds = DefaultMediaSessionRestartMaxSeconds
	}
	if strings.TrimSpace(c.MediaWorkerExecutable) == "" {
		c.MediaWorkerExecutable = DefaultMediaWorkerExecutable
	}
}

func (c Config) Validate() error {
	if err := validateLoopbackListen(c.ListenAddress); err != nil {
		return err
	}
	if strings.TrimSpace(c.DataDir) == "" {
		return errors.New("data_dir must not be empty")
	}
	if c.MediaProbeIntervalSeconds < 5 || c.MediaProbeIntervalSeconds > 3600 {
		return errors.New("media_probe_interval_seconds must be between 5 and 3600")
	}
	if c.MediaProbeTimeoutSeconds < 1 || c.MediaProbeTimeoutSeconds > 60 {
		return errors.New("media_probe_timeout_seconds must be between 1 and 60")
	}
	if c.MediaSessionRWTimeoutSeconds < 5 || c.MediaSessionRWTimeoutSeconds > 300 {
		return errors.New("media_session_rw_timeout_seconds must be between 5 and 300")
	}
	if c.MediaSessionRestartMinSeconds < 1 || c.MediaSessionRestartMinSeconds > 60 {
		return errors.New("media_session_restart_min_seconds must be between 1 and 60")
	}
	if c.MediaSessionRestartMaxSeconds < c.MediaSessionRestartMinSeconds || c.MediaSessionRestartMaxSeconds > 300 {
		return errors.New("media_session_restart_max_seconds must be between media_session_restart_min_seconds and 300")
	}
	if c.MediaWorkerExecutable != "" && (len(c.MediaWorkerExecutable) > 4096 || strings.ContainsRune(c.MediaWorkerExecutable, '\x00') || filepath.Base(c.MediaWorkerExecutable) != DefaultMediaWorkerExecutable) {
		return errors.New("media_worker_executable must name the GoreeCloud home-security-media-worker binary")
	}
	if c.MediaSessionsEnabled && strings.TrimSpace(c.MediaWorkerExecutable) == "" {
		return errors.New("media_worker_executable is required when media sessions are enabled")
	}
	seen := make(map[string]struct{}, len(c.Cameras))
	for i, camera := range c.Cameras {
		if err := camera.Validate(); err != nil {
			return fmt.Errorf("camera[%d]: %w", i, err)
		}
		if _, ok := seen[camera.ID]; ok {
			return fmt.Errorf("camera[%d]: duplicate camera id %q", i, camera.ID)
		}
		seen[camera.ID] = struct{}{}
	}
	return nil
}

func validateLoopbackListen(address string) error {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("listen_address must be host:port: %w", err)
	}
	if port == "" {
		return errors.New("listen_address must include a port")
	}
	if strings.EqualFold(host, "localhost") {
		return nil
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return errors.New("listen_address must use loopback while GoreeCloud Identity/Wardveil network authorization is not integrated")
	}
	return nil
}

func (c Camera) Validate() error {
	if !cameraIDPattern.MatchString(c.ID) {
		return fmt.Errorf("id %q must match %s", c.ID, cameraIDPattern.String())
	}
	if strings.TrimSpace(c.Name) == "" {
		return errors.New("name must not be empty")
	}
	u, err := url.Parse(c.StreamURL)
	if err != nil {
		return fmt.Errorf("stream_url: %w", err)
	}
	if u.Scheme != "rtsp" && u.Scheme != "rtsps" {
		return errors.New("stream_url must use rtsp or rtsps")
	}
	if u.Host == "" {
		return errors.New("stream_url must include a host")
	}
	if u.User != nil {
		return errors.New("stream_url must not contain credentials; use username_env/password_env secret references")
	}
	if (c.UsernameEnv == "") != (c.PasswordEnv == "") {
		return errors.New("username_env and password_env must be provided together")
	}
	if c.UsernameEnv != "" && !envNamePattern.MatchString(c.UsernameEnv) {
		return fmt.Errorf("username_env %q is not a valid environment-variable name", c.UsernameEnv)
	}
	if c.PasswordEnv != "" && !envNamePattern.MatchString(c.PasswordEnv) {
		return fmt.Errorf("password_env %q is not a valid environment-variable name", c.PasswordEnv)
	}
	return nil
}
