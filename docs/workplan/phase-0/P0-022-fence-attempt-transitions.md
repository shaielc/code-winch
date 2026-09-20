# P0-022: Fence attempt transitions by the run lease

**Phase:** 0 — Foundation repair
**Shape:** hardening
**Dependencies:** P0-008 (revision: its object is the attempt transitions P0-008's start use case writes through the unfenced `RunRepository.Save` while it holds a lease)

## Objective

A daemon that has lost a run's lease can no longer change that run's attempt
state: the write is refused, as its event appends and runner commands already
are.

## Scope

- Add an attempt-state write to `application.SupervisorStore` that checks the
  lease owner, token, and epoch in the same statement that saves the attempt,
  following `SaveDesiredState` (`internal/adapters/postgres/repository.go:440-449`).
  Implement it in the PostgreSQL and memory adapters.
- Expose it through `supervisor.Supervisor` and the `application.RunExecutor`
  port, and route every attempt transition an execution writes while it holds a
  lease through it — including any stop transition present at HEAD when this
  task lands. Today `StartService.apply` uses `RunRepository.Save`
  (`internal/application/start.go:307`), which checks only the version.
- Once any fenced call for an execution reports a stale lease, stop changing
  that run from the execution: no further transitions, appends, or commands.
- Keep `created → queued` on `RunRepository.Save`; no lease exists yet.
- Tests: the store write refuses a stale owner, token, or epoch in both
  adapters; a replacement owner takes the lease, and the old owner's harness
  exit leaves the attempt unchanged.

## Non-goals

- Calling `Supervisor.Reconcile` at daemon startup.
- A start interrupted before it holds a lease — P0-021.
- An event append failing while the lease is held — P0-020.

## Runtime reachability

- **Composition root:** `cmd/winchd` (the `supervisorControl` adapter).
- **Profile:** `harnessProfile=fake`, `sandboxProfile=local`, PostgreSQL storage.
- **Command:** `winch run start` and `winch run get`, with the run's lease taken
  over in the database.

## Write set

- `internal/application/supervisor.go` (port)
- `internal/application/start.go`
- `internal/application/start_test.go`
- `internal/supervisor/supervisor.go`
- `internal/supervisor/supervisor_test.go`
- `internal/adapters/postgres/repository.go`
- `internal/adapters/postgres/repository_integration_test.go`
- `internal/adapters/memory/memory.go`
- `internal/adapters/memory/memory_test.go`
- `cmd/winchd/main.go` (`supervisorControl`)

## Contract surfaces

- port: `application.SupervisorStore` — a lease-fenced attempt-state write
- port: `application.RunExecutor` — the same write, through the supervisor
- persisted aggregate: changes to `run_attempts.state` made by an execution
  require that execution's current lease

## Demonstration

Start a run that stays up, take its lease over in the database, and let the
harness exit:

    $ WINCH_FAKE_HARNESS_DELAY=10s winchd &
    $ RUN_ID=$(winch run create --workspace /tmp/ws --harness fake --sandbox local)
    $ winch run start "$RUN_ID"
    $ psql "$WINCH_DATABASE_URL" -c "UPDATE runs
        SET supervisor_lease_owner = 'demo-replacement',
            supervisor_lease_token = gen_random_uuid(),
            supervisor_lease_epoch = supervisor_lease_epoch + 1
        WHERE id = (SELECT id FROM runs ORDER BY created_at DESC LIMIT 1)"
    $ sleep 15 && winch run get "$RUN_ID"
    → expect: state still `running`; the daemon logs the refused write instead
      of completing a run it no longer owns

## Verification

- `make check` and `make test-integration` pass, including the new store tests
  against PostgreSQL.
- `make e2e` passes unchanged.
- Contract suites for fake harness and local sandbox still pass.

## Acceptance criteria

- [ ] After another owner takes a run's lease, the old owner's harness exit
  leaves the attempt `running` and appends no event.
- [ ] The attempt-state write refuses a stale owner, token, or epoch in both the
  PostgreSQL and memory stores.
- [ ] An execution that holds its lease still completes: the standing
  `create → start → poll events → get` scenario passes unchanged.
- [ ] Each of the three has a test that fails without the change.
- [ ] I1, I2, and I4 still hold.

## Deferrals

| Deferred | Owning task |
|---|---|
| Calling `Supervisor.Reconcile` at daemon startup | Phase 1 — the *Truthful terminal state for a run the daemon was stopped under* row in [`../phase-1/README.md`](../phase-1/README.md) |

## Traces to

- `docs/architecture.md` §2 (one writer per run) and §4 (the run supervisor
  leases execution ownership)
- `docs/contracts.md` §1 (run lifecycle)
- [`../post-mortems/2026-09-16-a-plan-defect-read-from-one-implementation.md`](../post-mortems/2026-09-16-a-plan-defect-read-from-one-implementation.md)
  — the audit finding and its reproduction
- P0-008 — the implementation this task revises
