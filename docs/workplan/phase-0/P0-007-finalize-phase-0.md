# P0-007: Finalize phase 0

**Phase:** 0 — Foundation repair
**Shape:** hardening
**Dependencies:** P0-001 (semantic: the CI job needs the PostgreSQL service before `make e2e` can run there), P0-009 (semantic: the input step must exist before the assembled scenario can assert it), P0-010 (semantic: the stream step must exist before the assembled scenario can assert it), P0-011 (semantic: the stop step must exist before the assembled scenario can assert it)

## Objective

CI refuses any change that breaks the complete `create → start → stream → input →
stop` round trip against the fake profile.

## Scope

- Add the **`create → start → stream → input → stop`** scenario in
  `test/e2e/roundtrip_test.go`, driving all five operations against one run. The
  per-operation scenarios P0-006 and P0-008 through P0-011 contributed stay as
  they are; this one asserts what none of them can: one run identity throughout,
  gap-free sequences across the whole scenario, live stream and polled events
  agreeing, and a terminal state with no `fake-harness` descendant.
- Add the `make e2e` step to `.github/workflows/go.yml`, reusing the PostgreSQL
  service P0-001 added.
- Update `deployments/README.md` so the run routes are described as bound, with
  the fake profile's limits stated.

## Non-goals

- New API operations or CLI subcommands, and any behavior not already reachable.
  This task adds one scenario over operations that all exist.
- Editing the per-operation scenario files. If one of them is wrong, that is a
  revision task against its owning brief.
- Repairing a step that does not pass. A defect found here is a revision task
  against the owning brief, not silent scope.
- Phase 1 planning.

## Runtime reachability

- **Composition root:** `cmd/winchd` with the fake harness and local sandbox
  registered, driven by `test/e2e/` against PostgreSQL.
- **Profile:** `harnessProfile=fake`, `sandboxProfile=local`, PostgreSQL storage.
- **Command:** `make e2e`, locally and in the Go workflow.

## Write set

- `test/e2e/roundtrip_test.go`
- `.github/workflows/go.yml`
- `deployments/README.md`

## Contract surfaces

None.

## Demonstration

    $ make e2e
    → expect: the full create → start → stream → input → stop scenario passes

    $ ps -eo pid,cmd | grep -c 'fake[-]harness'
    → expect: 0

Break one step on a scratch branch — return early from `StopRun`, for example —
and confirm the gate refuses it:

    $ make e2e
    → expect: nonzero exit naming the failed step

CI: the Go workflow log for this task's pull request shows the `make e2e` step
executed and passing.

## Verification

- `make check` and `make test-integration` pass.
- `make e2e` passes locally against a running daemon and PostgreSQL.
- The workflow run for this pull request shows `make e2e` executed, not skipped.

## Acceptance criteria

- [ ] `make e2e` drives create, start, stream, input, and stop in one scenario
  against one run.
- [ ] The scenario asserts gap-free ordering across the whole run and a terminal
  state with no surviving `fake-harness` process.
- [ ] `.github/workflows/go.yml` runs `make e2e` on every push and pull request.
- [ ] A deliberately broken step fails the gate; the observation is recorded in
  the pull request.
- [ ] `deployments/README.md` describes the run routes as bound.
- [ ] I4 holds: one standing scenario suite runs against the fake profile through
  the daemon's real entrypoint.
- [ ] I1 and I2 still hold.

## Deferrals

None.

## Traces to

- Invariant I4 — `skills/shared/workplan-model.md`
- `docs/state.md` §*Invariants that were never established* (no standing scenario
  suite; `make e2e` was never a target)
- `docs/roadmap.md` Phase 1 exit ("forced stop leaves no child processes")
