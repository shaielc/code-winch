package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/pkg/protocol"
)

var (
	// ErrStateConflict reports a command the run's current state does not allow.
	ErrStateConflict = errors.New("application: run state conflict")
	// ErrPreconditionFailed reports an expected version that is no longer current.
	ErrPreconditionFailed = errors.New("application: run version precondition failed")
	// ErrUnsupportedProfile reports a persisted harness/sandbox pair this
	// deployment did not construct. It names no profile, so an operator learns
	// that the pair is refused without the error echoing request content.
	ErrUnsupportedProfile = errors.New("application: unsupported harness and sandbox profile pair")
)

// ExecutionCommand is one fenced supervisor command. It mirrors the supervisor's
// own command shape because the application layer defines the port and the
// supervisor package, which depends on this package, implements it.
type ExecutionCommand struct {
	DesiredState                              domain.RunState
	HarnessDriver, SandboxDriver, ExecutionID string
	Message                                   protocol.RunnerMessage
}

// RunExecutor is the one-writer-per-run control surface. Every method is fenced
// by the lease, so a caller that lost ownership cannot make a durable change.
type RunExecutor interface {
	Acquire(context.Context, domain.RunID, string) (RunLease, error)
	Execute(context.Context, RunLease, ExecutionCommand) error
	Observe(context.Context, RunLease, uint64, []UnsequencedEvent) ([]protocol.Event, error)
	Release(context.Context, RunLease) error
}

// RunnerObservation is one runner-local output adapted by the composition root
// from the runner's own channel. Ordinal orders observations within an
// execution and is not a canonical event sequence.
type RunnerObservation struct {
	ExecutionID string
	Ordinal     uint64
	Type        string
	Event       *UnsequencedEvent
	Exit        *HarnessExit
}

// Observation types the runner emits.
const (
	ObservationStarted = "start"
	ObservationOutput  = "output"
	ObservationExited  = "exit"
)

// SupportedProfiles is the single harness/sandbox pair the composition root
// constructed. Naming it here rather than resolving a driver by name keeps
// registration explicit in the composition root (docs/code-structure.md §4)
// while still refusing a run that asks for anything else.
type SupportedProfiles struct{ Harness, Sandbox string }

// saveAttempts bounds the optimistic-concurrency retry. The supervisor bumps a
// run's version on every fenced write, so an attempt-state save legitimately
// races its own run's observations rather than another operator.
const saveAttempts = 8

type runExecution struct {
	runID domain.RunID
	lease RunLease
	mu    sync.Mutex
}

// StartService drives a created run to a terminal state: it validates the
// persisted profile pair, walks the run's attempt through the domain state
// machine, and turns runner observations into durably appended events.
type StartService struct {
	runs     RunRepository
	control  RunExecutor
	clock    Clock
	ids      IDSource
	profiles SupportedProfiles

	mu         sync.Mutex
	executions map[string]*runExecution
}

func NewStartService(runs RunRepository, control RunExecutor, clock Clock, ids IDSource, profiles SupportedProfiles) (*StartService, error) {
	if runs == nil || control == nil || clock == nil || ids == nil || profiles.Harness == "" || profiles.Sandbox == "" {
		return nil, errors.New("start service: repository, executor, clock, ID source, and a supported profile pair are required")
	}
	return &StartService{runs: runs, control: control, clock: clock, ids: ids, profiles: profiles, executions: map[string]*runExecution{}}, nil
}

// Start launches a created run and returns as soon as the harness is running.
// The run reaches a terminal state through Observe, not through this call.
//
// A non-zero expectedVersion is the caller's If-Match precondition. Start is not
// idempotent: replaying it against a run that already left `created` is refused
// as a state conflict rather than launching a second execution.
func (s *StartService) Start(ctx context.Context, id domain.RunID, expectedVersion uint64) (RunView, error) {
	record, version, err := s.runs.Get(ctx, id)
	if err != nil {
		return RunView{}, err
	}
	if len(record.Attempts) == 0 {
		return RunView{}, ErrInvalidRunRecord
	}
	if expectedVersion != 0 && expectedVersion != version {
		return RunView{}, fmt.Errorf("%w: run=%s", ErrPreconditionFailed, id)
	}
	// Refusing before any driver call is what keeps an unsupported request from
	// running unisolated: no sandbox is prepared and no process is launched.
	if record.HarnessProfile != s.profiles.Harness || record.SandboxProfile != s.profiles.Sandbox {
		return RunView{}, fmt.Errorf("%w: run=%s", ErrUnsupportedProfile, id)
	}
	if _, _, err = s.apply(ctx, id, domain.RunCommandStart); err != nil {
		return RunView{}, err
	}
	return s.launch(ctx, id)
}

