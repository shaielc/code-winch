package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/shaielc/code-winch/internal/domain"
)

type InputKind string

const (
	InputText          InputKind = "text"
	InputInterrupt     InputKind = "interrupt"
	InputTerminalBytes InputKind = "terminal_bytes"
	InputResize        InputKind = "resize"
)

const (
	InputErrorInvalid      = "INPUT_INVALID"
	InputErrorUnauthorized = "INPUT_UNAUTHORIZED"
	InputErrorUnsupported  = "INPUT_UNSUPPORTED"
	InputErrorStaleState   = "INPUT_STALE_STATE"
	InputErrorNotFound     = "INPUT_RUN_NOT_FOUND"
)

// InputError deliberately contains identifiers and stable classifications but
// never input content, credentials, or authorization details.
type InputError struct {
	Code  string
	RunID domain.RunID
	Kind  InputKind
}

func (e *InputError) Error() string {
	return fmt.Sprintf("input command rejected: code=%s run=%s kind=%s", e.Code, e.RunID, e.Kind)
}

type InputPayload struct {
	Text    string `json:"text,omitempty"`
	Bytes   []byte `json:"bytes,omitempty"`
	Rows    uint16 `json:"rows,omitempty"`
	Columns uint16 `json:"columns,omitempty"`
}

type InputRequest struct {
	CommandID        domain.CommandID
	RunID            domain.RunID
	IdempotencyKey   string
	ActorID          string
	Kind             InputKind
	Payload          InputPayload
	ExpectedState    domain.RunState
	ExpectedSequence *uint64
}

type InputCapabilities struct {
	State domain.RunState
	Modes map[InputKind]bool
}

type InputAcceptance struct {
	Request         InputRequest
	Capabilities    InputCapabilities
	DeliveryPayload []byte
}

type InputResult struct {
	CommandID domain.CommandID
	RunID     domain.RunID
	Kind      InputKind
	Accepted  bool
}

type InputCapabilityProvider interface {
	InputCapabilities(context.Context, domain.RunID) (InputCapabilities, error)
}

type InputAuthorizer interface {
	AuthorizeRawTerminalInput(context.Context, string, domain.RunID) error
}

// InputService performs bounded validation and delegates the state check,
// acceptance record, and outbox insert to one durable transaction.
type InputService struct {
	store        InputCommandStore
	capabilities InputCapabilityProvider
	authorizer   InputAuthorizer
}

func NewInputService(store InputCommandStore, capabilities InputCapabilityProvider, authorizer InputAuthorizer) *InputService {
	return &InputService{store: store, capabilities: capabilities, authorizer: authorizer}
}

func (s *InputService) Accept(ctx context.Context, request InputRequest) (InputResult, error) {
	if request.CommandID.IsZero() || request.RunID.IsZero() || request.IdempotencyKey == "" || len(request.IdempotencyKey) > 200 || request.ActorID == "" || !validInput(request) {
		return InputResult{}, &InputError{Code: InputErrorInvalid, RunID: request.RunID, Kind: request.Kind}
	}
	if request.Kind == InputTerminalBytes {
		if s.authorizer == nil || s.authorizer.AuthorizeRawTerminalInput(ctx, request.ActorID, request.RunID) != nil {
			return InputResult{}, &InputError{Code: InputErrorUnauthorized, RunID: request.RunID, Kind: request.Kind}
		}
	}
	capabilities, err := s.capabilities.InputCapabilities(ctx, request.RunID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return InputResult{}, &InputError{Code: InputErrorNotFound, RunID: request.RunID, Kind: request.Kind}
		}
		return InputResult{}, fmt.Errorf("accept input: load capabilities run=%s: %w", request.RunID, err)
	}
	if !capabilities.Modes[request.Kind] {
		return InputResult{}, &InputError{Code: InputErrorUnsupported, RunID: request.RunID, Kind: request.Kind}
	}
	payload, _ := json.Marshal(struct {
		CommandID string       `json:"commandId"`
		Kind      InputKind    `json:"kind"`
		Payload   InputPayload `json:"payload"`
	}{request.CommandID.String(), request.Kind, request.Payload})
	result, err := s.store.AcceptInput(ctx, InputAcceptance{Request: request, Capabilities: capabilities, DeliveryPayload: payload})
	if err != nil {
		return InputResult{}, err
	}
	return result, nil
}

func validInput(r InputRequest) bool {
	switch r.Kind {
	case InputText:
		return r.Payload.Text != "" && len(r.Payload.Bytes) == 0
	case InputInterrupt:
		return r.Payload.Text == "" && len(r.Payload.Bytes) == 0
	case InputTerminalBytes:
		return len(r.Payload.Bytes) > 0 && r.Payload.Text == ""
	case InputResize:
		return r.Payload.Rows > 0 && r.Payload.Columns > 0 && r.Payload.Text == "" && len(r.Payload.Bytes) == 0
	default:
		return false
	}
}
