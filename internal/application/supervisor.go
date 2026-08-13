package application

import (
	"context"

	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/pkg/protocol"
)

// RunControl is the durable supervisor checkpoint. Epoch is a monotonically
// increasing fencing number; LeaseToken is a secret capability and must never
// be logged or persisted in events.
type RunControl struct {
	RunID          domain.RunID
	DesiredState   domain.RunState
	HarnessDriver  string
	SandboxDriver  string
	ExecutionID    string
	LeaseOwner     string
	LeaseToken     string
	LeaseEpoch     uint64
	LeaseExpiresAt domain.Timestamp
	LastSequence   uint64
	LastOrdinal    uint64
	Version        uint64
}

type RunLease struct {
	RunID        domain.RunID
	Owner, Token string
	Epoch        uint64
	ExpiresAt    domain.Timestamp
}

// SupervisorStore makes lease checks and durable changes atomic. In particular,
// implementations must fence every mutation by both epoch and token.
type SupervisorStore interface {
	LoadControl(context.Context, domain.RunID) (RunControl, error)
	AcquireRunLease(context.Context, domain.RunID, string, string, domain.Timestamp, domain.Timestamp) (RunLease, error)
	RenewRunLease(context.Context, RunLease, domain.Timestamp) (RunLease, error)
	ReleaseRunLease(context.Context, RunLease) error
	SaveDesiredState(context.Context, RunLease, domain.RunState, string, string, string) (RunControl, error)
	AppendObservation(context.Context, RunLease, uint64, []UnsequencedEvent) ([]protocol.Event, error)
}

// EventRedactor is invoked before any observation crosses a persistence or
// publication boundary. It must return only safe-to-persist event data.
type EventRedactor interface {
	Redact(context.Context, UnsequencedEvent) (UnsequencedEvent, error)
}
