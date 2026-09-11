# P0-012: Ship controllable in-memory store profile

**Phase:** 0 — Foundation repair
**Shape:** seam
**Dependencies:** None

## Objective

`winchd` boots with `storeProfile=memory`, serves the run API from the existing
`internal/adapters/memory` adapters, and needs no PostgreSQL process — restoring
I3 for storage.

## Objective context

The adapters this profile needs already exist. `internal/adapters/memory` holds
concurrency-safe implementations of `CreateRunRepository`, `RunRepository`,
`EventStore`, and `SupervisorStore`, each with a `FailurePlan`, written for
use-case and contract tests. Nothing outside `_test.go` imports the package
(`docs/state.md` §*What is not implemented*). This task makes that existing code
a selectable runtime profile; it does not write a new store.

The profile is a **fake, not a persistence substrate**. It makes no durability
claim, so it is not the single-process developer database that
`docs/architecture.md` §5 reserves for SQLite, and it supersedes no ADR. What it
replaces is a test double's invisibility, which is I3's subject.

## Scope

- Add a `store_profile` configuration key (`memory` | `postgres`, default
  `postgres`) with validation, and require `database_url` only under
  `postgres`.
- Introduce a composition-root helper that selects one wired set of durable
  ports, parallel to the existing postgres path, over exactly the ports the run
  path consumes today: `CreateRunRepository`, `RunRepository`, `EventStore`,
  `SupervisorStore`, `Clock`, and `IDSource`. Both profiles must be able to
  supply every port in the set — see *Port asymmetry* below.
- Teach `cmd/winchd` to skip the database pool and migrations when
  `storeProfile=memory`.
- Make the profile controllable at runtime, over the three I3 controls that
  apply to a store:
  - **failure injection** — drive the existing `FailurePlan` on any port in the
    set, for any operation, from one configuration key;
  - **latency** — a delay applied to every memory port call;
  - **determinism** — a seed that fixes the clock's start and makes IDs a
    reproducible sequence, and the ability to leave it unset for wall-clock time
    and random IDs.
- Supply the deterministic clock and ID source the seed needs.
  `memory.Clock` never advances on its own and `memory.IDSource` panics when its
  preloaded queue drains (`internal/adapters/memory/memory.go:420-450`); neither
  survives a long-running process. Both live in `internal/adapters/memory`
  beside the adapters they pair with.
- State what the profile does **not** prove: no cross-restart durability, no
  multi-instance consistency, no SQL semantics, and **no transactional outbox**
  — the guarantee `docs/architecture.md` §6 requires of every event-publishing
  mutation.
- Add unit tests for profile selection, for startup with no `database_url`, and
  for each of the three controls.

## Non-goals

- **Any change to `test/e2e/`.** The e2e suite allocates one scenario file per
  seam task and allocates none to this one; `create → get` on the memory
  profile is P0-013's, as is making memory the default for `make e2e`. See
  *Deferrals*.
- Binding run use cases or replacing `unavailableBackend` — P0-006 and the
  revision tasks own that.
- Changing the postgres profile's behavior, its default, or removing it.
- Adding `OutboxStore` or `InputCommandStore` to the memory package — P0-015.
- Browser session cookies.

## Port asymmetry

The two adapter sets are not parallel at HEAD, which constrains the set above.
Asserted by interface assertion against both adapters:

| Port | `memory` | `postgres.Store` |
|---|---|---|
| `CreateRunRepository` | yes | yes |
| `RunRepository` | yes | yes |
| `EventStore` | yes | yes |
| `SupervisorStore` | yes | yes |
| `OutboxPublisher` | yes | **no** |
| `OutboxStore` | **no** | yes |
| `InputCommandStore` | **no** | yes |

`OutboxPublisher` (`internal/application/ports.go:83`) is the delivery side and
a transport concern; `OutboxStore` (`:97`) is the durable side. A store set must
declare `OutboxStore`, never `OutboxPublisher` — putting the publisher in it
forces the postgres branch to wire a memory object, because `postgres.Store` has
no `Publish` method and should not grow one.

Neither `OutboxStore` nor `InputCommandStore` exists in `internal/adapters/memory`,
and nothing in the run path consumes either at HEAD. They are therefore out of
the set and out of this task; P0-015 adds them when `run input` needs them.

## Runtime reachability

- **Composition root:** `cmd/winchd`.
- **Profile:** `storeProfile=memory`; harness and sandbox unset until run
  binding lands.
