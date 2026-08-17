package httpapi_test

import (
	"context"
	"errors"
	"testing"
	"time"

	fakeharness "github.com/shaielc/code-winch/internal/adapters/harness/fake"
	"github.com/shaielc/code-winch/internal/adapters/memory"
	fakesandbox "github.com/shaielc/code-winch/internal/adapters/sandbox/fake"
	"github.com/shaielc/code-winch/internal/adapters/transport/httpapi"
	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
)

// Every problem code in the OpenAPI contract must have a real producing path.
// These tests reach each httpapi.Err* sentinel by making the request that
// earns it, never by constructing the sentinel, so a mapping that stops
// matching the application's errors fails here.

type backendFixture struct {
	backend *httpapi.ApplicationBackend
	runs    *memory.RunRepository
	events  *memory.EventStore
	service *application.RunService
	ids     application.RandomIDs
}

func newBackendFixture(t *testing.T) *backendFixture {
	t.Helper()
	now, err := domain.NewTimestamp(time.Unix(1_700_000_000, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	registry := application.NewDriverRegistry()
	registry.RegisterHarness("fake", fakeharness.Driver{})
	registry.RegisterSandbox("fake", fakesandbox.New(application.SandboxCapabilities{Isolation: "in-memory"}))
	f := &backendFixture{runs: &memory.RunRepository{}, events: &memory.EventStore{}}
	f.service = application.NewRunService(f.runs, f.events, f.ids, memory.NewClock(now), registry, nil)
	inputs := application.NewInputService(&inputStore{}, capabilities{runs: f.runs}, nil)
	f.backend = httpapi.NewBackend(f.service, inputs, f.ids)
	return f
}

func (f *backendFixture) create(t *testing.T) httpapi.Run {
	t.Helper()
	run, err := f.backend.CreateRun(context.Background(), "operator", "key-create", httpapi.CreateRunRequest{
		WorkspacePath: "/workspace", HarnessProfile: "fake", SandboxProfile: "fake",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	return run
}

// capabilities answers from the stored run, the way the execution engine's
// provider does, so an input request meets the real state check.
type capabilities struct{ runs *memory.RunRepository }

func (c capabilities) InputCapabilities(ctx context.Context, id domain.RunID) (application.InputCapabilities, error) {
	record, _, err := c.runs.Get(ctx, id)
	if err != nil {
		return application.InputCapabilities{}, err
	}
	state := domain.RunStateCreated
	if len(record.Attempts) > 0 {
		state = record.Attempts[len(record.Attempts)-1].State
	}
	return application.InputCapabilities{State: state, Modes: map[application.InputKind]bool{
		application.InputText: true, application.InputResize: true,
	}}, nil
}

// inputStore records acceptances and replays a repeated key, which is what the
// durable store promises and what the idempotency check depends on.
type inputStore struct {
	byKey map[string]application.InputResult
}

func (s *inputStore) AcceptInput(_ context.Context, a application.InputAcceptance) (application.InputResult, error) {
	if s.byKey == nil {
		s.byKey = map[string]application.InputResult{}
	}
	key := a.Request.RunID.String() + "\x00" + a.Request.IdempotencyKey
	if existing, ok := s.byKey[key]; ok {
		return existing, nil
	}
	if a.Capabilities.State != domain.RunStateRunning {
		return application.InputResult{}, &application.InputError{
			Code: application.InputErrorStaleState, RunID: a.Request.RunID, Kind: a.Request.Kind,
		}
	}
	result := application.InputResult{CommandID: a.Request.CommandID, RunID: a.Request.RunID, Kind: a.Request.Kind, Accepted: true}
	s.byKey[key] = result
	return result, nil
}

// 404: a run ID that names nothing, and one that is not an ID at all.
func TestBackendProducesRunNotFound(t *testing.T) {
	f := newBackendFixture(t)
	missing := f.ids.NewRunID().String()
	for name, call := range map[string]func() error{
		"get":   func() error { _, e := f.backend.GetRun(context.Background(), "operator", missing); return e },
		"start": func() error { _, e := f.backend.StartRun(context.Background(), "operator", missing, "k", 1); return e },
		"stop": func() error {
			_, e := f.backend.StopRun(context.Background(), "operator", missing, "k", 1, httpapi.StopRunRequest{})
			return e
		},
		"events": func() error {
			_, e := f.backend.ListRunEvents(context.Background(), "operator", missing, 0, 10)
			return e
		},
		"unparse": func() error { _, e := f.backend.GetRun(context.Background(), "operator", "not-a-run-id"); return e },
	} {
		t.Run(name, func(t *testing.T) {
			if err := call(); !errors.Is(err, httpapi.ErrRunNotFound) {
				t.Fatalf("err = %v, want ErrRunNotFound", err)
			}
		})
	}
}

// 422: an unknown profile is refused before anything is launched.
func TestBackendProducesValidationFailed(t *testing.T) {
	f := newBackendFixture(t)
	for name, request := range map[string]httpapi.CreateRunRequest{
		"unknown harness": {WorkspacePath: "/workspace", HarnessProfile: "missing", SandboxProfile: "fake"},
		"unknown sandbox": {WorkspacePath: "/workspace", HarnessProfile: "fake", SandboxProfile: "missing"},
		"no workspace":    {WorkspacePath: "", HarnessProfile: "fake", SandboxProfile: "fake"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := f.backend.CreateRun(context.Background(), "operator", "key", request)
			if !errors.Is(err, httpapi.ErrValidation) {
				t.Fatalf("err = %v, want ErrValidation", err)
			}
		})
	}
}

// 412: a stale ETag on a conditional write, and on input.
func TestBackendProducesPreconditionFailed(t *testing.T) {
	f := newBackendFixture(t)
	run := f.create(t)
	stale := run.Version + 7

	if _, err := f.backend.StartRun(context.Background(), "operator", run.Id, "k1", stale); !errors.Is(err, httpapi.ErrPreconditionFailed) {
		t.Fatalf("start: err = %v, want ErrPreconditionFailed", err)
	}
	if _, err := f.backend.StopRun(context.Background(), "operator", run.Id, "k2", stale, httpapi.StopRunRequest{}); !errors.Is(err, httpapi.ErrPreconditionFailed) {
		t.Fatalf("stop: err = %v, want ErrPreconditionFailed", err)
	}
	text := "hello"
	if _, err := f.backend.SendRunInput(context.Background(), "operator", run.Id, "k3", stale, httpapi.RunInputRequest{Kind: "text", Text: &text}); !errors.Is(err, httpapi.ErrPreconditionFailed) {
		t.Fatalf("input: err = %v, want ErrPreconditionFailed", err)
	}
}

// 409 run_state_conflict: a command the run's state does not allow, and input
// to a run that is not running.
func TestBackendProducesStateConflict(t *testing.T) {
	f := newBackendFixture(t)
	run := f.create(t)

	// Stop is illegal from Created: the domain refuses it, not the transport.
	if _, err := f.backend.StopRun(context.Background(), "operator", run.Id, "k1", run.Version, httpapi.StopRunRequest{}); !errors.Is(err, httpapi.ErrStateConflict) {
		t.Fatalf("stop from created: err = %v, want ErrStateConflict", err)
	}

	// Input to a run that has not reached Running is a stale-state rejection.
	text := "hello"
	if _, err := f.backend.SendRunInput(context.Background(), "operator", run.Id, "k2", run.Version, httpapi.RunInputRequest{Kind: "text", Text: &text}); !errors.Is(err, httpapi.ErrStateConflict) {
		t.Fatalf("input to a created run: err = %v, want ErrStateConflict", err)
	}
}

// 409 idempotency_conflict: the same key reused for a different request. The
// store replays the first command, and answering with it would tell the caller
// their second, different request had been accepted.
func TestBackendProducesIdempotencyConflict(t *testing.T) {
	f := newBackendFixture(t)
	run := f.create(t)
	id, err := domain.ParseRunID(run.Id)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []domain.RunCommand{domain.RunCommandStart, domain.RunCommandAcquireLease, domain.RunCommandExecutionStarted} {
		if _, err = f.service.Advance(context.Background(), id, c); err != nil {
			t.Fatalf("advance %s: %v", c, err)
		}
	}
	view, err := f.service.GetRun(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	version := int64(view.Version)

	text := "hello"
	if _, err = f.backend.SendRunInput(context.Background(), "operator", run.Id, "shared-key", version, httpapi.RunInputRequest{Kind: "text", Text: &text}); err != nil {
		t.Fatalf("first input: %v", err)
	}
	// Same key, different request.
	rows, cols := 40, 120
	_, err = f.backend.SendRunInput(context.Background(), "operator", run.Id, "shared-key", version, httpapi.RunInputRequest{Kind: "resize", Rows: &rows, Columns: &cols})
	if !errors.Is(err, httpapi.ErrIdempotencyConflict) {
		t.Fatalf("err = %v, want ErrIdempotencyConflict", err)
	}
}

// A repeated key with the same request is a replay, not a conflict: it returns
// the command the first request recorded.
func TestBackendReplaysARepeatedIdenticalRequest(t *testing.T) {
	f := newBackendFixture(t)
	run := f.create(t)
	id, err := domain.ParseRunID(run.Id)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []domain.RunCommand{domain.RunCommandStart, domain.RunCommandAcquireLease, domain.RunCommandExecutionStarted} {
		if _, err = f.service.Advance(context.Background(), id, c); err != nil {
			t.Fatalf("advance %s: %v", c, err)
		}
	}
	view, err := f.service.GetRun(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	version := int64(view.Version)

	text := "hello"
	first, err := f.backend.SendRunInput(context.Background(), "operator", run.Id, "same-key", version, httpapi.RunInputRequest{Kind: "text", Text: &text})
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	again, err := f.backend.SendRunInput(context.Background(), "operator", run.Id, "same-key", version, httpapi.RunInputRequest{Kind: "text", Text: &text})
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if again.CommandId != first.CommandId {
		t.Fatalf("replay returned command %s, want the first request's %s", again.CommandId, first.CommandId)
	}
}

// The happy path, so the tests above are known to be rejecting for the reason
// under test rather than because nothing works.
func TestBackendCreateAndStartSucceed(t *testing.T) {
	f := newBackendFixture(t)
	run := f.create(t)
	if run.State != httpapi.RunState(domain.RunStateCreated) {
		t.Fatalf("state = %s, want created", run.State)
	}
	started, err := f.backend.StartRun(context.Background(), "operator", run.Id, "k", run.Version)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if started.State != httpapi.RunState(domain.RunStateQueued) {
		t.Fatalf("state = %s, want queued", started.State)
	}
}
