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
	if cfg.ListenAddress != DefaultListenAddress {
		t.Fatalf("listen = %q", cfg.ListenAddress)
	}
	if cfg.DataDir != "./data" {
		t.Fatalf("data_dir = %q", cfg.DataDir)
	}
	if cfg.MediaProbeIntervalSeconds != DefaultMediaProbeIntervalSeconds || cfg.MediaProbeTimeoutSeconds != DefaultMediaProbeTimeoutSeconds {
		t.Fatalf("media defaults = %#v", cfg)
	}
}

func TestCameraRejectsCredentialsInURL(t *testing.T) {
	camera := Camera{ID: "front-door", Name: "Front Door", StreamURL: "rtsp://admin:secret@camera.local/live", Enabled: true}
	if err := camera.Validate(); err == nil || !strings.Contains(err.Error(), "must not contain credentials") {
		t.Fatalf("expected credential rejection, got %v", err)
	}
}

func TestConfigRejectsNonLoopbackListen(t *testing.T) {
	cfg := Config{ListenAddress: "0.0.0.0:8787", DataDir: t.TempDir(), MediaProbeIntervalSeconds: 60, MediaProbeTimeoutSeconds: 8}
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "loopback") {
		t.Fatalf("expected loopback rejection, got %v", err)
	}
}

func TestConfigRejectsDuplicateCameraIDs(t *testing.T) {
	camera := Camera{ID: "garage", Name: "Garage", StreamURL: "rtsp://camera.local/live"}
	cfg := Config{ListenAddress: DefaultListenAddress, DataDir: t.TempDir(), MediaProbeIntervalSeconds: 60, MediaProbeTimeoutSeconds: 8, Cameras: []Camera{camera, camera}}
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "duplicate camera id") {
		t.Fatalf("expected duplicate rejection, got %v", err)
	}
}
