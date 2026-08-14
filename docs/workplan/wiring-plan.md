# Wiring plan: from components to a running daemon

Phase 1 has produced every component a minimal deployment needs, but nothing
that connects them. `cmd/winchd/main.go` is still `func main() {}`, and the two
seams that would join the pieces — the runner that owns process I/O, and the
run use cases behind the HTTP adapter — are owned by no task in
`docs/workplan/README.md`.

This document records what is missing and proposes the tasks that close it.
Goal: a single-user deployment that can **start a harness and interact with
it** through the browser, on the `local-trusted` profile.

## What exists

| Layer | Status |
|---|---|
| Domain IDs, run/attempt state machine, event envelope | complete (P0-003 – P0-005) |
| Runner protocol messages `pkg/protocol` | complete (P0-006) |
| Application ports, in-memory fakes | complete (P0-007) |
| Fake harness, harness/sandbox contract suites | complete (P0-008) |
| PostgreSQL repositories, migrations, outbox store | complete (P1-011, P1-012) |
| Local PTY sandbox driver | complete (P1-013) |
| Codex harness adapter (descriptor, launch, codec, exit map) | complete (P1-014) |
| Supervisor: lease acquire/renew/release, `Execute`, `Observe`, `Reconcile` | complete (P1-015) |
| Input acceptance and outbox intent | complete (P1-016) |
| HTTP API: 7 operations, auth, CSRF, limits, problems, ETags, audit | complete (P1-017) |
| WebSocket stream: resume, heartbeat, `caught_up`, slow-subscriber policy | complete (P1-018) |
| Terminal web slice | complete (P1-019) |

## What is missing

### A. The sandbox port cannot move bytes

`application.SandboxDriver` has `Prepare`, `Start`, `Inspect`, `Resize`,
`Stop`, and `Cleanup` — and no way to read output or write input. The local
driver opens a PTY (`local.go:106`) and keeps `*os.File` private; it is used
only for `Setsize` and `Close`. The shared sandbox contract suite asserts
nothing about I/O.

Consequence: no harness output can reach a codec, and no input frame can reach
a process. Both stated requirements are blocked here. This is a **port
contract change**, so `docs/code-structure.md` §3 must be updated in the same
change, per the workplan's contract rule.

### B. No runner implementation

`application.RunnerGateway` has exactly one implementation — the in-memory
recorder in `internal/adapters/memory`. Nothing translates a
`protocol.RunnerMessage` into sandbox driver calls, and nothing pumps output
through `HarnessCodec.Consume` into `Supervisor.Observe`. The runner described
in `docs/architecture.md` §3 as the owner of "processes, PTYs, containers,
workspace mounts, and signal delivery" does not exist as code.
`application.ReconciliationRunner` (needed by P1-020) likewise has no
implementation.

### C. No run use cases

`internal/application` has `InputService`, outbox, and workflow logic, but no
run orchestration. The supervisor exposes primitives (`Acquire`, `Execute`,
`Observe`), not `CreateRun`/`StartRun`/`StopRun`. There is no registry mapping
a harness or sandbox profile name to a driver, and no resolution of a profile
into a `SandboxSpec`.

### D. The HTTP adapter is unbound

`httpapi.Backend` has no non-test implementation; the adapter is tested against
a fake. Nothing maps application errors onto the five `httpapi.Err*` sentinels.

### E. Nothing consumes the outbox

P1-016 commits `run.input` intent; P1-012 delivers records to an
`OutboxPublisher`. No publisher routes `run.input` to the runner or `run.event`
to `httpapi.EventStream`, so accepted input is never delivered and persisted
events never reach a browser.

### F. No composition root or deployment profile

No config loading, no pool construction, no `MigrateUp` call at startup, no
server lifecycle, no static asset serving, no container image or compose file.
The web client calls `/api/v1` on its own origin (`web/src/api/runClient.ts`)
and the CSRF token is documented as arriving through "trusted local bootstrap
configuration" that does not exist yet.

## Proposed tasks

Five tasks, each independently reviewable. IDs continue the Phase 1 sequence;
add them to `tasks.json` and write briefs in `docs/workplan/phase-1/` before
starting, so the status gate can resolve them.

### P1-021 — Add sandbox I/O to the driver port

**Depends on:** P1-013 (and P0-008 for the suite)

Add a single attach operation to `SandboxDriver`, returning an
`io.ReadWriteCloser` bound to one execution handle — for the local driver, the
PTY. Implement it in the local and fake drivers, and extend the shared sandbox
contract suite so any driver that advertises the capability is checked for
read-after-write, close semantics, and behavior after the process exits.

- Update `docs/code-structure.md` §3 in the same change.
- Acceptance: contract suite covers attach for both drivers; a driver that
  cannot stream fails the suite rather than returning a nil stream; attach is
  idempotent or explicitly single-use, and closing it does not kill the process
  before `Stop` escalation runs.
- Verification: sandbox contract suite; PTY integration test that writes input
  and reads echoed output.

### P1-022 — Implement the in-process local runner

**Depends on:** P0-006, P1-013, P1-014, P1-021

New package `internal/runner/local`, implementing `application.RunnerGateway`
and `application.ReconciliationRunner`. It is the only component allowed to
hold execution handles and codecs.

- Handle `prepare`, `start`, `input`, `resize`, `stop`, `inspect`, `cleanup`
  from `protocol.RunnerMessage`; reject unknown kinds and stale lease tokens.
