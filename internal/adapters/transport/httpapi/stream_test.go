package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
)

type streamBackend struct {
	backendStub
	mu       sync.Mutex
	events   []Event
	getCalls int
	revokeAt int
	onList   func()
}

func (b *streamBackend) GetRun(_ context.Context, actor string, _ RunId) (Run, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.actor = actor
	b.getCalls++
	if b.revokeAt > 0 && b.getCalls >= b.revokeAt {
		return Run{}, ErrRunNotFound
	}
	return sampleRun(), nil
}

func (b *streamBackend) ListRunEvents(_ context.Context, actor string, _ RunId, after int64, limit int) (EventPage, error) {
	b.mu.Lock()
	b.actor = actor
	hook := b.onList
	b.onList = nil
	b.mu.Unlock()
	if hook != nil {
		hook()
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	page := EventPage{NextAfterSequence: after}
	for _, event := range b.events {
		if event.Sequence > after && len(page.Events) < limit {
			page.Events = append(page.Events, event)
			page.NextAfterSequence = event.Sequence
		}
	}
	for _, event := range b.events {
		if event.Sequence > page.NextAfterSequence {
			page.HasMore = true
			break
		}
	}
	return page, nil
}

func event(sequence int64) Event {
	return Event{EventId: "event", RunId: sampleRun().Id, Sequence: sequence, OccurredAt: time.Unix(sequence, 0), Kind: "test", SchemaVersion: 1, Source: map[string]any{"type": "test"}, Sensitivity: EventSensitivity("operational"), Payload: map[string]any{"safe": true}}
}

func streamServer(t *testing.T, backend *streamBackend, stream *EventStream, reauth time.Duration) *httptest.Server {
	t.Helper()
	cfg := Config{Token: testToken, CSRFToken: testCSRF, Actor: "local-user", RequestID: func() string { return "stream-request" }, EventStream: stream, HeartbeatInterval: time.Hour, ReauthorizeInterval: reauth, StreamWriteTimeout: time.Second}
	// The exact allowed origin must be known before constructing the handler.
	listener := httptest.NewUnstartedServer(nil)
	cfg.AllowedOrigin = "http://" + listener.Listener.Addr().String()
	handler, err := NewHandler(cfg, backend)
	if err != nil {
		t.Fatal(err)
	}
	listener.Config.Handler = handler
	listener.Start()
	t.Cleanup(listener.Close)
	return listener
}

func dialStream(t *testing.T, server *httptest.Server, after int64, origin string) *websocket.Conn {
	t.Helper()
	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/v1/runs/" + sampleRun().Id + "/events/stream?after_sequence=" + jsonNumber(after)
	header := http.Header{"Authorization": {"Bearer " + testToken}, "Origin": {origin}}
	c, resp, err := websocket.Dial(context.Background(), url, &websocket.DialOptions{HTTPHeader: header})
	if err != nil {
		if resp != nil {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("dial status=%d body=%s err=%v", resp.StatusCode, body, err)
		}
		t.Fatal(err)
	}
	return c
}

func jsonNumber(value int64) string { return strings.TrimSpace(string(mustJSON(value))) }

func readMessage(t *testing.T, c *websocket.Conn) streamMessage {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, data, err := c.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var message streamMessage
	if err = json.Unmarshal(data, &message); err != nil {
		t.Fatal(err)
	}
	return message
}

func TestStreamReconnectAndSnapshotLiveHandoff(t *testing.T) {
	stream := NewEventStream(4)
	b := &streamBackend{events: []Event{event(1), event(2)}}
	b.onList = func() {
		b.mu.Lock()
		b.events = append(b.events, event(3))
		b.mu.Unlock()
		stream.Publish(sampleRun().Id, event(3))
	}
	ts := streamServer(t, b, stream, time.Hour)
	c := dialStream(t, ts, 1, ts.URL)
	defer func() { _ = c.CloseNow() }()
	for sequence := int64(2); sequence <= 3; sequence++ {
		message := readMessage(t, c)
		if message.Type != "event" || message.Event == nil || message.Event.Sequence != sequence {
			t.Fatalf("message=%+v", message)
		}
	}
	if message := readMessage(t, c); message.Type != "caught_up" || message.LastSequence != 3 {
		t.Fatalf("caught_up=%+v", message)
	}
	b.mu.Lock()
	b.events = append(b.events, event(4))
	b.mu.Unlock()
	stream.Publish(sampleRun().Id, event(4))
	if message := readMessage(t, c); message.Event == nil || message.Event.Sequence != 4 {
		t.Fatalf("live=%+v", message)
	}
}

func TestEventStreamDisconnectsSlowSubscriberWithoutBlockingPublisher(t *testing.T) {
	stream := NewEventStream(1)
	sub, cancel := stream.subscribe(sampleRun().Id)
	defer cancel()
	stream.Publish(sampleRun().Id, event(1))
	done := make(chan struct{})
	go func() { stream.Publish(sampleRun().Id, event(2)); close(done) }()
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("publisher was backpressured")
	}
	select {
	case <-sub.done:
	case <-time.After(time.Second):
		t.Fatal("slow subscriber remained connected")
	}
}

func TestStreamRejectsOriginAndReauthorizes(t *testing.T) {
	stream := NewEventStream(2)
	b := &streamBackend{revokeAt: 2}
	ts := streamServer(t, b, stream, 10*time.Millisecond)
	url := "ws" + strings.TrimPrefix(ts.URL, "http") + "/api/v1/runs/" + sampleRun().Id + "/events/stream?after_sequence=0"
	_, resp, err := websocket.Dial(context.Background(), url, &websocket.DialOptions{HTTPHeader: http.Header{"Authorization": {"Bearer " + testToken}, "Origin": {"https://evil.example"}}})
	if err == nil || resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Fatalf("origin status=%v err=%v", resp, err)
	}
	c := dialStream(t, ts, 0, ts.URL)
	defer func() { _ = c.CloseNow() }()
	if message := readMessage(t, c); message.Type != "caught_up" {
		t.Fatalf("message=%+v", message)
	}
	if message := readMessage(t, c); message.Type != "disconnect" || !message.Resumable {
		t.Fatalf("disconnect=%+v", message)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, _, err = c.Read(ctx)
	if websocket.CloseStatus(err) != websocket.StatusPolicyViolation {
		t.Fatalf("reauthorization close=%v status=%v", err, websocket.CloseStatus(err))
	}
}