// launch holds the run's execution lock across the whole prepare/start sequence
// so a harness that exits immediately cannot apply its terminal transition
// before the run has reached `running`.
func (s *StartService) launch(ctx context.Context, id domain.RunID) (RunView, error) {
	lease, err := s.control.Acquire(ctx, id, s.ids.NewCommandID().String())
	if err != nil {
		return RunView{}, err
	}
	executionID := s.ids.NewCommandID().String()
	execution := &runExecution{runID: id, lease: lease}
	execution.mu.Lock()
	defer execution.mu.Unlock()
	s.mu.Lock()
	s.executions[executionID] = execution
	s.mu.Unlock()

	// The lease is taken before the attempt leaves `queued`, so a run only ever
	// reaches `preparing` while exactly one writer owns it.
	if _, _, err = s.apply(ctx, id, domain.RunCommandAcquireLease); err != nil {
		s.abandon(ctx, executionID, lease, domain.RunStateQueued)
		return RunView{}, err
	}
	if err = s.command(ctx, lease, executionID, domain.RunStatePreparing, "prepare", protocol.PreparePayload{WorkspaceID: id.String()}); err != nil {
		return s.failPreparation(ctx, id, executionID, lease, err)
	}
	if err = s.command(ctx, lease, executionID, domain.RunStateRunning, "start", protocol.StartPayload{LaunchProfile: s.profiles.Harness}); err != nil {
		return s.failPreparation(ctx, id, executionID, lease, err)
	}
	record, version, err := s.apply(ctx, id, domain.RunCommandExecutionStarted)
	if err != nil {
		return s.failPreparation(ctx, id, executionID, lease, err)
	}
	return RunView{Record: record, Version: version}, nil
}

// failPreparation records the failed attempt durably before reporting, so a run
// whose harness never launched is never left claiming to be preparing.
func (s *StartService) failPreparation(ctx context.Context, id domain.RunID, executionID string, lease RunLease, cause error) (RunView, error) {
	_, _, err := s.apply(ctx, id, domain.RunCommandPreparationFailed)
	s.abandon(ctx, executionID, lease, domain.RunStateFailed)
	if err != nil {
		return RunView{}, errors.Join(cause, err)
	}
	return RunView{}, cause
}

// abandon drops a registration the pump will never receive observations for,
// releases whatever the runner did prepare, and gives up the lease. Cleanup is
// idempotent and finds nothing when preparation never got that far, so running
// it unconditionally is what keeps a half-launched harness from being orphaned.
func (s *StartService) abandon(ctx context.Context, executionID string, lease RunLease, state domain.RunState) {
	s.mu.Lock()
	delete(s.executions, executionID)
	s.mu.Unlock()
	_ = s.command(ctx, lease, executionID, state, "cleanup", protocol.CleanupPayload{})
	_ = s.control.Release(ctx, lease)
}

func (s *StartService) command(ctx context.Context, lease RunLease, executionID string, state domain.RunState, kind string, payload any) error {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("start service: encode runner command: run=%s kind=%s", lease.RunID, kind)
	}
	return s.control.Execute(ctx, lease, ExecutionCommand{
		DesiredState:  state,
		HarnessDriver: s.profiles.Harness,
		SandboxDriver: s.profiles.Sandbox,
		ExecutionID:   executionID,
		Message: protocol.RunnerMessage{
			Version:   protocol.RunnerVersion{Major: protocol.RunnerProtocolMajor, Minor: protocol.RunnerProtocolMinor},
			Kind:      kind,
			CommandID: s.ids.NewCommandID().String(),
			Payload:   encoded,
		},
	})
}

// Observe turns one runner observation into durable state. The composition root
// feeds it from the runner's observation channel; observations for an execution
// this process does not own are ignored rather than treated as an error.
func (s *StartService) Observe(ctx context.Context, observation RunnerObservation) error {
	s.mu.Lock()
	execution := s.executions[observation.ExecutionID]
	s.mu.Unlock()
	if execution == nil {
		return nil
	}
	execution.mu.Lock()
	defer execution.mu.Unlock()
	switch observation.Type {
	case ObservationStarted:
		return s.append(ctx, execution, observation.Ordinal, s.lifecycle(domain.RunStateRunning, ""))
	case ObservationOutput:
		if observation.Event == nil {
			return nil
		}
		event := *observation.Event
		// The codec produces canonical event data without identity: assigning the
		// ID and the observation time is the application's job, not the adapter's.
		event.EventID = s.ids.NewEventID()
		event.OccurredAt = s.clock.Now()
		return s.append(ctx, execution, observation.Ordinal, event)
	case ObservationExited:
		return s.finish(ctx, execution, observation)
	}
	return nil
}

