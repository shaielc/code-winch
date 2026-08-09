# P3-028: Implement disposable workspace preparation

**Phase:** 3 — Docker isolation
**Dependencies:** P2-023
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Create policy-checked per-run copies/worktrees and return changes as reviewed artifacts.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Resolve paths beneath approved roots and reject traversal/symlink escapes
- Create uniquely owned disposable workspaces
- Capture changes without mutating the primary checkout and clean idempotently

## Non-goals

- Do not mount the primary checkout writable
- Do not follow repository-controlled links outside policy roots

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Traversal and symlink attacks are rejected
- [ ] Concurrent runs cannot share writable workspace state
- [ ] Crash reconciliation can identify and clean owned workspaces
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run path/symlink adversarial tests.
- Run concurrent workspace isolation tests.
- Run crash cleanup tests.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
