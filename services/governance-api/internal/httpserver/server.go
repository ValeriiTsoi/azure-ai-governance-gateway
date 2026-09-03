package httpserver

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Server struct {
	logger *slog.Logger
	mux    *http.ServeMux
}

func New(logger *slog.Logger) *Server {
	s := &Server{
		logger: logger,
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
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "governance-api",
	}); err != nil {
		s.logger.Error("failed to encode health response", "error", err)
	}
}