func (s *StartService) append(ctx context.Context, execution *runExecution, ordinal uint64, event UnsequencedEvent) error {
	_, err := s.control.Observe(ctx, execution.lease, ordinal, []UnsequencedEvent{event})
	return err
}

// finish appends the terminal lifecycle event under the lease that fenced the
// run's output, releases the sandbox, and records the terminal attempt state.
func (s *StartService) finish(ctx context.Context, execution *runExecution, observation RunnerObservation) error {
	exit := HarnessExit{Code: "EXECUTION_LOST"}
	if observation.Exit != nil {
		exit = *observation.Exit
	}
	state, transition := domain.RunStateFailed, domain.RunCommandFailedExit
	if exit.Successful {
		state, transition = domain.RunStateCompleted, domain.RunCommandSuccessfulExit
	}
	appendErr := s.append(ctx, execution, observation.Ordinal, s.lifecycle(state, exit.Code))
	// Cleanup rides the terminal desired-state write so both are fenced by the
	// same lease; the runner releases the sandbox and reaps the process group.
	cleanupErr := s.command(ctx, execution.lease, observation.ExecutionID, state, "cleanup", protocol.CleanupPayload{})
	_, _, applyErr := s.apply(ctx, execution.runID, transition)
	s.mu.Lock()
	delete(s.executions, observation.ExecutionID)
	s.mu.Unlock()
	releaseErr := s.control.Release(ctx, execution.lease)
	return errors.Join(appendErr, cleanupErr, applyErr, releaseErr)
}

// Shutdown releases every execution this process still owns: the sandbox is
// cleaned up so no harness outlives the daemon, and the lease is released so a
// replacement process can take the run over. It deliberately writes no terminal
// state — a run interrupted mid-flight has not ended, and deciding what became
// of it belongs to restart reconciliation, not to the shutdown path.
func (s *StartService) Shutdown(ctx context.Context) {
	s.mu.Lock()
	live := make(map[string]*runExecution, len(s.executions))
	for id, execution := range s.executions {
		live[id] = execution
	}
	s.executions = map[string]*runExecution{}
	s.mu.Unlock()
	for id, execution := range live {
		execution.mu.Lock()
		_ = s.command(ctx, execution.lease, id, domain.RunStateRunning, "cleanup", protocol.CleanupPayload{})
		_ = s.control.Release(ctx, execution.lease)
		execution.mu.Unlock()
	}
}

// lifecycle builds a run.lifecycle event. Its payload carries the new state, the
// attempt-free reason code, and nothing the harness produced, so it is safe at
// operational sensitivity.
func (s *StartService) lifecycle(state domain.RunState, reasonCode string) UnsequencedEvent {
	fields := map[string]string{"state": string(state)}
	if reasonCode != "" {
		fields["reasonCode"] = reasonCode
	}
	payload, _ := json.Marshal(fields)
	return UnsequencedEvent{
		EventID:       s.ids.NewEventID(),
		OccurredAt:    s.clock.Now(),
		Kind:          "run.lifecycle",
		SchemaVersion: 1,
		Source:        protocol.Source{Type: "supervisor"},
		Sensitivity:   protocol.SensitivityOperational,
		Payload:       payload,
	}
}

// apply re-reads the run before every transition. The supervisor bumps the same
// version column on each fenced write, so the version held before a runner
// command is routinely stale by the time the attempt state changes.
func (s *StartService) apply(ctx context.Context, id domain.RunID, command domain.RunCommand) (RunRecord, uint64, error) {
	var lastErr error
	for attempt := 0; attempt < saveAttempts; attempt++ {
		record, version, err := s.runs.Get(ctx, id)
		if err != nil {
			return RunRecord{}, 0, err
		}
		run, err := domain.RestoreRun(id, record.Attempts)
		if err != nil {
			return RunRecord{}, 0, ErrInvalidRunRecord
		}
		if err = run.Apply(command, domain.AttemptID{}); err != nil {
			return RunRecord{}, 0, fmt.Errorf("%w: run=%s command=%s", ErrStateConflict, id, command)
		}
		record.Attempts = run.Attempts()
		record.UpdatedAt = s.clock.Now().Time()
		next, err := s.runs.Save(ctx, record, version)
		if err == nil {
			return record, next, nil
		}
		if !errors.Is(err, ErrConflict) {
			return RunRecord{}, 0, err
		}
		lastErr = err
	}
	return RunRecord{}, 0, lastErr
}
