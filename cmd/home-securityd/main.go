package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/GoreeCloud/goreecloud-home-security/internal/api"
	"github.com/GoreeCloud/goreecloud-home-security/internal/camera"
	"github.com/GoreeCloud/goreecloud-home-security/internal/config"
	"github.com/GoreeCloud/goreecloud-home-security/internal/events"
	"github.com/GoreeCloud/goreecloud-home-security/internal/media"
)

const (
	version                = "unreleased-development"
	retentionSweepInterval = 6 * time.Hour
)

func main() {
	if err := run(); err != nil {
		slog.Error("home security stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	defaultConfig := os.Getenv("GOREECLOUD_HOME_SECURITY_CONFIG")
	if defaultConfig == "" {
		defaultConfig = "./config/local.json"
	}
	configPath := flag.String("config", defaultConfig, "path to JSON configuration")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println(version)
		return nil
	}
	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}
	registry := camera.NewRegistry(cfg.Cameras)
	journal, err := events.NewJournal(filepath.Join(cfg.DataDir, "events.jsonl"))
	if err != nil {
		return err
	}
	retention := time.Duration(cfg.EventRetentionDays) * 24 * time.Hour
	if _, err := journal.PruneBefore(time.Now().UTC().Add(-retention)); err != nil {
		return fmt.Errorf("enforce event retention at startup: %w", err)
	}
	handler := api.New(registry, journal).Handler()
	httpServer := &http.Server{Addr: cfg.ListenAddress, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	prober := media.NewFFProbe("ffprobe", time.Duration(cfg.MediaProbeTimeoutSeconds)*time.Second)
	mediaManager := media.NewManager(cfg.Cameras, registry, prober, time.Duration(cfg.MediaProbeIntervalSeconds)*time.Second)
	go mediaManager.Run(ctx)
	if cfg.MediaSessionsEnabled {
		session := media.NewWorkerSession(cfg.MediaWorkerExecutable, os.LookupEnv, time.Duration(cfg.MediaSessionRWTimeoutSeconds)*time.Second)
		supervisor := media.NewSupervisor(cfg.Cameras, registry, session, time.Duration(cfg.MediaSessionRestartMinSeconds)*time.Second, time.Duration(cfg.MediaSessionRestartMaxSeconds)*time.Second)
		go supervisor.Run(ctx)
	}
	retentionErrCh := make(chan error, 1)
	go func() {
		retentionErrCh <- runEventRetention(ctx, journal, retention)
	}()
	errCh := make(chan error, 1)
	go func() {
		slog.Info("GoreeCloud Home Security development API listening", "address", cfg.ListenAddress, "cameras", registry.Count(), "media_sessions_enabled", cfg.MediaSessionsEnabled, "event_retention_days", cfg.EventRetentionDays)
		errCh <- httpServer.ListenAndServe()
	}()
	select {
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	case err := <-retentionErrCh:
		if err == nil {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	}
}

func runEventRetention(ctx context.Context, journal *events.Journal, retention time.Duration) error {
	if retention < 24*time.Hour || retention > time.Duration(config.MaxEventRetentionDays)*24*time.Hour {
		return fmt.Errorf("event retention duration is outside configured bounds")
	}
	ticker := time.NewTicker(retentionSweepInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case now := <-ticker.C:
			if _, err := journal.PruneBefore(now.UTC().Add(-retention)); err != nil {
				return fmt.Errorf("enforce event retention: %w", err)
			}
		}
	}
}
