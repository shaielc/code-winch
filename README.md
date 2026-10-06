# Code Winch

Code Winch runs coding-agent harnesses in sandboxes you can watch and type at
from a browser. The smallest useful form is one container holding one agent,
serving its own page; from there it grows a control plane over many runs,
multiple agent vendors, rich output rendering, real isolation, and durable
multi-agent workflows.

The repository currently holds the design baseline and the delivery order for
that, and very little of the implementation — see **Project status** below.

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

**Nothing of the product runs yet.** There is no daemon, no sandbox, no browser
application, and no way to start a harness. What exists is the design set,
`cmd/fake-harness` (a controllable stand-in for a vendor CLI), the workplan
control panel, and the planning skills.
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

The root `Makefile` and both CI workflows are **currently broken**. Most targets
name a path that was removed along with the previous implementation —
`api/openapi/`, `internal/`, `cmd/winchd`, `web/src/`, `test/e2e/`,
`test/contract/`, and `deployments/compose.yml`, which the whole `[docker]` group
depends on — so `make check`, `make build`, and `make test-cycle` all fail before
reaching any code. `format`, `format-check`, `vet`, and `lint` are the exception
and would pass on a host with Go installed. `docs/state.md` lists each gate with
the line that names the missing path. Repairing them is task `P0-001`.

What runs today:

```sh
python3 -m unittest discover -s test                        # completion scripts
cd workplan-control-panel && python3 -m unittest discover -s tests
./scripts/list-available-tasks.sh                           # expect []
```

Once the gates are repaired, every `Makefile` target is marked `[host]` or
`[docker]`: `[host]` targets run on this machine and need Go 1.24+, npm, and
golangci-lint 2.1.6 (`go install
github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6`); `[docker]`
targets need only Docker and run in the `runner` container. `make check` is the
host gate CI runs, and `make test-cycle` is the Docker path — it covers neither
`lint` nor `api-check`, so a host run is still required before submitting.
