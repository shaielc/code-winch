# P4-038: Implement control-flow workflow activities

**Phase:** 4 — Top-level workflows
**Dependencies:** P2-023, P4-036
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Implement condition, bounded parallel, bounded foreach, artifact.publish, and declared compensation behavior.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Deterministic branch IDs and joins
- Configurable fail-fast/collect behavior
- Concurrency limits across replay and worker restarts
- Authorized artifact publication

## Non-goals

- Do not allow unbounded fan-out
- Do not infer compensation for undeclared side effects

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Fan-out never exceeds declared bounds
- [ ] Branch results join deterministically
- [ ] Restart preserves branch status and capacity accounting
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run branch ordering/property tests.
- Run concurrency-limit stress test.
- Run restart and compensation tests.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
