package domain

import "fmt"

// RunState is the persisted lifecycle state of one execution attempt.
type RunState string

const (
	RunStateCreated   RunState = "created"
	RunStateQueued    RunState = "queued"
	RunStatePreparing RunState = "preparing"
	RunStateRunning   RunState = "running"
	RunStateStopping  RunState = "stopping"
	RunStateCompleted RunState = "completed"
	RunStateFailed    RunState = "failed"
	RunStateCancelled RunState = "cancelled"
)

func (state RunState) IsTerminal() bool {
	return state == RunStateCompleted || state == RunStateFailed || state == RunStateCancelled
}

// RunCommand names a lifecycle fact or request accepted by the domain.
type RunCommand string

const (
	RunCommandStart             RunCommand = "start"
	RunCommandAcquireLease      RunCommand = "acquire_lease"
	RunCommandExecutionStarted  RunCommand = "execution_started"
	RunCommandPreparationFailed RunCommand = "preparation_failed"
	RunCommandStop              RunCommand = "stop"
	RunCommandSuccessfulExit    RunCommand = "successful_exit"
	RunCommandFailedExit        RunCommand = "failed_exit"
	RunCommandCancel            RunCommand = "cancel"
	RunCommandRetry             RunCommand = "retry"
)

const (
	ErrorCodeInvalidRun        = "invalid_run"
	ErrorCodeInvalidAttempt    = "invalid_attempt"
	ErrorCodeIllegalTransition = "illegal_run_transition"
)

// RunError contains only bounded lifecycle metadata, so it is safe to report
// without leaking run content or secrets.
type RunError struct {
	Code      string
	RunID     RunID
	AttemptID AttemptID
	State     RunState
	Command   RunCommand
	Summary   string
}

func (err *RunError) Error() string {
	return fmt.Sprintf("%s: %s (run=%s attempt=%s state=%s command=%s)",
		err.Code, err.Summary, err.RunID, err.AttemptID, err.State, err.Command)
}

// Attempt is an immutable view of an attempt in a run's history.
type Attempt struct {
	ID                AttemptID
	PreviousAttemptID AttemptID
	State             RunState
}

// Run owns an append-only attempt history. Only the current attempt can move;
// Retry appends a linked attempt instead of rewriting the failed attempt.
type Run struct {
	id       RunID
	attempts []Attempt
}

func NewRun(runID RunID, attemptID AttemptID) (*Run, error) {
	if runID.IsZero() {
		return nil, &RunError{Code: ErrorCodeInvalidRun, RunID: runID, AttemptID: attemptID, Summary: "run ID must be non-zero"}
	}
	if attemptID.IsZero() {
		return nil, &RunError{Code: ErrorCodeInvalidAttempt, RunID: runID, AttemptID: attemptID, Summary: "attempt ID must be non-zero"}
	}
	return &Run{id: runID, attempts: []Attempt{{ID: attemptID, State: RunStateCreated}}}, nil
}

func (run *Run) ID() RunID { return run.id }

func (run *Run) CurrentAttempt() Attempt { return run.attempts[len(run.attempts)-1] }

func (run *Run) Attempts() []Attempt {
	return append([]Attempt(nil), run.attempts...)
}

// Apply performs one lifecycle command. A non-zero nextAttemptID is required
// only for Retry and is ignored by no other command.
func (run *Run) Apply(command RunCommand, nextAttemptID AttemptID) error {
	current := run.CurrentAttempt()

	if command == RunCommandStop && (current.State == RunStateStopping || current.State.IsTerminal()) {
		return nil
	}
	if command == RunCommandRetry {
		return run.retry(current, nextAttemptID)
	}

	next, allowed := nextState(current.State, command)
	if !allowed {
		return run.transitionError(current, command)
	}
	run.attempts[len(run.attempts)-1].State = next
	return nil
}

func (run *Run) retry(current Attempt, nextAttemptID AttemptID) error {
	if current.State != RunStateFailed {
		return run.transitionError(current, RunCommandRetry)
	}
	if nextAttemptID.IsZero() || run.hasAttempt(nextAttemptID) {
		return &RunError{
			Code: ErrorCodeInvalidAttempt, RunID: run.id, AttemptID: nextAttemptID,
			State: current.State, Command: RunCommandRetry,
			Summary: "retry attempt ID must be non-zero and unique within the run",
		}
	}
	run.attempts = append(run.attempts, Attempt{
		ID: nextAttemptID, PreviousAttemptID: current.ID, State: RunStateQueued,
	})
	return nil
}

func (run *Run) hasAttempt(id AttemptID) bool {
	for _, attempt := range run.attempts {
		if attempt.ID == id {
			return true
		}
	}
	return false
}

func (run *Run) transitionError(current Attempt, command RunCommand) error {
	return &RunError{
		Code: ErrorCodeIllegalTransition, RunID: run.id, AttemptID: current.ID,
		State: current.State, Command: command,
		Summary: "command is not allowed from the current attempt state",
	}
}

func nextState(state RunState, command RunCommand) (RunState, bool) {
	switch state {
	case RunStateCreated:
		if command == RunCommandStart {
			return RunStateQueued, true
		}
		if command == RunCommandCancel {
			return RunStateCancelled, true
		}
	case RunStateQueued:
		if command == RunCommandAcquireLease {
			return RunStatePreparing, true
		}
		if command == RunCommandCancel {
			return RunStateCancelled, true
		}
	case RunStatePreparing:
		if command == RunCommandExecutionStarted {
			return RunStateRunning, true
		}
		if command == RunCommandPreparationFailed {
			return RunStateFailed, true
		}
	case RunStateRunning:
		if command == RunCommandStop {
			return RunStateStopping, true
		}
		if command == RunCommandSuccessfulExit {
			return RunStateCompleted, true
		}
		if command == RunCommandFailedExit {
			return RunStateFailed, true
		}
	case RunStateStopping:
		if command == RunCommandSuccessfulExit {
			return RunStateCompleted, true
		}
		if command == RunCommandFailedExit {
			return RunStateFailed, true
		}
	}
	return "", false
}
