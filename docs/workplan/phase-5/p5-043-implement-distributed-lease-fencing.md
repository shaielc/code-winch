# P5-043: Implement distributed lease fencing

**Phase:** 5 — Remote runners and hardening
**Shape:** hardening
**Dependencies:** P5-042 (semantic: two hosts must be able to reach one run before fencing between them can be demonstrated)

## Objective

A partitioned runner that comes back cannot append to a run someone else now
owns, and no partition produces two controllers of one execution.

## Scope

- Lease epochs monotonic across the deployment, so a takeover always receives a
  greater epoch and an older epoch is provably stale.
- Event rejection at the append boundary for any observation carrying a stale
  token or epoch, retained only in runner diagnostics and counted as a metric.
- Takeover protocol: the new owner verifies execution identity before adopting
  it, and marks the run failed rather than adopting an execution it cannot
  identify — the behavior `internal/supervisor` already encodes locally, now
  correct across hosts.
- Fencing of the side effects that outlive a lease: outbox claims, artifact
  writes, and cleanup.
- A partition-injection test harness so these paths are exercised
  deterministically rather than by chance.

## Non-goals

- Consensus. PostgreSQL remains the authority for ownership.
- Automatic re-parenting of a live execution across hosts.

## Runtime reachability

Any remote run; stale-event rejections surface as a metric and an audit fact.

## Owned surfaces

`internal/supervisor/fencing.go`, `internal/adapters/postgres/repository.go`
(epoch predicates), `test/integration/partition/`.

## Demonstration

    $ ID=$(winch run create --harness fake --scenario long-running --json | jq -r .id)
    $ winch run start $ID
    $ docker network disconnect <net> <runner-a>      # partition the owner
    $ winch runners revoke <runner-a> && winch run reconcile $ID
    → expect: ownership moves, with a strictly greater epoch

    $ docker network connect <net> <runner-a>
    $ winch run watch $ID --after-sequence 0 | grep -c '<runner-a-only-marker>'
    → expect: 0; nothing runner A produced after losing the lease is in history

    $ winch metrics | grep stale_events_rejected
    → expect: a non-zero count attributable to that window

    $ pgrep -f fake-harness   # on runner A
    → expect: nothing; the fenced runner cleans up rather than continuing

## Verification

- Standing scenario suite passes with fencing active.
- Partition matrix: partition before start, during output, during stop, and
  during cleanup.
- A test asserting a stale observation is never persisted, under concurrency.
- Epoch monotonicity test across daemon restarts.

## Acceptance criteria

- [ ] Two owners can never control one execution.
- [ ] Stale events cannot enter canonical history under any tested partition.
- [ ] A fenced runner stops and cleans up rather than continuing.
- [ ] Epochs are monotonic across restarts and takeovers.
- [ ] Rejections are counted and auditable without content.

## Deferrals

`None.`

## Traces to

`docs/architecture.md` §2 (one writer per run), §7;
`docs/contracts.md` §4; `docs/security.md` T05, LB10;
`docs/roadmap.md` Phase 5 exit
