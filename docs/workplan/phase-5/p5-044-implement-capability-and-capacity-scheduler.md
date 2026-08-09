# P5-044: Implement capability and capacity scheduler

**Phase:** 5 — Remote runners and hardening
**Dependencies:** P5-041, P5-043
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Assign queued runs only to healthy runners satisfying harness, sandbox, policy, and resource requirements.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Filter by versioned capabilities and security profile
- Account for CPU/memory/run slots and locality constraints
- Explain pending/unschedulable reasons and avoid starvation

## Non-goals

- Do not silently downgrade a security requirement
- Do not schedule based on stale heartbeat indefinitely

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] No incompatible runner receives an assignment
- [ ] Capacity is not overcommitted beyond policy
- [ ] Unschedulable state is stable and observable
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run scheduling matrix/property tests.
- Run stale-heartbeat and fairness tests.
- Run multi-runner capacity simulation.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
