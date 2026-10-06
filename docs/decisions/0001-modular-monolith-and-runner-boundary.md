# ADR-0001: Modular monolith with a serializable runner boundary

- **Status:** Superseded by [ADR-0005](0005-sandbox-resident-runner.md)
- **Date:** 2026-08-07

The modular-monolith decision below was never implemented. ADR-0005 keeps its
goal — a runner boundary that can move to another host — and reverses its
sequencing: the runner starts in its own process inside the sandbox, so there is
no in-process transport stage, and the control plane is what arrives later.

## Context

Process management benefits from local coordination, while sandbox isolation
and future multi-host capacity imply that runners may eventually be remote.
Starting with independent services would add deployment and distributed-state
cost before the contracts are understood.

## Decision

Start with a modular Go daemon containing API, application, supervisor,
workflow, and runner modules. Define runner commands/events as bounded,
versioned, serializable messages and prohibit application code from using
process/container handles directly.

## Consequences

The first deployment is simple and can use direct in-process transport. The
runner can later become a separate binary without changing use cases. We accept
some up-front protocol discipline and must test serialization even while calls
are local.

## Alternatives

- Microservices now: rejected due to premature operational complexity.
- One tightly coupled process manager: rejected because it obstructs remote
  runner isolation.
- Runner-only desktop application: rejected because workflows and shared web
  access require durable control-plane state.

## Revisit when

Untrusted workloads require runner host separation, one host cannot provide the
needed capacity, or runner release cadence must differ from the control plane.
