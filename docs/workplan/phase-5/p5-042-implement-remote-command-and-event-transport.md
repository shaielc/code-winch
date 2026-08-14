# P5-042: Implement remote command and event transport

**Phase:** 5 — Remote runners and hardening
**Shape:** swap
**Dependencies:** P5-041 (compile: the registered runner and its authenticated channel), P1-050 (compile: the `RunnerGateway` binding in the composition root being swapped)

## Objective

The standing scenarios pass unchanged with the harness executing on a different
host from the control plane.

## Scope

- A remote `RunnerGateway` sending `prepare`, `start`, `input`, `resize`,
  `stop`, `inspect`, and `cleanup` over the registered channel, with command IDs
  and lease tokens.
- Runner-to-control-plane observations carrying runner-local ordinal, command
  correlation ID, execution ID, and lease token. The control plane assigns
  canonical sequences; the runner never does.
- Backpressure: bounded payload size, bounded outstanding commands, and bounded
  event buffering, with a defined behavior when a bound is hit — never unbounded
  memory and never a silent drop.
- Reconnection: a dropped channel resumes without losing or duplicating an
  observation, using the same after-ordinal discipline the browser stream uses.
- Additive-field tolerance in both directions, verified against the committed
  protocol fixtures.

## Non-goals

- Ownership fencing across hosts — P5-043 hardens what this makes possible.
- Artifact bytes, which need their own transfer path — P5-045.
- Scheduling — P5-044.

## Runtime reachability

`make e2e PROFILE=remote` with `winch-runner` on a second container;
`winch run ls --runner <id>`.

## Owned surfaces

`internal/adapters/transport/runnerrpc/` (command and event streams),
`internal/runner/remote/`, `deployments/compose.remote.yml`,
`schemas/runner/v1/fixtures/`.

## Demonstration

    $ docker compose -f deployments/compose.yml -f deployments/compose.remote.yml up -d --build
    $ make e2e PROFILE=remote
    → expect: every standing scenario passes, unchanged
    $ make e2e PROFILE=fake && make e2e PROFILE=docker
    → expect: both still pass; local execution is not regressed

    $ ID=$(winch run create --harness fake --json | jq -r .id) && winch run start $ID
    $ docker compose exec runner pgrep -f fake-harness
    → expect: the process exists on the runner host, not the daemon host

    $ docker network disconnect <net> <runner-container> && sleep 5 && docker network connect …
    → expect: the stream resumes; `winch run watch $ID` shows no gap and no
      duplicate against persisted history

    # with a runner flooding observations faster than they can be persisted:
    → expect: bounded memory and a defined backpressure outcome, recorded as a
      metric rather than a silent drop

## Verification

- Standing scenario suite passes against the remote substrate and the local one.
- Protocol compatibility tests: old fixtures decode, unknown fields ignored.
- Reconnect and duplicate-delivery tests at each bound.
- A soak test asserting bounded memory under sustained event pressure.

## Acceptance criteria

- [ ] The same scenarios pass against local, Docker, and remote substrates.
- [ ] The runner assigns no canonical sequence number.
- [ ] Every bound has a defined, observable behavior when reached.
- [ ] Reconnect loses and duplicates nothing.
- [ ] Protocol evolution stays additive within a major version.

## Deferrals

| Deferred | Owning task |
|---|---|
| Rejecting stale-lease events from a partitioned runner | P5-043 |
| Artifact content transfer | P5-045 |

## Traces to

`docs/contracts.md` §4; `docs/architecture.md` §5 Phase B;
`docs/decisions/0001-modular-monolith-and-runner-boundary.md`; T05
