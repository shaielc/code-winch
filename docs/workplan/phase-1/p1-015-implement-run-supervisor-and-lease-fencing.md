# P1-015: Implement run supervisor and lease fencing

**Phase:** 1 — Local single-user vertical slice
**Dependencies:** P0-007, P1-011, P1-012, P1-013, P1-014
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Serialize commands per active run, enforce lifecycle, and fence execution ownership with lease epochs.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Mailbox/command serialization, lease acquire/renew/release, driver resolution, sequence assignment, and stale-token rejection
- Redact before persistence and publication
- Persist desired state before runner interaction

## Non-goals

- Do not keep correctness solely in goroutine memory
- Do not let runner assign canonical sequences

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Concurrent start/stop/input has one deterministic order
- [ ] Only the current lease epoch can append observations
- [ ] Restart can rehydrate from durable state
- [ ] Redaction precedes storage in every path
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run race-enabled supervisor tests.
- Run lease takeover and stale-event tests.
- Run crash-point tests around desired-state persistence.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
