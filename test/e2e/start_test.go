package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestCreateStartPollEventsThenGet is the standing scenario: a created run is
// started, the fake harness plays a transcript and exits on its own, the
// durable event page reports what it produced, and the run reads back terminal.
func TestCreateStartPollEventsThenGet(t *testing.T) {
	database := requireDatabase(t)
	transcript := writeTranscript(t, "echo hello from transcript", "exit")
	base := startDaemon(t, database, "WINCH_FAKE_HARNESS_TRANSCRIPT="+transcript)

	run := createRun(t, base, "start-1", "/tmp/ws", "fake", "local")
	if run["state"] != "created" {
		t.Fatalf("created run is not in the created state: %v", run["state"])
	}
	runID := run["id"].(string)

	started := request(t, http.MethodPost, base+"/api/v1/runs/"+runID+"/start", nil, map[string]string{
		"Idempotency-Key": "start-run-1",
		"If-Match":        etag(run),
	})
	if started.StatusCode != http.StatusAccepted {
		t.Fatalf("start status=%d body=%s", started.StatusCode, started.Body)
	}
	var accepted map[string]any
	if err := json.Unmarshal(started.Body, &accepted); err != nil {
		t.Fatal(err)
	}
	if accepted["state"] != "running" {
		t.Fatalf("start did not report a running run: %s", started.Body)
	}

	final := awaitTerminal(t, base, runID)
	if final["state"] != "completed" {
		t.Fatalf("run did not complete: %v", final["state"])
	}

	events := pollEvents(t, base, runID)
	assertGapFree(t, events)
	if kind := events[0]["kind"]; kind != "run.lifecycle" {
		t.Fatalf("first event is not a lifecycle event: %v", kind)
	}
	last := events[len(events)-1]
	if last["kind"] != "run.lifecycle" || last["payload"].(map[string]any)["state"] != "completed" {
		t.Fatalf("last event is not the completed lifecycle event: %v", last)
	}
	if !containsOutput(events, "hello from transcript") {
		t.Fatalf("transcript output is missing from the event history: %v", events)
	}
	// lastSequence is the run's own view of its history and must agree with it.
	if final["lastSequence"] != float64(len(events)) {
		t.Fatalf("lastSequence=%v does not match %d events", final["lastSequence"], len(events))
	}

	// Publish intent exists for every appended event, and nothing has drained it:
	// this daemon starts no outbox worker.
	if pending := outboxBacklog(t, database, events); pending != len(events) {
		t.Fatalf("outbox backlog=%d for %d events", pending, len(events))
	}
}

// TestStartRefusesUnsupportedProfilePair proves the pair is checked before any
// driver is touched: the run keeps its created state and produces no events, so
// nothing was prepared, launched, or run unisolated.
func TestStartRefusesUnsupportedProfilePair(t *testing.T) {
	database := requireDatabase(t)
	base := startDaemon(t, database)

	run := createRun(t, base, "unsupported-1", "/tmp/ws", "codex", "docker")
	refused := request(t, http.MethodPost, base+"/api/v1/runs/"+run["id"].(string)+"/start", nil, map[string]string{
		"Idempotency-Key": "start-unsupported-1",
		"If-Match":        etag(run),
	})
	if refused.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("start status=%d body=%s", refused.StatusCode, refused.Body)
	}
	if !bytes.Contains(refused.Body, []byte(`"code":"unsupported_profile"`)) {
		t.Fatalf("refusal is not the stable unsupported_profile problem: %s", refused.Body)
	}
	// A content-free refusal names neither profile back to the caller.
	if bytes.Contains(refused.Body, []byte("codex")) || bytes.Contains(refused.Body, []byte("docker")) {
		t.Fatalf("refusal echoed the rejected profiles: %s", refused.Body)
	}
	after := readRun(t, base, run["id"].(string))
	if after["state"] != "created" {
		t.Fatalf("refused run left the created state: %v", after["state"])
	}
	if events := pageEvents(t, base, run["id"].(string), 0); len(events) != 0 {
		t.Fatalf("refused run produced events: %v", events)
	}
}

