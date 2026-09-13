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
		if err := c.super.Release(ctx, active.lease); err != nil {
			c.logger.Error("run lease release failed", "component", "supervisor", "operation", "shutdown", "run_id", active.runID.String(), "error_code", "lease_release_failed")
		}
		delete(c.active, executionID)
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
	c.active[executionID] = activeExecution{runID: record.ID, lease: lease}
	prepare, _ := json.Marshal(protocol.PreparePayload{WorkspaceID: record.ID.String()})
	if err = c.super.Execute(ctx, lease, supervisor.Command{DesiredState: domain.RunStatePreparing, HarnessDriver: "fake", SandboxDriver: "local", ExecutionID: executionID, Message: protocol.RunnerMessage{Version: protocol.RunnerVersion{Major: protocol.RunnerProtocolMajor}, Kind: "prepare", CommandID: uuid.NewString(), Payload: prepare}}); err != nil {
		return err
	}
	if err = c.setState(ctx, record.ID, domain.RunStatePreparing); err != nil {
		return err
	}
	start, _ := json.Marshal(protocol.StartPayload{LaunchProfile: "fake"})
	if err = c.super.Execute(ctx, lease, supervisor.Command{DesiredState: domain.RunStateRunning, HarnessDriver: "fake", SandboxDriver: "local", ExecutionID: executionID, Message: protocol.RunnerMessage{Version: protocol.RunnerVersion{Major: protocol.RunnerProtocolMajor}, Kind: "start", CommandID: uuid.NewString(), Payload: start}}); err != nil {
		return err
	}
	return c.setState(ctx, record.ID, domain.RunStateRunning)
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
				c.mu.Unlock()
				continue
			}
			_ = c.runner.Cleanup(ctx, observation.ExecutionID)
			_ = c.super.Release(ctx, active.lease)
			delete(c.active, observation.ExecutionID)
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
