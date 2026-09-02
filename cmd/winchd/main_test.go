package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/shaielc/code-winch/internal/adapters/transport/httpapi"
	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/internal/platform/telemetry"
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
