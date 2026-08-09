# P4-034: Define and validate workflow graph schemas

**Phase:** 4 — Top-level workflows
**Dependencies:** P0-005, P1-010
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Specify versioned declarative definitions with stable step IDs, typed references, policies, and no embedded code.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Schema run.start/send/stop, event.wait, approval.wait, condition, parallel, foreach, and artifact.publish
- Validate cycles, references, types, bounded concurrency, retry/timeout, and compensation declarations
- Pin definition version and profiles in instances

## Non-goals

- Do not execute definitions in this task
- Do not allow arbitrary server code expressions

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Invalid graphs fail with path-specific diagnostics
- [ ] Old fixtures remain decodable
- [ ] Parallel/foreach declarations require explicit bounds
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run schema compatibility tests.
- Run graph/type validation table tests.
- Fuzz definition parser limits.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
