# P4-036: Implement workflow coordinator replay loop

**Phase:** 4 — Top-level workflows
**Dependencies:** P4-035
**Can run in parallel with:** Any task whose dependency list is satisfied and which does not edit the same contract surface.

## Objective

Drive ready steps from durable state and domain events so restart/replay is safe.

## Architectural context

Implement this as a thin increment behind the ports and boundaries defined in `docs/architecture.md`, `docs/code-structure.md`, and `docs/contracts.md`. Apply the threat model and secure defaults from `docs/security.md`; the phase gate remains authoritative in `docs/roadmap.md`.

## Scope

- Claim, dispatch, wait, retry/backoff, timeout, complete/fail, cancel, and compensation orchestration
- Use deterministic idempotency keys from instance, step, and attempt
- Consume run events through application ports rather than process control

## Non-goals

- Do not call sandbox or harness drivers directly
- Do not use sleeps for durable timers

## Deliverables

- Production code and configuration for the scoped behavior, placed according to `docs/code-structure.md`.
- Focused automated tests and deterministic fixtures; use injected time, IDs, runner, publisher, and secrets where relevant.
- Contract/schema/migration updates required by the scope, including generated outputs from their declared source of truth.
- Brief operator or developer documentation for any new command, configuration, failure mode, or security posture.

## Acceptance criteria

- [ ] Restart at every dispatch boundary duplicates no external effect
- [ ] Retries create explicit attempts
- [ ] Cancellation reaches every active branch and is observable
- [ ] Errors are stable and actionable; logs/traces include resource IDs but exclude content and secrets by default.
- [ ] The implementation does not introduce an outward dependency into the domain or a cross-adapter import.

## Required verification

- Run deterministic replay/crash matrix.
- Run fake-time timeout/retry tests.
- Run cancellation propagation tests.
- Run the repository format, lint, unit-test, and build checks affected by the change.

## Implementation notes

- Prefer the smallest end-to-end behavior that proves the contract; keep optional optimizations out of this task.
- Test failure and replay/idempotency behavior at every durable or asynchronous boundary introduced here.
- Update this brief only when architectural review changes its scope, dependencies, or acceptance criteria.
