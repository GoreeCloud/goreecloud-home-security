package media

import (
	"errors"
	"net/url"
	"strings"
)

type sdpInfo struct {
	sessionControl string
	videoControl   string
}

func parseSDP(payload []byte) (sdpInfo, error) {
	if len(payload) == 0 || len(payload) > 1<<20 {
		return sdpInfo{}, errors.New("invalid sdp size")
	}
	var result sdpInfo
	media := ""
	for _, raw := range strings.Split(strings.ReplaceAll(string(payload), "\r\n", "\n"), "\n") {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "m=") {
			fields := strings.Fields(strings.TrimPrefix(line, "m="))
			if len(fields) > 0 {
				media = fields[0]
			}
			continue
		}
		if !strings.HasPrefix(line, "a=control:") {
			continue
		}
		control := strings.TrimSpace(strings.TrimPrefix(line, "a=control:"))
		if control == "" || len(control) > 8192 || strings.ContainsAny(control, "\r\n\x00") {
			return sdpInfo{}, errors.New("invalid sdp control")
		}
		if media == "" && result.sessionControl == "" {
			result.sessionControl = control
		}
		if media == "video" && result.videoControl == "" {
			result.videoControl = control
		}
	}
	if result.videoControl == "" {
		return sdpInfo{}, errors.New("sdp contains no video control")
	}
	return result, nil
}

func resolveRTSPControl(base, control string) (string, error) {
	if control == "*" {
		return base, nil
	}
	ref, err := url.Parse(control)
	if err != nil {
		return "", err
	}
	if ref.IsAbs() {
		if ref.Scheme != "rtsp" && ref.Scheme != "rtsps" {
			return "", errors.New("unsupported control scheme")
		}
		return ref.String(), nil
	}
	baseURL, err := url.Parse(base)
	if err != nil || (baseURL.Scheme != "rtsp" && baseURL.Scheme != "rtsps") || baseURL.Host == "" {
		return "", errors.New("invalid rtsp base url")
	}
	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path += "/"
	}
	return baseURL.ResolveReference(ref).String(), nil
}
