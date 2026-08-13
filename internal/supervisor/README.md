# Run supervisor

The supervisor is the application-level serialization boundary for a run. A
caller first acquires a durable lease, then uses that returned lease for desired
state changes and observations. Lease tokens are capabilities: do not include
them in logs, errors, metrics, events, or API responses. Every takeover advances
the durable epoch, so delayed observations from a prior execution are rejected.

`Execute` saves desired state and the resolved harness/sandbox driver IDs before
sending a runner command. A send failure therefore leaves durable intent for
restart reconciliation rather than pretending the command never existed.
`Rehydrate` reads that checkpoint without requiring a resident goroutine.

On daemon startup, call `Reconcile` for each persisted nonterminal run after
its old lease expires. Reconciliation inspects the stored execution ID, uses
the previous ownership token as compare-and-swap evidence, and only then lets
the new lease supervise a running execution. Preparing or unknown executions
are failed and cleaned, observed exits are recorded truthfully, and a pending
stop is reissued. Every outcome emits a content-free `run.reconciled` event;
cleanup failures return `ORPHAN_CLEANUP_FAILED` with run and execution IDs.

All observation batches pass through the configured redactor before the store
atomically verifies the lease and runner ordinal, assigns run-local sequences,
persists the events, and creates publication outbox records. Redaction failures
and `secret` sensitivity are fail-closed. `ErrStaleLease` is actionable: stop
using that execution and reacquire; never retry its observations under a new
lease.
