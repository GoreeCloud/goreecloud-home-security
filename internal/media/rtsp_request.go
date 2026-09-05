package media

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"strconv"
	"strings"
	"time"
)

func (c *rtspClient) request(ctx context.Context, method, uri string, headers map[string]string) (rtspResponse, error) {
	response, err := c.requestOnce(ctx, method, uri, headers)
	if err != nil {
		return rtspResponse{}, err
	}
	if response.statusCode != 401 {
		return response, nil
	}
	if c.auth == nil || c.auth.username == "" {
		return rtspResponse{}, &ProbeError{Code: ReasonRTSPAuthRequired}
	}
	challenge, err := selectAuthChallenge(response.header.Values("Www-Authenticate"), c.secureTLS)
	if err != nil {
		return rtspResponse{}, err
	}
	challenge.username, challenge.password = c.auth.username, c.auth.password
	c.auth = challenge
	response, err = c.requestOnce(ctx, method, uri, headers)
	if err != nil {
		return rtspResponse{}, err
	}
	if response.statusCode == 401 || response.statusCode == 403 {
		return rtspResponse{}, &ProbeError{Code: ReasonRTSPAuthFailed}
	}
	return response, nil
}

func (c *rtspClient) requestOnce(ctx context.Context, method, uri string, headers map[string]string) (rtspResponse, error) {
	if err := c.writeRequest(ctx, method, uri, headers); err != nil {
		return rtspResponse{}, &ProbeError{Code: ReasonSessionFailed}
	}
	response, err := c.readResponse(ctx)
	if err != nil {
		if isTimeout(err) {
			return rtspResponse{}, &ProbeError{Code: ReasonSessionStalled}
		}
		return rtspResponse{}, &ProbeError{Code: ReasonSessionFailed}
	}
	return response, nil
}

func (c *rtspClient) writeRequest(ctx context.Context, method, uri string, headers map[string]string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	_ = c.conn.SetWriteDeadline(time.Now().Add(c.timeout))
	c.cseq++
	if _, err := fmt.Fprintf(c.writer, "%s %s RTSP/1.0\r\nCSeq: %d\r\nUser-Agent: GoreeCloud-Home-Security/unreleased-development\r\n", method, uri, c.cseq); err != nil {
		return err
	}
	if c.auth != nil && c.auth.scheme != "" {
		authorization, err := c.auth.authorization(method, uri)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(c.writer, "Authorization: %s\r\n", authorization); err != nil {
			return err
		}
	}
	for key, value := range headers {
		if strings.ContainsAny(key, "\r\n") || strings.ContainsAny(value, "\r\n") {
			return errors.New("invalid rtsp header")
		}
		if _, err := fmt.Fprintf(c.writer, "%s: %s\r\n", key, value); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(c.writer, "\r\n"); err != nil {
		return err
	}
	return c.writer.Flush()
}

func (c *rtspClient) readResponse(ctx context.Context) (rtspResponse, error) {
	if err := ctx.Err(); err != nil {
		return rtspResponse{}, err
	}
	_ = c.conn.SetReadDeadline(time.Now().Add(c.timeout))
	line, err := readBoundedLine(c.reader, 4096)
	if err != nil {
		return rtspResponse{}, err
	}
	line = strings.TrimSpace(line)
	parts := strings.SplitN(line, " ", 3)
	if len(parts) < 2 || !strings.HasPrefix(parts[0], "RTSP/") {
		return rtspResponse{}, errors.New("invalid rtsp status line")
	}
	status, err := strconv.Atoi(parts[1])
	if err != nil || status < 100 || status > 599 {
		return rtspResponse{}, errors.New("invalid rtsp status")
	}
	header, err := readBoundedMIMEHeader(c.reader)
	if err != nil {
		return rtspResponse{}, err
	}
	length := 0
	if raw := header.Get("Content-Length"); raw != "" {
		length, err = strconv.Atoi(raw)
		if err != nil || length < 0 || length > 1<<20 {
			return rtspResponse{}, errors.New("invalid rtsp content length")
		}
	}
	body := make([]byte, length)
	if length > 0 {
		if _, err := io.ReadFull(c.reader, body); err != nil {
			return rtspResponse{}, err
		}
	}
	return rtspResponse{statusCode: status, header: header, body: body}, nil
}

func parseSessionHeader(value string) (string, time.Duration, error) {
	if value == "" {
		return "", 0, errors.New("missing session header")
	}
	parts := strings.Split(value, ";")
	id := strings.TrimSpace(parts[0])
	if id == "" || len(id) > 256 || strings.ContainsAny(id, "\r\n\x00") {
		return "", 0, errors.New("invalid session id")
	}
	timeout := 60 * time.Second
	for _, part := range parts[1:] {
		part = strings.TrimSpace(part)
		if !strings.HasPrefix(strings.ToLower(part), "timeout=") {
			continue
		}
		seconds, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(strings.ToLower(part), "timeout=")))
		if err == nil && seconds >= 5 && seconds <= 3600 {
			timeout = time.Duration(seconds) * time.Second
		}
	}
	return id, timeout, nil
}

