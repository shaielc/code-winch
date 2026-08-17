package local_test

import (
	"context"
	"testing"
	"time"

	fakeharness "github.com/shaielc/code-winch/internal/adapters/harness/fake"
	fakesandbox "github.com/shaielc/code-winch/internal/adapters/sandbox/fake"
	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/runner/local"
	"github.com/shaielc/code-winch/pkg/protocol"
)

func registry(t *testing.T) *application.DriverRegistry {
	t.Helper()
	r := application.NewDriverRegistry()
	r.RegisterHarness("fake", fakeharness.Driver{})
	r.RegisterSandbox("first", fakesandbox.New(application.SandboxCapabilities{Isolation: "in-memory", Attach: true}))
	r.RegisterSandbox("second", fakesandbox.New(application.SandboxCapabilities{Isolation: "in-memory", Attach: true}))
	return r
}

func prepare(executionID, lease, runID string) protocol.RunnerMessage {
	return protocol.RunnerMessage{
		Version:     protocol.RunnerVersion{Major: protocol.RunnerProtocolMajor},
		Kind:        "prepare",
		CommandID:   "command-" + executionID,
		ExecutionID: executionID,
		LeaseToken:  lease,
		Payload:     []byte(`{"workspaceId":"` + runID + `"}`),
	}
}

func run(t *testing.T) string {
	t.Helper()
	return application.RandomIDs{}.NewRunID().String()
}

func TestPoolRoutesEachExecutionToItsProfilePair(t *testing.T) {
	pool := local.NewPool(registry(t))
	defer pool.Close()
	ctx := context.Background()
	for _, sandbox := range []string{"first", "second"} {
		if err := pool.Bind("execution-"+sandbox, "fake", sandbox); err != nil {
			t.Fatal(err)
		}
		if err := pool.Send(ctx, prepare("execution-"+sandbox, "lease-1", run(t))); err != nil {
			t.Fatalf("%s: %v", sandbox, err)
		}
	}
	// Each execution is known only to the runner that prepared it, which is
	// what proves the pool routed rather than broadcast.
	for _, sandbox := range []string{"first", "second"} {
		observed, err := pool.Inspect(ctx, "execution-"+sandbox)
		if err != nil {
			t.Fatal(err)
		}
		if observed.State != application.ExecutionPreparing || observed.ExecutionID != "execution-"+sandbox {
			t.Errorf("%s observed as %+v", sandbox, observed)
		}
	}
}

func TestPoolRefusesMessagesForUnboundExecutions(t *testing.T) {
	pool := local.NewPool(registry(t))
	defer pool.Close()
	if err := pool.Send(context.Background(), prepare("never-bound", "lease-1", run(t))); err == nil {
		t.Fatal("an unbound execution accepted a message")
	}
	if err := pool.Bind("bad-profile", "fake", "no-such-sandbox"); err == nil {
		t.Fatal("an unknown profile was bound")
	}
}

// A restarted daemon holds no executions, and reporting one as healthy would
// tell reconciliation to leave a dead run alone.
func TestUnknownExecutionsObserveAsUnknown(t *testing.T) {
	pool := local.NewPool(registry(t))
	defer pool.Close()
	observed, err := pool.Inspect(context.Background(), "from-a-previous-process")
	if err != nil {
		t.Fatal(err)
	}
	if observed.State != application.ExecutionUnknown {
		t.Errorf("observed as %+v", observed)
	}
	if err = pool.Cleanup(context.Background(), "from-a-previous-process"); err != nil {
		t.Errorf("cleanup of an unknown execution reported %v", err)
	}
}

func TestCleanupReleasesTheRoute(t *testing.T) {
	pool := local.NewPool(registry(t))
	defer pool.Close()
	ctx := context.Background()
	if err := pool.Bind("execution-1", "fake", "first"); err != nil {
		t.Fatal(err)
	}
	if err := pool.Send(ctx, prepare("execution-1", "lease-1", run(t))); err != nil {
		t.Fatal(err)
	}
	if err := pool.Cleanup(ctx, "execution-1"); err != nil {
		t.Fatal(err)
	}
	if err := pool.Cleanup(ctx, "execution-1"); err != nil {
		t.Errorf("a repeated cleanup reported %v", err)
	}
	if err := pool.Send(ctx, prepare("execution-1", "lease-1", run(t))); err == nil {
		t.Fatal("a cleaned-up execution still routed")
	}
}

func TestObservationsFromEveryRunnerMergeAndTheStreamEnds(t *testing.T) {
	pool := local.NewPool(registry(t))
	ctx := context.Background()
	runID := run(t)
	for _, sandbox := range []string{"first", "second"} {
		id := "execution-" + sandbox
		if err := pool.Bind(id, "fake", sandbox); err != nil {
			t.Fatal(err)
		}
		if err := pool.Send(ctx, prepare(id, "lease-1", runID)); err != nil {
			t.Fatal(err)
		}
		start := protocol.RunnerMessage{Version: protocol.RunnerVersion{Major: protocol.RunnerProtocolMajor}, Kind: "start", CommandID: "start-" + id, ExecutionID: id, LeaseToken: "lease-1", Payload: []byte(`{"launchProfile":"fake"}`)}
		if err := pool.Send(ctx, start); err != nil {
			t.Fatal(err)
		}
	}
	seen := map[string]bool{}
	deadline := time.After(5 * time.Second)
	for len(seen) < 2 {
		select {
		case observation := <-pool.Observations():
			if observation.Type == "start" {
				seen[observation.ExecutionID] = true
			}
		case <-deadline:
			t.Fatalf("only saw %v", seen)
		}
	}
	// Every execution must be released before Close, whose pumps finish only
	// once their streams are closed.
	for _, sandbox := range []string{"first", "second"} {
		if err := pool.Cleanup(ctx, "execution-"+sandbox); err != nil {
			t.Fatal(err)
		}
	}
	done := make(chan struct{})
	go func() {
		for range pool.Observations() {
		}
		close(done)
	}()
	pool.Close()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("the merged observation stream did not end")
	}
}
