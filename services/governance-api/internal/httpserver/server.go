package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"governance-api/internal/governance"
)

type DatabasePinger interface {
	Ping(context.Context) error
}

type GovernanceService interface {
	CreateRequest(
		context.Context,
		governance.CreateRequestInput,
	) (governance.Request, error)

	GetRequest(
		context.Context,
		string,
	) (governance.Request, error)
}

type Server struct {
	logger     *slog.Logger
	db         DatabasePinger
	governance GovernanceService
	mux        *http.ServeMux
}

func New(
	logger *slog.Logger,
	db DatabasePinger,
	governanceService GovernanceService,
) *Server {
	s := &Server{
		logger:     logger,
		db:         db,
		governance: governanceService,
		mux:        http.NewServeMux(),
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

	s.mux.HandleFunc(
		"POST /v1/governance/requests",
		s.createGovernanceRequest,
	)

	s.mux.HandleFunc(
		"GET /v1/governance/requests/{requestID}",
		s.getGovernanceRequest,
	)
}

func (s *Server) healthz(
	w http.ResponseWriter,
	_ *http.Request,
) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "governance-api",
	})
}

func (s *Server) readyz(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx, cancel := context.WithTimeout(
		r.Context(),
		2*time.Second,
	)
	defer cancel()

	if s.db == nil {
		writeJSON(
			w,
			http.StatusServiceUnavailable,
			map[string]string{
				"status":   "not_ready",
				"database": "unavailable",
			},
		)
		return
	}

	if err := s.db.Ping(ctx); err != nil {
		s.logger.Warn(
			"database readiness check failed",
			"error", err,
		)

		writeJSON(
			w,
			http.StatusServiceUnavailable,
			map[string]string{
				"status":   "not_ready",
				"database": "unavailable",
			},
		)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":   "ready",
		"database": "ok",
	})
}

func (s *Server) createGovernanceRequest(
	w http.ResponseWriter,
	r *http.Request,
) {
	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		1<<20,
	)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var input governance.CreateRequestInput

	if err := decoder.Decode(&input); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid JSON request body",
		)
		return
	}

	var extra any

	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		writeError(
			w,
			http.StatusBadRequest,
			"request body must contain exactly one JSON object",
		)
		return
	}

	result, err := s.governance.CreateRequest(
		r.Context(),
		input,
	)
	if err != nil {
		s.handleGovernanceError(w, err)
		return
	}

	s.logger.Info(
		"governance request evaluated",
		"request_id", result.RequestID,
		"caller_subject", result.CallerSubject,
		"classification", result.DataClassification,
		"decision", result.Policy.Decision,
	)

	writeJSON(
		w,
		http.StatusCreated,
		result,
	)
}

func (s *Server) getGovernanceRequest(
	w http.ResponseWriter,
	r *http.Request,
) {
	requestID := r.PathValue("requestID")

	result, err := s.governance.GetRequest(
		r.Context(),
		requestID,
	)
	if err != nil {
		s.handleGovernanceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleGovernanceError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, governance.ErrInvalidInput):
		writeError(
			w,
			http.StatusBadRequest,
			err.Error(),
		)

	case errors.Is(err, governance.ErrNotFound):
		writeError(
			w,
			http.StatusNotFound,
			"governance request not found",
		)

	default:
		s.logger.Error(
			"governance request failed",
			"error", err,
		)

		writeError(
			w,
			http.StatusInternalServerError,
			"internal server error",
		)
	}
}

func writeError(
	w http.ResponseWriter,
	status int,
	message string,
) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	payload any,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(payload)
}
