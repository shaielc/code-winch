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

Workflow definitions and their versioned instances are stored alongside
append-only step attempts, durable signals, timers, lineage, and transactional
workflow outbox intents. `ClaimReadySteps` uses `SKIP LOCKED` so concurrent
workers cannot claim the same attempt. An expired claim may be reclaimed with a
new UUID token; workers must pass the returned token to `CompleteStep`, which
rejects stale owners with `application.ErrConflict`. Operators should size the
lease longer than normal database round trips and monitor retries: lease expiry
permits duplicate execution, while the fencing token prevents a stale result
from changing durable history. Logs should include workflow, step, and attempt
identifiers, but not stored inputs, outputs, signals, or definition content.
