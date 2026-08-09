# P5-041: Implement authenticated runner registration

**Phase:** 5 — Remote runners and hardening
**Dependencies:** P0-006, P3-030
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Move the proven runner wire contract onto an authenticated transport with version/capability negotiation.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Separate runner identity from browser identity
- Mutual authentication, credential rotation hooks, supported protocol range, capability/load registration, and heartbeats
- Bound messages, commands, and buffers

## Non-goals

- Do not assign work across a major-version mismatch
- Do not trust self-declared identity without transport authentication

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Highest compatible version is selected
- [ ] Expired/revoked runner credentials cannot register
- [ ] Heartbeat loss changes runner availability observably
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run protocol negotiation tests.
- Run mTLS/rotation/revocation tests.
- Run message-limit/backpressure tests.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
