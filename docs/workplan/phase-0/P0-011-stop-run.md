# P0-011: Stop run

**Phase:** 0 — Foundation repair
**Shape:** seam
**Dependencies:** P0-008 (semantic: a running execution path must exist before stop is meaningful)

## Objective

A person can force-stop a running harness through the deployed CLI; the run
reaches a truthful persisted terminal state and leaves no owned child processes.

## Scope

- Add a stop use case beside `internal/application/start.go`, using its
  `RunRepository`, `RunExecutor`, execution registration, and per-execution
  lock. A stop must serialize with runner observations and normal exit, retain
  the live lease, and reject stale `If-Match` versions before changing a run.
  Do not acquire a second lease for an execution already owned by StartService.
- Apply `RunCommandStop` to a running attempt and send a `stop`
  `ExecutionCommand` with `RunStateStopping` and `protocol.StopPayload` through
  `supervisorControl`. `Supervisor.Execute` saves desired state before
  `runner.Send`; the local runner already calls `sandbox.Stop`, which escalates
  against the owned process group. Choose and document a bounded grace period.
  Keep the existing observation consumer responsible for terminal event,
  cleanup, attempt transition, and lease release. Coordinate that consumer with
  stop so a forced, nonzero exit after a requested stop has a truthful terminal
  result rather than a success inferred from the request.
- Follow the existing domain policy: `running → stopping → completed` for a
  successful exit, `running → stopping → failed` for a failed exit or cleanup
  failure. `cancelled` is for `created` or `queued` cancellation, not this
  running stop. Repeated stop in `stopping` or a terminal state is safe and
  does not start a second stop or rewrite terminal history. Handle the exit
  racing stop: return the observed terminal run and do not overwrite it.
  Surface unknown runs, stale versions, and invalid states through the existing
  HTTP problem mapping.
- Replace the deferred `runBackend.StopRun` in `cmd/winchd/main.go`, translating
  the API run ID and application errors as `StartRun` does. Preserve the
  existing `POST /api/v1/runs/{runId}/stop` request, response, `If-Match`,
  `Idempotency-Key`, and optional reason contract; do not redefine it.
- Add `winch run stop RUN_ID` to `cmd/winch/main.go` and its usage text.
  Follow `runStart`: read the current version, send a fresh idempotency key
  and quoted `If-Match`, then print the returned run. A stopped run can be
  read again with `winch run get`.
- Add `test/e2e/stop_test.go` using the daemon, PostgreSQL, fake harness,
  local sandbox, and helpers in `test/e2e/start_test.go`. Configure a
  P0-003 transcript that keeps the process alive until stop (verify the fake
  harness's `EarlyExit` behavior when selecting it), then exercise
  `create → start → stop → get`. Poll the persisted run with a deadline,
  assert the policy above and an intact run identity, and inspect only the
  process or process group belonging to this test's execution. Add focused
  tests for repeated stop, stop/exit race, escalation, and descendant reaping.

## Non-goals

- Any e2e step beyond `create → start → stop → get`. Input and WebSocket
  steps and the assembled round trip belong to P0-007.
- Wiring `make e2e` into CI or claiming fully bound routes in
  `deployments/README.md` (P0-007).
- The memory-store stop profile (P0-017), browser UI or session cookies
  (Phase 1), and restart reconciliation of an interrupted run (Phase 1).
- New CLI subcommands beyond stop, or changing the public stop schema.

## Runtime reachability

- **Composition root:** `cmd/winchd`, whose run backend delegates stop to the
  application service and whose existing observation goroutine records exit.
- **Profile:** fake harness, local sandbox, PostgreSQL storage.
- **Command:** `winch run stop`, `make e2e`.

## Write set

- `cmd/winchd/main.go`
- `internal/application/start.go` and focused stop service/tests
- `cmd/winch/main.go` and its tests
- `test/e2e/stop_test.go`

## Contract surfaces

- API: `POST /api/v1/runs/{runId}/stop` (existing operation, now bound)
- Persisted run attempt: `running → stopping → completed/failed`

## Demonstration

Run the PostgreSQL-backed daemon with `WINCH_FAKE_HARNESS_TRANSCRIPT` set to
a transcript that waits for stop, and the CLI configured with its API URL,
token, and CSRF token. Use a unique workspace and record the harness process
group for this run before stopping it:

    $ RUN_ID=$(winch run create --workspace "$WORKSPACE" --harness fake --sandbox local)
    $ winch run start "$RUN_ID"
    → expect: the same run reports running; its harness is still alive
    $ winch run stop "$RUN_ID"
    $ winch run get "$RUN_ID"
    → expect: the same run becomes completed on successful exit or failed on
      forced/nonzero exit; it does not stay running or stopping
    $ kill -0 -- "-$OWNED_PGID"
    → expect: no such process group (use the captured group for this run only;
      do not count process names across the host)
    $ make e2e
    → expect: the stop scenario and all existing scenarios pass

The `kill -0` check is Unix-specific; also confirm the test-owned leader was
reaped. If a group ID could have been reused, inspect the captured process
identities rather than treating the group lookup alone as proof.

## Verification

- `make check` and `make test-integration` pass.
- `make e2e` passes locally, including the new stop scenario.
- Focused application, local runner/sandbox, HTTP, and CLI tests cover
  conditional stop, idempotence, exit races, escalation, and owned descendants.

## Acceptance criteria

- [ ] A running execution accepts stop through the deployed API and CLI;
  desired state is persisted before the runner stop command, and a subsequent
  get reports the same run as `completed` for successful exit or `failed`
  for failed exit/cleanup, never falsely `cancelled` or indefinitely
  `stopping`.
- [ ] Stop escalates by the configured deadline, reaps the test-owned harness
  leader and descendants, cleans up its sandbox, and releases its lease;
  neither the e2e check nor the manual check matches unrelated processes.
- [ ] Repeated stop while stopping or terminal has no second process effect or
  history rewrite; a concurrent natural exit cannot be overwritten. Unknown,
  stale-version, and invalid-state requests return the existing mapped problem
  responses.
- [ ] `winch run stop` uses the current ETag and an idempotency key, and the
  manual `create → start → stop → get` check observes a persisted terminal run.
- [ ] `make e2e` runs the PostgreSQL-backed
  `create → start → stop → get` scenario locally, alongside existing scenarios.
- [ ] I1 and I2 still hold.

## Deferrals

| Deferred | Owning task |
|---|---|
| Assembling the complete round trip and gating `make e2e` in CI | P0-007 |
| Stop scenario on the memory store profile | P0-017 |

Browser session establishment and restart reconciliation are registered in
[`../phase-1/README.md`](../phase-1/README.md); they are outside this task.

## Traces to

- `docs/roadmap.md` Phase 1 exit ("forced stop leaves no child processes")
- `docs/state.md` §*The run round trip*
- `docs/contracts.md` §1 (running/stopping/terminal transitions)
- `docs/architecture.md` §7 (persist intent, escalate, clean up)
- Invariant I4 — this task contributes the stop scenario; P0-007 closes I4
