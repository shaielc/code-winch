# PostgreSQL storage

`Store` implements durable run, command, and event persistence. Connect with a
least-privilege PostgreSQL role, construct a `pgxpool.Pool`, and call
`MigrateUp` on every start before serving traffic. A `schema_migrations` ledger
records what a database already has, so each migration runs once and a start
against an already-migrated database is a no-op; an advisory lock serializes
concurrent migrators. Production migrations are forward-only; `MigrateDown`
reverses what the ledger records and is intended for clean-database
verification.

A run's own fields — workspace path, requested harness and sandbox profiles,
and its timestamps — are columns on `runs`, and the caller's clock owns the
timestamps the store writes. Creation is idempotent per request key: migration
006 adds `create_actor` and `create_idempotency_key` with a unique constraint,
following `input_commands` and `workflow_signals`, so replaying a key returns
the stored run and reusing it for a different run reports
`application.ErrIdempotencyConflict`. A run no client request created leaves
both NULL and dedupes against nothing. `resolved_configuration` holds only the
resolved configuration described in `docs/code-structure.md`; its writer
replaces it as one document, so run fields must never be kept there.

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
