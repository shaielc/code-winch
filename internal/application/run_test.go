package application_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	fakeharness "github.com/shaielc/code-winch/internal/adapters/harness/fake"
	"github.com/shaielc/code-winch/internal/adapters/memory"
	fakesandbox "github.com/shaielc/code-winch/internal/adapters/sandbox/fake"
	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
)

// service builds a RunService over the in-memory ports with no executor, so
// these tests observe the persistence half of each use case on its own. The
// runtime half is covered by internal/execution.
type fixture struct {
	service *application.RunService
	runs    *memory.RunRepository
	events  *memory.EventStore
	clock   *memory.Clock
	metrics *recordedMetrics
}

func newFixture(t *testing.T, admission application.AdmissionHook) *fixture {
	t.Helper()
	now, err := domain.NewTimestamp(time.Unix(1_700_000_000, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	registry := application.NewDriverRegistry()
	registry.RegisterHarness("fake", fakeharness.Driver{})
	registry.RegisterSandbox("fake", fakesandbox.New(application.SandboxCapabilities{Isolation: "in-memory"}))
	runs := &memory.RunRepository{}
	events := &memory.EventStore{}
	clock := memory.NewClock(now)
	metrics := &recordedMetrics{}
	return &fixture{
		service: application.NewRunService(runs, events, application.RandomIDs{}, clock, registry, admission).WithMetrics(metrics),
		runs:    runs,
		events:  events,
		clock:   clock,
		metrics: metrics,
	}
}

// advance moves the fixture's clock, so a duration a measurement reports is the
// one the test chose rather than however long the test itself took.
func (f *fixture) advance(t *testing.T, d time.Duration) {
	t.Helper()
	next, err := domain.NewTimestamp(f.clock.Now().Time().Add(d))
	if err != nil {
		t.Fatal(err)
	}
	f.clock.Set(next)
}

type queueObservation struct {
	id     domain.RunID
	state  domain.RunState
	waited time.Duration
}

// recordedMetrics captures what the service observed. The lifecycle tests race
// starts against stops, so it is safe for concurrent use.
type recordedMetrics struct {
	mu     sync.Mutex
	queued []queueObservation
	active []map[domain.RunState]int
}

func (m *recordedMetrics) QueueTime(_ context.Context, id domain.RunID, state domain.RunState, waited time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.queued = append(m.queued, queueObservation{id, state, waited})
}

func (m *recordedMetrics) ActiveRuns(_ context.Context, counts map[domain.RunState]int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	copied := make(map[domain.RunState]int, len(counts))
	for state, n := range counts {
		copied[state] = n
	}
	m.active = append(m.active, copied)
}

func (m *recordedMetrics) queueTimes() []queueObservation {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]queueObservation(nil), m.queued...)
}

// lastActive is the most recent gauge reading, and false when none was taken.
func (m *recordedMetrics) lastActive() (map[domain.RunState]int, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.active) == 0 {
		return nil, false
	}
	return m.active[len(m.active)-1], true
}

func (m *recordedMetrics) activeReadings() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.active)
}

