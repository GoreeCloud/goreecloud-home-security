package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/GoreeCloud/goreecloud-home-security/internal/camera"
	"github.com/GoreeCloud/goreecloud-home-security/internal/events"
)

type Server struct {
	cameras *camera.Registry
	journal *events.Journal
}

func New(cameras *camera.Registry, journal *events.Journal) *Server {
	return &Server{cameras: cameras, journal: journal}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /readyz", s.ready)
	mux.HandleFunc("GET /api/v1/status", s.status)
	mux.HandleFunc("GET /api/v1/cameras", s.listCameras)
	mux.HandleFunc("GET /api/v1/events", s.listEvents)
	return requestMetadata(mux)
}

func requestMetadata(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := make([]byte, 16)
		if _, err := rand.Read(requestID); err != nil {
			writeError(w, http.StatusInternalServerError, "request_id_unavailable", "request could not be initialized")
			return
		}
		w.Header().Set("X-Request-ID", hex.EncodeToString(requestID))
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) ready(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (s *Server) status(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"product":       "GoreeCloud Home Security",
		"lifecycle":     "development",
		"api_version":   "v1",
		"camera_count":  s.cameras.Count(),
		"network_scope": "loopback-only",
	})
}

func (s *Server) listCameras(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"cameras": s.cameras.Snapshot()})
}

func (s *Server) listEvents(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 1000 {
			writeError(w, http.StatusBadRequest, "invalid_limit", "limit must be an integer between 1 and 1000")
			return
		}
		limit = parsed
	}
	items, err := s.journal.List(limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "event_store_unavailable", "event store could not be read")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": items})
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
