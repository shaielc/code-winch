# P5-046: Add high-availability and recovery validation

**Phase:** 5 — Remote runners and hardening
**Shape:** hardening
**Dependencies:** P5-043 (semantic: ownership must be fenced before multiple API and worker replicas are safe), P5-045 (semantic: recovery must account for artifacts held remotely), P4-036 (semantic: coordinator restart is part of what recovery must prove)

## Objective

Losing an API replica, a worker, a runner, or the database does not lose a run's
history — and a restore is exercised rather than assumed.

## Scope

- Multiple stateless API replicas behind one entry point, with sticky-free
  WebSocket handling so a browser reconnects to any replica.
- Multiple supervisor and workflow workers claiming leases, with a documented
  and tested failover time.
- Backup and restore: scheduled backups, a documented restore procedure, and a
  drill that restores into a clean environment and replays the standing
  scenarios against it.
- Backup expiry that does not exceed any sensitivity class maximum, verified
  rather than configured.
- Failure-injection drills covering replica loss, worker loss, runner loss,
  database failover, and a partial-storage outage.
- An incident-response record naming on-call contacts and exercising credential
  revocation, runner revocation, audit preservation, containment, and user
  notification.

## Non-goals

- Multi-region topology.
- Zero-downtime schema migration as a general framework; this task documents the
  procedure that the current migrations require.

## Runtime reachability

`deployments/compose.ha.yml`; `make ha-drill` runs the injection suite against
it.

## Owned surfaces

`deployments/compose.ha.yml`, `Makefile` (`ha-drill`, `restore-drill`),
`test/resilience/`, `docs/operations/recovery.md`.

## Demonstration

    $ docker compose -f deployments/compose.ha.yml up -d --scale winchd=3
    $ make ha-drill
    → expect: every injected failure recovers within its documented budget

    # with a live run streaming to a browser:
    $ docker compose kill winchd-2
    → expect: the browser reconnects to another replica and resumes with no gap
      and no duplicate

    $ make restore-drill
    → expect: a backup restored into a clean environment, then the standing
      scenarios pass against it

    $ winch backup verify --expiry
    → expect: no backup retained longer than its highest included sensitivity
      class allows

    # incident drill:
    $ winch credential revoke <id> && winch runners revoke <id>
    → expect: both effective immediately; audit records preserved and readable

## Verification

- Standing scenario suite passes against the HA deployment and against the
  restored environment.
- Failover timing tests with a documented budget per component.
- Backup expiry verification test.
- An audit-preservation test asserting records survive every drill.

## Acceptance criteria

- [ ] No drill loses or duplicates a persisted event.
- [ ] Failover completes within its documented budget for each component.
- [ ] A restore is demonstrated, not described.
- [ ] Backup expiry respects every sensitivity class maximum.
- [ ] Credential and runner revocation are exercised end to end.

## Deferrals

| Deferred | Owning task |
|---|---|
| Alert thresholds over these failure modes | P5-047 |

## Traces to

`docs/architecture.md` §5 Phase C; `docs/security.md` §5, T14, LB09, LB10;
`docs/roadmap.md` Phase 5 exit
