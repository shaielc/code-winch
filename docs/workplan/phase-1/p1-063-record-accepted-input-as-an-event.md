# P1-063: Record accepted input as an event

**Phase:** 1 — Local single-user vertical slice
**Shape:** capability
**Dependencies:** P1-016 (compile: the `InputService.Accept` path and the `InputAcceptance` / `InputCommandStore` seam this extends), P1-050 (semantic: an appended event is only observable once the outbox worker, publisher, and events endpoints are wired in the composition root)

## Objective

An accepted input command appears in the run's own event history citing its
command ID, so a reader of `/runs/{id}/events` can see that a human sent
something and which command it was.

## Scope

- Extend `InputAcceptance` with the event to record, and append it inside the
  transaction `AcceptInput` already runs. That transaction holds the run row
  under `FOR UPDATE OF r` and already allocates `input_command_sequence`, so the
  event sequence comes from the same lock and the `run.events` outbox row is
  enqueued exactly as `Append` does it. Atomicity is the point: an accepted
  command with no event, or an event for a command that was not accepted, are
  both states no reader can recover from.
- Both implementations of the port — `internal/adapters/postgres` and
  `internal/adapters/memory` — so the fake profile and the real store agree.
- One event kind for every accepted input kind, carrying the command ID, the
  input kind, and the actor. A single kind keeps the citation uniform;
  `interrupt` and `resize` are not messages and should not be forced into a
  message family to get one.
- Payload by kind, classified at the sensitivity the content actually warrants:
  the text for `text` (`user-content`, which is what makes it a conversation
  entry for P2-022); rows and columns for `resize`; nothing beyond the kind for
  `interrupt`. For `terminal_bytes`, the byte count only — never the bytes.
  `docs/contracts.md` §3 already treats raw terminal input as a separately
  authorized capability because it bypasses redaction, and a password typed at a
  prompt is the ordinary case, not the exotic one.
- The replay path emits nothing. `AcceptInput` returns the originally recorded
  result when an idempotency key repeats, and that return happens before any
  write; a second event there would make the history disagree with the command
  table about how many times a person acted.

## Non-goals

- Attributing harness output back to the command that caused it. The harness
  emits bytes; `internal/adapters/harness/fake/fake.go:101` sends the command ID
  into the process and nothing carries it back, and Codex will not echo one
  either. Any link would be inferred from timing or ordering, and an inferred
  link in an audit trail is worse than an absent one. The citation this task
  builds is on the acceptance event, which the application knows with certainty.
- A new envelope field. `api/openapi/components/event.yaml` declares `kind` as
  an open string and `payload` as an open object, so the command ID travels in
  the payload and no contract changes.
- Rendering it — P2-022. Normalizing harness-produced events — P2-021. The
  approval and structured-answer input kinds — P2-056.
- Backfilling runs created before this lands.

## Runtime reachability

`cmd/winchd` with `WINCH_HARNESS_PROFILE=fake` on the compose stack:
`POST /api/v1/runs/{id}/input`, read back through `GET /runs/{id}/events` and
live on `/runs/{id}/events/stream`.

## Owned surfaces

`internal/application/input.go` and `input_test.go`; the `InputAcceptance`
struct in `internal/application/ports.go`; `AcceptInput` in both
`internal/adapters/postgres/repository.go` and
`internal/adapters/memory/memory.go`.

Its PostgreSQL integration test goes in a new
`internal/adapters/postgres/input_event_integration_test.go` rather than into
`repository_integration_test.go`, which P1-061 owns and is concurrently
rewriting; the two would otherwise collide for no reason.

No migration: the events and outbox tables already carry this, so this task
claims no migration number and cannot collide with one.

## Demonstration

    $ CMD=$(curl -fsS -X POST localhost:8080/api/v1/runs/$ID/input \
        -d '{"kind":"text","text":"echo hi"}' … | jq -r .commandId)
    $ curl -fsS localhost:8080/api/v1/runs/$ID/events | jq --arg c "$CMD" \
        '.events[] | select(.payload.commandId == $c)'
    → expect: exactly one event, naming the kind and the actor

Replaying the same idempotency key returns the first result and must not add a
second event:

    $ curl -fsS -X POST localhost:8080/api/v1/runs/$ID/input -H "Idempotency-Key: $K" …
    $ curl -fsS -X POST localhost:8080/api/v1/runs/$ID/input -H "Idempotency-Key: $K" …
    $ curl -fsS localhost:8080/api/v1/runs/$ID/events | jq --arg c "$CMD" \
        '[.events[] | select(.payload.commandId == $c)] | length'
    → expect: 1

Raw terminal input is recorded without its bytes:

    $ curl -fsS -X POST localhost:8080/api/v1/runs/$ID/input \
        -d '{"kind":"terminal_bytes","bytes":"'"$(printf 'hunter2' | base64)"'"}' …
    $ curl -fsS localhost:8080/api/v1/runs/$ID/events | grep -c hunter2
    → expect: 0

## Verification

- Service tests: each accepted kind produces exactly one event citing the
  command ID; each refusal path — unsupported kind, stale state, unauthorized
  raw terminal input, invalid request — produces none.
- A test asserting the replayed idempotency key adds no second event.
- PostgreSQL integration test that the command row and its event commit
  atomically: a forced failure after the command insert leaves neither behind.
- The accepted-input event reaches a subscribed WebSocket in sequence order,
  through the stream fixture P1-050's `stream_integration_test.go` establishes.

## Acceptance criteria

- [ ] Every accepted input command produces exactly one event in the run's
      history citing its command ID, for every kind the first slice accepts.
- [ ] A replayed idempotency key returns the original result and adds no event.
- [ ] A refused input produces no event.
- [ ] The event and the command row commit atomically; neither exists without
      the other after a crash between them.
- [ ] No event carries raw terminal-input bytes.
- [ ] `docs/contracts.md` §3's "the resulting event cites the command ID" is
      demonstrable by a command a person can run, against the fake profile.

## Deferrals

| Deferred | Owning task |
|---|---|
| Rendering the accepted input in a conversation view | P2-022 |
| Normalizing harness-produced events into structured kinds | P2-021 |
| The approval and structured-answer input kinds | P2-056 |

## Traces to

`docs/contracts.md` §2 (canonical envelope, event families), §3 (input
commands: "The resulting event cites the command ID");
`docs/architecture.md` §4 (application services)
