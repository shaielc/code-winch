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
