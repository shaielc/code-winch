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
	owner         string
	leaseDuration time.Duration
	locks         sync.Map
}

func New(store application.SupervisorStore, runner application.RunnerGateway, redactor application.EventRedactor, clock application.Clock, owner string, leaseDuration time.Duration) *Supervisor {
	return &Supervisor{store: store, runner: runner, redactor: redactor, clock: clock, owner: owner, leaseDuration: leaseDuration}
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
