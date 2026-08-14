# P5-044: Implement capability and capacity scheduler

**Phase:** 5 — Remote runners and hardening
**Shape:** capability
**Dependencies:** P5-041 (compile: the runner registry supplying capabilities and load), P2-059 (compile: the admission decision point this extends from one host to many)

## Objective

A run lands on a runner that can actually execute it, and a deployment with no
capable runner says so instead of failing at launch.

## Scope

- Placement by declared capability: required sandbox driver, harness adapter,
  isolation class, and any capability the profile demands.
- Placement by capacity: per-runner concurrency and resource headroom, extending
  P2-059's admission from one host to the fleet.
- No capable runner produces a queued run with a stated reason, not a failure;
  a run whose requirements no registered runner can ever satisfy is refused at
  creation.
- Draining: a runner marked draining accepts no new work and finishes what it
  has.
- Placement decisions are recorded with the run and visible, so an operator can
  answer "why did it go there".
- Fleet metrics: capacity, headroom, placement failures, and queue depth per
  capability class.

## Non-goals

- Preemption, bin-packing optimality, or cost-aware placement.
- Autoscaling runner hosts.

## Runtime reachability

Any run on a multi-runner deployment; `winch runners ls --capacity` and
`winch run get <id> --json | jq .placement`.

## Owned surfaces

`internal/application/scheduler.go`,
`api/openapi/components/run.yaml` (placement fields),
`deployments/compose.fleet.yml`.

## Demonstration

    $ docker compose -f deployments/compose.yml -f deployments/compose.fleet.yml up -d --scale runner=3
    $ for i in $(seq 6); do winch run create --sandbox docker --harness fake --json | jq -r .id; done | xargs -n1 winch run start
    $ winch runners ls --capacity
    → expect: work spread across capable runners, none over its declared limit

    $ winch run get $ID --json | jq -r '.placement.runnerId, .placement.reason'
    → expect: the runner and why it was chosen

    # with every docker-capable runner drained:
    $ winch run create --sandbox docker --harness fake && winch run ls
    → expect: queued with a stated reason, not failed

    # requesting a capability no runner declares:
    → expect: refused at creation with a stable code

    $ winch runners drain <id> && winch run ls --runner <id>
    → expect: existing runs finish; no new run is placed there

## Verification

- Standing scenario suite passes on a multi-runner deployment.
- Placement table test across capability and capacity combinations.
- Drain test asserting in-flight work completes and new work does not arrive.
- Metric tests for bounded labels on capability classes.

## Acceptance criteria

- [ ] A run is never placed on a runner lacking a required capability.
- [ ] Per-runner limits are never exceeded.
- [ ] No capable runner means queued with a reason; impossible means refused at
      creation.
- [ ] Placement decisions are recorded and explainable.
- [ ] Draining is observable and complete.

## Deferrals

| Deferred | Owning task |
|---|---|
| Alerting on capacity exhaustion | P5-047 |

## Traces to

`docs/architecture.md` §5 Phase C, §7; `docs/contracts.md` §4;
`docs/roadmap.md` Phase 5
