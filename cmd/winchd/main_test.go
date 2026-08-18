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
	"github.com/shaielc/code-winch/internal/application"
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
	}, unavailableBackend{})
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

// The declaration and the emitters are written in different places, so a
// status one side adds and the other never declares would be silently dropped
// in production. validateMetrics turns that drift into a refused start.
func TestDeclaredMetricsCoverEveryStatusTheDaemonEmits(t *testing.T) {
	registry := telemetry.NewRegistry()
	declareMetrics(registry)
	if err := validateMetrics(registry); err != nil {
		t.Fatalf("a declared metric was refused: %v", err)
	}
	// The registry pre-registers the metric names, so an undeclared *label* is
	// what a missing declaration actually looks like.
	if err := validateMetrics(telemetry.NewRegistry()); err == nil {
		t.Fatal("a daemon that declared nothing was allowed to start")
	}
}

// The redaction allowlist and these metric keys live in different packages, so
// a rename on either side silently drops the field rather than failing a build.
func TestRunMetricRecordsKeepTheirBoundedFields(t *testing.T) {
	var logs bytes.Buffer
	registry := telemetry.NewRegistry()
	declareMetrics(registry)
	metrics := runMetrics{registry: registry, logger: slog.New(telemetry.NewHandler(slog.NewJSONHandler(&logs, nil)))}
	id := application.RandomIDs{}.NewRunID()

	metrics.QueueTime(context.Background(), id, domain.RunStatePreparing, 1500*time.Millisecond)
	metrics.ActiveRuns(context.Background(), map[domain.RunState]int{domain.RunStateRunning: 2, domain.RunStateStopping: 1})

	lines := decodeRecords(t, logs.String())
	if len(lines) != 3 {
		t.Fatalf("%d records emitted, want 3: %s", len(lines), logs.String())
	}
	if lines[0]["run_id"] != id.String() || lines[0]["status"] != string(domain.RunStatePreparing) || lines[0]["duration_ms"] != float64(1500) {
		t.Errorf("queue time record lost a field: %v", lines[0])
	}
	// One record per declared status, in the declared order, so a gauge reads
	// the same on every emission.
	for i, want := range []struct {
		status string
		size   float64
	}{{string(domain.RunStateRunning), 2}, {string(domain.RunStateStopping), 1}} {
		if lines[i+1]["status"] != want.status || lines[i+1]["size"] != want.size {
			t.Errorf("active runs record %d = %v, want status %s size %v", i, lines[i+1], want.status, want.size)
		}
	}
}

// A status nobody declared must not reach the log even if an emitter passes
// it: the registry is the boundary on telemetry cardinality, not a suggestion.
func TestAnUndeclaredStatusIsRefusedRatherThanEmitted(t *testing.T) {
	var logs bytes.Buffer
	registry := telemetry.NewRegistry()
	declareMetrics(registry)
	metrics := runMetrics{registry: registry, logger: slog.New(telemetry.NewHandler(slog.NewJSONHandler(&logs, nil)))}

	// Running is a real state, and not one a run can leave the queue for.
	metrics.QueueTime(context.Background(), application.RandomIDs{}.NewRunID(), domain.RunStateRunning, time.Second)

	lines := decodeRecords(t, logs.String())
	if len(lines) != 1 {
		t.Fatalf("%d records emitted, want the refusal alone: %s", len(lines), logs.String())
	}
	if lines[0]["error_code"] != "undeclared_metric" {
		t.Fatalf("the refusal was not reported: %v", lines[0])
	}
	if strings.Contains(logs.String(), string(domain.RunStateRunning)) {
		t.Fatalf("the refused label value reached the log: %s", logs.String())
	}
}

func decodeRecords(t *testing.T, out string) []map[string]any {
	t.Helper()
	var records []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("decode %q: %v", line, err)
		}
		records = append(records, record)
	}
	return records
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
