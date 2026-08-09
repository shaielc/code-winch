# P1-010: Design initial OpenAPI run contract

**Phase:** 1 — Local single-user vertical slice
**Dependencies:** P0-002, P0-004, P0-005
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Define the HTTP contract for creating, reading, starting, stopping, and paging events for a run.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Add schemas, stable problem responses, idempotency headers/fields, pagination, and concurrency preconditions
- Generate Go server types and TypeScript client
- Add compatibility fixtures

## Non-goals

- Do not add WebSocket framing here
- Do not expose credentials or execution handles

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] All operations and domain errors have stable response shapes
- [ ] Generated Go and TypeScript code derives from one source
- [ ] Breaking schema changes are detected
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Validate OpenAPI.
- Regenerate both sides and require a clean diff.
- Run API compatibility tests.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
