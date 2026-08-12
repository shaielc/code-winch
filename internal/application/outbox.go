package application

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/shaielc/code-winch/internal/domain"
)

type OutboxClock interface{ Now() time.Time }

type OutboxWorkerConfig struct {
	WorkerID, LeaseToken string
	BatchSize            int
	LeaseDuration        time.Duration
	BaseBackoff          time.Duration
	MaxBackoff           time.Duration
	MaxAttempts          uint
}

type OutboxMetrics struct{ Backlog, Retries, Poison uint64 }

// OutboxWorker delivers a claimed batch. Publishing precedes the completion
// marker deliberately: a crash between them causes a safe duplicate delivery.
type OutboxWorker struct {
	store   OutboxStore
	pub     OutboxPublisher
	clock   OutboxClock
	cfg     OutboxWorkerConfig
	retries atomic.Uint64
	poison  atomic.Uint64
}

func NewOutboxWorker(store OutboxStore, pub OutboxPublisher, clock OutboxClock, cfg OutboxWorkerConfig) (*OutboxWorker, error) {
	if store == nil || pub == nil || clock == nil || cfg.WorkerID == "" || cfg.LeaseToken == "" || cfg.BatchSize <= 0 || cfg.LeaseDuration <= 0 || cfg.BaseBackoff <= 0 || cfg.MaxBackoff < cfg.BaseBackoff || cfg.MaxAttempts == 0 {
		return nil, fmt.Errorf("outbox worker: invalid configuration")
	}
	return &OutboxWorker{store: store, pub: pub, clock: clock, cfg: cfg}, nil
}

func (w *OutboxWorker) RunOnce(ctx context.Context) (int, error) {
	now, err := domain.NewTimestamp(w.clock.Now())
	if err != nil {
		return 0, fmt.Errorf("outbox worker: clock: %w", err)
	}
	until, _ := domain.NewTimestamp(now.Time().Add(w.cfg.LeaseDuration))
	records, err := w.store.ClaimOutbox(ctx, w.cfg.WorkerID, w.cfg.LeaseToken, now, until, w.cfg.BatchSize)
	if err != nil {
		return 0, fmt.Errorf("outbox worker: claim: %w", err)
	}
	for i, record := range records {
		if err = w.pub.Publish(ctx, record.Message); err == nil {
			if err = w.store.CompleteOutbox(ctx, record.Message.ID, record.LeaseToken, now); err != nil {
				return i, fmt.Errorf("outbox worker: complete message=%s: %w", record.Message.ID, err)
			}
			continue
		}
		w.retries.Add(1)
		if record.Attempts >= w.cfg.MaxAttempts {
			w.poison.Add(1)
			if poisonErr := w.store.PoisonOutbox(ctx, record.Message.ID, record.LeaseToken, now, "publish_failed"); poisonErr != nil {
				return i, fmt.Errorf("outbox worker: poison message=%s: %w", record.Message.ID, poisonErr)
			}
			continue
		}
		delay := w.cfg.BaseBackoff
		for n := uint(1); n < record.Attempts && delay < w.cfg.MaxBackoff; n++ {
			delay *= 2
			if delay > w.cfg.MaxBackoff {
				delay = w.cfg.MaxBackoff
			}
		}
		next, _ := domain.NewTimestamp(now.Time().Add(delay))
		// Store only a stable class, never publisher details which may contain content.
		if retryErr := w.store.RetryOutbox(ctx, record.Message.ID, record.LeaseToken, next, "publish_failed"); retryErr != nil {
			return i, fmt.Errorf("outbox worker: retry message=%s: %w", record.Message.ID, retryErr)
		}
	}
	return len(records), nil
}

func (w *OutboxWorker) Metrics(ctx context.Context) (OutboxMetrics, error) {
	now, err := domain.NewTimestamp(w.clock.Now())
	if err != nil {
		return OutboxMetrics{}, err
	}
	backlog, err := w.store.OutboxBacklog(ctx, now)
	return OutboxMetrics{Backlog: backlog, Retries: w.retries.Load(), Poison: w.poison.Load()}, err
}
