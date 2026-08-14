# P2-059: Bound queued and concurrent runs

**Phase:** 2 — Structured experience and second harness
**Shape:** hardening
**Dependencies:** P1-050 (compile: admission is a decision inside the run use cases and the registry that resolves profiles)

## Objective

A user cannot exhaust the host by starting runs in a loop: the `Queued` state
that has existed since Phase 0 finally means something, and queue depth is
visible.

## Scope

- Admission control on start: configured limits on concurrent running executions
  per deployment and per workspace, with a queued run admitted when a slot frees.
- Queue ordering is deterministic and starvation-free, and a queued run reports
  its position.
- Refusal rather than silent queueing past a configured queue depth, with a
  stable, retryable problem code.
- Cancelling a queued run is immediate and never starts a process.
- Emit the queue-time, active-run, and rejected-admission metrics that
  `docs/architecture.md` §7 lists and P1-048 declared.
- The run page and `winch run ls` show queued position and the reason a run is
  waiting.

## Non-goals

- Scheduling across multiple runners by capability and capacity — P5-044 extends
  this decision point rather than replacing it.
- Per-user quotas or billing.
- Preempting a running execution.

## Runtime reachability

`POST /api/v1/runs/{id}/start` under load on the compose stack; queue state in
`winch run ls`.

## Owned surfaces

`internal/application/admission.go` (implementing the start-path hook
P1-050 defines), `api/openapi/components/run.yaml` (queue fields —
shared with P2-055),
`web/src/features/runs/RunList.tsx`.

## Demonstration

    $ WINCH_MAX_CONCURRENT_RUNS=2 docker compose -f deployments/compose.yml up -d --build
    $ for i in $(seq 5); do winch run create --harness fake --scenario long-running --json | jq -r .id; done | \
        xargs -n1 winch run start
    $ winch run ls
    → expect: exactly 2 running, 3 queued with positions 1..3

    $ winch run stop <a-running-id> && sleep 1 && winch run ls
    → expect: the position-1 run is now running, and positions renumber

    $ for i in $(seq 100); do winch run start <id>; done
    → expect: past the configured queue depth, a stable retryable problem code
      rather than unbounded growth

    $ winch run stop <a-queued-id> && pgrep -f fake-harness | wc -l
    → expect: unchanged; a cancelled queued run never started a process

## Verification

- Standing scenario suite passes with limits configured.
- Concurrency test asserting the limit is never exceeded under parallel starts.
- Starvation test: a run queued first is not overtaken indefinitely.
- Metric tests asserting bounded labels and correct queue-time measurement.

## Acceptance criteria

- [ ] The concurrent-execution limit is never exceeded, including under
      simultaneous starts on one run and across runs.
- [ ] A queued run reports a position and the reason it is waiting.
- [ ] Cancelling a queued run starts no process.
- [ ] Queue time, active runs, and rejected admissions are emitted as metrics.

## Deferrals

| Deferred | Owning task |
|---|---|
| Capability and capacity scheduling across runners | P5-044 |
| Alerting on queue depth | P5-047 |

## Traces to

`docs/architecture.md` §7; `docs/contracts.md` §1 (`Queued`);
`docs/security.md` §1 (denial of service by an authorized user);
`docs/roadmap.md` Phase 2
