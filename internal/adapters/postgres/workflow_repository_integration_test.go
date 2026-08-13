//go:build integration

package postgres_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	pgstore "github.com/shaielc/code-winch/internal/adapters/postgres"
	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
)

var _ application.WorkflowRepository = (*pgstore.Store)(nil)

func workflowFixture(t *testing.T) (*pgstore.Store, application.WorkflowInstanceRecord, application.WorkflowStepAttempt, domain.Timestamp) {
	t.Helper()
	_, store := database(t)
	now, _ := domain.NewTimestamp(time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC))
	definition := application.WorkflowDefinitionRecord{DefinitionID: "release", Version: "1", Document: json.RawMessage(`{"schemaVersion":1,"definitionId":"release","version":"1","steps":[]}`), CreatedAt: now}
	if err := store.PutWorkflowDefinition(context.Background(), definition); err != nil {
		t.Fatal(err)
	}
	instance := application.WorkflowInstanceRecord{ID: id(t, domain.ParseWorkflowID, 1001), DefinitionID: "release", DefinitionVersion: "1", Status: "running", Inputs: json.RawMessage(`{"safe":true}`), HarnessProfile: "default", SandboxProfile: "locked", CreatedAt: now}
	attempt := application.WorkflowStepAttempt{ID: id(t, domain.ParseAttemptID, 1002), InstanceID: instance.ID, StepID: "build", Attempt: 1, State: "ready", AvailableAt: now}
	if err := store.CreateWorkflowInstance(context.Background(), instance, []application.WorkflowStepAttempt{attempt}); err != nil {
		t.Fatal(err)
	}
	return store, instance, attempt, now
}

func TestWorkflowStepLeaseContentionAndExpiredLeaseReclaim(t *testing.T) {
	store, instance, _, now := workflowFixture(t)
	until, _ := domain.NewTimestamp(now.Time().Add(time.Minute))
	type result struct {
		leases []application.WorkflowStepAttempt
		err    error
	}
	results := make(chan result, 16)
	for worker := 1; worker <= 16; worker++ {
		go func(worker int) {
			leases, err := store.ClaimReadyWorkflowSteps(context.Background(), fmt.Sprintf("worker-%d", worker), fmt.Sprintf("00000000-0000-0000-0000-%012d", 1100+worker), now, until, 1)
			results <- result{leases, err}
		}(worker)
	}
	var lease application.WorkflowStepAttempt
	claims := 0
	for range 16 {
		result := <-results
		if result.err != nil {
			t.Fatal(result.err)
		}
		claims += len(result.leases)
		if len(result.leases) == 1 {
			lease = result.leases[0]
		}
	}
	if claims != 1 || lease.LeaseEpoch != 1 {
		t.Fatalf("claims=%d epoch=%d, want one claim at epoch 1", claims, lease.LeaseEpoch)
	}
	afterExpiry, _ := domain.NewTimestamp(until.Time().Add(time.Second))
	secondUntil, _ := domain.NewTimestamp(afterExpiry.Time().Add(time.Minute))
	reclaimed, err := store.ClaimReadyWorkflowSteps(context.Background(), "recovery-worker", "00000000-0000-0000-0000-000000001199", afterExpiry, secondUntil, 1)
	if err != nil || len(reclaimed) != 1 || reclaimed[0].LeaseEpoch != 2 {
		t.Fatalf("reclaimed=%#v err=%v", reclaimed, err)
	}
	if err = store.CompleteWorkflowStep(context.Background(), lease, "completed", json.RawMessage(`{"stale":true}`), "", afterExpiry, nil); !errors.Is(err, application.ErrConflict) {
		t.Fatalf("stale worker completion error=%v", err)
	}
	if err = store.CompleteWorkflowStep(context.Background(), reclaimed[0], "completed", json.RawMessage(`{"ok":true}`), "", afterExpiry, []application.WorkflowOutboxMessage{{ID: "00000000-0000-0000-0000-000000001200", Topic: "workflow.step.completed", Payload: json.RawMessage(`{"workflowId":"00000000-0000-0000-0000-000000001001"}`)}}); err != nil {
		t.Fatal(err)
	}
	_, history, err := store.GetWorkflowInstance(context.Background(), instance.ID)
	if err != nil || len(history) != 1 || history[0].State != "completed" {
		t.Fatalf("history=%#v err=%v", history, err)
	}
}

func TestWorkflowCompletionCrashPointIsAtomicAndHistoryIsAppendOnly(t *testing.T) {
	store, instance, first, now := workflowFixture(t)
	until, _ := domain.NewTimestamp(now.Time().Add(time.Minute))
	leases, err := store.ClaimReadyWorkflowSteps(context.Background(), "worker", "00000000-0000-0000-0000-000000001300", now, until, 1)
	if err != nil || len(leases) != 1 {
		t.Fatalf("claim=%#v err=%v", leases, err)
	}
	// A duplicate second intent fails after the attempt update and first intent
	// insert. The transaction must roll both writes back, as after a process
	// crash at the same pre-commit point.
	intent := application.WorkflowOutboxMessage{ID: "00000000-0000-0000-0000-000000001301", Topic: "workflow.done", Payload: json.RawMessage(`{"safe":true}`)}
	err = store.CompleteWorkflowStep(context.Background(), leases[0], "completed", json.RawMessage(`{"ok":true}`), "", now, []application.WorkflowOutboxMessage{intent, intent})
	if !errors.Is(err, application.ErrConflict) {
		t.Fatalf("completion error=%v", err)
	}
	_, history, err := store.GetWorkflowInstance(context.Background(), instance.ID)
	if err != nil || history[0].State != "running" {
		t.Fatalf("post-crash history=%#v err=%v", history, err)
	}
	retry := application.WorkflowStepAttempt{ID: id(t, domain.ParseAttemptID, 1302), InstanceID: instance.ID, StepID: first.StepID, Attempt: 2, State: "ready", AvailableAt: now}
	if err = store.AppendWorkflowAttempt(context.Background(), retry); err != nil {
		t.Fatal(err)
	}
	_, history, err = store.GetWorkflowInstance(context.Background(), instance.ID)
	if err != nil || len(history) != 2 || history[0].Attempt != 1 || history[1].Attempt != 2 {
		t.Fatalf("history=%#v err=%v", history, err)
	}
}

func TestWorkflowSignalsAndTimersAreIdempotent(t *testing.T) {
	store, instance, _, now := workflowFixture(t)
	signal := application.WorkflowSignal{ID: "00000000-0000-0000-0000-000000001400", IdempotencyKey: "approval-1", InstanceID: instance.ID, Kind: "approved", Payload: json.RawMessage(`{"approved":true}`), ReceivedAt: now}
	created, err := store.PutWorkflowSignal(context.Background(), signal)
	if err != nil || !created {
		t.Fatalf("signal create=%v err=%v", created, err)
	}
	created, err = store.PutWorkflowSignal(context.Background(), signal)
	if err != nil || created {
		t.Fatalf("signal replay=%v err=%v", created, err)
	}
	timer := application.WorkflowTimer{ID: "00000000-0000-0000-0000-000000001401", InstanceID: instance.ID, StepID: "build", FireAt: now}
	created, err = store.PutWorkflowTimer(context.Background(), timer)
	if err != nil || !created {
		t.Fatalf("timer create=%v err=%v", created, err)
	}
	created, err = store.PutWorkflowTimer(context.Background(), timer)
	if err != nil || created {
		t.Fatalf("timer replay=%v err=%v", created, err)
	}
}
