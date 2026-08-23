# P0-009: Send input and drain outbox

**Phase:** 0 — Foundation repair
**Shape:** seam
**Dependencies:** P0-008 (semantic: a running execution path must exist before input is meaningful)

## Objective

A person can send text input to a running harness; the outbox worker drains
publish intent with polling as the demonstration surface — no WebSocket yet.

## Scope

- Implement `SendRunInput` through the existing `InputService` and supervisor
  path.
- Start `application.OutboxWorker` in the daemon lifecycle (start loop on boot,
  stop on shutdown within the existing drain budget).
- Register an `OutboxPublisher` that **completes delivery without calling
  `EventStream.Publish`** — for example a no-op or metrics-only publisher whose
  sole job is to let the worker mark outbox rows complete.
- Add `winch run input`.
- Extend the standing `make e2e` scenario with input and an outbox-backlog
  drain assertion.

## Non-goals

- `EventStream.Publish` or WebSocket subscribers.
- `winch run stream` or any WebSocket client.
- `StopRun` (P0-011).
- Proving live browser delivery.

## Runtime reachability

- **Composition root:** `cmd/winchd` (outbox worker loop).
- **Profile:** fake harness, local sandbox, PostgreSQL storage.
- **Command:** `winch run input`.

## Write set

- `cmd/winchd/main.go` (worker lifecycle)
- `internal/application/` (input use case, publisher adapter)
- `cmd/winch/` (`input` subcommand)
- Tests for input acceptance and outbox drain

## Contract surfaces

- API: `POST /api/v1/runs/{runId}/input`
- port: `OutboxPublisher` (non-streaming implementation)

## Demonstration

Start a run that waits for input (use a P0-003 transcript that blocks until
input):

    $ winch run create ...
    $ winch run start $RUN_ID
    $ winch run input $RUN_ID --text 'hello'
    → expect: input accepted

    $ winch run events $RUN_ID
    → expect: event reflecting the input response

Poll outbox backlog until drained (SQL against `outbox`, worker metrics, or an
exposed debug endpoint introduced in this task if needed):

    → expect: backlog reaches 0; no poisoned rows

## Verification

- `make check` and `make test-integration` pass.
- `make e2e` passes through the input and outbox-drain steps.
- Existing outbox worker unit tests still pass.

## Acceptance criteria

- [ ] `SendRunInput` accepts text input idempotently and the harness receives it.
- [ ] Outbox worker runs for the daemon lifetime and drains pending rows from
  P0-008 as well as new input-driven rows.
- [ ] `EventStream.Publish` still has no non-test caller on the live path.
- [ ] `winch run input` demonstrates the behavior; polling (not WebSocket) is the
  manual check.
- [ ] I1 and I2 still hold.

## Deferrals

| Deferred | Owning task |
|---|---|
| Live WebSocket delivery | P0-010 |
| Stop | P0-011 |

## Traces to

- `docs/state.md` §*What went wrong* (`NewOutboxWorker` called only from tests)
- `internal/application/outbox.go` (publish precedes complete)
