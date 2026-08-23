# Implementation workplan

Phase 0 repairs the gaps recorded in [`docs/state.md`](../state.md) under *What
went wrong* — invariant failures and quality gates that the previous plan never
established or that broke when the plan closed. Later phases (not written yet)
will implement the capabilities listed under *What is not implemented*.

Pick up work with `scripts/list-available-tasks.sh`. Each task is one pull
request; include `Task: <ID>` in the body. Automation stamps `completed` in
`tasks.json` when the pull request is approved — do not edit status fields by
hand.

## Inherited state

At plan creation (`workplan/2026_08_21`), HEAD matches the close report at
`ccad757`:

- The daemon starts, migrates, and serves; `make check` passes.
- Run routes return 500/404 because `unavailableBackend` is still bound
  (`cmd/winchd/main.go:159-183`).
- `winch dev run` drives a fake harness locally; the operator CLI is not built by
  `make build` or installed in the deployment image.
- The fake harness is not controllable and is exercised only by hand.
- CI runs `make check` only; storage integration tests and a complete fake run
  are not gated.
- `docs/workplan/tasks.json` was absent and blocked the task-status gate on every
  pull request.

## Phase 0 — Foundation repair

Restore invariants and gates the previous plan left broken, without binding the
run use cases. Every task leaves the daemon startable and deployable.

| ID | Title | Depends on | Shape |
|---|---|---|---|
| P0-001 | Repair CI test gates | — | hardening |
| P0-002 | Ship operator CLI in build and deployment | — | hardening |
| P0-003 | Controllable fake harness profile | — | capability |
| P0-004 | Exercise complete fake run in CI | P0-002 | hardening |
| P0-005 | Sanitize post-close plan references | — | hardening |

### Phase exit

Phase 0 is complete when:

- `make check` on CI also runs PostgreSQL-backed storage integration tests.
- `make test` does not descend into `web/node_modules`.
- `make build` produces `winch` and the deployment image installs it on `PATH`.
- The fake harness profile accepts scripted transcripts and injectable failure
  modes at runtime.
- CI executes a complete fake harness run (the criterion the previous plan
  marked complete but never proved).
- No document or comment cites a removed brief path or stale task ID.

**Still open after Phase 0** (Phase 1, not yet planned): run round trip through
the daemon, the standing `create → start → stream → input → stop` scenario
suite (I4), browser session establishment, and every gap under *What is not
implemented* in `docs/state.md`.

### Width

| Metric | Value |
|---|---|
| Critical path length | 2 (P0-002 → P0-004) |
| Average width | 5 tasks ÷ 2 = 2.5 |
| Write collisions | P0-001 ↔ P0-002 (`Makefile`) |
| Contract collisions | none |

P0-001 and P0-002 both touch `Makefile`; expect a rebase if they land in
parallel. No two concurrently available tasks share a contract surface.

## Invariants at plan start

| Invariant | Status at open | Phase 0 tasks |
|---|---|---|
| I1 — system runs | holds (daemon starts) | — |
| I2 — system deploys | holds (compose stack) | P0-002 extends deploy surface |
| I3 — controllable fake profile | **fails** | P0-003 |
| I4 — standing scenario suite | **fails** | deferred to Phase 1 (requires run binding) |
| I5 — operator CLI reachable | **partial** | P0-002, P0-004 |
| I6 — owned deferrals | holds once tracker exists | P0-005 |

## What comes next

Phase 1 will bind the run use cases, wire unreachable packages into the
composition root, and establish the end-to-end scenario suite against the fake
profile. It is derived from *What is not implemented* in `docs/state.md` and
will be written when Phase 0 closes or when explicitly requested.
