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

Workflow definitions, pinned instances, step-attempt history, lineage, signals,
timers, and workflow publish intents are installed by migration 005. Coordinator
workers claim ready attempts with `ClaimReadyWorkflowSteps`, using a unique UUID
token for every claim call and a bounded expiry. An expired attempt is reclaimed
with a higher fencing epoch; completion checks the token, epoch, and expiry in
the same transaction that writes workflow outbox intents. A stale worker receives
`application.ErrConflict` with workflow and step identifiers only. Operators
should investigate repeated expiry/reclaim cycles as worker-health failures;
lease tokens, definition/input documents, signal payloads, and outputs must not
be logged. Terminal attempt rows are database-protected and retries append a new
attempt number rather than changing history.
