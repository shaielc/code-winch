package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/coder/websocket"
)

const defaultSubscriberBuffer = 64

// EventStream is the process-local live notification boundary. Durable storage
// remains authoritative: subscribers repair gaps from ListRunEvents.
type EventStream struct {
	mu       sync.Mutex
	capacity int
	next     uint64
	runs     map[RunId]map[uint64]*eventSubscriber
}

// Close disconnects all subscribers so daemon shutdown can drain boundedly.
func (b *EventStream) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for runID, subscribers := range b.runs {
		for id, sub := range subscribers {
			sub.once.Do(func() { close(sub.done) })
			delete(subscribers, id)
		}
		delete(b.runs, runID)
	}
}

type eventSubscriber struct {
	events chan Event
	done   chan struct{}
	once   sync.Once
}

// NewEventStream constructs a broadcaster with a bounded queue per browser.
func NewEventStream(capacity int) *EventStream {
	if capacity <= 0 {
		capacity = defaultSubscriberBuffer
	}
	return &EventStream{capacity: capacity, runs: make(map[RunId]map[uint64]*eventSubscriber)}
}

// Publish announces an event after its durable commit. It never waits for a
// subscriber; a full queue is closed and the browser can resume from storage.
func (b *EventStream) Publish(runID RunId, event Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for id, sub := range b.runs[runID] {
		select {
		case sub.events <- event:
		default:
			sub.once.Do(func() { close(sub.done) })
			delete(b.runs[runID], id)
		}
	}
}

func (b *EventStream) subscribe(runID RunId) (*eventSubscriber, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.next++
	id := b.next
	sub := &eventSubscriber{events: make(chan Event, b.capacity), done: make(chan struct{})}
	if b.runs[runID] == nil {
		b.runs[runID] = make(map[uint64]*eventSubscriber)
	}
	b.runs[runID][id] = sub
	return sub, func() {
		b.mu.Lock()
		delete(b.runs[runID], id)
		b.mu.Unlock()
		sub.once.Do(func() { close(sub.done) })
	}
}

type streamMessage struct {
	Type         string `json:"type"`
	Event        *Event `json:"event,omitempty"`
	LastSequence int64  `json:"lastSequence"`
	Resumable    bool   `json:"resumable,omitempty"`
}

func (s *server) streamEvents(w http.ResponseWriter, r *http.Request, runID RunId) {
	if s.cfg.EventStream == nil {
		s.problem(w, r, http.StatusNotImplemented, "stream_unavailable", "Stream unavailable", "The event stream is not configured.")
		return
	}
	if r.Header.Get("Origin") != s.cfg.AllowedOrigin {
		s.problem(w, r, http.StatusForbidden, "origin_failed", "Forbidden", "The WebSocket origin is invalid.")
		return
	}
	after, err := strconv.ParseInt(r.URL.Query().Get("after_sequence"), 10, 64)
	if r.URL.Query().Get("after_sequence") == "" {
		after = 0
		err = nil
	}
	if err != nil || after < 0 {
		s.problem(w, r, http.StatusBadRequest, "bad_request", "Bad request", "after_sequence must be a non-negative integer.")
		return
	}
	// Subscribe before reading history. Events committed during the snapshot are
	// queued and deduplicated by sequence, eliminating the handoff race.
	sub, unsubscribe := s.cfg.EventStream.subscribe(runID)
	defer unsubscribe()
	if _, err = s.backend.GetRun(r.Context(), actor(r), runID); err != nil {
		s.backendProblem(w, r, err)
		return
	}
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{s.cfg.AllowedOrigin}})
	if err != nil {
		return
	}
	defer func() { _ = c.CloseNow() }()

	heartbeat := duration(s.cfg.HeartbeatInterval, 15*time.Second)
	reauthorize := duration(s.cfg.ReauthorizeInterval, time.Minute)
	writeTimeout := duration(s.cfg.StreamWriteTimeout, 5*time.Second)
	hb := time.NewTicker(heartbeat)
	defer hb.Stop()
	auth := time.NewTicker(reauthorize)
	defer auth.Stop()
	last := after
	write := func(m streamMessage) error {
		ctx, cancel := context.WithTimeout(r.Context(), writeTimeout)
		defer cancel()
		return c.Write(ctx, websocket.MessageText, mustJSON(m))
	}
	repair := func() error {
		for {
			page, readErr := s.backend.ListRunEvents(r.Context(), actor(r), runID, last, 200)
			if readErr != nil {
				return readErr
			}
			for i := range page.Events {
				e := page.Events[i]
				if e.Sequence <= last {
					continue
				}
				if e.Sequence != last+1 {
					return errors.New("non-contiguous durable event history")
				}
				if err := write(streamMessage{Type: "event", Event: &e, LastSequence: e.Sequence}); err != nil {
					return err
				}
				last = e.Sequence
			}
			if !page.HasMore {
				return nil
			}
		}
	}
	if err = repair(); err != nil {
		s.closeStream(r, c, runID, last, "history_unavailable")
		return
	}
	if err = write(streamMessage{Type: "caught_up", LastSequence: last}); err != nil {
		return
	}
	for {
		select {
		case <-r.Context().Done():
			return
		case <-sub.done:
			_ = write(streamMessage{Type: "disconnect", LastSequence: last, Resumable: true})
			_ = c.Close(websocket.StatusPolicyViolation, "slow_consumer")
			return
		case <-sub.events:
			if err = repair(); err != nil {
				s.closeStream(r, c, runID, last, "history_unavailable")
				return
			}
		case <-hb.C:
			if err = write(streamMessage{Type: "heartbeat", LastSequence: last}); err != nil {
				return
			}
		case <-auth.C:
			if _, err = s.backend.GetRun(r.Context(), actor(r), runID); err != nil {
				s.closeStream(r, c, runID, last, "authorization_revoked")
				return
			}
		}
	}
}

func duration(value, fallback time.Duration) time.Duration {
	if value > 0 {
		return value
	}
	return fallback
}
func mustJSON(v any) []byte { b, _ := json.Marshal(v); return b }
func (s *server) closeStream(ctx *http.Request, c *websocket.Conn, runID RunId, last int64, reason string) {
	s.cfg.Logger.LogAttrs(ctx.Context(), slog.LevelWarn, "event stream closed", slog.String("request_id", ctx.Context().Value(requestKey{}).(identity).requestID), slog.String("run_id", runID), slog.String("error_code", reason), slog.Int64("sequence", last))
	writeCtx, cancel := context.WithTimeout(context.Background(), duration(s.cfg.StreamWriteTimeout, 5*time.Second))
	_ = c.Write(writeCtx, websocket.MessageText, mustJSON(streamMessage{Type: "disconnect", LastSequence: last, Resumable: true}))
	cancel()
	_ = c.Close(websocket.StatusPolicyViolation, reason)
}
