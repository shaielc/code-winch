package runexec

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/internal/supervisor"
	"github.com/shaielc/code-winch/pkg/protocol"
)

type Store interface {
	application.RunRepository
	application.SupervisorStore
}

type Runner interface {
	application.RunnerGateway
	Observations() <-chan application.RunnerObservation
	Cleanup(context.Context, string) error
}

type activeExecution struct {
	runID domain.RunID
	lease application.RunLease
}

type Coordinator struct {
	store  Store
	runner Runner
	super  *supervisor.Supervisor
	ids    application.IDSource
	logger *slog.Logger
	mu     sync.Mutex
	active map[string]activeExecution
}

type fakeProfileRedactor struct{}

func (fakeProfileRedactor) Redact(_ context.Context, event application.UnsequencedEvent) (application.UnsequencedEvent, error) {
	if event.Sensitivity == protocol.SensitivitySecret || !json.Valid(event.Payload) {
		return application.UnsequencedEvent{}, errors.New("event rejected by persistence policy")
	}
	// Provider extensions are not needed by the provider-neutral fake profile;
	// drop them rather than persisting an unreviewed content channel.
	event.Extensions = nil
	return event, nil
}

func NewCoordinator(store Store, runner Runner, ids application.IDSource, logger *slog.Logger) *Coordinator {
	c := &Coordinator{store: store, runner: runner, ids: ids, logger: logger, active: map[string]activeExecution{}}
	c.super = supervisor.New(store, runner, fakeProfileRedactor{}, application.SystemClock{}, "winchd", 30*time.Second)
	go c.consume()
	return c
}

// Close terminates every owned execution before the runner closes its
// observation channel. It is bounded by the daemon's shutdown budget.
func (c *Coordinator) Close(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	c.mu.Lock()
	defer c.mu.Unlock()
	for executionID, active := range c.active {
		if err := c.runner.Cleanup(ctx, executionID); err != nil {
			c.logger.Error("run cleanup failed", "component", "supervisor", "operation", "shutdown", "run_id", active.runID.String(), "error_code", "cleanup_failed")
		}
		c.releaseLease(ctx, "shutdown", executionID, active)
	}
}

func (c *Coordinator) Start(ctx context.Context, record application.RunRecord) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	executionID, leaseToken := uuid.NewString(), uuid.NewString()
	lease, err := c.super.Acquire(ctx, record.ID, leaseToken)
	if err != nil {
		return err
	}
	active := activeExecution{runID: record.ID, lease: lease}
	c.active[executionID] = active
	if err = c.launch(ctx, executionID, lease, record); err != nil {
		c.abandon(ctx, executionID, active)
		return err
	}
	return nil
}

func (c *Coordinator) launch(ctx context.Context, executionID string, lease application.RunLease, record application.RunRecord) error {
	prepare, _ := json.Marshal(protocol.PreparePayload{WorkspaceID: record.ID.String()})
	if err := c.super.Execute(ctx, lease, supervisor.Command{DesiredState: domain.RunStatePreparing, HarnessDriver: "fake", SandboxDriver: "local", ExecutionID: executionID, Message: protocol.RunnerMessage{Version: protocol.RunnerVersion{Major: protocol.RunnerProtocolMajor}, Kind: "prepare", CommandID: uuid.NewString(), Payload: prepare}}); err != nil {
		return err
	}
	if err := c.setState(ctx, record.ID, domain.RunStatePreparing); err != nil {
		return err
	}
	start, _ := json.Marshal(protocol.StartPayload{LaunchProfile: "fake"})
	if err := c.super.Execute(ctx, lease, supervisor.Command{DesiredState: domain.RunStateRunning, HarnessDriver: "fake", SandboxDriver: "local", ExecutionID: executionID, Message: protocol.RunnerMessage{Version: protocol.RunnerVersion{Major: protocol.RunnerProtocolMajor}, Kind: "start", CommandID: uuid.NewString(), Payload: start}}); err != nil {
		return err
	}
	return c.setState(ctx, record.ID, domain.RunStateRunning)
}

