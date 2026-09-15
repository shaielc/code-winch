# P0-019: Harden start refusal, control convergence, and coordinator placement

**Phase:** 0 — Foundation repair
**Shape:** hardening
**Dependencies:** P0-008 (revision: this task's object is the coordinator, the
profile refusal, and the daemon fake configuration that P0-008 produced)

## Objective

A start refused for an unsupported profile says so in its own problem code, a
finished or abandoned run leaves a terminal supervisor control record, and the
run coordinator lives beside the supervisor instead of inside the application
layer.

## Scope

- Add a problem code for the refusal P0-008 already performs: an
  `unsupported_run_profile` response on `POST /api/v1/runs/{runId}/start`,
  declared in `api/openapi/code-winch.yaml` and mapped in the transport
  adapter, replacing the `run_state_conflict` a valid `created` run currently
  receives.
- Converge the supervisor control record with the run's lifecycle: every
  terminal transition the coordinator writes — successful exit, failed exit,
  and the abandoned launch — also records a terminal `desired_state`, so a
  restart no longer finds a completed run described as running.
- Move the coordinator out of `internal/application/` to
  `internal/supervisor/`, so the application layer stops importing outward, and
  move the `fake`/`local` driver names and the fake-profile redaction policy to
  the composition root that already selects the profile.
- Load the daemon's fake-harness controls through `internal/platform/config`
  rather than bare `os.Getenv`, and declare those configuration keys.
- Negative tests for each: the refusal's code and status, a terminal control
  record after each of the three terminal paths, and the dependency direction.

## Non-goals

- Restart reconciliation itself. `Supervisor.Reconcile` exists and nothing
  calls it; wiring it is not this task's work, and the converged control record
  is what makes it correct when someone does.
- Any new event kind, including a lifecycle event for an abandoned launch.
- Input, outbox delivery, WebSocket streaming, or stop — P0-009 to P0-011.
- The memory store profile — P0-014.
- Retry of a failed run. `domain.RunCommandRetry` exists and no API exposes it.

## Runtime reachability

- **Composition root:** `cmd/winchd`.
- **Profile:** `harnessProfile=fake`, `sandboxProfile=local`, PostgreSQL storage.
- **Command:** `winch run start` against a run whose persisted profiles are not
  `fake`/`local`; `winch run get` after a run completes.

## Write set

- `internal/supervisor/` (the coordinator moves here, with its tests)
- `internal/application/runexec/` (removed)
- `internal/adapters/transport/httpapi/server.go` (problem-code mapping)
- `api/openapi/code-winch.yaml` (`startRun` responses, problem component)
- `cmd/winchd/main.go` (driver names, redaction policy, configuration loading)
- `internal/platform/config/` (fake-harness configuration keys)
- `test/contract/openapi/` (fixture for the new response)

## Contract surfaces

- API: `POST /api/v1/runs/{runId}/start` — its response set and the
  `unsupported_run_profile` problem code
- port: `application.RunRuntime` implementation package
- persisted aggregate: the `runs` control record's `desired_state` on the
  transitions P0-008 writes — successful exit, failed exit, abandoned launch.
  A later seam that adds a transition of its own (P0-009, P0-011) writes it
  under the same rule and redefines nothing, so neither takes an edge.
- configuration keys: the daemon's fake-harness controls

## Demonstration

    $ winch run create --workspace /tmp/ws --harness fake --sandbox local
    $ psql -c "update runs set harness_profile='claude' where id='<uuid>'"
    $ winch run start $RUN_ID
    → expect: refusal naming the profile, code `unsupported_run_profile`,
      not `run_state_conflict`; no `fake-harness` process started

    $ winch run create --workspace /tmp/ws --harness fake --sandbox local
    $ winch run start $RUN_ID && sleep 2 && winch run get $RUN_ID
    $ psql -tAc "select desired_state from runs where id='<uuid>'"
    → expect: run `completed`, and `desired_state` terminal rather than
      `running`

    $ WINCH_FAKE_BINARY=/nonexistent winchd   # in a second daemon
    $ winch run start $RUN_ID
    $ psql -tAc "select desired_state from runs where id='<uuid>'"
    → expect: run `failed` (P0-008 behavior, kept) and `desired_state` terminal

    $ grep -rn "code-winch/internal/supervisor" --include=*.go . | grep -v _test
    → expect: no match under internal/application/

## Verification

- `make check` and `make test-integration` pass.
- `make e2e` passes unchanged — the start scenario is P0-008's and this task
  does not revise it.
- `make api-check` passes, and `api-compat` accepts the added response as
  non-breaking.
- Contract suites for fake harness and local sandbox still pass.

## Acceptance criteria

- [ ] Starting a run whose persisted profiles are not `fake`/`local` answers
  `unsupported_run_profile` with a declared status, and the response is
  declared for `startRun` in the OpenAPI document.
- [ ] After a run completes, fails, or is abandoned mid-launch, its persisted
  `desired_state` is terminal.
- [ ] No package under `internal/application/` imports `internal/supervisor`,
  and no `fake` or `local` driver name appears outside `cmd/winchd`.
- [ ] The daemon's fake-harness controls are read through the configuration
  loader, and the deployment documentation names them as configuration.
- [ ] Each of the four has a test that fails without the change.
- [ ] I1, I2, and I3 still hold.

## Deferrals

| Deferred | Owning task |
|---|---|
| Memory-profile equivalents of these guarantees | P0-014 |

## Traces to

- [`../post-mortems/2026-09-15-a-coordinator-with-nowhere-to-live.md`](../post-mortems/2026-09-15-a-coordinator-with-nowhere-to-live.md)
  — the two declaration defects in P0-008 that produced findings 2 to 5
- `docs/code-structure.md` §2 (dependency rule)
- `docs/contracts.md` §1 (run lifecycle — terminal states, which the control
  record must reach too)
- `api/openapi/code-winch.yaml` — `startRun` responses and the `Problem`
  component, the only authority for problem codes
- P0-008 — the implementation this task revises
