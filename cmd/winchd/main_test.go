package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/shaielc/code-winch/internal/adapters/memory"
	"github.com/shaielc/code-winch/internal/adapters/transport/httpapi"
	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/internal/platform/config"
	"github.com/shaielc/code-winch/internal/platform/telemetry"
	"github.com/shaielc/code-winch/internal/supervisor"
	"github.com/shaielc/code-winch/pkg/protocol"
)

const testSecret = "0123456789abcdef0123456789abcdef"

func TestStaticHandlerWithoutWebBuildStillLetsTheDaemonBoot(t *testing.T) {
	assets, served := staticHandler(filepath.Join(t.TempDir(), "absent"), testSecret)
	if served {
		t.Fatal("reported assets it does not have")
	}
	rec := httptest.NewRecorder()
	assets.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d, want 404", rec.Code)
	}
}

func TestAPIRunIDRoundTrip(t *testing.T) {
	id, _ := domain.ParseRunID("11111111-2222-3333-8444-555555555555")
	external := formatAPIRunID(id)
	if len(external) != 26 {
		t.Fatalf("API ID length=%d", len(external))
	}
	got, err := apiRunID(external)
	if err != nil || got != id {
		t.Fatalf("round trip got=%s err=%v", got, err)
	}
}

func TestBackendRejectsPersistedRunWithoutAttempts(t *testing.T) {
	id, _ := domain.ParseRunID("77777777-7777-7777-7777-777777777777")
	repository := &memory.RunRepository{}
	_, _ = repository.Save(context.Background(), application.RunRecord{ID: id}, 0)
	now, _ := domain.NewTimestamp(time.Now())
	service, err := application.NewRunService(repository, memory.NewClock(now), &memory.IDSource{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = (runBackend{runs: service}).GetRun(context.Background(), "actor", formatAPIRunID(id))
	if !errors.Is(err, application.ErrInvalidRunRecord) {
		t.Fatalf("malformed run error: %v", err)
	}
}

func TestStaticHandlerInjectsCSRFTokenIntoIndex(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<meta content="__WINCH_CSRF_TOKEN__">`), 0o600); err != nil {
		t.Fatal(err)
	}
	assets, served := staticHandler(dir, testSecret)
	if !served {
		t.Fatal("assets were not served")
	}
	rec := httptest.NewRecorder()
	assets.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	body := rec.Body.String()
	if !strings.Contains(body, testSecret) {
		t.Fatalf("token not injected: %s", body)
	}
	if strings.Contains(body, "__WINCH_CSRF_TOKEN__") {
		t.Fatalf("placeholder survived: %s", body)
	}
}

// The redaction allowlist and the API's log keys live in different packages, so
// a rename on either side silently drops the field rather than failing a build.
func TestRejectionLogKeepsCorrelationIDAndErrorCode(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(telemetry.NewHandler(slog.NewJSONHandler(&logs, nil)))
	api, err := httpapi.NewHandler(httpapi.Config{
		Token: testSecret, CSRFToken: testSecret, AllowedOrigin: "http://localhost:8080",
		Actor: "local-user", Logger: logger, RequestID: func() string { return "correlation-canary" },
	}, runBackend{})
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/runs", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d, want 401", rec.Code)
	}
	var line map[string]any
	if err := json.Unmarshal([]byte(strings.SplitN(logs.String(), "\n", 2)[0]), &line); err != nil {
		t.Fatalf("no log line: %v (%s)", err, logs.String())
	}
	if line["request_id"] != "correlation-canary" {
		t.Errorf("correlation ID dropped: %v", line)
	}
	if line["error_code"] != "unauthorized" {
		t.Errorf("error code dropped: %v", line)
	}
}

func TestServeReturnsWithinTheShutdownDeadline(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	stream := httpapi.NewEventStream(1)
	server := &http.Server{Addr: "127.0.0.1:0", Handler: http.NotFoundHandler(), ReadHeaderTimeout: readHeaderTimeout}
	done := make(chan error, 1)
	go func() { done <- serve(ctx, server, stream, time.Second) }()
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("serve: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("serve did not return within the drain deadline")
	}
}

func TestFakeHarnessProfileAlwaysTerminatesOnItsOwn(t *testing.T) {
	// A daemon-started run cannot answer an interactive prompt, so the profile
	// must never leave the harness reading its terminal.
	if !fakeHarnessConfig(config.FakeHarnessConfig{}).EarlyExit {
		t.Fatal("the unconfigured fake profile can block on input")
	}
	scripted := fakeHarnessConfig(config.FakeHarnessConfig{Binary: "/opt/fake-harness", Transcript: "/etc/t.txt", Delay: 5 * time.Millisecond, ForceFailure: true, MalformedLine: true})
	if !scripted.EarlyExit || scripted.Executable != "/opt/fake-harness" || scripted.Transcript != "/etc/t.txt" {
		t.Fatalf("scripted profile: %#v", scripted)
	}
	if scripted.Delay != 5*time.Millisecond || !scripted.ForceFailure || !scripted.MalformedLine {
		t.Fatalf("injections were dropped: %#v", scripted)
	}
}

func TestAPIEventCarriesTheEnvelopeWithoutReshapingIt(t *testing.T) {
	id, _ := domain.ParseRunID("11111111-2222-3333-8444-555555555555")
	external := formatAPIRunID(id)
	occurred := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	stored := protocol.Event{EventID: "88888888-8888-4888-8888-888888888888", RunID: id.String(), Sequence: 7, OccurredAt: occurred, Kind: "stream.raw", SchemaVersion: 1, Source: protocol.Source{Type: "harness", Adapter: "fake", Version: "1.0.0"}, Sensitivity: protocol.SensitivityUserContent, Payload: []byte(`{"stream":"stdout","encoding":"utf-8","data":"hi"}`), Extensions: map[string]json.RawMessage{"example.vendor/v1": json.RawMessage(`{"a":1}`)}}
	converted, err := apiEvent(external, stored)
	if err != nil {
		t.Fatal(err)
	}
	// The API's run ID is the Crockford form, never the stored UUID.
	if converted.RunId != external {
		t.Fatalf("run ID = %s", converted.RunId)
	}
	if converted.Sequence != 7 || converted.Kind != "stream.raw" || converted.SchemaVersion != 1 || !converted.OccurredAt.Equal(occurred) {
		t.Fatalf("envelope: %#v", converted)
	}
	if converted.Sensitivity != httpapi.EventSensitivity(protocol.SensitivityUserContent) {
		t.Fatalf("sensitivity = %s", converted.Sensitivity)
	}
	if converted.Payload["data"] != "hi" || converted.Source["adapter"] != "fake" {
		t.Fatalf("payload=%v source=%v", converted.Payload, converted.Source)
	}
	if converted.Extensions == nil || (*converted.Extensions)["example.vendor/v1"] == nil {
		t.Fatalf("extensions were dropped: %v", converted.Extensions)
	}
}

func TestAPIEventReportsAnUnrepresentableStoredEvent(t *testing.T) {
	id, _ := domain.ParseRunID("11111111-2222-3333-8444-555555555555")
	// A scalar payload cannot become the object the contract promises, so it is
	// reported rather than silently answered as an empty payload.
	if _, err := apiEvent(formatAPIRunID(id), protocol.Event{EventID: "e", Payload: []byte(`"just a string"`)}); err == nil {
		t.Fatal("a non-object payload was accepted")
	}
}

func TestStartProblemsAreStableAndDoNotLeakCauses(t *testing.T) {
	cases := map[error]error{
		application.ErrNotFound:           httpapi.ErrRunNotFound,
		application.ErrUnsupportedProfile: httpapi.ErrUnsupportedProfile,
		application.ErrPreconditionFailed: httpapi.ErrPreconditionFailed,
		application.ErrStateConflict:      httpapi.ErrStateConflict,
		supervisor.ErrStaleLease:          httpapi.ErrStateConflict,
	}
	for cause, want := range cases {
		if got := startProblem(fmt.Errorf("wrapped: %w", cause)); !errors.Is(got, want) {
			t.Fatalf("%v mapped to %v", cause, got)
		}
	}
	other := errors.New("database unavailable")
	if got := startProblem(other); !errors.Is(got, other) {
		t.Fatalf("an unrecognised cause was rewritten to %v", got)
	}
}
