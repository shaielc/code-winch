# P0-014: Revise start execution for memory store profile

**Phase:** 0 — Foundation repair
**Shape:** swap
**Dependencies:** P0-008 (revision: P0-008 declared PostgreSQL storage under *Runtime reachability* and wired supervisor/event append through postgres), P0-013 (semantic: e2e harness defaults to memory and create/read must pass first)

## Objective

Start run execution and event polling work with `storeProfile=memory`; the
`create → start → poll events → get` e2e scenario passes without PostgreSQL.

## Scope

- Revise daemon wiring from P0-008 so `SupervisorStore`, `EventStore`, and
  outbox enqueue use the memory profile's repositories when selected.
- Revise `test/e2e/start_test.go` to run against the memory-backed harness
  (no database fixture).
- Keep the `fake`/`local` harness pair and profile validation from P0-008
  unchanged.
- Ensure demonstration and verification commands name memory storage, not SQL.

## Non-goals

- Input, outbox worker delivery, WebSocket stream, or stop — P0-015 through
  P0-017.
- A driver registry or second harness/sandbox pair.
- Daemon restart reconciliation demonstration.

## Runtime reachability

- **Composition root:** `cmd/winchd`.
- **Profile:** `storeProfile=memory`, `harnessProfile=fake`, `sandboxProfile=local`.
- **Command:** `winch run start`, `winch run events`, `make e2e` (start scenario).

## Write set

- `cmd/winchd/main.go`
- `test/e2e/start_test.go`
- Tests for start transitions and durable event append on the memory profile

## Contract surfaces

None.

## Demonstration

    $ make e2e -run TestStart
    → expect: passes with no PostgreSQL process running

    $ winch run create --workspace /tmp/ws --harness fake --sandbox local
    $ winch run start $RUN_ID
    $ winch run events $RUN_ID
    → expect: ordered events including harness output; run reaches a terminal state

## Verification

- `make check` passes.
- `make e2e` passes locally through the start scenario without a database.

## Acceptance criteria

- [ ] `POST /api/v1/runs/{runId}/start` and `GET /api/v1/runs/{runId}/events`
      work with `storeProfile=memory`.
- [ ] `test/e2e/start_test.go` runs without PostgreSQL setup.
- [ ] Outbox backlog behavior from P0-008 is observable on the memory profile.
- [ ] I1 and I2 still hold.

## Deferrals

| Deferred | Owning task |
|---|---|
| Input and outbox drain on memory profile | P0-015 |
| Live WebSocket stream on memory profile | P0-016 |
| Stop on memory profile | P0-017 |

## Traces to

- Invariant I4 — `skills/shared/workplan-model.md`
- [`../post-mortems/2026-08-23-workplan-create-phase-0.md`](../post-mortems/2026-08-23-workplan-create-phase-0.md)
  defect 5
- P0-008 — postgres storage declaration this task revises
