# P1-018: Implement resumable WebSocket event stream

**Phase:** 1 — Local single-user vertical slice
**Dependencies:** P1-012, P1-017
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Deliver ordered persisted history plus live deltas without allowing a slow browser to block execution.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Authenticate and validate Origin
- Support `after_sequence`, heartbeat, `caught_up`, gap repair, and resumable disconnect indication
- Reauthorize long-lived connections and bound subscriber buffers

## Non-goals

- Do not use the socket as event storage
- Do not put long-lived bearer credentials in URLs

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Reconnect receives every persisted event exactly once at the view boundary
- [ ] Snapshot/live handoff has no race-created gap
- [ ] Slow subscribers are disconnected without backpressuring a run
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run reconnect and snapshot/live race tests.
- Run slow-consumer/backpressure tests.
- Run Origin and reauthorization tests.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
