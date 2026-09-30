# P0-009: Send input and drain outbox

**Phase:** 0 — Foundation repair
**Shape:** seam
**Dependencies:** P0-008 (semantic: a running execution path must exist before input is meaningful)

## Objective

A person can send text input to a running harness; the outbox worker drains
publish intent with polling as the demonstration surface — no WebSocket yet.

## Scope

- Replace `runBackend.SendRunInput`'s deferred error in `cmd/winchd/main.go`
  with `InputService.Accept` backed by the PostgreSQL `InputCommandStore`.
  Translate the API's text request, actor, idempotency key, and `If-Match`
  version into a new command ID and the service's expected running state;
  validate the current version **inside the store transaction** for a new
  command, after its existing same-key lookup and before insertion. The store
  checks state and event sequence but not run version (`repository.go:331-361`);
  add the expected version to acceptance and compare it under the run-row lock.
  A stale new key returns the existing precondition problem. Preserve same-key
  replay after version or state advances, returning the recorded command ID/kind
  without a second intent. `InputService.Accept` currently rejects unsupported
  modes before the store can find a replay (`input.go:106-121`); adapt this
  ordering so valid replays reach the transactional lookup while new requests
  still refuse unsupported or terminal modes. Map the service's stable
  rejection codes to the existing HTTP problem responses without logging text.
  The HTTP handler already requires `If-Match` and returns `202`
  (`internal/adapters/transport/httpapi/server.go:249-266`).
- Supply `InputCapabilityProvider` from the active fake/local execution and
  persisted run state. Refuse missing, terminal, and unsupported input modes;
  text is the supported slice of this command. The store already atomically
  records acceptance and a `run.input` outbox row, including replay by
  `(run ID, idempotency key)` (`internal/adapters/postgres/repository.go:324-374`).
  Ensure a new text request has a live input-capable fake/local run and active
  execution, with state checked again at acceptance. Preserve command-ID
  correlation in the resulting event and update the relevant contract
  documentation alongside implementation if a schema or API meaning changes.
- Start `application.OutboxWorker` in `cmd/winchd` after constructing the
  store, supervisor, and runner. Poll `RunOnce` with a bounded idle interval;
  use a unique worker/lease identity and valid retry settings. On cancellation,
  stop claiming, bound in-flight publishing and wait within `ShutdownTimeout`
  before closing the pool or runner; leave unfinished claims for lease
  expiry/retry. Report stable error classes and IDs without payload content.
- Route `run.input` outbox messages by their stable command ID to the active
  execution through the supervisor's fenced runner path, using the stored
  command to resolve its run and the runner's `input` payload. Use the active
  execution's lease token and execution ID; absent or stale execution fails
  delivery. `Supervisor.Execute` persists desired state before sending
  (`supervisor.go:149-165`), so use a fenced, serialized input path that does
  not rewrite run state merely to deliver a command. A successful
  runner handoff precedes outbox completion; failures retry or poison under the
  existing worker policy. Acknowledge `run.events` only after its durable append
  (already done by the store), **without calling `EventStream.Publish`**. Do
  not treat an unknown topic as successful delivery. A no-op publisher for all
  topics would silently drop accepted input (`OutboxMessage` has no run ID;
  `internal/application/ports.go:76-108`).
- Keep the fake harness reading input after a nonterminating transcript. Today
  `fakeHarnessConfig` sets `EarlyExit: true` unconditionally
  (`cmd/winchd/main.go:260-264`); make the input scenario's profile stay alive
  while preserving the existing transcript-driven exit scenario. Keep the
  transcript free of the input marker so later output proves stdin delivery. Add
  `winch run input RUN_ID --text ...` with an optional explicit idempotency
  key; read the current ETag as `run start` does and print the accepted command
  ID and kind so a person can retry an ambiguous request with the same key.
  A retry may read a newer ETag, including after the run has ended; the service
  must still return the recorded result for that key. State clearly that a
  generated key applies to one attempt and the caller must supply the original
  key to retry an ambiguous response.
- Add the **`create → start → input → poll events`** scenario in
  `test/e2e/input_test.go`: start a run whose nonterminating transcript reaches
  the stdin reader, send a unique text marker, poll
  `GET /runs/{runId}/events` from a cursor captured before input until the
  marker appears in a later harness output event with command-ID correlation, and assert the run's `run.input` and `run.events` rows complete
  with no poisoned rows. Check a repeated key returns the same command ID and
  creates only one command and outbox intent; reject a new-key input after the
  run becomes terminal. Outbox delivery is at least once, so tests must not
  claim exactly-once physical handoff after a worker crash.
  Update `test/e2e/start_test.go`'s assertion that every event remains pending:
  it explicitly assumes no worker (`start_test.go:72-75`) and must instead
  wait for those rows to complete without changing its start scenario.

