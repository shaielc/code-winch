package application

import (
	"context"
	"errors"

	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/pkg/protocol"
)

// ErrNotFound and ErrConflict let use cases make stable decisions without
// depending on an adapter's implementation details.
var (
	ErrNotFound = errors.New("application port: resource not found")
	ErrConflict = errors.New("application port: concurrent update conflict")
)

// Workspace and RunRecord are persistence DTOs. Slice fields must be copied by
// port implementations so callers cannot mutate stored state by aliasing it.
type Workspace struct {
	ID   domain.WorkspaceID
	Name string
}

type RunRecord struct {
	ID       domain.RunID
	Attempts []domain.Attempt
}

type WorkspaceRepository interface {
	Put(context.Context, Workspace) error
	Get(context.Context, domain.WorkspaceID) (Workspace, error)
}

// RunRepository uses an explicit version for optimistic concurrency. A zero
// expected version creates a record; successful writes return the new version.
type RunRepository interface {
	Save(context.Context, RunRecord, uint64) (uint64, error)
	Get(context.Context, domain.RunID) (RunRecord, uint64, error)
}

// UnsequencedEvent is canonical event data before the store assigns its
// gap-free, run-local sequence number.
type UnsequencedEvent struct {
	EventID       domain.EventID
	OccurredAt    domain.Timestamp
	Kind          string
	SchemaVersion uint
	Source        protocol.Source
	Sensitivity   protocol.Sensitivity
	Payload       []byte
	Extensions    map[string][]byte
}

type EventStore interface {
	Append(context.Context, domain.RunID, uint64, []UnsequencedEvent) ([]protocol.Event, error)
	Read(context.Context, domain.RunID, uint64, int) ([]protocol.Event, error)
}

type OutboxMessage struct {
	ID      domain.CommandID
	Topic   string
	Payload []byte
}

type OutboxPublisher interface {
	Publish(context.Context, OutboxMessage) error
}

// OutboxRecord is durable publish intent. LeaseToken fences completion and
// retry updates made by workers whose lease has subsequently expired.
type OutboxRecord struct {
	Message    OutboxMessage
	LeaseToken string
	Attempts   uint
}

// OutboxStore is the durable side of at-least-once delivery. Implementations
// must claim records exclusively and only accept updates with the active token.
type OutboxStore interface {
	ClaimOutbox(context.Context, string, string, domain.Timestamp, domain.Timestamp, int) ([]OutboxRecord, error)
	CompleteOutbox(context.Context, domain.CommandID, string, domain.Timestamp) error
	RetryOutbox(context.Context, domain.CommandID, string, domain.Timestamp, string) error
	PoisonOutbox(context.Context, domain.CommandID, string, domain.Timestamp, string) error
	OutboxBacklog(context.Context, domain.Timestamp) (uint64, error)
}

type SecretReference struct {
	CredentialID domain.CredentialID
	Provider     string
	Key          string
}

// ResolvedSecret is deliberately opaque and short-lived. It must not be
// persisted, included in an error, or formatted by implementations.
type ResolvedSecret struct{ Bytes []byte }

type SecretReferenceStore interface {
	Put(context.Context, SecretReference, ResolvedSecret) error
	Resolve(context.Context, SecretReference) (ResolvedSecret, error)
}

type RunnerGateway interface {
	Send(context.Context, protocol.RunnerMessage) error
}
