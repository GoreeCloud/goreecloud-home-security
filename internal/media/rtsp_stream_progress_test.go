package media

import (
	"bufio"
	"context"
	"io"
	"net"
	"testing"
	"time"
)

func TestRTSPStreamSustainedRTPProgressExtendsStallDeadline(t *testing.T) {
	workerConn, cameraConn := net.Pipe()
	defer workerConn.Close()
	defer cameraConn.Close()

	client := &rtspClient{
		conn:    workerConn,
		reader:  bufio.NewReader(workerConn),
		writer:  bufio.NewWriter(workerConn),
		timeout: 150 * time.Millisecond,
	}
	statusR, statusW := io.Pipe()
	defer statusR.Close()

	workerDone := make(chan error, 1)
	go func() {
		workerDone <- client.stream(
			context.Background(),
			"rtsp://camera.test/stream/",
			"session-1",
			60*time.Second,
			0,
			statusW,
		)
		_ = statusW.Close()
	}()

	writeRTP := func(payload byte) {
		t.Helper()
		packet := []byte{'$', 0, 0, 4, payload, payload, payload, payload}
		if _, err := cameraConn.Write(packet); err != nil {
			t.Fatalf("write RTP packet: %v", err)
		}
	}

	writeRTP(1)
	scanner := bufio.NewScanner(statusR)
	if !scanner.Scan() {
		t.Fatalf("expected media-ready event: %v", scanner.Err())
	}
	event, err := DecodeWorkerEvent(scanner.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if event.Type != WorkerEventMediaReady {
		t.Fatalf("event=%#v", event)
	}

	// Keep sending video RTP beyond the original first-packet stall deadline. The worker
	// must treat each negotiated video packet as progress rather than stalling relative
	// to the first packet only.
	time.Sleep(90 * time.Millisecond)
	writeRTP(2)
	time.Sleep(90 * time.Millisecond)
	writeRTP(3)

	select {
	case err := <-workerDone:
		t.Fatalf("worker stalled while video RTP was still progressing: %v", err)
	default:
	}

	// Once progress stops, the same bounded timeout must terminate the session with the
	// categorical stall reason rather than hanging indefinitely.
	select {
	case err := <-workerDone:
		if ErrorCode(err) != ReasonSessionStalled {
			t.Fatalf("error=%v code=%s", err, ErrorCode(err))
		}
	case <-time.After(750 * time.Millisecond):
		t.Fatal("worker did not report a bounded media stall")
	}
}
