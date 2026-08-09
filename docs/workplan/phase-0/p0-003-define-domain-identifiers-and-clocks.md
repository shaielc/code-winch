# P0-003: Define domain identifiers and clocks

**Phase:** 0 — Contracts and development foundation
**Dependencies:** P0-001
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Implement dependency-free strongly typed identifiers, timestamps, and injectable ID/clock ports used by later domain work.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Define IDs for workspace, run, attempt, event, command, artifact, credential, and workflow records
- Validate parsing and zero/invalid values
- Provide fake clock and deterministic ID sources for tests

## Non-goals

- Do not import persistence, HTTP, or vendor libraries

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Identifiers cannot be accidentally interchanged at compile time
- [ ] Invalid external representations are rejected
- [ ] Tests can produce deterministic IDs and time without sleeps
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run domain unit and property/table tests.
- Run dependency-boundary checks.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
