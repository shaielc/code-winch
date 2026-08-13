package application

import (
	"context"
	"encoding/json"

	"github.com/shaielc/code-winch/internal/domain"
)

// WorkflowDefinitionRecord is an immutable, versioned workflow definition.
// Document contains validated canonical JSON; repository errors must never echo it.
type WorkflowDefinitionRecord struct {
	DefinitionID string
	Version      string
	Document     json.RawMessage
	CreatedAt    domain.Timestamp
}

type WorkflowInstanceRecord struct {
	ID                domain.WorkflowID
	DefinitionID      string
	DefinitionVersion string
	Status            string
	Inputs            json.RawMessage
	HarnessProfile    string
	SandboxProfile    string
	ParentInstanceID  domain.WorkflowID
	ParentStepID      string
	CreatedAt         domain.Timestamp
}

// WorkflowStepAttempt is append-only after it reaches a terminal state. Epoch
// and LeaseToken fence mutations made by a worker after its lease is reclaimed.
type WorkflowStepAttempt struct {
	ID             domain.AttemptID
	InstanceID     domain.WorkflowID
	StepID         string
	Attempt        uint
	State          string
	AvailableAt    domain.Timestamp
	StartedAt      *domain.Timestamp
	FinishedAt     *domain.Timestamp
	Output         json.RawMessage
	ErrorCode      string
	LeaseOwner     string
	LeaseToken     string
	LeaseEpoch     uint64
	LeaseExpiresAt *domain.Timestamp
}

type WorkflowSignal struct {
	ID, IdempotencyKey string
	InstanceID         domain.WorkflowID
	Kind               string
	Payload            json.RawMessage
	ReceivedAt         domain.Timestamp
}

type WorkflowTimer struct {
	ID         string
	InstanceID domain.WorkflowID
	StepID     string
	FireAt     domain.Timestamp
}

type WorkflowOutboxMessage struct {
	ID, Topic string
	Payload   json.RawMessage
}

// WorkflowRepository is the durable boundary used by the coordinator. Creating
// an instance and its initial attempts is atomic, as is completing an attempt
// and recording its publish intents.
type WorkflowRepository interface {
	PutWorkflowDefinition(context.Context, WorkflowDefinitionRecord) error
	CreateWorkflowInstance(context.Context, WorkflowInstanceRecord, []WorkflowStepAttempt) error
	GetWorkflowInstance(context.Context, domain.WorkflowID) (WorkflowInstanceRecord, []WorkflowStepAttempt, error)
	ClaimReadyWorkflowSteps(context.Context, string, string, domain.Timestamp, domain.Timestamp, int) ([]WorkflowStepAttempt, error)
	CompleteWorkflowStep(context.Context, WorkflowStepAttempt, string, json.RawMessage, string, domain.Timestamp, []WorkflowOutboxMessage) error
	AppendWorkflowAttempt(context.Context, WorkflowStepAttempt) error
	PutWorkflowSignal(context.Context, WorkflowSignal) (bool, error)
	PutWorkflowTimer(context.Context, WorkflowTimer) (bool, error)
}
