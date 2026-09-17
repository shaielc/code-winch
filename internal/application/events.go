package application

import (
	"context"
	"errors"

	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/pkg/protocol"
)

// EventPage is one ordered, gap-free window over a run's durable events.
// NextAfterSequence is the cursor a caller passes back to continue, and is the
// caller's own cursor when the page is empty rather than a reset to zero.
type EventPage struct {
	Events            []protocol.Event
	NextAfterSequence uint64
	HasMore           bool
}

// EventService reads persisted events. It deliberately reads from the store
// rather than from a live subscription: a poll must answer the same ordered
// history whether or not the run is still executing.
type EventService struct {
	runs   RunRepository
	events EventStore
}

func NewEventService(runs RunRepository, events EventStore) (*EventService, error) {
	if runs == nil || events == nil {
		return nil, errors.New("event service: repository and event store are required")
	}
	return &EventService{runs: runs, events: events}, nil
}

// List answers a page of events for an existing run. The run is read first so
// an unknown run is a not-found answer rather than an empty page, which a
// client could not tell from a run that has produced nothing yet.
func (s *EventService) List(ctx context.Context, id domain.RunID, after uint64, limit int) (EventPage, error) {
	if limit <= 0 {
		return EventPage{}, errors.New("event service: limit must be positive")
	}
	if _, _, err := s.runs.Get(ctx, id); err != nil {
		return EventPage{}, err
	}
	// Reading one past the page decides hasMore exactly, rather than inferring it
	// from a full page and reporting more when the last event fit precisely.
	events, err := s.events.Read(ctx, id, after, limit+1)
	if err != nil {
		return EventPage{}, err
	}
	page := EventPage{Events: events, NextAfterSequence: after}
	if len(events) > limit {
		page.Events, page.HasMore = events[:limit], true
	}
	if n := len(page.Events); n > 0 {
		page.NextAfterSequence = page.Events[n-1].Sequence
	}
	return page, nil
}
