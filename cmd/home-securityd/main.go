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
)

const version = "unreleased-development"

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

	handler := api.New(registry, journal).Handler()
	httpServer := &http.Server{
		Addr:              cfg.ListenAddress,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		slog.Info("GoreeCloud Home Security development API listening", "address", cfg.ListenAddress, "cameras", registry.Count())
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	}
}
