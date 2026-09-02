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
}

type RunService struct {
	repository RunRepository
	clock      Clock
	ids        IDSource
}

func NewRunService(repository RunRepository, clock Clock, ids IDSource) (*RunService, error) {
	if repository == nil || clock == nil || ids == nil {
		return nil, errors.New("run service: repository, clock, and ID source are required")
	}
	return &RunService{repository: repository, clock: clock, ids: ids}, nil
}

func (s *RunService) Create(ctx context.Context, command CreateRunCommand) (RunView, error) {
	if command.WorkspacePath == "" || command.HarnessProfile == "" || command.SandboxProfile == "" {
		return RunView{}, errors.New("run service: required run fields are missing")
	}
	run, err := domain.NewRun(s.ids.NewRunID(), s.ids.NewAttemptID())
	if err != nil {
		return RunView{}, err
	}
	now := s.clock.Now().Time()
	record := RunRecord{ID: run.ID(), Attempts: run.Attempts(), CreatedAt: now, UpdatedAt: now, WorkspacePath: command.WorkspacePath, HarnessProfile: command.HarnessProfile, SandboxProfile: command.SandboxProfile}
	version, err := s.repository.Save(ctx, record, 0)
	return RunView{Record: record, Version: version}, err
}

func (s *RunService) Get(ctx context.Context, id domain.RunID) (RunView, error) {
	record, version, err := s.repository.Get(ctx, id)
	return RunView{Record: record, Version: version}, err
}

type SystemClock struct{}

func (SystemClock) Now() domain.Timestamp {
	value, _ := domain.NewTimestamp(time.Now().UTC())
	return value
}
