# Phase 1 — Browser-reachable single-user product

**Status:** not derived. This phase has no tasks in
[`../tasks.json`](../tasks.json) and no task briefs. It exists so that phase-0
briefs can defer work to it under I6, and so that the deferrals are registered
somewhere a reader can find them.

## Objective

A person opens a browser, authenticates, and drives a run to completion without
touching the CLI.

## Scope

Candidate scope, from `docs/roadmap.md` Phase 1 and the gaps
[`docs/state.md`](../../state.md) records under *What is not implemented*. It is
not decomposed into tasks and the boundaries will move when it is.

- **Browser session establishment** — a login path that issues the
  `winch_session` cookie the handler already accepts, so the SPA can call the
  API it is served alongside.
- **A controllable in-memory store profile, and I4's first rung** — phase 0 runs
  on PostgreSQL throughout, so the standing scenario suite has no fake profile to
  start from and skips entirely without a database. This phase derives the store
  profile from the ports the run path consumes, then re-runs the same suite
  against `storeProfile=postgres` unchanged, which is the swap I4 actually asks
  for. Phase 0's withdrawn attempt and why it could not be built are recorded in
  [`../post-mortems/2026-09-11-a-store-profile-derived-from-test-doubles.md`](../post-mortems/2026-09-11-a-store-profile-derived-from-test-doubles.md).
- **Credentials and workspaces** — the `Workspace` and `Credential` aggregates
  from `docs/architecture.md` §6, replacing the raw `workspacePath` string a run
  carries today, and giving `application.SecretReferenceStore` a real
  implementation.
- **Restart reconciliation, demonstrated** — `internal/supervisor` already
  implements it and no scenario proves it against the running daemon.
- **Resize and the remaining input kinds** — the API models `resize` and
  `terminal_bytes`; the web app renders a disabled resize control.
- **Retry and queue admission** — `domain.Run`'s `Failed → Queued` transition is
  implemented and reachable through no API, and `RunStateQueued` is unbounded by
  any admission control.

## Deferrals in

The register. Every phase-0 brief that names this phase as its owner appears
here, with the gap stated in terms of the system, as does anything phase 0 exits
without — an invariant it does not close counts, whether or not a brief said so.
A deferral naming phase 1 and absent from this table is not owned.

| Deferred | From | The gap |
|---|---|---|
| A runnable, controllable in-memory store profile (I3 for storage) | phase 0 exit | `internal/adapters/memory` has no non-test importer; `winchd` cannot start without PostgreSQL, so there is no fake profile to run the product on |
| The standing scenario suite's all-fake first rung (I4) | phase 0 exit | Every file in `test/e2e/` requires `PG_TEST_DATABASE_URL` and skips without it, so "same scenario, more real" has no baseline to compare against |
| Same round trip against `storeProfile=postgres` as a swap (I4) | phase 0 exit | Postgres is the only profile, so the suite proves no parity — it is the baseline and the substrate at once |
| Browser `winch_session` establishment | P0-011 | `httpapi.SetSessionCookie` exists and has no caller, so a browser is served the SPA and gets 401 from every API call |

When this phase is derived, each row becomes a task or is re-deferred
explicitly. Deriving the phase and leaving a row unowned is a coverage defect.

## Phase exit

Provisional, from `docs/roadmap.md` Phase 1: a browser reconnects to a live run
without losing ordered history, daemon restart produces a truthful terminal
state, and forced stop leaves no child processes. Phase 0 demonstrates the third
clause through the CLI; the first two are this phase's.

## Traces to

- `docs/roadmap.md` Phase 1
- `docs/architecture.md` §1 (browser-based login), §6 (`Workspace`,
  `Credential` aggregates)
- [`docs/state.md`](../../state.md) — *The run round trip*, *Credentials,
  workspaces, and login*, *Approvals, retry, and queue admission*
- [`../post-mortems/2026-09-11-a-store-profile-derived-from-test-doubles.md`](../post-mortems/2026-09-11-a-store-profile-derived-from-test-doubles.md)
  — why phase 0's in-memory store profile was withdrawn and redrawn here
