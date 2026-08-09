# P2-025: Implement second harness adapter

**Phase:** 2 — Structured experience and second harness
**Dependencies:** P2-021, P2-024
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Add a meaningfully different provider to validate structured and capability-based adapter boundaries.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Implement descriptor, launch, codec, input, exit, and namespaced extensions
- Exercise at least one capability absent from the first adapter and one unsupported capability
- Use deterministic sanitized fixtures

## Non-goals

- Do not add provider conditionals to generic UI, supervisor, or workflow code
- Do not weaken shared contracts to fit the provider

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Both adapters pass the same shared suite
- [ ] Generic views consume both without provider branches
- [ ] Provider-only data remains available only through its extension renderer
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run both harness contract suites.
- Run cross-provider generic renderer tests.
- Run second fake-CLI integration test.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
