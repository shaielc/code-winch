package execution_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	fakeharness "github.com/shaielc/code-winch/internal/adapters/harness/fake"
	"github.com/shaielc/code-winch/internal/adapters/memory"
	fakesandbox "github.com/shaielc/code-winch/internal/adapters/sandbox/fake"
	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/internal/execution"
	"github.com/shaielc/code-winch/internal/runner/local"
	"github.com/shaielc/code-winch/internal/supervisor"
	"github.com/shaielc/code-winch/pkg/protocol"
)

// runner records what the engine sent and, with it, the desired state that was
// already durable when the message arrived. That ordering is the property the
// supervisor exists to provide, so every test asserting a launch also asserts
// it implicitly.
type runner struct {
	mu       sync.Mutex
	control  application.SupervisorStore
	runID    domain.RunID
	bound    map[string]string
	sent     []protocol.RunnerMessage
	states   []domain.RunState
	cleaned  []string
	failKind string
	bindErr  error
	// onStart lets a test emit runner output while the launch is still
	// running, which is what a real harness that greets you does.
	onStart func(executionID string)
}

func newRunner(control application.SupervisorStore, runID domain.RunID) *runner {
	return &runner{control: control, runID: runID, bound: map[string]string{}}
}

func (r *runner) Bind(executionID, harness, sandbox string) error {
	if r.bindErr != nil {
		return r.bindErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bound[executionID] = harness + "/" + sandbox
	return nil
}

func (r *runner) Send(ctx context.Context, m protocol.RunnerMessage) error {
	control, err := r.control.LoadControl(ctx, r.runID)
	if err != nil {
		return err
	}
	r.mu.Lock()
	if m.Kind == r.failKind {
		r.mu.Unlock()
		return errors.New("runner refused " + m.Kind)
	}
	r.sent = append(r.sent, m)
	r.states = append(r.states, control.DesiredState)
	onStart := r.onStart
	r.mu.Unlock()
	if m.Kind == "start" && onStart != nil {
		onStart(m.ExecutionID)
	}
	return nil
}

func (r *runner) Cleanup(_ context.Context, executionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cleaned = append(r.cleaned, executionID)
	return nil
}

func (r *runner) kinds() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, 0, len(r.sent))
	for _, m := range r.sent {
		out = append(out, m.Kind)
	}
	return out
}

func (r *runner) last() protocol.RunnerMessage {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.sent[len(r.sent)-1]
}

type harness struct {
	runs     *memory.RunRepository
	control  *memory.SupervisorStore
	service  *application.RunService
	engine   *execution.Engine
	runner   *runner
	observe  chan local.Observation
	consumed sync.WaitGroup
	runID    domain.RunID
}

func drivers(t *testing.T) *application.DriverRegistry {
	t.Helper()
	registry := application.NewDriverRegistry()
	registry.RegisterHarness("fake", fakeharness.Driver{})
	registry.RegisterSandbox("fake", fakesandbox.New(application.SandboxCapabilities{Isolation: "in-memory", Attach: true}))
	return registry
}

