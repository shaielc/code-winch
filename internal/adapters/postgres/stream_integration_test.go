//go:build integration

package postgres_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	fakeharness "github.com/shaielc/code-winch/internal/adapters/harness/fake"
	fakesandbox "github.com/shaielc/code-winch/internal/adapters/sandbox/fake"
	"github.com/shaielc/code-winch/internal/adapters/transport/httpapi"
	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/execution"
	"github.com/shaielc/code-winch/internal/runner/local"
	"github.com/shaielc/code-winch/internal/supervisor"
	"github.com/shaielc/code-winch/pkg/protocol"
)

// The daemon's delivery settings. The interval bounds how long a committed
// event waits before live subscribers are told about it, and is shortened here
// only so the test is not paced by the production tick.
const (
	outboxInterval = 10 * time.Millisecond
	outboxBatch    = 64
)

// apiSecret stands in for the deployment-local bearer and CSRF secrets, which
// the handler refuses below 32 bytes.
const apiSecret = "0123456789abcdef0123456789abcdef"

// liveEvents adapts the outbox publisher to the WebSocket broadcaster. It is
// the composition root's adapter, repeated because package main cannot be
// imported; the projection it applies is the exported one the API also uses for
// history, so a subscriber cannot see a shape the replay would not produce.
type liveEvents struct{ stream *httpapi.EventStream }

func (l liveEvents) Notify(event protocol.Event) {
	l.stream.Publish(event.RunID, httpapi.APIEvent(event))
}

// wallClock is the outbox worker's clock, which measures lease and backoff
// deadlines in wall time rather than in domain timestamps.
type wallClock struct{}

func (wallClock) Now() time.Time { return time.Now().UTC() }

// daemon is cmd/winchd's runtime over a real database and a real listener: the
// same store, supervisor, engine, runner pool, outbox worker, publisher, live
// stream, and API handler, wired in the same order. Only the process concerns —
// configuration, signals, static assets — are left out.
type daemon struct {
	server *httptest.Server
	origin string
	ids    application.RandomIDs
}

