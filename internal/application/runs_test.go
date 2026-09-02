package application_test

import (
	"context"
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
	attemptID, _ := domain.ParseAttemptID("22222222-2222-2222-2222-222222222222")
	now, _ := domain.NewTimestamp(time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC))
	service, err := application.NewRunService(&memory.RunRepository{}, fixedClock{now}, &memory.IDSource{RunIDs: []domain.RunID{runID}, AttemptIDs: []domain.AttemptID{attemptID}})
	if err != nil {
		t.Fatal(err)
	}
	created, err := service.Create(context.Background(), application.CreateRunCommand{WorkspacePath: "/tmp/ws", HarnessProfile: "fake", SandboxProfile: "local"})
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
}
