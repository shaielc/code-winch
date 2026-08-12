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
Command persistence remains adapter-local until its dedicated delivery task
introduces an application port.
