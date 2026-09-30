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
  validate the current version for a new command, while preserving the store's
  same-key replay even if the run version has since advanced. Return the
  recorded command ID/kind on replay. Map the service's stable
  rejection codes to the existing HTTP problem responses without logging text.
  The HTTP handler already requires `If-Match` and returns `202`
  (`internal/adapters/transport/httpapi/server.go:249-266`).
- Supply `InputCapabilityProvider` from the active fake/local execution and
  persisted run state. Refuse missing, terminal, and unsupported input modes;
  text is the supported slice of this command. The store already atomically
  records acceptance and a `run.input` outbox row, including replay by
  `(run ID, idempotency key)` (`internal/adapters/postgres/repository.go:324-374`).
  Preserve the design's command-ID correlation in the resulting event and
  update the relevant contract documentation alongside implementation if a
  schema or API meaning changes.
- Start `application.OutboxWorker` in `cmd/winchd` after constructing the
  store, supervisor, and runner. Poll `RunOnce` with a bounded idle interval;
  use a unique worker/lease identity and valid retry settings. Stop the loop
  and wait for it on every daemon exit within `ShutdownTimeout`, before closing
  the pool or runner; report worker errors without payload content.
- Route `run.input` outbox messages by their stable command ID to the active
  execution through the supervisor's fenced runner path, using the stored
  command to resolve its run and the runner's `input` payload. A successful
  runner handoff precedes outbox completion; failures retry or poison under the
  existing worker policy. Acknowledge `run.events` only after its durable append
  (already done by the store), **without calling `EventStream.Publish`**. Do
  not treat an unknown topic as successful delivery. A no-op publisher for all
  topics would silently drop accepted input (`OutboxMessage` has no run ID;
  `internal/application/ports.go:76-108`).
- Keep the fake harness reading input after a nonterminating transcript. Today
  `fakeHarnessConfig` sets `EarlyExit: true` unconditionally
  (`cmd/winchd/main.go:260-264`); make the input scenario's profile stay alive
  while preserving the existing transcript-driven exit scenario. Add
  `winch run input RUN_ID --text ...` with an optional explicit idempotency
  key; read the current ETag as `run start` does and print the accepted command
  ID so a person can retry an ambiguous request with the same key.
- Add the **`create → start → input → poll events`** scenario in
  `test/e2e/input_test.go`: start a run whose nonterminating transcript reaches
  the stdin reader, send a unique text marker, poll
  `GET /runs/{runId}/events` by sequence until that marker appears in a new
  harness event, and assert the run's `run.input` and `run.events` rows complete
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
- `internal/application/input.go` and focused input-delivery code (capabilities,
  topic routing and supervisor handoff)
- `internal/supervisor/` (fenced active-execution input path, if needed)
- `cmd/winch/main.go` (`input` subcommand)
- `test/e2e/input_test.go`
- `test/e2e/start_test.go` (worker-aware outbox assertion)
- Tests for replay, rejection, routing, retry, and shutdown drain
- Contract documentation if the implemented event/API contract changes

## Contract surfaces

- API: `POST /api/v1/runs/{runId}/input`
- port: `OutboxPublisher` (non-streaming topic-aware implementation)

## Demonstration

Configure a nonterminating fake transcript (no `exit`), and start a run that
remains `running` at its stdin reader. Use a marker absent from the transcript
and startup events, such as `echo input-marker-73`:

    $ winch run create --workspace /tmp/ws --harness fake --sandbox local
    $ winch run start $RUN_ID
    $ winch run input $RUN_ID --text 'echo input-marker-73' --idempotency-key input-73
    → expect: accepted=true and a command ID

    $ winch run events $RUN_ID
    → expect: a new harness output event containing `input-marker-73`

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
  shutdown within the configured budget.

## Acceptance criteria

- [ ] A running fake/local run accepts text through `SendRunInput`; the same
  `(run ID, idempotency key)` returns the same command ID without a second
  intent. Stale, missing, terminal, and unsupported new requests are refused
  with stable, content-free problems.
- [ ] Accepted `run.input` intent reaches the active harness before completion;
  failures do not mark it delivered. The worker drains earlier and new
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
