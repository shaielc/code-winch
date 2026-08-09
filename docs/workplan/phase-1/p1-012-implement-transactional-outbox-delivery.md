# P1-012: Implement transactional outbox delivery

**Phase:** 1 — Local single-user vertical slice
**Dependencies:** P1-011
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Atomically record publish intent with mutations and deliver it at least once without corrupting event order.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Add outbox migration, claim/lease query, worker, retry/backoff, and completion markers
- Publish run events only after commit
- Expose backlog, retry, and poison-record metrics

## Non-goals

- Do not treat the live publisher as authoritative storage
- Do not require a broker

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] A crash between commit and publish loses no event
- [ ] Duplicate delivery is safe for subscribers
- [ ] Multiple workers cannot concurrently own the same record
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run crash-point integration tests.
- Run multi-worker claim tests.
- Run retry/backoff tests with fake time.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
