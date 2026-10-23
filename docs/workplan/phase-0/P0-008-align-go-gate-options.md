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

Today `make test-cycle` runs `make format-check vet test build e2e` in the
toolchain and has no lint. `make check` runs `format-check vet lint test build`
and has no e2e, which needs a started test environment (`test/e2e/sandbox.go`
fails when `WINCH_E2E_URL` is unset). Proposed approach, for the refiner to
confirm against code at dispatch:

- Define the Go check list once in the `Makefile` and have both options use it,
  so the lists cannot drift again.
- Install golangci-lint at the CI-pinned version (v2.1.6) in
  `deployments/Dockerfile.toolchain` and run `lint` in `make test-cycle`.
  `.golangci.yml` and the `CHECKED_PACKAGES` set (including `./test/e2e/...`)
  are shared, not duplicated.
- Make `make check` run the e2e scenarios by starting the branch's test
  environment, running them through the proxy URL, and always tearing the
  environment down, as `make test-cycle` does. The host option then needs Docker
  for the e2e step only; Go and golangci-lint still run on the host.
- Keep `make e2e`, `make docker-e2e`, `make test-env` and `make test-env-down`
  working as the way to run only the scenarios against a started environment.
- Update `docs/testing.md`, `deployments/README.md` and `.github/workflows/go.yml`
  together: the "currently differ" paragraph and its #149 link go away, the
  Docker description names lint, and CI states what each job adds now that the
  options are equivalent (the refiner decides whether CI keeps both jobs, and
  says why).
- Reconcile any other statement of the gate contents found by
  `rg "make check|test-cycle|lint"` over `AGENTS.md`, `docs/` and `skills/`.

## Required constraints

- Both options run the identical check list, stated once; a test or `make -n`
  comparison in the brief's verification shows it.
- The e2e step never runs in, replaces, or tears down the operator's deployment
  (`make main-up`): it uses the per-branch test project only.
- Teardown runs on failure, and the exit status is the first failure's.
- No live provider account is required; CI keeps passing without new secrets.

## Non-goals

Web and control-panel gates, the e2e scenarios' content, making the e2e run
without any container, and changes to the proxy or compose topology.

## Runtime reachability

`make check` and `make test-cycle` from the repository root, and
`.github/workflows/go.yml`. No product composition root changes.

## Write set

- `Makefile`
- `deployments/Dockerfile.toolchain`
- `deployments/README.md`
- `docs/testing.md`
- `.github/workflows/go.yml`
- `docs/state.md` (only if it states gate contents that change)

## Contract surfaces

The meaning of "the Go gate" in `docs/testing.md` and the commands `make check`
and `make test-cycle`. No product contract.

## Demonstration

    $ make -n check | grep -c .      # and: make -n test-cycle
    $ make check
    $ make test-cycle
    → expect: both list and run format-check, vet, lint, test, build and e2e
      in that order, and each leaves no test environment running
      (`docker compose -p <project> ps` is empty)
    $ introduce a lint error in a Go file under ./test/e2e
    → expect: both options fail at lint, and the environment is not left up

## Verification

- The comparison of the two options' check lists passes.
- Negative case: a lint violation fails both options; a failing e2e scenario
  fails both and still tears the environment down.
- `make test-cycle` builds the toolchain image from scratch with the pinned lint.
- Say which Go option(s) ran and what could not run here.

## Acceptance criteria

- [ ] Both options run the same checks, defined in one place in the `Makefile`.
- [ ] Lint runs inside the Docker option; e2e runs in the host option.
- [ ] A lint failure and an e2e failure each fail both options, and teardown
      still occurs.
- [ ] `docs/testing.md`, `deployments/README.md` and CI describe the options as
      equivalent and no longer cite #149 as a current difference.
- [ ] `make e2e`/`make docker-e2e` against a pre-started environment still work.

## Deferrals

None.

## Traces to

Issue #149 and the review threads on #148; `docs/testing.md`;
`deployments/README.md`; `skills/shared/workplan-model.md` (Phase 0 meta tasks).
