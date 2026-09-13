// Package local pumps a harness codec over an attached sandbox execution.
package local

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/pkg/protocol"
)

// A stream reaching EOF and the driver recording an exit code are separate
// events: a PTY slave closes the moment the process exits, but the driver's
// wait may not have run yet. Reading Inspect once therefore reports a running
// execution with no code, which maps to a spurious failure. The pump waits for
// the code instead, and falls back only if the sandbox never produces one.
const (
	exitWaitTimeout  = 5 * time.Second
	exitWaitInterval = time.Millisecond
	unknownExitCode  = -1
)

var knownKinds = map[string]bool{
	"prepare": true, "start": true, "input": true, "resize": true,
	"stop": true, "inspect": true, "cleanup": true,
}

// Observation is runner-local output. Ordinal orders observations but is not a
// canonical event sequence; persistence remains the application's concern.
type Observation struct {
	ExecutionID string
	Ordinal     uint64
	Type        string
	Event       *application.UnsequencedEvent
	Exit        *application.HarnessExit
}

// execution keeps three independent locks so that no caller can starve the
// pump:
//
//	mu      guards the fields below and is never held across a blocking call
//	codecMu serializes codec calls, since HarnessCodec promises no concurrency
//	        safety and this is the one place Encode and Consume overlap
//	writeMu serializes stream writes and is acquired without mu
//
// The pump is the only reader draining the stream. Holding mu across a write
// would let an input that blocks on a full terminal buffer stop the pump from
// emitting, and the two would wait on each other forever.
type execution struct {
	mu       sync.Mutex
	codecMu  sync.Mutex
	writeMu  sync.Mutex
	lease    string
	prepared application.PreparedSandbox
	handle   application.ExecutionHandle
	codec    application.HarnessCodec
	runID    domain.RunID
	stream   io.ReadWriteCloser
	starting bool
	started  bool
	ordinal  atomic.Uint64
}

type Runner struct {
	sandbox      application.SandboxDriver
	harness      application.HarnessDriver
	observations chan Observation
	closed       chan struct{}
	closeOnce    sync.Once
	lifecycle    sync.Mutex
	pumps        sync.WaitGroup
	mu           sync.Mutex
	executions   map[string]*execution
}

func New(sandbox application.SandboxDriver, harness application.HarnessDriver) *Runner {
	return &Runner{
		sandbox:      sandbox,
		harness:      harness,
		observations: make(chan Observation, 64),
		closed:       make(chan struct{}),
		executions:   map[string]*execution{},
	}
}

func (r *Runner) Observations() <-chan Observation { return r.observations }

// Close releases every pump and closes the observation channel once they have
// all finished, so a consumer ranging over Observations terminates. Callers
// must stop sending commands first; Close is safe to call more than once.
func (r *Runner) Close() {
	r.closeOnce.Do(func() {
		r.lifecycle.Lock()
		close(r.closed)
		r.lifecycle.Unlock()
		r.pumps.Wait()
		close(r.observations)
	})
}

// emit delivers an observation unless the runner is closing, so a pump can
// never outlive Close by blocking on a channel nobody is reading. Pumps are the
// only senders, which is what makes closing the channel in Close safe.
func (r *Runner) emit(o Observation) {
	select {
	case r.observations <- o:
	case <-r.closed:
	}
}

// startPump registers a pump under the lifecycle lock so that no pump can be
// added after Close has begun waiting for them.
func (r *Runner) startPump(id string, e *execution) bool {
	r.lifecycle.Lock()
	defer r.lifecycle.Unlock()
	select {
	case <-r.closed:
		return false
	default:
	}
	r.pumps.Add(1)
	go r.pump(id, e)
	return true
}

func protocolError(m protocol.RunnerMessage, code, field string) error {
	return &protocol.RunnerProtocolError{Code: code, CommandID: m.CommandID, ExecutionID: m.ExecutionID, Field: field}
}

