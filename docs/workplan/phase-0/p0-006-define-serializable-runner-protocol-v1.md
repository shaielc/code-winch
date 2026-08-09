# P0-006: Define serializable runner protocol v1

**Phase:** 0 — Contracts and development foundation
**Dependencies:** P0-003, P0-005
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Create bounded, versioned wire schemas for the in-process runner boundary before any process implementation.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Model prepare, start, input, resize, stop, inspect, and cleanup commands
- Model execution observations with local ordinal, command correlation, execution ID, and lease token
- Define protocol negotiation and explicit payload/buffer limits

## Non-goals

- Do not implement network transport
- Do not expose OS process or container handles

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] All runner messages serialize and round-trip
- [ ] A major-version mismatch is rejected
- [ ] Unknown minor-version fields are ignored and preserved where required
- [ ] Oversized messages fail with a stable error
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run protocol fixture compatibility tests.
- Run boundary-size and negotiation tests.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