func (c *rtspClient) stream(ctx context.Context, playURI, sessionID string, sessionTimeout time.Duration, rtpChannel byte, status io.Writer) error {
	lastMediaAt := time.Now()
	keepaliveEvery := sessionTimeout / 2
	if keepaliveEvery > 30*time.Second {
		keepaliveEvery = 30 * time.Second
	}
	if keepaliveEvery < 5*time.Second {
		keepaliveEvery = 5 * time.Second
	}
	nextKeepalive := time.Now().Add(keepaliveEvery)
	ready := false

	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		now := time.Now()
		stallAt := lastMediaAt.Add(c.timeout)
		if !now.Before(stallAt) {
			return &ProbeError{Code: ReasonSessionStalled}
		}
		if !now.Before(nextKeepalive) {
			if err := c.writeRequest(ctx, "OPTIONS", playURI, map[string]string{"Session": sessionID}); err != nil {
				return &ProbeError{Code: ReasonSessionFailed}
			}
			nextKeepalive = time.Now().Add(keepaliveEvery)
			continue
		}
		deadline := stallAt
		if nextKeepalive.Before(deadline) {
			deadline = nextKeepalive
		}
		_ = c.conn.SetReadDeadline(deadline)
		first, err := c.reader.Peek(1)
		if err != nil {
			if isTimeout(err) {
				continue
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return &ProbeError{Code: ReasonSessionFailed}
		}
		if first[0] == '$' {
			header := make([]byte, 4)
			if _, err := io.ReadFull(c.reader, header); err != nil {
				return &ProbeError{Code: ReasonSessionFailed}
			}
			length := int(header[2])<<8 | int(header[3])
			if length <= 0 {
				return &ProbeError{Code: ReasonRTSPProtocolFailed}
			}
			if _, err := io.CopyN(io.Discard, c.reader, int64(length)); err != nil {
				return &ProbeError{Code: ReasonSessionFailed}
			}
			if header[1] != rtpChannel {
				continue
			}
			lastMediaAt = time.Now()
			if !ready {
				if err := EncodeWorkerEvent(status, WorkerEvent{Version: WorkerEventVersion, Type: WorkerEventMediaReady}); err != nil {
					return &ProbeError{Code: ReasonSessionFailed}
				}
				ready = true
			}
			continue
		}
		if err := c.readControlMessage(); err != nil {
			return err
		}
	}
}

