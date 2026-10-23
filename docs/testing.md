# Testing and gates

How this repository is tested: the gates, which are required and when, how to
run them, how CI covers them, and how a demonstration is run. `AGENTS.md`, the
skills, and the briefs point here and do not restate it. A brief's
**Verification** lists only the checks specific to its task.

## Gates

| Gate | Covers | Options | Required when |
|---|---|---|---|
| Go | Format, vet, lint, unit and contract tests, build, and the standing e2e scenarios | Host or Docker | Always |
| Web | Format, lint, typecheck, unit tests, build of the attach page | Host or Docker | Always |
| Control panel | The workplan control panel's tests | Host | When the change touches `workplan-control-panel/` |
| Completion scripts | The task-status and completion scripts' tests | Host | CI only |

The two Go options run the same `GO_CHECKS` list from the Makefile: format,
vet, lint, unit and contract tests, build, and e2e. The Docker toolchain pins
golangci-lint 2.1.6, the same version required on the host and in CI.

## Running the gates

Run the option your environment supports. Running both is not needed. Say which
you ran.

| Gate | Host | Docker |
|---|---|---|
| Go | `make host-test-env && make check` (Go 1.24+, golangci-lint 2.1.6, Node 20+) | `make test-cycle` |
| Web | `cd web && npm run format:check && npm run lint && npm run typecheck && npm test && npm run build` | `make web-test` |
| Control panel | `cd workplan-control-panel && python3 -m unittest discover -s tests` (Python 3.11+, Git) | |
| Completion scripts | `python3 -m unittest discover -s test` | |

`make check` expects the host test sandbox to be running. Start it with
`make host-test-env`, then stop it with `make host-test-env-down` after the gate
or a manual demonstration. It builds the attach page and Go binaries, starts a
loopback-only sandbox at `http://127.0.0.1:7081`, and writes its PID and log to a
repository-specific directory under `/tmp`. Set `HOST_TEST_PORT` and
`WINCH_E2E_URL` together to use another address.

`make test-cycle` builds the pinned toolchain image, starts the isolated Docker
test environment, runs that same `GO_CHECKS` list (including lint and e2e) in
the `toolchain`, and always tears the environment down.

To run only the scenarios on the host: `make host-test-env`, `make e2e`, then
`make host-test-env-down`. To run only the scenarios in Docker: `make test-env`,
then `make docker-e2e`, then `make test-env-down` (see
`deployments/README.md`). The Docker test environment is named after the
current branch.

## The standing scenario suite

`test/e2e` is the end-to-end suite. `make check` runs it against the host test
sandbox and `make test-cycle` runs it against the composed image. A change
keeps it green. A task that makes a substrate real runs the same scenario
against that substrate as well as the fake profile.

## CI

`.github/workflows/` runs the gates:

- `go.yml`: the completion-script tests, then the Go gate in both options, on
  pushes and pull requests. The host option starts the host test sandbox
  (`make host-test-env`), runs `make check`, and always stops the sandbox
  (`make host-test-env-down`); the Docker option is `make test-cycle`. The job
  sets up Node because `make host-test-env` builds the attach page.
- `web.yml`: the web gate, on pull requests and pushes to `main`.
- `control-panel.yml`: the control-panel tests, when files under
  `workplan-control-panel/` change.

CI does not run a brief's demonstration.

## Demonstrations

A demonstration is a manual check, run as the brief writes it. One that starts
the system runs in an isolated environment (see `deployments/README.md`), never
in the operator's deployment.
