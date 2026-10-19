# Testing and gates

How this repository is tested: the gates, what each needs, what to run in your
environment, how CI covers the rest, and where a demonstration runs.
`AGENTS.md`, the skills, and the briefs point here and do not restate it. A
brief's **Verification** lists only the checks specific to its task.

## Gates

| Gate | Command | Needs | Covers |
|---|---|---|---|
| Completion scripts | `python3 -m unittest discover -s test` | Python 3 | The task-status and completion scripts |
| Control panel | `cd workplan-control-panel && python3 -m unittest discover -s tests` | Python 3.11+, Git | The workplan control panel |
| Go, host | `make check` | Go 1.24+, golangci-lint 2.1.6 | Format, vet, lint, unit and contract tests, build |
| Go, Docker | `make test-cycle` | Docker | Starts the isolated test environment; format, vet, unit and contract tests, build, and the composed-image e2e scenarios in the `toolchain`; tears it down |
| Web | `cd web && npm run format:check && npm run lint && npm run typecheck && npm test && npm run build` | Node and npm, or `make web-test` with Docker | The attach page |

Install golangci-lint with `go install
github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6`.

`make check` and `make test-cycle` overlap on format, vet, tests, and build.
Lint runs only in `make check`; the composed-image e2e scenarios run only in
`make test-cycle`.

## What to run

Run the gates your environment supports. Running both Go paths is not a
requirement:

- Go and golangci-lint on the host: `make check`.
- Only Docker: `make test-cycle`.
- The Python suites always, and the web gate when the change touches `web/`.

Say which you ran. A gate you did not run is covered if CI ran it at your head
(see CI below); otherwise report it as unrun, never as a pass.

To run only the scenarios, start the environment with `make test-env`, then
`make e2e` (host Go, through the local proxy) or `make docker-e2e` (toolchain),
then `make test-env-down` (see `deployments/README.md`). The test environment is
named after the current branch.

## The standing scenario suite

`test/e2e` is the end-to-end suite, run against the composed image by
`make test-cycle`. A change keeps it green. A task that makes a substrate real
runs the same scenario against that substrate as well as the fake profile.

## CI

`.github/workflows/` runs the gates above:

- `go.yml`: the completion-script tests, `make check`, and `make test-cycle`, on
  pushes and pull requests.
- `web.yml`: the web gate, on pull requests and pushes to `main`.
- `control-panel.yml`: the control-panel tests, when files under
  `workplan-control-panel/` change.

CI does not run a brief's demonstration.

## Demonstrations

A demonstration is a manual check. One that starts the system runs in the
per-branch environment: `make test-env`, then
`http://127.0.0.1:$PROXY_PORT/<project>/sandbox/` (see `deployments/README.md`),
then `make test-env-down`.

The default compose project (`code-winch`, `127.0.0.1:8080`) is also the
operator's deployment (`make main-up`), so
`docker compose -f deployments/compose.yml up` and `down` can rebuild or remove
it. A check that needs the host-published port, such as loopback-only
publication, says so, first confirms that no `code-winch` deployment holds it,
and stops rather than replacing one.
