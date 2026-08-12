package fake_test

import (
	"context"
	"testing"
	"time"

	"github.com/shaielc/code-winch/internal/adapters/harness/fake"
	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/application/testkit"
	"github.com/shaielc/code-winch/internal/domain"
	harnesscontract "github.com/shaielc/code-winch/test/contract/harness"
)

func runSpec(t *testing.T) application.RunSpec {
	t.Helper()
	id, err := domain.ParseRunID("00000000-0000-0000-0000-000000000008")
	if err != nil {
		t.Fatal(err)
	}
	return application.RunSpec{RunID: id}
}

func TestHarnessContract(t *testing.T) {
	harnesscontract.Run(t, func(*testing.T) application.HarnessDriver { return fake.Driver{} }, runSpec(t), []byte("{\"kind\":\"assistant.message.delta\",\"payload\":{\"text\":\"hello\"}}\n"))
}

func TestCompleteRunUsesFakeTime(t *testing.T) {
	now, _ := domain.NewTimestamp(time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC))
	clock := testkit.NewClock(now)
	zero := 0
	process := fake.NewProcess(clock, []fake.Step{
		{Data: []byte("{\"kind\":\"run.started\",\"payload\":{}}\n")},
		{After: 2 * time.Second, Data: []byte("{\"kind\":\"assistant.message.delta\",\"payload\":{\"text\":\"done\"}}\n"), ExitCode: &zero},
	})
	codec, _ := (fake.Driver{}).NewCodec(context.Background(), runSpec(t))
	chunks, exit := process.Poll()
	if exit != nil || len(chunks) != 1 {
		t.Fatalf("initial poll: chunks=%d exit=%v", len(chunks), exit)
	}
	events, err := codec.Consume(application.OutputChunk{Data: chunks[0]})
	if err != nil || len(events) != 1 {
		t.Fatalf("initial event: %v %#v", err, events)
	}
	clock.Advance(2 * time.Second)
	chunks, exit = process.Poll()
	if exit == nil || *exit != 0 || len(chunks) != 1 {
		t.Fatalf("final poll: chunks=%d exit=%v", len(chunks), exit)
	}
	events, err = codec.Consume(application.OutputChunk{Data: chunks[0]})
	if err != nil || len(events) != 1 || events[0].Kind != "assistant.message.delta" {
		t.Fatalf("final event: %v %#v", err, events)
	}
}
