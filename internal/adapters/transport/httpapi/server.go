package httpapi

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBodyLimit = int64(128 << 10)
	sessionCookie    = "winch_session"
	csrfHeader       = "X-CSRF-Token"
)

var (
	ErrRunNotFound         = errors.New("run not found")
	ErrStateConflict       = errors.New("run state conflict")
	ErrIdempotencyConflict = errors.New("idempotency conflict")
	ErrPreconditionFailed  = errors.New("precondition failed")
	ErrValidation          = errors.New("validation failed")
)

// Backend is the inward application boundary used by the HTTP adapter. Actor is
// supplied on every call so implementations authorize the object, rather than
// treating knowledge of a run ID as authorization.
type Backend interface {
	CreateRun(context.Context, string, string, CreateRunRequest) (Run, error)
	GetRun(context.Context, string, RunId) (Run, error)
	StartRun(context.Context, string, RunId, string, int64) (Run, error)
	StopRun(context.Context, string, RunId, string, int64, StopRunRequest) (Run, error)
	ListRunEvents(context.Context, string, RunId, int64, int) (EventPage, error)
	SendRunInput(context.Context, string, RunId, string, int64, RunInputRequest) (InputAccepted, error)
}

type Config struct {
	// Token is a deployment-local secret used for both bearer and cookie auth.
	Token               string
	CSRFToken           string
	AllowedOrigin       string
	Actor               string
	BodyLimit           int64
	Logger              *slog.Logger
	RequestID           func() string
	EventStream         *EventStream
	HeartbeatInterval   time.Duration
	ReauthorizeInterval time.Duration
	StreamWriteTimeout  time.Duration
}

type server struct {
	backend Backend
	cfg     Config
}

// NewHandler constructs an authenticated API rooted at /api/v1.
func NewHandler(cfg Config, backend Backend) (http.Handler, error) {
	if backend == nil || len(cfg.Token) < 32 || len(cfg.CSRFToken) < 32 || cfg.Actor == "" || cfg.RequestID == nil {
		return nil, errors.New("httpapi: backend, actor, request ID source, and 32-byte authentication secrets are required")
	}
	origin, err := url.Parse(cfg.AllowedOrigin)
	if err != nil || origin.Scheme == "" || origin.Host == "" || origin.Path != "" {
		return nil, errors.New("httpapi: allowed origin must be an absolute origin without a path")
	}
	if cfg.BodyLimit <= 0 {
		cfg.BodyLimit = defaultBodyLimit
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	s := &server{backend: backend, cfg: cfg}
	generated := HandlerWithOptions(s, ChiServerOptions{
		BaseURL: "/api/v1",
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, _ error) {
			s.problem(w, r, http.StatusBadRequest, "bad_request", "Bad request", "The request parameters are invalid.")
		},
	})
	root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const prefix, suffix = "/api/v1/runs/", "/events/stream"
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, prefix) && strings.HasSuffix(r.URL.Path, suffix) {
			runID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, prefix), suffix)
			if runID != "" && !strings.Contains(runID, "/") {
				s.streamEvents(w, r, RunId(runID))
				return
			}
		}
		generated.ServeHTTP(w, r)
	})
	return s.security(root), nil
}

// SetSessionCookie establishes the secure local browser session. The CSRF
// token is intentionally not placed in a cookie and must be delivered through
// trusted local bootstrap configuration.
func SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: token, Path: "/api/v1", Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode})
}

type requestKey struct{}
type identity struct{ requestID, actor string }

func (s *server) security(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := identity{requestID: s.cfg.RequestID()}
		r = r.WithContext(context.WithValue(r.Context(), requestKey{}, id))
		w.Header().Set("X-Request-Id", id.requestID)
		if r.URL.Path == "/api/v1/health" {
			next.ServeHTTP(w, r)
			return
		}
		if !s.authenticated(r) {
			s.problem(w, r, http.StatusUnauthorized, "unauthorized", "Unauthorized", "Valid authentication is required.")
			return
		}
		id.actor = s.cfg.Actor
		r = r.WithContext(context.WithValue(r.Context(), requestKey{}, id))
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			if !equal(r.Header.Get(csrfHeader), s.cfg.CSRFToken) || r.Header.Get("Origin") != s.cfg.AllowedOrigin {
				s.problem(w, r, http.StatusForbidden, "csrf_failed", "Forbidden", "The CSRF token or request origin is invalid.")
				return
			}
			if r.ContentLength > s.cfg.BodyLimit {
				s.problem(w, r, http.StatusRequestEntityTooLarge, "payload_too_large", "Payload too large", "The request body exceeds the configured limit.")
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, s.cfg.BodyLimit)
		}
		next.ServeHTTP(w, r)
	})
}

