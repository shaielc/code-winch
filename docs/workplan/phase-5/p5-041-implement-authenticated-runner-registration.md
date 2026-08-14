# P5-041: Implement authenticated runner registration

**Phase:** 5 — Remote runners and hardening
**Shape:** seam
**Dependencies:** P1-049 (compile: `internal/runner/local` is what the standalone binary hosts)

## Objective

A separate `winch-runner` process registers with the control plane over an
authenticated channel, publishes its capabilities, and heartbeats — appearing in
`winch runners ls` as a real, revocable identity.

## Scope

- `cmd/winch-runner`: the standalone runner composition root
  `docs/code-structure.md` §1 reserves and no task has ever named. It hosts the
  same `internal/runner/local` the daemon embeds.
- Machine identity separate from browser identity: mutually authenticated
  transport, credential rotation, and immediate revocation.
- Registration reporting runner ID, supported protocol range, harness and
  sandbox capabilities, load, and lease epoch.
- Protocol negotiation selecting the highest mutually supported major/minor
  version; a major mismatch refuses assignment rather than degrading.
- Heartbeats with a miss budget, after which the runner is marked unavailable
  and its runs become reconcilable.
- A registry of runners readable through the API, with capability and health.

## Non-goals

- Dispatching commands to a remote runner — P5-042. This task registers and
  heartbeats; the embedded runner still executes.
- Scheduling by capability — P5-044.
- Sandbox profile enforcement, which is independent of where a runner lives.

## Runtime reachability

`winch-runner --endpoint … --identity …` against the compose stack;
`winch runners ls` and `winch runners revoke`.

## Owned surfaces

`cmd/winch-runner/`, `internal/application/runner_registry.go`,
`internal/adapters/transport/runnerrpc/`,
`api/openapi/paths/runners.yaml`,
`internal/adapters/postgres/migrations/014_*.sql`.

## Demonstration

    $ winch-runner --endpoint https://localhost:8080 --identity /etc/winch/runner.pem &
    $ winch runners ls
    → expect: one runner, its capabilities, and a recent heartbeat

    $ winch-runner --identity /etc/winch/wrong.pem
    → expect: registration refused; nothing appears in the registry

    $ winch runners revoke <id>
    → expect: the runner's next heartbeat is rejected and it stops being
      assignable, without restarting the control plane

    $ kill -STOP %1 && sleep <miss-budget>; winch runners ls
    → expect: marked unavailable, with the miss count and no content

    # a runner advertising an unsupported major version:
    → expect: assignment refused with a stable code naming the version range

## Verification

- Standing scenario suite passes with a registered runner present.
- Negotiation table test across supported and unsupported version pairs.
- Identity rotation and revocation tests, including a revoked runner retrying.
- Heartbeat miss and recovery tests with a fake clock.

## Acceptance criteria

- [ ] A runner cannot register without a valid machine identity.
- [ ] Revocation takes effect without a control-plane restart.
- [ ] Version negotiation refuses a major mismatch rather than degrading.
- [ ] Registry entries contain identifiers, capabilities, and health only.
- [ ] `cmd/winch-runner` builds and runs from the same runner code the daemon
      embeds.

## Deferrals

| Deferred | Owning task |
|---|---|
| Sending commands and receiving events remotely | P5-042 |
| Choosing which runner gets a run | P5-044 |

## Traces to

`docs/architecture.md` §5 Phase B; `docs/contracts.md` §4;
`docs/security.md` §9, T05, LB10; `docs/code-structure.md` §1
