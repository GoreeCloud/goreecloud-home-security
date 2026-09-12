package media

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"strings"
	"sync"
	"testing"
	"time"
)

type testRTSPRequest struct {
	method, uri string
	header      textproto.MIMEHeader
}

func readTestRequest(r *bufio.Reader) (testRTSPRequest, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return testRTSPRequest{}, err
	}
	parts := strings.SplitN(strings.TrimSpace(line), " ", 3)
	if len(parts) != 3 {
		return testRTSPRequest{}, fmt.Errorf("bad request line %q", line)
	}
	h, err := textproto.NewReader(r).ReadMIMEHeader()
	if err != nil {
		return testRTSPRequest{}, err
	}
	return testRTSPRequest{method: parts[0], uri: parts[1], header: h}, nil
}

func writeTestResponse(w *bufio.Writer, cseq string, status int, extra map[string]string, body string) error {
	if _, err := fmt.Fprintf(w, "RTSP/1.0 %d Test\r\nCSeq: %s\r\n", status, cseq); err != nil {
		return err
	}
	for k, v := range extra {
		if _, err := fmt.Fprintf(w, "%s: %s\r\n", k, v); err != nil {
			return err
		}
	}
	if body != "" {
		if _, err := fmt.Fprintf(w, "Content-Length: %d\r\nContent-Type: application/sdp\r\n", len(body)); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w, "\r\n"); err != nil {
		return err
	}
	if body != "" {
		if _, err := io.WriteString(w, body); err != nil {
			return err
		}
	}
	return w.Flush()
}

func TestRTSPWorkerDigestAuthReachesMediaReadyWithoutSecretInRequestURI(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	serverErr := make(chan error, 1)
	var mu sync.Mutex
	seenAuth := ""
	seenURIs := []string{}
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			serverErr <- err
			return
		}
		defer conn.Close()
		r := bufio.NewReader(conn)
		w := bufio.NewWriter(conn)
		first, err := readTestRequest(r)
		if err != nil {
			serverErr <- err
			return
		}
		if first.method != "DESCRIBE" {
			serverErr <- fmt.Errorf("first method %s", first.method)
			return
		}
		if err := writeTestResponse(w, first.header.Get("CSeq"), 401, map[string]string{"WWW-Authenticate": `Digest realm="camera", nonce="abcdef0123456789", algorithm=MD5, qop="auth"`}, ""); err != nil {
			serverErr <- err
			return
		}
		second, err := readTestRequest(r)
		if err != nil {
			serverErr <- err
			return
		}
		mu.Lock()
		seenAuth = second.header.Get("Authorization")
		seenURIs = append(seenURIs, second.uri)
		mu.Unlock()
		base := "rtsp://" + ln.Addr().String() + "/stream/"
		sdp := "v=0\r\nm=video 0 RTP/AVP 96\r\na=control:trackID=0\r\n"
		if err := writeTestResponse(w, second.header.Get("CSeq"), 200, map[string]string{"Content-Base": base}, sdp); err != nil {
			serverErr <- err
			return
		}
		setup, err := readTestRequest(r)
		if err != nil {
			serverErr <- err
			return
		}
		mu.Lock()
		seenURIs = append(seenURIs, setup.uri)
		mu.Unlock()
		if err := writeTestResponse(w, setup.header.Get("CSeq"), 200, map[string]string{"Session": "session-1;timeout=60", "Transport": "RTP/AVP/TCP;unicast;interleaved=0-1"}, ""); err != nil {
			serverErr <- err
			return
		}
		play, err := readTestRequest(r)
		if err != nil {
			serverErr <- err
			return
		}
		mu.Lock()
		seenURIs = append(seenURIs, play.uri)
		mu.Unlock()
		if err := writeTestResponse(w, play.header.Get("CSeq"), 200, nil, ""); err != nil {
			serverErr <- err
			return
		}
		if _, err := conn.Write([]byte{'$', 0, 0, 4, 1, 2, 3, 4}); err != nil {
			serverErr <- err
			return
		}
		_, _ = io.Copy(io.Discard, conn)
		serverErr <- nil
	}()

	ctx, cancel := context.WithCancel(context.Background())
	statusR, statusW := io.Pipe()
	d := WorkerDescriptor{Version: WorkerDescriptorVersion, CameraID: "front-door", StreamURL: "rtsp://" + ln.Addr().String() + "/stream", Username: "alice", Password: "top-secret", RWTimeoutSeconds: 5}
	workerDone := make(chan error, 1)
	go func() { workerDone <- RunRTSPWorker(ctx, d, statusW); _ = statusW.Close() }()
	scanner := bufio.NewScanner(statusR)
	if !scanner.Scan() {
		t.Fatalf("no status event: %v", scanner.Err())
	}
	event, err := DecodeWorkerEvent(scanner.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if event.Type != WorkerEventMediaReady {
		t.Fatalf("event=%#v", event)
	}
	cancel()
	select {
	case err := <-workerDone:
		if err != context.Canceled {
			t.Fatalf("worker error=%v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("worker did not stop")
	}
	mu.Lock()
	auth, uris := seenAuth, append([]string(nil), seenURIs...)
	mu.Unlock()
	if !strings.HasPrefix(auth, "Digest ") || strings.Contains(auth, "top-secret") {
		t.Fatalf("authorization=%q", auth)
	}
	for _, uri := range uris {
		if strings.Contains(uri, "alice") || strings.Contains(uri, "top-secret") {
			t.Fatalf("secret in uri %q", uri)
		}
	}
	select {
	case err := <-serverErr:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("server did not stop")
	}
}

func TestRTSPWorkerRejectsBasicAuthOverPlainRTSP(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		r := bufio.NewReader(conn)
		w := bufio.NewWriter(conn)
		req, err := readTestRequest(r)
		if err != nil {
			return
		}
		_ = writeTestResponse(w, req.header.Get("CSeq"), 401, map[string]string{"WWW-Authenticate": `Basic realm="camera"`}, "")
	}()
	d := WorkerDescriptor{Version: WorkerDescriptorVersion, CameraID: "front-door", StreamURL: "rtsp://" + ln.Addr().String() + "/stream", Username: "alice", Password: "secret", RWTimeoutSeconds: 5}
	err = RunRTSPWorker(context.Background(), d, io.Discard)
	if ErrorCode(err) != ReasonRTSPBasicInsecure {
		t.Fatalf("error=%v code=%s", err, ErrorCode(err))
	}
}

func TestParseSDPRequiresVideoControl(t *testing.T) {
	_, err := parseSDP([]byte("v=0\r\nm=audio 0 RTP/AVP 0\r\na=control:audio\r\n"))
	if err == nil {
		t.Fatal("expected video control failure")
	}
}
