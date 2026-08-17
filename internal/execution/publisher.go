package execution

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/pkg/protocol"
)

// LiveEvents is the process-local notification boundary an event reaches after
// its durable commit. Delivery is best effort by design: durable storage stays
// authoritative and subscribers repair gaps from it.
type LiveEvents interface {
	Notify(event protocol.Event)
}

// InputCommandLookup resolves the run an accepted input command belongs to. The
// outbox record carries the command ID, and the run it targets is recorded with
// the command itself.
type InputCommandLookup interface {
	RunForCommand(context.Context, domain.CommandID) (domain.RunID, error)
}

// InputDelivery hands one accepted command to the running harness.
type InputDelivery interface {
	Deliver(context.Context, domain.RunID, domain.CommandID, application.InputKind, application.InputPayload) error
}

// Publisher is the outbox's delivery side: events go to live subscribers and
// input commands go to the runner. Returning an error leaves the record claimed
// for retry, so nothing here may discard a message it did not deliver.
type Publisher struct {
	Events   LiveEvents
	Commands InputCommandLookup
	Input    InputDelivery
}

// deliveredInput mirrors the payload the input service records. It is decoded
// rather than re-derived so the delivered command is the accepted one.
type deliveredInput struct {
	CommandID string                   `json:"commandId"`
	Kind      application.InputKind    `json:"kind"`
	Payload   application.InputPayload `json:"payload"`
}

func (p Publisher) Publish(ctx context.Context, message application.OutboxMessage) error {
	switch message.Topic {
	case application.TopicRunEvents:
		return p.publishEvent(message)
	case application.TopicRunInput:
		return p.publishInput(ctx, message)
	default:
		return fmt.Errorf("outbox publisher: unroutable topic=%s", message.Topic)
	}
}

func (p Publisher) publishEvent(message application.OutboxMessage) error {
	var event protocol.Event
	if err := json.Unmarshal(message.Payload, &event); err != nil {
		return fmt.Errorf("outbox publisher: undecodable event: message=%s", message.ID)
	}
	if p.Events != nil {
		p.Events.Notify(event)
	}
	return nil
}

func (p Publisher) publishInput(ctx context.Context, message application.OutboxMessage) error {
	if p.Commands == nil || p.Input == nil {
		return fmt.Errorf("outbox publisher: input delivery is not configured: message=%s", message.ID)
	}
	var command deliveredInput
	if err := json.Unmarshal(message.Payload, &command); err != nil {
		return fmt.Errorf("outbox publisher: undecodable input command: message=%s", message.ID)
	}
	runID, err := p.Commands.RunForCommand(ctx, message.ID)
	if err != nil {
		return err
	}
	return p.Input.Deliver(ctx, runID, message.ID, command.Kind, command.Payload)
}

var _ application.OutboxPublisher = Publisher{}