// TestStartRequiresCurrentVersion covers the conditional command: a stale ETag
// is refused, and a replayed start is refused because the run already left the
// only state it can be started from.
func TestStartRequiresCurrentVersion(t *testing.T) {
	database := requireDatabase(t)
	base := startDaemon(t, database, "WINCH_FAKE_HARNESS_TRANSCRIPT="+writeTranscript(t, "exit"))

	run := createRun(t, base, "conditional-1", "/tmp/ws", "fake", "local")
	runID := run["id"].(string)
	stale := request(t, http.MethodPost, base+"/api/v1/runs/"+runID+"/start", nil, map[string]string{
		"Idempotency-Key": "start-stale-1",
		"If-Match":        `"9999"`,
	})
	if stale.StatusCode != http.StatusPreconditionFailed {
		t.Fatalf("stale start status=%d body=%s", stale.StatusCode, stale.Body)
	}
	accepted := request(t, http.MethodPost, base+"/api/v1/runs/"+runID+"/start", nil, map[string]string{
		"Idempotency-Key": "start-conditional-1",
		"If-Match":        etag(run),
	})
	if accepted.StatusCode != http.StatusAccepted {
		t.Fatalf("start status=%d body=%s", accepted.StatusCode, accepted.Body)
	}
	awaitTerminal(t, base, runID)
	current := readRun(t, base, runID)
	replay := request(t, http.MethodPost, base+"/api/v1/runs/"+runID+"/start", nil, map[string]string{
		"Idempotency-Key": "start-conditional-1",
		"If-Match":        etag(current),
	})
	if replay.StatusCode != http.StatusConflict {
		t.Fatalf("replayed start status=%d body=%s", replay.StatusCode, replay.Body)
	}
}

func requireDatabase(t *testing.T) string {
	t.Helper()
	database := os.Getenv("PG_TEST_DATABASE_URL")
	if database == "" {
		t.Skip("PG_TEST_DATABASE_URL is required")
	}
	return database
}

// startDaemon runs the real binary against the real database and returns its
// base URL. The fake harness binary is taken from the same directory as the
// daemon, which is where `make build` puts both.
func startDaemon(t *testing.T, database string, env ...string) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	_ = listener.Close()
	binary := os.Getenv("WINCHD_BIN")
	if binary == "" {
		binary = "../../bin/winchd"
	}
	absolute, err := filepath.Abs(binary)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	command := exec.CommandContext(ctx, absolute)
	command.Env = append(os.Environ(),
		"WINCH_ADDR="+address,
		"WINCH_DATABASE_URL="+database,
		"WINCH_ALLOWED_ORIGIN=http://"+address,
		"WINCH_TOKEN="+secret,
		"WINCH_CSRF_TOKEN="+secret,
		"WINCH_STATIC_DIR="+t.TempDir(),
		"WINCH_FAKE_HARNESS_BINARY="+filepath.Join(filepath.Dir(absolute), "fake-harness"),
	)
	command.Env = append(command.Env, env...)
	var logs bytes.Buffer
	command.Stdout, command.Stderr = &logs, &logs
	if err = command.Start(); err != nil {
		cancel()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cancel()
		_ = command.Wait()
	})
	base := "http://" + address
	waitHealthy(t, base, &logs)
	return base
}

func writeTranscript(t *testing.T, lines ...string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "transcript.txt")
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func request(t *testing.T, method, url string, body []byte, headers map[string]string) response {
	t.Helper()
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+secret)
	if method != http.MethodGet {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", secret)
		req.Header.Set("Origin", origin(url))
	}
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = r.Body.Close() }()
	var out bytes.Buffer
	if _, err = out.ReadFrom(r.Body); err != nil {
		t.Fatal(err)
	}
	return response{r.StatusCode, out.Bytes()}
}

func origin(url string) string {
	rest := strings.TrimPrefix(url, "http://")
	host, _, _ := strings.Cut(rest, "/")
	return "http://" + host
}

