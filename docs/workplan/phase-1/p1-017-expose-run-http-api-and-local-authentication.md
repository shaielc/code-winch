# P1-017: Expose run HTTP API and local authentication

**Phase:** 1 — Local single-user vertical slice
**Dependencies:** P1-010, P1-015, P1-016
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Wire run use cases to authenticated HTTP endpoints for the single-user profile.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Implement create/get/start/stop/input and event paging
- Map domain errors to OpenAPI problems
- Add secure local session behavior, CSRF protection, request limits, and audit-safe logging

## Non-goals

- Do not return credential values
- Do not rely on possession of a run ID as authorization

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Generated contract tests pass
- [ ] All mutations enforce authentication and CSRF policy
- [ ] Logs include IDs but exclude message/secret content by default
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run HTTP contract tests.
- Run auth/CSRF and payload-limit tests.
- Run log canary test.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