func (c *rtspClient) readControlMessage() error {
	line, err := readBoundedLine(c.reader, 4096)
	if err != nil {
		return &ProbeError{Code: ReasonSessionFailed}
	}
	line = strings.TrimSpace(line)
	header, err := readBoundedMIMEHeader(c.reader)
	if err != nil {
		return &ProbeError{Code: ReasonSessionFailed}
	}
	length := 0
	if raw := header.Get("Content-Length"); raw != "" {
		length, err = strconv.Atoi(raw)
		if err != nil || length < 0 || length > 1<<20 {
			return &ProbeError{Code: ReasonRTSPProtocolFailed}
		}
	}
	if length > 0 {
		if _, err := io.CopyN(io.Discard, c.reader, int64(length)); err != nil {
			return &ProbeError{Code: ReasonSessionFailed}
		}
	}
	if strings.HasPrefix(line, "RTSP/") {
		parts := strings.SplitN(line, " ", 3)
		if len(parts) >= 2 && (parts[1] == "401" || parts[1] == "403") {
			return &ProbeError{Code: ReasonRTSPAuthFailed}
		}
		return nil
	}
	parts := strings.SplitN(line, " ", 3)
	if len(parts) != 3 || !strings.HasPrefix(parts[2], "RTSP/") {
		return &ProbeError{Code: ReasonRTSPProtocolFailed}
	}
	cseq := header.Get("CSeq")
	if cseq == "" || strings.ContainsAny(cseq, "\r\n") {
		return &ProbeError{Code: ReasonRTSPProtocolFailed}
	}
	_ = c.conn.SetWriteDeadline(time.Now().Add(c.timeout))
	if _, err := fmt.Fprintf(c.writer, "RTSP/1.0 200 OK\r\nCSeq: %s\r\n\r\n", cseq); err != nil {
		return &ProbeError{Code: ReasonSessionFailed}
	}
	if err := c.writer.Flush(); err != nil {
		return &ProbeError{Code: ReasonSessionFailed}
	}
	return nil
}

func parseInterleavedRTPChannel(value string) (byte, error) {
	if value == "" {
		return 0, errors.New("missing transport header")
	}
	for _, part := range strings.Split(value, ";") {
		part = strings.TrimSpace(part)
		if !strings.HasPrefix(strings.ToLower(part), "interleaved=") {
			continue
		}
		raw := strings.TrimSpace(part[len("interleaved="):])
		pair := strings.SplitN(raw, "-", 2)
		if len(pair) != 2 {
			return 0, errors.New("invalid interleaved transport")
		}
		channel, err := strconv.Atoi(pair[0])
		if err != nil || channel < 0 || channel > 255 {
			return 0, errors.New("invalid interleaved rtp channel")
		}
		return byte(channel), nil
	}
	return 0, errors.New("transport is not interleaved tcp")
}

func readBoundedLine(r *bufio.Reader, max int) (string, error) {
	line, err := r.ReadSlice('\n')
	if errors.Is(err, bufio.ErrBufferFull) || len(line) > max {
		return "", errors.New("rtsp line exceeds size limit")
	}
	if err != nil {
		return "", err
	}
	return string(line), nil
}

func readBoundedMIMEHeader(r *bufio.Reader) (textproto.MIMEHeader, error) {
	header := make(textproto.MIMEHeader)
	total := 0
	for count := 0; count < 128; count++ {
		line, err := readBoundedLine(r, 8192)
		if err != nil {
			return nil, err
		}
		total += len(line)
		if total > 64<<10 {
			return nil, errors.New("rtsp headers exceed size limit")
		}
		trimmed := strings.TrimRight(line, "\r\n")
		if trimmed == "" {
			return header, nil
		}
		if strings.HasPrefix(trimmed, " ") || strings.HasPrefix(trimmed, "\t") {
			return nil, errors.New("rtsp folded headers are not supported")
		}
		colon := strings.IndexByte(trimmed, ':')
		if colon <= 0 {
			return nil, errors.New("invalid rtsp header")
		}
		key := textproto.CanonicalMIMEHeaderKey(strings.TrimSpace(trimmed[:colon]))
		if key == "" {
			return nil, errors.New("invalid rtsp header name")
		}
		value := strings.TrimSpace(trimmed[colon+1:])
		header.Add(key, value)
	}
	return nil, errors.New("too many rtsp headers")
}

func isTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}
