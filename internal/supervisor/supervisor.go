// Package supervisor serializes per-run commands and fences all durable effects.
package supervisor

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/pkg/protocol"
)

var ErrStaleLease = errors.New("run supervisor: stale lease")

type Command struct {
	DesiredState                              domain.RunState
	HarnessDriver, SandboxDriver, ExecutionID string
	Message                                   protocol.RunnerMessage
}

type Supervisor struct {
	store         application.SupervisorStore
	runner        application.RunnerGateway
	redactor      application.EventRedactor
	clock         application.Clock
	ids           application.IDSource
	owner         string
	leaseDuration time.Duration
	locks         sync.Map
}

func New(store application.SupervisorStore, runner application.RunnerGateway, redactor application.EventRedactor, clock application.Clock, owner string, leaseDuration time.Duration) *Supervisor {
	return &Supervisor{store: store, runner: runner, redactor: redactor, clock: clock, owner: owner, leaseDuration: leaseDuration}
}

// WithReconciliationIDs enables durable reconciliation events. It is separate
// from New to preserve the small supervisor construction surface for callers
// that do not run restart recovery.
func (s *Supervisor) WithReconciliationIDs(ids application.IDSource) *Supervisor {
	s.ids = ids
	return s
}

const (
	ErrCodeExecutionLost    = "EXECUTION_LOST"
	ErrCodeIdentityMismatch = "EXECUTION_IDENTITY_MISMATCH"
	ErrCodeCleanupFailed    = "ORPHAN_CLEANUP_FAILED"
)

// Reconcile compares one nonterminal checkpoint with the runner and takes a
// new fenced lease. The old runner token is used only as takeover evidence and
// is never included in diagnostics.
func (s *Supervisor) Reconcile(ctx context.Context, id domain.RunID, token string, runner application.ReconciliationRunner) (application.RunControl, error) {
	before, err := s.store.LoadControl(ctx, id)
	if err != nil || before.DesiredState.IsTerminal() {
		return before, err
	}
	lease, err := s.Acquire(ctx, id, token)
	if err != nil {
		return before, err
	}
	code, state := "EXECUTION_RESUMED", before.DesiredState
	obs := application.RunnerExecutionObservation{State: application.ExecutionUnknown}
	if before.ExecutionID != "" {
		obs, err = runner.Inspect(ctx, before.ExecutionID)
	}
	if err != nil || obs.State == application.ExecutionUnknown || obs.ExecutionID != before.ExecutionID {
		// Runner errors can contain adapter details. Persist the stable,
		// content-free outcome instead of propagating them.
		err = nil
		code, state = ErrCodeExecutionLost, domain.RunStateFailed
	} else if obs.State == application.ExecutionPreparing {
		code, state = ErrCodeExecutionLost, domain.RunStateFailed
	} else if obs.State == application.ExecutionExited {
		if obs.ExitSuccessful && before.DesiredState != domain.RunStateStopping {
			state = domain.RunStateCompleted
		} else if before.DesiredState == domain.RunStateStopping {
			state = domain.RunStateCancelled
		} else {
			state = domain.RunStateFailed
		}
		code = "EXECUTION_EXIT_RECONCILED"
	} else if obs.OwnershipToken != before.LeaseToken || runner.Takeover(ctx, before.ExecutionID, before.LeaseToken, lease.Token) != nil {
		code, state = ErrCodeIdentityMismatch, domain.RunStateFailed
	} else if before.DesiredState == domain.RunStateStopping {
		code = "STOP_RECONCILIATION_CONTINUED"
		msg := protocol.RunnerMessage{Version: protocol.RunnerVersion{Major: protocol.RunnerProtocolMajor}, Kind: "stop", CommandID: "reconcile-stop-" + id.String(), ExecutionID: before.ExecutionID, LeaseToken: lease.Token, Payload: []byte(`{"graceMilliseconds":0}`)}
		if err = s.runner.Send(ctx, msg); err != nil {
			return before, fmt.Errorf("run supervisor: stop reconciliation failed: run=%s: %w", id, err)
		}
	}
	updated, saveErr := s.store.SaveDesiredState(ctx, lease, state, before.HarnessDriver, before.SandboxDriver, before.ExecutionID)
	if saveErr != nil {
		return updated, fence(saveErr, id)
	}
	if state.IsTerminal() && before.ExecutionID != "" {
		if cleanupErr := runner.Cleanup(ctx, before.ExecutionID); cleanupErr != nil {
			code = ErrCodeCleanupFailed
			err = fmt.Errorf("run supervisor: %s: run=%s execution=%s", code, id, before.ExecutionID)
		}
	}
	if s.ids == nil {
		return updated, errors.New("run supervisor: reconciliation event ID source is required")
	}
	payload := []byte(fmt.Sprintf(`{"code":%q,"executionId":%q,"state":%q}`, code, before.ExecutionID, state))
	_, eventErr := s.Observe(ctx, lease, before.LastOrdinal+1, []application.UnsequencedEvent{{EventID: s.ids.NewEventID(), OccurredAt: s.clock.Now(), Kind: "run.reconciled", SchemaVersion: 1, Source: protocol.Source{Type: "supervisor"}, Sensitivity: protocol.SensitivityPublic, Payload: payload}})
	if eventErr != nil {
		return updated, eventErr
	}
	return updated, err
}

