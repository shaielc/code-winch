# P0-008: Start run execution

**Phase:** 0 — Foundation repair
**Shape:** seam
**Dependencies:** P0-006 (compile: delegating backend and store exist), P0-003 (contract: `harnessProfile=fake` launch surface)

## Objective

A person can start a created run; the fake harness executes under the local
sandbox, events are durably stored, and output is visible by polling HTTP — with
outbox rows enqueued but not yet delivered.

## Scope

- Construct the `fake` harness and `local` sandbox drivers directly in
  `cmd/winchd`. Phase 0 supports exactly this pair, so no name-to-driver
  registry is built (`docs/code-structure.md:119` keeps registration explicit in
  the composition root either way).
- Validate the run's persisted `harnessProfile`/`sandboxProfile` against that
  pair and refuse anything else with a stable, content-free error, so an
  unsupported request is rejected rather than silently run unisolated
  (ADR-0003).
- Construct `supervisor.Supervisor` and `runner/local.Runner`, wiring runner
  observations to `Supervisor.Observe` and durable `EventStore` append (which also
  enqueues outbox records per the existing postgres adapter).
- Implement `StartRun` and `ListRunEvents` on the delegating backend.
- Add `winch run start` and `winch run events` (HTTP poll; no WebSocket client).
- Add the **`create → start → poll events → get`** scenario in
  `test/e2e/start_test.go`: start a created run, let the fake harness exit on
  its own, assert ordered, gap-free events from the polling endpoint, and read
  the run back to assert its terminal state.

## Non-goals

- A driver registry, name-to-driver resolution, or any second harness or
  sandbox. Phase 0 ships one supported pair and says so.
- **Any e2e step beyond `create → start → poll events → get`.** No input, no
  WebSocket, and no stop command — the scenario's run ends because its
  transcript ends.
- Starting the outbox worker or any `OutboxPublisher` implementation.
- `EventStream.Publish` or WebSocket streaming.
- `SendRunInput`, `StopRun`, or outbox backlog draining.
- Daemon restart reconciliation demonstration.

## Runtime reachability

- **Composition root:** `cmd/winchd`.
- **Profile:** `harnessProfile=fake`, `sandboxProfile=local`, PostgreSQL storage.
- **Command:** `winch run start`, `winch run events`.

## Write set

- `cmd/winchd/main.go` (registry registration only — no wholesale rewrite)
- `internal/application/` (start use case)
- `cmd/winch/` (`start`, `events` subcommands)
- `test/e2e/start_test.go`
- Tests for start transitions and durable event append

## Contract surfaces

- API: `POST /api/v1/runs/{runId}/start`, `GET /api/v1/runs/{runId}/events`
- driver namespace: `harnessProfile=fake`, `sandboxProfile=local`

## Demonstration

    $ winch run create --workspace /tmp/ws --harness fake --sandbox local
    $ winch run start $RUN_ID
    → expect: run reaches `running` then a terminal state; `fake-harness` output
      appears while running

    $ winch run events $RUN_ID
    → expect: ordered events including harness stream output

    $ ps -eo pid,cmd | grep -c 'fake[-]harness'
    → expect: 0 after the run completes on its own (no stop command yet)

Outbox backlog (SQL or a debug metric if exposed) remains **non-zero** — publish
intent exists but no worker drains it yet.

## Verification

- `make check` and `make test-integration` pass.
- `make e2e` passes through create, get, start, and polled events.
- Contract suites for fake harness and local sandbox still pass.

## Acceptance criteria

- [ ] `StartRun` drives a fake harness to completion through supervisor and
  local runner.
- [ ] `ListRunEvents` returns gap-free, ordered events from PostgreSQL.
- [ ] Starting a run whose persisted profiles are not `fake`/`local` is refused
  with a stable error code, and no harness process is launched.
- [ ] Outbox rows are created on append but no worker runs in the daemon process.
- [ ] `winch run start` and `winch run events` demonstrate the behavior.
- [ ] I1 and I2 still hold.

## Deferrals

| Deferred | Owning task |
|---|---|
| Outbox worker and delivery | P0-009 |
| Input | P0-009 |
| Live WebSocket stream | P0-010 |
| Stop | P0-011 |

## Traces to

- `docs/architecture.md` §4 (run supervisor, harness adapters)
- `docs/state.md` §*What went wrong* (supervisor and runner unreachable from
  composition root)
