# P3-033: Display and test effective security posture

**Phase:** 3 — Docker isolation
**Dependencies:** P3-030, P3-031, P3-032
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Make the resolved filesystem, network, credential, image, and isolation posture visible and resistant to UI ambiguity.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Expose safe effective-posture fields through API
- Render warnings for local/unavailable controls and exact profile details for Docker
- Add adversarial end-to-end scenarios for forbidden overrides and malicious output

## Non-goals

- Do not imply Docker is a hard hostile multi-tenant boundary
- Do not expose secret material or unsafe daemon details

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Users can distinguish local-trusted, standard, and readonly posture before start
- [ ] Rejected overrides are explained without leaking policy internals
- [ ] UI labels match effective driver observations
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run API posture contract tests.
- Run browser security-posture tests.
- Run adversarial override end-to-end test.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
