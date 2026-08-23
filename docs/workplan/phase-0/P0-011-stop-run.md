# P0-011: Stop run

**Phase:** 0 — Foundation repair
**Shape:** seam
**Dependencies:** P0-010 (semantic: the full observe path must exist before forced stop is demonstrated end to end)

## Objective

A person can force-stop a running harness; the run reaches a terminal state and
leaves no child processes.

## Scope

- Implement `StopRun` through the supervisor and runner stop path.
- Add `winch run stop`.
- Update `deployments/README.md` so run routes are described as fully bound.

## Non-goals

- Browser UI or session cookies.
- Daemon restart reconciliation demonstration (covered by later hardening).
- New CLI subcommands beyond stop.

## Runtime reachability

- **Composition root:** `cmd/winchd`.
- **Profile:** fake harness, local sandbox, PostgreSQL storage.
- **Command:** `winch run stop`.

## Write set

- `internal/application/` (stop use case)
- `cmd/winch/` (`stop` subcommand)
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

## Verification

- `make check` and `make test-integration` pass.

## Acceptance criteria

- [ ] `StopRun` forces harness termination and persists a terminal state.
- [ ] No `fake-harness` descendant remains after stop.
- [ ] `winch run stop` demonstrates the behavior.
- [ ] All `httpapi.Backend` methods are implemented; no stub deferrals remain
  for run operations.
- [ ] I1 and I2 still hold.

## Deferrals

| Deferred | Owning task |
|---|---|
| Browser `winch_session` establishment | Phase 1 (not yet planned) |

## Traces to

- `docs/roadmap.md` Phase 1 exit ("forced stop leaves no child processes")
- `docs/state.md` §*The run round trip*
