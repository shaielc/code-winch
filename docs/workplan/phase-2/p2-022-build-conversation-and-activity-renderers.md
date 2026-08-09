# P2-022: Build conversation and activity renderers

**Phase:** 2 — Structured experience and second harness
**Dependencies:** P1-019, P2-021
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Project authorized canonical events into conversation turns and lifecycle/tool/approval/usage cards.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Coalesce deltas by stable message ID
- Declare supported versions, bounded history, renderer version, and fallback
- Sanitize Markdown and constrain links

## Non-goals

- Do not read database objects or launch data
- Do not execute provider content

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Old and unknown event versions display a safe fallback
- [ ] Out-of-order or duplicate deltas cannot corrupt the view
- [ ] Renderer failure cannot affect execution
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run projection golden tests.
- Run XSS/CSP malicious-payload tests.
- Run renderer failure-isolation tests.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
