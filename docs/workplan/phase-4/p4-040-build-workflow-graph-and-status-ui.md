# P4-040: Build workflow graph and status UI

**Phase:** 4 — Top-level workflows
**Dependencies:** P4-039
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Let users start, inspect, and cancel workflows and every active branch.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Definition/version selection, input validation, graph/step attempt status, timers/retries, lineage links, approval surfaces, and cancellation
- Reconnect through persisted snapshots/event deltas
- Accessible status and error representations

## Non-goals

- Do not hide partial branch failures
- Do not permit UI-only enforcement of policy

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] A user can inspect and cancel every active branch
- [ ] Reload/reconnect reconstructs the same graph state
- [ ] Run links navigate to canonical run views
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run renderer/component tests.
- Run browser workflow replay/reconnect test.
- Run accessibility checks.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
