// Package execution drives a run that a client has started: it takes the
// supervisor lease, hands the runner its prepare and start commands, turns
// runner observations into durable events, and moves the run's domain state as
// the execution reports progress.
//
// It is a composition-layer orchestrator rather than an adapter, which is why
// it may import the concrete supervisor and runner packages: it is the wiring
// `cmd/winchd` would otherwise contain inline, kept here so it can be tested.
package execution

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/internal/runner/local"
	"github.com/shaielc/code-winch/internal/supervisor"
	"github.com/shaielc/code-winch/pkg/protocol"
)

// Runner is the subset of the runner gateway the engine drives. Bind precedes
// the prepare message so the gateway can route every later message for that
// execution without being told the profiles again.
type Runner interface {
	Send(context.Context, protocol.RunnerMessage) error
	Bind(executionID, harnessProfile, sandboxProfile string) error
	Cleanup(context.Context, string) error
}

// RunStates applies a lifecycle command the runtime observed. It is satisfied
// by the run service, which owns the state machine and its persistence.
type RunStates interface {
	Advance(context.Context, domain.RunID, domain.RunCommand) (application.RunView, error)
}

type Config struct {
	Runs       application.RunRepository
	States     RunStates
	Control    application.SupervisorStore
	Supervisor *supervisor.Supervisor
	Runner     Runner
	IDs        application.IDSource
	Clock      application.Clock
	Logger     *slog.Logger
	// StopGrace is how long a harness has to exit after a stop before the
	// sandbox escalates.
	StopGrace time.Duration
}

const defaultStopGrace = 5 * time.Second

type Engine struct {
	cfg  Config
	mu   sync.Mutex
	live map[string]*execution
	runs map[domain.RunID]*execution
}

// execution is one live run. ordinal is the engine's own runner-ordinal
// counter: the supervisor requires strictly increasing ordinals per run, and
// the engine emits lifecycle events of its own between the runner's, so it
// allocates every ordinal rather than forwarding the runner's.
//
// appending covers allocating an ordinal *and* writing under it. Allocating
// atomically is not enough: the observation loop and the launch and shutdown
// paths both write, and the store rejects an ordinal that is no longer greater
// than the last one written, so a write that overtook its predecessor would
// silently drop the slower event.
type execution struct {
	id        string
	runID     domain.RunID
	lease     application.RunLease
	harness   string
	sandbox   string
	appending sync.Mutex
	ordinal   atomic.Uint64
}

func (e *execution) next() uint64 { return e.ordinal.Add(1) }

func New(cfg Config) (*Engine, error) {
	if cfg.Runs == nil || cfg.States == nil || cfg.Control == nil || cfg.Supervisor == nil || cfg.Runner == nil || cfg.IDs == nil || cfg.Clock == nil {
		return nil, errors.New("execution engine: incomplete configuration")
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.StopGrace <= 0 {
		cfg.StopGrace = defaultStopGrace
	}
	return &Engine{cfg: cfg, live: map[string]*execution{}, runs: map[domain.RunID]*execution{}}, nil
}

func (e *Engine) track(x *execution) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.live[x.id] = x
	e.runs[x.runID] = x
}

func (e *Engine) untrack(x *execution) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.live, x.id)
	if e.runs[x.runID] == x {
		delete(e.runs, x.runID)
	}
}

func (e *Engine) byExecution(id string) *execution {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.live[id]
}

func (e *Engine) byRun(id domain.RunID) *execution {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.runs[id]
}

func message(kind, commandID string, payload any) protocol.RunnerMessage {
	body, _ := json.Marshal(payload)
	return protocol.RunnerMessage{
		Version:   protocol.RunnerVersion{Major: protocol.RunnerProtocolMajor, Minor: protocol.RunnerProtocolMinor},
		Kind:      kind,
		CommandID: commandID,
		Payload:   body,
	}
}

