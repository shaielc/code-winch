//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	fakeharness "github.com/shaielc/code-winch/internal/adapters/harness/fake"
	fakesandbox "github.com/shaielc/code-winch/internal/adapters/sandbox/fake"
	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/internal/execution"
	"github.com/shaielc/code-winch/internal/supervisor"
	"github.com/shaielc/code-winch/pkg/protocol"
)

// lostRunner is what a restarted daemon's runner pool answers: it holds no
// executions, so every inspection reports the execution as unknown.
type lostRunner struct{}

func (lostRunner) Inspect(_ context.Context, id string) (application.RunnerExecutionObservation, error) {
	return application.RunnerExecutionObservation{ExecutionID: id, State: application.ExecutionUnknown}, nil
}
func (lostRunner) Takeover(context.Context, string, string, string) error { return nil }
func (lostRunner) Cleanup(context.Context, string) error                  { return nil }

type silentGateway struct{}

func (silentGateway) Send(context.Context, protocol.RunnerMessage) error { return nil }

// TestStartupReconciliationSettlesAnInterruptedRun is acceptance criterion 6
// against the real database: a run whose daemon was killed mid-flight must not
// still report that it is running, and it must stop accepting input that no
// execution will ever receive.
func TestStartupReconciliationSettlesAnInterruptedRun(t *testing.T) {
	_, store := database(t)
	ctx := context.Background()

	registry := application.NewDriverRegistry()
	registry.RegisterHarness("fake", fakeharness.Driver{})
	registry.RegisterSandbox("fake", fakesandbox.New(application.SandboxCapabilities{Isolation: "in-memory"}))
	ids := application.RandomIDs{}
	clock := application.SystemClock{}
	runs := application.NewRunService(store, store, ids, clock, registry, nil)

	// The daemon that died: it created a run, drove it to running, and left a
	// lease and an execution behind.
	view, err := runs.CreateRun(ctx, application.CreateRunCommand{
		WorkspacePath: "/workspace", HarnessProfile: "fake", SandboxProfile: "fake",
	})
	if err != nil {
		t.Fatal(err)
	}
	id := view.Record.ID
	for _, c := range []domain.RunCommand{domain.RunCommandStart, domain.RunCommandAcquireLease, domain.RunCommandExecutionStarted} {
		if _, err = runs.Advance(ctx, id, c); err != nil {
			t.Fatalf("advance %s: %v", c, err)
		}
	}
	now, err := domain.NewTimestamp(time.Now().UTC().Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	expired, err := domain.NewTimestamp(time.Now().UTC().Add(-time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	lease, err := store.AcquireRunLease(ctx, id, "dead-daemon", ids.NewCommandID().String(), now, expired)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.SaveDesiredState(ctx, lease, domain.RunStateRunning, "fake", "fake", "execution-orphan"); err != nil {
		t.Fatal(err)
	}

	// Before the sweep the run lies to every reader.
	before, err := runs.GetRun(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if got := before.Record.Attempts[0].State; got != domain.RunStateRunning {
		t.Fatalf("before the sweep state = %s, want running", got)
	}
	inFlight, err := store.InFlight(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(inFlight) != 1 || inFlight[0] != id {
		t.Fatalf("in flight = %v, want [%s]", inFlight, id)
	}

	// The daemon restarts.
	restarted := supervisor.New(store, silentGateway{}, execution.ClassificationRedactor{}, clock, "restarted-daemon", time.Minute).WithReconciliationIDs(ids)
	settled, err := execution.NewReconciler(store, runs, restarted, lostRunner{}, ids, nil).Sweep(ctx)
	if err != nil {
		t.Fatalf("sweep: %v", err)
	}
	if settled != 1 {
		t.Fatalf("settled = %d, want 1", settled)
	}

	after, err := runs.GetRun(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	state := after.Record.Attempts[len(after.Record.Attempts)-1].State
	if state != domain.RunStateFailed {
		t.Fatalf("state = %s, want failed", state)
	}

	// Nothing is in flight any more, so a second sweep has no work.
	if inFlight, err = store.InFlight(ctx); err != nil {
		t.Fatal(err)
	}
	if len(inFlight) != 0 {
		t.Fatalf("in flight after the sweep = %v, want none", inFlight)
	}

	// The reconciliation is durable and visible in the run's history.
	events, err := store.Read(ctx, id, 0, 100)
	if err != nil {
		t.Fatal(err)
	}
	reconciled := false
	for _, e := range events {
		if e.Kind == "run.reconciled" {
			reconciled = true
		}
	}
	if !reconciled {
		t.Fatal("no run.reconciled event was recorded")
	}

	// Input is refused now that the run is terminal, rather than accepted and
	// retried against an execution that does not exist.
	inputs := application.NewInputService(store, execution.Capabilities{Runs: store}, nil)
	_, err = inputs.Accept(ctx, application.InputRequest{
		CommandID: ids.NewCommandID(), RunID: id, IdempotencyKey: "after-reconcile",
		ActorID: "operator", Kind: application.InputText,
		Payload: application.InputPayload{Text: "hello"}, ExpectedState: domain.RunStateRunning,
	})
	if err == nil {
		t.Fatal("input was accepted for a run reconciliation had already failed")
	}
}

// InFlight must select on the run's latest attempt, not on any attempt, or a
// retried run would be listed forever by its failed first attempt.
func TestInFlightSelectsOnTheLatestAttempt(t *testing.T) {
	_, store := database(t)
	ctx := context.Background()
	runID := id(t, domain.ParseRunID, 900)
	first := id(t, domain.ParseAttemptID, 901)
	second := id(t, domain.ParseAttemptID, 902)

	record := application.RunRecord{ID: runID, Attempts: []domain.Attempt{{ID: first, State: domain.RunStateFailed}}}
	version, err := store.Save(ctx, record, 0)
	if err != nil {
		t.Fatal(err)
	}
	inFlight, err := store.InFlight(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(inFlight) != 0 {
		t.Fatalf("a failed run is in flight: %v", inFlight)
	}

	record.Attempts = append(record.Attempts, domain.Attempt{ID: second, PreviousAttemptID: first, State: domain.RunStateRunning})
	if _, err = store.Save(ctx, record, version); err != nil {
		t.Fatal(err)
	}
	if inFlight, err = store.InFlight(ctx); err != nil {
		t.Fatal(err)
	}
	if len(inFlight) != 1 || inFlight[0] != runID {
		t.Fatalf("in flight = %v, want [%s]", inFlight, runID)
	}
}
