# P0-008: Align the host and Docker Go gate options

**Phase:** 0 — Plan the implementation phases
**Shape:** capability
**Dependencies:** None

Resolves issue [#149](https://github.com/shaielc/code-winch/issues/149). This
Phase 0 meta task changes repository tooling; its observable capability is
reached through the Makefile targets and CI.

## Objective

`make check` (host) and `make test-cycle` (Docker) run the same Go checks,
including the composed-image e2e scenarios, so either option proves the Go gate.

## Scope

Baseline: `617a89e` on `main`, whose tracker and Phase 0 coverage identify this
task. The original PR base `0ebe55b` instead assigned this ID to child-status
planning and lacked `docs/testing.md`. At the corrected baseline, `AGENTS.md`
delegates verification to that file. The Makefile's `check` and `test-cycle`
targets show unequal gates; `test-cycle` also hides teardown-only failure.

- Define `format-check vet lint test build e2e` once, in an internal
  `go-checks` target used by both options. Enforce serial, fail-fast execution
  even with `make -j`; reuse existing recipes, `CHECKED_PACKAGES` and
  `E2E_TEST`.
- Wrap both options in setup → checks → teardown. Use `test-env` and
  `test-env-down`; attempt teardown whenever setup was attempted, including
  partial setup failure. Skip checks if setup fails. Preserve the first
  non-zero setup/check status; return teardown's status if those succeeded.
  A toolchain-image failure before setup need not tear down an unstarted env.
- Keep the existing host proxy/container service URLs. Pass `E2E_TEST` as one
  command-line assignment to the container's `make go-checks`, e.g.
  `"E2E_TEST=$(E2E_TEST)"` in the outer recipe. Do not forward the host's
  `WINCH_E2E_URL` or all of `MAKEFLAGS` into Docker.
- Install golangci-lint v2.1.6 in `deployments/Dockerfile.toolchain`, matching
  CI; keep `GOLANGCI_LINT=golangci-lint` usable. Preserve the focused
  `e2e`, `docker-e2e`, `test-env` and `test-env-down` building blocks.
- Update `docs/testing.md` and `deployments/README.md` with the shared check
  sequence, lifecycle and Docker requirement for either whole-gate option;
  remove the current-difference paragraph and its #149 link.
- Keep both serial CI executions, named for host-toolchain and shipped-image
  coverage. Two recurring full runs catch installation drift that this task's
  one-time parity demonstration cannot cover.
- Add a lightweight lifecycle regression harness under `test/` using the
  existing Python unittest runner; its command spies cover the matrix below.

## Required constraints

- Both options consume the single ordered check list.
- Test setup and cleanup affect only the branch's test project, never the
  operator's deployment or the shared proxy; topology is in `deployments/README.md`.
- Cleanup and failure precedence follow Scope; no failure is silently discarded.

## Non-goals

Product or web/control-panel changes, scenario content, container-free e2e,
proxy/Compose topology changes, and rewriting historical `docs/state.md` evidence.

## Runtime reachability

`make check`, `make test-cycle`, and `.github/workflows/go.yml`.

## Write set

- `Makefile`
- `deployments/Dockerfile.toolchain`
- `deployments/README.md`
- `docs/testing.md`
- `.github/workflows/go.yml`
- `test/test_go_gate.py` (lifecycle/command regression harness)

## Contract surfaces

The meaning of the Go gate in `docs/testing.md` and the two public make targets.
No product contract.

## Demonstration

After implementation, run each command independently from the repository root:

    $ make check
    $ make test-cycle
    → expect: the shared sequence runs in order; each command succeeds
    $ branch=$(git rev-parse --abbrev-ref HEAD)
    $ project=$(printf '%s' "$branch" | tr 'A-Z' 'a-z' | tr -c 'a-z0-9_-' '-')
    $ WINCH_TEST_PROJECT="$project" docker compose -f deployments/compose.yml \
        -f deployments/compose.test.yml -p "$project" ps -a -q
    → expect: no containers after each run (including stopped containers)

## Verification

These are implementation acceptance checks, not claims of passing at refinement.

- Compare `make -n check` and `make -n test-cycle`: both invoke `go-checks`.
  The regression harness must also assert the actual six-command order and
  fail-fast behavior with `make -j4`.
- For both public targets, run `python3 -m unittest discover -s test` with
  spies substituted via `MAKE`, `COMPOSE`, `COMPOSE_TEST` and `COMPOSE_PROXY`.
  The harness runs the real Makefile wrapper, logs calls, and makes the setup,
  check and teardown commands return the following injected codes:

  | Setup | Checks | Teardown | Expected calls | Wrapper status |
  |---|---|---|---|---|
  | 0 | 0 | 0 | setup, checks, teardown | 0 |
  | 41 | — | 0 | setup, teardown | 41 |
  | 41 | — | 43 | setup, teardown | 41 |
  | 0 | 42 | 0 | setup, checks, teardown | 42 |
  | 0 | 42 | 43 | setup, checks, teardown | 42 |
  | 0 | 0 | 43 | setup, checks, teardown | 43 |

  A shell spy records the actual recipe's status: outer GNU make returns 2
  on recipe failure, which cannot prove precedence. Stub image build success;
  separately fail it and assert setup/checks were skipped. Spies prove command
  flow; real runs below prove container integration.
- Lint injection: add a temporary gofmt-formatted file under `test/e2e` with
  its package declaration and `func gateLintFailure() {}`. Run each option;
  expect unused-function lint failure. Remove the file before e2e checks.
- E2E injection: run `make check E2E_TEST=false` and
  `make test-cycle E2E_TEST=false`; expect failure at e2e. The Docker command spy
  must receive one `E2E_TEST=false` argument alongside `go-checks`.
  After each lint/e2e failure, the Demonstration query must show no containers.
- Before the successful Docker demonstration, build with
  `WINCH_TEST_PROJECT="$project" docker compose -f deployments/compose.yml -f deployments/compose.test.yml -p "$project" --profile test build --no-cache toolchain`;
  run the same Compose prefix with `--profile test run --rm toolchain golangci-lint version`
  and expect v2.1.6. With one `make test-env`, run `make e2e` and
  `make docker-e2e`, then `make test-env-down`.
- Search `rg "make check|test-cycle|lint" AGENTS.md docs skills deployments .github Makefile`;
  reconcile conflicting current gate descriptions within scope.

## Acceptance criteria

- [ ] Both options run the single shared sequence serially, including Docker
      lint and host e2e; lint/e2e failures stop both and leave no test containers.
- [ ] The lifecycle matrix passes, including setup failure, teardown-only
      failure and competing failures.
- [ ] The shipped image has the pinned linter; focused e2e commands still work.
- [ ] Testing/deployment docs describe equivalent options; CI explicitly covers
      both toolchains and no longer describes #149 as a current difference.

## Deferrals

None.

## Traces to

Issue #149 and review threads on #148; `docs/testing.md`;
`deployments/README.md`; `skills/shared/workplan-model.md` (Phase 0 meta tasks).
