# ADR-0002: Canonical events separated from rendering

- **Status:** Proposed
- **Date:** 2026-08-07

## Context

Harnesses emit terminals, structured JSON, messages, tool calls, and provider-
specific records. Users may want terminal, conversational, activity, or diff
views. Persisting presentation output would couple history to one UI/version.

## Decision

Adapters translate output into immutable canonical events with optional
namespaced extensions. Renderers are disposable, versioned projections over
authorized events. Raw stream events may be retained by policy for fidelity,
but are not the only semantic representation.

## Consequences

New views can render old runs and workflows can react to semantic events.
Adapters bear normalization complexity, schemas require compatibility policy,
and a malformed parser must degrade to diagnostic/raw events rather than losing
output.

## Alternatives

- Store HTML/view models: rejected as unsafe and version-coupled.
- Store raw terminal bytes only: rejected because tools and workflows need
  semantics.
- Expose every provider schema directly: rejected because generic UI and
  workflow steps would become provider-specific.

## Revisit when

Canonical schemas consistently lose required provider semantics or projection
cost makes on-demand/cached rendering inadequate.
