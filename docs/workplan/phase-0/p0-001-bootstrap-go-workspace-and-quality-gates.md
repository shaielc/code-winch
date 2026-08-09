# P0-001: Bootstrap Go workspace and quality gates

**Phase:** 0 — Contracts and development foundation
**Dependencies:** None
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Create the Go module, daemon composition-root skeleton, and repeatable format, vet, lint, and unit-test commands.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Add the root Go module and the minimum `cmd/winchd` entry point
- Define repository-wide Go formatting, vet, lint, and test commands
- Add CI checks that invoke the same local commands

## Non-goals

- Do not create empty architecture directories
- Do not select production framework dependencies yet

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] A clean checkout can build and test the daemon
- [ ] CI fails on an unformatted or failing Go test
- [ ] The daemon skeleton contains no domain or adapter behavior
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run formatting verification.
- Run Go unit tests and vet.
- Build `cmd/winchd`.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
