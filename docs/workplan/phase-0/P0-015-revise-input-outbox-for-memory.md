# P0-015: Revise input and outbox for memory store profile

**Phase:** 0 — Foundation repair
**Shape:** swap
**Dependencies:** P0-009 (revision: P0-009 wired outbox worker and input against postgres storage), P0-014 (semantic: a running execution path on the memory profile must exist)

## Objective

Send input and drain the outbox with `storeProfile=memory`; the
`create → start → input → poll events` e2e scenario passes without PostgreSQL.

## Scope

- Revise daemon lifecycle wiring from P0-009 so `OutboxWorker`, input command
  persistence, and the no-op/metrics `OutboxPublisher` use memory repositories
  when the memory profile is selected.
- Revise `test/e2e/input_test.go` to run against the memory-backed harness.
- Keep polling as the delivery demonstration surface; no WebSocket changes.

## Non-goals

- `EventStream.Publish` or WebSocket subscribers — P0-016.
- `StopRun` — P0-017.
- Proving live browser delivery.

## Runtime reachability

- **Composition root:** `cmd/winchd` (outbox worker loop).
- **Profile:** `storeProfile=memory`, fake harness, local sandbox.
- **Command:** `winch run input`, `make e2e` (input scenario).

## Write set

- `cmd/winchd/main.go`
- `test/e2e/input_test.go`
- Tests for input acceptance and outbox drain on the memory profile

## Contract surfaces

None.

## Demonstration

    $ make e2e -run TestInput
    → expect: passes with no PostgreSQL process running

    $ winch run create --workspace /tmp/ws --harness fake --sandbox local
    $ winch run start $RUN_ID
    $ winch run input $RUN_ID --text hello
    $ winch run events $RUN_ID
    → expect: harness response in events; outbox backlog drains to zero

## Verification

- `make check` passes.
- `make e2e` passes locally through the input scenario without a database.

## Acceptance criteria

- [ ] `POST /api/v1/runs/{runId}/input` works with `storeProfile=memory`.
- [ ] Outbox worker starts and drains on daemon shutdown within the memory
      profile.
- [ ] `test/e2e/input_test.go` runs without PostgreSQL setup.
- [ ] I1 and I2 still hold.

## Deferrals

| Deferred | Owning task |
|---|---|
| Live WebSocket stream on memory profile | P0-016 |
| Stop on memory profile | P0-017 |

## Traces to

- Invariant I4 — `skills/shared/workplan-model.md`
- [`../post-mortems/2026-08-23-workplan-create-phase-0.md`](../post-mortems/2026-08-23-workplan-create-phase-0.md)
  defect 5
- P0-009 — postgres storage declaration this task revises
