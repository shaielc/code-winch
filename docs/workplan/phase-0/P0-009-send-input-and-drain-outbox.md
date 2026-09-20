# P0-009: Send input and drain outbox

**Phase:** 0 — Foundation repair
**Shape:** seam
**Dependencies:** P0-008 (semantic: a running execution path must exist before input is meaningful)

## Objective

A person can durably and idempotently send text to a running fake harness with
`winch run input`; a daemon-owned worker delivers the input and drains all
publish intent while durable event polling shows the harness response.

## Scope

- Replace `runBackend.SendRunInput`'s P0-009-owned stub in
  `cmd/winchd/main.go` with an adapter from the generated HTTP request to the
  existing `application.InputService`. Read the run first, enforce the supplied
  `If-Match` version, mint the command ID, require the running state, and map
  `InputResult` and the stable `InputError` classes to the existing HTTP
  response types. This task's maintained command sends text only; retain the
  application service's existing typed-input validation rather than widening
  the HTTP contract.
- Make the execution created by `application.StartService` provide the current
  input capabilities and accept a delivered `run.input` command through its
  existing live execution, lease, and `RunExecutor`/supervisor path. The fake
  harness supports text input; delivery must use the accepted command ID as the
  runner input ID so an outbox redelivery is stable. Serialize delivery with
  observation handling for that execution, and return content-free errors when
  no live input-capable execution exists.
- Make a `run.input` outbox message self-routing by including its run ID in the
  delivery envelope produced by `InputService`. Add an application-layer
  `OutboxPublisher` that validates the topic and envelope, sends `run.input` to
  the live execution, and deliberately acknowledges `run.events` without
  calling `httpapi.EventStream.Publish`: durable `GET /events` polling, not
  live fan-out, is this slice's event delivery surface. Unknown or malformed
  messages fail publication and follow the worker's existing retry/poison
  policy rather than being silently completed.
- Construct `InputService`, the polling-only publisher, and `OutboxWorker` in
  `cmd/winchd`. Run `OutboxWorker.RunOnce` repeatedly with a bounded idle
  interval from daemon startup until cancellation. On shutdown, stop claiming
  new work and wait for an in-flight iteration within `WINCH_SHUTDOWN_TIMEOUT`;
  worker failures are logged with stable, content-free fields and do not stop
  the HTTP server. Use explicit, valid worker lease/backoff/batch constants and
  process-unique owner/token values; do not add configuration that the
  demonstration does not need.
- Remove the unconditional early-exit override in `fakeHarnessConfig`: an
  empty or non-terminating transcript must leave the fake harness reading
  stdin so the deployed profile can demonstrate input. Keep the P0-008
  transcript ending in `exit` as the non-interactive start scenario.
- Add `winch run input RUN_ID --text TEXT [--idempotency-key KEY]`. Like
  `run start`, it first reads the run and supplies its current ETag, then prints
  the accepted command result. Reject a missing run ID or empty text locally.
- Add `test/e2e/input_test.go` using the existing PostgreSQL daemon fixture: run
  `create → start` with a fake transcript that reaches end-of-file without
  exiting, post text input twice with one idempotency key, poll durable events
  until the echoed response appears, and query PostgreSQL until the rows for
  that run have neither pending nor poisoned outbox records. Assert both posts
  return the same command ID and only one input command exists.
- Cover the focused failure boundaries: stale `If-Match` and input to a
  non-running run are refused without an input command/outbox row; malformed or
  unknown outbox messages are retried and never completed; shutdown does not
  abandon an in-flight worker iteration; and logs contain neither input text
  nor outbox payloads.

## Non-goals

- WebSocket publication, `EventStream.Publish`, or subscribers; P0-010 replaces
  the polling-only publisher with live fan-out.
- Input kinds other than the `text` command demonstrated here, or raw-terminal
  authorization policy.
- Outbox recovery guarantees beyond the existing retry, poison, lease, and
  at-least-once behavior.
- `winch run stream` or any WebSocket client.
- `StopRun` (P0-011).
- The memory-store profile (P0-015) or live browser delivery.

## Runtime reachability

- **Composition root:** `cmd/winchd` constructs the input service and owns the
  outbox worker loop and its shutdown.
- **Profile:** fake harness, local sandbox, PostgreSQL storage; the selected
  transcript reaches EOF and then waits for text input.
- **Command:** `winch run input`, with `winch run events` polling the durable
  response.

## Write set

