# P5-042: Implement remote command and event transport

**Phase:** 5 — Remote runners and hardening
**Dependencies:** P1-015, P5-041
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Reliably correlate bounded runner commands and observations while preserving control-plane sequencing authority.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Command IDs, acknowledgements, retry rules, outstanding limits, runner-local ordinals, and diagnostic retention
- Reconnect/resume behavior
- Artifact handoff references but not scheduler policy

## Non-goals

- Do not assign canonical sequence numbers on runners
- Do not insert stale observations into run history

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Transport retry duplicates no command effect
- [ ] Reconnection detects missing local ordinals
- [ ] Backpressure is bounded and visible
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run disconnect/reorder/duplicate fault tests.
- Run buffer exhaustion tests.
- Run end-to-end remote fake-runner test.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
