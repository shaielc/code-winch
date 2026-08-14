# P4-036: Implement workflow coordinator replay loop

**Phase:** 4 — Top-level workflows
**Shape:** seam
**Dependencies:** P4-035 (compile: the repository, step leases, timers, and signals the loop claims and advances)

## Objective

A workflow instance advances through its graph, survives a coordinator restart
mid-step, and resumes without duplicating an effect.

## Scope

- A worker loop claiming ready steps under lease, evaluating readiness from
  persisted state only, and releasing or expiring leases without losing work.
- Deterministic idempotency keys derived from workflow instance, step, and
  attempt, so replay after a crash cannot issue an external effect twice.
- Retry with the definition's policy, timeout handling, and durable timers.
- Signal delivery: a waiting step wakes when its signal arrives, and a signal
  arriving before the wait is not lost.
- A no-op activity registry plus one trivial built-in step, so the loop is
  runnable and observable before real activities exist.
- Cancellation: cancelling an instance stops scheduling and marks every active
  branch.

## Non-goals

- Real activities. `run.*`, `event.wait`, `approval.wait`, `condition`,
  `parallel`, `foreach`, and `artifact.publish` are P4-037 and P4-038.
- An HTTP surface — P4-039. This task is driven by the CLI.
- Replacing the database coordinator with a dedicated engine; that stays a
  deferred decision behind the `WorkflowRuntime` port.

## Runtime reachability

The workflow worker inside `winchd`; `winch workflow start|get` driving a
definition composed only of built-in no-op steps.

## Owned surfaces

`internal/workflow/coordinator.go`, `internal/workflow/registry.go`,
`cmd/winchd/main.go` (worker registration), `cmd/winch/workflow.go`,
`test/scenarios/workflow-*.json`.

## Demonstration

    $ winch workflow start --definition test/scenarios/workflow-noop.json --json | jq -r .id
    $ winch workflow get $WID --json | jq -r '.steps[] | "\(.id) \(.state)"'
    → expect: every step reaching a terminal state in graph order

    $ winch workflow start --definition test/scenarios/workflow-slow.json &
    $ docker compose -f deployments/compose.yml kill winchd && docker compose up -d winchd
    $ winch workflow get $WID --json | jq '[.steps[] | select(.attempts > 1)] | length'
    → expect: the interrupted step resumes; no step records a duplicated effect

    $ winch workflow cancel $WID
    → expect: every active branch marked cancelled, and no further scheduling

## Verification

- Standing scenario suite passes with the worker running.
- Multi-worker contention test: two coordinators, one instance, no double claim.
- Crash-point tests at each durable boundary, asserting exactly-once effects.
- Fake-clock tests for timers, retry backoff, and lease expiry.

## Acceptance criteria

- [ ] Coordinator restart mid-step does not duplicate an effect.
- [ ] A signal arriving before its wait step is not lost.
- [ ] Two workers cannot own one step attempt.
- [ ] Cancellation reaches every active branch.
- [ ] Readiness is computed from persisted state alone.

## Deferrals

| Deferred | Owning task |
|---|---|
| Run-issuing activities | P4-037 |
| Control-flow, approval, and artifact activities | P4-038 |
| HTTP and UI surfaces | P4-039, P4-040 |

## Traces to

`docs/architecture.md` §4 (workflow coordinator); `docs/contracts.md` §7;
`docs/roadmap.md` Phase 4 exit