// setup builds the engine over in-memory ports and a recording runner, then
// creates one run in the `created` state.
func setup(t *testing.T) *harness {
	t.Helper()
	ids := application.RandomIDs{}
	now, err := domain.NewTimestamp(time.Unix(1_700_000_000, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	clock := memory.NewClock(now)
	runs := &memory.RunRepository{}
	events := &memory.EventStore{}
	control := &memory.SupervisorStore{}
	service := application.NewRunService(runs, events, ids, clock, drivers(t), nil)

	view, err := service.CreateRun(context.Background(), application.CreateRunCommand{WorkspacePath: "/workspace", HarnessProfile: "fake", SandboxProfile: "fake"})
	if err != nil {
		t.Fatal(err)
	}
	gateway := newRunner(control, view.Record.ID)
	engine, err := execution.New(execution.Config{
		Runs:       runs,
		States:     service,
		Control:    control,
		Supervisor: supervisor.New(control, gateway, execution.ClassificationRedactor{}, clock, "test-daemon", time.Minute).WithReconciliationIDs(ids),
		Runner:     gateway,
		IDs:        ids,
		Clock:      clock,
	})
	if err != nil {
		t.Fatal(err)
	}
	service.WithExecutor(engine)
	h := &harness{runs: runs, control: control, service: service, engine: engine, runner: gateway, observe: make(chan local.Observation, 16), runID: view.Record.ID}
	h.consumed.Add(1)
	go func() { defer h.consumed.Done(); engine.Consume(context.Background(), h.observe) }()
	t.Cleanup(func() {
		close(h.observe)
		h.consumed.Wait()
	})
	return h
}

func (h *harness) start(t *testing.T) error {
	t.Helper()
	_, err := h.service.StartRun(context.Background(), h.runID, 1)
	return err
}

func (h *harness) state(t *testing.T) domain.RunState {
	t.Helper()
	record, _, err := h.runs.Get(context.Background(), h.runID)
	if err != nil {
		t.Fatal(err)
	}
	if len(record.Attempts) == 0 {
		t.Fatal("run has no attempt")
	}
	return record.Attempts[len(record.Attempts)-1].State
}

func (h *harness) lifecycle(t *testing.T) []string {
	t.Helper()
	var out []string
	for _, event := range h.control.Events(h.runID) {
		if event.Kind != "run.lifecycle" {
			continue
		}
		var payload struct {
			State      string `json:"state"`
			ReasonCode string `json:"reasonCode"`
		}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			t.Fatal(err)
		}
		entry := payload.State
		if payload.ReasonCode != "" {
			entry += ":" + payload.ReasonCode
		}
		out = append(out, entry)
	}
	return out
}

func (h *harness) leaseHeld(t *testing.T) bool {
	t.Helper()
	control, err := h.control.LoadControl(context.Background(), h.runID)
	if err != nil {
		t.Fatal(err)
	}
	return control.LeaseToken != ""
}

// waitFor polls a condition the observation loop satisfies asynchronously.
func waitFor(t *testing.T, what string, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !condition() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(time.Millisecond)
	}
}

func TestLaunchReachesRunningAndRecordsIntentBeforeEveryRunnerCall(t *testing.T) {
	h := setup(t)
	if err := h.start(t); err != nil {
		t.Fatal(err)
	}
	if got := h.state(t); got != domain.RunStateRunning {
		t.Errorf("run state is %q", got)
	}
	if got := strings.Join(h.runner.kinds(), ","); got != "prepare,start" {
		t.Errorf("runner received %q", got)
	}
	// Each message must arrive after the state it implements is durable.
	if want := []domain.RunState{domain.RunStatePreparing, domain.RunStateRunning}; !equalStates(h.runner.states, want) {
		t.Errorf("desired states at send time were %v", h.runner.states)
	}
	if got := strings.Join(h.lifecycle(t), ","); got != "preparing,running" {
		t.Errorf("lifecycle events were %q", got)
	}
	if !h.leaseHeld(t) {
		t.Error("a running execution must hold its lease")
	}
}

func TestLaunchFailureFailsTheRunAndReleasesEverything(t *testing.T) {
	h := setup(t)
	h.runner.failKind = "start"
	if err := h.start(t); err == nil {
		t.Fatal("a refused start must be reported")
	}
	if got := h.state(t); got != domain.RunStateFailed {
		t.Errorf("run state is %q, want failed", got)
	}
	if got := strings.Join(h.lifecycle(t), ","); got != "preparing,failed:LAUNCH_FAILED" {
		t.Errorf("lifecycle events were %q", got)
	}
	if h.leaseHeld(t) {
		t.Error("a failed launch must release its lease")
	}
	if len(h.runner.cleaned) != 1 {
		t.Errorf("execution cleaned %d times", len(h.runner.cleaned))
	}
}

func TestLaunchRefusesASecondExecutionForOneRun(t *testing.T) {
	h := setup(t)
	if err := h.start(t); err != nil {
		t.Fatal(err)
	}
	if err := h.engine.Launch(context.Background(), h.runID); err == nil {
		t.Fatal("a run with a live execution must not be launched again")
	}
}

func TestOutputIsRecordedAndAnExitCompletesTheRun(t *testing.T) {
	h := setup(t)
	if err := h.start(t); err != nil {
		t.Fatal(err)
	}
	executionID := h.runner.last().ExecutionID
	h.observe <- local.Observation{ExecutionID: executionID, Type: "output", Event: &application.UnsequencedEvent{
		Kind: "stream.raw", SchemaVersion: 1, Source: protocol.Source{Type: "harness"},
		Sensitivity: protocol.SensitivityUserContent, Payload: []byte(`{"stream":"stdout","encoding":"utf-8","data":"hi"}`),
	}}
	waitFor(t, "the output event", func() bool { return len(h.control.Events(h.runID)) == 3 })

	events := h.control.Events(h.runID)
	if events[2].Kind != "stream.raw" || events[2].Sequence != 3 {
		t.Errorf("third event is %s at sequence %d", events[2].Kind, events[2].Sequence)
	}
	if events[2].EventID == "" || events[2].OccurredAt.IsZero() {
		t.Error("the engine must stamp an event ID and time on runner output")
	}

	h.observe <- local.Observation{ExecutionID: executionID, Type: "exit", Exit: &application.HarnessExit{Successful: true, Code: "OK"}}
	waitFor(t, "the run to complete", func() bool { return h.state(t) == domain.RunStateCompleted })
	if got := strings.Join(h.lifecycle(t), ","); got != "preparing,running,completed:OK" {
		t.Errorf("lifecycle events were %q", got)
	}
	waitFor(t, "the lease to be released", func() bool { return !h.leaseHeld(t) })
	if len(h.runner.cleaned) != 1 {
		t.Errorf("execution cleaned %d times", len(h.runner.cleaned))
	}
}

