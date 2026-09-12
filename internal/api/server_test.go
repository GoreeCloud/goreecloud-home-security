package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GoreeCloud/goreecloud-home-security/internal/camera"
	"github.com/GoreeCloud/goreecloud-home-security/internal/config"
	"github.com/GoreeCloud/goreecloud-home-security/internal/events"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	j, err := events.NewJournal(filepath.Join(t.TempDir(), "events.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	r := camera.NewRegistry([]config.Camera{{
		ID: "front-door", Name: "Front Door", StreamURL: "rtsp://camera.local/live",
		UsernameEnv: "CAMERA_USER", PasswordEnv: "CAMERA_PASS", Enabled: true,
	}})
	return New(r, j)
}

func TestCameraAPIIsSanitized(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/cameras", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	for _, forbidden := range []string{"rtsp://", "CAMERA_USER", "CAMERA_PASS"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("response leaked %q: %s", forbidden, body)
		}
	}
	if rec.Header().Get("Cache-Control") != "no-store" || rec.Header().Get("X-Request-ID") == "" {
		t.Fatalf("missing privacy/request headers: %#v", rec.Header())
	}
}

func TestInvalidEventLimitUsesStructuredError(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events?limit=all", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
	var body map[string]map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["error"]["code"] != "invalid_limit" {
		t.Fatalf("unexpected error: %#v", body)
	}
}

func TestStatusIncludesSanitizedSessionCounts(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/status", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "sessions_running") || strings.Contains(body, "rtsp://") {
		t.Fatalf("status body = %s", body)
	}
}
