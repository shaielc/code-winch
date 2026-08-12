# Code Winch

Code Winch is a proposed control plane and web interface for running coding-agent
harnesses. The repository currently contains the design baseline for an
implementation that can start with local processes and evolve toward isolated
containers, multiple agent vendors, rich output rendering, and durable
multi-agent workflows.

## Design documents

- [System architecture](docs/architecture.md)
- [Repository and package structure](docs/code-structure.md)
- [Harness and event contracts](docs/contracts.md)
- [Sandbox and security model](docs/security.md)
- [Delivery roadmap](docs/roadmap.md)
- [Implementation workplan](docs/workplan/README.md)
- [Architecture decisions](docs/decisions/README.md)
- [Task runner deployment](runner/README.md)

## Project status

The Go workspace, an intentionally empty daemon composition root, and a
bootstrapped web workspace are in place; domain, adapter, and product behavior
have not been implemented yet. See [`web/README.md`](web/README.md) for the web
workspace's deterministic development and generated-client checks. The
documents distinguish initial interfaces from later deployment choices so that
the first vertical slice does not prematurely require Docker, a workflow
engine, or a particular coding-agent provider.

## Go development

Go 1.24 or newer and golangci-lint 2.1.6 are required. Install the linter with
`go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6`.
The repository exposes the same quality gates used in CI through the root
`Makefile`:

```sh
make format        # format Go source files
make format-check  # verify formatting without changing files
make vet           # run go vet
make lint          # run the pinned golangci-lint release
make test          # run Go unit tests
make build         # build cmd/winchd
make check         # run all non-mutating CI checks
```

CI installs the pinned linter before invoking these commands. Run `make format`
when `make format-check` reports file names, then rerun `make check` before
submitting a change. Set `GOLANGCI_LINT` to an alternate binary path when your
linter is not on `PATH`.