func (s *Supervisor) lock(id domain.RunID) func() {
	value, _ := s.locks.LoadOrStore(id, &sync.Mutex{})
	mu := value.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

// Acquire takes durable ownership. A takeover always receives a greater epoch.
func (s *Supervisor) Acquire(ctx context.Context, id domain.RunID, token string) (application.RunLease, error) {
	now := s.clock.Now()
	until, err := domain.NewTimestamp(now.Time().Add(s.leaseDuration))
	if err != nil {
		return application.RunLease{}, err
	}
	lease, err := s.store.AcquireRunLease(ctx, id, s.owner, token, now, until)
	if errors.Is(err, application.ErrConflict) {
		return lease, fmt.Errorf("%w: run=%s", ErrStaleLease, id)
	}
	return lease, err
}

func (s *Supervisor) Renew(ctx context.Context, lease application.RunLease) (application.RunLease, error) {
	until, err := domain.NewTimestamp(s.clock.Now().Time().Add(s.leaseDuration))
	if err != nil {
		return lease, err
	}
	lease, err = s.store.RenewRunLease(ctx, lease, until)
	return lease, fence(err, lease.RunID)
}
func (s *Supervisor) Release(ctx context.Context, lease application.RunLease) error {
	return fence(s.store.ReleaseRunLease(ctx, lease), lease.RunID)
}

// Execute persists intent before interacting with the runner. The run-local
// mutex gives concurrent callers one order while the durable lease provides
// correctness across restart and takeover.
func (s *Supervisor) Execute(ctx context.Context, lease application.RunLease, command Command) error {
	unlock := s.lock(lease.RunID)
	defer unlock()
	if _, err := s.store.SaveDesiredState(ctx, lease, command.DesiredState, command.HarnessDriver, command.SandboxDriver, command.ExecutionID); err != nil {
		return fence(err, lease.RunID)
	}
	command.Message.LeaseToken = lease.Token
	command.Message.ExecutionID = command.ExecutionID
	if err := s.runner.Send(ctx, command.Message); err != nil {
		return fmt.Errorf("run supervisor: runner command failed: run=%s kind=%s: %w", lease.RunID, command.Message.Kind, err)
	}
	return nil
}

// Observe rejects replayed ordinals and stale ownership before assigning the
// canonical sequence. Redaction is completed before the atomic store call.
func (s *Supervisor) Observe(ctx context.Context, lease application.RunLease, ordinal uint64, values []application.UnsequencedEvent) ([]protocol.Event, error) {
	unlock := s.lock(lease.RunID)
	defer unlock()
	redacted := make([]application.UnsequencedEvent, len(values))
	for i := range values {
		v, err := s.redactor.Redact(ctx, values[i])
		if err != nil {
			return nil, fmt.Errorf("run supervisor: redaction failed: run=%s event=%s", lease.RunID, values[i].EventID)
		}
		if v.Sensitivity == protocol.SensitivitySecret {
			return nil, fmt.Errorf("run supervisor: redaction rejected secret event: run=%s event=%s", lease.RunID, values[i].EventID)
		}
		redacted[i] = v
	}
	events, err := s.store.AppendObservation(ctx, lease, ordinal, redacted)
	return events, fence(err, lease.RunID)
}

func (s *Supervisor) Rehydrate(ctx context.Context, id domain.RunID) (application.RunControl, error) {
	return s.store.LoadControl(ctx, id)
}
func fence(err error, id domain.RunID) error {
	if errors.Is(err, application.ErrConflict) {
		return fmt.Errorf("%w: run=%s", ErrStaleLease, id)
	}
	return err
}
