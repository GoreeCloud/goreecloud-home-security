package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAppliesSafeDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"cameras":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ListenAddress != DefaultListenAddress || cfg.DataDir != "./data" {
		t.Fatalf("base defaults = %#v", cfg)
	}
	if cfg.EventRetentionDays != DefaultEventRetentionDays {
		t.Fatalf("event retention default = %d", cfg.EventRetentionDays)
	}
	if cfg.MediaProbeIntervalSeconds != DefaultMediaProbeIntervalSeconds || cfg.MediaProbeTimeoutSeconds != DefaultMediaProbeTimeoutSeconds {
		t.Fatalf("probe defaults = %#v", cfg)
	}
	if cfg.MediaSessionsEnabled {
		t.Fatal("media sessions must be opt-in")
	}
	if cfg.MediaSessionRWTimeoutSeconds != DefaultMediaSessionRWTimeoutSeconds || cfg.MediaSessionRestartMinSeconds != DefaultMediaSessionRestartMinSeconds || cfg.MediaSessionRestartMaxSeconds != DefaultMediaSessionRestartMaxSeconds {
		t.Fatalf("session defaults = %#v", cfg)
	}
	if cfg.MediaWorkerExecutable != DefaultMediaWorkerExecutable {
		t.Fatalf("worker executable = %q", cfg.MediaWorkerExecutable)
	}
}

func TestCameraRejectsCredentialsInURL(t *testing.T) {
	camera := Camera{ID: "front-door", Name: "Front Door", StreamURL: "rtsp://admin:secret@camera.local/live", Enabled: true}
	if err := camera.Validate(); err == nil || !strings.Contains(err.Error(), "must not contain credentials") {
		t.Fatalf("expected credential rejection, got %v", err)
	}
}

func TestConfigRejectsNonLoopbackListen(t *testing.T) {
	cfg := validConfig(t)
	cfg.ListenAddress = "0.0.0.0:8787"
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "loopback") {
		t.Fatalf("expected loopback rejection, got %v", err)
	}
}

func TestConfigRejectsDuplicateCameraIDs(t *testing.T) {
	camera := Camera{ID: "garage", Name: "Garage", StreamURL: "rtsp://camera.local/live"}
	cfg := validConfig(t)
	cfg.Cameras = []Camera{camera, camera}
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "duplicate camera id") {
		t.Fatalf("expected duplicate rejection, got %v", err)
	}
}

func TestConfigRejectsInvalidSessionBackoff(t *testing.T) {
	cfg := validConfig(t)
	cfg.MediaSessionRestartMinSeconds = 20
	cfg.MediaSessionRestartMaxSeconds = 10
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "restart_max") {
		t.Fatalf("expected session backoff rejection, got %v", err)
	}
}

func TestConfigRejectsNonGoreeCloudWorkerNameWhenEnabled(t *testing.T) {
	cfg := validConfig(t)
	cfg.MediaSessionsEnabled = true
	cfg.MediaWorkerExecutable = "/tmp/other-worker"
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "home-security-media-worker") {
		t.Fatalf("expected worker rejection, got %v", err)
	}
}

func TestConfigRejectsEventRetentionOutsideBounds(t *testing.T) {
	for _, days := range []int{-1, MaxEventRetentionDays + 1} {
		cfg := validConfig(t)
		cfg.EventRetentionDays = days
		if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "event_retention_days") {
			t.Fatalf("retention days %d: expected bounded retention rejection, got %v", days, err)
		}
	}
}

func validConfig(t *testing.T) Config {
	t.Helper()
	return Config{
		ListenAddress:                 DefaultListenAddress,
		DataDir:                       t.TempDir(),
		EventRetentionDays:            DefaultEventRetentionDays,
		MediaProbeIntervalSeconds:     DefaultMediaProbeIntervalSeconds,
		MediaProbeTimeoutSeconds:      DefaultMediaProbeTimeoutSeconds,
		MediaSessionRWTimeoutSeconds:  DefaultMediaSessionRWTimeoutSeconds,
		MediaSessionRestartMinSeconds: DefaultMediaSessionRestartMinSeconds,
		MediaSessionRestartMaxSeconds: DefaultMediaSessionRestartMaxSeconds,
		MediaWorkerExecutable:         DefaultMediaWorkerExecutable,
	}
}
