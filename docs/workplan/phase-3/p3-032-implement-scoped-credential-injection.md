# P3-032: Implement scoped credential injection

**Phase:** 3 — Docker isolation
**Dependencies:** P2-026, P3-029, P3-030
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Resolve opaque references only at launch and inject least-privilege credentials through profile-approved mechanisms.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Authorize credential/provider/workspace binding
- Prefer temporary files or brokered access and track all temporary material
- Redact known values before persistence/logging and delete material during cleanup

## Non-goals

- Do not mount home, SSH/cloud config, or agent sockets
- Do not return credential values after entry

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] A run receives only explicitly referenced authorized credentials
- [ ] Canary values occur zero times in events/default logs
- [ ] Temporary material is gone after success, failure, and daemon restart
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run authorization and scope tests.
- Run secret-canary scan.
- Run crash cleanup integration test.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
