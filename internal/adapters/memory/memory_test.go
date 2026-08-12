package memory_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/shaielc/code-winch/internal/adapters/memory"
	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/pkg/protocol"
)

var (
	_ application.WorkspaceRepository  = (*memory.WorkspaceRepository)(nil)
	_ application.RunRepository        = (*memory.RunRepository)(nil)
	_ application.EventStore           = (*memory.EventStore)(nil)
	_ application.OutboxPublisher      = (*memory.OutboxPublisher)(nil)
	_ application.SecretReferenceStore = (*memory.SecretReferenceStore)(nil)
	_ application.RunnerGateway        = (*memory.RunnerGateway)(nil)
	_ application.Clock                = (*memory.Clock)(nil)
	_ application.IDSource             = (*memory.IDSource)(nil)
)

func runID(t *testing.T, n int) domain.RunID {
	t.Helper()
	id, err := domain.ParseRunID(fmt.Sprintf("00000000-0000-0000-0000-%012d", n))
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func eventID(t *testing.T, n int) domain.EventID {
	t.Helper()
	id, err := domain.ParseEventID(fmt.Sprintf("00000000-0000-0000-0000-%012d", n))
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func commandID(t *testing.T, n int) domain.CommandID {
	t.Helper()
	id, err := domain.ParseCommandID(fmt.Sprintf("00000000-0000-0000-0000-%012d", n))
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestEventStoreSequenceConflictAndCopies(t *testing.T) {
	store := &memory.EventStore{}
	now, _ := domain.NewTimestamp(time.Date(2026, 8, 12, 1, 2, 3, 0, time.UTC))
	payload := []byte(`{"text":"safe"}`)
	input := application.UnsequencedEvent{EventID: eventID(t, 1), OccurredAt: now, Kind: "test", SchemaVersion: 1, Source: protocol.Source{Type: "test"}, Sensitivity: protocol.SensitivityUserContent, Payload: payload}

	added, err := store.Append(context.Background(), runID(t, 1), 0, []application.UnsequencedEvent{input})
	if err != nil {
		t.Fatal(err)
	}
	payload[2] = 'X'
	added[0].Payload[2] = 'Y'
	read, err := store.Read(context.Background(), runID(t, 1), 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(read[0].Payload); got != `{"text":"safe"}` {
		t.Fatalf("stored payload aliased caller: %s", got)
	}
	if read[0].Sequence != 1 {
		t.Fatalf("sequence = %d", read[0].Sequence)
	}

	_, err = store.Append(context.Background(), runID(t, 1), 0, []application.UnsequencedEvent{input})
	if !errors.Is(err, application.ErrConflict) {
		t.Fatalf("expected sequence conflict, got %v", err)
	}
	duplicate := input
	_, err = store.Append(context.Background(), runID(t, 1), 1, []application.UnsequencedEvent{duplicate})
	if !errors.Is(err, application.ErrConflict) {
		t.Fatalf("expected event ID conflict, got %v", err)
	}
	if len(store.AppendCalls()) != 3 {
		t.Fatal("all append effects were not recorded")
	}
}

func TestRunRepositoryOptimisticConcurrency(t *testing.T) {
	repository := &memory.RunRepository{}
	record := application.RunRecord{ID: runID(t, 1)}
	version, err := repository.Save(context.Background(), record, 0)
	if err != nil || version != 1 {
		t.Fatalf("create: version=%d err=%v", version, err)
	}
	if _, err = repository.Save(context.Background(), record, 0); !errors.Is(err, application.ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
	if _, err = repository.Save(context.Background(), record, 1); err != nil {
		t.Fatal(err)
	}
}

func TestIdempotentExternalBoundariesStillRecordCalls(t *testing.T) {
	publisher := &memory.OutboxPublisher{}
	message := application.OutboxMessage{ID: commandID(t, 1), Topic: "run.input", Payload: []byte("payload")}
	for range 2 {
		if err := publisher.Publish(context.Background(), message); err != nil {
			t.Fatal(err)
		}
	}
	if len(publisher.Calls()) != 2 {
		t.Fatal("publisher did not record replay")
	}
	if len(publisher.Published()) != 1 {
		t.Fatal("publisher duplicated an idempotent effect")
	}

	runner := &memory.RunnerGateway{}
	runnerMessage := protocol.RunnerMessage{CommandID: commandID(t, 2).String(), Payload: []byte(`{}`)}
	for range 2 {
		if err := runner.Send(context.Background(), runnerMessage); err != nil {
			t.Fatal(err)
		}
	}
	if len(runner.Calls()) != 2 {
		t.Fatal("runner did not record replay")
	}
	if len(runner.Delivered()) != 1 {
		t.Fatal("runner duplicated an idempotent effect")
	}
}

func TestFailurePlanIsDeterministic(t *testing.T) {
	injected := errors.New("planned outage")
	publisher := &memory.OutboxPublisher{}
	publisher.Failures.Inject("publish", injected, nil)
	message := application.OutboxMessage{ID: commandID(t, 1)}
	if err := publisher.Publish(context.Background(), message); !errors.Is(err, injected) {
		t.Fatalf("first error = %v", err)
	}
	if err := publisher.Publish(context.Background(), message); err != nil {
		t.Fatalf("second error = %v", err)
	}
	if len(publisher.Calls()) != 1 {
		t.Fatal("failed call should not be an external effect")
	}
}

func TestConcurrentPublisherAccess(t *testing.T) {
	publisher := &memory.OutboxPublisher{}
	const count = 100
	var group sync.WaitGroup
	for i := 1; i <= count; i++ {
		group.Add(1)
		go func(n int) {
			defer group.Done()
			_ = publisher.Publish(context.Background(), application.OutboxMessage{ID: commandID(t, n)})
		}(i)
	}
	group.Wait()
	if got := len(publisher.Calls()); got != count {
		t.Fatalf("calls = %d", got)
	}
}

func TestClockAndIDSourceAreControlled(t *testing.T) {
	now, _ := domain.NewTimestamp(time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC))
	clock := memory.NewClock(now)
	if clock.Now() != now {
		t.Fatal("clock changed without instruction")
	}
	source := &memory.IDSource{RunIDs: []domain.RunID{runID(t, 1), runID(t, 2)}}
	if source.NewRunID() != runID(t, 1) || source.NewRunID() != runID(t, 2) {
		t.Fatal("IDs returned out of order")
	}
}
