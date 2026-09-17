package application_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// control adapts the real supervisor to the executor port, the way the daemon's
// composition root does. Testing against the supervisor rather than a double is
// what makes the lease fencing part of what these tests cover.
type control struct{ *supervisor.Supervisor }

func (c control) Execute(ctx context.Context, lease application.RunLease, command application.ExecutionCommand) error {
	return c.Supervisor.Execute(ctx, lease, supervisor.Command{DesiredState: command.DesiredState, HarnessDriver: command.HarnessDriver, SandboxDriver: command.SandboxDriver, ExecutionID: command.ExecutionID, Message: command.Message})
}

// recordingControl keeps the sequenced events the supervisor assigned, which is
// the only place a caller sees them: the start service consumes them itself.
type recordingControl struct {
	application.RunExecutor
	mu     sync.Mutex
	events []protocol.Event
}

func (c *recordingControl) Observe(ctx context.Context, lease application.RunLease, ordinal uint64, values []application.UnsequencedEvent) ([]protocol.Event, error) {
	events, err := c.RunExecutor.Observe(ctx, lease, ordinal, values)
	c.mu.Lock()
	c.events = append(c.events, events...)
	c.mu.Unlock()
	return events, err
}

func (c *recordingControl) Events() []protocol.Event {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]protocol.Event(nil), c.events...)
}

// countingIDs issues distinct UUIDs without a test having to arrange one per
// call, which a run's event stream makes impractical to predict.
type countingIDs struct{ n int }

func (s *countingIDs) next() string {
	s.n++
	return fmt.Sprintf("00000000-0000-4000-8000-%012d", s.n)
}
func (s *countingIDs) NewWorkspaceID() domain.WorkspaceID {
	id, _ := domain.ParseWorkspaceID(s.next())
	return id
}
func (s *countingIDs) NewRunID() domain.RunID {
	id, _ := domain.ParseRunID(s.next())
	return id
}
func (s *countingIDs) NewAttemptID() domain.AttemptID {
	id, _ := domain.ParseAttemptID(s.next())
	return id
}
func (s *countingIDs) NewEventID() domain.EventID {
	id, _ := domain.ParseEventID(s.next())
	return id
}
func (s *countingIDs) NewCommandID() domain.CommandID {
	id, _ := domain.ParseCommandID(s.next())
	return id
}
func (s *countingIDs) NewArtifactID() domain.ArtifactID {
	id, _ := domain.ParseArtifactID(s.next())
	return id
}
func (s *countingIDs) NewCredentialID() domain.CredentialID {
	id, _ := domain.ParseCredentialID(s.next())
	return id
}
func (s *countingIDs) NewWorkflowID() domain.WorkflowID {
	id, _ := domain.ParseWorkflowID(s.next())
	return id
}

type startHarness struct {
	runs    *memory.RunRepository
	store   *memory.SupervisorStore
	gateway *memory.RunnerGateway
	control *recordingControl
	service *application.StartService
	runID   domain.RunID
}

func newStartHarness(t *testing.T, harness, sandbox string) *startHarness {
	t.Helper()
	now, _ := domain.NewTimestamp(time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC))
	ids := &countingIDs{}
	runID, attemptID := ids.NewRunID(), ids.NewAttemptID()
	runs := &memory.RunRepository{}
	if _, _, err := runs.Create(context.Background(), application.RunRecord{ID: runID, Attempts: []domain.Attempt{{ID: attemptID, State: domain.RunStateCreated}}, WorkspacePath: "/tmp/ws", HarnessProfile: harness, SandboxProfile: sandbox}, application.CreateRunIdentity{Actor: "actor", IdempotencyKey: "key"}); err != nil {
		t.Fatal(err)
	}
	store := &memory.SupervisorStore{}
	gateway := &memory.RunnerGateway{}
	clock := memory.NewClock(now)
	executor := &recordingControl{RunExecutor: control{supervisor.New(store, gateway, application.SensitivityRedactor{}, clock, "test-owner", time.Minute)}}
	service, err := application.NewStartService(runs, executor, clock, ids, application.SupportedProfiles{Harness: "fake", Sandbox: "local"})
	if err != nil {
		t.Fatal(err)
	}
	return &startHarness{runs: runs, store: store, gateway: gateway, control: executor, service: service, runID: runID}
}

func (h *startHarness) state(t *testing.T) domain.RunState {
	t.Helper()
	record, _, err := h.runs.Get(context.Background(), h.runID)
	if err != nil {
		t.Fatal(err)
	}
	return record.Attempts[len(record.Attempts)-1].State
}

func (h *startHarness) executionID(t *testing.T) string {
	t.Helper()
	control, err := h.store.LoadControl(context.Background(), h.runID)
	if err != nil {
		t.Fatal(err)
	}
	return control.ExecutionID
}