## Non-goals

- **Any e2e step beyond `create → start → input → poll events`.** No WebSocket
  assertions and no stop command; the scenario asserts delivery by polling.
- `EventStream.Publish` or WebSocket subscribers.
- `winch run stream` or any WebSocket client.
- `StopRun` (P0-011).
- Proving live browser delivery.
- Memory-store wiring and a database-free input scenario (P0-015).
- New input kinds or a general runner restart/reconciliation mechanism.

## Runtime reachability

- **Composition root:** `cmd/winchd` (outbox worker loop).
- **Profile:** fake harness, local sandbox, PostgreSQL storage.
- **Command:** `winch run input`.

## Write set

- `cmd/winchd/main.go` (input wiring, fake profile, worker lifecycle)
- `internal/application/input.go`, `internal/application/ports.go`, and focused
  input-delivery code (replay ordering, expected version, capabilities, topic
  routing and supervisor handoff)
- `internal/adapters/postgres/repository.go` (transactional version check and
  replay before new-command rejection)
- `internal/supervisor/` (fenced active-execution input path, if needed)
- `cmd/winch/main.go` (`input` subcommand)
- `test/e2e/input_test.go`
- `test/e2e/start_test.go` (worker-aware outbox assertion)
- Tests for replay, rejection, routing, retry, and shutdown drain
- Contract documentation if the implemented event/API contract changes

## Contract surfaces

- API: `POST /api/v1/runs/{runId}/input`
- ports: input acceptance expected-version semantics and `OutboxPublisher`
  (non-streaming topic-aware implementation)

## Demonstration

Configure a nonterminating fake transcript (no `exit`) with no
`input-marker-73` in its content, and start a run that remains `running` at
its stdin reader. Record the highest event sequence before sending the unique
text `echo input-marker-73`:

    $ winch run create --workspace /tmp/ws --harness fake --sandbox local
    $ winch run start $RUN_ID
    $ winch run input $RUN_ID --text 'echo input-marker-73' --idempotency-key input-73
    → expect: accepted=true and a command ID

    $ winch run events $RUN_ID --after-sequence $BEFORE_INPUT_SEQUENCE
    → expect: a later harness output event containing `input-marker-73`,
      correlated to the accepted command ID

    $ winch run input $RUN_ID --text 'echo input-marker-73' --idempotency-key input-73
    → expect: the same command ID; no second intent or marker event in this
      uninterrupted demonstration

Poll PostgreSQL for this run's `run.input` and `run.events` rows (the API event
page reads durable history independently of outbox delivery):

    → expect: all matching rows have `completed_at` set and none has
      `poisoned_at` set, within a bounded timeout

## Verification

- `make check` and `make test-integration` pass.
- `make e2e` passes through the input and outbox-drain steps.
- Existing outbox worker unit tests and the start scenario pass with the
  worker active. Focused tests cover same-key replay, stale/terminal refusal,
  `run.input` delivery before completion, unknown-topic failure, and worker
  shutdown within the configured budget. Force stale `If-Match` for a new key,
  replay an accepted key after version/state advances, and verify a stale or
  failed handoff never falsely completes its row.

## Acceptance criteria

- [ ] A running fake/local run accepts text through `SendRunInput`; the same
  `(run ID, idempotency key)` returns its recorded command ID and kind after
  version/state advances without a second intent. A new key with stale
  `If-Match`, missing or terminal run, or unsupported mode is refused with a
  stable, content-free problem. Concurrent version changes cannot slip between
  validation and acceptance.
- [ ] Accepted `run.input` intent reaches the currently leased harness before
  completion and produces a later correlated output event; stale/absent
  execution or send failure does not mark it delivered. The worker drains earlier and new
  `run.events` intents, with no poison on the happy path, and stops within the
  daemon shutdown budget.
- [ ] `EventStream.Publish` still has no non-test caller on the live path.
- [ ] `winch run input` exposes accepted command identity and polling reveals
  the unique response; the same-key replay and SQL check establish one
  acceptance intent and outbox completion separately.
- [ ] I1 and I2 still hold.

## Deferrals

| Deferred | Owning task |
|---|---|
| Live WebSocket delivery | P0-010 |
| Stop | P0-011 |

## Traces to

- `docs/state.md` §*What went wrong* (`NewOutboxWorker` called only from tests)
- `internal/application/outbox.go` (publish precedes complete)
- `docs/contracts.md` §3 (durable input acceptance, replay and command ID)
