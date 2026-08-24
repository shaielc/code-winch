# P0-016: Revise WebSocket stream for memory store profile

**Phase:** 0 — Foundation repair
**Shape:** swap
**Dependencies:** P0-010 (revision: P0-010 wired EventStream publish through postgres-backed execution), P0-015 (semantic: outbox drain on the memory profile must exist before live publish is meaningful)

## Objective

The live WebSocket event stream works with `storeProfile=memory`; the
`create → start → stream` e2e scenario passes without PostgreSQL.

## Scope

- Revise daemon wiring from P0-010 so `EventStream.Publish` is driven from the
  memory-profile outbox path when selected.
- Revise `test/e2e/stream_test.go` to run against the memory-backed harness.
- Add or revise `winch run stream` documentation if it still assumes SQL
  inspection.

## Non-goals

- `StopRun` — P0-017.
- The assembled round-trip scenario or CI gate — P0-018.
- Browser UI or session cookies.

## Runtime reachability

- **Composition root:** `cmd/winchd` (`httpapi.EventStream`).
- **Profile:** `storeProfile=memory`, fake harness, local sandbox.
- **Command:** `winch run stream`, `make e2e` (stream scenario).

## Write set

- `cmd/winchd/main.go`
- `internal/application/` (only if publish wiring is profile-branching)
- `test/e2e/stream_test.go`
- Tests for WebSocket delivery on the memory profile

## Contract surfaces

- API: `GET /api/v1/runs/{runId}/events/stream` (WebSocket)

## Demonstration

    $ make e2e -run TestStream
    → expect: passes with no PostgreSQL process running

    $ winch run create --workspace /tmp/ws --harness fake --sandbox local
    $ winch run start $RUN_ID
    $ winch run stream $RUN_ID
    → expect: live events in order, gap-free sequence numbers

## Verification

- `make check` passes.
- `make e2e` passes locally through the stream scenario without a database.

## Acceptance criteria

- [ ] WebSocket subscribers receive events for runs on `storeProfile=memory`.
- [ ] `test/e2e/stream_test.go` runs without PostgreSQL setup.
- [ ] Polled and streamed events agree on the memory profile.
- [ ] I1 and I2 still hold.

## Deferrals

| Deferred | Owning task |
|---|---|
| Assembled round trip and CI on memory profile | P0-018 |
| Stop on memory profile | P0-017 |

## Traces to

- Invariant I4 — `skills/shared/workplan-model.md`
- [`../post-mortems/2026-08-23-workplan-create-phase-0.md`](../post-mortems/2026-08-23-workplan-create-phase-0.md)
  defect 5
- P0-010 — postgres storage declaration this task revises
