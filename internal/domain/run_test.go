package domain

import (
	"errors"
	"testing"
)

const secondAttemptID = "223e4567-e89b-12d3-a456-426614174000"

func mustRunIDs(t *testing.T) (RunID, AttemptID, AttemptID) {
	t.Helper()
	runID, err := ParseRunID(validID)
	if err != nil {
		t.Fatal(err)
	}
	attemptID, err := ParseAttemptID(validID)
	if err != nil {
		t.Fatal(err)
	}
	nextID, err := ParseAttemptID(secondAttemptID)
	if err != nil {
		t.Fatal(err)
	}
	return runID, attemptID, nextID
}

func TestRunEveryStateCommandPair(t *testing.T) {
	states := []RunState{
		RunStateCreated, RunStateQueued, RunStatePreparing, RunStateRunning,
		RunStateStopping, RunStateCompleted, RunStateFailed, RunStateCancelled,
	}
	commands := []RunCommand{
		RunCommandStart, RunCommandAcquireLease, RunCommandExecutionStarted,
		RunCommandPreparationFailed, RunCommandStop, RunCommandSuccessfulExit,
		RunCommandFailedExit, RunCommandCancel, RunCommandRetry,
	}
	type transition struct {
		state   RunState
		command RunCommand
	}
	want := map[transition]RunState{
		{RunStateCreated, RunCommandStart}:               RunStateQueued,
		{RunStateCreated, RunCommandCancel}:              RunStateCancelled,
		{RunStateQueued, RunCommandAcquireLease}:         RunStatePreparing,
		{RunStateQueued, RunCommandCancel}:               RunStateCancelled,
		{RunStatePreparing, RunCommandExecutionStarted}:  RunStateRunning,
		{RunStatePreparing, RunCommandPreparationFailed}: RunStateFailed,
		{RunStateRunning, RunCommandStop}:                RunStateStopping,
		{RunStateRunning, RunCommandSuccessfulExit}:      RunStateCompleted,
		{RunStateRunning, RunCommandFailedExit}:          RunStateFailed,
		{RunStateStopping, RunCommandStop}:               RunStateStopping,
		{RunStateStopping, RunCommandSuccessfulExit}:     RunStateCompleted,
		{RunStateStopping, RunCommandFailedExit}:         RunStateFailed,
		{RunStateCompleted, RunCommandStop}:              RunStateCompleted,
		{RunStateFailed, RunCommandStop}:                 RunStateFailed,
		{RunStateFailed, RunCommandRetry}:                RunStateQueued,
		{RunStateCancelled, RunCommandStop}:              RunStateCancelled,
	}

	for _, state := range states {
		for _, command := range commands {
			t.Run(string(state)+"/"+string(command), func(t *testing.T) {
				runID, attemptID, nextID := mustRunIDs(t)
				run := &Run{id: runID, attempts: []Attempt{{ID: attemptID, State: state}}}
				before := run.Attempts()
				err := run.Apply(command, nextID)
				expected, allowed := want[transition{state, command}]
				if !allowed {
					var domainErr *RunError
					if !errors.As(err, &domainErr) || domainErr.Code != ErrorCodeIllegalTransition {
						t.Fatalf("error = %v, want %s RunError", err, ErrorCodeIllegalTransition)
					}
					if got := run.Attempts(); len(got) != len(before) || got[0] != before[0] {
						t.Fatalf("illegal transition mutated attempts: before=%v after=%v", before, got)
					}
					return
				}
				if err != nil {
					t.Fatalf("Apply() error = %v", err)
				}
				if got := run.CurrentAttempt().State; got != expected {
					t.Fatalf("state = %q, want %q", got, expected)
				}
			})
		}
	}
}

func TestRetryAppendsLinkedAttemptWithoutChangingHistory(t *testing.T) {
	runID, attemptID, nextID := mustRunIDs(t)
	run := &Run{id: runID, attempts: []Attempt{{ID: attemptID, State: RunStateFailed}}}

	if err := run.Apply(RunCommandRetry, nextID); err != nil {
		t.Fatal(err)
	}
	attempts := run.Attempts()
	if len(attempts) != 2 {
		t.Fatalf("attempt count = %d, want 2", len(attempts))
	}
	if attempts[0] != (Attempt{ID: attemptID, State: RunStateFailed}) {
		t.Fatalf("prior attempt changed: %v", attempts[0])
	}
	if attempts[1] != (Attempt{ID: nextID, PreviousAttemptID: attemptID, State: RunStateQueued}) {
		t.Fatalf("new attempt = %v", attempts[1])
	}

	// Returned history is a snapshot, not a mutation path into the aggregate.
	attempts[0].State = RunStateCreated
	if run.Attempts()[0].State != RunStateFailed {
		t.Fatal("Attempts exposed mutable history")
	}
}

func TestRetryRejectsInvalidAttemptIDWithoutMutation(t *testing.T) {
	runID, attemptID, _ := mustRunIDs(t)
	for name, nextID := range map[string]AttemptID{"zero": {}, "duplicate": attemptID} {
		t.Run(name, func(t *testing.T) {
			run := &Run{id: runID, attempts: []Attempt{{ID: attemptID, State: RunStateFailed}}}
			err := run.Apply(RunCommandRetry, nextID)
			var domainErr *RunError
			if !errors.As(err, &domainErr) || domainErr.Code != ErrorCodeInvalidAttempt {
				t.Fatalf("error = %v, want %s RunError", err, ErrorCodeInvalidAttempt)
			}
			if got := run.Attempts(); len(got) != 1 || got[0].State != RunStateFailed {
				t.Fatalf("invalid retry mutated history: %v", got)
			}
		})
	}
}

func TestNewRunValidatesIDs(t *testing.T) {
	runID, attemptID, _ := mustRunIDs(t)
	tests := []struct {
		name      string
		runID     RunID
		attemptID AttemptID
		code      string
	}{
		{"run ID", RunID{}, attemptID, ErrorCodeInvalidRun},
		{"attempt ID", runID, AttemptID{}, ErrorCodeInvalidAttempt},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewRun(test.runID, test.attemptID)
			var domainErr *RunError
			if !errors.As(err, &domainErr) || domainErr.Code != test.code {
				t.Fatalf("error = %v, want code %q", err, test.code)
			}
		})
	}

	run, err := NewRun(runID, attemptID)
	if err != nil {
		t.Fatal(err)
	}
	if run.ID() != runID || run.CurrentAttempt() != (Attempt{ID: attemptID, State: RunStateCreated}) {
		t.Fatalf("NewRun() = %#v", run)
	}
}
