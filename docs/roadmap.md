# Delivery roadmap

## Principles

Build thin, observable vertical slices. A phase is complete only when its
contracts and failure behavior are tested; merely presenting a happy-path UI is
not sufficient. Avoid implementing remote distribution before the in-process
runner boundary is proven serializable.

## Phase 0: contracts and development foundation

- Establish Go and web workspaces, formatting, linting, tests, migrations, and
  generated OpenAPI client checks.
- Implement domain IDs, run state machine, event envelope, and in-memory ports.
- Add a deterministic fake harness and contract-test kits.
- Record threat model and data-retention defaults.

**Exit:** state/event compatibility fixtures exist and CI can exercise a run
without a real provider.

## Phase 1: local single-user vertical slice

- Local PTY sandbox driver with honest `unisolated` capability.
- One harness adapter, start/input/resize/stop, and restart reconciliation.
- PostgreSQL run/event storage and transactional outbox.
- HTTP snapshot API, resumable WebSocket stream, and terminal renderer.
- Minimal local authentication and credential-reference storage.

**Exit:** a browser can reconnect to a live run without losing ordered history;
daemon restart produces a truthful terminal state; forced stop leaves no child
processes.

## Phase 2: structured experience and second harness

- Canonical messages, tools, approvals, artifacts, file-change, and usage
  events.
- Conversation/activity/diff renderers with safe fallbacks.
- Second provider adapter proving the harness port and capability UI.
- User/workspace authorization, audit trail, retention, and export/delete.

**Exit:** the same web views consume two providers without provider conditionals
in generic components, while provider extensions remain available.

## Phase 3: Docker isolation

- Docker driver, image policy, disposable worktrees, resource limits, stop and
  orphan reconciliation.
- Named sandbox profiles, network deny/allowlist controls, and scoped credential
  injection.
- Adversarial integration tests and visible security posture in the UI.

**Exit:** contract tests pass for local and Docker drivers; escape-prone options
cannot be requested through per-run overrides; cleanup survives daemon failure.

## Phase 4: top-level workflows

- Versioned declarative definitions and database-backed workflow runtime.
- Run/send/wait/approval/condition/parallel steps, retries, timeouts, and
  idempotency.
- Workflow graph/status UI and run lineage.

**Exit:** coordinator restart/replay does not duplicate a run or input; a user
can inspect and cancel every active branch.

## Phase 5: remote runners and hardening

- Versioned authenticated runner protocol, registration, capabilities,
  heartbeats, leases, backpressure, and artifact handoff.
- Scheduling by sandbox/harness capability and resource capacity.
- Stronger isolation driver evaluation, HA API/workers, backup/restore, SLOs,
  and operational runbooks.

**Exit:** loss or partition of a runner cannot cause two owners to control one
execution, stale events cannot enter canonical history, and capacity/failure is
observable.

## Deferred decisions and triggers

| Decision | Defer until | Trigger to revisit |
|---|---|---|
| Dedicated workflow engine | Database coordinator shows limits | workflow volume, long timers, operational retry burden |
| Broker for live events | PostgreSQL outbox/polling is measured | fan-out/latency load exceeds targets |
| Kubernetes deployment | remote runner scheduling is stable | multi-host operations require it |
| Third-party adapter plugins | two built-in adapters validate contract | external ecosystem demand |
| MicroVM sandbox | container risk is quantified | hostile multi-tenant workloads |
| CRDT/shared terminal input | single-controller semantics proven | true simultaneous collaboration demand |
| Out-of-process renderer isolation | a renderer runs server-side or is not built in | any renderer that is experimental, third-party, or executes outside the browser |
| Interactive harness login proxying | non-interactive login patterns cover the shipped adapters | an adapter offers no OAuth, device flow, or token entry |
| Renderer output caching | projection cost is measured against real event volumes | renderer latency or CPU becomes a visible cost |
| Communication/API proxy | egress policy by DNS/IP/SNI is in place and measured | credential non-disclosure, billing, or redaction requires a local endpoint |
| SQLite single-process developer profile | the PostgreSQL compose stack is a friction point | contributors cannot or will not run a database |

Each row is a decision no task owns by design. A brief may cite one instead of
naming an owning task ID; anything not in this table needs an owner.

## Initial acceptance metrics

- no missing/duplicated persisted event sequences under reconnect/restart tests;
- p95 accepted-event-to-browser latency below 500 ms on a local deployment;
- stop completes within configured escalation deadline and leaves zero owned
  resources in integration tests;
- secrets have zero occurrences in persisted events and default logs in canary
  tests; and
- adapter and sandbox contract suites are mandatory for every implementation.
