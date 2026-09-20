# P0-011: Stop run

**Phase:** 0 — Foundation repair
**Shape:** seam
**Dependencies:** P0-008 (semantic: a running execution path must exist before stop is meaningful)

## Objective

A person can force-stop a running fake harness through the operator CLI; the
run durably follows `running → stopping → completed`, and the harness and
all of its child processes are reaped.

## Scope

- Extend the execution service in `internal/application/start.go` with the stop
  use case rather than creating a second owner of its in-memory execution map.
  Serialize stop with observation handling through the existing per-execution
  lock, apply `RunCommandStop` before sending the runner command, and use the
  execution's existing fenced lease and ID. This preserves the architecture's
  persist-intent-before-runner-interaction rule and its one-writer guarantee.
- Send the existing runner `stop` command with a bounded grace period. Mark the
  execution as intentionally stopping before the command can produce an exit
  observation, so that the observation pump treats the resulting successful
  fake-harness exit as the contract's expected `stopping → completed`
  transition. Keep cleanup and lease release in the existing terminal
  observation path; a stop-command or cleanup failure must not report a
  successful terminal outcome.
- Preserve the API contract already mounted by `httpapi`: require the current
  ETag, map missing runs, stale versions or leases, and illegal states to the
  existing stable problem responses, and make a replay in `stopping` or any
  terminal state safe without sending another runner command. Bind this use
  case in `runBackend.StopRun`, replacing the owned deferred error.
- Add `winch run stop RUN_ID`. Like `run start`, it first reads the run to obtain
  the current ETag, sends a fresh or explicitly supplied idempotency key to
  `POST /api/v1/runs/{runId}/stop`, and prints the accepted run. Add the command
  to CLI usage and the operator command list in `deployments/README.md`.
- Add application tests for ordering (durable `stopping` intent precedes the
  runner command), the expected terminal transition, stop/observation
  serialization, stop-command failure, stale-version refusal, and safe replay
  from `stopping` and terminal states. Keep the existing local-runner process
  group stop tests green; they provide the lower-level escalation proof.
- Add `TestCreateStartStopThenGet` in `test/e2e/stop_test.go`. Configure a
  transcript action with enough delay to keep the real fake-harness process
  running, create and start the run through the daemon, record the exact
  fake-harness PID while it is a descendant of that test's daemon, stop via
  HTTP, and poll `get` until the persisted state is `completed`. Then wait for
  that recorded PID to disappear from `/proc`; do not use a global `ps | grep`
  count, which matches unrelated shells and cannot identify an orphan from this
  run.

## Non-goals

- Cancelling a run in `created` or `queued`; this task stops an active
  `running` execution and otherwise follows the already-defined lifecycle.
- **Any e2e step beyond `create → start → stop → get`.** No input and no
  WebSocket steps; P0-007 owns the assembled round trip.
- Wiring `make e2e` into CI, or describing every run route as fully bound in
  deployment documentation; P0-007 owns phase closure.
- Changing the public stop request, response, ETag, idempotency, runner message,
  stop-escalation, or run-lifecycle contracts. They already exist at HEAD; this
  task implements them.
- The memory-store profile; P0-017 revises this PostgreSQL seam for that
  profile.
- Browser UI or session cookies.
- Daemon restart reconciliation, including resuming a persisted `stopping`
  intent after restart; P0-008's phase-1 deferral owns that system-level gap.

## Runtime reachability

- **Composition root:** `cmd/winchd`, where `runBackend.StopRun` calls the same
  execution service that owns the live run.
- **Profile:** `harnessProfile=fake`, `sandboxProfile=local`, PostgreSQL storage.
- **Command:** `winch run stop RUN_ID`; `make e2e` exercises the same route
  through the real daemon, runner, local sandbox, and fake-harness binaries.

## Write set

- `internal/application/start.go`
- `internal/application/start_test.go`
- `cmd/winchd/main.go`
- `cmd/winch/main.go`
- `cmd/winch/main_test.go`
- `deployments/README.md`
- `test/e2e/stop_test.go`

