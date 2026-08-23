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
- [System state](docs/state.md)
- [Architecture decisions](docs/decisions/README.md)
- [Task runner deployment](runner/README.md)

## Project status

[`docs/state.md`](docs/state.md) is the authoritative account of what runs, what
went wrong, and what has no working code behind it. In short: the daemon starts,
migrates, and serves an authenticated HTTP API and the browser app, and
`winch dev run` drives a harness under a local PTY by hand — but the run use
cases that would join them are unbound, so no run can be created through the
product. No implementation plan is in flight.

See [`web/README.md`](web/README.md) for the web workspace's deterministic
development and generated-client checks, and
[`deployments/README.md`](deployments/README.md) for the local stack.

## Go development

Every target in the root `Makefile` is marked `[host]` or `[docker]`. `[host]`
targets run directly on this machine; `[docker]` targets need Docker only and
run inside the `runner` container.

### On the host

Go 1.24 or newer and golangci-lint 2.1.6 are required. Install the linter with
`go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6`.
These are the same quality gates CI runs:

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

### In Docker

With no Go toolchain installed, the containerised toolchain runs the Go gates
and the integration suite instead:

```sh
make test-cycle      # build, start, verify, integration-test, tear down
make runner-verify   # gofmt, vet, unit tests, and compile in the runner
make runner-shell    # a shell in the runner
```

`runner-verify` is `check` minus `lint` and `api-check`, which need golangci-lint
and npm that the image does not carry — so a host run is still required before
submitting. See [deployments/README.md](deployments/README.md) for the step-by-step
form and the database it uses.
