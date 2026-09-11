package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shaielc/code-winch/internal/adapters/memory"
	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
)

type fixedClock struct{ at domain.Timestamp }

func (c fixedClock) Now() domain.Timestamp { return c.at }

func TestCreateAndGetRun(t *testing.T) {
	runID, _ := domain.ParseRunID("11111111-1111-1111-1111-111111111111")
	runID2, _ := domain.ParseRunID("33333333-3333-3333-3333-333333333333")
	runID3, _ := domain.ParseRunID("44444444-4444-4444-4444-444444444444")
	attemptID, _ := domain.ParseAttemptID("22222222-2222-2222-2222-222222222222")
	attemptID2, _ := domain.ParseAttemptID("55555555-5555-5555-5555-555555555555")
	attemptID3, _ := domain.ParseAttemptID("66666666-6666-6666-6666-666666666666")
	now, _ := domain.NewTimestamp(time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC))
	service, err := application.NewRunService(&memory.RunRepository{}, fixedClock{now}, &memory.IDSource{RunIDs: []domain.RunID{runID, runID2, runID3}, AttemptIDs: []domain.AttemptID{attemptID, attemptID2, attemptID3}})
	if err != nil {
		t.Fatal(err)
	}
	created, err := service.Create(context.Background(), application.CreateRunCommand{WorkspacePath: "/tmp/ws", HarnessProfile: "fake", SandboxProfile: "local", Actor: "actor", IdempotencyKey: "key"})
	if err != nil {
		t.Fatal(err)
	}
	read, err := service.Get(context.Background(), runID)
	if err != nil {
		t.Fatal(err)
	}
	if created.Version != 1 || read.Version != 1 || read.Record.Attempts[0].State != domain.RunStateCreated || read.Record.WorkspacePath != "/tmp/ws" {
		t.Fatalf("unexpected run: %#v", read)
	}
	replayed, err := service.Create(context.Background(), application.CreateRunCommand{WorkspacePath: "/tmp/ws", HarnessProfile: "fake", SandboxProfile: "local", Actor: "actor", IdempotencyKey: "key"})
	if err != nil || replayed.Record.ID != runID {
		t.Fatalf("replay: %#v %v", replayed, err)
	}
	_, err = service.Create(context.Background(), application.CreateRunCommand{WorkspacePath: "/different", HarnessProfile: "fake", SandboxProfile: "local", Actor: "actor", IdempotencyKey: "key"})
	if !errors.Is(err, application.ErrIdempotencyConflict) {
		t.Fatalf("conflicting replay: %v", err)
	}
}

func TestGetRejectsRunWithoutAttempts(t *testing.T) {
	id, _ := domain.ParseRunID("77777777-7777-7777-7777-777777777777")
	repository := &memory.RunRepository{}
	_, _ = repository.Save(context.Background(), application.RunRecord{ID: id}, 0)
	now, _ := domain.NewTimestamp(time.Now())
	service, _ := application.NewRunService(repository, fixedClock{now}, &memory.IDSource{})
	_, err := service.Get(context.Background(), id)
	if !errors.Is(err, application.ErrInvalidRunRecord) {
		t.Fatalf("invalid record error: %v", err)
	}
}