func TestUnsuccessfulExitFailsTheRun(t *testing.T) {
	h := setup(t)
	if err := h.start(t); err != nil {
		t.Fatal(err)
	}
	h.observe <- local.Observation{ExecutionID: h.runner.last().ExecutionID, Type: "exit", Exit: &application.HarnessExit{Code: "PROCESS_FAILED"}}
	waitFor(t, "the run to fail", func() bool { return h.state(t) == domain.RunStateFailed })
	if got := strings.Join(h.lifecycle(t), ","); got != "preparing,running,failed:PROCESS_FAILED" {
		t.Errorf("lifecycle events were %q", got)
	}
}

func TestObservationsForAnUntrackedExecutionAreIgnored(t *testing.T) {
	h := setup(t)
	if err := h.start(t); err != nil {
		t.Fatal(err)
	}
	before := len(h.control.Events(h.runID))
	h.observe <- local.Observation{ExecutionID: "not-a-live-execution", Type: "output", Event: &application.UnsequencedEvent{
		Kind: "stream.raw", SchemaVersion: 1, Sensitivity: protocol.SensitivityUserContent, Payload: []byte(`{}`),
	}}
	// Nothing to wait for, so drive one observation that must land afterwards.
	h.observe <- local.Observation{ExecutionID: h.runner.last().ExecutionID, Type: "output", Event: &application.UnsequencedEvent{
		Kind: "diagnostic.emitted", SchemaVersion: 1, Sensitivity: protocol.SensitivityOperational, Payload: []byte(`{}`),
	}}
	waitFor(t, "the tracked event", func() bool { return len(h.control.Events(h.runID)) == before+1 })
	for _, event := range h.control.Events(h.runID) {
		if event.Kind == "stream.raw" {
			t.Fatal("an unfenced observation was recorded")
		}
	}
}

