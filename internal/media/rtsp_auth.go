package media

import (
	"crypto/md5" // #nosec G501 -- RTSP Digest MD5 is an interoperability requirement for legacy cameras.
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

type rtspAuth struct {
	scheme    string
	realm     string
	nonce     string
	opaque    string
	algorithm string
	qop       string
	username  string
	password  string
	nonceCnt  uint32
}

func selectAuthChallenge(values []string, secureTLS bool) (*rtspAuth, error) {
	var basic *rtspAuth
	for _, value := range values {
		challenge, err := parseAuthChallenge(value)
		if err != nil {
			continue
		}
		if challenge.scheme == "digest" {
			if err := challenge.validateDigest(); err == nil {
				return challenge, nil
			}
		}
		if challenge.scheme == "basic" {
			basic = challenge
		}
	}
	if basic != nil {
		if !secureTLS {
			return nil, &ProbeError{Code: ReasonRTSPBasicInsecure}
		}
		return basic, nil
	}
	return nil, &ProbeError{Code: ReasonRTSPAuthUnsupported}
}

func parseAuthChallenge(value string) (*rtspAuth, error) {
	value = strings.TrimSpace(value)
	space := strings.IndexByte(value, ' ')
	if space <= 0 {
		return nil, errors.New("invalid auth challenge")
	}
	scheme := strings.ToLower(strings.TrimSpace(value[:space]))
	params, err := parseAuthParams(value[space+1:])
	if err != nil {
		return nil, err
	}
	a := &rtspAuth{scheme: scheme, realm: params["realm"], nonce: params["nonce"], opaque: params["opaque"], algorithm: params["algorithm"], qop: params["qop"]}
	for _, value := range []string{a.realm, a.nonce, a.opaque, a.algorithm, a.qop} {
		if len(value) > 8192 || strings.ContainsAny(value, "\r\n\x00") {
			return nil, errors.New("invalid auth challenge value")
		}
	}
	if scheme == "basic" {
		if a.realm == "" {
			return nil, errors.New("basic realm missing")
		}
		return a, nil
	}
	if scheme != "digest" {
		return nil, errors.New("unsupported auth scheme")
	}
	if a.algorithm == "" {
		a.algorithm = "MD5"
	}
	return a, nil
}

func parseAuthParams(value string) (map[string]string, error) {
	result := map[string]string{}
	for i := 0; i < len(value); {
		for i < len(value) && (value[i] == ' ' || value[i] == '\t' || value[i] == ',') {
			i++
		}
		if i >= len(value) {
			break
		}
		start := i
		for i < len(value) && value[i] != '=' && value[i] != ',' {
			i++
		}
		if i >= len(value) || value[i] != '=' {
			return nil, errors.New("invalid auth parameter")
		}
		key := strings.ToLower(strings.TrimSpace(value[start:i]))
		i++
		if key == "" {
			return nil, errors.New("invalid auth parameter name")
		}
		var val string
		if i < len(value) && value[i] == '"' {
			i++
			var b strings.Builder
			for i < len(value) {
				if value[i] == '\\' && i+1 < len(value) {
					i++
					b.WriteByte(value[i])
					i++
					continue
				}
				if value[i] == '"' {
					i++
					break
				}
				b.WriteByte(value[i])
				i++
			}
			val = b.String()
		} else {
			start = i
			for i < len(value) && value[i] != ',' {
				i++
			}
			val = strings.TrimSpace(value[start:i])
		}
		result[key] = val
	}
	return result, nil
}

func (a *rtspAuth) validateDigest() error {
	if a.realm == "" || a.nonce == "" {
		return errors.New("digest realm or nonce missing")
	}
	algorithm := strings.ToUpper(a.algorithm)
	switch algorithm {
	case "MD5", "MD5-SESS", "SHA-256", "SHA-256-SESS":
	default:
		return errors.New("unsupported digest algorithm")
	}
	if a.qop != "" {
		foundAuth := false
		for _, q := range strings.Split(a.qop, ",") {
			if strings.EqualFold(strings.TrimSpace(q), "auth") {
				foundAuth = true
			}
		}
		if !foundAuth {
			return errors.New("unsupported digest qop")
		}
		a.qop = "auth"
	}
	return nil
}

func (a *rtspAuth) authorization(method, uri string) (string, error) {
	if a.scheme == "basic" {
		return "Basic " + base64.StdEncoding.EncodeToString([]byte(a.username+":"+a.password)), nil
	}
	if a.scheme != "digest" {
		return "", errors.New("auth scheme not initialized")
	}
	if err := a.validateDigest(); err != nil {
		return "", err
	}
	algorithm := strings.ToUpper(a.algorithm)
	baseAlgorithm := strings.TrimSuffix(algorithm, "-SESS")
	hash := func(value string) (string, error) {
		switch baseAlgorithm {
		case "MD5":
			sum := md5.Sum([]byte(value)) // #nosec G401 -- required for RTSP Digest MD5 interoperability.
			return hex.EncodeToString(sum[:]), nil
		case "SHA-256":
			sum := sha256.Sum256([]byte(value))
			return hex.EncodeToString(sum[:]), nil
		default:
			return "", errors.New("unsupported digest algorithm")
		}
	}
	cnonceBytes := make([]byte, 16)
	if _, err := rand.Read(cnonceBytes); err != nil {
		return "", err
	}
	cnonce := hex.EncodeToString(cnonceBytes)
	ha1, _ := hash(a.username + ":" + a.realm + ":" + a.password)
	if strings.HasSuffix(algorithm, "-SESS") {
		ha1, _ = hash(ha1 + ":" + a.nonce + ":" + cnonce)
	}
	ha2, _ := hash(method + ":" + uri)
	a.nonceCnt++
	nc := fmt.Sprintf("%08x", a.nonceCnt)
	var response string
	if a.qop == "auth" {
		response, _ = hash(ha1 + ":" + a.nonce + ":" + nc + ":" + cnonce + ":auth:" + ha2)
	} else {
		response, _ = hash(ha1 + ":" + a.nonce + ":" + ha2)
	}
	quote := func(v string) string { return strings.ReplaceAll(strings.ReplaceAll(v, "\\", "\\\\"), "\"", "\\\"") }
	parts := []string{fmt.Sprintf("username=\"%s\"", quote(a.username)), fmt.Sprintf("realm=\"%s\"", quote(a.realm)), fmt.Sprintf("nonce=\"%s\"", quote(a.nonce)), fmt.Sprintf("uri=\"%s\"", quote(uri)), fmt.Sprintf("response=\"%s\"", response), fmt.Sprintf("algorithm=%s", algorithm)}
	if a.opaque != "" {
		parts = append(parts, fmt.Sprintf("opaque=\"%s\"", quote(a.opaque)))
	}
	if a.qop == "auth" {
		parts = append(parts, "qop=auth", "nc="+nc, fmt.Sprintf("cnonce=\"%s\"", cnonce))
	}
	return "Digest " + strings.Join(parts, ", "), nil
}
