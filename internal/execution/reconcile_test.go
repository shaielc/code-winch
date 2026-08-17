package execution_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shaielc/code-winch/internal/adapters/memory"
	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/internal/execution"
	"github.com/shaielc/code-winch/internal/supervisor"
	"github.com/shaielc/code-winch/pkg/protocol"
)

// lostRunner is the runner a restarted daemon has: it holds no executions, so
// every inspection reports the execution as unknown. local.Pool answers
// exactly this way after a restart, which is what makes the sweep necessary.
type lostRunner struct {
	inspected []string
	cleaned   []string
	adopt     *application.RunnerExecutionObservation
}

func (r *lostRunner) Inspect(_ context.Context, id string) (application.RunnerExecutionObservation, error) {
	r.inspected = append(r.inspected, id)
	if r.adopt != nil {
		return *r.adopt, nil
	}
	return application.RunnerExecutionObservation{ExecutionID: id, State: application.ExecutionUnknown}, nil
}
func (r *lostRunner) Takeover(context.Context, string, string, string) error { return nil }
func (r *lostRunner) Cleanup(_ context.Context, id string) error {
	r.cleaned = append(r.cleaned, id)
	return nil
}

// sweepFixture is a daemon's durable state without the daemon: the run
// repository and supervisor store survive, and nothing is live.
type sweepFixture struct {
	runs    *memory.RunRepository
	events  *memory.EventStore
	control *memory.SupervisorStore
	service *application.RunService
	runner  *lostRunner
	clock   *memory.Clock
	ids     application.RandomIDs
}

