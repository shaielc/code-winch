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
)

// deliveredCommand is what the runner side of the publisher was handed.
type deliveredCommand struct {
	runID   domain.RunID
	command domain.CommandID
	kind    application.InputKind
	payload application.InputPayload
}

// harnessInbox stands in for the restarted daemon's runner. Counting at this
// boundary is what makes "exactly once" observable: the engine's own delivery
// would report only the last outcome, not how many times it was asked.
type harnessInbox struct{ delivered []deliveredCommand }

func (h *harnessInbox) Deliver(_ context.Context, runID domain.RunID, command domain.CommandID, kind application.InputKind, payload application.InputPayload) error {
	h.delivered = append(h.delivered, deliveredCommand{runID, command, kind, payload})
	return nil
}

// restartClock is the outbox worker's wall clock. Lease expiry is measured
// against it, so advancing it is how the test reaches the moment a restarted
// daemon may reclaim the work a dead one left claimed.
type restartClock struct{ now time.Time }

func (c *restartClock) Now() time.Time { return c.now }

// TestInputAcceptedBeforeACrashIsDeliveredExactlyOnceAfterARestart covers the
// delivery half of the outbox contract against the real store: a command
// accepted durably outlives the daemon that accepted it, and the daemon that
// takes over delivers it once — not zero times, which loses an operator's
// input, and not twice, which runs it twice.
func TestInputAcceptedBeforeACrashIsDeliveredExactlyOnceAfterARestart(t *testing.T) {
	pool, store := database(t)
	ctx := context.Background()

	registry := application.NewDriverRegistry()
	registry.RegisterHarness("fake", fakeharness.Driver{})
	registry.RegisterSandbox("fake", fakesandbox.New(application.SandboxCapabilities{Isolation: "in-memory"}))
	ids := application.RandomIDs{}
	runs := application.NewRunService(store, store, ids, application.SystemClock{}, registry, nil)

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

	inputs := application.NewInputService(store, execution.Capabilities{Runs: store}, nil)
	accepted, err := inputs.Accept(ctx, application.InputRequest{
		CommandID: ids.NewCommandID(), RunID: id, IdempotencyKey: "before-the-crash",
		ActorID: "operator", Kind: application.InputText,
		Payload: application.InputPayload{Text: "echo hi"}, ExpectedState: domain.RunStateRunning,
	})
	if err != nil {
		t.Fatal(err)
	}

	// The daemon dies between acceptance and delivery: its worker claimed the
	// intent and never completed it, leaving a live lease behind.
	now, err := domain.NewTimestamp(time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	until, err := domain.NewTimestamp(now.Time().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	claimed, err := store.ClaimOutbox(ctx, "dead-daemon", ids.NewCommandID().String(), now, until, 10)
	if err != nil || len(claimed) != 1 {
		t.Fatalf("the crashed daemon claimed %d records: %v", len(claimed), err)
	}

	inbox := &harnessInbox{}
	clock := &restartClock{now: now.Time()}
	restarted, err := application.NewOutboxWorker(store, execution.Publisher{Commands: store, Input: inbox}, clock, application.OutboxWorkerConfig{
		WorkerID: "restarted-daemon", LeaseToken: ids.NewCommandID().String(), BatchSize: 10,
		LeaseDuration: time.Minute, BaseBackoff: time.Second, MaxBackoff: 8 * time.Second, MaxAttempts: 3,
	})
	if err != nil {
		t.Fatal(err)
	}

	// A restart that races the dead lease finds nothing to do, so it cannot
	// deliver a second copy alongside the daemon it is replacing.
	if delivered, runErr := restarted.RunOnce(ctx); runErr != nil || delivered != 0 {
		t.Fatalf("claimed %d records under a live lease: %v", delivered, runErr)
	}

	clock.now = now.Time().Add(2 * time.Minute)
	if delivered, runErr := restarted.RunOnce(ctx); runErr != nil || delivered != 1 {
		t.Fatalf("after the lease expired the restart claimed %d records: %v", delivered, runErr)
	}

	// The completion is durable, so a further restart delivers nothing more.
	if delivered, runErr := restarted.RunOnce(ctx); runErr != nil || delivered != 0 {
		t.Fatalf("a repeated pass claimed %d records: %v", delivered, runErr)
	}

	if len(inbox.delivered) != 1 {
		t.Fatalf("the runner was handed %d copies, want exactly 1", len(inbox.delivered))
	}
	got := inbox.delivered[0]
	if got.runID != id || got.command != accepted.CommandID || got.kind != application.InputText || got.payload.Text != "echo hi" {
		t.Fatalf("delivered run=%s command=%s kind=%s payload=%+v", got.runID, got.command, got.kind, got.payload)
	}

	var pending, completed int
	if err = pool.QueryRow(ctx, `SELECT count(*) FILTER (WHERE completed_at IS NULL), count(*) FILTER (WHERE completed_at IS NOT NULL) FROM outbox WHERE topic='run.input'`).Scan(&pending, &completed); err != nil {
		t.Fatal(err)
	}
	if pending != 0 || completed != 1 {
		t.Fatalf("input intents: pending=%d completed=%d, want 0 and 1", pending, completed)
	}
}
