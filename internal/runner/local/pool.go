package local

import (
	"context"
	"sync"

	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/pkg/protocol"
)

// Pool is the runner gateway a composition root registers. A Runner is bound to
// one harness and one sandbox driver, while a run chooses its pair per request,
// so the pool keeps one Runner per pair and routes every later message by
// execution ID. Adding a driver is therefore still only a registration file:
// the pool reads the registry and never names an adapter.
type Pool struct {
	drivers      *application.DriverRegistry
	observations chan Observation

	mu      sync.Mutex
	runners map[string]*Runner
	routes  map[string]*Runner
	closed  bool

	forwarders sync.WaitGroup
	closeOnce  sync.Once
}

func NewPool(drivers *application.DriverRegistry) *Pool {
	return &Pool{
		drivers:      drivers,
		observations: make(chan Observation, 64),
		runners:      map[string]*Runner{},
		routes:       map[string]*Runner{},
	}
}

// Observations merges every runner's stream. It closes after Close, so a
// consumer ranging over it terminates.
func (p *Pool) Observations() <-chan Observation { return p.observations }

// Bind resolves the profiles and attaches an execution ID to their runner. It
// must precede the prepare message, which is the first message carrying that
// execution ID.
func (p *Pool) Bind(executionID, harnessProfile, sandboxProfile string) error {
	harness, err := p.drivers.Harness(harnessProfile)
	if err != nil {
		return err
	}
	sandbox, err := p.drivers.Sandbox(sandboxProfile)
	if err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return &protocol.RunnerProtocolError{Code: "RUNNER_CLOSED", ExecutionID: executionID, Field: "executionId"}
	}
	key := harnessProfile + "\x00" + sandboxProfile
	runner := p.runners[key]
	if runner == nil {
		runner = New(sandbox, harness)
		p.runners[key] = runner
		p.forwarders.Add(1)
		go p.forward(runner)
	}
	p.routes[executionID] = runner
	return nil
}

func (p *Pool) forward(runner *Runner) {
	defer p.forwarders.Done()
	for observation := range runner.Observations() {
		p.observations <- observation
	}
}

func (p *Pool) route(executionID string) *Runner {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.routes[executionID]
}

func (p *Pool) Send(ctx context.Context, m protocol.RunnerMessage) error {
	runner := p.route(m.ExecutionID)
	if runner == nil {
		return &protocol.RunnerProtocolError{Code: "RUNNER_EXECUTION_NOT_FOUND", CommandID: m.CommandID, ExecutionID: m.ExecutionID, Field: "executionId"}
	}
	return runner.Send(ctx, m)
}

func (p *Pool) Inspect(ctx context.Context, executionID string) (application.RunnerExecutionObservation, error) {
	runner := p.route(executionID)
	if runner == nil {
		// A restarted daemon holds no local executions, which the supervisor
		// reads as a lost execution rather than a healthy one.
		return application.RunnerExecutionObservation{ExecutionID: executionID, State: application.ExecutionUnknown}, nil
	}
	return runner.Inspect(ctx, executionID)
}

func (p *Pool) Takeover(ctx context.Context, executionID, oldToken, newToken string) error {
	runner := p.route(executionID)
	if runner == nil {
		return &protocol.RunnerProtocolError{Code: "RUNNER_EXECUTION_NOT_FOUND", ExecutionID: executionID, Field: "executionId"}
	}
	return runner.Takeover(ctx, executionID, oldToken, newToken)
}

// Cleanup releases the execution and its route, and like Runner.Cleanup it is
// safe to repeat.
func (p *Pool) Cleanup(ctx context.Context, executionID string) error {
	p.mu.Lock()
	runner := p.routes[executionID]
	delete(p.routes, executionID)
	p.mu.Unlock()
	if runner == nil {
		return nil
	}
	return runner.Cleanup(ctx, executionID)
}

// Close stops every runner and closes the merged stream once their pumps have
// finished. The observation consumer must keep reading until then, since the
// pumps drain through it.
func (p *Pool) Close() {
	p.closeOnce.Do(func() {
		p.mu.Lock()
		p.closed = true
		runners := make([]*Runner, 0, len(p.runners))
		for _, runner := range p.runners {
			runners = append(runners, runner)
		}
		p.mu.Unlock()
		for _, runner := range runners {
			runner.Close()
		}
		p.forwarders.Wait()
		close(p.observations)
	})
}

var _ application.RunnerGateway = (*Pool)(nil)
var _ application.ReconciliationRunner = (*Pool)(nil)
