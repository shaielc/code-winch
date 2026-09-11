# P0-012: Ship controllable in-memory store profile

**Phase:** 0 — Foundation repair
**Shape:** seam
**Dependencies:** P0-009 (contract: P0-009 defines the `OutboxPublisher` port surface and starts the outbox worker in the composition root, settling which durable ports the run path consumes)

## Objective

`winchd` boots with `storeProfile=memory` and serves every run operation the
daemon supports, from `internal/adapters/memory`, with no PostgreSQL process —
restoring I3 for storage.

## Objective context

The memory profile is a **supported way to run the product** (I3), held to the
same standard as the postgres profile for everything it claims. What it does not
claim is durability: it keeps nothing across a restart, so it is not the durable
single-process developer database that `docs/architecture.md` §5 reserves for
SQLite, and it supersedes no ADR. That is a limit on what it *proves*, not a
discount on what it has to *do* — per I3, every operation the running system
reaches must work on this profile or the task is not complete.

Most of the implementation exists. `internal/adapters/memory` holds
concurrency-safe `CreateRunRepository`, `RunRepository`, `EventStore`, and
`SupervisorStore`, each with a `FailurePlan`, written for use-case and contract
tests; nothing outside `_test.go` imports the package (`docs/state.md` §*What is
not implemented*). This task makes that code a selectable runtime profile and
closes the gaps that stop it from serving the whole run path.

## Why this runs after P0-009

The durable ports the run path consumes are not settled until P0-009 lands. It
starts `application.OutboxWorker` in the daemon lifecycle, adds the input path,
and **declares `port: OutboxPublisher` as its own contract surface**. A store set
wired before that either omits ports the run path will consume, or reaches into a
surface P0-009 owns — the earlier draft of this brief did the second, and listed
`OutboxPublisher` among the "durable ports" it wired. `OutboxPublisher`
(`internal/application/ports.go:83`) is the delivery side and a transport
concern; `OutboxStore` (`:97`) is the durable side. Only the second belongs to a
store profile, and P0-009 settles the first.

## Scope

- Add a `store_profile` configuration key (`memory` | `postgres`, default
  `postgres`) with validation, and require `database_url` only under `postgres`.
- Introduce a composition-root helper that selects one wired set of durable
  ports, parallel to the existing postgres path, over every durable port the run
  path consumes once P0-009 has landed: `CreateRunRepository`, `RunRepository`,
  `EventStore`, `SupervisorStore`, `InputCommandStore`, `OutboxStore`, `Clock`,
  and `IDSource`. Both profiles supply every port in the set from their own
  adapter. `OutboxPublisher` is **not** in the set — it is delivery, and P0-009
  owns it.
- Add the memory implementations the set is missing: `memory.OutboxStore` and
  `memory.InputCommandStore` do not exist, and once P0-009 puts those ports in
  the run path a profile without them does not work.
- Supply a deterministic clock and ID source fit for a long-running process.
  `memory.Clock` never advances on its own and `memory.IDSource` panics when its
  preloaded queue drains (`internal/adapters/memory/memory.go:420-450`); neither
  survives a daemon.
- Teach `cmd/winchd` to skip the database pool and migrations when
  `storeProfile=memory`.
- Make the profile controllable at runtime through a **memory scenario file**
  selected by configuration, covering the three I3 controls that apply to a
  store: failure injection over any port and operation in the set, latency, and a
  determinism seed that can be left unset for wall-clock time and random IDs. See
  *Where the controls live* below.
- State what the profile does **not** prove: no cross-restart durability, no
  multi-instance consistency, no SQL semantics, and no transactional outbox — the
  guarantee `docs/architecture.md` §6 requires of every event-publishing
  mutation. Memory satisfies `OutboxStore`'s signature, not its atomicity with
  the event append.
- Add unit tests for profile selection, for startup with no `database_url`, for
  the new ports, and for each of the three controls.

## Non-goals

- **Any change to `test/e2e/`.** The phase's e2e table allocates one scenario
  file per seam task and allocates none to this one; `create → get` on the memory
  profile is P0-013's, as is making memory the default for `make e2e`.
- Defining, wiring, or replacing `OutboxPublisher` — P0-009's contract surface.
- Changing the postgres profile's behavior, its default, or removing it.
- A durable single-process store. SQLite, if it is ever wanted, is a separate
  task against `docs/architecture.md` §5.
- Browser session cookies.

## Where the controls live

