# P2-026: Add workspace authorization and audit trail

**Phase:** 2 — Structured experience and second harness
**Dependencies:** P1-017, P1-018, P2-023
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Enforce user/workspace authorization across runs, inputs, artifacts, credentials, streams, and audit reads.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Centralize authorization decisions
- Reauthorize stream access on relevant policy changes
- Record append-only security events with actor, action, resource, result, and safe metadata

## Non-goals

- Do not infer authorization from resource IDs
- Do not include content or secrets in audit metadata by default

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Every endpoint and stream has allow/deny tests
- [ ] Revocation terminates or blocks ongoing access promptly
- [ ] Audit records cover launches, stops, approvals, credential use, exports, and policy changes
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run authorization matrix tests.
- Run confused-deputy/ID-enumeration tests.
- Run stream revocation test.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