func (f *fixture) create(t *testing.T) application.RunView {
	t.Helper()
	view, err := f.service.CreateRun(context.Background(), application.CreateRunCommand{
		WorkspacePath: "/workspace", HarnessProfile: "fake", SandboxProfile: "fake",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	return view
}

func (f *fixture) state(t *testing.T, id domain.RunID) domain.RunState {
	t.Helper()
	record, _, err := f.runs.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(record.Attempts) == 0 {
		t.Fatal("run has no attempt")
	}
	return record.Attempts[len(record.Attempts)-1].State
}

// TestLifecycleTransitionsAreReachableThroughTheService walks every edge of
// docs/contracts.md §1 that Phase 1 reaches. Client-requested edges go through
// StartRun and StopRun; edges the runtime observes go through Advance. The
// cancel edges belong to P2-059 and the retry edge to P2-057, so neither has a
// service entry point yet.
func TestLifecycleTransitionsAreReachableThroughTheService(t *testing.T) {
	type step struct {
		apply func(*fixture, domain.RunID, uint64) (application.RunView, error)
		want  domain.RunState
	}
	start := func(f *fixture, id domain.RunID, v uint64) (application.RunView, error) {
		return f.service.StartRun(context.Background(), id, v)
	}
	stop := func(f *fixture, id domain.RunID, v uint64) (application.RunView, error) {
		return f.service.StopRun(context.Background(), id, v)
	}
	advance := func(c domain.RunCommand) func(*fixture, domain.RunID, uint64) (application.RunView, error) {
		return func(f *fixture, id domain.RunID, _ uint64) (application.RunView, error) {
			return f.service.Advance(context.Background(), id, c)
		}
	}

	for name, path := range map[string][]step{
		"preparation failed": {
			{start, domain.RunStateQueued},
			{advance(domain.RunCommandAcquireLease), domain.RunStatePreparing},
			{advance(domain.RunCommandPreparationFailed), domain.RunStateFailed},
		},
		"successful exit": {
			{start, domain.RunStateQueued},
			{advance(domain.RunCommandAcquireLease), domain.RunStatePreparing},
			{advance(domain.RunCommandExecutionStarted), domain.RunStateRunning},
			{advance(domain.RunCommandSuccessfulExit), domain.RunStateCompleted},
		},
		"failed exit": {
			{start, domain.RunStateQueued},
			{advance(domain.RunCommandAcquireLease), domain.RunStatePreparing},
			{advance(domain.RunCommandExecutionStarted), domain.RunStateRunning},
			{advance(domain.RunCommandFailedExit), domain.RunStateFailed},
		},
		"stop then expected exit": {
			{start, domain.RunStateQueued},
			{advance(domain.RunCommandAcquireLease), domain.RunStatePreparing},
			{advance(domain.RunCommandExecutionStarted), domain.RunStateRunning},
			{stop, domain.RunStateStopping},
			{advance(domain.RunCommandSuccessfulExit), domain.RunStateCompleted},
		},
		"stop then forced-stop failure": {
			{start, domain.RunStateQueued},
			{advance(domain.RunCommandAcquireLease), domain.RunStatePreparing},
			{advance(domain.RunCommandExecutionStarted), domain.RunStateRunning},
			{stop, domain.RunStateStopping},
			{advance(domain.RunCommandFailedExit), domain.RunStateFailed},
		},
	} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t, nil)
			view := f.create(t)
			if got := f.state(t, view.Record.ID); got != domain.RunStateCreated {
				t.Fatalf("created state = %s", got)
			}
			version := view.Version
			for i, s := range path {
				next, err := s.apply(f, view.Record.ID, version)
				if err != nil {
					t.Fatalf("step %d to %s: %v", i, s.want, err)
				}
				version = next.Version
				if got := f.state(t, view.Record.ID); got != s.want {
					t.Fatalf("step %d: state = %s, want %s", i, got, s.want)
				}
			}
		})
	}
}

