# P0-004: Exercise complete fake run in CI

**Phase:** 0 — Foundation repair
**Shape:** hardening
**Dependencies:** P0-002 (semantic: CI must build `winch` before this test can invoke it)

## Objective

CI executes a complete fake harness run end to end, proving the criterion the
previous plan marked complete but never enforced.

## Scope

- Add an automated test (Go `exec` test, scripted workflow step, or dedicated
  package under `test/`) that builds or uses the shipped `winch` and
  `fake-harness` binaries and drives `winch dev run --harness fake --sandbox
  local` with a short stdin script through successful exit.
- Wire the test into the default CI path (`make check` or the workflow step added
  in P0-001) without requiring PostgreSQL or the daemon.
- Assert stop escalation: no `fake-harness` descendant remains after the run
  (mirror the manual check in `docs/state.md`).

## Non-goals

- Driving runs through the HTTP API or WebSocket (Phase 1).
- Replacing contract tests in `test/contract/` — this supplements them with a
  process-level check of the shipped binaries.
- Browser or `web/` testing.

## Runtime reachability

- **Composition root:** `cmd/winch` with `internal/runner/local` and
  `internal/adapters/harness/fake`.
- **Profile:** `harnessProfile=fake`, `sandboxProfile=local`.
- **Command:** `winch dev run --harness fake --sandbox local`.

## Write set

- `test/` or `internal/runner/local/` (new CI-oriented test file)
- `Makefile` and/or `.github/workflows/go.yml` (only if the test is not picked
  up by existing targets)

## Contract surfaces

None.

## Demonstration

    $ make check
    → expect: new test passes locally without `PG_TEST_DATABASE_URL`

The test log should show a successful fake run, for example:

    [started]
    fake harness ready: run_id=...
    ...
    [exit] successful=true code=OK

## Verification

- `make check` passes on a clean checkout without a database.
- CI workflow log includes the new test and it passes on pull requests.

## Acceptance criteria

- [ ] A Go test invokes the real `winch` and `fake-harness` binaries (not only
  in-process adapters) and completes a fake run.
- [ ] The test runs in CI on every push and pull request without extra manual
  setup.
- [ ] The test asserts no `fake-harness` process remains after exit.
- [ ] The false completion claim in the previous plan ("fake harness drives a
  complete run in CI") is true at HEAD.

## Deferrals

| Deferred | Owning task |
|---|---|
| Standing API scenario `create → start → stream → input → stop` | Phase 1 (not yet planned) |

## Traces to

- `docs/state.md` §*Statuses that do not hold at HEAD* (fake harness not exercised
  in CI)
- Invariant I3/I4 partial — process-level fake run proof before the API suite
  exists
