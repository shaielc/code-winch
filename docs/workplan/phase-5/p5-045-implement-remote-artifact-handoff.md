# P5-045: Implement remote artifact handoff

**Phase:** 5 — Remote runners and hardening
**Dependencies:** P2-023, P5-042
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Transfer content-addressed artifacts without proxying unbounded bodies through control-plane event streams.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Signed storage handoff or bounded chunk protocol
- Digest verification, resume, expiry, authorization, quotas, and cleanup
- Associate artifact only after verified completion

## Non-goals

- Do not trust runner-provided digest or media type without verification
- Do not include credentials in artifact events

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Interrupted uploads resume or fail cleanly
- [ ] Digest mismatch never creates a published artifact
- [ ] Expired grants cannot upload or download
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run interrupted/resumed transfer tests.
- Run digest/quota/expiry tests.
- Run unauthorized handoff tests.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
