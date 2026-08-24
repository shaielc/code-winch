# P0-018: Revise phase 0 closure for memory-backed e2e

**Phase:** 0 — Foundation repair
**Shape:** hardening
**Dependencies:** P0-007 (revision: P0-007 wired `make e2e` into CI against PostgreSQL and named postgres under *Runtime reachability*), P0-016 (semantic: stream step must pass on memory before the assembled scenario can assert it), P0-017 (semantic: stop step must pass on memory before the assembled scenario can assert it), P0-012 (contract: `storeProfile=memory` namespace)

## Objective

CI refuses any change that breaks the complete `create → start → stream → input →
stop` round trip on the all-fake profile with `storeProfile=memory` and no
PostgreSQL service — closing I4's first rung.

## Scope

- Revise `test/e2e/roundtrip_test.go` from P0-007 to run against the
  memory-backed harness.
- Revise `.github/workflows/go.yml` so the `make e2e` step does **not** require
  a PostgreSQL service; keep postgres integration tests on the separate
  `make test-integration` leg from P0-001.
- Revise `deployments/README.md` to name memory as the default store profile for
  local development and e2e, with postgres documented as the swap substrate.
- Confirm per-operation scenario files from P0-013 through P0-017 are unchanged
  except where a revision task already touched them; this task owns only the
  assembled scenario and CI wiring.

## Non-goals

- New API operations or CLI subcommands.
- Editing per-operation scenario files — defects there are revision tasks
  against their owning brief.
- Removing postgres integration tests or the postgres store profile.
- Phase 1 planning.

## Runtime reachability

- **Composition root:** `cmd/winchd` with fake harness and local sandbox,
  `storeProfile=memory`.
- **Profile:** all-fake — memory store, fake harness, local sandbox.
- **Command:** `make e2e`, locally and in the Go workflow.

## Write set

- `test/e2e/roundtrip_test.go`
- `.github/workflows/go.yml`
- `deployments/README.md`

## Contract surfaces

None.

## Demonstration

    $ make e2e
    → expect: full create → start → stream → input → stop passes with no
      PostgreSQL process running

    $ ps -eo pid,cmd | grep -c 'fake[-]harness'
    → expect: 0

CI: the Go workflow log shows `make e2e` executed without a database service and
passing.

## Verification

- `make check` passes.
- `make e2e` passes locally without a database.
- `make test-integration` still passes with `PG_TEST_DATABASE_URL` in CI.
- The workflow run for this pull request shows `make e2e` executed, not skipped.

## Acceptance criteria

- [ ] `make e2e` drives all five operations on `storeProfile=memory` in one
      scenario.
- [ ] `.github/workflows/go.yml` runs `make e2e` without a PostgreSQL service.
- [ ] Gap-free ordering, stream/poll agreement, and terminal state with no
      surviving `fake-harness` process are asserted.
- [ ] I4 holds: the standing suite's first rung is the all-fake profile including
      memory storage.
- [ ] I1 and I2 still hold.

## Deferrals

| Deferred | Owning task |
|---|---|
| Same round trip against `storeProfile=postgres` (I4 swap) | Phase 1 — registered in [`../phase-1/README.md`](../phase-1/README.md) |

## Traces to

- Invariant I4 — `skills/shared/workplan-model.md`
- [`../post-mortems/2026-08-23-workplan-create-phase-0.md`](../post-mortems/2026-08-23-workplan-create-phase-0.md)
  defect 5
- P0-007 — postgres-backed CI gate this task revises
