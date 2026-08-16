// Package local pumps a harness codec over an attached sandbox execution.
package local

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
	"github.com/shaielc/code-winch/pkg/protocol"
)

// Observation is runner-local output. Ordinal orders observations but is not a
// canonical event sequence; persistence remains the application's concern.
type Observation struct {
	ExecutionID string
	Ordinal     uint64
	Type        string
	Event       *application.UnsequencedEvent
	Exit        *application.HarnessExit
}

type execution struct {
	mu       sync.Mutex
	lease    string
	prepared application.PreparedSandbox
	handle   application.ExecutionHandle
	codec    application.HarnessCodec
	runID    domain.RunID
	stream   io.ReadWriteCloser
	started  bool
	ordinal  uint64
}

type Runner struct {
	sandbox      application.SandboxDriver
	harness      application.HarnessDriver
	observations chan Observation
	mu           sync.Mutex
	executions   map[string]*execution
}

func New(sandbox application.SandboxDriver, harness application.HarnessDriver) *Runner {
	return &Runner{sandbox: sandbox, harness: harness, observations: make(chan Observation, 64), executions: map[string]*execution{}}
}
func (r *Runner) Observations() <-chan Observation { return r.observations }

func protocolError(m protocol.RunnerMessage, code, field string) error {
	return &protocol.RunnerProtocolError{Code: code, CommandID: m.CommandID, ExecutionID: m.ExecutionID, Field: field}
}

func (r *Runner) Send(ctx context.Context, m protocol.RunnerMessage) error {
	known := map[string]bool{"prepare": true, "start": true, "input": true, "resize": true, "stop": true, "inspect": true, "cleanup": true}
	if !known[m.Kind] {
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
	defer e.mu.Unlock()
	if e.lease != m.LeaseToken {
		return protocolError(m, "RUNNER_STALE_LEASE", "leaseToken")
	}
	switch m.Kind {
	case "start":
		return r.startLocked(ctx, m, e)
	case "input":
		if !e.started || e.stream == nil {
			return protocolError(m, "RUNNER_NOT_READY", "executionId")
		}
		var p protocol.InputPayload
		if json.Unmarshal(m.Payload, &p) != nil {
			return protocolError(m, "RUNNER_INVALID_PAYLOAD", "payload")
		}
		frames, err := e.codec.Encode(application.InputMessage{ID: p.InputID, Text: p.Text})
		if err != nil {
			return err
		}
		for _, frame := range frames {
			if _, err = e.stream.Write(frame.Data); err != nil {
				return fmt.Errorf("local runner: code=INPUT_WRITE_FAILED execution_id=%s: %w", m.ExecutionID, err)
			}
		}
		return nil
	case "resize":
		var p protocol.ResizePayload
		if json.Unmarshal(m.Payload, &p) != nil {
			return protocolError(m, "RUNNER_INVALID_PAYLOAD", "payload")
		}
		return r.sandbox.Resize(ctx, e.handle, application.TerminalSize{Rows: p.Rows, Cols: p.Columns})
	case "stop":
		var p protocol.StopPayload
		if json.Unmarshal(m.Payload, &p) != nil {
			return protocolError(m, "RUNNER_INVALID_PAYLOAD", "payload")
		}
		return r.sandbox.Stop(ctx, e.handle, application.StopPolicy{GracePeriod: time.Duration(p.GraceMilliseconds) * time.Millisecond})
	case "inspect":
		_, err := r.sandbox.Inspect(ctx, e.handle)
		return err
	case "cleanup":
		return r.cleanupLocked(ctx, m.ExecutionID, e)
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

func (r *Runner) startLocked(ctx context.Context, m protocol.RunnerMessage, e *execution) error {
	if e.started {
		return protocolError(m, "RUNNER_ALREADY_STARTED", "executionId")
	}
	var p protocol.StartPayload
	if json.Unmarshal(m.Payload, &p) != nil {
		return protocolError(m, "RUNNER_INVALID_PAYLOAD", "payload")
	}
	launch, err := r.harness.BuildLaunch(ctx, application.RunSpec{RunID: e.runID}, nil)
	if err != nil {
		return err
	}
	e.handle, err = r.sandbox.Start(ctx, e.prepared, launch)
	if err != nil {
		return err
	}
	e.stream, err = r.sandbox.Attach(ctx, e.handle)
	if err != nil {
		return err
	}
	e.started = true
	e.ordinal++
	r.observations <- Observation{ExecutionID: m.ExecutionID, Ordinal: e.ordinal, Type: "start"}
	go r.pump(m.ExecutionID, e)
	return nil
}

func (r *Runner) pump(id string, e *execution) {
	buf := make([]byte, 32*1024)
	for {
		n, err := e.stream.Read(buf)
		if n > 0 {
			events, codecErr := e.codec.Consume(application.OutputChunk{Data: append([]byte(nil), buf[:n]...)})
			if codecErr == nil {
				for i := range events {
					e.mu.Lock()
					e.ordinal++
					o := e.ordinal
					e.mu.Unlock()
					event := events[i]
					r.observations <- Observation{ExecutionID: id, Ordinal: o, Type: "output", Event: &event}
				}
			}
		}
		if err != nil {
			break
		}
	}
	events, _ := e.codec.Flush()
	for i := range events {
		e.mu.Lock()
		e.ordinal++
		o := e.ordinal
		e.mu.Unlock()
		event := events[i]
		r.observations <- Observation{ExecutionID: id, Ordinal: o, Type: "output", Event: &event}
	}
	observed, err := r.sandbox.Inspect(context.Background(), e.handle)
	code := -1
	if err == nil && observed.ExitCode != nil {
		code = *observed.ExitCode
	}
	mapped := r.harness.MapExit(code)
	e.mu.Lock()
	e.ordinal++
	o := e.ordinal
	e.mu.Unlock()
	r.observations <- Observation{ExecutionID: id, Ordinal: o, Type: "exit", Exit: &mapped}
}

func (r *Runner) cleanupLocked(ctx context.Context, id string, e *execution) error {
	if e.stream != nil {
		_ = e.stream.Close()
	}
	err := r.sandbox.Cleanup(ctx, e.prepared)
	r.mu.Lock()
	delete(r.executions, id)
	r.mu.Unlock()
	return err
}

func (r *Runner) Inspect(ctx context.Context, id string) (application.RunnerExecutionObservation, error) {
	r.mu.Lock()
	e := r.executions[id]
	r.mu.Unlock()
	if e == nil {
		return application.RunnerExecutionObservation{}, errors.New("local runner: code=EXECUTION_NOT_FOUND")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	state := application.ExecutionPreparing
	if e.started {
		state = application.ExecutionRunning
		o, err := r.sandbox.Inspect(ctx, e.handle)
		if err != nil {
			return application.RunnerExecutionObservation{}, err
		}
		if !o.Running {
			state = application.ExecutionExited
		}
	}
	return application.RunnerExecutionObservation{ExecutionID: id, OwnershipToken: e.lease, State: state}, nil
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
func (r *Runner) Cleanup(ctx context.Context, id string) error {
	r.mu.Lock()
	e := r.executions[id]
	r.mu.Unlock()
	if e == nil {
		return nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	return r.cleanupLocked(ctx, id, e)
}

var _ application.RunnerGateway = (*Runner)(nil)
var _ application.ReconciliationRunner = (*Runner)(nil)
