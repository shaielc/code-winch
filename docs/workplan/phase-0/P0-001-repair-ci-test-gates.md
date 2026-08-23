# P0-001: Repair CI test gates

**Phase:** 0 — Foundation repair
**Shape:** hardening
**Dependencies:** None

## Objective

CI proves PostgreSQL storage guarantees on every pull request, and the unit-test
gate scans only this repository's Go packages.

## Scope

- Add a PostgreSQL service (or equivalent) to `.github/workflows/go.yml` and run
  `make test-integration` (or an equivalent command) with `PG_TEST_DATABASE_URL`
  set.
- Narrow `make test` so `go test ./...` does not descend into
  `web/node_modules` (or any other vendored Go tree under `web/`).
- Document the CI integration step in `deployments/README.md` if the testing
  procedure changes.

## Non-goals

- Adding an end-to-end scenario suite (`test/e2e`, `make e2e`) — owned by
  P0-007.
- Changing integration test assertions beyond what is required to run them in CI.
- Running integration tests inside `make check` on developer machines without a
  database (keep the existing `make test-integration` split).

## Runtime reachability

- **Composition root:** none — this task changes gates only.
- **Profile:** CI (`ubuntu-latest` workflow) and local `make test` /
  `make test-integration`.
- **Command:** `make check` (unit leg), `make test-integration` (integration leg).

## Write set

- `.github/workflows/go.yml`
- `Makefile`
- `deployments/README.md` (if the testing procedure changes)

## Contract surfaces

None.

## Demonstration

    $ make test
    → expect: tests pass and no package path under web/node_modules appears in output

    $ PG_TEST_DATABASE_URL=postgres://winch@127.0.0.1:55432/winch_test?sslmode=disable \
      make test-integration
    → expect: ok  github.com/shaielc/code-winch/internal/adapters/postgres

Simulate CI locally (or observe a workflow run on a pull request for this task):

    → expect: the Go workflow runs storage integration tests and they pass

## Verification

- `make check` passes without a database.
- `make test-integration` passes with `PG_TEST_DATABASE_URL` set.
- `.github/workflows/go.yml` job log shows integration tests executed.

## Acceptance criteria

- [ ] `.github/workflows/go.yml` runs the PostgreSQL integration suite on every
  push and pull request.
- [ ] `make test` no longer includes packages under `web/node_modules`.
- [ ] All existing unit and integration tests pass unchanged.
- [ ] I1 and I2 still hold: `make build` succeeds and the compose stack
  description in `deployments/README.md` remains accurate.

## Deferrals

| Deferred | Owning task |
|---|---|
| End-to-end scenario suite and `make e2e` | P0-007 |

## Traces to

- Invariant I2 (deployment substrate proven in CI) — `skills/shared/workplan-model.md`
- `docs/state.md` §*Gates that do not cover what they appear to* (CI never runs
  integration suite; `make test` walks `node_modules`)
