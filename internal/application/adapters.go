package application

import (
	"context"
	"time"

	"github.com/shaielc/code-winch/internal/domain"
)

// HarnessDescriptor is the provider-neutral description used during adapter
// selection. Capabilities are explicit; callers must not infer them from an ID.
type HarnessDescriptor struct {
	ID           string
	Version      string
	InputModes   []string
	OutputModes  []string
	Capabilities map[string]bool
}

type RunSpec struct{ RunID domain.RunID }
type ResolvedCredentials map[string]ResolvedSecret
type LaunchSpec struct {
	Command string
	Args    []string
	Env     map[string]string
}
type OutputChunk struct{ Data []byte }
type InputMessage struct {
	ID   string
	Text string
}
type InputFrame struct{ Data []byte }

type HarnessDriver interface {
	Describe(context.Context) (HarnessDescriptor, error)
	BuildLaunch(context.Context, RunSpec, ResolvedCredentials) (LaunchSpec, error)
	NewCodec(context.Context, RunSpec) (HarnessCodec, error)
}

type HarnessCodec interface {
	Consume(OutputChunk) ([]UnsequencedEvent, error)
	Encode(InputMessage) ([]InputFrame, error)
	Flush() ([]UnsequencedEvent, error)
}

type SandboxCapabilities struct {
	Resize         bool
	NetworkPolicy  bool
	ResourceLimits bool
}
type SandboxSpec struct {
	ID             string
	DisableNetwork bool
	MemoryBytes    uint64
}
type PreparedSandbox struct{ ID string }
type ExecutionHandle struct{ ID string }
type ObservedExecution struct {
	Running  bool
	ExitCode *int
}
type StopPolicy struct{ GracePeriod time.Duration }

type SandboxDriver interface {
	Capabilities(context.Context) SandboxCapabilities
	Prepare(context.Context, SandboxSpec) (PreparedSandbox, error)
	Start(context.Context, PreparedSandbox, LaunchSpec) (ExecutionHandle, error)
	Inspect(context.Context, ExecutionHandle) (ObservedExecution, error)
	Stop(context.Context, ExecutionHandle, StopPolicy) error
	Cleanup(context.Context, PreparedSandbox) error
}
