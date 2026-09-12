package media

import (
	"context"
	"sync"
	"time"

	"github.com/GoreeCloud/goreecloud-home-security/internal/camera"
	"github.com/GoreeCloud/goreecloud-home-security/internal/config"
)

type Manager struct {
	cameras  []config.Camera
	registry *camera.Registry
	prober   Prober
	interval time.Duration
	now      func() time.Time
}

func NewManager(
	cameras []config.Camera,
	registry *camera.Registry,
	prober Prober,
	interval time.Duration,
) *Manager {
	if interval <= 0 {
		interval = time.Minute
	}
	return &Manager{
		cameras:  append([]config.Camera(nil), cameras...),
		registry: registry,
		prober:   prober,
		interval: interval,
		now:      time.Now,
	}
}

func (m *Manager) Run(ctx context.Context) {
	m.ProbeAll(ctx)

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.ProbeAll(ctx)
		}
	}
}

func (m *Manager) ProbeAll(ctx context.Context) {
	var wg sync.WaitGroup
	for _, cfg := range m.cameras {
		cfg := cfg
		if !cfg.Enabled {
			_ = m.registry.UpdateMedia(cfg.ID, camera.MediaStatus{State: camera.MediaDisabled})
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			m.probeOne(ctx, cfg)
		}()
	}
	wg.Wait()
}

func (m *Manager) probeOne(ctx context.Context, cfg config.Camera) {
	started := m.now().UTC()
	_ = m.registry.UpdateMedia(cfg.ID, camera.MediaStatus{
		State:       camera.MediaProbing,
		LastProbeAt: &started,
	})

	metadata, err := m.prober.Probe(ctx, cfg)
	finished := m.now().UTC()
	if err != nil {
		reason := ErrorCode(err)
		state := camera.MediaUnavailable
		if reason == ReasonCredentialedProbeBlocked {
			state = camera.MediaBlocked
		}
		_ = m.registry.UpdateMedia(cfg.ID, camera.MediaStatus{
			State:       state,
			LastProbeAt: &finished,
			Reason:      reason,
		})
		return
	}

	_ = m.registry.UpdateMedia(cfg.ID, camera.MediaStatus{
		State:       camera.MediaOnline,
		VideoCodec:  metadata.VideoCodec,
		AudioCodec:  metadata.AudioCodec,
		Width:       metadata.Width,
		Height:      metadata.Height,
		FPS:         metadata.FPS,
		LastProbeAt: &finished,
	})
}
