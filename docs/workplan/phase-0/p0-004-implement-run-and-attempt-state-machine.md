# P0-004: Implement run and attempt state machine

**Phase:** 0 — Contracts and development foundation
**Dependencies:** P0-003
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Encode lifecycle transitions and invariants from the run contract as dependency-free domain behavior.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Model Created, Queued, Preparing, Running, Stopping, Completed, Failed, and Cancelled
- Make terminal states immutable and stop idempotent
- Represent retries as linked attempts rather than history mutation

## Non-goals

- Do not execute processes or persist state
- Do not infer adapter input capability from lifecycle alone

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Every documented transition has a test
- [ ] Illegal transitions return stable domain errors without mutation
- [ ] Retry preserves prior attempt history
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run table tests for every state/command pair.
- Run tests for idempotent stop and terminal immutability.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
