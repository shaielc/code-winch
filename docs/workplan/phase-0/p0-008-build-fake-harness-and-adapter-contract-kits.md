# P0-008: Build fake harness and adapter contract kits

**Phase:** 0 — Contracts and development foundation
**Dependencies:** P0-005, P0-006, P0-007
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Provide a deterministic executable harness plus reusable harness and sandbox contract suites.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Fake output chunking, structured events, exit codes, delays controlled by fake time, and malformed records
- Harness suite checks descriptors, codecs, arbitrary byte boundaries, input encoding, flush, and fallback diagnostics
- Sandbox suite checks lifecycle idempotency and only advertised capabilities

## Non-goals

- Do not require a real provider account, Docker, or network
- Do not encode one provider's assumptions in shared assertions

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] The fake harness drives a complete run in CI
- [ ] Codec tests split input at every byte boundary
- [ ] A lying capability implementation fails the shared suite
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run harness contracts.
- Run sandbox contracts against a test double.
- Run fake-harness end-to-end process test.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