// Launch prepares and starts the execution for a run the client has queued.
// Every runner interaction is preceded by a durable desired state, so a daemon
// that dies mid-launch leaves a checkpoint that reconciliation can read.
func (e *Engine) Launch(ctx context.Context, id domain.RunID) error {
	record, _, err := e.cfg.Runs.Get(ctx, id)
	if err != nil {
		return err
	}
	if existing := e.byRun(id); existing != nil {
		return fmt.Errorf("execution engine: run already has a live execution: run=%s", id)
	}
	x := &execution{id: e.cfg.IDs.NewCommandID().String(), runID: id, harness: record.HarnessProfile, sandbox: record.SandboxProfile}
	x.lease, err = e.cfg.Supervisor.Acquire(ctx, id, e.cfg.IDs.NewCommandID().String())
	if err != nil {
		return err
	}
	if err = e.cfg.Runner.Bind(x.id, x.harness, x.sandbox); err != nil {
		if releaseErr := e.cfg.Supervisor.Release(ctx, x.lease); releaseErr != nil {
			e.cfg.Logger.Warn("lease not released", "component", "execution", "operation", "bind", "run_id", id.String(), "error_code", "lease_not_released")
		}
		return err
	}
	e.track(x)
	if _, err = e.cfg.States.Advance(ctx, id, domain.RunCommandAcquireLease); err != nil {
		return e.abandon(ctx, x, err)
	}
	e.emit(ctx, x, domain.RunStatePreparing, "")
	prepare := message("prepare", e.cfg.IDs.NewCommandID().String(), protocol.PreparePayload{WorkspaceID: id.String()})
	if err = e.command(ctx, x, domain.RunStatePreparing, prepare); err != nil {
		return e.abandon(ctx, x, err)
	}
	start := message("start", e.cfg.IDs.NewCommandID().String(), protocol.StartPayload{LaunchProfile: x.harness})
	if err = e.command(ctx, x, domain.RunStateRunning, start); err != nil {
		return e.abandon(ctx, x, err)
	}
	if _, err = e.cfg.States.Advance(ctx, id, domain.RunCommandExecutionStarted); err != nil {
		return e.abandon(ctx, x, err)
	}
	e.emit(ctx, x, domain.RunStateRunning, "")
	return nil
}

// abandon records a launch that failed as a failed run rather than leaving it
// queued forever, and releases everything the partial launch took.
func (e *Engine) abandon(ctx context.Context, x *execution, cause error) error {
	if _, err := e.cfg.States.Advance(ctx, x.runID, domain.RunCommandPreparationFailed); err != nil {
		e.cfg.Logger.Warn("run state not advanced", "component", "execution", "operation", "abandon", "run_id", x.runID.String(), "error_code", "state_not_advanced")
	}
	e.emit(ctx, x, domain.RunStateFailed, "LAUNCH_FAILED")
	e.release(ctx, x, domain.RunStateFailed)
	return cause
}

// command persists the desired state and then sends the runner message, in that
// order, under the run's lease.
func (e *Engine) command(ctx context.Context, x *execution, desired domain.RunState, m protocol.RunnerMessage) error {
	return e.cfg.Supervisor.Execute(ctx, x.lease, supervisor.Command{
		DesiredState:  desired,
		HarnessDriver: x.harness,
		SandboxDriver: x.sandbox,
		ExecutionID:   x.id,
		Message:       m,
	})
}

// Stop asks the runner to end the execution. The run reaches its terminal state
// from the exit observation, not from here, so a stop that races an exit does
// not invent an outcome.
func (e *Engine) Stop(ctx context.Context, id domain.RunID) error {
	x := e.byRun(id)
	if x == nil {
		// Nothing is running: the execution already exited, and the domain
		// state the caller just wrote is the whole effect of the stop.
		return nil
	}
	stop := message("stop", e.cfg.IDs.NewCommandID().String(), protocol.StopPayload{GraceMilliseconds: uint32(e.cfg.StopGrace.Milliseconds())})
	return e.command(ctx, x, domain.RunStateStopping, stop)
}

// Deliver sends one accepted input command to the running harness. It preserves
// the recorded desired state: delivering input is not a lifecycle change.
func (e *Engine) Deliver(ctx context.Context, id domain.RunID, commandID domain.CommandID, kind application.InputKind, payload application.InputPayload) error {
	x := e.byRun(id)
	if x == nil {
		return fmt.Errorf("execution engine: no live execution for run=%s", id)
	}
	control, err := e.cfg.Control.LoadControl(ctx, id)
	if err != nil {
		return err
	}
	var m protocol.RunnerMessage
	switch kind {
	case application.InputText:
		m = message("input", commandID.String(), protocol.InputPayload{InputID: commandID.String(), Text: payload.Text})
	case application.InputResize:
		m = message("resize", commandID.String(), protocol.ResizePayload{Rows: payload.Rows, Columns: payload.Columns})
	default:
		return fmt.Errorf("execution engine: undeliverable input kind: run=%s kind=%s", id, kind)
	}
	return e.command(ctx, x, control.DesiredState, m)
}

// Shutdown ends every live execution and records why. A local execution cannot
// outlive the daemon that owns its processes, so leaving the run marked running
// would be a claim about a process that is about to be killed. It must run
// before the runner is closed: the runner's pumps only finish once their
// streams are released here.
func (e *Engine) Shutdown(ctx context.Context) {
	for _, x := range e.snapshot() {
		if _, err := e.cfg.States.Advance(ctx, x.runID, domain.RunCommandFailedExit); err != nil {
			if _, err = e.cfg.States.Advance(ctx, x.runID, domain.RunCommandPreparationFailed); err != nil {
				e.cfg.Logger.Warn("run state not advanced", "component", "execution", "operation", "shutdown", "run_id", x.runID.String(), "error_code", "state_not_advanced")
			}
		}
		e.emit(ctx, x, domain.RunStateFailed, "DAEMON_SHUTDOWN")
		e.release(ctx, x, domain.RunStateFailed)
	}
}

