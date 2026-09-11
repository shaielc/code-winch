# Phase 0 — Foundation repair

**Status:** derived. Its tasks are in [`../tasks.json`](../tasks.json).

## Objective

The product completes a run round trip. A person creates a run through the
daemon API or the operator CLI, starts it, watches it live, sends it input,
stops it, and reads a truthful terminal state back — against the fake harness
under the local sandbox, with PostgreSQL as the store — and CI refuses any
change that breaks that path.

## Scope

Phase 0 repairs what [`docs/state.md`](../../state.md) records under *What went
wrong*: invariants the previous plan set down and never built on, gates that do
not cover what they appear to, and the use-case layer that was never written.

- **Gates and shipping** — CI proves storage guarantees; the unit-test gate stops
  walking `node_modules`; the operator CLI is built and installed in the
  deployment image (P0-001, P0-002).
- **The fake profile** — the fake harness becomes controllable at runtime, and a
  complete fake run is exercised in CI against the shipped binaries (P0-003,
  P0-004).
- **The run round trip** — one seam per operator operation, each binding its own
  method on the delegating `httpapi.Backend`, each adding its own CLI command,
  each contributing its own e2e scenario (P0-006, P0-008 to P0-011).
- **Closure** — the assembled round trip runs as one scenario and gates CI
  (P0-007); contributor documentation stops describing a plan that no longer
  exists (P0-005).

## Non-goals

- Browser login and session establishment, credentials, and workspaces — see
  [`../phase-1/README.md`](../phase-1/README.md).
- Structured event families, renderers, and a second harness provider
  (`docs/roadmap.md` Phase 2).
- Container isolation, workflows, and remote runners (Phases 3 to 5).
- **A runnable in-memory store profile.** Phase 0 runs on PostgreSQL
  throughout, so I3 does not hold for storage at phase exit and I4's first rung
  is the postgres scenario rather than a fake one. The gap is registered in
  [`../phase-1/README.md`](../phase-1/README.md) *Deferrals in* and explained in
  [`../post-mortems/2026-09-11-a-store-profile-derived-from-test-doubles.md`](../post-mortems/2026-09-11-a-store-profile-derived-from-test-doubles.md).

## Tasks

| ID | Title | Depends on | Shape | CLI / e2e |
|---|---|---|---|---|
| P0-001 | Repair CI test gates | — | hardening | — |
| P0-002 | Ship operator CLI in build and deployment | — | hardening | — |
| P0-003 | Controllable fake harness profile | — | capability | flags on `dev run` |
| P0-004 | Exercise complete fake run in CI | P0-002, P0-003 | hardening | `dev run` in CI |
| P0-005 | Sanitize post-close plan references | — | hardening | — |
| P0-006 | Create and read runs | P0-001 | seam | `run create/get`; starts `make e2e` |
| P0-007 | Finalize phase 0 | P0-001, P0-009, P0-010, P0-011 | hardening | `make e2e` in CI |
| P0-008 | Start run execution | P0-003, P0-006 | seam | `run start/events` |
| P0-009 | Send input and drain outbox | P0-008 | seam | `run input` |
| P0-010 | Live WebSocket event stream | P0-009 | seam | `run stream` |
| P0-011 | Stop run | P0-008 | seam | `run stop` |

### Dependency graph

```
Independent at open:
  P0-001   P0-002   P0-003   P0-005

P0-002 + P0-003 ──► P0-004

P0-001 ──► P0-006 ──┬──► P0-008 ──┬──► P0-009 ──► P0-010 ──┐
                    │             │                        │
P0-003 ──────────────┘             └──► P0-011 ─────────────┤
P0-001 ────────────────────────────────────────────────────►┴──► P0-007
```

- **P0-004** waits on the profile it exercises (P0-003) and on the CLI it
  invokes (P0-002).
- **P0-006 → P0-001** is semantic in the CI sense: the e2e path needs a
  PostgreSQL service the repaired gate provides. Every scenario in this phase
  inherits that requirement, because every scenario runs on postgres.
- **P0-011 → P0-008** only. Stop needs a running execution, not input or a
  WebSocket.
- **P0-007** asserts the assembled round trip on postgres, and is the phase's
  last task.

### The e2e suite

Each seam task owns one scenario file, and every scenario runs against
PostgreSQL. P0-007 adds the assembled one.

