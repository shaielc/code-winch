# P0-012: Ship controllable in-memory store profile

**Phase:** 0 — Foundation repair
**Shape:** capability
**Dependencies:** None

## Objective

`winchd` boots with `storeProfile=memory`, wires `internal/adapters/memory` as
the durable ports, and needs no PostgreSQL process — restoring I3 for storage.

## Scope

- Add a `store_profile` configuration key (`memory` | `postgres`) with
  validation and documented defaults for local development.
- Introduce a composition-root helper that constructs the memory repositories
  (`RunRepository`, `EventStore`, `OutboxPublisher`, `SupervisorStore`, clock,
  and ID source) as one wired set, parallel to the existing postgres path.
- Teach `cmd/winchd` to skip the database pool and migrations when
  `storeProfile=memory`.
- Expose controllability for the memory profile: document how to drive
  `FailurePlan` injection at runtime (configuration or environment), and state
  what the profile does **not** prove (no cross-restart durability, no
  multi-instance consistency, no SQL semantics).
- Add unit tests that the memory profile constructs every port the run path will
  need once P0-006 lands.

## Non-goals

- Binding run use cases or replacing `unavailableBackend` — P0-006 and the
  revision tasks own that.
- Changing the postgres profile's behavior or removing it.
- Revising per-operation e2e scenarios — P0-013 through P0-018.
- Browser session cookies.

## Runtime reachability

- **Composition root:** `cmd/winchd`.
- **Profile:** `storeProfile=memory`; harness and sandbox unset until run binding
  lands.
- **Command:** `winchd` with `store_profile: memory` (or equivalent env var).

## Write set

- `internal/platform/config/` (`store_profile` key)
- `cmd/winchd/` (profile selection; may be a new file beside `main.go`)
- `internal/adapters/memory/` (only if a facade or wiring helper is needed)
- `deployments/README.md` (memory store profile limits and controls)
- Tests for profile selection and startup without a database

## Contract surfaces

- configuration: `store_profile`
- driver namespace: `storeProfile=memory`

## Demonstration

    $ WINCH_STORE_PROFILE=memory winchd
    → expect: listener starts, no database connection or migration log line,
      health/API routes respond

    $ WINCH_STORE_PROFILE=postgres winchd
    → expect: unchanged postgres startup (pool ping and schema check)

With failure injection configured (exact surface as implemented):

    $ WINCH_STORE_PROFILE=memory WINCH_MEMORY_INJECT_RUN_SAVE=not_found winchd
    → expect: documented injection path is reachable from configuration

## Verification

- `make check` passes.
- New tests cover memory-profile startup and postgres-profile regression.
- No change to `make test-integration` requirements (postgres adapter tests
  unchanged).

## Acceptance criteria

- [ ] `storeProfile=memory` reaches every memory port the run round trip needs.
- [ ] `winchd` starts without `database_url` when the memory profile is selected.
- [ ] Injectable failures and profile limits are documented and covered by tests.
- [ ] I3 holds for the in-memory store profile: supported, controllable, and
      honest about limits.
- [ ] I1 and I2 still hold for both profiles.

## Deferrals

| Deferred | Owning task |
|---|---|
| Run API and CLI on the memory profile | P0-013 |
| Default e2e harness to memory | P0-013 |

## Traces to

- Invariant I3 — `skills/shared/workplan-model.md`
- [`../post-mortems/2026-08-23-workplan-create-phase-0.md`](../post-mortems/2026-08-23-workplan-create-phase-0.md)
  defect 5
- `docs/state.md` §*What is not implemented* (`internal/adapters/memory` is
  test-only)
