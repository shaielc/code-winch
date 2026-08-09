# P0-002: Bootstrap web workspace and generated-client checks

**Phase:** 0 — Contracts and development foundation
**Dependencies:** None
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Create the TypeScript/React workspace and its deterministic quality gates without implementing product UI.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Set up the web application, type checking, linting, unit tests, and production build
- Reserve `web/src/api` for generated API code
- Add a CI check that regeneration leaves the tree clean

## Non-goals

- Do not hand-write a duplicate API client
- Do not build run screens in this task

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] The web app builds from a clean checkout
- [ ] Type, lint, and unit-test failures block CI
- [ ] Generated-client drift is detectable
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Install dependencies with the locked package manager.
- Run lint, typecheck, tests, and production build.
- Regenerate the client and check `git diff --exit-code`.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
