# P2-024: Implement approval binding and policy evaluation

**Phase:** 2 — Structured experience and second harness
**Dependencies:** P1-016, P2-021
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Normalize approvals into expiring, single-use decisions bound to the exact requested operation.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Compute operation digests over command/tool, working directory, and file/network scope
- Intersect deployment, user/workspace, workflow, and sandbox policies
- Audit request, decision, expiry, and attempted replay

## Non-goals

- Do not let a workflow broaden permissions
- Do not accept provider approval IDs without local binding

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Modified, expired, or reused approvals are rejected
- [ ] Auto-deny/require-user/narrow-allow paths are deterministic
- [ ] Audit records contain no secret values
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run policy matrix tests.
- Run digest tamper/replay/expiry tests with fake time.
- Run audit canary tests.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
