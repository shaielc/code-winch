# P3-031: Enforce container network policy

**Phase:** 3 — Docker isolation
**Dependencies:** P3-029, P3-030
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Provide deny-by-default egress with explicit administrator allowlists and metadata endpoint protection.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Create per-run network identity
- Enforce DNS/IP/port rules outside the container namespace where practical
- Block link-local/cloud metadata and record safe network audit facts
- Clean network resources idempotently

## Non-goals

- Do not claim TLS content inspection
- Do not trust in-container firewall rules as the sole boundary

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Deny profile has no egress including DNS and metadata
- [ ] Allowlist permits only declared destinations/ports
- [ ] Policy survives container attempts to alter its namespace
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run adversarial egress integration tests.
- Run DNS rebinding/metadata endpoint tests.
- Run network cleanup reconciliation tests.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
