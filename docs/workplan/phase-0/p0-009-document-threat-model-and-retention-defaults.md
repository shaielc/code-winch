# P0-009: Document threat model and retention defaults

**Phase:** 0 — Contracts and development foundation
**Dependencies:** None
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Turn security assumptions into reviewable, test-oriented operational defaults for the first deployment.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Inventory assets, actors, trust boundaries, abuse cases, and mitigations
- Define sensitivity-based retention, export, telemetry, and deletion defaults
- Define launch blockers for shared deployment and owners for residual risks

## Non-goals

- Do not claim the local driver provides isolation
- Do not store example plaintext credentials

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Threats cover browser, control plane, runner, sandbox, repository, network, secrets, and renderer
- [ ] Each mitigation maps to a future task or accepted residual risk
- [ ] Retention defaults specify behavior for every sensitivity class
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run documentation link/lint checks.
- Review against `docs/security.md` checklist.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