func newSweepFixture(t *testing.T) *sweepFixture {
	t.Helper()
	now, err := domain.NewTimestamp(time.Unix(1_700_000_000, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	f := &sweepFixture{
		runs:    &memory.RunRepository{},
		events:  &memory.EventStore{},
		control: &memory.SupervisorStore{},
		runner:  &lostRunner{},
		clock:   memory.NewClock(now),
	}
	f.service = application.NewRunService(f.runs, f.events, f.ids, f.clock, drivers(t), nil)
	return f
}

func (f *sweepFixture) reconciler(t *testing.T) *execution.Reconciler {
	t.Helper()
	s := supervisor.New(f.control, &silentGateway{}, execution.ClassificationRedactor{}, f.clock, "restarted-daemon", time.Minute).WithReconciliationIDs(f.ids)
	return execution.NewReconciler(f.runs, f.service, s, f.runner, f.ids, nil)
}

// interrupted leaves a run in state, as a daemon killed mid-run would: the
// attempt says it is in flight and no execution exists to back that up.
func (f *sweepFixture) interrupted(t *testing.T, state domain.RunState, executionID string) domain.RunID {
	t.Helper()
	view, err := f.service.CreateRun(context.Background(), application.CreateRunCommand{
		WorkspacePath: "/workspace", HarnessProfile: "fake", SandboxProfile: "fake",
	})
	if err != nil {
		t.Fatal(err)
	}
	id := view.Record.ID
	for _, c := range pathTo(state) {
		if _, err = f.service.Advance(context.Background(), id, c); err != nil {
			t.Fatalf("advance %s: %v", c, err)
		}
	}
	// The dead daemon's lease and execution are still recorded, and its lease
	// has long since expired.
	lease, err := f.control.AcquireRunLease(context.Background(), id, "dead-daemon", f.ids.NewCommandID().String(), f.clock.Now(), f.clock.Now())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.control.SaveDesiredState(context.Background(), lease, state, "fake", "fake", executionID); err != nil {
		t.Fatal(err)
	}
	return id
}

func pathTo(state domain.RunState) []domain.RunCommand {
	switch state {
	case domain.RunStateQueued:
		return []domain.RunCommand{domain.RunCommandStart}
	case domain.RunStatePreparing:
		return []domain.RunCommand{domain.RunCommandStart, domain.RunCommandAcquireLease}
	case domain.RunStateRunning:
		return []domain.RunCommand{domain.RunCommandStart, domain.RunCommandAcquireLease, domain.RunCommandExecutionStarted}
	case domain.RunStateStopping:
		return []domain.RunCommand{domain.RunCommandStart, domain.RunCommandAcquireLease, domain.RunCommandExecutionStarted, domain.RunCommandStop}
	}
	return nil
}

func (f *sweepFixture) state(t *testing.T, id domain.RunID) domain.RunState {
	t.Helper()
	record, _, err := f.runs.Get(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	return record.Attempts[len(record.Attempts)-1].State
}

// silentGateway satisfies the supervisor's runner gateway. The sweep reaches it
// only on the stop-continuation path, which a lost execution never takes.
type silentGateway struct{}

func (silentGateway) Send(context.Context, protocol.RunnerMessage) error { return nil }

// A run interrupted by a daemon kill must not still claim to be in flight
// after the daemon comes back. This is acceptance criterion 6.
func TestSweepSettlesRunsInterruptedByADaemonKill(t *testing.T) {
	for name, tc := range map[string]struct {
		state domain.RunState
		want  domain.RunState
	}{
		"running":   {domain.RunStateRunning, domain.RunStateFailed},
		"preparing": {domain.RunStatePreparing, domain.RunStateFailed},
		"stopping":  {domain.RunStateStopping, domain.RunStateFailed},
		"queued":    {domain.RunStateQueued, domain.RunStateCancelled},
	} {
		t.Run(name, func(t *testing.T) {
			f := newSweepFixture(t)
			id := f.interrupted(t, tc.state, "execution-"+name)
			if got := f.state(t, id); got != tc.state {
				t.Fatalf("before the sweep state = %s, want %s", got, tc.state)
			}

			settled, err := f.reconciler(t).Sweep(context.Background())
			if err != nil {
				t.Fatalf("sweep: %v", err)
			}
			if settled != 1 {
				t.Fatalf("settled = %d, want 1", settled)
			}
			got := f.state(t, id)
			if got != tc.want {
				t.Fatalf("state = %s, want %s", got, tc.want)
			}
			if !got.IsTerminal() {
				t.Fatalf("state = %s, want terminal", got)
			}
		})
	}
}

// The sweep is answerable only for runs in flight. A run waiting for a client
// to start it, and a run that already finished, are both already truthful.
func TestSweepLeavesCreatedAndTerminalRunsAlone(t *testing.T) {
	f := newSweepFixture(t)
	created, err := f.service.CreateRun(context.Background(), application.CreateRunCommand{
		WorkspacePath: "/workspace", HarnessProfile: "fake", SandboxProfile: "fake",
	})
	if err != nil {
		t.Fatal(err)
	}
	finished := f.interrupted(t, domain.RunStateRunning, "execution-finished")
	if _, err = f.service.Advance(context.Background(), finished, domain.RunCommandSuccessfulExit); err != nil {
		t.Fatal(err)
	}

	settled, err := f.reconciler(t).Sweep(context.Background())
	if err != nil {
		t.Fatalf("sweep: %v", err)
	}
	if settled != 0 {
		t.Fatalf("settled = %d, want 0", settled)
	}
	if got := f.state(t, created.Record.ID); got != domain.RunStateCreated {
		t.Fatalf("created run moved to %s", got)
	}
	if got := f.state(t, finished); got != domain.RunStateCompleted {
		t.Fatalf("completed run moved to %s", got)
	}
	if len(f.runner.inspected) != 0 {
		t.Fatalf("the sweep inspected %v", f.runner.inspected)
	}
}

// An execution the runner still holds is adopted, not failed. No local runner
// can answer this way after a restart; the branch is the seam a remote runner
// grows into, and it must not be a lie in the other direction.
func TestSweepLeavesAnAdoptedExecutionRunning(t *testing.T) {
	f := newSweepFixture(t)
	id := f.interrupted(t, domain.RunStateRunning, "execution-live")
	control, err := f.control.LoadControl(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	f.runner.adopt = &application.RunnerExecutionObservation{
		ExecutionID: "execution-live", OwnershipToken: control.LeaseToken, State: application.ExecutionRunning,
	}

	settled, err := f.reconciler(t).Sweep(context.Background())
	if err != nil {
		t.Fatalf("sweep: %v", err)
	}
	if settled != 0 {
		t.Fatalf("settled = %d, want 0: an adopted run is not settled", settled)
	}
	if got := f.state(t, id); got != domain.RunStateRunning {
		t.Fatalf("state = %s, want running", got)
	}
}

// One run that cannot be reconciled must not stop the daemon telling the truth
// about the others.
func TestSweepContinuesPastOneFailure(t *testing.T) {
	f := newSweepFixture(t)
	first := f.interrupted(t, domain.RunStateRunning, "execution-one")
	second := f.interrupted(t, domain.RunStateRunning, "execution-two")
	// The first run the sweep reaches cannot be read back for its attempt.
	f.runs.Failures.Inject("get", nil, nil, errors.New("storage unavailable"))

	settled, err := f.reconciler(t).Sweep(context.Background())
	if err != nil {
		t.Fatalf("sweep: %v", err)
	}
	if settled != 1 {
		t.Fatalf("settled = %d, want 1", settled)
	}
	states := []domain.RunState{f.state(t, first), f.state(t, second)}
	failed := 0
	for _, s := range states {
		if s == domain.RunStateFailed {
			failed++
		}
	}
	if failed != 1 {
		t.Fatalf("states = %v, want exactly one failed", states)
	}
}

// Listing is the one failure that leaves nothing reconciled, so it is the one
// the daemon refuses to start on.
func TestSweepReportsAListingFailure(t *testing.T) {
	f := newSweepFixture(t)
	f.runs.Failures.Inject("in_flight", errors.New("storage unavailable"))
	if _, err := f.reconciler(t).Sweep(context.Background()); err == nil {
		t.Fatal("a listing failure was not reported")
	}
}