- On `start`: `HarnessDriver.BuildLaunch` → `SandboxDriver.Prepare`/`Start` →
  attach → goroutine pumping chunks into `HarnessCodec.Consume`, emitting
  observations with a monotonic runner-local ordinal.
- On `input`: `HarnessCodec.Encode` → write frames to the attached stream.
- On exit: `Flush`, then `HarnessDriver.MapExit` into a lifecycle observation.
- Acceptance: a run against the P0-008 fake harness produces ordered
  observations for start, output, and exit; input written before the process is
  ready fails with a stable error rather than being dropped; stop escalation
  leaves no child process; the runner never assigns canonical sequences.
- Verification: fake-harness end-to-end runner test; race-enabled pump test;
  orphan cleanup test.

This task also unblocks **P1-020**, which currently has no runner to inspect.

### P1-023 — Implement run use cases and the driver registry

**Depends on:** P0-004, P1-011, P1-015, P1-022

New `internal/application/run.go` (or `runs` package) providing `CreateRun`,
`GetRun`, `StartRun`, `StopRun`, and `ListRunEvents`, plus a registry resolving
harness and sandbox profile names to registered drivers and a profile-to-
`SandboxSpec` resolution that stores the effective non-secret configuration
with the run.

- `CreateRun` persists a `RunRecord` in `Created` through `RunRepository`,
  returning its version for ETag use.
- `StartRun` acquires a lease, resolves drivers, and issues `prepare`+`start`
  through the supervisor; desired state is persisted before runner interaction.
- `StopRun` applies the escalation policy and is idempotent in `Stopping` and
  terminal states.
- Unknown profile names fail before launch with a stable error.
- Acceptance: every documented transition is reachable through the service;
  concurrent start/stop on one run has one deterministic outcome; a run records
  its resolved profile and contains no secret values.
- Verification: service tests against in-memory ports; PostgreSQL integration
  test for create/start/stop; capability-combination table test.

### P1-024 — Bind the HTTP backend and outbox delivery

**Depends on:** P1-016, P1-017, P1-018, P1-023

Two thin adapters, deliberately kept free of business logic:

1. `httpapi.Backend` over the run services and `InputService`, mapping
   `ErrNotFound`/`ErrConflict`/`InputError` onto `ErrRunNotFound`,
   `ErrStateConflict`, `ErrIdempotencyConflict`, `ErrPreconditionFailed`, and
   `ErrValidation`.
2. An `OutboxPublisher` that routes `run.input` records to the runner through
   the supervisor and `run.event` records to `httpapi.EventStream.Publish`.

- Acceptance: every problem code in the OpenAPI contract is produced by a real
  application error path; input accepted before a crash is delivered after
  restart without duplication; published events reach a subscribed WebSocket in
  sequence order.
- Verification: HTTP integration tests against a real backend; crash/retry
  delivery test; end-to-end HTTP-to-stream ordering test.

### P1-025 — Daemon composition root and local deployment profile

**Depends on:** P1-019, P1-024

Fill in `cmd/winchd`, and make it deployable.

- Configuration from environment with safe defaults and startup validation:
  database DSN, API token, CSRF token, allowed origin, listen address, harness
  executable and profile, workspace root, lease duration.
- Startup: pool, `postgres.MigrateUp`, ID/clock sources, driver registration
  (codex + fake harness; local sandbox), supervisor, runner, outbox worker,
  event stream, HTTP handler, static web assets, graceful shutdown that drains
  subscribers and releases leases.
- Deliver a bootstrap endpoint or injected config so the browser can obtain its
  CSRF token without it being placed in a cookie.
- `deployments/`: Dockerfile for `winchd`, compose file with PostgreSQL, a
  `make run` target, and a documented vite dev proxy for the web workspace.
- Acceptance: `docker compose up` yields a browser session that creates a run,
  starts it against the fake harness, streams output, sends input, and stops it;
  the daemon refuses to start with weak or missing secrets; restart does not
  lose accepted input.
- Verification: browser end-to-end scenario against the fake harness;
  configuration validation tests; a documented manual check against the real
  Codex CLI.

## Sequencing

```
P1-021 ──► P1-022 ──► P1-023 ──► P1-024 ──► P1-025
                 └──► P1-020 (already planned, currently blocked)
```

Strictly serial. P1-021 and P1-022 carry the design risk; P1-024 and P1-025 are
mechanical once the layers below exist. P1-021 and P1-022 can be merged into
one task if the port change proves small, and P1-024 can be folded into P1-025,
giving a three-task path at the cost of larger reviews.

## Constraints carried into these tasks

- **Credentials remain out of band.** `BuildLaunch` accepts
  `ResolvedCredentials`, but no task supplies them until P3-032, and nothing
  owns acquisition at all. The minimal deployment works only because
  `local-trusted` runs as the host user with the CLI already logged in. Record
  this as an explicit deployment precondition; it is the reason this profile
  cannot be shared.
- **No workspace management.** `CreateRun` takes a workspace path from the
  request body. `application.WorkspaceRepository` exists but no task registers
  or validates workspaces, so P1-023 should treat the path as policy-checked
  input against a configured root and leave the aggregate to a later task.
- **Restart truthfulness needs P1-020.** Until it lands, a restarted daemon
  will misreport in-flight runs. Acceptable for a development deployment, not
  for the Phase 1 exit gate.
- **Telemetry has no foundation.** These five tasks will each invent their own
  logging unless the platform task proposed in
  [`architecture-coverage-review.md`](architecture-coverage-review.md) lands
  first. It is cheapest to do it before P1-022.
