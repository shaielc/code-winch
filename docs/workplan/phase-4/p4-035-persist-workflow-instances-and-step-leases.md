# P4-035: Persist workflow instances and step leases

**Phase:** 4 — Top-level workflows
**Dependencies:** P1-012, P4-034
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Add database state for definitions, instances, attempts, timers, signals, lineage, and worker claims.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Ordered migrations and repositories
- Atomic ready-step claim with lease expiry and fencing
- Durable signals/timers and outbox records

## Non-goals

- Do not hold correctness only in a worker process
- Do not overwrite completed attempt history

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Multiple workers cannot own one step attempt
- [ ] Expired leases can be safely reclaimed
- [ ] Instance and attempt history is append-only/reconstructable
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run PostgreSQL repository tests.
- Run multi-worker lease contention test.
- Run migration and crash-point tests.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