| Task | Scenario | File |
|---|---|---|
| P0-006 | `create → get` | `test/e2e/create_test.go` |
| P0-008 | `create → start → poll events → get` | `test/e2e/start_test.go` |
| P0-009 | `create → start → input → poll events` | `test/e2e/input_test.go` |
| P0-010 | `create → start → stream` | `test/e2e/stream_test.go` |
| P0-011 | `create → start → stop → get` | `test/e2e/stop_test.go` |
| P0-007 | `create → start → stream → input → stop` | `test/e2e/roundtrip_test.go` |

Every file here requires `PG_TEST_DATABASE_URL` and skips without it, so the
suite asserts nothing on a machine with no database. Making it runnable against
a fake profile is I4's first rung, which this phase does not build.

### Width

| Metric | Value |
|---|---|
| Critical path | 6 — `P0-001 → P0-006 → P0-008 → P0-009 → P0-010 → P0-007` |
| Average width | 11 ÷ 6 ≈ 1.8 |
| Available at open | P0-001, P0-002, P0-003, P0-005 |
| Contract collisions | none between concurrently-available tasks |

Write collisions — a cost, not an edge. Whoever takes the second one rebases.

| Pair | Files |
|---|---|
| P0-001 ↔ P0-002 | `Makefile`, `deployments/README.md` |
| P0-001 ↔ P0-003 | `deployments/README.md` (if P0-001 changes the testing procedure) |
| P0-001 ↔ P0-004 | `Makefile`, `.github/workflows/go.yml` |
| P0-002 ↔ P0-003 | `deployments/README.md` |
| P0-002 ↔ P0-006 | `Makefile` |
| P0-002 ↔ P0-007 | `deployments/README.md` |
| P0-003 ↔ P0-006 | `cmd/winch/main.go` |
| P0-004 ↔ P0-006 | `Makefile` |
| P0-004 ↔ P0-007 | `.github/workflows/go.yml` |
| P0-009 ↔ P0-011 | `cmd/winchd/main.go`, `internal/application/`, `cmd/winch/` |

`cmd/winchd/main.go` attracts every seam, since each binds its own method on the
delegating backend. `docs/code-structure.md:119` keeps adapter registration
explicit in the composition root.

## Deferrals in

None. No other phase defers work into phase 0.

## Phase exit

Every task above is `completed`, `make e2e` runs the full
`create → start → stream → input → stop` scenario against PostgreSQL on every
push and pull request, and `make test-integration` still gates the postgres
adapter.

The phase exits with I3 unmet for storage and I4's first rung unbuilt. That is a
stated shortfall, not a silent one: both are registered as deferrals into phase
1.

## Invariants

| Invariant | At open | Closed by |
|---|---|---|
| I1 — system runs | holds (daemon starts) | kept by every task |
| I2 — system deploys | holds (compose stack) | P0-002 |
| I3 — controllable fake profile | **fails** | P0-003 (harness) only; **storage is not closed in this phase** — deferred to phase 1 |
| I4 — standing scenario suite | **fails** | P0-006, P0-008 to P0-011 contribute; P0-007 assembles it on postgres; **the all-fake first rung is not built** — deferred to phase 1 |
| I5 — operator CLI reachable | **partial** | P0-002, then each seam adds its own command |
| I6 — owned deferrals | — | P0-005; phase deferrals are registered in the owning phase brief |
| I7 — plan stays wide | — | width recorded above; recheck on every extend |

A first attempt to close I3 for storage inside this phase — P0-012 through
P0-018, a store-profile task plus one revision task per seam — was withdrawn
after an audit found the store profile had been derived from the shape of
`internal/adapters/memory`'s test doubles rather than from the ports the run path
consumes. See
[`../post-mortems/2026-09-11-a-store-profile-derived-from-test-doubles.md`](../post-mortems/2026-09-11-a-store-profile-derived-from-test-doubles.md).
The work is real and still wanted; it is deferred into phase 1 for a fresh
derivation rather than left in the plan in a form that cannot be built.

## Traces to

- [`docs/state.md`](../../state.md) — *What went wrong*, *Invariants that were
  never established*, *Gates that do not cover what they appear to*
- `docs/roadmap.md` Phase 1 exit (forced stop leaves no child processes)
- `docs/architecture.md` §4 (application services, run supervisor)
