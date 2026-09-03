package httpserver

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type DatabasePinger interface {
	Ping(context.Context) error
}

type Server struct {
	logger *slog.Logger
	db     DatabasePinger
	mux    *http.ServeMux
}

func New(logger *slog.Logger, db DatabasePinger) *Server {
	s := &Server{
		logger: logger,
		db:     db,
		mux:    http.NewServeMux(),
	}

	s.routes()

	return s
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.healthz)
	s.mux.HandleFunc("GET /readyz", s.readyz)
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "governance-api",
	})
}

func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if s.db == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status":   "not_ready",
			"database": "unavailable",
		})
		return
	}

	if err := s.db.Ping(ctx); err != nil {
		s.logger.Warn("database readiness check failed", "error", err)

		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status":   "not_ready",
			"database": "unavailable",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":   "ready",
		"database": "ok",
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(payload)
}
