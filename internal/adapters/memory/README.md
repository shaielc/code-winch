# In-memory application adapters

This package contains concurrency-safe, deterministic adapters for application
and contract tests. They are test infrastructure, not production persistence.

- `Calls`/`*Calls` methods return defensive snapshots of attempted successful
  boundary calls. `Published` and `Delivered` expose idempotently deduplicated
  external effects.
- `RunRepository` rejects stale versions and `EventStore` rejects stale expected
  sequences with `application.ErrConflict`.
- Each adapter exposes a `Failures` plan. Inject errors under the lower-case
  method name (`save`, `append`, `read`, `publish`, `resolve`, or `send`). Plans
  are consumed FIFO; inject `nil` to represent a planned successful call.
- Secret values and payloads are defensively copied. Errors include bounded
  resource identifiers and sequence metadata, never payload or secret content.
- `Clock` changes only through `Set`; `IDSource` consumes preloaded typed IDs in
  order and panics with a content-free message when a queue is exhausted.
