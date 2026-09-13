package application

import (
	"context"
	"errors"
	"time"

	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/pkg/protocol"
)

type RunView struct {
	Record  RunRecord
	Version uint64
}

type CreateRunCommand struct {
	WorkspacePath, HarnessProfile, SandboxProfile string
	Actor, IdempotencyKey                         string
}

type RunService struct {
	repository interface {
		RunRepository
		CreateRunRepository
	}
	clock Clock
	ids   IDSource
}

var ErrUnsupportedRunProfiles = errors.New("application: unsupported run profiles")
var ErrRunStateConflict = errors.New("application: run state conflict")

// RunRuntime starts the single runtime profile selected by the composition
// root. Profile selection remains outside the provider-neutral use case.
type RunRuntime interface {
	Validate(RunRecord) error
	Start(context.Context, RunRecord) error
}

func NewRunService(repository interface {
	RunRepository
	CreateRunRepository
}, clock Clock, ids IDSource) (*RunService, error) {
	if repository == nil || clock == nil || ids == nil {
		return nil, errors.New("run service: repository, clock, and ID source are required")
	}
	return &RunService{repository: repository, clock: clock, ids: ids}, nil
}

func (s *RunService) Create(ctx context.Context, command CreateRunCommand) (RunView, error) {
	if command.WorkspacePath == "" || command.HarnessProfile == "" || command.SandboxProfile == "" || command.Actor == "" || command.IdempotencyKey == "" {
		return RunView{}, errors.New("run service: required run fields are missing")
	}
	run, err := domain.NewRun(s.ids.NewRunID(), s.ids.NewAttemptID())
	if err != nil {
		return RunView{}, err
	}
	now := s.clock.Now().Time()
	record := RunRecord{ID: run.ID(), Attempts: run.Attempts(), CreatedAt: now, UpdatedAt: now, WorkspacePath: command.WorkspacePath, HarnessProfile: command.HarnessProfile, SandboxProfile: command.SandboxProfile}
	stored, version, err := s.repository.Create(ctx, record, CreateRunIdentity{Actor: command.Actor, IdempotencyKey: command.IdempotencyKey})
	return RunView{Record: stored, Version: version}, err
}

func (s *RunService) Get(ctx context.Context, id domain.RunID) (RunView, error) {
	record, version, err := s.repository.Get(ctx, id)
	if err == nil && len(record.Attempts) == 0 {
		return RunView{}, ErrInvalidRunRecord
	}
	return RunView{Record: record, Version: version}, err
}

// Start accepts a created run and hands its execution to the configured
// runtime. The queued state is durable before any process can be launched.
func (s *RunService) Start(ctx context.Context, id domain.RunID, expected uint64, runtime RunRuntime) (RunView, error) {
	record, version, err := s.repository.Get(ctx, id)
	if err != nil {
		return RunView{}, err
	}
	if version != expected {
		return RunView{}, ErrConflict
	}
	if err = runtime.Validate(record); err != nil {
		return RunView{}, err
	}
	if err = ApplyRunTransition(&record, domain.RunCommandStart); err != nil {
		return RunView{}, ErrRunStateConflict
	}
	record.UpdatedAt = s.clock.Now().Time()
	_, err = s.repository.Save(ctx, record, version)
	if err != nil {
		return RunView{}, err
	}
	if err = runtime.Start(context.WithoutCancel(ctx), record); err != nil {
		return RunView{}, err
	}
	return s.Get(ctx, id)
}

// ApplyRunTransition routes persisted phase-zero lifecycle changes through the
// domain state machine. Phase-zero records have exactly one attempt.
func ApplyRunTransition(record *RunRecord, command domain.RunCommand) error {
	if len(record.Attempts) != 1 {
		return ErrInvalidRunRecord
	}
	run, err := domain.NewRun(record.ID, record.Attempts[0].ID)
	if err != nil {
		return err
	}
	var history []domain.RunCommand
	switch record.Attempts[0].State {
	case domain.RunStateCreated:
	case domain.RunStateQueued:
		history = []domain.RunCommand{domain.RunCommandStart}
	case domain.RunStatePreparing:
		history = []domain.RunCommand{domain.RunCommandStart, domain.RunCommandAcquireLease}
	case domain.RunStateRunning:
		history = []domain.RunCommand{domain.RunCommandStart, domain.RunCommandAcquireLease, domain.RunCommandExecutionStarted}
	default:
		return ErrRunStateConflict
	}
	for _, previous := range history {
		if err = run.Apply(previous, domain.AttemptID{}); err != nil {
			return err
		}
	}
	if err = run.Apply(command, domain.AttemptID{}); err != nil {
		return err
	}
	record.Attempts = run.Attempts()
	return nil
}

func (s *RunService) Events(ctx context.Context, id domain.RunID, after uint64, limit int, events EventStore) ([]protocol.Event, error) {
	if _, _, err := s.repository.Get(ctx, id); err != nil {
		return nil, err
	}
	return events.Read(ctx, id, after, limit)
}

type SystemClock struct{}

func (SystemClock) Now() domain.Timestamp {
	value, _ := domain.NewTimestamp(time.Now().UTC())
	return value
}
