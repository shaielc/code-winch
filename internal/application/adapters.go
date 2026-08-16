package application

import (
	"context"
	"io"
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

// HarnessExit is the provider-neutral interpretation of a harness process
// exit. Message is a stable, content-free operator diagnostic.
type HarnessExit struct {
	Successful bool
	Code       string
	Message    string
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
	MapExit(int) HarnessExit
}

type HarnessCodec interface {
	Consume(OutputChunk) ([]UnsequencedEvent, error)
	Encode(InputMessage) ([]InputFrame, error)
	Flush() ([]UnsequencedEvent, error)
}

type SandboxCapabilities struct {
	// Isolation is a stable, user-facing description of the security boundary.
	// Drivers which run directly on the host must report "unisolated".
	Isolation      string
	Resize         bool
	NetworkPolicy  bool
	ResourceLimits bool
	// Attach provides a bidirectional byte stream for the execution. Attach is
	// single-use when AttachSingleUse is true.
	Attach          bool
	AttachSingleUse bool
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
type TerminalSize struct {
	Rows uint16
	Cols uint16
}

type SandboxDriver interface {
	Capabilities(context.Context) SandboxCapabilities
	Prepare(context.Context, SandboxSpec) (PreparedSandbox, error)
	Start(context.Context, PreparedSandbox, LaunchSpec) (ExecutionHandle, error)
	Attach(context.Context, ExecutionHandle) (io.ReadWriteCloser, error)
	Inspect(context.Context, ExecutionHandle) (ObservedExecution, error)
	Resize(context.Context, ExecutionHandle, TerminalSize) error
	Stop(context.Context, ExecutionHandle, StopPolicy) error
	Cleanup(context.Context, PreparedSandbox) error
}
