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
command and extends the same standing `make e2e` scenario. Every task leaves the
daemon startable and deployable.

| ID | Title | Depends on | Shape | CLI / e2e |
|---|---|---|---|---|
| P0-001 | Repair CI test gates | — | hardening | — |
| P0-002 | Ship operator CLI in build and deployment | — | hardening | — |
| P0-003 | Controllable fake harness profile | — | capability | flags on `dev run` |
| P0-004 | Exercise complete fake run in CI | P0-002, P0-003 | hardening | `dev run` in CI |
| P0-005 | Sanitize post-close plan references | — | hardening | — |
| P0-006 | Create and read runs | — | seam | `run create/get`; starts `make e2e` |
| P0-008 | Start run execution | P0-003, P0-006 | seam | `run start/events`; e2e +start |
| P0-009 | Send input and drain outbox | P0-008 | seam | `run input`; e2e +input |
| P0-010 | Live WebSocket event stream | P0-009 | seam | `run stream`; e2e +stream |
| P0-011 | Stop run | P0-008 | seam | `run stop`; e2e complete; CI gate |

P0-007 was removed: the standing scenario grows in P0-006 through P0-011 rather
than in a terminal umbrella task.

### Dependency graph

```
Independent at open:
  P0-001   P0-002   P0-003   P0-005   P0-006

P0-002 + P0-003 ──► P0-004

P0-003 + P0-006 ──► P0-008 ──┬──► P0-009 ──► P0-010
                              └──► P0-011 (stop; parallel with 009/010 chain)
```

- **P0-001, P0-002, P0-003** have no edges between them.
- **P0-004** waits for the controllable fake (P0-003) it exercises in CI.
- **P0-006** does not wait for P0-001 — local `make test-integration` and
  `make e2e` suffice during development; P0-011 wires `make e2e` into CI.
- **P0-011** depends only on **P0-008** — stop does not require input or
  WebSocket; it can land in parallel with P0-009 → P0-010.

### Phase exit

Phase 0 is complete when all tasks above are `completed` and `make e2e` runs
the full `create → start → stream → input → stop` scenario in CI (wired in
P0-011).

**Still open after Phase 0** (Phase 1, not yet planned): browser session
establishment, credentials and workspaces, structured events and a second
provider, and every other gap under *What is not implemented* in `docs/state.md`.

### Width

| Metric | Value |
|---|---|
| Critical path length | 5 (P0-003 → P0-006 → P0-008 → P0-009 → P0-010) |
| Average width | 10 tasks ÷ 5 = 2.0 |
| Write collisions | P0-001 ↔ P0-002 (`Makefile`); P0-009 ↔ P0-011 (`cmd/winch/`, `test/e2e/`) if taken in parallel |
| Contract collisions | none (P0-003 → P0-008 resolves fake harness launch surface) |

**Available at plan open:** P0-001, P0-002, P0-003, P0-005, P0-006. After
P0-008 completes, P0-009 and P0-011 are both available.

## Invariants at plan start

| Invariant | Status at open | Phase 0 tasks |
|---|---|---|
| I1 — system runs | holds (daemon starts) | P0-006–P0-011 |
| I2 — system deploys | holds (compose stack) | P0-002 |
| I3 — controllable fake profile | **fails** | P0-003 |
| I4 — standing scenario suite | **fails** | P0-006–P0-011 (iterative); CI in P0-011 |
| I5 — operator CLI reachable | **partial** | P0-002, P0-004, P0-006–P0-011 |
| I6 — owned deferrals | holds once tracker exists | P0-005 |

## What comes next

Phase 1 will cover the remaining design-set gaps: browser login and session
cookies, credentials and workspaces, structured events, a second harness
provider, and the roadmap exit criteria not satisfied by the fake-profile run
round trip. It will be written when Phase 0 closes or when explicitly requested.
