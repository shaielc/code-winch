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
	ID                                            domain.RunID
	Attempts                                      []domain.Attempt
	WorkspacePath, HarnessProfile, SandboxProfile string
	ResolvedConfiguration                         []byte
	CreatedAt, UpdatedAt                          domain.Timestamp
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
	// LastSequence is the highest sequence committed for a run, or zero when it
	// has none yet. Implementations maintain this counter as they append, so a
	// caller learns how far a history goes without reading the history.
	// Reporting an unknown run is implementation-defined — a store that does
	// not track runs cannot tell one from a run with no events — so callers
	// establish that the run exists before asking.
	LastSequence(context.Context, domain.RunID) (uint64, error)
}

type OutboxMessage struct {
	ID      domain.CommandID
	Topic   string
	Payload []byte
}

// Outbox topics. Store implementations write these literals when they record
// delivery intent in the same transaction as the change that caused it, and a
// publisher routes on them.
const (
	// TopicRunEvents carries one canonical event, already sequenced.
	TopicRunEvents = "run.events"
	// TopicRunInput carries one accepted input command, by command ID.
	TopicRunInput = "run.input"
)

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

// InputCommandStore atomically records a command result and its delivery
// intent. Implementations return the already-recorded result on replay.
type InputCommandStore interface {
	AcceptInput(context.Context, InputAcceptance) (InputResult, error)
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
