# P1-019: Build terminal run web slice

**Phase:** 1 — Local single-user vertical slice
**Dependencies:** P0-002, P1-017, P1-018
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Let a browser create, observe, control, and reconnect to one run with a safe terminal projection.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Run list/detail shell, terminal view, input, resize, stop control, connection state, and sequence cursor persistence
- Safely interpret ANSI in a sandboxed component
- Display harness/sandbox capabilities and a prominent `local-trusted` warning

## Non-goals

- Do not render raw HTML
- Do not hide unsupported controls without explanation

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] A browser can reconnect without losing or duplicating visible output
- [ ] Unsupported controls are disabled with capability context
- [ ] Malicious ANSI/URLs cannot execute script or escape CSP
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run renderer unit and malicious-payload tests.
- Run browser end-to-end run/reconnect/stop test.
- Run accessibility checks.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
