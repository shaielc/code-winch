# P5-047: Define SLOs, dashboards, alerts, and runbooks

**Phase:** 5 — Remote runners and hardening
**Dependencies:** P5-044, P5-046
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Operationalize queue, execution, event, workflow, cleanup, and security health with content-safe telemetry.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Metrics and traces for documented reliability indicators
- Dashboards and actionable alerts
- Runbooks for runner loss, lease contention, event backlog, cleanup failure, restore, secret incident, and runner revocation
- Measure initial acceptance targets

## Non-goals

- Do not place message content or secrets in default telemetry
- Do not alert without an owner and response action

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Acceptance metrics are automatically measurable
- [ ] Every alert links to a tested runbook
- [ ] Telemetry canaries prove content exclusion
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run dashboard/alert rule validation.
- Execute representative runbook drills.
- Run telemetry secret-canary scan.
- Run local event-latency and stop-cleanup benchmark.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
