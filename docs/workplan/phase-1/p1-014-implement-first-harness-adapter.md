# P1-014: Implement first harness adapter

**Phase:** 1 — Local single-user vertical slice
**Dependencies:** P0-008, P1-013
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Integrate one selected coding-agent provider behind the harness port, using sanitized fixtures and capability discovery.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Descriptor, configuration validation, launch construction, incremental codec, input encoding, exit mapping, and provider extension namespace
- Map malformed structured output to diagnostics/raw output without data loss
- Provide fake CLI or sanitized transcript fixtures

## Non-goals

- Do not access persistence, web sessions, containers, or secrets directly
- Do not add provider branches to generic application code

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Harness contract suite passes for arbitrary chunk boundaries
- [ ] Every advertised capability has a positive and negative test
- [ ] No live provider account is required in CI
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run harness contract suite.
- Run sanitized transcript golden tests.
- Run adapter/fake-CLI integration test.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
