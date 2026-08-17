package execution_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/shaielc/code-winch/internal/adapters/memory"
	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/internal/execution"
	"github.com/shaielc/code-winch/pkg/protocol"
)

type notifier struct{ events []protocol.Event }

func (n *notifier) Notify(event protocol.Event) { n.events = append(n.events, event) }

type lookup struct {
	runID domain.RunID
	err   error
}

func (l lookup) RunForCommand(context.Context, domain.CommandID) (domain.RunID, error) {
	return l.runID, l.err
}

type delivery struct {
	runID   domain.RunID
	command domain.CommandID
	kind    application.InputKind
	payload application.InputPayload
	err     error
}

func (d *delivery) Deliver(_ context.Context, runID domain.RunID, command domain.CommandID, kind application.InputKind, payload application.InputPayload) error {
	d.runID, d.command, d.kind, d.payload = runID, command, kind, payload
	return d.err
}

func runID(t *testing.T) domain.RunID {
	t.Helper()
	return application.RandomIDs{}.NewRunID()
}

func TestEventsReachLiveSubscribers(t *testing.T) {
	run := runID(t)
	event := protocol.Event{EventID: "e-1", RunID: run.String(), Sequence: 7, Kind: "stream.raw", SchemaVersion: 1, Sensitivity: protocol.SensitivityUserContent, Payload: []byte(`{"data":"hi"}`)}
	body, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	live := &notifier{}
	publisher := execution.Publisher{Events: live}
	if err = publisher.Publish(context.Background(), application.OutboxMessage{Topic: application.TopicRunEvents, Payload: body}); err != nil {
		t.Fatal(err)
	}
	if len(live.events) != 1 || live.events[0].Sequence != 7 || live.events[0].RunID != run.String() {
		t.Fatalf("published %+v", live.events)
	}
}

func TestInputCommandsReachTheRunnerAsAccepted(t *testing.T) {
	run := runID(t)
	command := application.RandomIDs{}.NewCommandID()
	body, err := json.Marshal(map[string]any{"commandId": command.String(), "kind": "text", "payload": map[string]any{"text": "echo hi"}})
	if err != nil {
		t.Fatal(err)
	}
	target := &delivery{}
	publisher := execution.Publisher{Commands: lookup{runID: run}, Input: target}
	if err = publisher.Publish(context.Background(), application.OutboxMessage{ID: command, Topic: application.TopicRunInput, Payload: body}); err != nil {
		t.Fatal(err)
	}
	if target.runID != run || target.command != command || target.kind != application.InputText || target.payload.Text != "echo hi" {
		t.Fatalf("delivered run=%s command=%s kind=%s payload=%+v", target.runID, target.command, target.kind, target.payload)
	}
}

// The outbox retries whatever the publisher reports as failed, so a message it
// could not deliver must never be reported as delivered.
func TestUndeliverableMessagesAreReportedAsFailures(t *testing.T) {
	command := application.RandomIDs{}.NewCommandID()
	body := []byte(`{"commandId":"x","kind":"text","payload":{"text":"hi"}}`)
	cases := map[string]execution.Publisher{
		"no lookup":     {Input: &delivery{}},
		"no delivery":   {Commands: lookup{runID: runID(t)}},
		"lookup failed": {Commands: lookup{err: application.ErrNotFound}, Input: &delivery{}},
		"runner failed": {Commands: lookup{runID: runID(t)}, Input: &delivery{err: errors.New("no live execution")}},
	}
	for name, publisher := range cases {
		if err := publisher.Publish(context.Background(), application.OutboxMessage{ID: command, Topic: application.TopicRunInput, Payload: body}); err == nil {
			t.Errorf("%s: reported as delivered", name)
		}
	}
	publisher := execution.Publisher{Events: &notifier{}, Commands: lookup{}, Input: &delivery{}}
	if err := publisher.Publish(context.Background(), application.OutboxMessage{Topic: "run.something-else", Payload: body}); err == nil {
		t.Error("an unroutable topic was reported as delivered")
	}
	if err := publisher.Publish(context.Background(), application.OutboxMessage{Topic: application.TopicRunEvents, Payload: []byte("not json")}); err == nil {
		t.Error("an undecodable event was reported as delivered")
	}
}

// The topics are written by a store in the same transaction as the change they
// describe, so the publisher must route the literals the store writes.
func TestTopicsMatchTheStoreLiterals(t *testing.T) {
	if application.TopicRunEvents != "run.events" || application.TopicRunInput != "run.input" {
		t.Fatalf("topics are %q and %q", application.TopicRunEvents, application.TopicRunInput)
	}
}

func TestCapabilitiesOnlyOfferInputARunningHarnessCanTake(t *testing.T) {
	ids := application.RandomIDs{}
	now, err := domain.NewTimestamp(memoryNow())
	if err != nil {
		t.Fatal(err)
	}
	runs := &memory.RunRepository{}
	service := application.NewRunService(runs, &memory.EventStore{}, ids, memory.NewClock(now), drivers(t), nil)
	view, err := service.CreateRun(context.Background(), application.CreateRunCommand{WorkspacePath: "/workspace", HarnessProfile: "fake", SandboxProfile: "fake"})
	if err != nil {
		t.Fatal(err)
	}
	capabilities := execution.Capabilities{Runs: runs}

	got, err := capabilities.InputCapabilities(context.Background(), view.Record.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.State != domain.RunStateCreated || len(got.Modes) != 0 {
		t.Errorf("a created run offers %+v", got)
	}

	for _, command := range []domain.RunCommand{domain.RunCommandStart, domain.RunCommandAcquireLease, domain.RunCommandExecutionStarted} {
		if _, err = service.Advance(context.Background(), view.Record.ID, command); err != nil {
			t.Fatal(err)
		}
	}
	if got, err = capabilities.InputCapabilities(context.Background(), view.Record.ID); err != nil {
		t.Fatal(err)
	}
	if got.State != domain.RunStateRunning || !got.Modes[application.InputText] || !got.Modes[application.InputResize] {
		t.Errorf("a running run offers %+v", got)
	}
	if got.Modes[application.InputInterrupt] || got.Modes[application.InputTerminalBytes] {
		t.Error("a mode with no delivery path must not be offered")
	}

	if _, err = capabilities.InputCapabilities(context.Background(), ids.NewRunID()); !errors.Is(err, application.ErrNotFound) {
		t.Errorf("an unknown run reported %v", err)
	}
}

func memoryNow() time.Time { return time.Unix(1_700_000_000, 0).UTC() }
