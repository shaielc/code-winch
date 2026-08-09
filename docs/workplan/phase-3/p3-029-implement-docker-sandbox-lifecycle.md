# P3-029: Implement Docker sandbox lifecycle

**Phase:** 3 — Docker isolation
**Dependencies:** P0-008, P3-028
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Implement prepare/start/inspect/stop/cleanup using deterministic labels and opaque handles.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Pinned image digests, non-root user, dropped capabilities, no-new-privileges, seccomp, read-only root, bounded mounts
- CPU, memory, PID, disk where supported, and wall-clock limits
- Stop escalation and orphan discovery by labels

## Non-goals

- Do not mount Docker socket, host home, devices, or privileged namespaces
- Do not claim unsupported engine controls

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Shared sandbox contracts pass
- [ ] Inspect/stop/cleanup are idempotent
- [ ] Daemon crash leaves resources discoverable and cleanable
- [ ] Effective image digest and posture are observable
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run opt-in Docker contract suite.
- Run resource/privilege assertion tests.
- Run daemon-crash orphan cleanup test.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