func (e *Engine) snapshot() []*execution {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]*execution, 0, len(e.live))
	for _, x := range e.live {
		out = append(out, x)
	}
	return out
}

// Consume turns runner observations into durable events until the stream
// closes, which the runner does only after every pump has finished.
func (e *Engine) Consume(ctx context.Context, observations <-chan local.Observation) {
	for observation := range observations {
		e.observe(ctx, observation)
	}
}

func (e *Engine) observe(ctx context.Context, observation local.Observation) {
	x := e.byExecution(observation.ExecutionID)
	if x == nil {
		// An observation for an execution this daemon no longer tracks has no
		// lease to be written under, and writing it unfenced is what the lease
		// exists to prevent.
		return
	}
	switch observation.Type {
	case "output":
		if observation.Event != nil {
			e.persist(ctx, x, *observation.Event)
		}
	case "exit":
		exit := application.HarnessExit{Code: "UNKNOWN"}
		if observation.Exit != nil {
			exit = *observation.Exit
		}
		e.finish(ctx, x, exit)
	}
}

func (e *Engine) persist(ctx context.Context, x *execution, event application.UnsequencedEvent) {
	if event.EventID.IsZero() {
		event.EventID = e.cfg.IDs.NewEventID()
	}
	if event.OccurredAt.Time().IsZero() {
		event.OccurredAt = e.cfg.Clock.Now()
	}
	x.appending.Lock()
	defer x.appending.Unlock()
	if _, err := e.cfg.Supervisor.Observe(ctx, x.lease, x.next(), []application.UnsequencedEvent{event}); err != nil {
		// The message may name the redaction that refused the event, so only
		// the stable class is logged.
		e.cfg.Logger.Error("event not recorded", "component", "execution", "operation", "observe", "run_id", x.runID.String(), "error_code", "event_not_recorded")
	}
}

// finish records the harness exit as the run's terminal state and releases the
// execution.
func (e *Engine) finish(ctx context.Context, x *execution, exit application.HarnessExit) {
	command := domain.RunCommandFailedExit
	state := domain.RunStateFailed
	if exit.Successful {
		command, state = domain.RunCommandSuccessfulExit, domain.RunStateCompleted
	}
	view, err := e.cfg.States.Advance(ctx, x.runID, command)
	if err != nil {
		e.cfg.Logger.Warn("run state not advanced", "component", "execution", "operation", "exit", "run_id", x.runID.String(), "error_code", "state_not_advanced")
	} else if len(view.Record.Attempts) > 0 {
		// A stop that raced the exit, or a run already terminal, keeps whatever
		// state the machine actually reached.
		state = view.Record.Attempts[len(view.Record.Attempts)-1].State
	}
	e.emit(ctx, x, state, exit.Code)
	e.release(ctx, x, state)
}

// release ends the run's execution: it records the terminal desired state under
// the lease, cleans the sandbox up, and drops the lease so no later write can
// be made in this run's name.
func (e *Engine) release(ctx context.Context, x *execution, state domain.RunState) {
	if _, err := e.cfg.Control.SaveDesiredState(ctx, x.lease, state, x.harness, x.sandbox, x.id); err != nil {
		e.cfg.Logger.Warn("desired state not recorded", "component", "execution", "operation", "release", "run_id", x.runID.String(), "error_code", "desired_state_not_recorded")
	}
	if err := e.cfg.Runner.Cleanup(ctx, x.id); err != nil {
		e.cfg.Logger.Warn("execution not cleaned up", "component", "execution", "operation", "release", "run_id", x.runID.String(), "error_code", "cleanup_failed")
	}
	if err := e.cfg.Supervisor.Release(ctx, x.lease); err != nil {
		e.cfg.Logger.Warn("lease not released", "component", "execution", "operation", "release", "run_id", x.runID.String(), "error_code", "lease_not_released")
	}
	e.untrack(x)
}

// emit records a control-plane lifecycle event so a reader of the event history
// sees the run's state changes beside the harness output that caused them.
func (e *Engine) emit(ctx context.Context, x *execution, state domain.RunState, reason string) {
	payload := map[string]string{"state": string(state)}
	if reason != "" {
		payload["reasonCode"] = reason
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}
	e.persist(ctx, x, application.UnsequencedEvent{
		Kind:          "run.lifecycle",
		SchemaVersion: 1,
		Source:        protocol.Source{Type: "control-plane"},
		Sensitivity:   protocol.SensitivityOperational,
		Payload:       body,
	})
}

var _ application.RunExecutor = (*Engine)(nil)
