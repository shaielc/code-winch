# Code Winch

Code Winch runs coding-agent harnesses in sandboxes you can watch and type at
from a browser. The smallest useful form is one container holding one agent,
serving its own page; from there it grows a control plane over many runs,
multiple agent vendors, rich output rendering, real isolation, and durable
multi-agent workflows.

The repository builds the first runnable slice of that design: a standalone
sandbox attach surface. See **Project status** below.

## Design documents

- [System architecture](docs/architecture.md)
- [Repository and package structure](docs/code-structure.md)
- [Harness and event contracts](docs/contracts.md)
- [Sandbox and security model](docs/security.md)
- [Delivery roadmap](docs/roadmap.md)
- [System state](docs/state.md)
- [Architecture decisions](docs/decisions/README.md)
- [Workplan control panel](workplan-control-panel/README.md)

## Project status

The sandbox image starts, serves a browser attach page on loopback, and reports
its effective profile and unenforced controls. It carries `cmd/fake-harness`,
but does not start it yet. There is no daemon or control plane.
[`docs/state.md`](docs/state.md) is the authoritative account, with a command or
a `file:line` behind every claim.

[`docs/roadmap.md`](docs/roadmap.md) is the order that gets fixed in. Its first
stage is one container holding a runner and the fake harness, started by hand,
serving a page you can type at — deliberately smaller than anything the design
documents describe, because the previous attempt built downward from the control
plane and never produced a configuration anybody could start.

[`docs/workplan/`](docs/workplan/README.md) decomposes that first stage into seven
tasks. `./scripts/list-available-tasks.sh` says which are available; `P0-001`,
which makes the sandbox start and repairs the gates below, is the only one until
it lands.

## Building and testing

Start the product on loopback, then inspect it with the maintained CLI:

```sh
docker compose -f deployments/compose.yml up --build -d
docker compose -f deployments/compose.yml exec sandbox winch status
docker compose -f deployments/compose.yml down
```

The repository gates are:

```sh
python3 -m unittest discover -s test                        # completion scripts
cd workplan-control-panel && python3 -m unittest discover -s tests
make check                                                  # host Go gate
cd web && npm run format:check && npm run lint && npm run typecheck && npm test && npm run build
make test-cycle                                             # Docker-only path: test env, gates, e2e, teardown
```

Host targets need Go 1.24+, npm, and golangci-lint 2.1.6 (`go install
github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6`); `[docker]`
targets need only Docker and run in the `toolchain` container. `make check` is
the host gate CI runs for Go; `make test-cycle` is the Docker path, which also runs the
composed-image scenarios in an isolated `code-winch-test` project, and does not run
golangci-lint. To run only the scenarios, use `make test-env`, then `make e2e` (host,
with `WINCH_E2E_URL` set to the URL it prints) or `make docker-e2e` (in the `toolchain`
container), then `make test-env-down` (see `deployments/README.md`).
