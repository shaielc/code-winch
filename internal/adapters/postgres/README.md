# PostgreSQL storage

`Store` implements durable run, command, and event persistence. Connect with a
least-privilege PostgreSQL role, construct a `pgxpool.Pool`, and call
`MigrateUp` during deployment before serving traffic. Production migrations are
forward-only; `MigrateDown` is intended for clean-database verification.

Event appends atomically compare `expectedSequence`, reserve exactly the batch
range, and insert the batch. A conflict returns `application.ErrConflict` with
only resource identifiers. Secret-sensitivity events and invalid JSON are
rejected before commit. Resolved run configuration passed to
`SaveResolvedConfiguration` must contain non-secret values only; callers retain
responsibility for replacing credentials with opaque references before storage.
Every committed run event also creates a `run.events` outbox row in the same
transaction. Start one or more `application.OutboxWorker` loops with unique
worker IDs and UUID lease tokens. Claims use expiring, fenced leases and
`SKIP LOCKED`; delivery is at least once, so subscribers must deduplicate by
message/event ID. Failed publishes use bounded exponential backoff and become
poison records after the configured attempt limit. Worker metrics expose the
ready backlog, retry count, and poison count without payload content.
