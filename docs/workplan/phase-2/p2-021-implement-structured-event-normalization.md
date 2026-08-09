# P2-021: Implement structured event normalization

**Phase:** 2 — Structured experience and second harness
**Dependencies:** P0-005, P1-014
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Complete typed messages, tools, approvals, artifacts, file changes, usage, and safe provider-extension handling.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Add domain validators and adapter mapping helpers
- Enforce event-size and sensitivity policies
- Preserve unknown kinds/extensions with diagnostic fallback

## Non-goals

- Do not make extensions authoritative for generic workflow behavior
- Do not discard malformed raw evidence silently

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Every family has valid/invalid fixtures
- [ ] Malformed provider data degrades without terminating the run
- [ ] Sensitivity is explicit for all persisted events
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run schema and normalization tests.
- Fuzz incremental parsers and extension decoding.
- Run event-size boundary tests.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
