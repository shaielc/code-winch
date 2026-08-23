# P0-010: Live WebSocket event stream

**Phase:** 0 — Foundation repair
**Shape:** seam
**Dependencies:** P0-009 (semantic: outbox worker must run before stream publisher replaces the polling-only publisher)

## Objective

A person can subscribe to a run's live event stream over WebSocket; outbox
delivery fans out through `EventStream.Publish` after durable commit.

## Scope

- Replace the P0-009 `OutboxPublisher` with an implementation that decodes
  outbox payloads and calls `httpapi.EventStream.Publish`.
- Ensure new events after this wiring are visible on the existing WebSocket route
  (`GET /api/v1/runs/{runId}/events/stream`).
- Add `winch run stream` (WebSocket client with bearer auth and resume via
  `after_sequence` if the API supports it).
- Extend the standing `make e2e` scenario with WebSocket stream assertions.

## Non-goals

- `SendRunInput` or `StopRun` changes beyond what streaming requires.
- Browser session cookies or web UI work.
- New event kinds or renderers.

## Runtime reachability

- **Composition root:** `cmd/winchd` (stream publisher).
- **Profile:** fake harness, local sandbox, PostgreSQL storage.
- **Command:** `winch run stream`.

## Write set

- `internal/application/` or `internal/adapters/transport/httpapi/` (stream
  publisher adapter)
- `cmd/winchd/main.go` (publisher registration swap)
- `cmd/winch/` (`stream` subcommand)
- Tests proving publish follows durable commit and subscribers receive events

## Contract surfaces

- API: `GET /api/v1/runs/{runId}/events/stream`
- port: `OutboxPublisher` (streaming implementation)

## Demonstration

    $ winch run create ...
    $ winch run start $RUN_ID &
    $ winch run stream $RUN_ID
    → expect: live harness output lines appear on the WebSocket without polling
      `winch run events`

Disconnect and reconnect with `after_sequence` if supported:

    → expect: no duplicate output for already-delivered sequences

## Verification

- `make check` and `make test-integration` pass.
- `make e2e` passes through the WebSocket stream step.
- Existing `httpapi` stream tests still pass.

## Acceptance criteria

- [ ] `EventStream.Publish` has a non-test caller on the live outbox path.
- [ ] WebSocket subscribers receive events after outbox publish succeeds.
- [ ] `winch run stream` demonstrates live delivery by hand.
- [ ] P0-009 polling path (`winch run events`) still works.
- [ ] I1 and I2 still hold.

## Deferrals

| Deferred | Owning task |
|---|---|
| Stop | P0-011 |

## Traces to

- `docs/state.md` §*What went wrong* (`EventStream.Publish` has no non-test
  caller)
- `internal/adapters/transport/httpapi/stream.go`