// A stop the domain accepts without changing anything must not be written.
// Persisting it would move the version, and the version is the ETag a client
// holds; docs/contracts.md §1 makes terminal states immutable.
func TestRedundantStopLeavesTheVersionAndTheRecordAlone(t *testing.T) {
	for _, from := range []domain.RunCommand{domain.RunCommandSuccessfulExit, domain.RunCommandFailedExit} {
		f := newFixture(t, nil)
		view := f.create(t)
		id := view.Record.ID
		version := view.Version
		for _, c := range []domain.RunCommand{domain.RunCommandStart, domain.RunCommandAcquireLease, domain.RunCommandExecutionStarted, domain.RunCommandStop, from} {
			next, err := f.service.Advance(context.Background(), id, c)
			if err != nil {
				t.Fatalf("advance %s: %v", c, err)
			}
			version = next.Version
		}
		terminal := f.state(t, id)
		if !terminal.IsTerminal() {
			t.Fatalf("state = %s, want terminal", terminal)
		}
		writes := len(f.runs.SaveCalls())

		for i := range 3 {
			stopped, err := f.service.StopRun(context.Background(), id, version)
			if err != nil {
				t.Fatalf("stop %d of a %s run: %v", i, terminal, err)
			}
			if stopped.Version != version {
				t.Fatalf("stop %d of a %s run moved the version %d -> %d", i, terminal, version, stopped.Version)
			}
			if got := f.state(t, id); got != terminal {
				t.Fatalf("stop %d changed the state %s -> %s", i, terminal, got)
			}
		}
		if got := len(f.runs.SaveCalls()); got != writes {
			t.Fatalf("%d redundant stops wrote %d times", 3, got-writes)
		}
	}
}

// Stopping accepts a stop for the same reason, and must not move the version
// either: the run is already going down and the client's ETag is still good.
func TestRedundantStopInStoppingLeavesTheVersionAlone(t *testing.T) {
	f := newFixture(t, nil)
	view := f.create(t)
	id := view.Record.ID
	version := view.Version
	for _, c := range []domain.RunCommand{domain.RunCommandStart, domain.RunCommandAcquireLease, domain.RunCommandExecutionStarted} {
		next, err := f.service.Advance(context.Background(), id, c)
		if err != nil {
			t.Fatalf("advance %s: %v", c, err)
		}
		version = next.Version
	}
	stopped, err := f.service.StopRun(context.Background(), id, version)
	if err != nil {
		t.Fatalf("first stop: %v", err)
	}
	if stopped.Version == version {
		t.Fatal("the first stop did not move the version")
	}
	again, err := f.service.StopRun(context.Background(), id, stopped.Version)
	if err != nil {
		t.Fatalf("second stop: %v", err)
	}
	if again.Version != stopped.Version {
		t.Fatalf("the second stop moved the version %d -> %d", stopped.Version, again.Version)
	}
	if got := f.state(t, id); got != domain.RunStateStopping {
		t.Fatalf("state = %s, want stopping", got)
	}
}

