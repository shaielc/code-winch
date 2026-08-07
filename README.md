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
- [Architecture decisions](docs/decisions/README.md)

## Project status

This is an architecture-only foundation. No runtime is implemented yet. The
documents distinguish initial interfaces from later deployment choices so that
the first vertical slice does not prematurely require Docker, a workflow
engine, or a particular coding-agent provider.