func newDaemon(t *testing.T) *daemon {
	t.Helper()
	_, store := database(t)
	ctx, cancel := context.WithCancel(context.Background())
	// Durable work outlives the request context, as it does in the daemon.
	runtimeCtx := context.WithoutCancel(ctx)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	registry := application.NewDriverRegistry()
	registry.RegisterHarness("fake", fakeharness.Driver{})
	registry.RegisterSandbox("fake", fakesandbox.New(application.SandboxCapabilities{Isolation: "in-memory", Attach: true}))

	stream := httpapi.NewEventStream(64)
	ids := application.RandomIDs{}
	clock := application.SystemClock{}
	instance := ids.NewCommandID().String()

	runners := local.NewPool(registry)
	runSupervisor := supervisor.New(store, runners, execution.ClassificationRedactor{}, clock, instance, time.Minute).WithReconciliationIDs(ids)
	runs := application.NewRunService(store, store, ids, clock, registry, nil)
	engine, err := execution.New(execution.Config{
		Runs: store, States: runs, Control: store, Supervisor: runSupervisor,
		Runner: runners, IDs: ids, Clock: clock, Logger: logger, StopGrace: time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	runs.WithExecutor(engine)
	inputs := application.NewInputService(store, execution.Capabilities{Runs: store}, nil)
	outbox, err := application.NewOutboxWorker(store, execution.Publisher{
		Events: liveEvents{stream}, Commands: store, Input: engine,
	}, wallClock{}, application.OutboxWorkerConfig{
		WorkerID: instance, LeaseToken: ids.NewCommandID().String(), BatchSize: outboxBatch,
		LeaseDuration: 30 * time.Second, BaseBackoff: 10 * time.Millisecond,
		MaxBackoff: time.Second, MaxAttempts: 8,
	})
	if err != nil {
		t.Fatal(err)
	}

	// The exact allowed origin must be known before constructing the handler.
	listener := httptest.NewUnstartedServer(nil)
	origin := "http://" + listener.Listener.Addr().String()
	api, err := httpapi.NewHandler(httpapi.Config{
		Token: apiSecret, CSRFToken: apiSecret, AllowedOrigin: origin, Actor: "local-user",
		Logger: logger, RequestID: func() string { return "integration-request" }, EventStream: stream,
		HeartbeatInterval: time.Hour, ReauthorizeInterval: time.Hour, StreamWriteTimeout: time.Second,
	}, httpapi.NewBackend(runs, inputs, ids))
	if err != nil {
		t.Fatal(err)
	}
	listener.Config.Handler = api
	listener.Start()

	var runtime sync.WaitGroup
	runtime.Add(2)
	go func() { defer runtime.Done(); engine.Consume(runtimeCtx, runners.Observations()) }()
	go func() { defer runtime.Done(); deliver(ctx, outbox) }()
	t.Cleanup(func() {
		// Shutdown order is the daemon's: subscribers are disconnected before
		// in-flight requests are drained, because a stream handler parked on a
		// subscriber would otherwise outlive the listener; and executions are
		// ended before the runner closes, because its pumps finish only once
		// their streams are released.
		stream.Close()
		listener.Close()
		cancel()
		engine.Shutdown(runtimeCtx)
		runners.Close()
		runtime.Wait()
	})
	return &daemon{server: listener, origin: origin, ids: ids}
}

// deliver drains the outbox until the daemon stops. A full batch is followed
// immediately by another claim so a burst of output is not paced by the tick.
func deliver(ctx context.Context, worker *application.OutboxWorker) {
	ticker := time.NewTicker(outboxInterval)
	defer ticker.Stop()
	for {
		for {
			delivered, err := worker.RunOnce(ctx)
			if err != nil || delivered < outboxBatch {
				break
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// call issues an authenticated request carrying the CSRF token, origin, and
// idempotency key every mutating endpoint requires.
func (d *daemon) call(t *testing.T, method, path, etag string, body any) (*http.Response, []byte) {
	t.Helper()
	var payload io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		payload = bytes.NewReader(encoded)
	}
	request, err := http.NewRequest(method, d.server.URL+path, payload)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Authorization", "Bearer "+apiSecret)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", apiSecret)
	request.Header.Set("Origin", d.origin)
	request.Header.Set("Idempotency-Key", d.ids.NewCommandID().String())
	if etag != "" {
		request.Header.Set("If-Match", etag)
	}
	response, err := d.server.Client().Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = response.Body.Close() }()
	received, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	return response, received
}

// command performs a lifecycle request and returns the run representation with
// the ETag a following command must present.
func (d *daemon) command(t *testing.T, method, path, etag string, body any, want int) (httpapi.Run, string) {
	t.Helper()
	response, payload := d.call(t, method, path, etag, body)
	if response.StatusCode != want {
		t.Fatalf("%s %s: status=%d want=%d body=%s", method, path, response.StatusCode, want, payload)
	}
	var run httpapi.Run
	if err := json.Unmarshal(payload, &run); err != nil {
		t.Fatalf("%s %s: %v (%s)", method, path, err, payload)
	}
	return run, response.Header.Get("ETag")
}

func (d *daemon) subscribe(t *testing.T, runID httpapi.RunId) *websocket.Conn {
	t.Helper()
	url := "ws" + strings.TrimPrefix(d.server.URL, "http") + "/api/v1/runs/" + runID + "/events/stream?after_sequence=0"
	header := http.Header{"Authorization": {"Bearer " + apiSecret}, "Origin": {d.origin}}
	c, response, err := websocket.Dial(context.Background(), url, &websocket.DialOptions{HTTPHeader: header})
	if err != nil {
		if response != nil {
			body, _ := io.ReadAll(response.Body)
			t.Fatalf("subscribe: status=%d body=%s err=%v", response.StatusCode, body, err)
		}
		t.Fatal(err)
	}
	return c
}

// streamMessage is the frame the stream handler writes. A subscriber outside
// the process knows the stream only by this JSON, so the test declares it
// rather than sharing the handler's type.
type streamMessage struct {
	Type         string         `json:"type"`
	Event        *httpapi.Event `json:"event,omitempty"`
	LastSequence int64          `json:"lastSequence"`
}

func readStream(t *testing.T, c *websocket.Conn) streamMessage {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, data, err := c.Read(ctx)
	if err != nil {
		t.Fatalf("stream read: %v", err)
	}
	var message streamMessage
	if err = json.Unmarshal(data, &message); err != nil {
		t.Fatal(err)
	}
	return message
}

func lifecycle(state string) func(httpapi.Event) bool {
	return func(e httpapi.Event) bool { return e.Kind == "run.lifecycle" && e.Payload["state"] == state }
}

func harnessEvent(name string) func(httpapi.Event) bool {
	return func(e httpapi.Event) bool { return e.Kind == name }
}

func eventKinds(events []httpapi.Event) string {
	out := make([]string, 0, len(events))
	for _, event := range events {
		out = append(out, event.Kind)
	}
	return "[" + strings.Join(out, " ") + "]"
}

// TestCommittedEventsReachASubscribedWebSocketInSequenceOrder is acceptance
// criterion 5 over the whole path this task creates: an event is committed to
// the durable store, claimed by the outbox worker, routed by the publisher into
// the live stream, and only then delivered to a subscriber holding a real
// WebSocket. Nothing here publishes into the stream directly, which is what
// separates it from the in-process stream tests.
func TestCommittedEventsReachASubscribedWebSocketInSequenceOrder(t *testing.T) {
	d := newDaemon(t)
	run, etag := d.command(t, http.MethodPost, "/api/v1/runs", "", httpapi.CreateRunRequest{
		WorkspacePath: "/workspace", HarnessProfile: "fake", SandboxProfile: "fake",
	}, http.StatusCreated)

	// Subscribed before the start, so the stream carries the run rather than
	// replaying it afterwards.
	c := d.subscribe(t, run.Id)
	defer func() { _ = c.CloseNow() }()
	if message := readStream(t, c); message.Type != "caught_up" || message.LastSequence != 0 {
		t.Fatalf("first message = %+v, want caught_up at sequence 0", message)
	}

	var received []httpapi.Event
	await := func(match func(httpapi.Event) bool) httpapi.Event {
		t.Helper()
		for {
			message := readStream(t, c)
			if message.Type != "event" || message.Event == nil {
				t.Fatalf("message = %+v, want an event", message)
			}
			received = append(received, *message.Event)
			if match(*message.Event) {
				return *message.Event
			}
		}
	}

	_, etag = d.command(t, http.MethodPost, "/api/v1/runs/"+run.Id+"/start", etag, nil, http.StatusAccepted)
	await(lifecycle("preparing"))
	await(lifecycle("running"))

	text := "echo hi"
	response, payload := d.call(t, http.MethodPost, "/api/v1/runs/"+run.Id+"/input", etag,
		httpapi.RunInputRequest{Kind: httpapi.RunInputRequestKindText, Text: &text})
	if response.StatusCode != http.StatusAccepted {
		t.Fatalf("input: status=%d body=%s", response.StatusCode, payload)
	}
	var accepted httpapi.InputAccepted
	if err := json.Unmarshal(payload, &accepted); err != nil {
		t.Fatalf("input: %v (%s)", err, payload)
	}

	// The fake harness echoes what it is given, so the output proves the input
	// travelled the other half of the same path: outbox, publisher, runner.
	output := await(harnessEvent("raw.output"))
	encoded, ok := output.Payload["data"].(string)
	if !ok {
		t.Fatalf("harness output carries no data: %+v", output.Payload)
	}
	echoed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("harness output is not base64: %v", err)
	}
	if !strings.Contains(string(echoed), accepted.CommandId) || !strings.Contains(string(echoed), text) {
		t.Fatalf("harness output %q cites neither the accepted command %s nor the input", echoed, accepted.CommandId)
	}
	await(harnessEvent("diagnostic"))

	// The output moved the run's version, so the stop presents the ETag the run
	// carries now rather than the one the start returned.
	_, etag = d.command(t, http.MethodGet, "/api/v1/runs/"+run.Id, "", nil, http.StatusOK)
	_, _ = d.command(t, http.MethodPost, "/api/v1/runs/"+run.Id+"/stop", etag, nil, http.StatusAccepted)
	await(lifecycle("completed"))

	for i, event := range received {
		if event.Sequence != int64(i+1) {
			t.Fatalf("subscriber received sequence %d in position %d: %s", event.Sequence, i, eventKinds(received))
		}
		if event.RunId != run.Id {
			t.Fatalf("subscriber received run %s, want %s", event.RunId, run.Id)
		}
	}

	// What arrived live is exactly what was committed, so the stream is not
	// merely ordered but complete.
	response, payload = d.call(t, http.MethodGet, "/api/v1/runs/"+run.Id+"/events?limit=200", "", nil)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("events: status=%d body=%s", response.StatusCode, payload)
	}
	var page httpapi.EventPage
	if err = json.Unmarshal(payload, &page); err != nil {
		t.Fatalf("events: %v (%s)", err, payload)
	}
	if len(page.Events) != len(received) {
		t.Fatalf("history holds %s but the subscriber received %s", eventKinds(page.Events), eventKinds(received))
	}
	for i := range page.Events {
		if page.Events[i].EventId != received[i].EventId || page.Events[i].Sequence != received[i].Sequence {
			t.Fatalf("position %d: history %s/%d, subscriber %s/%d", i, page.Events[i].EventId, page.Events[i].Sequence, received[i].EventId, received[i].Sequence)
		}
	}
}
