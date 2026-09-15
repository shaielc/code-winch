package runexec

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/shaielc/code-winch/internal/adapters/memory"
	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/internal/supervisor"
	"github.com/shaielc/code-winch/pkg/protocol"
)

type closeStore struct {
	*memory.RunRepository
	*memory.SupervisorStore
}

type closeRunner struct {
	*memory.RunnerGateway
	observations chan application.RunnerObservation
	cleaned      chan string
}

func (r *closeRunner) Observations() <-chan application.RunnerObservation { return r.observations }
func (r *closeRunner) Cleanup(_ context.Context, id string) error {
	r.cleaned <- id
	return nil
}

func TestFakeProfileRedactorRejectsSecretsAndDropsExtensions(t *testing.T) {
	redactor := fakeProfileRedactor{}
	_, err := redactor.Redact(context.Background(), application.UnsequencedEvent{Sensitivity: protocol.SensitivitySecret, Payload: []byte(`{}`)})
	if err == nil {
		t.Fatal("secret event was accepted")
	}
	event, err := redactor.Redact(context.Background(), application.UnsequencedEvent{Sensitivity: protocol.SensitivityPublic, Payload: []byte(`{"safe":true}`), Extensions: map[string][]byte{"unreviewed": []byte(`true`)}})
	if err != nil {
		t.Fatal(err)
	}
	if event.Extensions != nil {
		t.Fatalf("extensions were persisted: %#v", event.Extensions)
	}
}

func TestCloseCleansActiveExecutionBeforeRunnerShutdown(t *testing.T) {
	runID, _ := domain.ParseRunID("11111111-1111-1111-1111-111111111111")
	store := &closeStore{RunRepository: &memory.RunRepository{}, SupervisorStore: &memory.SupervisorStore{}}
	runner := &closeRunner{RunnerGateway: &memory.RunnerGateway{}, observations: make(chan application.RunnerObservation), cleaned: make(chan string, 1)}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	coordinator := NewCoordinator(store, runner, &memory.IDSource{}, logger)
	lease, err := coordinator.super.Acquire(context.Background(), runID, "lease")
	if err != nil {
		t.Fatal(err)
	}
	coordinator.active["execution"] = activeExecution{runID: runID, lease: lease}
	coordinator.Close(time.Second)
	select {
	case id := <-runner.cleaned:
		if id != "execution" {
			t.Fatalf("cleaned execution %q", id)
		}
	default:
		t.Fatal("active execution was not cleaned")
	}
	if _, err = coordinator.super.Renew(context.Background(), lease); !errors.Is(err, supervisor.ErrStaleLease) {
		t.Fatalf("lease remained active after close: %v", err)
	}
}

// queuedRun seeds a store with one run that the start use case has already
// moved to queued, which is the state the coordinator is handed.
func queuedRun(t *testing.T, store *closeStore) (domain.RunID, uint64, application.RunRecord) {
	t.Helper()
	runID, _ := domain.ParseRunID("55555555-5555-5555-5555-555555555555")
	attemptID, _ := domain.ParseAttemptID("66666666-6666-6666-6666-666666666666")
	run, err := domain.NewRun(runID, attemptID)
	if err != nil {
		t.Fatal(err)
	}
	record := application.RunRecord{ID: runID, Attempts: run.Attempts(), WorkspacePath: "/tmp/ws", HarnessProfile: "fake", SandboxProfile: "local"}
	if err = application.ApplyRunTransition(&record, domain.RunCommandStart); err != nil {
		t.Fatal(err)
	}
	stored, version, err := store.Create(context.Background(), record, application.CreateRunIdentity{Actor: "actor", IdempotencyKey: "key"})
	if err != nil {
		t.Fatal(err)
	}
	return runID, version, stored
}

func TestStartAbandonsTheRunWhenTheRunnerRefusesACommand(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		failures []error
	}{
		{name: "prepare refused", failures: []error{errors.New("prepare refused")}},
		{name: "launch refused", failures: []error{nil, errors.New("launch refused")}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			store := &closeStore{RunRepository: &memory.RunRepository{}, SupervisorStore: &memory.SupervisorStore{}}
			runner := &closeRunner{RunnerGateway: &memory.RunnerGateway{}, observations: make(chan application.RunnerObservation), cleaned: make(chan string, 1)}
			runner.Failures.Inject("send", testCase.failures...)
			runID, _, record := queuedRun(t, store)
			coordinator := NewCoordinator(store, runner, &memory.IDSource{}, slog.New(slog.NewTextHandler(io.Discard, nil)))

			if err := coordinator.Start(context.Background(), record); err == nil {
				t.Fatal("start reported success after the runner refused a command")
			}

			// A queued or preparing run is startable-looking but unstartable: the
			// domain refuses a second start, so a launch that fails without a
			// terminal transition wedges the run for good.
			stored, _, err := store.Get(context.Background(), runID)
			if err != nil {
				t.Fatal(err)
			}
			if state := stored.Attempts[len(stored.Attempts)-1].State; state != domain.RunStateFailed {
				t.Fatalf("abandoned run is in state %q, not failed", state)
			}
			select {
			case id := <-runner.cleaned:
				if id == "" {
					t.Fatal("cleanup named no execution")
				}
			default:
				t.Fatal("the execution was not cleaned up")
			}
			control, err := store.LoadControl(context.Background(), runID)
			if err != nil {
				t.Fatal(err)
			}
			if control.LeaseToken != "" {
				t.Fatal("the supervisor lease is still held after the failure")
			}
			if len(coordinator.active) != 0 {
				t.Fatalf("coordinator still tracks %d execution(s)", len(coordinator.active))
			}
		})
	}
}