func (r *Runner) Send(ctx context.Context, m protocol.RunnerMessage) error {
	if !knownKinds[m.Kind] {
		return protocolError(m, "RUNNER_UNKNOWN_MESSAGE_KIND", "kind")
	}
	if m.ExecutionID == "" || m.LeaseToken == "" {
		return protocolError(m, "RUNNER_REQUIRED_FIELD", "executionId")
	}
	if m.Kind == "prepare" {
		return r.prepare(ctx, m)
	}
	r.mu.Lock()
	e := r.executions[m.ExecutionID]
	r.mu.Unlock()
	if e == nil {
		return protocolError(m, "RUNNER_EXECUTION_NOT_FOUND", "executionId")
	}
	e.mu.Lock()
	stale := e.lease != m.LeaseToken
	started, handle, stream := e.started, e.handle, e.stream
	e.mu.Unlock()
	if stale {
		return protocolError(m, "RUNNER_STALE_LEASE", "leaseToken")
	}
	switch m.Kind {
	case "start":
		return r.start(ctx, m, e)
	case "input":
		if !started || stream == nil {
			return protocolError(m, "RUNNER_NOT_READY", "executionId")
		}
		return r.input(m, e, stream)
	case "resize":
		if !started {
			return protocolError(m, "RUNNER_NOT_READY", "executionId")
		}
		var p protocol.ResizePayload
		if json.Unmarshal(m.Payload, &p) != nil {
			return protocolError(m, "RUNNER_INVALID_PAYLOAD", "payload")
		}
		return r.sandbox.Resize(ctx, handle, application.TerminalSize{Rows: p.Rows, Cols: p.Columns})
	case "stop":
		var p protocol.StopPayload
		if json.Unmarshal(m.Payload, &p) != nil {
			return protocolError(m, "RUNNER_INVALID_PAYLOAD", "payload")
		}
		if !started { // Nothing was launched, so stopping is the no-op it is in the drivers.
			return nil
		}
		return r.sandbox.Stop(ctx, handle, application.StopPolicy{GracePeriod: time.Duration(p.GraceMilliseconds) * time.Millisecond})
	case "inspect":
		_, err := r.Inspect(ctx, m.ExecutionID)
		return err
	case "cleanup":
		return r.Cleanup(ctx, m.ExecutionID)
	}
	return nil
}

