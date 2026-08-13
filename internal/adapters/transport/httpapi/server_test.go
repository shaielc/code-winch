package httpapi

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	testToken = "local-auth-token-with-at-least-32-bytes"
	testCSRF  = "local-csrf-token-with-at-least-32-bytes"
)

type backendStub struct {
	actor string
	input RunInputRequest
}

func sampleRun() Run {
	return Run{Id: "01J00000000000000000000000", State: Created, Version: 1, CreatedAt: time.Unix(1, 0), UpdatedAt: time.Unix(1, 0), WorkspacePath: "/safe", HarnessProfile: "fake", SandboxProfile: "local"}
}
func (b *backendStub) CreateRun(_ context.Context, actor, _ string, _ CreateRunRequest) (Run, error) {
	b.actor = actor
	return sampleRun(), nil
}
func (b *backendStub) GetRun(_ context.Context, actor string, _ RunId) (Run, error) {
	b.actor = actor
	return sampleRun(), nil
}
func (b *backendStub) StartRun(_ context.Context, actor string, _ RunId, _ string, _ int64) (Run, error) {
	b.actor = actor
	return sampleRun(), nil
}
func (b *backendStub) StopRun(_ context.Context, actor string, _ RunId, _ string, _ int64, _ StopRunRequest) (Run, error) {
	b.actor = actor
	return sampleRun(), nil
}
func (b *backendStub) ListRunEvents(_ context.Context, actor string, _ RunId, _ int64, _ int) (EventPage, error) {
	b.actor = actor
	return EventPage{Events: []Event{}}, nil
}
func (b *backendStub) SendRunInput(_ context.Context, actor string, _ RunId, _ string, _ int64, input RunInputRequest) (InputAccepted, error) {
	b.actor, b.input = actor, input
	return InputAccepted{Accepted: true, CommandId: "command", Kind: InputAcceptedKindText, RunId: sampleRun().Id}, nil
}

func newTestHandler(t *testing.T, backend Backend, log io.Writer, limit int64) http.Handler {
	t.Helper()
	h, err := NewHandler(Config{Token: testToken, CSRFToken: testCSRF, AllowedOrigin: "https://localhost:8443", Actor: "local-user", BodyLimit: limit, Logger: slog.New(slog.NewTextHandler(log, nil)), RequestID: func() string { return "request-1" }}, backend)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func request(method, path, body string) *http.Request {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Authorization", "Bearer "+testToken)
	r.Header.Set("Origin", "https://localhost:8443")
	r.Header.Set(csrfHeader, testCSRF)
	return r
}

func TestAuthenticationCSRFAndObjectActor(t *testing.T) {
	b := &backendStub{}
	h := newTestHandler(t, b, io.Discard, 1024)

	unauth := httptest.NewRecorder()
	h.ServeHTTP(unauth, httptest.NewRequest(http.MethodGet, "/api/v1/runs/01J00000000000000000000000", nil))
	if unauth.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d", unauth.Code)
	}

	noCSRF := httptest.NewRecorder()
	r := request(http.MethodPost, "/api/v1/runs/01J00000000000000000000000/start", "")
	r.Header.Del(csrfHeader)
	h.ServeHTTP(noCSRF, r)
	if noCSRF.Code != http.StatusForbidden {
		t.Fatalf("missing CSRF status = %d", noCSRF.Code)
	}

	read := httptest.NewRecorder()
	h.ServeHTTP(read, request(http.MethodGet, "/api/v1/runs/01J00000000000000000000000", ""))
	if read.Code != http.StatusOK || b.actor != "local-user" {
		t.Fatalf("read status=%d actor=%q", read.Code, b.actor)
	}
}

func TestInputIsBoundedAndContentNeverLogged(t *testing.T) {
	b := &backendStub{}
	var logs bytes.Buffer
	h := newTestHandler(t, b, &logs, 256)
	path := "/api/v1/runs/01J00000000000000000000000/input"

	large := httptest.NewRecorder()
	h.ServeHTTP(large, request(http.MethodPost, path, `{"kind":"text","text":"`+strings.Repeat("x", 300)+`"}`))
	if large.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("large payload status = %d", large.Code)
	}

	const canary = "SECRET-CONTENT-CANARY"
	ok := httptest.NewRecorder()
	r := request(http.MethodPost, path, `{"kind":"text","text":"`+canary+`"}`)
	r.Header.Set("Idempotency-Key", "input-1")
	r.Header.Set("If-Match", `"1"`)
	h.ServeHTTP(ok, r)
	if ok.Code != http.StatusAccepted || b.input.Text == nil || *b.input.Text != canary {
		t.Fatalf("input status=%d body=%s", ok.Code, ok.Body.String())
	}
	if strings.Contains(logs.String(), canary) {
		t.Fatalf("logs contained request content: %s", logs.String())
	}
	if !strings.Contains(logs.String(), "run_id=01J00000000000000000000000") {
		t.Fatalf("logs omitted run ID: %s", logs.String())
	}
}

func TestSessionCookieSecurityAttributes(t *testing.T) {
	w := httptest.NewRecorder()
	SetSessionCookie(w, testToken)
	cookie := w.Result().Cookies()[0]
	if !cookie.Secure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode || cookie.Path != "/api/v1" {
		t.Fatalf("insecure cookie: %#v", cookie)
	}
}

func TestConfigurationRejectsWeakSecrets(t *testing.T) {
	_, err := NewHandler(Config{Token: "short", CSRFToken: "short", AllowedOrigin: "http://localhost", Actor: "user", RequestID: func() string { return "id" }}, &backendStub{})
	if err == nil {
		t.Fatal("NewHandler accepted weak local secrets")
	}
}
