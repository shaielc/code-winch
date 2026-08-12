package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
)

type clock struct{ now time.Time }

func (c *clock) Now() time.Time { return c.now }

type publisher struct {
	err   error
	calls int
}

func (p *publisher) Publish(_ context.Context, _ application.OutboxMessage) error {
	p.calls++
	return p.err
}

type store struct {
	records                      []application.OutboxRecord
	completed, retried, poisoned int
	retryAt                      time.Time
	completeErr                  error
}

func (s *store) ClaimOutbox(_ context.Context, _, token string, _, _ domain.Timestamp, limit int) ([]application.OutboxRecord, error) {
	out := append([]application.OutboxRecord(nil), s.records...)
	if len(out) > limit {
		out = out[:limit]
	}
	for i := range out {
		out[i].LeaseToken = token
	}
	return out, nil
}
func (s *store) CompleteOutbox(context.Context, domain.CommandID, string, domain.Timestamp) error {
	s.completed++
	return s.completeErr
}
func (s *store) RetryOutbox(_ context.Context, _ domain.CommandID, _ string, at domain.Timestamp, _ string) error {
	s.retried++
	s.retryAt = at.Time()
	return nil
}
func (s *store) PoisonOutbox(context.Context, domain.CommandID, string, domain.Timestamp, string) error {
	s.poisoned++
	return nil
}
func (s *store) OutboxBacklog(context.Context, domain.Timestamp) (uint64, error) {
	return uint64(len(s.records)), nil
}

func message(t *testing.T, attempts uint) application.OutboxRecord {
	t.Helper()
	id, err := domain.ParseCommandID("00000000-0000-0000-0000-000000000001")
	if err != nil {
		t.Fatal(err)
	}
	return application.OutboxRecord{Message: application.OutboxMessage{ID: id, Topic: "run.events", Payload: []byte(`{"safe":true}`)}, Attempts: attempts}
}
func worker(t *testing.T, s application.OutboxStore, p application.OutboxPublisher, c *clock) *application.OutboxWorker {
	t.Helper()
	w, err := application.NewOutboxWorker(s, p, c, application.OutboxWorkerConfig{WorkerID: "worker-1", LeaseToken: "00000000-0000-0000-0000-000000000099", BatchSize: 10, LeaseDuration: time.Minute, BaseBackoff: time.Second, MaxBackoff: 8 * time.Second, MaxAttempts: 3})
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func TestOutboxWorkerCompletesOnlyAfterPublish(t *testing.T) {
	s := &store{records: []application.OutboxRecord{message(t, 1)}}
	p := &publisher{}
	c := &clock{now: time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC)}
	n, err := worker(t, s, p, c).RunOnce(context.Background())
	if err != nil || n != 1 || p.calls != 1 || s.completed != 1 {
		t.Fatalf("n=%d calls=%d complete=%d err=%v", n, p.calls, s.completed, err)
	}
}

func TestOutboxWorkerCrashPointRedeliversWithStableID(t *testing.T) {
	s := &store{records: []application.OutboxRecord{message(t, 1)}, completeErr: errors.New("database unavailable")}
	p := &publisher{}
	c := &clock{now: time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC)}
	w := worker(t, s, p, c)
	if _, err := w.RunOnce(context.Background()); err == nil {
		t.Fatal("completion failure was hidden")
	}
	// Simulate restart after lease expiry: the durable intent is claimed again.
	s.completeErr = nil
	c.now = c.now.Add(2 * time.Minute)
	if _, err := w.RunOnce(context.Background()); err != nil || p.calls != 2 || s.completed != 2 {
		t.Fatalf("calls=%d completions=%d err=%v", p.calls, s.completed, err)
	}
}

func TestOutboxWorkerRetriesWithFakeTimeAndPoisons(t *testing.T) {
	c := &clock{now: time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC)}
	p := &publisher{err: errors.New("payload must not leak")}
	s := &store{records: []application.OutboxRecord{message(t, 2)}}
	w := worker(t, s, p, c)
	_, err := w.RunOnce(context.Background())
	if err != nil || s.retried != 1 || !s.retryAt.Equal(c.now.Add(2*time.Second)) {
		t.Fatalf("retry=%d at=%v err=%v", s.retried, s.retryAt, err)
	}
	s.records = []application.OutboxRecord{message(t, 3)}
	_, err = w.RunOnce(context.Background())
	metrics, _ := w.Metrics(context.Background())
	if err != nil || s.poisoned != 1 || metrics.Retries != 2 || metrics.Poison != 1 {
		t.Fatalf("poison=%d metrics=%+v err=%v", s.poisoned, metrics, err)
	}
}
