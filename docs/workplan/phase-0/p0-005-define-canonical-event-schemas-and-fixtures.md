# P0-005: Define canonical event schemas and fixtures

**Phase:** 0 — Contracts and development foundation
**Dependencies:** P0-003
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Specify the versioned canonical event envelope and initial event-family payload schemas.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Add schemas for lifecycle, raw stream, messages, tools, approvals, file changes, artifacts, usage, diagnostics, and workflow linkage
- Define sensitivity and source metadata
- Check in compatibility fixtures including unknown kinds, fields, and extensions

## Non-goals

- Do not persist rendered HTML
- Do not put provider fields in the core payload

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Every event validates against its declared schema
- [ ] Unknown additive fields and namespaced extensions survive compatible decoding
- [ ] Consumers order by sequence, never timestamp
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Validate all schemas and fixtures.
- Round-trip old and unknown-field fixtures.
- Run schema compatibility checks.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