func TestStopSendsAStopWithoutInventingATerminalState(t *testing.T) {
	h := setup(t)
	if err := h.start(t); err != nil {
		t.Fatal(err)
	}
	_, version, err := h.runs.Get(context.Background(), h.runID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = h.service.StopRun(context.Background(), h.runID, version); err != nil {
		t.Fatal(err)
	}
	if got := h.state(t); got != domain.RunStateStopping {
		t.Errorf("run state is %q, want stopping", got)
	}
	if got := strings.Join(h.runner.kinds(), ","); got != "prepare,start,stop" {
		t.Errorf("runner received %q", got)
	}
	if !h.leaseHeld(t) {
		t.Error("a stopping run still owns its execution")
	}
	// The exit, not the stop, is what ends the run.
	h.observe <- local.Observation{ExecutionID: h.runner.last().ExecutionID, Type: "exit", Exit: &application.HarnessExit{Successful: true, Code: "OK"}}
	waitFor(t, "the run to complete", func() bool { return h.state(t) == domain.RunStateCompleted })
}

func TestDeliverEncodesInputAndLeavesTheDesiredStateAlone(t *testing.T) {
	h := setup(t)
	if err := h.start(t); err != nil {
		t.Fatal(err)
	}
	commandID := application.RandomIDs{}.NewCommandID()
	if err := h.engine.Deliver(context.Background(), h.runID, commandID, application.InputText, application.InputPayload{Text: "echo hi"}); err != nil {
		t.Fatal(err)
	}
	sent := h.runner.last()
	if sent.Kind != "input" {
		t.Fatalf("input was sent as %q", sent.Kind)
	}
	var payload protocol.InputPayload
	if err := json.Unmarshal(sent.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Text != "echo hi" || payload.InputID != commandID.String() {
		t.Errorf("input payload is %+v", payload)
	}
	if err := h.engine.Deliver(context.Background(), h.runID, commandID, application.InputResize, application.InputPayload{Rows: 24, Columns: 80}); err != nil {
		t.Fatal(err)
	}
	if got := h.runner.last().Kind; got != "resize" {
		t.Errorf("resize was sent as %q", got)
	}
	// Delivery must not rewrite what the run is supposed to be doing.
	control, err := h.control.LoadControl(context.Background(), h.runID)
	if err != nil {
		t.Fatal(err)
	}
	if control.DesiredState != domain.RunStateRunning {
		t.Errorf("desired state is %q after input", control.DesiredState)
	}
}

func TestDeliverRefusesKindsTheRunnerCannotCarry(t *testing.T) {
	h := setup(t)
	if err := h.start(t); err != nil {
		t.Fatal(err)
	}
	commandID := application.RandomIDs{}.NewCommandID()
	if err := h.engine.Deliver(context.Background(), h.runID, commandID, application.InputInterrupt, application.InputPayload{}); err == nil {
		t.Error("interrupt has no runner message yet and must be refused")
	}
	if err := h.engine.Deliver(context.Background(), h.runID, commandID, application.InputTerminalBytes, application.InputPayload{Bytes: []byte{3}}); err == nil {
		t.Error("raw terminal bytes have no runner message yet and must be refused")
	}
}

func TestDeliverWithoutALiveExecutionIsAnError(t *testing.T) {
	h := setup(t)
	err := h.engine.Deliver(context.Background(), h.runID, application.RandomIDs{}.NewCommandID(), application.InputText, application.InputPayload{Text: "hi"})
	if err == nil {
		t.Fatal("input for a run that is not executing must not be reported as delivered")
	}
}

func TestShutdownEndsLiveExecutionsTruthfully(t *testing.T) {
	h := setup(t)
	if err := h.start(t); err != nil {
		t.Fatal(err)
	}
	h.engine.Shutdown(context.Background())
	if got := h.state(t); got != domain.RunStateFailed {
		t.Errorf("run state is %q, want failed", got)
	}
	if got := strings.Join(h.lifecycle(t), ","); got != "preparing,running,failed:DAEMON_SHUTDOWN" {
		t.Errorf("lifecycle events were %q", got)
	}
	if h.leaseHeld(t) {
		t.Error("shutdown must release the lease")
	}
	if len(h.runner.cleaned) != 1 {
		t.Errorf("execution cleaned %d times", len(h.runner.cleaned))
	}
}

// A harness that produces output the moment it starts writes through the
// observation loop while the launch is still emitting its own lifecycle
// events. Both paths allocate ordinals from the same execution, and the store
// refuses an ordinal that no longer leads, so an unserialized write loses
// whichever event lost the race — in practice the `running` marker.
func TestLifecycleEventsSurviveOutputArrivingDuringTheLaunch(t *testing.T) {
	h := setup(t)
	const noisy = 200
	emitting := make(chan struct{})
	var pumped sync.WaitGroup
	pumped.Add(1)
	// The output keeps arriving while the launch emits, so the two writers
	// overlap rather than merely follow one another.
	h.runner.onStart = func(executionID string) {
		go func() {
			defer pumped.Done()
			close(emitting)
			for i := 0; i < noisy; i++ {
				h.observe <- local.Observation{ExecutionID: executionID, Type: "output", Event: &application.UnsequencedEvent{
					Kind: "stream.raw", SchemaVersion: 1, Source: protocol.Source{Type: "harness"},
					Sensitivity: protocol.SensitivityUserContent, Payload: []byte(`{"stream":"stdout","encoding":"utf-8","data":"noise"}`),
				}}
			}
		}()
		<-emitting
	}
	if err := h.start(t); err != nil {
		t.Fatal(err)
	}
	pumped.Wait()
	waitFor(t, "every event", func() bool { return len(h.control.Events(h.runID)) == noisy+2 })
	if got := strings.Join(h.lifecycle(t), ","); got != "preparing,running" {
		t.Errorf("lifecycle events were %q", got)
	}
	// Sequences stay gap-free whatever order the writers arrived in.
	for i, event := range h.control.Events(h.runID) {
		if event.Sequence != uint64(i+1) {
			t.Fatalf("event %d has sequence %d", i, event.Sequence)
		}
	}
}

func TestNewRejectsAnIncompleteConfiguration(t *testing.T) {
	if _, err := execution.New(execution.Config{}); err == nil {
		t.Fatal("an engine without ports must not be constructed")
	}
}

func equalStates(got, want []domain.RunState) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
