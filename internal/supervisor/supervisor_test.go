package supervisor_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/shaielc/code-winch/internal/adapters/memory"
	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/internal/supervisor"
	"github.com/shaielc/code-winch/pkg/protocol"
)

type clock struct{ now domain.Timestamp }

func (c clock) Now() domain.Timestamp { return c.now }

type runner struct {
	mu     sync.Mutex
	store  *memory.SupervisorStore
	run    domain.RunID
	states []domain.RunState
}

type reconciliationRunner struct {
	runner
	observed                application.RunnerExecutionObservation
	takeovers               int
	cleanups                int
	takeoverErr, cleanupErr error
}

func (r *reconciliationRunner) Inspect(context.Context, string) (application.RunnerExecutionObservation, error) {
	return r.observed, nil
}
func (r *reconciliationRunner) Takeover(_ context.Context, execution, oldToken, newToken string) error {
	r.takeovers++
	if execution != r.observed.ExecutionID || oldToken != r.observed.OwnershipToken || newToken == "" {
		return errors.New("identity mismatch")
	}
	return r.takeoverErr
}
func (r *reconciliationRunner) Cleanup(context.Context, string) error {
	r.cleanups++
	return r.cleanupErr
}

func (r *runner) Send(ctx context.Context, m protocol.RunnerMessage) error {
	c, e := r.store.LoadControl(ctx, r.run)
	if e != nil {
		return e
	}
	r.mu.Lock()
	r.states = append(r.states, c.DesiredState)
	r.mu.Unlock()
	return nil
}

type redact struct{ canary string }

