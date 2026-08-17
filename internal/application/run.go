package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/pkg/protocol"
)

type RunView struct {
	Record       RunRecord
	Version      uint64
	LastSequence uint64
}
type CreateRunCommand struct{ WorkspacePath, HarnessProfile, SandboxProfile string }
type AdmissionHook interface {
	AdmitStart(context.Context, RunView) error
}
type AdmissionFunc func(context.Context, RunView) error

func (f AdmissionFunc) AdmitStart(ctx context.Context, r RunView) error { return f(ctx, r) }
func PermissiveAdmission(context.Context, RunView) error                { return nil }

// RunExecutor performs the runtime half of a lifecycle command: the service
// persists what the run should be doing, and the executor makes it so. A
// service without one accepts starts that never reach a harness, so the
// composition root always sets it; only tests of persistence alone leave it
// unset.
type RunExecutor interface {
	Launch(context.Context, domain.RunID) error
	Stop(context.Context, domain.RunID) error
}

type RunService struct {
	runs      RunRepository
	events    EventStore
	ids       IDSource
	clock     Clock
	drivers   *DriverRegistry
	admission AdmissionHook
	executor  RunExecutor
}

func NewRunService(r RunRepository, e EventStore, ids IDSource, clock Clock, drivers *DriverRegistry, admission AdmissionHook) *RunService {
	if admission == nil {
		admission = AdmissionFunc(PermissiveAdmission)
	}
	return &RunService{runs: r, events: e, ids: ids, clock: clock, drivers: drivers, admission: admission}
}

// WithExecutor attaches the runtime. It is separate from NewRunService because
// the executor needs the service to advance run states, so the two are wired
// after both exist.
func (s *RunService) WithExecutor(executor RunExecutor) *RunService {
	s.executor = executor
	return s
}

func (s *RunService) CreateRun(ctx context.Context, c CreateRunCommand) (RunView, error) {
	if c.WorkspacePath == "" || c.HarnessProfile == "" || c.SandboxProfile == "" {
		return RunView{}, ErrUnknownProfile
	}
	h, err := s.drivers.Harness(c.HarnessProfile)
	if err != nil {
		return RunView{}, err
	}
	sb, err := s.drivers.Sandbox(c.SandboxProfile)
	if err != nil {
		return RunView{}, err
	}
	desc, err := h.Describe(ctx)
	if err != nil {
		return RunView{}, fmt.Errorf("describe harness: %w", err)
	}
	caps := sb.Capabilities(ctx)
	resolved, _ := json.Marshal(map[string]any{"harness": map[string]any{"id": desc.ID, "version": desc.Version}, "sandbox": map[string]any{"profile": c.SandboxProfile, "isolation": caps.Isolation}})
	run, err := domain.NewRun(s.ids.NewRunID(), s.ids.NewAttemptID())
	if err != nil {
		return RunView{}, err
	}
	now := s.clock.Now()
	rec := RunRecord{ID: run.ID(), Attempts: run.Attempts(), WorkspacePath: c.WorkspacePath, HarnessProfile: c.HarnessProfile, SandboxProfile: c.SandboxProfile, ResolvedConfiguration: resolved, CreatedAt: now, UpdatedAt: now}
	v, err := s.runs.Save(ctx, rec, 0)
	return RunView{Record: rec, Version: v}, err
}
func (s *RunService) GetRun(ctx context.Context, id domain.RunID) (RunView, error) {
	rec, v, err := s.runs.Get(ctx, id)
	if err != nil {
		return RunView{}, err
	}
	last, err := s.events.LastSequence(ctx, id)
	if err != nil {
		return RunView{}, err
	}
	return RunView{rec, v, last}, nil
}

// shortestPath drives a fresh attempt to a stored state. The persisted model
// keeps only the state of each attempt, not the route it took there, so any
// route reaching the same state restores the same record.
var shortestPath = map[domain.RunState][]domain.RunCommand{
	domain.RunStateQueued:    {domain.RunCommandStart},
	domain.RunStatePreparing: {domain.RunCommandStart, domain.RunCommandAcquireLease},
	domain.RunStateRunning:   {domain.RunCommandStart, domain.RunCommandAcquireLease, domain.RunCommandExecutionStarted},
	domain.RunStateStopping:  {domain.RunCommandStart, domain.RunCommandAcquireLease, domain.RunCommandExecutionStarted, domain.RunCommandStop},
	domain.RunStateCancelled: {domain.RunCommandCancel},
	domain.RunStateCompleted: {domain.RunCommandStart, domain.RunCommandAcquireLease, domain.RunCommandExecutionStarted, domain.RunCommandSuccessfulExit},
	domain.RunStateFailed:    {domain.RunCommandStart, domain.RunCommandAcquireLease, domain.RunCommandPreparationFailed},
}