func (r *Runner) prepare(ctx context.Context, m protocol.RunnerMessage) error {
	var p protocol.PreparePayload
	if json.Unmarshal(m.Payload, &p) != nil {
		return protocolError(m, "RUNNER_INVALID_PAYLOAD", "payload")
	}
	runID, err := domain.ParseRunID(p.WorkspaceID)
	if err != nil {
		return protocolError(m, "RUNNER_INVALID_PAYLOAD", "workspaceId")
	}
	prepared, err := r.sandbox.Prepare(ctx, application.SandboxSpec{ID: m.ExecutionID})
	if err != nil {
		return err
	}
	codec, err := r.harness.NewCodec(ctx, application.RunSpec{RunID: runID})
	if err != nil {
		_ = r.sandbox.Cleanup(ctx, prepared)
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if old := r.executions[m.ExecutionID]; old != nil {
		_ = r.sandbox.Cleanup(ctx, prepared)
		return protocolError(m, "RUNNER_EXECUTION_EXISTS", "executionId")
	}
	r.executions[m.ExecutionID] = &execution{lease: m.LeaseToken, prepared: prepared, codec: codec, runID: runID}
	return nil
}

func (r *Runner) start(ctx context.Context, m protocol.RunnerMessage, e *execution) error {
	var p protocol.StartPayload
	if json.Unmarshal(m.Payload, &p) != nil {
		return protocolError(m, "RUNNER_INVALID_PAYLOAD", "payload")
	}
	e.mu.Lock()
	if e.started || e.starting {
		e.mu.Unlock()
		return protocolError(m, "RUNNER_ALREADY_STARTED", "executionId")
	}
	e.starting = true
	prepared, runID := e.prepared, e.runID
	e.mu.Unlock()

	handle, stream, err := r.launch(ctx, prepared, runID)
	if err != nil {
		e.mu.Lock()
		e.starting = false
		e.mu.Unlock()
		return err
	}
	e.mu.Lock()
	e.handle, e.stream, e.started, e.starting = handle, stream, true, false
	e.mu.Unlock()
	if !r.startPump(m.ExecutionID, e) {
		return protocolError(m, "RUNNER_CLOSED", "executionId")
	}
	return nil
}

func (r *Runner) launch(ctx context.Context, prepared application.PreparedSandbox, runID domain.RunID) (application.ExecutionHandle, io.ReadWriteCloser, error) {
	launch, err := r.harness.BuildLaunch(ctx, application.RunSpec{RunID: runID}, nil)
	if err != nil {
		return application.ExecutionHandle{}, nil, err
	}
	handle, err := r.sandbox.Start(ctx, prepared, launch)
	if err != nil {
		return application.ExecutionHandle{}, nil, err
	}
	stream, err := r.sandbox.Attach(ctx, handle)
	if err != nil {
		return application.ExecutionHandle{}, nil, err
	}
	return handle, stream, nil
}

func (r *Runner) input(m protocol.RunnerMessage, e *execution, stream io.ReadWriteCloser) error {
	var p protocol.InputPayload
	if json.Unmarshal(m.Payload, &p) != nil {
		return protocolError(m, "RUNNER_INVALID_PAYLOAD", "payload")
	}
	e.codecMu.Lock()
	frames, err := e.codec.Encode(application.InputMessage{ID: p.InputID, Text: p.Text})
	e.codecMu.Unlock()
	if err != nil {
		return err
	}
	// writeMu orders concurrent input without e.mu. A write that blocks on a
	// full terminal buffer must not stop the pump from draining the other side.
	e.writeMu.Lock()
	defer e.writeMu.Unlock()
	for _, frame := range frames {
		if _, err = stream.Write(frame.Data); err != nil {
			// The cause names the device path, which must not reach a log.
			return fmt.Errorf("local runner: code=INPUT_WRITE_FAILED execution_id=%s", m.ExecutionID)
		}
	}
	return nil
}

func (r *Runner) pump(id string, e *execution) {
	defer r.pumps.Done()
	e.mu.Lock()
	stream, handle := e.stream, e.handle
	e.mu.Unlock()
	r.emit(Observation{ExecutionID: id, Ordinal: e.ordinal.Add(1), Type: "start"})

	buf := make([]byte, 32*1024)
	for {
		n, err := stream.Read(buf)
		if n > 0 {
			e.codecMu.Lock()
			events, codecErr := e.codec.Consume(application.OutputChunk{Data: append([]byte(nil), buf[:n]...)})
			e.codecMu.Unlock()
			if codecErr == nil {
				r.emitEvents(id, e, events)
			}
		}
		if err != nil {
			break
		}
	}
	// Flush precedes the exit observation so a partial record the harness wrote
	// on its way out is reported rather than dropped.
	e.codecMu.Lock()
	events, _ := e.codec.Flush()
	e.codecMu.Unlock()
	r.emitEvents(id, e, events)

	exit := r.harness.MapExit(r.awaitExitCode(handle))
	r.emit(Observation{ExecutionID: id, Ordinal: e.ordinal.Add(1), Type: "exit", Exit: &exit})
}

func (r *Runner) emitEvents(id string, e *execution, events []application.UnsequencedEvent) {
	for i := range events {
		event := events[i]
		r.emit(Observation{ExecutionID: id, Ordinal: e.ordinal.Add(1), Type: "output", Event: &event})
	}
}

// awaitExitCode polls until the sandbox has reaped the process. Inspect failing
// means the execution is already gone, which carries no code either.
func (r *Runner) awaitExitCode(handle application.ExecutionHandle) int {
	deadline := time.Now().Add(exitWaitTimeout)
	for {
		observed, err := r.sandbox.Inspect(context.Background(), handle)
		if err != nil {
			return unknownExitCode
		}
		if observed.ExitCode != nil {
			return *observed.ExitCode
		}
		if !time.Now().Before(deadline) {
			return unknownExitCode
		}
		select {
		case <-time.After(exitWaitInterval):
		case <-r.closed:
			return unknownExitCode
		}
	}
}

func (r *Runner) Inspect(ctx context.Context, id string) (application.RunnerExecutionObservation, error) {
	r.mu.Lock()
	e := r.executions[id]
	r.mu.Unlock()
	if e == nil {
		return application.RunnerExecutionObservation{}, errors.New("local runner: code=EXECUTION_NOT_FOUND")
	}
	e.mu.Lock()
	lease, started, handle := e.lease, e.started, e.handle
	e.mu.Unlock()
	state := application.ExecutionPreparing
	if started {
		state = application.ExecutionRunning
		observed, err := r.sandbox.Inspect(ctx, handle)
		if err != nil {
			return application.RunnerExecutionObservation{}, err
		}
		if !observed.Running {
			state = application.ExecutionExited
		}
	}
	return application.RunnerExecutionObservation{ExecutionID: id, OwnershipToken: lease, State: state}, nil
}

func (r *Runner) Takeover(_ context.Context, id, oldToken, newToken string) error {
	r.mu.Lock()
	e := r.executions[id]
	r.mu.Unlock()
	if e == nil {
		return errors.New("local runner: code=EXECUTION_NOT_FOUND")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.lease != oldToken {
		return errors.New("local runner: code=STALE_LEASE")
	}
	e.lease = newToken
	return nil
}

// Cleanup releases the execution and is safe to repeat; the second call finds
// nothing to release. The sandbox call happens without e.mu, since stopping an
// execution blocks for the grace period.
func (r *Runner) Cleanup(ctx context.Context, id string) error {
	r.mu.Lock()
	e := r.executions[id]
	delete(r.executions, id)
	r.mu.Unlock()
	if e == nil {
		return nil
	}
	e.mu.Lock()
	stream, prepared := e.stream, e.prepared
	e.mu.Unlock()
	if stream != nil {
		_ = stream.Close()
	}
	return r.sandbox.Cleanup(ctx, prepared)
}

var _ application.RunnerGateway = (*Runner)(nil)
var _ application.ReconciliationRunner = (*Runner)(nil)