func (r redact) Redact(_ context.Context, v application.UnsequencedEvent) (application.UnsequencedEvent, error) {
	v.Payload = []byte(`{"text":"[REDACTED]"}`)
	return v, nil
}
func id(t *testing.T, n byte) domain.RunID {
	t.Helper()
	s := "00000000-0000-0000-0000-000000000001"
	b := []byte(s)
	b[len(b)-1] = '0' + n
	v, e := domain.ParseRunID(string(b))
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func eid(t *testing.T) domain.EventID {
	v, e := domain.ParseEventID("00000000-0000-0000-0000-000000000009")
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func setup(t *testing.T) (*supervisor.Supervisor, *memory.SupervisorStore, application.RunLease, *runner) {
	now, _ := domain.NewTimestamp(time.Unix(100, 0).UTC())
	st := &memory.SupervisorStore{}
	rr := &runner{store: st, run: id(t, 1)}
	s := supervisor.New(st, rr, redact{"secret"}, clock{now}, "worker", time.Minute)
	l, e := s.Acquire(context.Background(), rr.run, "token-1")
	if e != nil {
		t.Fatal(e)
	}
	return s, st, l, rr
}
func TestIntentIsDurableBeforeRunnerAndCommandsSerialize(t *testing.T) {
	s, st, l, r := setup(t)
	states := []domain.RunState{domain.RunStateQueued, domain.RunStateStopping}
	var wg sync.WaitGroup
	for _, state := range states {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := s.Execute(context.Background(), l, supervisor.Command{DesiredState: state, HarnessDriver: "fake", SandboxDriver: "local", ExecutionID: "exec", Message: protocol.RunnerMessage{Version: protocol.RunnerVersion{Major: 1}, Kind: "stop", CommandID: "cmd", Payload: []byte(`{}`)}})
			if err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	c, _ := st.LoadControl(context.Background(), l.RunID)
	if len(r.states) != 2 || r.states[1] != c.DesiredState {
		t.Fatalf("runner observed non-durable/order states: %v final=%s", r.states, c.DesiredState)
	}
}
func TestTakeoverFencesStaleObservationsAndRehydrates(t *testing.T) {
	s, st, old, _ := setup(t)
	if err := s.Release(context.Background(), old); err != nil {
		t.Fatal(err)
	}
	fresh, err := s.Acquire(context.Background(), old.RunID, "token-2")
	if err != nil {
		t.Fatal(err)
	}
	v := application.UnsequencedEvent{EventID: eid(t), OccurredAt: old.ExpiresAt, Kind: "message", SchemaVersion: 1, Source: protocol.Source{Type: "test"}, Sensitivity: protocol.SensitivityUserContent, Payload: []byte(`{"text":"secret"}`)}
	if _, err = s.Observe(context.Background(), old, 1, []application.UnsequencedEvent{v}); !errors.Is(err, supervisor.ErrStaleLease) {
		t.Fatalf("stale err=%v", err)
	}
	events, err := s.Observe(context.Background(), fresh, 1, []application.UnsequencedEvent{v})
	if err != nil {
		t.Fatal(err)
	}
	if string(events[0].Payload) != `{"text":"[REDACTED]"}` || events[0].Sequence != 1 {
		t.Fatalf("event=%+v", events[0])
	}
	c, err := supervisor.New(st, nil, redact{}, clock{}, "new", time.Minute).Rehydrate(context.Background(), old.RunID)
	if err != nil || c.LastSequence != 1 || c.LeaseEpoch != 2 {
		t.Fatalf("control=%+v err=%v", c, err)
	}
}
func TestPersistenceFailurePreventsRunnerInteraction(t *testing.T) {
	s, _, l, r := setup(t)
	bad := l
	bad.Token = "stale"
	err := s.Execute(context.Background(), bad, supervisor.Command{DesiredState: domain.RunStateRunning})
	if !errors.Is(err, supervisor.ErrStaleLease) {
		t.Fatal(err)
	}
	if len(r.states) != 0 {
		t.Fatal("runner called before persistence")
	}
}

func TestRestartMatrixReconcilesWithoutInventingSuccess(t *testing.T) {
	tests := []struct {
		name                      string
		desired                   domain.RunState
		observed                  application.ExecutionState
		success                   bool
		want                      domain.RunState
		wantTakeover, wantCleanup int
	}{
		{"prepare is lost", domain.RunStatePreparing, application.ExecutionPreparing, false, domain.RunStateFailed, 0, 1},
		{"running is resumed", domain.RunStateRunning, application.ExecutionRunning, false, domain.RunStateRunning, 1, 0},
		{"exited success is observed", domain.RunStateRunning, application.ExecutionExited, true, domain.RunStateCompleted, 0, 1},
		{"stop continues", domain.RunStateStopping, application.ExecutionRunning, false, domain.RunStateStopping, 1, 0},
		{"unknown fails", domain.RunStateRunning, application.ExecutionUnknown, false, domain.RunStateFailed, 0, 1},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runID := id(t, byte(i+1))
			oldNow, _ := domain.NewTimestamp(time.Unix(100, 0).UTC())
			newNow, _ := domain.NewTimestamp(time.Unix(200, 0).UTC())
			store := &memory.SupervisorStore{}
			seed := supervisor.New(store, &memory.RunnerGateway{}, redact{}, clock{oldNow}, "old", time.Minute)
			old, err := seed.Acquire(context.Background(), runID, "old-token")
			if err != nil {
				t.Fatal(err)
			}
			_, err = store.SaveDesiredState(context.Background(), old, tt.desired, "fake", "local", "exec-1")
			if err != nil {
				t.Fatal(err)
			}
			rr := &reconciliationRunner{observed: application.RunnerExecutionObservation{ExecutionID: "exec-1", OwnershipToken: "old-token", State: tt.observed, ExitSuccessful: tt.success}}
			rr.store, rr.run = store, runID
			event, _ := domain.ParseEventID("00000000-0000-0000-0000-000000000009")
			restarted := supervisor.New(store, rr, redact{}, clock{newNow}, "new", time.Minute).WithReconciliationIDs(&memory.IDSource{EventIDs: []domain.EventID{event}})
			got, err := restarted.Reconcile(context.Background(), runID, "new-token", rr)
			if err != nil {
				t.Fatal(err)
			}
			if got.DesiredState != tt.want || rr.takeovers != tt.wantTakeover || rr.cleanups != tt.wantCleanup {
				t.Fatalf("control=%+v takeovers=%d cleanups=%d", got, rr.takeovers, rr.cleanups)
			}
			if tt.desired == domain.RunStateStopping && len(rr.states) != 1 {
				t.Fatal("pending stop was not continued")
			}
		})
	}
}

func TestReconcileRejectsLeaseAndExecutionIdentityTakeover(t *testing.T) {
	runID := id(t, 1)
	oldNow, _ := domain.NewTimestamp(time.Unix(100, 0).UTC())
	newNow, _ := domain.NewTimestamp(time.Unix(200, 0).UTC())
	store := &memory.SupervisorStore{}
	seed := supervisor.New(store, &memory.RunnerGateway{}, redact{}, clock{oldNow}, "old", time.Minute)
	old, _ := seed.Acquire(context.Background(), runID, "old-token")
	_, _ = store.SaveDesiredState(context.Background(), old, domain.RunStateRunning, "fake", "local", "exec-1")
	rr := &reconciliationRunner{observed: application.RunnerExecutionObservation{ExecutionID: "exec-1", OwnershipToken: "different", State: application.ExecutionRunning}}
	rr.store, rr.run = store, runID
	event, _ := domain.ParseEventID("00000000-0000-0000-0000-000000000009")
	s := supervisor.New(store, rr, redact{}, clock{newNow}, "new", time.Minute).WithReconciliationIDs(&memory.IDSource{EventIDs: []domain.EventID{event}})
	got, err := s.Reconcile(context.Background(), runID, "new-token", rr)
	if err != nil {
		t.Fatal(err)
	}
	if got.DesiredState != domain.RunStateFailed || rr.takeovers != 0 || rr.cleanups != 1 {
		t.Fatalf("unsafe adoption: %+v", got)
	}
}

func TestReconcileProminentlyDiagnosesOrphanCleanupFailure(t *testing.T) {
	runID := id(t, 1)
	oldNow, _ := domain.NewTimestamp(time.Unix(100, 0).UTC())
	newNow, _ := domain.NewTimestamp(time.Unix(200, 0).UTC())
	store := &memory.SupervisorStore{}
	seed := supervisor.New(store, &memory.RunnerGateway{}, redact{}, clock{oldNow}, "old", time.Minute)
	old, _ := seed.Acquire(context.Background(), runID, "old-token")
	_, _ = store.SaveDesiredState(context.Background(), old, domain.RunStatePreparing, "fake", "local", "exec-1")
	rr := &reconciliationRunner{observed: application.RunnerExecutionObservation{ExecutionID: "exec-1", OwnershipToken: "old-token", State: application.ExecutionPreparing}, cleanupErr: errors.New("busy")}
	rr.store, rr.run = store, runID
	event, _ := domain.ParseEventID("00000000-0000-0000-0000-000000000009")
	s := supervisor.New(store, rr, redact{}, clock{newNow}, "new", time.Minute).WithReconciliationIDs(&memory.IDSource{EventIDs: []domain.EventID{event}})
	_, err := s.Reconcile(context.Background(), runID, "new-token", rr)
	if err == nil || !strings.Contains(err.Error(), supervisor.ErrCodeCleanupFailed) || strings.Contains(err.Error(), "old-token") {
		t.Fatalf("diagnostic=%v", err)
	}
}