func TestStartDrivesRunToRunningAndCommandsTheRunner(t *testing.T) {
	h := newStartHarness(t, "fake", "local")
	view, err := h.service.Start(context.Background(), h.runID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got := view.Record.Attempts[0].State; got != domain.RunStateRunning {
		t.Fatalf("start reported state %q", got)
	}
	if got := h.state(t); got != domain.RunStateRunning {
		t.Fatalf("persisted state %q", got)
	}
	kinds := []string{}
	for _, message := range h.gateway.Calls() {
		kinds = append(kinds, message.Kind)
		if message.LeaseToken == "" || message.ExecutionID == "" {
			t.Fatalf("runner command is not fenced: %#v", message)
		}
		if err = message.Validate(); err != nil {
			t.Fatalf("runner command is not valid: %v", err)
		}
	}
	if len(kinds) != 2 || kinds[0] != "prepare" || kinds[1] != "start" {
		t.Fatalf("runner commands: %v", kinds)
	}
	current, err := h.store.LoadControl(context.Background(), h.runID)
	if err != nil {
		t.Fatal(err)
	}
	if current.DesiredState != domain.RunStateRunning || current.HarnessDriver != "fake" || current.SandboxDriver != "local" {
		t.Fatalf("durable control: %#v", current)
	}
}

func TestStartRefusesUnsupportedProfilePairBeforeTouchingTheRunner(t *testing.T) {
	h := newStartHarness(t, "codex", "docker")
	_, err := h.service.Start(context.Background(), h.runID, 1)
	if !errors.Is(err, application.ErrUnsupportedProfile) {
		t.Fatalf("refusal: %v", err)
	}
	// Content-free: the error names the run, never the rejected profiles.
	if message := err.Error(); strings.Contains(message, "codex") || strings.Contains(message, "docker") {
		t.Fatalf("refusal echoed the rejected profiles: %s", message)
	}
	if got := h.state(t); got != domain.RunStateCreated {
		t.Fatalf("refused run moved to %q", got)
	}
	if calls := h.gateway.Calls(); len(calls) != 0 {
		t.Fatalf("refused run reached the runner: %v", calls)
	}
}

func TestStartRefusesAStaleVersionAndASecondStart(t *testing.T) {
	h := newStartHarness(t, "fake", "local")
	if _, err := h.service.Start(context.Background(), h.runID, 7); !errors.Is(err, application.ErrPreconditionFailed) {
		t.Fatalf("stale version: %v", err)
	}
	if _, err := h.service.Start(context.Background(), h.runID, 1); err != nil {
		t.Fatal(err)
	}
	_, _, err := h.runs.Get(context.Background(), h.runID)
	if err != nil {
		t.Fatal(err)
	}
	// A zero expected version skips the precondition, so what refuses the replay
	// is the state machine rather than the ETag.
	if _, err = h.service.Start(context.Background(), h.runID, 0); !errors.Is(err, application.ErrStateConflict) {
		t.Fatalf("replayed start: %v", err)
	}
}

func TestStartRecordsAFailedAttemptWhenTheRunnerRefusesToPrepare(t *testing.T) {
	h := newStartHarness(t, "fake", "local")
	h.gateway.Failures.Inject("send", errors.New("runner unavailable"))
	if _, err := h.service.Start(context.Background(), h.runID, 1); err == nil {
		t.Fatal("a refused prepare must not report success")
	}
	if got := h.state(t); got != domain.RunStateFailed {
		t.Fatalf("state after a refused prepare: %q", got)
	}
	kinds := []string{}
	for _, message := range h.gateway.Calls() {
		kinds = append(kinds, message.Kind)
	}
	// The refused prepare is not recorded by the gateway, so the one command it
	// did see is the cleanup that releases whatever the runner had reached.
	if len(kinds) != 1 || kinds[0] != "cleanup" {
		t.Fatalf("a failed launch did not release the sandbox: %v", kinds)
	}
	current, err := h.store.LoadControl(context.Background(), h.runID)
	if err != nil {
		t.Fatal(err)
	}
	if current.LeaseToken != "" {
		t.Fatal("a failed launch kept the run lease")
	}
}

func TestObservationsAppendOrderedEventsAndReachATerminalState(t *testing.T) {
	h := newStartHarness(t, "fake", "local")
	if _, err := h.service.Start(context.Background(), h.runID, 1); err != nil {
		t.Fatal(err)
	}
	executionID := h.executionID(t)
	output := application.UnsequencedEvent{Kind: "stream.raw", SchemaVersion: 1, Source: protocol.Source{Type: "harness", Adapter: "fake"}, Sensitivity: protocol.SensitivityUserContent, Payload: []byte(`{"stream":"stdout","encoding":"utf-8","data":"hello\r\n"}`)}
	observations := []application.RunnerObservation{
		{ExecutionID: executionID, Ordinal: 1, Type: application.ObservationStarted},
		{ExecutionID: executionID, Ordinal: 2, Type: application.ObservationOutput, Event: &output},
		{ExecutionID: executionID, Ordinal: 3, Type: application.ObservationExited, Exit: &application.HarnessExit{Successful: true, Code: "OK"}},
	}
	for _, observation := range observations {
		if err := h.service.Observe(context.Background(), observation); err != nil {
			t.Fatalf("observe %s: %v", observation.Type, err)
		}
	}
	if got := h.state(t); got != domain.RunStateCompleted {
		t.Fatalf("state after a successful exit: %q", got)
	}
	events := h.control.Events()
	if len(events) != 3 {
		t.Fatalf("appended %d events", len(events))
	}
	for i, event := range events {
		if event.Sequence != uint64(i+1) {
			t.Fatalf("event %d has sequence %d", i, event.Sequence)
		}
		if event.EventID == "" || event.OccurredAt.IsZero() {
			t.Fatalf("event %d carries no identity: %#v", i, event)
		}
	}
	if events[0].Kind != "run.lifecycle" || events[1].Kind != "stream.raw" || events[2].Kind != "run.lifecycle" {
		t.Fatalf("event kinds: %v", []string{events[0].Kind, events[1].Kind, events[2].Kind})
	}
	var terminal map[string]string
	if err := json.Unmarshal(events[2].Payload, &terminal); err != nil {
		t.Fatal(err)
	}
	if terminal["state"] != "completed" || terminal["reasonCode"] != "OK" {
		t.Fatalf("terminal lifecycle payload: %v", terminal)
	}
	// The run is released, so a later observation for the same execution is
	// ignored rather than appended behind a terminal state.
	if err := h.service.Observe(context.Background(), application.RunnerObservation{ExecutionID: executionID, Ordinal: 4, Type: application.ObservationOutput, Event: &output}); err != nil {
		t.Fatal(err)
	}
	if len(h.control.Events()) != 3 {
		t.Fatal("an observation after release was appended")
	}
}

func TestFailedExitIsRecordedAsAFailedRun(t *testing.T) {
	h := newStartHarness(t, "fake", "local")
	if _, err := h.service.Start(context.Background(), h.runID, 1); err != nil {
		t.Fatal(err)
	}
	executionID := h.executionID(t)
	if err := h.service.Observe(context.Background(), application.RunnerObservation{ExecutionID: executionID, Ordinal: 1, Type: application.ObservationExited, Exit: &application.HarnessExit{Code: "PROCESS_FAILED"}}); err != nil {
		t.Fatal(err)
	}
	if got := h.state(t); got != domain.RunStateFailed {
		t.Fatalf("state after an unsuccessful exit: %q", got)
	}
	kinds := []string{}
	for _, message := range h.gateway.Calls() {
		kinds = append(kinds, message.Kind)
	}
	if len(kinds) != 3 || kinds[2] != "cleanup" {
		t.Fatalf("a terminal run did not release its sandbox: %v", kinds)
	}
}

func TestObservationsForAnUnknownExecutionAreIgnored(t *testing.T) {
	h := newStartHarness(t, "fake", "local")
	if err := h.service.Observe(context.Background(), application.RunnerObservation{ExecutionID: "not-this-process", Ordinal: 1, Type: application.ObservationStarted}); err != nil {
		t.Fatalf("unknown execution: %v", err)
	}
	if got := h.state(t); got != domain.RunStateCreated {
		t.Fatalf("an unowned observation changed the run: %q", got)
	}
}

func TestShutdownReleasesLiveExecutionsWithoutInventingATerminalState(t *testing.T) {
	h := newStartHarness(t, "fake", "local")
	if _, err := h.service.Start(context.Background(), h.runID, 1); err != nil {
		t.Fatal(err)
	}
	h.service.Shutdown(context.Background())
	if got := h.state(t); got != domain.RunStateRunning {
		t.Fatalf("shutdown invented a terminal state: %q", got)
	}
	kinds := []string{}
	for _, message := range h.gateway.Calls() {
		kinds = append(kinds, message.Kind)
	}
	if len(kinds) != 3 || kinds[2] != "cleanup" {
		t.Fatalf("shutdown left the sandbox running: %v", kinds)
	}
	current, err := h.store.LoadControl(context.Background(), h.runID)
	if err != nil {
		t.Fatal(err)
	}
	if current.LeaseToken != "" {
		t.Fatal("shutdown held the run lease")
	}
}
