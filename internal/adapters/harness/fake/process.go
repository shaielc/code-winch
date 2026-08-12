package fake

import (
	"time"

	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
)

// Step is emitted once its offset has elapsed. ExitCode ends the process;
// Data may be chunked in any way a test needs, including malformed records.
type Step struct {
	After    time.Duration
	Data     []byte
	ExitCode *int
}

type Process struct {
	clock application.Clock
	start domain.Timestamp
	steps []Step
	next  int
}

func NewProcess(clock application.Clock, steps []Step) *Process {
	copySteps := append([]Step(nil), steps...)
	return &Process{clock: clock, start: clock.Now(), steps: copySteps}
}

// Poll returns all newly due output and an exit code, if one became due. It
// never sleeps: tests advance their injected clock and poll again.
func (p *Process) Poll() ([][]byte, *int) {
	var output [][]byte
	elapsed := p.clock.Now().Time().Sub(p.start.Time())
	for p.next < len(p.steps) && p.steps[p.next].After <= elapsed {
		step := p.steps[p.next]
		p.next++
		if step.Data != nil {
			output = append(output, append([]byte(nil), step.Data...))
		}
		if step.ExitCode != nil {
			code := *step.ExitCode
			return output, &code
		}
	}
	return output, nil
}
