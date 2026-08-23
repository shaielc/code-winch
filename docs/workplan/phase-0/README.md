# Phase 0 — Foundation repair

**Status:** derived. Its tasks are in [`../tasks.json`](../tasks.json).

## Objective

The product completes a run round trip. A person creates a run through the
daemon API or the operator CLI, starts it, watches it live, sends it input,
stops it, and reads a truthful terminal state back — against the fake harness
under the local sandbox — and CI refuses any change that breaks that path.

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

P0-001 ──► P0-006 ──┐
                    ├──► P0-008 ──┬──► P0-009 ──► P0-010 ──┐
P0-003 ──────────────┘             │                        │
                                   └──► P0-011 ─────────────┤
P0-001 ────────────────────────────────────────────────────►┴──► P0-007
```

- **P0-004** waits on the profile it exercises (P0-003) and on the CLI it
  invokes (P0-002).
- **P0-006 → P0-001** is semantic in the CI sense: P0-006's claim is that a run
  is *durably* persisted, and nothing proves that on a pull request until the Go
  workflow has a database. Against an in-memory store profile this edge would
  not exist; see the post-mortem's defect 5.
- **P0-011 → P0-008** only. Stop needs a running execution, not input or a
  WebSocket, so it proceeds in parallel with the P0-009 → P0-010 chain.
- **P0-007** is last by construction: it asserts the assembled round trip, so it
  waits on every operation, and it wires `make e2e` into CI, so it waits on the
  database P0-001 puts there.

### The e2e suite

Each task owns one scenario file and states in its non-goals which steps it does
not add. P0-007 adds a sixth file that drives one run through all five
operations and may not edit the others.

| Task | Scenario | File |
|---|---|---|
| P0-006 | `create → get` | `test/e2e/create_test.go` (plus the shared harness) |
| P0-008 | `create → start → poll events → get` | `test/e2e/start_test.go` |
| P0-009 | `create → start → input → poll events` | `test/e2e/input_test.go` |
| P0-010 | `create → start → stream` | `test/e2e/stream_test.go` |
| P0-011 | `create → start → stop → get` | `test/e2e/stop_test.go` |
| P0-007 | `create → start → stream → input → stop` | `test/e2e/roundtrip_test.go` |

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

`cmd/winchd/main.go` attracts P0-006, P0-008, P0-009, P0-010, and P0-011, but
only P0-009 and P0-011 are ever available at the same time; the rest are a
chain. `docs/code-structure.md:119` keeps adapter registration explicit in the
composition root, so this contention is a property of the design set and cannot
be architected away — only kept small, one binding per task.

## Deferrals in

None. No other phase defers work into phase 0.

## Phase exit

Every task above is `completed` and `make e2e` runs the full
`create → start → stream → input → stop` scenario on every push and pull
request.

## Invariants

| Invariant | At open | Closed by |
|---|---|---|
| I1 — system runs | holds (daemon starts) | kept by every task |
| I2 — system deploys | holds (compose stack) | P0-002 |
| I3 — controllable fake profile | **fails** | P0-003 (harness only; storage is deferred to phase 1) |
| I4 — standing scenario suite | **fails** | P0-006, P0-008 to P0-011 contribute; P0-007 closes |
| I5 — operator CLI reachable | **partial** | P0-002, then each seam adds its own command |
| I6 — owned deferrals | — | P0-005; phase deferrals are registered in the owning phase brief |
| I7 — plan stays wide | — | width recorded above; recheck on every extend |

I3 closes only for the harness. The in-memory store profile is deferred to phase
1, which means no phase-0 configuration runs without PostgreSQL — recorded as
defect 5 in
[`../post-mortems/2026-08-23-workplan-create-phase-0.md`](../post-mortems/2026-08-23-workplan-create-phase-0.md).

## Traces to

- [`docs/state.md`](../../state.md) — *What went wrong*, *Invariants that were
  never established*, *Gates that do not cover what they appear to*
- `docs/roadmap.md` Phase 1 exit (forced stop leaves no child processes)
- `docs/architecture.md` §4 (application services, run supervisor)
