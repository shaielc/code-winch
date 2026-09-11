# P0-017: Revise stop run for memory store profile

**Phase:** 0 — Foundation repair
**Shape:** swap
**Dependencies:** P0-011 (revision: P0-011 declared PostgreSQL storage and wired stop through postgres-backed supervisor state), P0-014 (semantic: a running execution path on the memory profile must exist)

## Objective

Force-stop works with `storeProfile=memory`; the `create → start → stop → get`
e2e scenario passes without PostgreSQL.

## Scope

- Revise daemon wiring from P0-011 so `StopRun` persists terminal state through
  the memory profile's repositories when selected.
- Revise `test/e2e/stop_test.go` to run against the memory-backed harness.
- Keep child-process reaping assertions from P0-011 unchanged.

## Non-goals

- The assembled round-trip scenario or CI gate — P0-018.
- Browser UI or session cookies.
- Daemon restart reconciliation demonstration.

## Runtime reachability

- **Composition root:** `cmd/winchd`.
- **Profile:** `storeProfile=memory`, fake harness, local sandbox.
- **Command:** `winch run stop`, `make e2e` (stop scenario).

## Write set

- `cmd/winchd/main.go`
- `test/e2e/stop_test.go`
- Tests for stop escalation and child reaping on the memory profile

## Contract surfaces

- API: `POST /api/v1/runs/{runId}/stop`

## Demonstration

    $ make e2e -run TestStop
    → expect: passes with no PostgreSQL process running

    $ winch run create --workspace /tmp/ws --harness fake --sandbox local
    $ winch run start $RUN_ID
    $ winch run stop $RUN_ID
    $ winch run get $RUN_ID
    → expect: terminal state persisted; no surviving `fake-harness` process

## Verification

- `make check` passes.
- `make e2e` passes locally through the stop scenario without a database.

## Acceptance criteria

- [ ] `POST /api/v1/runs/{runId}/stop` works with `storeProfile=memory`.
- [ ] `test/e2e/stop_test.go` runs without PostgreSQL setup.
- [ ] Terminal state and process reaping match P0-011's claims on the memory
      profile.
- [ ] I1 and I2 still hold.

## Deferrals

| Deferred | Owning task |
|---|---|
| Assembled round trip and CI on memory profile | P0-018 |
| Browser `winch_session` establishment | Phase 1 — registered in [`../phase-1/README.md`](../phase-1/README.md) |

## Traces to

- Invariant I4 — `skills/shared/workplan-model.md`
- `docs/roadmap.md` Phase 1 exit ("forced stop leaves no child processes")
- [`../post-mortems/2026-08-23-workplan-create-phase-0.md`](../post-mortems/2026-08-23-workplan-create-phase-0.md)
  defect 5
- P0-011 — postgres storage declaration this task revises
