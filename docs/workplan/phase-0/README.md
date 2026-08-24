# Phase 0 — Foundation repair

**Status:** derived. Its tasks are in [`../tasks.json`](../tasks.json).

## Objective

The product completes a run round trip. A person creates a run through the
daemon API or the operator CLI, starts it, watches it live, sends it input,
stops it, and reads a truthful terminal state back — against the fake harness
under the local sandbox and the in-memory store — and CI refuses any change that
breaks that path.

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
- **Memory store repair** — the run round trip was derived against PostgreSQL,
  skipping I3 and I4's first rung. P0-012 ships `storeProfile=memory`; P0-013
  through P0-018 revise each seam and the CI gate so `winchd` and `make e2e` run
  without a database (see the post-mortem's defect 5).
- **Closure** — the assembled round trip runs as one scenario and gates CI
  (P0-007, then P0-018 for the memory-backed gate); contributor documentation
  stops describing a plan that no longer exists (P0-005).

## Non-goals

- Browser login and session establishment, credentials, and workspaces — see
  [`../phase-1/README.md`](../phase-1/README.md).
- Structured event families, renderers, and a second harness provider
  (`docs/roadmap.md` Phase 2).
- Container isolation, workflows, and remote runners (Phases 3 to 5).
- PostgreSQL as the default e2e substrate — that is an I4 swap task in phase 1.

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
| P0-012 | Ship controllable in-memory store profile | — | capability | `winchd` with `storeProfile=memory` |
| P0-013 | Revise create/read for memory store profile | P0-006, P0-012 | swap | `make e2e` create/get without DB |
| P0-014 | Revise start execution for memory store profile | P0-008, P0-013 | swap | start scenario without DB |
| P0-015 | Revise input and outbox for memory store profile | P0-009, P0-014 | swap | input scenario without DB |
| P0-016 | Revise WebSocket stream for memory store profile | P0-010, P0-015 | swap | stream scenario without DB |
| P0-017 | Revise stop run for memory store profile | P0-011, P0-014 | swap | stop scenario without DB |
| P0-018 | Revise phase 0 closure for memory-backed e2e | P0-007, P0-016, P0-017, P0-012 | hardening | `make e2e` in CI without DB |

### Dependency graph

```
Independent at open:
  P0-001   P0-002   P0-003   P0-005   P0-012

P0-002 + P0-003 ──► P0-004

P0-001 ──► P0-006 ──┬──► P0-008 ──┬──► P0-009 ──► P0-010 ──┐
                    │             │                        │
P0-003 ──────────────┘             └──► P0-011 ─────────────┤
P0-001 ────────────────────────────────────────────────────►┴──► P0-007

Memory repair (revision edges; each waits on the postgres seam it revises):
P0-012 ──┬──► P0-013 ◄── P0-006
         │       │
         │       └──► P0-014 ◄── P0-008 ──┬──► P0-015 ◄── P0-009 ──► P0-016 ◄── P0-010
         │                                │
         │                                └──► P0-017 ◄── P0-011
         │
P0-007 + P0-016 + P0-017 ──► P0-018
```

- **P0-004** waits on the profile it exercises (P0-003) and on the CLI it
  invokes (P0-002).
- **P0-006 → P0-001** is semantic in the CI sense for the postgres profile only;
  P0-013 removes that requirement from the e2e path once it lands.
- **P0-011 → P0-008** only. Stop needs a running execution, not input or a
  WebSocket.
- **P0-007** asserts the assembled round trip on postgres; **P0-018** revises
  that gate to the all-fake memory profile.
- **P0-012** is available at open and can land before the postgres seams, but
  P0-013 through P0-018 carry `revision` edges and wait on P0-006 through
  P0-011 respectively.

### The e2e suite

Each postgres seam task owns one scenario file. The memory repair tasks revise
those files to run without PostgreSQL. P0-007 adds the assembled postgres
scenario; P0-018 revises it for memory.

| Task | Scenario | File |
|---|---|---|
| P0-006 / P0-013 | `create → get` | `test/e2e/create_test.go` |
| P0-008 / P0-014 | `create → start → poll events → get` | `test/e2e/start_test.go` |
| P0-009 / P0-015 | `create → start → input → poll events` | `test/e2e/input_test.go` |
| P0-010 / P0-016 | `create → start → stream` | `test/e2e/stream_test.go` |
| P0-011 / P0-017 | `create → start → stop → get` | `test/e2e/stop_test.go` |
| P0-007 / P0-018 | `create → start → stream → input → stop` | `test/e2e/roundtrip_test.go` |

### Width

| Metric | Value |
|---|---|
| Critical path | 10 — `P0-001 → P0-006 → P0-008 → P0-009 → P0-010 → P0-013 → P0-014 → P0-015 → P0-016 → P0-018` |
| Average width | 18 ÷ 10 ≈ 1.8 |
| Available at open | P0-001, P0-002, P0-003, P0-005, P0-012 |
| Contract collisions | none between concurrently-available tasks |

Write collisions — a cost, not an edge. Whoever takes the second one rebases.

| Pair | Files |
|---|---|
| P0-001 ↔ P0-002 | `Makefile`, `deployments/README.md` |
| P0-001 ↔ P0-003 | `deployments/README.md` (if P0-001 changes the testing procedure) |
| P0-001 ↔ P0-004 | `Makefile`, `.github/workflows/go.yml` |
| P0-001 ↔ P0-018 | `.github/workflows/go.yml` |
| P0-002 ↔ P0-003 | `deployments/README.md` |
| P0-002 ↔ P0-006 | `Makefile` |
| P0-002 ↔ P0-007 | `deployments/README.md` |
| P0-003 ↔ P0-006 | `cmd/winch/main.go` |
| P0-004 ↔ P0-006 | `Makefile` |
| P0-004 ↔ P0-007 | `.github/workflows/go.yml` |
| P0-006 ↔ P0-012 | `cmd/winchd/main.go` |
| P0-009 ↔ P0-011 | `cmd/winchd/main.go`, `internal/application/`, `cmd/winch/` |
| P0-013 ↔ P0-014 | `cmd/winchd/main.go`, `test/e2e/` |
| P0-007 ↔ P0-018 | `test/e2e/roundtrip_test.go`, `.github/workflows/go.yml` |

`cmd/winchd/main.go` attracts both the postgres seams and the memory repairs, but
revision edges serialize the repairs after their targets. `docs/code-structure.md:119`
keeps adapter registration explicit in the composition root.

## Deferrals in

None. No other phase defers work into phase 0.

## Phase exit

Every task above is `completed`, `make e2e` runs the full
`create → start → stream → input → stop` scenario on `storeProfile=memory` on
every push and pull request without a PostgreSQL service, and
`make test-integration` still gates the postgres adapter.

## Invariants

| Invariant | At open | Closed by |
|---|---|---|
| I1 — system runs | holds (daemon starts) | kept by every task |
| I2 — system deploys | holds (compose stack) | P0-002 |
| I3 — controllable fake profile | **fails** | P0-003 (harness); P0-012 (storage) |
| I4 — standing scenario suite | **fails** | P0-006, P0-008 to P0-011 contribute; P0-007 on postgres; P0-018 closes memory first rung |
| I5 — operator CLI reachable | **partial** | P0-002, then each seam adds its own command |
| I6 — owned deferrals | — | P0-005; phase deferrals are registered in the owning phase brief |
| I7 — plan stays wide | — | width recorded above; recheck on every extend |

The postgres seams (P0-006 through P0-011) were derived before the memory
profile existed — recorded as defect 5 in
[`../post-mortems/2026-08-23-workplan-create-phase-0.md`](../post-mortems/2026-08-23-workplan-create-phase-0.md).
P0-012 through P0-018 repair that without changing the launched briefs.

## Traces to

- [`docs/state.md`](../../state.md) — *What went wrong*, *Invariants that were
  never established*, *Gates that do not cover what they appear to*
- `docs/roadmap.md` Phase 1 exit (forced stop leaves no child processes)
- `docs/architecture.md` §4 (application services, run supervisor)
