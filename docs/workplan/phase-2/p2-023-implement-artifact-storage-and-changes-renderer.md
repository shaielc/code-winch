# P2-023: Implement artifact storage and changes renderer

**Phase:** 2 — Structured experience and second harness
**Dependencies:** P1-011, P1-017, P2-021
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Store content-addressed artifacts outside event envelopes and safely render file-change/diff summaries.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Define artifact metadata and storage port
- Verify digest, media type, size, access, retention, and signed/streamed download
- Render bounded text diffs with binary/large fallback

## Non-goals

- Do not place large binary bodies in PostgreSQL events
- Do not trust filenames, media types, or diff content

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Digest mismatch and traversal filenames are rejected
- [ ] Artifact access is authorized independently
- [ ] Large/binary content has a safe downloadable fallback
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run artifact repository integration tests.
- Run traversal/content-sniffing/size-limit tests.
- Run changes-renderer golden tests.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