// A retried run has more than one attempt. Rebuilding it must reproduce the
// stored attempt chain rather than fail, because every lifecycle command loads
// the run this way. P2-057 adds the entry point that creates such a record;
// the reconstruction has to hold before it does.
func TestARetriedRunIsRebuiltFromItsStoredAttempts(t *testing.T) {
	f := newFixture(t, nil)
	view := f.create(t)
	id := view.Record.ID
	record, version, err := f.runs.Get(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	first := record.Attempts[0].ID
	second := application.RandomIDs{}.NewAttemptID()
	record.Attempts = []domain.Attempt{
		{ID: first, State: domain.RunStateFailed},
		{ID: second, PreviousAttemptID: first, State: domain.RunStateRunning},
	}
	if version, err = f.runs.Save(context.Background(), record, version); err != nil {
		t.Fatal(err)
	}

	stopped, err := f.service.StopRun(context.Background(), id, version)
	if err != nil {
		t.Fatalf("stop a retried run: %v", err)
	}
	if got := f.state(t, id); got != domain.RunStateStopping {
		t.Fatalf("state = %s, want stopping", got)
	}
	if got := stopped.Record.Attempts; len(got) != 2 ||
		got[0].ID != first || got[0].State != domain.RunStateFailed ||
		got[1].ID != second || got[1].PreviousAttemptID != first {
		t.Fatalf("attempt chain not preserved: %+v", got)
	}
}

// A record whose earlier attempt is not failed cannot have been produced by
// retry, so it is reported rather than quietly repaired into a valid one.
func TestACorruptAttemptChainIsRejected(t *testing.T) {
	f := newFixture(t, nil)
	view := f.create(t)
	id := view.Record.ID
	record, version, err := f.runs.Get(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	first := record.Attempts[0].ID
	record.Attempts = []domain.Attempt{
		{ID: first, State: domain.RunStateCompleted},
		{ID: application.RandomIDs{}.NewAttemptID(), PreviousAttemptID: first, State: domain.RunStateRunning},
	}
	if version, err = f.runs.Save(context.Background(), record, version); err != nil {
		t.Fatal(err)
	}
	if _, err = f.service.StopRun(context.Background(), id, version); err == nil {
		t.Fatal("a completed attempt followed by another was accepted")
	}
}

// Concurrent start and stop must leave one outcome, not a mixture: the run is
// queued, exactly one call succeeded, and the loser was refused.
func TestConcurrentStartAndStopResolveToOneOutcome(t *testing.T) {
	f := newFixture(t, nil)
	view := f.create(t)
	id := view.Record.ID

	var wg sync.WaitGroup
	errs := make([]error, 2)
	wg.Add(2)
	go func() { defer wg.Done(); _, errs[0] = f.service.StartRun(context.Background(), id, view.Version) }()
	go func() { defer wg.Done(); _, errs[1] = f.service.StopRun(context.Background(), id, view.Version) }()
	wg.Wait()

	if errs[0] != nil {
		t.Fatalf("start: %v", errs[0])
	}
	if errs[1] == nil {
		t.Fatal("stop from created was accepted")
	}
	if got := f.state(t, id); got != domain.RunStateQueued {
		t.Fatalf("state = %s, want queued", got)
	}
}

// Two clients holding the same ETag race one stop. One wins; the other is told
// its precondition failed rather than silently applying a second time.
func TestConcurrentStopsResolveToOneWinner(t *testing.T) {
	f := newFixture(t, nil)
	view := f.create(t)
	id := view.Record.ID
	version := view.Version
	for _, c := range []domain.RunCommand{domain.RunCommandStart, domain.RunCommandAcquireLease, domain.RunCommandExecutionStarted} {
		next, err := f.service.Advance(context.Background(), id, c)
		if err != nil {
			t.Fatalf("advance %s: %v", c, err)
		}
		version = next.Version
	}

	const racers = 8
	var wg sync.WaitGroup
	errs := make([]error, racers)
	wg.Add(racers)
	for i := range racers {
		go func() { defer wg.Done(); _, errs[i] = f.service.StopRun(context.Background(), id, version) }()
	}
	wg.Wait()

	accepted := 0
	for i, err := range errs {
		switch {
		case err == nil:
			accepted++
		case errors.Is(err, application.ErrConflict):
		default:
			t.Fatalf("racer %d: %v", i, err)
		}
	}
	if accepted != 1 {
		t.Fatalf("%d of %d concurrent stops were accepted, want 1", accepted, racers)
	}
	if got := f.state(t, id); got != domain.RunStateStopping {
		t.Fatalf("state = %s, want stopping", got)
	}
}

// An unknown profile is refused when the run is created, so no run can carry
// one. The start path checks again because a profile can leave the registry
// between create and start.
func TestUnknownProfileIsRejectedBeforeLaunch(t *testing.T) {
	f := newFixture(t, nil)
	for name, command := range map[string]application.CreateRunCommand{
		"harness": {WorkspacePath: "/workspace", HarnessProfile: "missing", SandboxProfile: "fake"},
		"sandbox": {WorkspacePath: "/workspace", HarnessProfile: "fake", SandboxProfile: "missing"},
		"path":    {WorkspacePath: "", HarnessProfile: "fake", SandboxProfile: "fake"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := f.service.CreateRun(context.Background(), command); !errors.Is(err, application.ErrUnknownProfile) {
				t.Fatalf("err = %v, want ErrUnknownProfile", err)
			}
		})
	}

	// A run created against a registry that later loses the driver is refused
	// at start, before anything is launched.
	view := f.create(t)
	empty := newFixture(t, nil)
	record, _, err := f.runs.Get(context.Background(), view.Record.ID)
	if err != nil {
		t.Fatal(err)
	}
	record.HarnessProfile = "retired"
	version, err := empty.runs.Save(context.Background(), record, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = empty.service.StartRun(context.Background(), record.ID, version); !errors.Is(err, application.ErrUnknownProfile) {
		t.Fatalf("err = %v, want ErrUnknownProfile", err)
	}
	if got := empty.state(t, record.ID); got != domain.RunStateCreated {
		t.Fatalf("state = %s, want created: the run must not move", got)
	}
}

// The admission hook decides on the start path. Its default is permissive, and
// a refusal stops the run before any state is written — P2-059 supplies the
// policy, not a new branch in StartRun.
func TestAdmissionRefusalStopsStartBeforeAnyWrite(t *testing.T) {
	refused := errors.New("admission: at capacity")
	var seen application.RunView
	f := newFixture(t, application.AdmissionFunc(func(_ context.Context, v application.RunView) error {
		seen = v
		return refused
	}))
	view := f.create(t)
	writes := len(f.runs.SaveCalls())

	if _, err := f.service.StartRun(context.Background(), view.Record.ID, view.Version); !errors.Is(err, refused) {
		t.Fatalf("err = %v, want the admission refusal", err)
	}
	if got := f.state(t, view.Record.ID); got != domain.RunStateCreated {
		t.Fatalf("state = %s, want created", got)
	}
	if got := len(f.runs.SaveCalls()); got != writes {
		t.Fatalf("a refused start wrote %d times", got-writes)
	}
	if seen.Record.ID != view.Record.ID {
		t.Fatalf("admission saw run %s, want %s", seen.Record.ID, view.Record.ID)
	}
}

func TestPermissiveAdmissionIsTheDefault(t *testing.T) {
	f := newFixture(t, nil)
	view := f.create(t)
	if _, err := f.service.StartRun(context.Background(), view.Record.ID, view.Version); err != nil {
		t.Fatalf("start with no admission hook: %v", err)
	}
	if got := f.state(t, view.Record.ID); got != domain.RunStateQueued {
		t.Fatalf("state = %s, want queued", got)
	}
}

// A stale ETag is a conflict, not a second application of the command.
func TestAStaleVersionIsRefused(t *testing.T) {
	f := newFixture(t, nil)
	view := f.create(t)
	if _, err := f.service.StartRun(context.Background(), view.Record.ID, view.Version+7); !errors.Is(err, application.ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
	if got := f.state(t, view.Record.ID); got != domain.RunStateCreated {
		t.Fatalf("state = %s, want created", got)
	}
}

func TestAMissingRunIsNotFound(t *testing.T) {
	f := newFixture(t, nil)
	missing := application.RandomIDs{}.NewRunID()
	if _, err := f.service.GetRun(context.Background(), missing); !errors.Is(err, application.ErrNotFound) {
		t.Fatalf("get: err = %v, want ErrNotFound", err)
	}
	if _, err := f.service.StartRun(context.Background(), missing, 1); !errors.Is(err, application.ErrNotFound) {
		t.Fatalf("start: err = %v, want ErrNotFound", err)
	}
	if _, err := f.service.ListRunEvents(context.Background(), missing, 0, 10); !errors.Is(err, application.ErrNotFound) {
		t.Fatalf("events: err = %v, want ErrNotFound", err)
	}
}

// The resolved profile is stored for reproducibility and must carry no secret.
func TestCreateStoresTheResolvedProfile(t *testing.T) {
	f := newFixture(t, nil)
	view := f.create(t)
	record, _, err := f.runs.Get(context.Background(), view.Record.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(record.ResolvedConfiguration) == 0 {
		t.Fatal("no resolved configuration stored")
	}
	for _, forbidden := range []string{"token", "secret", "password", "credential"} {
		if bytesContain(record.ResolvedConfiguration, forbidden) {
			t.Fatalf("resolved configuration mentions %q: %s", forbidden, record.ResolvedConfiguration)
		}
	}
}

// GetRun reports how far a run's history goes without reading that history.
// Counting the events instead made a hot path — StartRun, StopRun and every
// input all pass through here — cost O(history), and the count silently
// stopped moving once the history outgrew the read limit.
func TestGetRunReportsTheLastSequenceWithoutReadingTheHistory(t *testing.T) {
	f := newFixture(t, nil)
	view := f.create(t)
	id := view.Record.ID

	ids := application.RandomIDs{}
	now, err := domain.NewTimestamp(time.Unix(1_700_000_000, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	const committed = 3
	values := make([]application.UnsequencedEvent, committed)
	for i := range values {
		values[i] = application.UnsequencedEvent{
			EventID: ids.NewEventID(), OccurredAt: now, Kind: "run.lifecycle", SchemaVersion: 1,
		}
	}
	if _, err = f.events.Append(context.Background(), id, 0, values); err != nil {
		t.Fatalf("append: %v", err)
	}

	// From here any Read fails, so a GetRun that still reads cannot pass.
	f.events.Failures.Inject("read", errors.New("GetRun read the history"))
	got, err := f.service.GetRun(context.Background(), id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.LastSequence != committed {
		t.Fatalf("lastSequence = %d, want %d", got.LastSequence, committed)
	}

	// The injected failure is still queued, proving GetRun never consumed it.
	if _, err = f.service.ListRunEvents(context.Background(), id, 0, 10); err == nil {
		t.Fatal("the read failure was consumed by GetRun rather than by a read")
	}
}

// docs/architecture.md §7 asks for queue time. It is the wait a run spends in
// Queued before anything picks it up, so it is measured when the run leaves
// Queued and not before: a run still queued has no queue time yet.
func TestQueueTimeIsMeasuredWhenTheRunLeavesQueued(t *testing.T) {
	const queued = 7 * time.Second
	for name, leave := range map[string]struct {
		command domain.RunCommand
		want    domain.RunState
	}{
		"lease acquired": {domain.RunCommandAcquireLease, domain.RunStatePreparing},
		"cancelled":      {domain.RunCommandCancel, domain.RunStateCancelled},
	} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t, nil)
			view := f.create(t)
			id := view.Record.ID
			if _, err := f.service.StartRun(context.Background(), id, view.Version); err != nil {
				t.Fatalf("start: %v", err)
			}
			if got := f.metrics.queueTimes(); len(got) != 0 {
				t.Fatalf("a still-queued run reported %d queue times", len(got))
			}

			f.advance(t, queued)
			if _, err := f.service.Advance(context.Background(), id, leave.command); err != nil {
				t.Fatalf("advance %s: %v", leave.command, err)
			}
			got := f.metrics.queueTimes()
			if len(got) != 1 {
				t.Fatalf("%d queue times reported, want 1", len(got))
			}
			if got[0].id != id {
				t.Fatalf("queue time reported for run %s, want %s", got[0].id, id)
			}
			if got[0].waited != queued {
				t.Fatalf("waited = %s, want %s", got[0].waited, queued)
			}
			if got[0].state != leave.want {
				t.Fatalf("state = %s, want %s", got[0].state, leave.want)
			}

			// Later transitions are not queue time: a second reading would
			// make the metric mean "time in the previous state".
			f.advance(t, queued)
			_, _ = f.service.Advance(context.Background(), id, domain.RunCommandPreparationFailed)
			if got = f.metrics.queueTimes(); len(got) != 1 {
				t.Fatalf("%d queue times after leaving Queued, want 1", len(got))
			}
		})
	}
}

// The active-runs gauge counts only the states a daemon holds an execution in,
// and it counts them from the store, so a reading is the truth about every run
// rather than about the one that just moved.
func TestActiveRunsCountsOnlyRunningAndStoppingRuns(t *testing.T) {
	f := newFixture(t, nil)
	first := f.create(t).Record.ID
	second := f.create(t).Record.ID

	drive := func(id domain.RunID, commands ...domain.RunCommand) {
		t.Helper()
		for _, c := range commands {
			if _, err := f.service.Advance(context.Background(), id, c); err != nil {
				t.Fatalf("advance %s: %v", c, err)
			}
		}
	}

	// Nothing has reached an execution yet, so no reading has been taken.
	drive(first, domain.RunCommandStart, domain.RunCommandAcquireLease)
	if got, taken := f.metrics.lastActive(); taken {
		t.Fatalf("a queued and a preparing run reported %v", got)
	}

	for _, step := range []struct {
		commands []domain.RunCommand
		id       domain.RunID
		want     map[domain.RunState]int
	}{
		{[]domain.RunCommand{domain.RunCommandExecutionStarted}, first,
			map[domain.RunState]int{domain.RunStateRunning: 1, domain.RunStateStopping: 0}},
		{[]domain.RunCommand{domain.RunCommandStart, domain.RunCommandAcquireLease, domain.RunCommandExecutionStarted}, second,
			map[domain.RunState]int{domain.RunStateRunning: 2, domain.RunStateStopping: 0}},
		{[]domain.RunCommand{domain.RunCommandStop}, second,
			map[domain.RunState]int{domain.RunStateRunning: 1, domain.RunStateStopping: 1}},
		{[]domain.RunCommand{domain.RunCommandSuccessfulExit}, second,
			map[domain.RunState]int{domain.RunStateRunning: 1, domain.RunStateStopping: 0}},
		{[]domain.RunCommand{domain.RunCommandFailedExit}, first,
			map[domain.RunState]int{domain.RunStateRunning: 0, domain.RunStateStopping: 0}},
	} {
		drive(step.id, step.commands...)
		got, taken := f.metrics.lastActive()
		if !taken {
			t.Fatalf("no reading after %v", step.commands)
		}
		if len(got) != len(step.want) {
			t.Fatalf("after %v the reading is %v, want %v", step.commands, got, step.want)
		}
		for state, want := range step.want {
			if got[state] != want {
				t.Fatalf("after %v the reading is %v, want %v", step.commands, got, step.want)
			}
		}
	}
}

// A refused or conflicting command changes nothing, so it must measure
// nothing: a metric that moves on a rejected request reports work that never
// happened.
func TestARefusedCommandReportsNoMeasurement(t *testing.T) {
	f := newFixture(t, nil)
	view := f.create(t)
	id := view.Record.ID
	for _, c := range []domain.RunCommand{domain.RunCommandStart, domain.RunCommandAcquireLease, domain.RunCommandExecutionStarted} {
		if _, err := f.service.Advance(context.Background(), id, c); err != nil {
			t.Fatalf("advance %s: %v", c, err)
		}
	}
	readings := f.metrics.activeReadings()
	before := len(f.metrics.queueTimes())

	if _, err := f.service.StopRun(context.Background(), id, view.Version); !errors.Is(err, application.ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
	if _, err := f.service.Advance(context.Background(), id, domain.RunCommandAcquireLease); err == nil {
		t.Fatal("acquire_lease from running was accepted")
	}
	if got := len(f.metrics.queueTimes()); got != before {
		t.Fatalf("%d queue times after refusals, want %d", got, before)
	}
	if got := f.metrics.activeReadings(); got != readings {
		t.Fatalf("%d gauge readings after refusals, want %d", got, readings)
	}
}

func bytesContain(haystack []byte, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if string(haystack[i:i+len(needle)]) == needle {
			return true
		}
	}
	return false
}
