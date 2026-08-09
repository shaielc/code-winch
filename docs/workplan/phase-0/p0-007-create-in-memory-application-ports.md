# P0-007: Create in-memory application ports

**Phase:** 0 — Contracts and development foundation
**Dependencies:** P0-003, P0-005, P0-006
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Define inward-facing ports and deterministic in-memory implementations needed to exercise use cases without infrastructure.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Define repositories, event store, outbox publisher, secret reference store, runner gateway, clock, and ID ports
- Implement concurrency-safe in-memory fakes
- Record calls and inject failures for tests

## Non-goals

- Do not let domain import application ports
- Do not imitate PostgreSQL-specific behavior unless contractually required

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Application tests can inspect all external effects
- [ ] Fakes support idempotency and expected-sequence conflict scenarios
- [ ] Failure injection is deterministic
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run race-enabled tests for in-memory implementations.
- Run application package dependency checks.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