func (s *server) authenticated(r *http.Request) bool {
	bearer := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if bearer != r.Header.Get("Authorization") && equal(bearer, s.cfg.Token) {
		return true
	}
	cookie, err := r.Cookie(sessionCookie)
	return err == nil && equal(cookie.Value, s.cfg.Token)
}

func equal(a, b string) bool {
	return len(a) == len(b) && subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func (s *server) GetHealth(w http.ResponseWriter, _ *http.Request) {
	s.writeJSON(w, http.StatusOK, HealthResponse{Status: Ok})
}

func (s *server) CreateRun(w http.ResponseWriter, r *http.Request, p CreateRunParams) {
	var body CreateRunRequest
	if !decode(r, &body) {
		s.problem(w, r, http.StatusBadRequest, "bad_request", "Bad request", "The request body is not valid JSON.")
		return
	}
	if body.WorkspacePath == "" || body.HarnessProfile == "" || body.SandboxProfile == "" {
		s.problem(w, r, http.StatusUnprocessableEntity, "validation_failed", "Validation failed", "Required run fields are missing.")
		return
	}
	result, err := s.backend.CreateRun(r.Context(), actor(r), p.IdempotencyKey, body)
	if err != nil {
		s.backendProblem(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/v1/runs/"+result.Id)
	s.runResponse(w, http.StatusCreated, p.IdempotencyKey, result)
}

func (s *server) GetRun(w http.ResponseWriter, r *http.Request, id RunId) {
	result, err := s.backend.GetRun(r.Context(), actor(r), id)
	if err != nil {
		s.backendProblem(w, r, err)
		return
	}
	s.runResponse(w, http.StatusOK, "", result)
}

func (s *server) StartRun(w http.ResponseWriter, r *http.Request, id RunId, p StartRunParams) {
	version, ok := etagVersion(p.IfMatch)
	if !ok {
		s.problem(w, r, http.StatusPreconditionRequired, "precondition_required", "Precondition required", "Supply the run's current ETag in If-Match.")
		return
	}
	result, err := s.backend.StartRun(r.Context(), actor(r), id, p.IdempotencyKey, version)
	if err != nil {
		s.backendProblem(w, r, err)
		return
	}
	s.audit(r, "run.start", id)
	s.runResponse(w, http.StatusAccepted, p.IdempotencyKey, result)
}

func (s *server) StopRun(w http.ResponseWriter, r *http.Request, id RunId, p StopRunParams) {
	version, ok := etagVersion(p.IfMatch)
	if !ok {
		s.problem(w, r, http.StatusPreconditionRequired, "precondition_required", "Precondition required", "Supply the run's current ETag in If-Match.")
		return
	}
	var body StopRunRequest
	if r.ContentLength != 0 && !decode(r, &body) {
		s.problem(w, r, 400, "bad_request", "Bad request", "The request body is not valid JSON.")
		return
	}
	result, err := s.backend.StopRun(r.Context(), actor(r), id, p.IdempotencyKey, version, body)
	if err != nil {
		s.backendProblem(w, r, err)
		return
	}
	s.audit(r, "run.stop", id)
	s.runResponse(w, http.StatusAccepted, p.IdempotencyKey, result)
}

func (s *server) ListRunEvents(w http.ResponseWriter, r *http.Request, id RunId, p ListRunEventsParams) {
	after, limit := int64(0), 50
	if p.AfterSequence != nil {
		after = *p.AfterSequence
	}
	if p.Limit != nil {
		limit = *p.Limit
	}
	if after < 0 || limit < 1 || limit > 200 {
		s.problem(w, r, 400, "bad_request", "Bad request", "Paging parameters are outside the allowed range.")
		return
	}
	result, err := s.backend.ListRunEvents(r.Context(), actor(r), id, after, limit)
	if err != nil {
		s.backendProblem(w, r, err)
		return
	}
	s.writeJSON(w, http.StatusOK, result)
}

func (s *server) StreamRunEvents(w http.ResponseWriter, r *http.Request, id RunId, _ StreamRunEventsParams) {
	s.streamEvents(w, r, id)
}

func (s *server) SendRunInput(w http.ResponseWriter, r *http.Request, id RunId, p SendRunInputParams) {
	version, ok := etagVersion(p.IfMatch)
	if !ok {
		s.problem(w, r, 428, "precondition_required", "Precondition required", "Supply the run's current ETag in If-Match.")
		return
	}
	var body RunInputRequest
	if !decode(r, &body) {
		s.problem(w, r, 400, "bad_request", "Bad request", "The request body is not valid JSON.")
		return
	}
	result, err := s.backend.SendRunInput(r.Context(), actor(r), id, p.IdempotencyKey, version, body)
	if err != nil {
		s.backendProblem(w, r, err)
		return
	}
	s.audit(r, "run.input", id)
	s.writeJSON(w, http.StatusAccepted, result)
}

func decode(r *http.Request, target any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return false
	}
	return decoder.Decode(&struct{}{}) == io.EOF
}

func etagVersion(value string) (int64, bool) {
	if len(value) < 3 || value[0] != '"' || value[len(value)-1] != '"' {
		return 0, false
	}
	v, err := strconv.ParseInt(value[1:len(value)-1], 10, 64)
	return v, err == nil && v > 0
}

func actor(r *http.Request) string { return r.Context().Value(requestKey{}).(identity).actor }

func (s *server) runResponse(w http.ResponseWriter, status int, key string, run Run) {
	w.Header().Set("ETag", fmt.Sprintf("\"%d\"", run.Version))
	if key != "" {
		w.Header().Set("Idempotency-Key", key)
	}
	s.writeJSON(w, status, run)
}

func (s *server) backendProblem(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrRunNotFound):
		s.problem(w, r, 404, "run_not_found", "Run not found", "The requested run was not found.")
	case errors.Is(err, ErrStateConflict):
		s.problem(w, r, 409, "run_state_conflict", "Run state conflict", "The command is invalid for the run's current state.")
	case errors.Is(err, ErrIdempotencyConflict):
		s.problem(w, r, 409, "idempotency_conflict", "Idempotency conflict", "Use a new Idempotency-Key for a different request.")
	case errors.Is(err, ErrPreconditionFailed):
		s.problem(w, r, 412, "precondition_failed", "Precondition failed", "Read the run and retry with its current ETag.")
	case errors.Is(err, ErrValidation):
		s.problem(w, r, 422, "validation_failed", "Validation failed", "Correct the invalid fields and retry.")
	default:
		s.problem(w, r, 500, "internal_error", "Internal error", "The request could not be completed.")
	}
}

func (s *server) problem(w http.ResponseWriter, r *http.Request, status int, code, title, detail string) {
	id, _ := r.Context().Value(requestKey{}).(identity)
	s.cfg.Logger.LogAttrs(r.Context(), slog.LevelWarn, "http request rejected", slog.String("operation", r.Method), slog.String("request_id", id.requestID), slog.String("code", code), slog.Int("status", status))
	w.Header().Set("Content-Type", "application/problem+json")
	s.writeJSON(w, status, Problem{Type: "https://code-winch.dev/problems/" + strings.ReplaceAll(code, "_", "-"), Title: title, Status: status, Code: code, Detail: detail, RequestId: id.requestID})
}

func (s *server) audit(r *http.Request, operation string, runID RunId) {
	id := r.Context().Value(requestKey{}).(identity)
	s.cfg.Logger.LogAttrs(r.Context(), slog.LevelInfo, "run command accepted", slog.String("operation", operation), slog.String("request_id", id.requestID), slog.String("run_id", runID), slog.String("actor_id", id.actor))
}

func (s *server) writeJSON(w http.ResponseWriter, status int, value any) {
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