// hydrate rebuilds a run from its stored attempts by replaying commands, so the
// domain validates the record rather than trusting it.
//
// Retry is the only command that appends an attempt and the domain accepts it
// only from Failed, so every attempt but the last is necessarily Failed. Each
// one is therefore driven to Failed before the next is appended; a stored
// record that disagrees is corrupt and is reported rather than repaired.
func hydrate(rec RunRecord) (*domain.Run, error) {
	if len(rec.Attempts) == 0 {
		return nil, errors.New("run has no attempt")
	}
	r, err := domain.NewRun(rec.ID, rec.Attempts[0].ID)
	if err != nil {
		return nil, err
	}
	for i, a := range rec.Attempts[1:] {
		if previous := rec.Attempts[i].State; previous != domain.RunStateFailed {
			return nil, fmt.Errorf("run %s attempt %d is %s, but only a failed attempt can be retried", rec.ID, i+1, previous)
		}
		if err = replay(r, domain.RunStateFailed); err != nil {
			return nil, err
		}
		if err = r.Apply(domain.RunCommandRetry, a.ID); err != nil {
			return nil, err
		}
	}
	return r, replay(r, rec.Attempts[len(rec.Attempts)-1].State)
}

// replay drives the current attempt to state. A first attempt starts at
// Created and a retried one at Queued, so the leading Start is skipped for the
// latter; every other command is reachable from both entry points.
func replay(r *domain.Run, state domain.RunState) error {
	for _, c := range shortestPath[state] {
		if c == domain.RunCommandStart && r.CurrentAttempt().State == domain.RunStateQueued {
			continue
		}
		if err := r.Apply(c, domain.AttemptID{}); err != nil {
			return err
		}
	}
	return nil
}
func (s *RunService) transition(ctx context.Context, id domain.RunID, expected uint64, command domain.RunCommand) (RunView, error) {
	rec, v, err := s.runs.Get(ctx, id)
	if err != nil {
		return RunView{}, err
	}
	if v != expected {
		return RunView{}, ErrConflict
	}
	run, err := hydrate(rec)
	if err != nil {
		return RunView{}, err
	}
	if command == domain.RunCommandStart {
		if _, err = s.drivers.Harness(rec.HarnessProfile); err != nil {
			return RunView{}, err
		}
		if _, err = s.drivers.Sandbox(rec.SandboxProfile); err != nil {
			return RunView{}, err
		}
		if err = s.admission.AdmitStart(ctx, RunView{rec, v, 0}); err != nil {
			return RunView{}, err
		}
	}
	before := run.CurrentAttempt()
	if err = run.Apply(command, domain.AttemptID{}); err != nil {
		return RunView{}, err
	}
	// A stop in Stopping or a terminal state is accepted and changes nothing.
	// Saving anyway would move the version, and the version is the ETag: a
	// client holding a valid one for a finished run would then be refused
	// because somebody else stopped it again.
	if run.CurrentAttempt() == before {
		return RunView{rec, v, 0}, nil
	}
	rec.Attempts = run.Attempts()
	rec.UpdatedAt = s.clock.Now()
	nv, err := s.runs.Save(ctx, rec, v)
	return RunView{rec, nv, 0}, err
}

// StartRun queues the run durably, then launches it. The view it returns is
// read back after the launch so the caller's ETag matches the run the launch
// left behind rather than the queued one it replaced.
func (s *RunService) StartRun(ctx context.Context, id domain.RunID, v uint64) (RunView, error) {
	view, err := s.transition(ctx, id, v, domain.RunCommandStart)
	if err != nil || s.executor == nil {
		return view, err
	}
	if launchErr := s.executor.Launch(ctx, id); launchErr != nil {
		// The executor has already recorded the failed run; report the view it
		// left behind together with the failure that produced it.
		if view, err = s.GetRun(ctx, id); err != nil {
			return RunView{}, err
		}
		return view, launchErr
	}
	return s.GetRun(ctx, id)
}

func (s *RunService) StopRun(ctx context.Context, id domain.RunID, v uint64) (RunView, error) {
	view, err := s.transition(ctx, id, v, domain.RunCommandStop)
	if err != nil || s.executor == nil {
		return view, err
	}
	if err = s.executor.Stop(ctx, id); err != nil {
		return view, err
	}
	return s.GetRun(ctx, id)
}

// Advance applies a lifecycle command the runtime observed rather than one a
// client requested. There is no expected version to check: the supervisor
// writes to the same row while an execution runs, so a version that moved in
// between is retried instead of reported as a conflict.
func (s *RunService) Advance(ctx context.Context, id domain.RunID, command domain.RunCommand) (RunView, error) {
	var conflict error
	for attempt := 0; attempt < advanceAttempts; attempt++ {
		rec, v, err := s.runs.Get(ctx, id)
		if err != nil {
			return RunView{}, err
		}
		run, err := hydrate(rec)
		if err != nil {
			return RunView{}, err
		}
		if err = run.Apply(command, domain.AttemptID{}); err != nil {
			return RunView{}, err
		}
		rec.Attempts = run.Attempts()
		rec.UpdatedAt = s.clock.Now()
		nv, err := s.runs.Save(ctx, rec, v)
		if errors.Is(err, ErrConflict) {
			conflict = err
			continue
		}
		if err != nil {
			return RunView{}, err
		}
		return RunView{Record: rec, Version: nv}, nil
	}
	return RunView{}, conflict
}

const advanceAttempts = 5

func (s *RunService) ListRunEvents(ctx context.Context, id domain.RunID, after uint64, limit int) ([]protocol.Event, error) {
	if _, _, err := s.runs.Get(ctx, id); err != nil {
		return nil, err
	}
	return s.events.Read(ctx, id, after, limit)
}

type SystemClock struct{}

func (SystemClock) Now() domain.Timestamp { t, _ := domain.NewTimestamp(time.Now().UTC()); return t }
