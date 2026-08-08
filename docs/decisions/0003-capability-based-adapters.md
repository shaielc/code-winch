# ADR-0003: Capability-based harness and sandbox ports

- **Status:** Proposed
- **Date:** 2026-08-07

## Context

Agent CLIs vary in structured input, terminal behavior, login, approvals, and
resume support. Sandbox backends vary in filesystem, network, limits, snapshots,
and process controls. A common interface can accidentally promise behavior an
adapter cannot provide.

## Decision

Use separate harness and sandbox ports. Every adapter publishes a versioned
descriptor and explicit capabilities. Application policy and UI operate on the
resolved combination and reject unsupported requests rather than silently
degrading security or behavior.

## Consequences

Local, Docker, and future backends compose with multiple harnesses. Capability
combinations require validation and UI states, but unsupported security controls
remain visible. Shared contract suites verify every advertised capability.

## Alternatives

- One adapter per harness/sandbox pair: rejected due to combinatorial growth.
- Boolean `isDocker` branches: rejected because future backends need richer
  semantics.
- Lowest common denominator: rejected because it prevents rich providers and
  stronger sandboxes from exposing useful features.

## Revisit when

A backend cannot be expressed without extensive pair-specific coupling or
capability negotiation becomes ambiguous in practice.