// createRun mints a fresh idempotency key per call. Creation deduplicates per
// (actor, key) for the life of the run it created, so a fixed key would replay
// the previous suite run's finished run instead of creating one to start.
func createRun(t *testing.T, base, name, workspace, harness, sandbox string) map[string]any {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"workspacePath": workspace, "harnessProfile": harness, "sandboxProfile": sandbox})
	key := fmt.Sprintf("%s-%d", name, time.Now().UnixNano())
	created := request(t, http.MethodPost, base+"/api/v1/runs", body, map[string]string{"Idempotency-Key": key})
	if created.StatusCode != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", created.StatusCode, created.Body)
	}
	var run map[string]any
	if err := json.Unmarshal(created.Body, &run); err != nil {
		t.Fatal(err)
	}
	return run
}

func readRun(t *testing.T, base, runID string) map[string]any {
	t.Helper()
	read := request(t, http.MethodGet, base+"/api/v1/runs/"+runID, nil, nil)
	if read.StatusCode != http.StatusOK {
		t.Fatalf("get status=%d body=%s", read.StatusCode, read.Body)
	}
	var run map[string]any
	if err := json.Unmarshal(read.Body, &run); err != nil {
		t.Fatal(err)
	}
	return run
}

func etag(run map[string]any) string {
	return fmt.Sprintf("\"%d\"", int64(run["version"].(float64)))
}

// awaitTerminal polls the run itself rather than watching the harness: the run
// ends because its transcript ends, so nothing here asks it to stop.
func awaitTerminal(t *testing.T, base, runID string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	var run map[string]any
	for time.Now().Before(deadline) {
		run = readRun(t, base, runID)
		switch run["state"] {
		case "completed", "failed", "cancelled":
			return run
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("run did not reach a terminal state: %v", run)
	return nil
}

func pageEvents(t *testing.T, base, runID string, after int64) []map[string]any {
	t.Helper()
	read := request(t, http.MethodGet, fmt.Sprintf("%s/api/v1/runs/%s/events?after_sequence=%d&limit=%d", base, runID, after, 2), nil, nil)
	if read.StatusCode != http.StatusOK {
		t.Fatalf("events status=%d body=%s", read.StatusCode, read.Body)
	}
	var page struct {
		Events            []map[string]any `json:"events"`
		NextAfterSequence int64            `json:"nextAfterSequence"`
		HasMore           bool             `json:"hasMore"`
	}
	if err := json.Unmarshal(read.Body, &page); err != nil {
		t.Fatal(err)
	}
	if page.NextAfterSequence < after {
		t.Fatalf("cursor moved backwards: %d < %d", page.NextAfterSequence, after)
	}
	return page.Events
}

// pollEvents walks the run's whole history through the polling endpoint with a
// page size small enough that paging is actually exercised.
func pollEvents(t *testing.T, base, runID string) []map[string]any {
	t.Helper()
	var all []map[string]any
	for cursor := int64(0); ; {
		page := pageEvents(t, base, runID, cursor)
		if len(page) == 0 {
			break
		}
		all = append(all, page...)
		cursor = int64(page[len(page)-1]["sequence"].(float64))
	}
	if len(all) == 0 {
		t.Fatal("the run produced no events")
	}
	return all
}

func assertGapFree(t *testing.T, events []map[string]any) {
	t.Helper()
	for i, event := range events {
		if got := int64(event["sequence"].(float64)); got != int64(i+1) {
			t.Fatalf("event %d has sequence %d; the page is not gap-free and ordered", i, got)
		}
	}
}

func containsOutput(events []map[string]any, want string) bool {
	for _, event := range events {
		payload, ok := event["payload"].(map[string]any)
		if !ok {
			continue
		}
		if data, ok := payload["data"].(string); ok && strings.Contains(data, want) {
			return true
		}
	}
	return false
}

// outboxBacklog counts undelivered publish intent for the given events straight
// from the table: no worker drains it and no metric exposes it yet. Events are
// matched by ID because an outbox record carries the event's own identifier.
func outboxBacklog(t *testing.T, database string, events []map[string]any) int {
	t.Helper()
	ids := make([]string, 0, len(events))
	for _, event := range events {
		ids = append(ids, event["eventId"].(string))
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, database)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	var count int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM outbox WHERE topic='run.events' AND completed_at IS NULL AND poisoned_at IS NULL AND id::text = ANY($1)`, ids).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}
