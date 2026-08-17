package execution

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/internal/supervisor"
)

// InFlightRuns is the half of the run repository a sweep needs: which runs
// claim to be doing something, and what each one currently says it is doing.
type InFlightRuns interface {
	InFlight(context.Context) ([]domain.RunID, error)
	Get(context.Context, domain.RunID) (application.RunRecord, uint64, error)
}

// Reconciler makes the durable state of a restarted daemon truthful.
//
// A local execution cannot outlive the daemon that owns its processes, so
// every run still marked in flight when this daemon starts is either a run
// whose harness died with the previous daemon or, in a future deployment, one
// whose execution can be adopted. The supervisor decides which; this type
// finds the candidates, asks it about each, and moves the run's attempt to
// match the answer.
//
// Both halves are needed. Supervisor.Reconcile records its conclusion as the
// run's desired state, but the API reports the attempt state, so a sweep that
// only reconciled would leave a dead run still answering "running" to every
// reader.
type Reconciler struct {
	runs       InFlightRuns
	states     RunStates
	supervisor *supervisor.Supervisor
	runner     application.ReconciliationRunner
	ids        application.IDSource
	logger     *slog.Logger
}

func NewReconciler(runs InFlightRuns, states RunStates, s *supervisor.Supervisor, runner application.ReconciliationRunner, ids application.IDSource, logger *slog.Logger) *Reconciler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Reconciler{runs: runs, states: states, supervisor: s, runner: runner, ids: ids, logger: logger}
}

// Sweep reconciles every in-flight run and reports how many it settled. One
// run's failure does not abandon the rest: a daemon that cannot reconcile a
// single run must still tell the truth about the others, so failures are
// logged by stable code and the sweep continues. Only an error listing the
// runs at all is returned, because then nothing was reconciled.
func (r *Reconciler) Sweep(ctx context.Context) (int, error) {
	ids, err := r.runs.InFlight(ctx)
	if err != nil {
		return 0, err
	}
	settled := 0
	for _, id := range ids {
		if r.reconcile(ctx, id) {
			settled++
		}
	}
	if len(ids) > 0 {
		r.logger.Info("runs reconciled", "component", "execution", "operation", "reconcile", "status", "ready", "size", settled)
	}
	return settled, nil
}

func (r *Reconciler) reconcile(ctx context.Context, id domain.RunID) bool {
	control, err := r.supervisor.Reconcile(ctx, id, r.ids.NewCommandID().String(), r.runner)
	if err != nil {
		// Reconcile's error can name the adapter that refused; log the class.
		r.logger.Warn("run not reconciled", "component", "execution", "operation", "reconcile", "run_id", id.String(), "error_code", "reconcile_failed")
		return false
	}
	if !control.DesiredState.IsTerminal() {
		// The execution was adopted and is still running. No local runner can
		// do this today, so this is the seam a remote runner grows into.
		return false
	}
	if _, err = r.advance(ctx, id, control.DesiredState); err != nil {
		r.logger.Warn("run state not advanced", "component", "execution", "operation", "reconcile", "run_id", id.String(), "error_code", "state_not_advanced")
		return false
	}
	return true
}

// advance moves the attempt to the terminal state reconciliation reached.
//
// The supervisor's vocabulary is wider than the attempt machine's: it can
// conclude Cancelled for an execution that exited under a stop, which the
// domain reaches only from Created or Queued. Where the two disagree the
// attempt records the only terminal the domain permits from where it stands,
// because a run whose daemon died did stop for a reason the operator needs to
// see, and refusing to write anything would leave the louder lie in place.
func (r *Reconciler) advance(ctx context.Context, id domain.RunID, reconciled domain.RunState) (application.RunView, error) {
	record, _, err := r.runs.Get(ctx, id)
	if err != nil {
		return application.RunView{}, err
	}
	if len(record.Attempts) == 0 {
		return application.RunView{}, fmt.Errorf("execution reconciler: run has no attempt: run=%s", id)
	}
	command := domain.RunCommandFailedExit
	switch record.Attempts[len(record.Attempts)-1].State {
	case domain.RunStateQueued:
		// Nothing ever ran, so no harness failed it: the run was abandoned
		// before it reached one, and Cancelled is the domain's word for that.
		command = domain.RunCommandCancel
	case domain.RunStatePreparing:
		command = domain.RunCommandPreparationFailed
	case domain.RunStateRunning, domain.RunStateStopping:
		if reconciled == domain.RunStateCompleted {
			command = domain.RunCommandSuccessfulExit
		}
	}
	return r.states.Advance(ctx, id, command)
}
