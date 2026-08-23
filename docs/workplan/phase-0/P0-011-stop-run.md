# P0-011: Stop run

**Phase:** 0 — Foundation repair
**Shape:** seam
**Dependencies:** P0-008 (semantic: a running execution path must exist before stop is meaningful)

## Objective

A person can force-stop a running harness; the run reaches a terminal state and
leaves no child processes.

## Scope

- Implement `StopRun` through the supervisor and runner stop path.
- Add `winch run stop`.
- Add the **`create → start → stop → get`** scenario in `test/e2e/stop_test.go`:
  start a run whose transcript runs until stopped, stop it, read the run back to
  assert a persisted terminal state, and assert no surviving `fake-harness`
  process.

## Non-goals

- **Any e2e step beyond `create → start → stop → get`.** No input and no
  WebSocket steps; assembling the complete round trip into one scenario is
  P0-007.
- Wiring `make e2e` into CI, and describing the run routes as fully bound in
  `deployments/README.md` — both are P0-007, which is the first point at which
  either is true.
- Browser UI or session cookies.
- Daemon restart reconciliation demonstration (covered by later hardening).
- New CLI subcommands beyond stop.

## Runtime reachability

- **Composition root:** `cmd/winchd`.
- **Profile:** fake harness, local sandbox, PostgreSQL storage.
- **Command:** `winch run stop`, `make e2e`.

## Write set

- `cmd/winchd/main.go` (binding `StopRun` on the delegating backend)
- `internal/application/` (stop use case)
- `cmd/winch/` (`stop` subcommand)
- `test/e2e/stop_test.go`
- Tests for stop escalation and child reaping

## Contract surfaces

- API: `POST /api/v1/runs/{runId}/stop`

## Demonstration

Use a P0-003 transcript that runs until stopped:

    $ winch run create ...
    $ winch run start $RUN_ID
    $ winch run stop $RUN_ID
    → expect: run reaches `cancelled` or `completed` per policy; harness exits

    $ ps -eo pid,cmd | grep -c 'fake[-]harness'
    → expect: 0

    $ make e2e
    → expect: the `create → start → stop → get` scenario passes, alongside the
      scenarios landed so far

## Verification

- `make check` and `make test-integration` pass.
- `make e2e` passes locally, including the new stop scenario.

## Acceptance criteria

- [ ] `StopRun` forces harness termination and persists a terminal state.
- [ ] No `fake-harness` descendant remains after stop.
- [ ] `winch run stop` demonstrates the behavior.
- [ ] `make e2e` runs the `create → start → stop → get` scenario locally.
- [ ] I1 and I2 still hold.

## Deferrals

| Deferred | Owning task |
|---|---|
| Assembling the complete round trip and gating `make e2e` in CI | P0-007 |
| Browser `winch_session` establishment | Phase 1 — registered in [`../phase-1/README.md`](../phase-1/README.md) |

## Traces to

- `docs/roadmap.md` Phase 1 exit ("forced stop leaves no child processes")
- `docs/state.md` §*The run round trip*
- Invariant I4 — this task contributes the stop scenario; P0-007 closes I4
