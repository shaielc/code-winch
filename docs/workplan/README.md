# Implementation workplan

Phase 0 repairs the gaps recorded in [`docs/state.md`](../state.md) under *What
went wrong* — invariant failures and quality gates that the previous plan never
established or that broke when the plan closed. Later phases (not written yet)
will implement the remaining capabilities listed under *What is not implemented*.

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
- No `test/e2e` directory or `make e2e` target exists.

## Phase 0 — Foundation repair

Restore invariants and gates the previous plan left broken. The run round trip is
split into one seam per operator operation; each seam includes its minimal CLI
command. Every task leaves the daemon startable and deployable.

| ID | Title | Depends on | Shape | CLI added |
|---|---|---|---|---|
| P0-001 | Repair CI test gates | — | hardening | — |
| P0-002 | Ship operator CLI in build and deployment | — | hardening | — |
| P0-003 | Controllable fake harness profile | — | capability | (flags on `dev run`) |
| P0-004 | Exercise complete fake run in CI | P0-002 | hardening | — |
| P0-005 | Sanitize post-close plan references | — | hardening | — |
| P0-006 | Create and read runs | P0-001 | seam | `run create`, `run get` |
| P0-008 | Start run execution | P0-003, P0-006 | seam | `run start`, `run events` |
| P0-009 | Send input and drain outbox | P0-008 | seam | `run input` |
| P0-010 | Live WebSocket event stream | P0-009 | seam | `run stream` |
| P0-011 | Stop run | P0-010 | seam | `run stop` |
| P0-007 | Standing end-to-end scenario suite | P0-001, P0-011 | capability | — |

### Run round trip chain

```
P0-001 ──► P0-006 create/read ──► P0-008 start ──► P0-009 input+outbox
                                                      ──► P0-010 websocket
                                                            ──► P0-011 stop
                                                                  ──► P0-007 e2e
P0-003 ────────────────► P0-008
```

- **P0-008** persists events and enqueues outbox rows; no worker runs yet.
- **P0-009** starts the outbox worker with a non-streaming publisher; polling
  (`run events`, backlog drain) is the manual check.
- **P0-010** swaps in a publisher that calls `EventStream.Publish`; WebSocket is
  the manual check.
- **P0-011** completes the API surface.

### Phase exit

Phase 0 is complete when all tasks above are `completed`, including `make e2e`
running `create → start → stream → input → stop` in CI.

**Still open after Phase 0** (Phase 1, not yet planned): browser session
establishment, credentials and workspaces, structured events and a second
provider, and every other gap under *What is not implemented* in `docs/state.md`.

### Width

| Metric | Value |
|---|---|
| Critical path length | 7 (P0-001 → P0-006 → P0-008 → P0-009 → P0-010 → P0-011 → P0-007) |
| Average width | 11 tasks ÷ 7 ≈ 1.6 |
| Write collisions | P0-001 ↔ P0-002 (`Makefile`); P0-001 ↔ P0-007 (`Makefile`, `.github/workflows/go.yml`); P0-006–P0-011 share `cmd/winch/` (sequential by dependency, not parallel) |
| Contract collisions | none (P0-003 → P0-008 resolves fake harness launch surface) |

**Available at plan open:** P0-001, P0-002, P0-003, P0-005. The run chain
serializes P0-006 through P0-011; P0-007 follows P0-011.

## Invariants at plan start

| Invariant | Status at open | Phase 0 tasks |
|---|---|---|
| I1 — system runs | holds (daemon starts) | P0-006–P0-011 |
| I2 — system deploys | holds (compose stack) | P0-002 |
| I3 — controllable fake profile | **fails** | P0-003 |
| I4 — standing scenario suite | **fails** | P0-007 |
| I5 — operator CLI reachable | **partial** | P0-002, P0-004, P0-006–P0-011 |
| I6 — owned deferrals | holds once tracker exists | P0-005 |

## What comes next

Phase 1 will cover the remaining design-set gaps: browser login and session
cookies, credentials and workspaces, structured events, a second harness
provider, and the roadmap exit criteria not satisfied by the fake-profile run
round trip. It will be written when Phase 0 closes or when explicitly requested.
