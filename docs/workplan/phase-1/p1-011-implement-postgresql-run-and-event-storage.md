# P1-011: Implement PostgreSQL run and event storage

**Phase:** 1 — Local single-user vertical slice
**Dependencies:** P0-004, P0-005, P0-007
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Persist runs, attempts, commands, and gap-free event sequences with optimistic concurrency.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Add ordered migrations and repositories
- Append batches using expected sequence in one transaction
- Page events after a sequence and record fully resolved non-secret configuration

## Non-goals

- Do not publish events from repository code
- Do not persist resolved secret values

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Concurrent appenders cannot create gaps or duplicate sequences
- [ ] Terminal history is immutable
- [ ] Migration up/down or forward-only policy is tested on a clean database
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run repository integration tests against PostgreSQL.
- Run concurrent append stress test.
- Scan persisted fixtures for secret canaries.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
