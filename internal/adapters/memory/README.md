# In-memory application adapters

This package contains concurrency-safe adapters used by the shipped `memory`
store profile as well as by application and contract tests.

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

The daemon exposes the run repository's next-save failure through
`WINCH_MEMORY_INJECT_RUN_SAVE=not_found` (or `conflict`). The plan is consumed
once, so a person can make one request fail and retry it without restarting the
daemon. Other adapters retain their programmatic `FailurePlan` controls for
scenario tests.

This profile proves application behavior and error handling in one daemon
process. It does **not** prove cross-restart durability, multi-instance
consistency, or PostgreSQL/SQL semantics. Restarting the daemon discards every
record.
