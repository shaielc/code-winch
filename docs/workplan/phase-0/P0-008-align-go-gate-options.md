# P0-008: Align the host and Docker Go gate options

**Phase:** 0 — Plan the implementation phases
**Shape:** capability
**Dependencies:** None

Resolves issue [#149](https://github.com/shaielc/code-winch/issues/149). This is
a Phase 0 meta task because it changes the repository's toolchain and gates, not
the product. The shape is the nearest fit: it adds observable behavior to the
maintained gate commands, and the model's reachability rule is met by the
`Makefile` targets and CI.

## Objective

`make check` (host) and `make test-cycle` (Docker) run the same Go checks —
format, vet, lint, unit and contract tests, build, and the composed-image e2e
scenarios — so that "I ran the Go gate" means the same thing for either option.

## Scope

Observed at refinement HEAD:

- `Makefile` defines `check: format-check vet lint test build`, while
  `test-cycle` invokes `make format-check vet test build e2e` in the toolchain.
  `GO_PACKAGES` omits `./test/e2e/...`; `CHECKED_PACKAGES` includes it for vet
  and lint. `E2E_TEST` runs that package separately.
- `test-env` creates the branch-named Compose project and shared proxy;
  `compose.test.yml` removes the sandbox's host port and gives the toolchain
  `WINCH_E2E_URL=http://sandbox:8080`. The host's default URL instead reaches
  the same sandbox through the loopback-only proxy. `main-up` uses the separate
  main Compose overlay and is not part of either path.
- `deployments/Dockerfile.toolchain` contains Go, Git, and Make but no
  golangci-lint. CI installs golangci-lint v2.1.6 before the host option and then
  runs both options serially.
- `make -n check` currently lists only the five host checks. `make -n
  test-cycle` lists the image build and lifecycle recipe whose inner command
  omits lint. A real run is not a safe pre-change demonstration because the two
  commands are known not to provide the promised parity.

Implement the alignment as follows:

- In `Makefile`, state the ordered list `format-check vet lint test build e2e`
  once and add one internal target that invokes that list. Both public options
  call that target: directly on the host for `check`, and inside `toolchain` for
  `test-cycle`. Continue using the existing `.golangci.yml`,
  `CHECKED_PACKAGES`, and `E2E_TEST`; do not duplicate package lists or recipes.
- Give both public options the same lifecycle wrapper: call `test-env`, run the
  shared target only after setup succeeds, and call `test-env-down` on every
  path after setup was attempted. Preserve the first non-zero setup/check
  status; if those succeeded, propagate a teardown failure. The host target
  uses the proxy URL already exported by the `Makefile`; the Docker target uses
  the service URL already supplied by `compose.test.yml`.
- Install golangci-lint v2.1.6 in `deployments/Dockerfile.toolchain` and keep
  `GOLANGCI_LINT=golangci-lint` usable in that image. Do not install Node or
  move web checks into the Go image.
- Keep `make e2e`, `make docker-e2e`, `make test-env`, and
  `make test-env-down` as explicit scenario/environment building blocks. They
  continue to assume or manage exactly what their current names document;
  parity applies to the two whole-gate options, not these focused targets.
- Update `docs/testing.md` and `deployments/README.md` to describe both whole
  options as equivalent, list all six ordered checks, state that either option
  now needs Docker for the isolated e2e environment, and remove the current
  difference and its #149 link.
- Keep both serial steps in `.github/workflows/go.yml`: the host step proves the
  locally installed toolchain option and the Docker step proves the shipped
  toolchain image, including its pinned linter. Rename the steps or add a short
  comment so that purpose is explicit; equivalence does not make either
  execution substrate untested.
- The refinement search found no other current gate-content statement needing
  a write. `AGENTS.md` delegates to `docs/testing.md`; `docs/state.md` records
  historical closure observations and must not be rewritten as if the new gate
  had already run. Re-run the search during implementation and change only a
  newly conflicting current statement.

## Required constraints

- Both options run the identical check list, stated once; a test or `make -n`
  comparison in the brief's verification shows it.
- The e2e step never runs in, replaces, or tears down the operator's deployment
  (`make main-up`): it uses the per-branch test project only.
- Teardown runs on failure, and the exit status is the first failure's.
- A teardown failure is returned when setup and checks succeeded; cleanup is
  not allowed to fail silently.
- No live provider account is required; CI keeps passing without new secrets.

## Non-goals

Web and control-panel gates, the e2e scenarios' content, making the e2e run
without any container, changes to the proxy or compose topology, and changing
the historical verification record in `docs/state.md`.

## Runtime reachability

`make check` and `make test-cycle` from the repository root, and
`.github/workflows/go.yml`. No product composition root changes.

## Write set

- `Makefile`
- `deployments/Dockerfile.toolchain`
- `deployments/README.md`
- `docs/testing.md`
- `.github/workflows/go.yml`

## Contract surfaces

The meaning of "the Go gate" in `docs/testing.md` and the commands `make check`
and `make test-cycle`. No product contract.

## Demonstration

    $ make -n check
    $ make -n test-cycle
    → expect: each expands the same shared ordered check target; the only
      difference is whether that target runs on the host or in `toolchain`
    $ make check
    $ make test-cycle
    $ branch=$(git rev-parse --abbrev-ref HEAD); project=$(printf '%s' "$branch" |
        tr 'A-Z' 'a-z' | tr -c 'a-z0-9_-' '-')
    → expect: each runs format-check, vet, lint, test, build, and e2e in that
      order and succeeds; after each, `docker compose -f deployments/compose.yml
      -f deployments/compose.test.yml -p "$project" ps -q` prints nothing

For the failure demonstration, add a temporary, gofmt-formatted Go file in
`test/e2e` containing an unused unexported function, run each option separately,
remove the file between runs, and confirm both fail in `lint`. Then run each
option separately with `E2E_TEST=false`; confirm both return non-zero at `e2e`.
After every failure, run the `docker compose ... ps -q` command above and expect
no output. Remove the temporary file before committing.

## Verification

- `make -n check` and `make -n test-cycle` show both paths invoking the single
  ordered Go-check target.
- `make check` and `make test-cycle` pass independently; after each, the
  branch-named test project has no containers.
- Negative case: a temporary unused function under `test/e2e` makes both
  options fail specifically at lint; `E2E_TEST=false` makes both fail at e2e;
  all four runs leave the branch-named test project with no containers.
- Rebuild the toolchain image without cache and verify
  `golangci-lint version` in it reports v2.1.6 before the Docker-option run.
- Run `make e2e` and `make docker-e2e` against one environment started by
  `make test-env`, then run `make test-env-down`.
- Run `rg "make check|test-cycle|lint" AGENTS.md docs skills deployments
  .github Makefile` and confirm current descriptions agree. Say which commands
  could not run in the implementation environment.

## Acceptance criteria

- [ ] Both options run the same checks, defined in one place in the `Makefile`.
- [ ] Lint runs inside the Docker option; e2e runs in the host option.
- [ ] A lint failure and an e2e failure each fail both options, and teardown
      still occurs; setup/check failure remains the returned status, while a
      teardown-only failure is not hidden.
- [ ] `docs/testing.md`, `deployments/README.md` and CI describe the options as
      equivalent and no longer cite #149 as a current difference; CI exercises
      both execution substrates.
- [ ] `make e2e`/`make docker-e2e` against a pre-started environment still work.

## Deferrals

None.

## Traces to

Issue #149 and the review threads on #148; `docs/testing.md`;
`deployments/README.md`; `skills/shared/workplan-model.md` (Phase 0 meta tasks).
