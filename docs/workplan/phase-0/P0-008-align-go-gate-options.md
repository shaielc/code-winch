# P0-008: Align the host and Docker Go gate options

**Phase:** 0 — Plan the implementation phases
**Shape:** capability
**Dependencies:** None

Resolves issue [#149](https://github.com/shaielc/code-winch/issues/149),
including its request for a Docker-free host test environment.

## Objective

`make check` (host) and `make test-cycle` (Docker) run the same Go checks:
format, vet, lint, unit and contract tests, build, and the standing e2e scenarios.
Host start/stop commands make those scenarios and manual demonstrations
usable without Docker.

## Scope

The Makefile's `check` target omits e2e; `test-cycle` omits lint.
The toolchain image lacks golangci-lint, and test-environment startup
currently requires Docker.

- Define `format-check vet lint test build e2e` once in the Makefile and
  have both options use it.
- Install golangci-lint v2.1.6, matching CI, in the toolchain image.
- Add `make host-test-env` and `make host-test-env-down` to start and stop
  a host environment usable by `make e2e` and manual demonstrations.
- Update `docs/testing.md` and `deployments/README.md` to describe the
  equivalent gates and how to use each test environment.

## Required constraints

Both gate options use a single check-list definition.

## Non-goals

Product behavior and web/control-panel gates.

## Runtime reachability

`make check`, `make test-cycle`, `make host-test-env`,
`make host-test-env-down`, and `make e2e`.

## Write set

- `Makefile`
- `deployments/Dockerfile.toolchain`
- `deployments/README.md`
- `docs/testing.md`
- `.github/workflows/go.yml` (if the gate invocation needs adapting)

## Contract surfaces

The Go gate and test-environment commands documented in `docs/testing.md`
and `deployments/README.md`.

## Demonstration

Without Docker:

    $ make host-test-env
    $ make check
    $ make e2e
    → expect: the gate and standing scenarios pass; the sandbox is available
      for a manual demonstration
    $ make host-test-env-down
    → expect: the host test sandbox stops

With Docker, run `make test-cycle`; expect the same check sequence to pass.

## Verification

- Run the demonstration and confirm both options use the shared check list.
- Introduce a temporary lint violation, then a failing e2e test; confirm each
  fails both gate options. Remove the temporary changes.
- Build the toolchain image and confirm the pinned linter is available.
- Check that the testing and deployment documentation matches the commands.

## Acceptance criteria

- [ ] Both options run the same checks from one definition, including Docker
      lint and host e2e; lint/e2e failures fail the gate.
- [ ] Host start/stop commands support e2e and manual demonstrations without Docker.
- [ ] Testing and deployment documentation describes both options.

## Deferrals

None.

## Traces to

Issue #149 and its [host-environment comment](https://github.com/shaielc/code-winch/issues/149#issuecomment-6338329527);
review threads on #148; `docs/testing.md`; `deployments/README.md`;
`skills/shared/workplan-model.md` (Phase 0 meta tasks).
