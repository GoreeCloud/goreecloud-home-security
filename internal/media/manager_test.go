package media

import (
	"context"
	"errors"
	"testing"

	"github.com/GoreeCloud/goreecloud-home-security/internal/camera"
	"github.com/GoreeCloud/goreecloud-home-security/internal/config"
)

type fakeProber struct {
	result Metadata
	err    error
}

func (f fakeProber) Probe(context.Context, config.Camera) (Metadata, error) { return f.result, f.err }

func TestManagerPublishesOnlineStatusWithoutURLs(t *testing.T) {
	cfg := []config.Camera{{ID: "front-door", Name: "Front Door", StreamURL: "rtsp://camera.local/private-path", Enabled: true}}
	registry := camera.NewRegistry(cfg)
	manager := NewManager(cfg, registry, fakeProber{result: Metadata{VideoCodec: "h264", Width: 1280, Height: 720, FPS: 15}}, 0)
	manager.ProbeAll(context.Background())
	snapshot := registry.Snapshot()
	if snapshot[0].Media.State != camera.MediaOnline {
		t.Fatalf("status = %#v", snapshot[0].Media)
	}
	if snapshot[0].Media.VideoCodec != "h264" {
		t.Fatalf("status = %#v", snapshot[0].Media)
	}
}

func TestManagerUsesCategoricalFailureOnly(t *testing.T) {
	cfg := []config.Camera{{ID: "garage", Name: "Garage", StreamURL: "rtsp://camera.local/live", Enabled: true}}
	registry := camera.NewRegistry(cfg)
	manager := NewManager(cfg, registry, fakeProber{err: &ProbeError{Code: ReasonCredentialedProbeBlocked}}, 0)
	manager.ProbeAll(context.Background())
	status := registry.Snapshot()[0].Media
	if status.State != camera.MediaBlocked || status.Reason != ReasonCredentialedProbeBlocked {
		t.Fatalf("status = %#v", status)
	}
}

func TestUnknownErrorsCollapseToProbeFailed(t *testing.T) {
	cfg := []config.Camera{{ID: "garage", Name: "Garage", StreamURL: "rtsp://camera.local/live", Enabled: true}}
	registry := camera.NewRegistry(cfg)
	manager := NewManager(cfg, registry, fakeProber{err: errors.New("rtsp://secret-host/private path refused")}, 0)
	manager.ProbeAll(context.Background())
	status := registry.Snapshot()[0].Media
	if status.Reason != ReasonProbeFailed {
		t.Fatalf("status = %#v", status)
	}
}
