# P1-020: Reconcile runs after daemon restart

**Phase:** 1 — Local single-user vertical slice
**Dependencies:** P1-015
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Compare persisted intent with local execution observations and converge every nonterminal run truthfully.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Inspect known execution handles through the runner boundary
- Resume ownership when safe; otherwise mark lost/failed with diagnostics
- Continue pending stop escalation and cleanup
- Emit observable reconciliation transitions

## Non-goals

- Do not invent success when process state is unknown
- Do not adopt an execution without valid identity/lease evidence

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Restart during prepare, run, input delivery, and stop has a defined outcome
- [ ] No execution gains two supervisors
- [ ] Orphaned owned resources are cleaned or prominently diagnosed
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run restart matrix integration tests.
- Run lease-takeover tests.
- Run orphan reconciliation tests.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
