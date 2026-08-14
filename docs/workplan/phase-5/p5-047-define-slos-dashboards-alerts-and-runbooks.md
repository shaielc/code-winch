# P5-047: Define SLOs, dashboards, alerts, and runbooks

**Phase:** 5 — Remote runners and hardening
**Shape:** capability
**Dependencies:** P1-048 (compile: the metric registry and bounded-label rules these build on), P5-044 (semantic: fleet capacity is one of the things being alerted on), P5-046 (semantic: the failure modes with budgets come from the drills)

## Objective

An operator on call can see whether the system is healthy, is paged when it is
not, and has a written procedure for each alert.

## Scope

- Service level objectives with stated measurement windows for the
  `docs/roadmap.md` acceptance metrics: event-to-browser latency, stop
  completion within the escalation deadline, sequence integrity, and secret
  canary counts.
- Dashboards over the metrics P1-048 declared and later tasks emit: queue time,
  startup time, active runs, dropped subscribers, parser failures, workflow
  retries, cleanup failures, stale-event rejections, and placement failures.
- Alerts tied to SLO burn rather than raw thresholds, each with a runbook naming
  the first diagnostic step and the containment action.
- Runbooks for the drills in P5-046 and for credential and runner revocation.
- A telemetry conformance check in CI asserting no metric label or log field
  outside the bounded sets in `docs/security.md` §5 — the enforcement that
  every brief's acceptance criteria have assumed since Phase 0.
- Secret canary monitoring running continuously, not only in tests.

## Non-goals

- Choosing a vendor. The exporter is a port; a local stack proves it.
- Business or usage analytics.

## Runtime reachability

`deployments/compose.observability.yml` brings up the exporter and dashboards;
`make telemetry-check` runs the conformance check.

## Owned surfaces

`deployments/compose.observability.yml`, `deployments/dashboards/`,
`deployments/alerts/`, `docs/operations/runbooks/`,
`Makefile` (`telemetry-check`).

## Demonstration

    $ docker compose -f deployments/compose.yml -f deployments/compose.observability.yml up -d
    $ make e2e && open http://localhost:3000
    → expect: every dashboard panel populated by the scenario run, none empty

    $ make telemetry-check
    → expect: pass; then add a free-form log field on a branch and expect a
      named failure

    # inject a stalled outbox:
    → expect: the backlog alert fires, and its runbook's first diagnostic step
      identifies the cause

    $ winch canary-scan --continuous &
    → expect: a zero count, and a raised signal within its budget when a canary
      is deliberately introduced

    $ ls docs/operations/runbooks/
    → expect: one runbook per alert, each naming a first step and a containment
      action

## Verification

- Standing scenario suite passes with the observability stack attached.
- A test asserting each alert has a runbook and each runbook an alert.
- Telemetry conformance check with positive and negative fixtures.
- Dashboard panels validated against the metric registry so a renamed metric
  fails the build rather than emptying a panel.

## Acceptance criteria

- [ ] Every acceptance metric in `docs/roadmap.md` has an SLO with a measurement
      window.
- [ ] Every alert has a runbook, and no runbook is orphaned.
- [ ] Content and secrets cannot enter telemetry, enforced by a CI check rather
      than by convention.
- [ ] Canary monitoring runs continuously and raises a signal within its budget.
- [ ] A renamed or removed metric fails the build.

## Deferrals

`None.`

## Traces to

`docs/architecture.md` §7; `docs/security.md` §5, T04, LB09;
`docs/roadmap.md` initial acceptance metrics
