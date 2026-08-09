# P5-046: Add high-availability and recovery validation

**Phase:** 5 — Remote runners and hardening
**Dependencies:** P4-036, P5-043, P5-045
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Prove stateless API replicas and leased workers recover through process, database, and runner failures.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Multi-replica supervisor/workflow tests
- Backup/restore procedure and restoration verification
- Graceful shutdown/drain and rolling-upgrade compatibility
- Fault-injection scenarios

## Non-goals

- Do not claim HA without measured recovery tests
- Do not make PostgreSQL non-authoritative

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Replica loss duplicates no ownership/effect
- [ ] Restore recovers runs, events, workflows, and artifact references consistently
- [ ] Supported rolling versions interoperate
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run multi-replica failover suite.
- Run backup/restore drill.
- Run rolling-version compatibility test.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
