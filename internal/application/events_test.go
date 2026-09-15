package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shaielc/code-winch/internal/adapters/memory"
	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/pkg/protocol"
)

func newEventHarness(t *testing.T, count int) (*application.EventService, domain.RunID) {
	t.Helper()
	ids := &countingIDs{}
	runID, attemptID := ids.NewRunID(), ids.NewAttemptID()
	runs := &memory.RunRepository{}
	if _, err := runs.Save(context.Background(), application.RunRecord{ID: runID, Attempts: []domain.Attempt{{ID: attemptID, State: domain.RunStateRunning}}}, 0); err != nil {
		t.Fatal(err)
	}
	store := &memory.EventStore{}
	now, _ := domain.NewTimestamp(time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC))
	values := make([]application.UnsequencedEvent, count)
	for i := range values {
		values[i] = application.UnsequencedEvent{EventID: ids.NewEventID(), OccurredAt: now, Kind: "stream.raw", SchemaVersion: 1, Source: protocol.Source{Type: "harness"}, Sensitivity: protocol.SensitivityUserContent, Payload: []byte(`{"stream":"stdout","encoding":"utf-8","data":"x"}`)}
	}
	if count > 0 {
		if _, err := store.Append(context.Background(), runID, 0, values); err != nil {
			t.Fatal(err)
		}
	}
	service, err := application.NewEventService(runs, store)
	if err != nil {
		t.Fatal(err)
	}
	return service, runID
}

func TestListReturnsOrderedGapFreePagesAndAnExactCursor(t *testing.T) {
	service, runID := newEventHarness(t, 5)
	var all []protocol.Event
	cursor, pages := uint64(0), 0
	for {
		page, err := service.List(context.Background(), runID, cursor, 2)
		if err != nil {
			t.Fatal(err)
		}
		pages++
		all = append(all, page.Events...)
		if !page.HasMore {
			// The last page reports no more even though it is exactly full at the
			// boundary, which is what reading one past the page buys.
			if len(all) != 5 {
				t.Fatalf("collected %d events", len(all))
			}
			break
		}
		if len(page.Events) != 2 {
			t.Fatalf("page of %d with more to come", len(page.Events))
		}
		cursor = page.NextAfterSequence
	}
	if pages != 3 {
		t.Fatalf("paged in %d requests", pages)
	}
	for i, event := range all {
		if event.Sequence != uint64(i+1) {
			t.Fatalf("event %d has sequence %d", i, event.Sequence)
		}
	}
}

func TestListKeepsTheCallersCursorOnAnEmptyPage(t *testing.T) {
	service, runID := newEventHarness(t, 2)
	page, err := service.List(context.Background(), runID, 2, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Events) != 0 || page.HasMore || page.NextAfterSequence != 2 {
		t.Fatalf("empty page: %#v", page)
	}
}

func TestListDistinguishesAnUnknownRunFromAnEmptyOne(t *testing.T) {
	service, _ := newEventHarness(t, 0)
	unknown, _ := domain.ParseRunID("99999999-9999-4999-8999-999999999999")
	if _, err := service.List(context.Background(), unknown, 0, 50); !errors.Is(err, application.ErrNotFound) {
		t.Fatalf("unknown run: %v", err)
	}
}

func TestListRejectsANonPositiveLimit(t *testing.T) {
	service, runID := newEventHarness(t, 1)
	if _, err := service.List(context.Background(), runID, 0, 0); err == nil {
		t.Fatal("a zero limit must be rejected")
	}
}
