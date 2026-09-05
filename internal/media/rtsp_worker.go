package media

import (
	"bufio"
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/textproto"
	"net/url"
	"time"
)

const (
	WorkerExitInvalidDescriptor = 20
	WorkerExitConnectFailed     = 21
	WorkerExitAuthRequired      = 22
	WorkerExitAuthFailed        = 23
	WorkerExitAuthUnsupported   = 24
	WorkerExitBasicInsecure     = 25
	WorkerExitProtocolFailed    = 26
	WorkerExitSessionStalled    = 27
	WorkerExitSessionFailed     = 28
)

func WorkerExitCodeForError(err error) int {
	switch ErrorCode(err) {
	case ReasonRTSPConnectFailed:
		return WorkerExitConnectFailed
	case ReasonRTSPAuthRequired:
		return WorkerExitAuthRequired
	case ReasonRTSPAuthFailed:
		return WorkerExitAuthFailed
	case ReasonRTSPAuthUnsupported:
		return WorkerExitAuthUnsupported
	case ReasonRTSPBasicInsecure:
		return WorkerExitBasicInsecure
	case ReasonRTSPProtocolFailed:
		return WorkerExitProtocolFailed
	case ReasonSessionStalled:
		return WorkerExitSessionStalled
	default:
		return WorkerExitSessionFailed
	}
}

func WorkerReasonForExitCode(code int) string {
	switch code {
	case WorkerExitConnectFailed:
		return ReasonRTSPConnectFailed
	case WorkerExitAuthRequired:
		return ReasonRTSPAuthRequired
	case WorkerExitAuthFailed:
		return ReasonRTSPAuthFailed
	case WorkerExitAuthUnsupported:
		return ReasonRTSPAuthUnsupported
	case WorkerExitBasicInsecure:
		return ReasonRTSPBasicInsecure
	case WorkerExitProtocolFailed, WorkerExitInvalidDescriptor:
		return ReasonRTSPProtocolFailed
	case WorkerExitSessionStalled:
		return ReasonSessionStalled
	default:
		return ReasonSessionFailed
	}
}

type rtspResponse struct {
	statusCode int
	header     textproto.MIMEHeader
	body       []byte
}

type rtspClient struct {
	conn      net.Conn
	reader    *bufio.Reader
	writer    *bufio.Writer
	timeout   time.Duration
	cseq      int
	auth      *rtspAuth
	secureTLS bool
}

func RunRTSPWorker(ctx context.Context, descriptor WorkerDescriptor, status io.Writer) error {
	if err := descriptor.Validate(); err != nil {
		return &ProbeError{Code: ReasonRTSPProtocolFailed}
	}
	u, err := url.Parse(descriptor.StreamURL)
	if err != nil {
		return &ProbeError{Code: ReasonRTSPProtocolFailed}
	}
	timeout := time.Duration(descriptor.RWTimeoutSeconds) * time.Second
	conn, err := dialRTSP(ctx, u, timeout)
	if err != nil {
		return &ProbeError{Code: ReasonRTSPConnectFailed}
	}
	defer conn.Close()

	client := &rtspClient{conn: conn, reader: bufio.NewReaderSize(conn, 64<<10), writer: bufio.NewWriterSize(conn, 16<<10), timeout: timeout, secureTLS: u.Scheme == "rtsps"}
	if descriptor.Username != "" {
		client.auth = &rtspAuth{username: descriptor.Username, password: descriptor.Password}
	}

	cancelDone := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = conn.SetDeadline(time.Now())
		case <-cancelDone:
		}
	}()
	defer close(cancelDone)

	describe, err := client.request(ctx, "DESCRIBE", descriptor.StreamURL, map[string]string{"Accept": "application/sdp"})
	if err != nil {
		return err
	}
	if describe.statusCode != 200 {
		return &ProbeError{Code: ReasonRTSPProtocolFailed}
	}
	baseURI := describe.header.Get("Content-Base")
	if baseURI == "" {
		baseURI = describe.header.Get("Content-Location")
	}
	if baseURI == "" {
		baseURI = descriptor.StreamURL
	}
	sdp, err := parseSDP(describe.body)
	if err != nil {
		return &ProbeError{Code: ReasonRTSPProtocolFailed}
	}
	trackURI, err := resolveRTSPControl(baseURI, sdp.videoControl)
	if err != nil {
		return &ProbeError{Code: ReasonRTSPProtocolFailed}
	}
	playURI := baseURI
	if sdp.sessionControl != "" && sdp.sessionControl != "*" {
		playURI, err = resolveRTSPControl(baseURI, sdp.sessionControl)
		if err != nil {
			return &ProbeError{Code: ReasonRTSPProtocolFailed}
		}
	}

	setup, err := client.request(ctx, "SETUP", trackURI, map[string]string{"Transport": "RTP/AVP/TCP;unicast;interleaved=0-1"})
	if err != nil {
		return err
	}
	if setup.statusCode != 200 {
		return &ProbeError{Code: ReasonRTSPProtocolFailed}
	}
	sessionID, sessionTimeout, err := parseSessionHeader(setup.header.Get("Session"))
	if err != nil {
		return &ProbeError{Code: ReasonRTSPProtocolFailed}
	}
	rtpChannel, err := parseInterleavedRTPChannel(setup.header.Get("Transport"))
	if err != nil {
		return &ProbeError{Code: ReasonRTSPProtocolFailed}
	}

	play, err := client.request(ctx, "PLAY", playURI, map[string]string{"Session": sessionID, "Range": "npt=0.000-"})
	if err != nil {
		return err
	}
	if play.statusCode != 200 {
		return &ProbeError{Code: ReasonRTSPProtocolFailed}
	}
	return client.stream(ctx, playURI, sessionID, sessionTimeout, rtpChannel, status)
}

func dialRTSP(ctx context.Context, u *url.URL, timeout time.Duration) (net.Conn, error) {
	port := u.Port()
	if port == "" {
		if u.Scheme == "rtsps" {
			port = "322"
		} else {
			port = "554"
		}
	}
	address := net.JoinHostPort(u.Hostname(), port)
	dialer := &net.Dialer{Timeout: timeout, KeepAlive: 30 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "rtsps" {
		return conn, nil
	}
	tlsConn := tls.Client(conn, &tls.Config{MinVersion: tls.VersionTLS12, ServerName: u.Hostname()})
	_ = tlsConn.SetDeadline(time.Now().Add(timeout))
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		_ = conn.Close()
		return nil, err
	}
	_ = tlsConn.SetDeadline(time.Time{})
	return tlsConn, nil
}
