# Phase 1 — Browser-reachable single-user product

**Status:** not derived. This phase has no tasks in
[`../tasks.json`](../tasks.json) and no task briefs. It exists so that phase-0
briefs can defer work to it under I6, and so that the deferrals are registered
somewhere a reader can find them.

## Objective

A person opens a browser, authenticates, and drives a run to completion without
touching the CLI; and the product runs end to end with no database, so the
standing scenario suite has a substrate to prove parity against.

## Scope

Candidate scope, from `docs/roadmap.md` Phase 1 and the gaps
[`docs/state.md`](../../state.md) records under *What is not implemented*. It is
not decomposed into tasks and the boundaries will move when it is.

- **Browser session establishment** — a login path that issues the
  `winch_session` cookie the handler already accepts, so the SPA can call the
  API it is served alongside.
- **A controllable in-memory store profile** — `winchd` boots against
  `internal/adapters/memory` with no database, and the phase-0 scenarios pass
  against it unchanged. This is I4's first rung, built after the fact.
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
here, with the gap stated in terms of the system. A deferral naming phase 1 and
absent from this table is not owned.

| Deferred | From | The gap |
|---|---|---|
| Controllable in-memory store profile | P0-003 | `internal/adapters/memory` implements every port and no runtime configuration reaches it, so no phase-0 configuration runs without PostgreSQL |
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
- [`../post-mortems/2026-08-23-workplan-create-phase-0.md`](../post-mortems/2026-08-23-workplan-create-phase-0.md)
  defect 5 — why the in-memory profile arrives here instead of in phase 0
