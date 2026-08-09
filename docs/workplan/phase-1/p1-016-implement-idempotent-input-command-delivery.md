# P1-016: Implement idempotent input command delivery

**Phase:** 1 — Local single-user vertical slice
**Dependencies:** P1-012, P1-015
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Accept typed input once, persist acceptance/outbox intent, and correlate resulting events to the command.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Support text, interrupt, terminal bytes, and resize for the first slice
- Validate expected state/optional last sequence and adapter capabilities
- Authorize raw terminal input separately

## Non-goals

- Do not accept input solely because a run is `Running`
- Do not retry non-idempotently under a new command ID

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Repeating an idempotency key returns the original result
- [ ] Unsupported or stale-state input produces a stable error
- [ ] Accepted input survives a daemon crash before delivery
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run duplicate/concurrent input tests.
- Run capability and authorization matrix tests.
- Run crash/retry integration tests.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
