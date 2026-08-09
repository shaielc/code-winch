# P1-013: Implement local PTY sandbox driver

**Phase:** 1 — Local single-user vertical slice
**Dependencies:** P0-006, P0-008
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Launch and control trusted local processes through the sandbox port with truthful capabilities.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Prepare, start with PTY, inspect, resize, stop escalation, and idempotent cleanup
- Track process groups/children and opaque execution handles
- Advertise `unisolated` and explicitly mark unsupported filesystem/network controls

## Non-goals

- Do not claim host isolation
- Do not inject arbitrary secrets outside the resolved launch scope

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Shared sandbox contracts pass
- [ ] Resize changes PTY dimensions when advertised
- [ ] Forced stop leaves no owned child process
- [ ] Inspect/stop/cleanup tolerate retries
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run PTY integration tests.
- Run sandbox contract suite.
- Run orphan/child cleanup tests.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
