# P4-039: Expose workflow HTTP API

**Phase:** 4 — Top-level workflows
**Dependencies:** P2-026, P4-037, P4-038
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Provide authenticated endpoints for definitions, instances, start, inspect, signal/approve, and cancel.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Extend OpenAPI and generated clients
- Enforce workspace/profile authorization
- Return graph status, attempts, deadlines, outputs by reference, and run lineage

## Non-goals

- Do not return artifact bodies or secrets inline
- Do not allow definition mutation after an instance pins it

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] API contract and authorization tests pass
- [ ] Cancellation is idempotent
- [ ] Pinned instances remain stable after a new definition version
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Validate/regenerate OpenAPI.
- Run workflow HTTP contract tests.
- Run authorization/version-pinning tests.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
