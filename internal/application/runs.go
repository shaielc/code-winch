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

// RunRuntime starts the single runtime profile selected by the composition
// root. Profile selection remains outside the provider-neutral use case.
type RunRuntime interface {
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
	if record.HarnessProfile != "fake" || record.SandboxProfile != "local" {
		return RunView{}, ErrUnsupportedRunProfiles
	}
	if len(record.Attempts) == 0 || record.Attempts[len(record.Attempts)-1].State != domain.RunStateCreated {
		return RunView{}, ErrConflict
	}
	record.Attempts[len(record.Attempts)-1].State = domain.RunStateQueued
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
