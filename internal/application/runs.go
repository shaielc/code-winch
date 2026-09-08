package application

import (
	"context"
	"errors"
	"time"

	"github.com/shaielc/code-winch/internal/domain"
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

type SystemClock struct{}

func (SystemClock) Now() domain.Timestamp {
	value, _ := domain.NewTimestamp(time.Now().UTC())
	return value
}
