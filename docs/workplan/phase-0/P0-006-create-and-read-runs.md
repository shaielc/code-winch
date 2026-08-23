# P0-006: Create and read runs

**Phase:** 0 — Foundation repair
**Shape:** seam
**Dependencies:** P0-001 (semantic: this task's durability claim is only checkable in CI once the Go workflow has a PostgreSQL service; without it the storage tests skip)

## Objective

A person can create a run and read it back through the daemon API and operator CLI;
the run is persisted but does not execute.

## Scope

- Construct `postgres.Store` at daemon startup and wire it as `RunRepository`.
- Implement application logic for `CreateRun` and `GetRun`, mapping to the domain
  state machine's initial `created` state.
- Replace `unavailableBackend` with a delegating `httpapi.Backend` that handles
  only create and get; other methods return a stable, content-free error naming
  the owning task ID.
- Introduce the composition-root pattern later tasks extend (store handle,
  delegating backend, shared clock/ID sources) without starting processes or
  background workers.
- Add `winch run create` and `winch run get`.
- Create `test/e2e/`, a `make e2e` target, and the shared scenario harness later
  scenarios reuse: daemon startup, database setup, and an API client.
- Add the **`create → get`** scenario in `test/e2e/create_test.go`: create a run
  through the daemon API, read it back, and assert it is persisted in state
  `created` with no events.

## Non-goals

- Supervisor, runner, harness/sandbox drivers, outbox worker, or
  `EventStream.Publish`.
- `StartRun`, `SendRunInput`, `StopRun`, or event listing.
- **Any e2e step beyond `create → get`.** Start, events, input, stream, and stop
  are each added by the task that introduces them (P0-008 through P0-011);
  the assembled round trip and the CI gate are P0-007.
- Browser session cookies.

## Runtime reachability

- **Composition root:** `cmd/winchd`.
- **Profile:** PostgreSQL storage; no harness execution.
- **Command:** `winch run create`, `winch run get`.

## Write set

- `cmd/winchd/main.go`
- `internal/application/` (run create/get use cases)
- `cmd/winch/` (run create/get subcommands)
- `test/e2e/` (new directory: shared scenario harness and `create_test.go`)
- `Makefile` (`make e2e` target)
- Tests for create/get persistence and API mapping

## Contract surfaces

- API: `POST /api/v1/runs`, `GET /api/v1/runs/{runId}`
- port: partial `httpapi.Backend` (create and get only)

## Demonstration

    $ winch run create --workspace /tmp/ws --harness fake --sandbox local
    → expect: prints a run ID

    $ winch run get $RUN_ID
    → expect: state `created`, no events, no child processes

    $ winch run start $RUN_ID
    → expect: stable error naming P0-008 as owner

## Verification

- `make check` passes.
- `make test-integration` passes with `PG_TEST_DATABASE_URL` set.
- `make e2e` passes locally through the create/get steps (daemon + PostgreSQL
  required).
- `unavailableBackend` is removed; create/get no longer return 500 for binding.

## Acceptance criteria

- [ ] `POST /api/v1/runs` returns 201 and the run row exists in PostgreSQL.
- [ ] `GET /api/v1/runs/{runId}` returns the persisted run.
- [ ] `winch run create` and `winch run get` exercise the same behavior.
- [ ] `test/e2e/` exists and `make e2e` runs the `create → get` scenario.
- [ ] Stubbed backend methods name P0-008, P0-009, P0-010, or P0-011 as owners.
- [ ] I1 and I2 still hold.

## Deferrals

| Deferred | Owning task |
|---|---|
| Start execution | P0-008 |
| Input | P0-009 |
| Outbox worker | P0-009 |
| Live WebSocket stream | P0-010 |
| Stop | P0-011 |

## Traces to

- `docs/state.md` §*Twenty-four tasks completed against a system that never ran
  a run* (`postgres.New` never called for runs)
- Invariant I1 — composition root reaches introduced behavior
