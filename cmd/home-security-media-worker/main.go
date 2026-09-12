package main

import (
	"context"
	"flag"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/GoreeCloud/goreecloud-home-security/internal/media"
)

func main() {
	flags := flag.NewFlagSet("home-security-media-worker", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	descriptorFD := flags.Int("descriptor-fd", -1, "protected descriptor file descriptor")
	statusFD := flags.Int("status-fd", -1, "sanitized status file descriptor")
	if err := flags.Parse(os.Args[1:]); err != nil || flags.NArg() != 0 || *descriptorFD != media.WorkerDescriptorFD || *statusFD != media.WorkerStatusFD {
		os.Exit(media.WorkerExitInvalidDescriptor)
	}
	descriptorFile := os.NewFile(uintptr(*descriptorFD), "home-security-descriptor")
	statusFile := os.NewFile(uintptr(*statusFD), "home-security-status")
	if descriptorFile == nil || statusFile == nil {
		os.Exit(media.WorkerExitInvalidDescriptor)
	}
	descriptor, err := media.DecodeWorkerDescriptor(descriptorFile)
	_ = descriptorFile.Close()
	if err != nil {
		_ = statusFile.Close()
		os.Exit(media.WorkerExitInvalidDescriptor)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	err = media.RunRTSPWorker(ctx, descriptor, statusFile)
	_ = statusFile.Close()
	if err == nil || ctx.Err() != nil {
		return
	}
	os.Exit(media.WorkerExitCodeForError(err))
}