The daemon's configuration gains exactly two keys, because both are
composition-root decisions: `store_profile` (which adapter set the root wires)
and `memory_scenario` (where that profile's controls are read from). Everything
else — which operation fails with which error, how much latency, which seed —
lives in the scenario file.

Three reasons the controls do not become daemon configuration keys:

- `internal/platform/config` describes how to deploy the product. It should not
  grow a field per adapter debug knob, and cross-field validation of the form
  "only meaningful when another key has one particular value" is the signal that
  a key is in the wrong place.
- `internal/adapters/memory` already defines `FailurePlan` and the operation
  names it keys on, so it owns their spelling. A later task adding a port adds
  operations to the file, with no new configuration key and no new contract
  surface to collide over.
- I3 names "scripted transcripts or scenario files selected at runtime" as the
  controllability mechanism, and P0-003 set the precedent by putting the fake
  harness's controls on the harness's own surface rather than in the daemon's
  configuration.

An illustrative shape; the task fixes the schema:

```yaml
seed: 1          # omit for wall-clock time and random IDs
latency: 250ms   # applied to every memory port call
failures:        # consumed in order, per operation
  - run.save: conflict
  - run.get: not_found
```

## Runtime reachability

- **Composition root:** `cmd/winchd`.
- **Profile:** `storeProfile=memory`, with the fake harness and local sandbox
  that P0-003 and P0-009 establish.
- **Command:** `winchd` with `store_profile: memory` (or `WINCH_STORE_PROFILE`),
  driven by hand through `winch run create`, `run get`, `run start`, and
  `run input`.

## Write set

- `internal/platform/config/` (`store_profile`, `memory_scenario`)
- `cmd/winchd/` (profile selection; may be a new file beside `main.go`)
- `internal/adapters/memory/` (`OutboxStore`, `InputCommandStore`, deterministic
  clock and ID source, scenario file loading, latency hook)
- `internal/adapters/memory/README.md` (controls and limits)
- `deployments/README.md` (memory store profile limits and controls)
- Tests for profile selection, database-free startup, the new ports, and each
  control

Write collision with P0-013 and P0-014 on `cmd/winchd/main.go`. No overlap with
`test/e2e/`.

## Contract surfaces

- configuration: `store_profile`
- configuration: `memory_scenario` — valid only when `store_profile=memory`,
  rejected otherwise
- schema: the memory scenario file, owned by `internal/adapters/memory`
- port: `InputCommandStore` and `OutboxStore` memory implementations
- driver namespace: `storeProfile=memory`

## Demonstration

    $ WINCH_STORE_PROFILE=memory winchd
    → expect: listener starts, no database connection or migration log line,
      GET /api/v1/health returns {"status":"ok"}

    $ WINCH_STORE_PROFILE=postgres winchd
    → expect: unchanged postgres startup (pool ping and schema check, status
      applied then current)

Against the memory daemon, a person drives the run path by hand through the
maintained CLI and gets the same observable result as on postgres:

    $ winch run create --workspace /tmp/ws --harness fake --sandbox local
    $ winch run start <RUN_ID>
    $ winch run input <RUN_ID> --text hello
    $ winch run get <RUN_ID>
    → expect: the run reads back with the harness response recorded and the
      outbox backlog drained to zero — the same assertions P0-009's scenario
      makes against PostgreSQL

Each control, driven from a scenario file:

    $ cat > /tmp/fail.yaml <<'Y'
    failures:
      - run.save: conflict
    Y
    $ WINCH_STORE_PROFILE=memory WINCH_MEMORY_SCENARIO=/tmp/fail.yaml winchd
    $ winch run create --workspace /tmp/ws --harness fake --sandbox local
    $ winch run create --workspace /tmp/ws --harness fake --sandbox local
    → expect: the first create fails with the injected error, the second
      succeeds — the plan is consumed, so a person retries without restarting

    $ printf 'latency: 250ms\n' > /tmp/slow.yaml
    $ WINCH_STORE_PROFILE=memory WINCH_MEMORY_SCENARIO=/tmp/slow.yaml winchd
    $ winch run get <RUN_ID>
    → expect: the read visibly takes at least the configured delay

    $ printf 'seed: 1\n' > /tmp/seeded.yaml
    $ WINCH_STORE_PROFILE=memory WINCH_MEMORY_SCENARIO=/tmp/seeded.yaml winchd
    $ winch run create --workspace /tmp/ws --harness fake --sandbox local
    → expect: the same run ID and createdAt across restarts with the same seed,
      and different values with no scenario file

## Verification

- `make check` passes.
- New unit tests cover profile selection, startup with no `database_url`, the
  memory `InputCommandStore` and `OutboxStore`, and failure, latency, and
  determinism control.
- `make e2e` still passes **unchanged** against PostgreSQL, with the profile seam
  in the path — this task adds no scenario and revises none.
- No change to `make test-integration` requirements (postgres adapter tests
  unchanged).

## Acceptance criteria

- [ ] `storeProfile=memory` reaches every port in the declared set, and each one
      is read by the code that uses it — a port constructed into a field nothing
      consumes is not reached.
- [ ] Both profiles supply every port in the set from their own adapter; neither
      profile is backed by the other's implementation.
- [ ] Every run operation the daemon supports at this point — create, get, start,
      input, and the outbox drain — works on the memory profile and returns the
      same errors the postgres profile returns. No operation is missing, stubbed,
      or degraded under the memory profile (I3).
- [ ] `winchd` starts without `database_url` when the memory profile is selected,
      and logs no database or migration line.
- [ ] Failure, latency, and determinism are each drivable from the scenario file
      without editing source, and each is covered by a test.
- [ ] I3 holds for the in-memory store profile: supported, controllable over the
      three controls that apply to a store, honest about its limits — the absent
      transactional-outbox atomicity among them — and working for everything it
      claims.
- [ ] I1 and I2 still hold for both profiles.
- [ ] The daemon's configuration gained no key outside *Contract surfaces*.

## Deferrals

| Deferred | Owning task |
|---|---|
| `create → get` e2e scenario on the memory profile, and the shared e2e harness | P0-013 |
| Memory profile as the default for local development and `make e2e` | P0-013 |
| WebSocket stream and stop on the memory profile | P0-016, P0-017 |
| `make e2e` gating CI without a database | P0-018 |

## Traces to

- Invariant I3 — `skills/shared/workplan-model.md`
- `docs/state.md` §*What is not implemented* (`internal/adapters/memory` is
  test-only; `postgres.New` is never called)
- `docs/architecture.md` §5 (SQLite, not this profile, is the durable
  single-process developer option) and §6 (transactional outbox)
- P0-009 §*Contract surfaces* (`port: OutboxPublisher`)
