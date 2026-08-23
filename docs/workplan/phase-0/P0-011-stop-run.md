# P0-011: Stop run

**Phase:** 0 — Foundation repair
**Shape:** seam
**Dependencies:** P0-008 (semantic: a running execution path must exist before stop is meaningful)

## Objective

A person can force-stop a running harness; the run reaches a terminal state and
leaves no child processes. The standing `make e2e` scenario is complete and
gates CI.

## Scope

- Implement `StopRun` through the supervisor and runner stop path.
- Add `winch run stop`.
- Extend the standing `make e2e` scenario with stop and child-reaping
  assertions.
- Wire `make e2e` into `.github/workflows/go.yml` (requires PostgreSQL service
  from P0-001 when that task has landed; this task adds the e2e step itself).
- Update `deployments/README.md` so run routes are described as fully bound.

## Non-goals

- Browser UI or session cookies.
- Daemon restart reconciliation demonstration (covered by later hardening).
- New CLI subcommands beyond stop.

## Runtime reachability

- **Composition root:** `cmd/winchd`.
- **Profile:** fake harness, local sandbox, PostgreSQL storage.
- **Command:** `winch run stop`, `make e2e`.

## Write set

- `internal/application/` (stop use case)
- `cmd/winch/` (`stop` subcommand)
- `test/e2e/` (final scenario step: stop)
- `.github/workflows/go.yml` (`make e2e` in CI)
- `deployments/README.md`
- Tests for stop escalation and child reaping

## Contract surfaces

- API: `POST /api/v1/runs/{runId}/stop`

## Demonstration

Use a P0-003 transcript that runs until stopped:

    $ winch run create ...
    $ winch run start $RUN_ID
    $ winch run stop $RUN_ID
    → expect: run reaches `cancelled` or `completed` per policy; harness exits

    $ ps -eo pid,cmd | grep -c 'fake[-]harness'
    → expect: 0

    $ make e2e
    → expect: full `create → start → stream → input → stop` scenario passes

## Verification

- `make check` and `make test-integration` pass.
- CI workflow log shows `make e2e` passing on pull requests.

## Acceptance criteria

- [ ] `StopRun` forces harness termination and persists a terminal state.
- [ ] No `fake-harness` descendant remains after stop.
- [ ] `winch run stop` demonstrates the behavior.
- [ ] `make e2e` runs the complete standing scenario locally and in CI.
- [ ] I4 holds: the end-to-end scenario suite exists against the all-fake profile.
- [ ] I1 and I2 still hold.

## Deferrals

| Deferred | Owning task |
|---|---|
| Browser `winch_session` establishment | Phase 1 (not yet planned) |

## Traces to

- `docs/roadmap.md` Phase 1 exit ("forced stop leaves no child processes")
- `docs/state.md` §*The run round trip*
- Invariant I4 — standing scenario suite completed iteratively through P0-006–P0-011
