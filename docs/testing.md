# Testing and gates

How this repository is tested: the gates, which are required and when, how to
run them, how CI covers them, and how a demonstration is run. `AGENTS.md`, the
skills, and the briefs point here and do not restate it. A brief's
**Verification** lists only the checks specific to its task.

## Gates

| Gate | Covers | Options | Required when |
|---|---|---|---|
| Go | Format, vet, lint, unit and contract tests, build; the Docker option also runs the composed-image e2e scenarios | Host or Docker | Always |
| Web | Format, lint, typecheck, unit tests, build of the attach page | Host or Docker | Always |
| Control panel | The workplan control panel's tests | Host | When the change touches `workplan-control-panel/` |
| Completion scripts | The task-status and completion scripts' tests | Host | CI only |

The two Go options overlap on format, vet, tests, and build, and currently
differ: lint runs only in the host option and the e2e scenarios only in the
Docker option ([#149](https://github.com/shaielc/code-winch/issues/149)).

## Running the gates

Run the option your environment supports. Running both is not needed. Say which
you ran.

| Gate | Host | Docker |
|---|---|---|
| Go | `make check` (Go 1.24+, golangci-lint 2.1.6) | `make test-cycle` |
| Web | `cd web && npm run format:check && npm run lint && npm run typecheck && npm test && npm run build` | `make web-test` |
| Control panel | `cd workplan-control-panel && python3 -m unittest discover -s tests` (Python 3.11+, Git) | |
| Completion scripts | `python3 -m unittest discover -s test` | |

`make test-cycle` starts the isolated test environment, runs the gates and the
composed-image e2e scenarios in the `toolchain`, and tears it down.

To run only the scenarios: `make test-env`, then `make docker-e2e`, then
`make test-env-down` (see `deployments/README.md`). The test environment is
named after the current branch.

## The standing scenario suite

`test/e2e` is the end-to-end suite, run against the composed image by
`make test-cycle`. A change keeps it green. A task that makes a substrate real
runs the same scenario against that substrate as well as the fake profile.

## CI

`.github/workflows/` runs the gates:

- `go.yml`: the completion-script tests, the Go gate in both options
  (`make check` and `make test-cycle`), on pushes and pull requests.
- `web.yml`: the web gate, on pull requests and pushes to `main`.
- `control-panel.yml`: the control-panel tests, when files under
  `workplan-control-panel/` change.

CI does not run a brief's demonstration.

## Demonstrations

A demonstration is a manual check, run as the brief writes it. One that starts
the system runs in an isolated environment (see `deployments/README.md`), never
in the operator's deployment.
