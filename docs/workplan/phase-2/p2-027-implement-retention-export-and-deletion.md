# P2-027: Implement retention, export, and deletion

**Phase:** 2 — Structured experience and second harness
**Dependencies:** P2-023, P2-026
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Apply sensitivity-aware lifecycle policies while preserving audit and referential integrity requirements.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Retention worker for events, artifacts, and renderer caches
- Authorized portable export with manifest/digests
- Deletion/tombstone workflow for runs, workspaces, credentials, and workflow links

## Non-goals

- Do not promise deletion the storage backend cannot prove
- Do not export secrets or unauthorized extensions

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Each sensitivity class follows documented defaults
- [ ] Export is complete, authorized, and integrity-verifiable
- [ ] Deletion is retryable and reports partial failures
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run fake-time retention tests.
- Run export round-trip and authorization tests.
- Run deletion crash/retry integration tests.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
