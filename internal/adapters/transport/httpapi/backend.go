package httpapi

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/pkg/protocol"
)

type ApplicationBackend struct {
	runs  *application.RunService
	input *application.InputService
	ids   application.IDSource
}

func NewBackend(runs *application.RunService, input *application.InputService, ids application.IDSource) *ApplicationBackend {
	return &ApplicationBackend{runs, input, ids}
}
func mapError(err error) error {
	var re *domain.RunError
	var ie *application.InputError
	switch {
	case errors.Is(err, application.ErrNotFound):
		return ErrRunNotFound
	case errors.Is(err, application.ErrConflict):
		return ErrPreconditionFailed
	case errors.Is(err, application.ErrUnknownProfile):
		return ErrValidation
	case errors.As(err, &re):
		return ErrStateConflict
	case errors.As(err, &ie):
		if ie.Code == application.InputErrorNotFound {
			return ErrRunNotFound
		}
		if ie.Code == application.InputErrorInvalid {
			return ErrValidation
		}
		return ErrStateConflict
	default:
		return err
	}
}
func parseRun(id RunId) (domain.RunID, error) {
	v, e := domain.ParseRunID(id)
	if e != nil {
		return v, ErrRunNotFound
	}
	return v, nil
}
func apiRun(v application.RunView) Run {
	r := v.Record
	state := domain.RunStateCreated
	if len(r.Attempts) > 0 {
		state = r.Attempts[len(r.Attempts)-1].State
	}
	seq := int64(v.LastSequence)
	return Run{Id: r.ID.String(), WorkspacePath: r.WorkspacePath, HarnessProfile: r.HarnessProfile, SandboxProfile: r.SandboxProfile, State: RunState(state), Version: int64(v.Version), CreatedAt: r.CreatedAt.Time(), UpdatedAt: r.UpdatedAt.Time(), LastSequence: &seq}
}
func (b *ApplicationBackend) CreateRun(ctx context.Context, _ string, _ string, r CreateRunRequest) (Run, error) {
	v, e := b.runs.CreateRun(ctx, application.CreateRunCommand{WorkspacePath: r.WorkspacePath, HarnessProfile: r.HarnessProfile, SandboxProfile: r.SandboxProfile})
	return apiRun(v), mapError(e)
}
func (b *ApplicationBackend) GetRun(ctx context.Context, _ string, id RunId) (Run, error) {
	rid, e := parseRun(id)
	if e != nil {
		return Run{}, e
	}
	v, e := b.runs.GetRun(ctx, rid)
	return apiRun(v), mapError(e)
}
func (b *ApplicationBackend) StartRun(ctx context.Context, _ string, id RunId, _ string, version int64) (Run, error) {
	rid, e := parseRun(id)
	if e != nil {
		return Run{}, e
	}
	v, e := b.runs.StartRun(ctx, rid, uint64(version))
	return apiRun(v), mapError(e)
}
func (b *ApplicationBackend) StopRun(ctx context.Context, _ string, id RunId, _ string, version int64, _ StopRunRequest) (Run, error) {
	rid, e := parseRun(id)
	if e != nil {
		return Run{}, e
	}
	v, e := b.runs.StopRun(ctx, rid, uint64(version))
	return apiRun(v), mapError(e)
}

// APIEvent projects a canonical event onto the wire type. Live publication and
// history use the one projection, so a subscriber cannot see a shape the
// replayed history would not produce.
func APIEvent(v protocol.Event) Event {
	var payload map[string]any
	_ = json.Unmarshal(v.Payload, &payload)
	sourceBytes, _ := json.Marshal(v.Source)
	var source map[string]any
	_ = json.Unmarshal(sourceBytes, &source)
	return Event{EventId: v.EventID, RunId: v.RunID, Sequence: int64(v.Sequence), OccurredAt: v.OccurredAt, Kind: v.Kind, SchemaVersion: int(v.SchemaVersion), Sensitivity: EventSensitivity(v.Sensitivity), Payload: payload, Source: source}
}

func (b *ApplicationBackend) ListRunEvents(ctx context.Context, _ string, id RunId, after int64, limit int) (EventPage, error) {
	rid, e := parseRun(id)
	if e != nil {
		return EventPage{}, e
	}
	values, e := b.runs.ListRunEvents(ctx, rid, uint64(after), limit+1)
	if e != nil {
		return EventPage{}, mapError(e)
	}
	more := len(values) > limit
	if more {
		values = values[:limit]
	}
	out := EventPage{Events: make([]Event, 0, len(values)), HasMore: more, NextAfterSequence: after}
	for _, v := range values {
		out.Events = append(out.Events, APIEvent(v))
		out.NextAfterSequence = int64(v.Sequence)
	}
	return out, nil
}
func (b *ApplicationBackend) SendRunInput(ctx context.Context, actor string, id RunId, key string, version int64, r RunInputRequest) (InputAccepted, error) {
	if b.input == nil {
		return InputAccepted{}, ErrStateConflict
	}
	rid, e := parseRun(id)
	if e != nil {
		return InputAccepted{}, e
	}
	view, e := b.runs.GetRun(ctx, rid)
	if e != nil {
		return InputAccepted{}, mapError(e)
	}
	if view.Version != uint64(version) {
		return InputAccepted{}, ErrPreconditionFailed
	}
	p := application.InputPayload{}
	if r.Text != nil {
		p.Text = *r.Text
	}
	if r.Bytes != nil {
		p.Bytes = *r.Bytes
	}
	if r.Rows != nil {
		p.Rows = uint16(*r.Rows)
	}
	if r.Columns != nil {
		p.Columns = uint16(*r.Columns)
	}
	result, e := b.input.Accept(ctx, application.InputRequest{CommandID: b.ids.NewCommandID(), RunID: rid, IdempotencyKey: key, ActorID: actor, Kind: application.InputKind(r.Kind), Payload: p, ExpectedState: domain.RunStateRunning})
	if e != nil {
		return InputAccepted{}, mapError(e)
	}
	// Accept answers a repeated key with the command it already recorded. A
	// different kind under the same key is therefore not a replay but the key
	// reused for another request, which the contract refuses rather than
	// silently answering with the first request's command.
	if result.Kind != application.InputKind(r.Kind) {
		return InputAccepted{}, ErrIdempotencyConflict
	}
	return InputAccepted{Accepted: result.Accepted, CommandId: result.CommandID.String(), Kind: InputAcceptedKind(result.Kind), RunId: result.RunID.String()}, nil
}

var _ Backend = (*ApplicationBackend)(nil)
