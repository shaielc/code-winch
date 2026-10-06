# Repository and package structure

## 1. Proposed monorepo

```text
code-winch/
├── cmd/
│   ├── winch-sandbox/           # sandbox-resident runner + attach surface;
│   │                            # the first composition root (ADR-0005)
│   ├── winchd/                  # control-plane composition root; arrives with
│   │                            # the stage that observes sandboxes
│   └── fake-harness/            # controllable stand-in for a vendor CLI
├── internal/
│   ├── domain/                  # dependency-free entities, value types, state machines
│   ├── application/             # use cases and ports
│   ├── supervisor/              # per-run serialization, leases, reconciliation
│   ├── workflow/                # definitions, coordinator, runtime port
│   ├── runner/                  # harness process ownership, codec pumping,
│   │                            # local ordinals, session record store
│   ├── adapters/
│   │   ├── harness/             # one package per coding-agent integration
│   │   ├── sandbox/             # local, docker, future backends
│   │   ├── persistence/         # PostgreSQL repositories and outbox
│   │   ├── secrets/             # OS keychain/file/vault adapters
│   │   └── transport/           # HTTP, WebSocket, runner RPC
│   └── platform/                # config, telemetry, clock, IDs
├── pkg/protocol/                # versioned runner/event wire schemas only
├── web/
│   ├── src/attach/              # the sandbox's one-session surface
│   ├── src/app/                 # routes and application shell
│   ├── src/features/            # run/workflow/auth vertical UI slices
│   ├── src/renderers/            # terminal, conversation, tool, diff projections
│   │                            # shared by both surfaces
│   └── src/api/                  # generated client and stream reconnection
├── api/openapi/                 # public API source of truth
├── schemas/                     # event and runner protocol schemas
├── migrations/                  # ordered database migrations
├── deployments/                 # local compose and production examples
├── docs/                        # architecture and operational design
└── test/
    ├── contract/                # adapter/protocol compatibility suites
    ├── integration/             # database, Docker, PTY tests
    └── e2e/                     # browser-to-fake-harness scenarios
```

Directories should be created when their first implementation is added; this
document is not a request for empty scaffolding. The order they appear in is
not the order they are built: `cmd/winch-sandbox`, `internal/runner`,
`internal/adapters/harness`, `web/src/attach`, and `deployments/` come first,
and `internal/domain`, `internal/supervisor`, and `internal/workflow` arrive with
the control plane. [`docs/roadmap.md`](roadmap.md) is the order.

Whether `web/src/attach` and `web/src/app` stay one workspace with two entry
points or become two workspaces sharing a renderer package is roadmap decision
D2, open until the control plane's surface needs its first shared component.

## 2. Dependency rule

```mermaid
flowchart LR
  T[Transport adapters] --> A[Application]
  P[Persistence adapters] --> A
  H[Harness adapters] --> A
  S[Sandbox adapters] --> A
  A --> D[Domain]
  W[Workflow] --> A
```

Dependencies point inward. Domain code imports no adapters, database packages,
web frameworks, Docker clients, or provider SDKs. Application packages define
ports; outer adapters implement them; a `cmd` composition root wires concrete
implementations. Cross-adapter imports are prohibited.

There are two composition roots and they wire different subsets. The sandbox
root wires a harness adapter, the runner, its session store, and the attach
transport — and nothing from `internal/domain`, `internal/supervisor`, or
`internal/workflow`, because a sandbox has a session and not a run. The
control-plane root wires everything else. A package the sandbox root needs
therefore may not depend on the run aggregate, and that constraint is what keeps
a standalone sandbox startable.

## 3. Principal ports

The names below describe responsibilities rather than freezing Go signatures.