// abandon ends a start that never reached running. Cleanup precedes the durable
// transition so no harness process outlives a start the API has already
// refused, and the attempt is recorded failed so the run is terminal rather
// than parked in a nonterminal state behind a lease nobody will renew.
func (c *Coordinator) abandon(ctx context.Context, executionID string, active activeExecution) {
	if err := c.runner.Cleanup(ctx, executionID); err != nil {
		c.logger.Error("run cleanup failed", "component", "supervisor", "operation", "abandon", "run_id", active.runID.String(), "error_code", "cleanup_failed")
	}
	if err := c.failRun(ctx, active.runID); err != nil {
		c.logger.Error("launch failure transition failed", "component", "supervisor", "operation", "abandon", "run_id", active.runID.String(), "error_code", "terminal_transition_failed")
	}
	c.releaseLease(ctx, "abandon", executionID, active)
}

// failRun records the attempt as failed. The domain leaves queued only through
// the lease this coordinator already holds, so a start that failed before it
// prepared takes that step first instead of staying queued forever.
func (c *Coordinator) failRun(ctx context.Context, id domain.RunID) error {
	record, _, err := c.store.Get(ctx, id)
	if err != nil {
		return err
	}
	if len(record.Attempts) > 0 && record.Attempts[len(record.Attempts)-1].State == domain.RunStateQueued {
		if err = c.setState(ctx, id, domain.RunStatePreparing); err != nil {
			return err
		}
	}
	return c.setState(ctx, id, domain.RunStateFailed)
}

// releaseLease hands the durable lease back and forgets the execution. Both
// happen even when an earlier step failed: holding either would only block the
// next owner of a run that is no longer executing.
func (c *Coordinator) releaseLease(ctx context.Context, operation, executionID string, active activeExecution) {
	if err := c.super.Release(ctx, active.lease); err != nil {
		c.logger.Error("run lease release failed", "component", "supervisor", "operation", operation, "run_id", active.runID.String(), "error_code", "lease_release_failed")
	}
	delete(c.active, executionID)
}

func (*Coordinator) Validate(record application.RunRecord) error {
	if record.HarnessProfile != "fake" || record.SandboxProfile != "local" {
		return application.ErrUnsupportedRunProfiles
	}
	return nil
}

func (c *Coordinator) consume() {
	for observation := range c.runner.Observations() {
		c.mu.Lock()
		active, ok := c.active[observation.ExecutionID]
		if !ok {
			c.mu.Unlock()
			continue
		}
		ctx := context.Background()
		if observation.Event != nil {
			event := *observation.Event
			if event.EventID.IsZero() {
				event.EventID = c.ids.NewEventID()
			}
			if event.OccurredAt.Time().IsZero() {
				event.OccurredAt = application.SystemClock{}.Now()
			}
			if _, err := c.super.Observe(ctx, active.lease, observation.Ordinal, []application.UnsequencedEvent{event}); err != nil {
				c.logger.Error("run observation failed", "component", "supervisor", "operation", "observe", "run_id", active.runID.String(), "error_code", "observation_failed")
			}
		}
		if observation.Exit != nil {
			state := domain.RunStateFailed
			if observation.Exit.Successful {
				state = domain.RunStateCompleted
			}
			if err := c.setState(ctx, active.runID, state); err != nil {
				c.logger.Error("terminal transition failed", "component", "supervisor", "operation", "transition", "run_id", active.runID.String(), "error_code", "terminal_transition_failed")
			}
			// The execution is over whether or not the transition was written, so
			// the runner's record of it and the lease both go back here.
			_ = c.runner.Cleanup(ctx, observation.ExecutionID)
			c.releaseLease(ctx, "transition", observation.ExecutionID, active)
		}
		c.mu.Unlock()
	}
}

func (c *Coordinator) setState(ctx context.Context, id domain.RunID, state domain.RunState) error {
	record, version, err := c.store.Get(ctx, id)
	if err != nil {
		return err
	}
	var command domain.RunCommand
	switch state {
	case domain.RunStatePreparing:
		command = domain.RunCommandAcquireLease
	case domain.RunStateRunning:
		command = domain.RunCommandExecutionStarted
	case domain.RunStateCompleted:
		command = domain.RunCommandSuccessfulExit
	case domain.RunStateFailed:
		if record.Attempts[len(record.Attempts)-1].State == domain.RunStatePreparing {
			command = domain.RunCommandPreparationFailed
		} else {
			command = domain.RunCommandFailedExit
		}
	default:
		return application.ErrRunStateConflict
	}
	if err = application.ApplyRunTransition(&record, command); err != nil {
		return err
	}
	record.UpdatedAt = time.Now().UTC()
	_, err = c.store.Save(ctx, record, version)
	return err
}
