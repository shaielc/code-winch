# Execution

The execution engine is what makes a started run actually run. The run service
records what a run should be doing; this package makes it so, and reports back
what happened.

It owns three flows:

- **Launch** takes the supervisor lease, binds an execution to its harness and
  sandbox pair, and sends `prepare` then `start`. Every runner message is
  preceded by a durable desired state, so a daemon that dies mid-launch leaves
  a checkpoint rather than a mystery. A launch that fails leaves the run
  `failed`, not `queued` forever.
- **Observation** turns runner output into sequenced, durable events under the
  run's lease, and turns a harness exit into the run's terminal state. It also
  emits the `run.lifecycle` events that let a reader of the history see the
  state changes beside the output that caused them.
- **Delivery** hands an accepted input command to the running harness. It never
  changes the run's desired state: input is not a lifecycle change.

Ordinals are allocated here rather than forwarded from the runner, because the
engine interleaves its own lifecycle events with the runner's output and the
store requires each write to lead the last. Allocation and the write that uses
it are serialized per execution: an ordinal that loses its race is rejected by
the store, and the event would be silently dropped.

It is a composition-layer orchestrator, not an adapter, so unlike the adapters
it may import the concrete supervisor and runner packages. This is the wiring
`cmd/winchd` would otherwise hold inline, kept here so it can be tested against
the in-memory ports.

## Limits

- `ClassificationRedactor` reads an event's declared sensitivity and rewrites
  nothing. Scanning payloads for credential values is P3-032.
- Input is offered as `text` and `resize` only; the kinds with no runner
  message yet are refused at acceptance rather than dropped at delivery.
- A run interrupted by a daemon that was killed rather than stopped is left for
  the restart reconciliation sweep, which is not wired yet (P1-020, P1-050).