```go
type HarnessDriver interface {
    Describe(ctx context.Context) (HarnessDescriptor, error)
    BuildLaunch(ctx context.Context, RunSpec, ResolvedCredentials) (LaunchSpec, error)
    NewCodec(ctx context.Context, RunSpec) (HarnessCodec, error)
}

type HarnessCodec interface {
    Consume(OutputChunk) ([]UnsequencedEvent, error)
    Encode(InputMessage) ([]InputFrame, error)
    Flush() ([]UnsequencedEvent, error)
}

type SandboxDriver interface {
    Capabilities(ctx context.Context) SandboxCapabilities
    Prepare(ctx context.Context, SandboxSpec) (PreparedSandbox, error)
    Start(ctx context.Context, PreparedSandbox, LaunchSpec) (ExecutionHandle, error)
    Attach(ctx context.Context, ExecutionHandle) (io.ReadWriteCloser, error)
    Inspect(ctx context.Context, ExecutionHandle) (ObservedExecution, error)
    Stop(ctx context.Context, ExecutionHandle, StopPolicy) error
    Cleanup(ctx context.Context, PreparedSandbox) error
}

type EventStore interface {
    Append(ctx context.Context, RunID, ExpectedSequence, []UnsequencedEvent) ([]RunEvent, error)
    Read(ctx context.Context, RunID, AfterSequence, Limit) ([]RunEvent, error)
}

type WorkflowRuntime interface {
    Start(ctx context.Context, WorkflowInstance) error
    Signal(ctx context.Context, WorkflowID, WorkflowSignal) error
    ClaimReadySteps(ctx context.Context, WorkerID, Limit) ([]StepLease, error)
}
```

Sandbox capabilities explicitly report whether attached I/O is available,
whether attachment is single-use, and whether the sandbox hosts the runner and
serves an attach surface. The runner is the sole owner of opaque execution
handles and harness codecs; it pumps attached bytes into codecs, emits
runner-local ordinals, never canonical event sequence numbers, and persists
those records to the session store it owns. On a hosting sandbox the runner lives
inside the sandbox, so "attached I/O" is a process-local pipe or PTY rather than
a protocol hop, and the protocol hop is between the control plane and the runner
instead.

`ResolvedCredentials` is short-lived and can be used only during sandbox
preparation/launch. It must never appear in `RunSpec`, persisted events, or
logs. An execution handle is opaque outside its sandbox adapter.

## 4. Adding a harness

Each harness package contains:

1. a descriptor with stable ID, adapter version, supported input/output modes,
   login modes, and capabilities;
2. configuration validation and launch-spec construction;
3. an incremental codec tolerant of arbitrary byte chunk boundaries;
4. mappings from native messages/exits to canonical events;
5. namespaced extension schemas for provider-only data; and
6. a contract-test fixture using a fake CLI or recorded, sanitized transcript.

Harness registration is explicit in the composition root. Loading arbitrary
in-process plugins is deferred because it weakens supply-chain and isolation
controls. Later third-party integrations should use a versioned subprocess
plugin protocol.

## 5. Adding a sandbox

A sandbox package owns resource naming, preparation, start/attach, inspection,
stop escalation, and cleanup. It must pass the shared sandbox contract suite:

- workspace visibility and write policy;
- environment/secret injection behavior;
- stdout/stderr ordering guarantees as declared by capability;
- terminal resizing if advertised;
- stop escalation and orphan cleanup;
- resource limits and network policy enforcement;
- idempotent inspect/stop/cleanup operations; and
- the attach surface's contract in `docs/contracts.md` §8, if the driver
  advertises hosting the runner.

The `local` adapter must label unsupported controls honestly; it must not claim
filesystem or network isolation.

## 6. Configuration layers

Configuration resolves in this order: compiled safe defaults, system config,
workspace policy, named harness/sandbox profile, and allowed per-run overrides.
Policy decides which fields are overridable. The fully resolved non-secret
configuration is stored with a run for reproducibility. Secrets remain opaque
references and are resolved only at launch.

## 7. Testing boundaries

- **Domain tests:** table/property tests for state machines and invariants.
- **Adapter contract tests:** the same suite against each harness and sandbox.
- **Protocol compatibility tests:** old fixtures decode; new optional fields are
  ignored; supported version negotiation is verified.
- **Integration tests:** real PTY, PostgreSQL, and opt-in Docker execution.
- **End-to-end tests:** browser + sandbox + deterministic fake harness, and —
  once the control plane exists — browser + daemon + sandbox. The first form is
  the standing suite the later stages re-run against more real substrates.
- **Security tests:** traversal, secret redaction, authorization, escape-prone
  sandbox configurations, and malicious renderer payloads.

Fake time, ID generation, runner, secret store, and event publisher are injected
through ports so tests do not rely on sleeps or actual provider accounts.
