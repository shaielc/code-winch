# P0-008: Align the host and Docker Go gate options

**Phase:** 0 — Plan the implementation phases
**Shape:** capability
**Dependencies:** None

Resolves issue [#149](https://github.com/shaielc/code-winch/issues/149).
This Phase 0 meta task changes repository tooling, reached through the
Makefile targets and CI.

## Objective

`make check` (host) and `make test-cycle` (Docker) run the same Go checks:
format, vet, lint, unit and contract tests, build, and composed-image e2e.

## Scope

The Makefile's `check` target omits e2e; `test-cycle` omits lint.
`deployments/Dockerfile.toolchain` also lacks golangci-lint.

- Define `format-check vet lint test build e2e` once in the Makefile and
  have both options use it. Reuse the existing recipes and package lists.
- Install golangci-lint v2.1.6, matching CI, in the toolchain image so the
  Docker option can run the existing lint target.
- Add the existing e2e scenarios to `make check`: start the branch's test
  environment with `test-env`, use its existing proxy URL, and tear it down
  with `test-env-down`, including when a check fails. The host option then
  requires Docker for the test environment.
- Preserve `make e2e`, `make docker-e2e`, `make test-env` and
  `make test-env-down` as focused commands with their existing responsibilities.
- Update `docs/testing.md` and `deployments/README.md` to describe equivalent
  options and remove the current-difference paragraph and its #149 link.
  Keep the two existing CI steps: one covers the host tools, the other the
  shipped toolchain image. State that purpose in `.github/workflows/go.yml`.

## Required constraints

- Both options run the same check list, defined once.
- E2e uses only the branch's test project, never the operator's deployment.
- Teardown runs on failure and preserves the first failure; a teardown failure
  must fail the command if the checks succeeded.

## Non-goals

Web/control-panel gates, changes to scenario content, container-free e2e,
and proxy or Compose topology changes.

## Runtime reachability

`make check`, `make test-cycle`, and `.github/workflows/go.yml`.

## Write set

- `Makefile`
- `deployments/Dockerfile.toolchain`
- `deployments/README.md`
- `docs/testing.md`
- `.github/workflows/go.yml`

## Contract surfaces

The meaning of the Go gate in `docs/testing.md` and the two public make targets.
No product contract.

## Demonstration

    $ make check
    $ make test-cycle
    → expect: each runs the shared check sequence successfully and leaves
      no branch test environment running

## Verification

- Inspect the Makefile to confirm both options use the same check list;
  run the demonstration.
- Temporarily add a gofmt-formatted unused function under `test/e2e` to provoke
  lint failure, then replace it with a formatted failing Go test for e2e. For each case,
  run both options and confirm failure at the expected step and cleanup.
  The toolchain mounts the same source, so no Docker override forwarding is
  needed. Remove temporary files afterwards.
- Build the toolchain image from scratch and check its pinned linter version.
  Confirm the focused e2e commands still work against a started test environment.
- Search current gate descriptions in `AGENTS.md`, `docs/` and `skills/`
  for contradictions; preserve historical observations in `docs/state.md`.

## Acceptance criteria

- [ ] Both options run the same checks from one definition, including Docker
      lint and host e2e.
- [ ] Lint and e2e failures fail both options and still clean up the test environment.
- [ ] Testing/deployment documentation describes equivalent options, and CI
      identifies the toolchain each step covers.
- [ ] Focused e2e and environment commands retain their existing behavior.

## Deferrals

None.

## Traces to

Issue #149 and review threads on #148; `docs/testing.md`;
`deployments/README.md`; `skills/shared/workplan-model.md` (Phase 0 meta tasks).
