package execution

import (
	"context"

	"github.com/shaielc/code-winch/internal/application"
	"github.com/shaielc/code-winch/internal/domain"
)

// Capabilities reports which input kinds a run can accept right now. The set is
// deliberately narrow: it lists only the kinds the engine can actually deliver,
// so a command is refused at acceptance rather than accepted, recorded, and
// then dropped at the runner.
//
// Absent kinds and their owners:
//
//	interrupt      no runner message carries it yet — P1-052
//	terminal_bytes raw terminal input needs an authorizer — P2-026
type Capabilities struct{ Runs application.RunRepository }

func (c Capabilities) InputCapabilities(ctx context.Context, id domain.RunID) (application.InputCapabilities, error) {
	record, _, err := c.Runs.Get(ctx, id)
	if err != nil {
		return application.InputCapabilities{}, err
	}
	state := domain.RunStateCreated
	if len(record.Attempts) > 0 {
		state = record.Attempts[len(record.Attempts)-1].State
	}
	modes := map[application.InputKind]bool{}
	if state == domain.RunStateRunning {
		modes[application.InputText] = true
		modes[application.InputResize] = true
	}
	return application.InputCapabilities{State: state, Modes: modes}, nil
}

var _ application.InputCapabilityProvider = Capabilities{}