- `cmd/winchd/main.go`
- `cmd/winchd/main_test.go`
- `internal/application/input.go`
- `internal/application/input_test.go`
- `internal/application/outbox.go`
- `internal/application/outbox_test.go`
- `internal/application/start.go`
- `internal/application/start_test.go`
- `cmd/winch/main.go`
- `cmd/winch/main_test.go`
- `test/e2e/input_test.go`

## Contract surfaces

- API implementation: `POST /api/v1/runs/{runId}/input` (the existing OpenAPI
  operation and schemas do not change)
- port behavior: `InputCapabilityProvider` and `OutboxPublisher`
- outbox protocol: `run.input` delivery envelope; `run.events` is acknowledged
  without live publication in this slice

## Demonstration

Start the deployment with `WINCH_FAKE_HARNESS_TRANSCRIPT` naming an empty file
so the fake harness waits for stdin, and export the CLI API/authentication
settings. Then:

    $ RUN_ID=$(bin/winch run create --workspace /tmp/ws --harness fake --sandbox local)
    $ bin/winch run start "$RUN_ID"
    $ INPUT=$(bin/winch run input "$RUN_ID" --text 'echo hello' --idempotency-key demo-input-1)
    $ printf '%s\n' "$INPUT"
    → expect: accepted=true, kind=text, and a command ID

Poll the maintained durable-event command until the response arrives:

    $ until bin/winch run events "$RUN_ID" | tee /tmp/winch-events | grep -q 'hello'; do sleep 0.1; done
    → expect: a stream.raw event whose data contains "hello"

Inspect this run's worker result without using it as the delivery surface:

    $ COMMAND_ID=$(printf '%s\n' "$INPUT" | jq -r .commandId)
    $ psql "$WINCH_DATABASE_URL" -v command_id="$COMMAND_ID" -Atc "SELECT count(*) FILTER (WHERE completed_at IS NULL AND poisoned_at IS NULL), count(*) FILTER (WHERE poisoned_at IS NOT NULL) FROM outbox WHERE run_id = (SELECT run_id FROM input_commands WHERE id = :'command_id')"
    → expect: 0|0

## Verification

- `make check` passes.
- `make test-integration` passes, including atomic input acceptance and outbox
  claim/complete behavior against PostgreSQL.
- `make e2e` passes both the unchanged start scenario and the new
  `create → start → input → poll events` scenario.
- `go test ./internal/application ./cmd/winchd ./cmd/winch` passes the focused
  input-delivery, worker-lifecycle, CLI, negative, and content-canary tests.

## Acceptance criteria

- [ ] A current-version text request to a running run returns `202`; retrying
  the same `(run ID, actor, idempotency key)` returns the original command ID,
  creates no second input command, and the fake harness receives the text once
  through the live supervisor path.
- [ ] A stale ETag or a run that is not input-capable is refused with the
  existing stable HTTP problem response; no input command or outbox intent is
  persisted and no text reaches the harness.
- [ ] The daemon starts one outbox loop, drains both pre-existing `run.events`
  intent from P0-008 and new `run.input`/response-event intent to zero, and
  stops or finishes its in-flight iteration inside the configured shutdown
  budget.
- [ ] A malformed or unknown outbox message is not marked complete; the
  existing worker retry/poison behavior is observable without payload content
  appearing in errors or logs.
- [ ] The live path has no call to `EventStream.Publish`; `winch run input`
  followed by repeated `winch run events` calls visibly produces the harness
  response through durable polling.
- [ ] The unchanged P0-008 start scenario passes, and the daemon and deployment
  remain startable (I1 and I2).

## Deferrals

| Deferred | Owning task |
|---|---|
| Live WebSocket delivery | P0-010 |
| Stop | P0-011 |
| Memory-store input and outbox parity | P0-015 |

## Traces to

- `docs/state.md` §*What went wrong* (`NewOutboxWorker` and
  `NewInputService` called only from tests)
- `docs/architecture.md` §6 (every published mutation uses a transactional
  outbox)
- `docs/contracts.md` §1 (input only in an input-capable state)
- `docs/contracts.md` §3 (idempotency, acceptance plus outbox transaction,
  supported kinds, and content-free diagnostics)
- `docs/security.md` §3 T03 and §5 (duplicate/reordered input and
  content-free telemetry)
- `internal/application/input.go` (`InputService.Accept`)
- `internal/application/outbox.go` (publish precedes complete)
- `internal/application/start.go` (live execution and supervisor command path)
