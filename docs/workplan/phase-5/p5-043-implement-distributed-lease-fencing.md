# P5-043: Implement distributed lease fencing

**Phase:** 5 — Remote runners and hardening
**Dependencies:** P5-042
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Prevent two control-plane or runner owners from controlling one execution during delay, loss, or partition.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Lease token/epoch on every command and observation
- Renewal, expiry, takeover, stale rejection, and explicit runner self-fencing
- Persist diagnostics for rejected stale events outside canonical history

## Non-goals

- Do not depend on wall-clock agreement alone
- Do not automatically adopt an execution across epochs

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Partition scenarios never yield two accepted owners
- [ ] Stale events cannot enter canonical history
- [ ] Takeover has a truthful observed outcome
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run deterministic partition model tests.
- Run epoch takeover integration tests.
- Verify stale diagnostic separation.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
