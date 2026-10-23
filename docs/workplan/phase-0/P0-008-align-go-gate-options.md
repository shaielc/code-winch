# P0-008: Align the host and Docker Go gate options

**Phase:** 0 — Plan the implementation phases
**Shape:** capability
**Dependencies:** None

Resolves issue [#149](https://github.com/shaielc/code-winch/issues/149),
including its request for a Docker-free host test environment.
This Phase 0 meta task changes repository tooling.

## Objective

`make check` (host) and `make test-cycle` (Docker) run the same Go checks:
format, vet, lint, unit and contract tests, build, and the standing e2e scenarios.
Explicit host start/stop commands also make those scenarios and manual
demonstrations usable without Docker.

## Scope

The Makefile's `check` target omits e2e; `test-cycle` omits lint.
The toolchain image lacks golangci-lint, and test-environment startup currently
requires Docker.

- Define `format-check vet lint test build e2e` once in the Makefile and
  have both options use it. Reuse the existing recipes and package lists.
- Install golangci-lint v2.1.6, matching CI, in the toolchain image so the
  Docker option can run the existing lint target.
- Add `make host-test-env` and `make host-test-env-down` to start and stop
  the local sandbox without Docker, using the existing Go and Node/npm build
  prerequisites. Serve the built web assets on a loopback URL and print that
  URL for use as `WINCH_E2E_URL` by `make e2e` and manual demonstrations.
- Add the existing e2e scenarios to `make check`, targeting the explicitly
  started environment through `WINCH_E2E_URL`. Starting and stopping the host
  environment are separate user commands; `make check` does neither.
- Preserve `make e2e`, `make docker-e2e`, `make test-env` and
  `make test-env-down` with their existing responsibilities.
- Update `docs/testing.md` and `deployments/README.md` with the equivalent
  check lists, host start/stop commands and each option's environment needs.
  Remove the current-difference paragraph and its #149 link. Keep both CI paths:
  explicitly start/stop the host environment around `make check`; retain
  `make test-cycle` for the shipped toolchain image.

## Required constraints

- Both options run the same check list, defined once.
- Tests and demonstrations use their test environment, never the operator's
  deployment.
- Host environment lifecycle is explicit and independent of `make check`.

## Non-goals

Web/control-panel gates, changes to scenario content, proxy or Compose topology
changes, and new automatic cleanup or failure-recovery behavior.

## Runtime reachability

`make check`, `make test-cycle`, `make host-test-env`,
`make host-test-env-down`, `make e2e`, and `.github/workflows/go.yml`.

## Write set

- `Makefile`
- `deployments/Dockerfile.toolchain`
- `deployments/README.md`
- `docs/testing.md`
- `.github/workflows/go.yml`

## Contract surfaces

The meaning of the Go gate in `docs/testing.md` and the make commands above.
No product contract.

## Demonstration

On a host with the existing toolchain prerequisites and no Docker daemon:

    $ make host-test-env
    → expect: sandbox starts and reports a loopback URL
    $ make check WINCH_E2E_URL=<reported-url>
    $ make e2e WINCH_E2E_URL=<reported-url>
    → expect: the gate and standing scenarios pass; the URL remains available
      for a manual demonstration
    $ make host-test-env-down
    → expect: the host test sandbox stops

On a Docker-capable host, run `make test-cycle`; expect the same check sequence
to pass against the composed image.

## Verification

- Inspect the Makefile to confirm both options use the same check list;
  run the demonstration.
- Temporarily add a gofmt-formatted unused function under `test/e2e` to provoke
  lint failure, then replace it with a formatted failing Go test for e2e.
  Run both options and confirm failure at the expected step. Remove temporary
  files afterwards and stop the host environment explicitly.
- Build the toolchain image from scratch and check its pinned linter version.
  Confirm `make docker-e2e` still works against a started Docker test environment.
- Check current gate descriptions in `AGENTS.md`, `docs/` and `skills/`
  for contradictions; preserve historical observations in `docs/state.md`.

## Acceptance criteria

- [ ] Both options run the same checks from one definition, including Docker
      lint and host e2e; lint/e2e failures fail the corresponding gate.
- [ ] Explicit host start/stop commands support the standing scenarios and
      manual demonstrations without Docker; `make check` does not manage them.
- [ ] Testing/deployment documentation describes both environments, and CI
      exercises the host and Docker paths.
- [ ] Existing focused e2e and Docker environment commands retain their behavior.

## Deferrals

None.

## Traces to

Issue #149 and its [host-environment comment](https://github.com/shaielc/code-winch/issues/149#issuecomment-6338329527);
review threads on #148; `docs/testing.md`; `deployments/README.md`;
`skills/shared/workplan-model.md` (Phase 0 meta tasks).