- **Command:** `winchd` with `store_profile: memory` (or `WINCH_STORE_PROFILE`),
  driven by hand through `winch run create` and `winch run get`.

## Write set

- `internal/platform/config/` (`store_profile` and the three memory control keys)
- `cmd/winchd/` (profile selection; may be a new file beside `main.go`)
- `internal/adapters/memory/` (deterministic clock and ID source; latency hook)
- `internal/adapters/memory/README.md` (controls and limits)
- `deployments/README.md` (memory store profile limits and controls)
- Tests for profile selection, database-free startup, and each control

Write collision with P0-006 and P0-013 on `cmd/winchd/main.go`. No overlap with
`test/e2e/`.

## Contract surfaces

- configuration: `store_profile`
- configuration: `memory_failures`, `memory_latency`, `memory_seed` — valid only
  when `store_profile=memory`, rejected otherwise
- driver namespace: `storeProfile=memory`

`memory_failures` takes `operation=error` pairs so one key covers every port in
the set and every operation on it, rather than one key per operation.

## Demonstration

    $ WINCH_STORE_PROFILE=memory winchd
    → expect: listener starts, no database connection or migration log line,
      GET /api/v1/health returns {"status":"ok"}

    $ WINCH_STORE_PROFILE=postgres winchd
    → expect: unchanged postgres startup (pool ping and schema check, status
      applied then current)

A person drives the profile and each control by hand through the maintained CLI
against that running daemon. `winch run create` and `winch run get` come from
P0-006, which is already completed, so this task carries no edge for them:

    $ winch run create --workspace /tmp/ws --harness fake --sandbox local
    $ winch run get <RUN_ID>
    → expect: the created run reads back with state "created" from memory

    $ WINCH_STORE_PROFILE=memory WINCH_MEMORY_FAILURES=run.save=conflict winchd
    $ winch run create --workspace /tmp/ws --harness fake --sandbox local
    $ winch run create --workspace /tmp/ws --harness fake --sandbox local
    → expect: the first create fails with the injected error, the second
      succeeds — the plan is consumed, so a person retries without restarting

    $ WINCH_STORE_PROFILE=memory WINCH_MEMORY_LATENCY=250ms winch run get <RUN_ID>
    → expect: the read visibly takes at least the configured delay

    $ WINCH_STORE_PROFILE=memory WINCH_MEMORY_SEED=1 winchd
    $ winch run create --workspace /tmp/ws --harness fake --sandbox local
    → expect: the same run ID and createdAt across restarts with the same seed,
      and different values with the seed unset

## Verification

- `make check` passes.
- New unit tests cover profile selection, startup with no `database_url`, and
  failure, latency, and determinism control.
- `make e2e` still passes **unchanged** against PostgreSQL, with the profile
  seam in the path — this task adds no scenario and revises none.
- No change to `make test-integration` requirements (postgres adapter tests
  unchanged).

## Acceptance criteria

- [ ] `storeProfile=memory` reaches every port in the declared set, and each one
      is read by the code that uses it — a port constructed into a field nothing
      consumes is not reached.
- [ ] Both profiles supply every port in the set from their own adapter; neither
      profile is backed by the other's implementation.
- [ ] `winchd` starts without `database_url` when the memory profile is
      selected, and logs no database or migration line.
- [ ] Failure, latency, and determinism are each drivable from configuration
      without editing source, and each is covered by a test.
- [ ] I3 holds for the in-memory store profile: supported, controllable over the
      three controls that apply to a store, and honest about its limits — the
      absent transactional outbox among them.
- [ ] I1 and I2 still hold for both profiles.
- [ ] No configuration key outside *Contract surfaces* was added.

## Deferrals

| Deferred | Owning task |
|---|---|
| `create → get` e2e scenario on the memory profile, and the shared e2e harness | P0-013 |
| Memory profile as the default for local development and `make e2e` | P0-013 |
| Run API and CLI beyond create/get on the memory profile | P0-014, P0-016, P0-017 |
| `memory.OutboxStore` and `memory.InputCommandStore` | P0-015 |
| `make e2e` gating CI without a database | P0-018 |

## Traces to

- Invariant I3 — `skills/shared/workplan-model.md`
- `docs/state.md` §*What is not implemented* (`internal/adapters/memory` is
  test-only; `postgres.New` is never called)
- `docs/architecture.md` §5 (SQLite, not this profile, is the durable
  single-process developer option) and §6 (transactional outbox)
