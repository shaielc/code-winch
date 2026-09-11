# P0-013: Revise create/read for memory store profile

**Phase:** 0 — Foundation repair
**Shape:** swap
**Dependencies:** P0-006 (revision: P0-006 wired postgres as the only store profile and the e2e harness assumes PostgreSQL), P0-012 (compile: memory profile wiring exists)

## Objective

Create and read runs through the daemon API and operator CLI use
`storeProfile=memory` as the default local and e2e profile, with no database
required for `make e2e` through the `create → get` step.

## Scope

- Revise `cmd/winchd` composition-root wiring from P0-006 so the memory profile
  is the default for local development and e2e, and postgres remains selectable.
- Revise the shared `test/e2e` harness P0-006 introduced: start `winchd` with
  `storeProfile=memory`, drop PostgreSQL setup from the harness, and keep an
  explicit switch for postgres-backed scenarios when needed.
- Revise `test/e2e/create_test.go` to assert persistence through the memory
  profile (in-process state, not SQL).
- Revise `winch run create` and `winch run get` demonstration docs if they
  still name PostgreSQL as required.
- Remove the semantic justification for **P0-006 → P0-001** from the e2e path:
  `make e2e` through create/get must pass without `PG_TEST_DATABASE_URL`.

## Non-goals

- Changing application use-case logic beyond store selection.
- Start, input, stream, stop, or the assembled round-trip scenario — P0-014
  through P0-018.
- Removing postgres integration tests or the postgres store profile.

## Runtime reachability

- **Composition root:** `cmd/winchd`.
- **Profile:** `storeProfile=memory`, `harnessProfile` unset (no execution).
- **Command:** `winch run create`, `winch run get`, `make e2e` (create/get only).

## Write set

- `cmd/winchd/main.go`
- `test/e2e/` (shared harness and `create_test.go`)
- `deployments/README.md` (if the default profile changes)
- Tests for create/get against the memory profile

## Contract surfaces

- configuration: default `store_profile` for local and e2e runs

## Demonstration

    $ make e2e -run TestCreate
    → expect: passes with no PostgreSQL process running

    $ winch run create --workspace /tmp/ws --harness fake --sandbox local
    $ winch run get $RUN_ID
    → expect: state `created`, persisted in the in-memory store

## Verification

- `make check` passes.
- `make e2e` passes locally through create/get without a database.
- `make test-integration` still passes with `PG_TEST_DATABASE_URL` set (postgres
  adapter unchanged).

## Acceptance criteria

- [ ] `POST /api/v1/runs` and `GET /api/v1/runs/{runId}` work with
      `storeProfile=memory`.
- [ ] `test/e2e/create_test.go` runs without PostgreSQL setup.
- [ ] `make e2e` through create/get does not require `PG_TEST_DATABASE_URL`.
- [ ] The postgres profile still works when explicitly selected.
- [ ] I1 and I2 still hold.

## Deferrals

| Deferred | Owning task |
|---|---|
| Start execution on memory profile | P0-014 |
| Postgres-backed e2e as I4 swap | Phase 1 — registered in [`../phase-1/README.md`](../phase-1/README.md) |

## Traces to

- Invariant I3 and I4 — `skills/shared/workplan-model.md`
- [`../post-mortems/2026-08-23-workplan-create-phase-0.md`](../post-mortems/2026-08-23-workplan-create-phase-0.md)
  defect 5
- P0-006 — postgres-only store wiring this task revises