## Contract surfaces

- API: `POST /api/v1/runs/{runId}/stop` (existing operation implemented here)
- persisted run lifecycle: `running → stopping → completed`, with failed
  stop/cleanup following the existing `stopping → failed` rule

## Demonstration

Start the deployed PostgreSQL profile with a fake transcript whose first action
is delayed long enough for the stop command, then use the shipped CLI:

    $ export WINCH_API_URL=http://localhost:8080
    $ export WINCH_TOKEN=<the deployment token>
    $ export WINCH_CSRF_TOKEN=<the deployment CSRF token>
    $ RUN_ID=$(bin/winch run create --workspace /tmp/ws --harness fake --sandbox local)
    $ bin/winch run start "$RUN_ID"
    $ bin/winch run stop "$RUN_ID"
    $ bin/winch run get "$RUN_ID"
    → expect: start prints `running`, stop prints `stopping` or `completed`,
      and get reaches `completed`

Run the focused real-process scenario. It captures this run's harness PID before
stop and fails unless that exact PID disappears, so it cannot pass by counting
an unrelated process:

    $ PG_TEST_DATABASE_URL=<disposable-postgres-url> \
        WINCHD_BIN="$PWD/bin/winchd" \
        go test ./test/e2e -run '^TestCreateStartStopThenGet$' -count=1 -v
    → expect: PASS; the run persisted `completed` and its recorded
      fake-harness PID no longer exists

## Verification

- `make check` passes.
- `make test-integration` passes.
- `make e2e` passes locally, including `TestCreateStartStopThenGet` against
  PostgreSQL and the real local process substrate.
- `go test ./internal/application ./internal/runner/local
  ./internal/adapters/sandbox/local ./cmd/winch ./cmd/winchd` passes, covering
  stop orchestration, escalation/reaping, CLI dispatch, and daemon binding.

## Acceptance criteria

- [ ] A stop of a running execution durably records `stopping` before the
      runner is called, then records `completed` after the expected stopped
      exit; a runner-stop or cleanup failure cannot be reported as completed.
- [ ] Stop is serialized with runner observations under the execution's fenced
      lease, and stale ETags/leases and illegal source states are refused with
      the existing stable HTTP problems.
- [ ] Replaying stop while the run is `stopping` or terminal is safe and does
      not send another stop command or launch another execution.
- [ ] The local sandbox's bounded stop escalation reaps the process group, and
      the e2e scenario proves the exact fake-harness process started for the run
      is gone after stop.
- [ ] `winch run stop RUN_ID` drives the deployed route using the current ETag
      and an idempotency key, and prints the accepted run state.
- [ ] `make e2e` runs `create → start → stop → get` through the real
      daemon, PostgreSQL store, local runner, and fake-harness process.
- [ ] I1 and I2 still hold.

## Deferrals

| Deferred | Owning task |
|---|---|
| Assembling the complete round trip and gating `make e2e` in CI | P0-007 |
| Stop on the memory-store profile | P0-017 |
| Browser `winch_session` establishment | Phase 1 — registered in [`../phase-1/README.md`](../phase-1/README.md) |

## Traces to

- `docs/architecture.md` §4 (application services and the run supervisor) and
  §7 (persist stop intent before runner interaction; escalation and cleanup)
- `docs/contracts.md` §1 (`running → stopping → completed/failed`, terminal
  immutability, and idempotent stop)
- `docs/roadmap.md` Phase 1 exit ("forced stop leaves no child processes")
- `docs/state.md` §*The run round trip*
- Invariant I4 — this task contributes the stop scenario; P0-007 closes the
  PostgreSQL round trip and P0-017 later revises this scenario for memory
- [`../post-mortems/2026-09-16-a-plan-defect-read-from-one-implementation.md`](../post-mortems/2026-09-16-a-plan-defect-read-from-one-implementation.md)
  (the original global `ps | grep` demonstration could not identify this run's
  process and matched the invoking shell)
