# P1-050: Implement run use cases and HTTP binding

**Phase:** 1 — Local single-user vertical slice
**Shape:** seam
**Dependencies:** P1-048 (compile: the composition root and config this registers into), P1-049 (semantic: a started run cannot execute until a runner exists), P1-015 (compile: supervisor primitives), P1-016 (compile: `InputService`), P1-017 (compile: the `httpapi.Backend` interface being implemented)

## Objective

`curl -X POST /api/v1/runs` followed by `/start` produces a live harness whose
events are readable from `/runs/{id}/events` and whose input reaches the
process.

## Scope

- `internal/application/run.go`: `CreateRun`, `GetRun`, `StartRun`, `StopRun`,
  and `ListRunEvents`. Desired state is persisted before any runner interaction;
  `StopRun` is idempotent in `Stopping` and terminal states.
- A driver registry with **append-only registration**: each harness and sandbox
  package registers itself from its own file and the root reads the registry, so
  adding the second adapter (P2-025) or the Docker driver (P3-029) does not edit
  a central switch. Unknown profile names fail before launch with a stable code.
- Profile resolution into a `SandboxSpec`, storing the resolved non-secret
  configuration with the run for reproducibility.
- `httpapi.Backend` over those services, mapping application errors onto the
  five `httpapi.Err*` sentinels so every problem code in the OpenAPI contract has
  a real producing path.
- An `OutboxPublisher` routing `run.input` records to the runner through the
  supervisor and `run.event` records to `httpapi.EventStream.Publish`.
- Register the runner, drivers, supervisor, outbox worker, and full API handler
  in `cmd/winchd`; run a reconciliation sweep at startup so P1-020's
  `Reconcile` finally has a runner to inspect.
- Emit the queue-time, startup-time, and active-run metrics P1-048 declared.
- Split `api/openapi/code-winch.yaml` into per-resource path files under
  `api/openapi/paths/` and shared schemas under `api/openapi/components/`,
  assembled by `$ref`. Every Phase 2 task adds a resource; without the split
  they all edit one file and the phase's width is fiction. The generated Go and
  TypeScript output must be unchanged by the split itself.
- Define a single admission hook on the start path with a permissive default,
  so P2-059 implements a policy rather than editing the start path.

## Non-goals

- Workspace registration. This task treats `workspacePath` as policy-checked
  input beneath a single configured approved root — see Deferrals.
- Credential acquisition or injection. `local-trusted` works only because the
  host user is already logged in; record that as a deployment precondition.
- Retry, admission limits, approvals — P2-057, P2-059, P2-024.

## Runtime reachability

`cmd/winchd`; `POST /api/v1/runs`, `/runs/{id}/start`, `/stop`, `/input`,
`/events`, `/events/stream` on the compose stack, with
`WINCH_HARNESS_PROFILE=fake`.

## Owned surfaces

`internal/application/run.go`, `internal/application/registry.go`,
`internal/adapters/transport/httpapi/backend.go`, `cmd/winchd/main.go`,
`internal/adapters/postgres/migrations/006_*.sql`,
`api/openapi/` (the whole document, once, as it is split).

## Demonstration

    $ docker compose -f deployments/compose.yml up --build -d
    $ ID=$(curl -fsS -X POST localhost:8080/api/v1/runs -H 'Content-Type: application/json' \
        -H "X-CSRF-Token: $T" -H "Origin: http://localhost:8080" \
        -d '{"workspacePath":"/workspace","harnessProfile":"fake","sandboxProfile":"local"}' | jq -r .id)
    $ curl -fsS -X POST localhost:8080/api/v1/runs/$ID/start -H 'If-Match: "1"' …
    $ curl -fsS localhost:8080/api/v1/runs/$ID/events | jq '.events[].kind'
    → expect: lifecycle events, then harness output, in gap-free sequence order

    $ curl -fsS -X POST localhost:8080/api/v1/runs/$ID/input -d '{"kind":"text","text":"echo hi"}' …
    $ curl -fsS localhost:8080/api/v1/runs/$ID/events?after_sequence=N
    → expect: "hi" appears, citing the input command ID

    $ curl -X POST localhost:8080/api/v1/runs/unknown-profile-run/start …
    → expect: 422 with a stable code, and no process started

## Verification

- Service tests against the in-memory ports for every documented transition,
  including concurrent start/stop resolving to one deterministic outcome.
- PostgreSQL integration test for create/start/stop and resolved-profile storage.
- HTTP integration tests producing each problem code from a real error path.
- Delivery test: input accepted before a crash is delivered exactly once after
  restart.

## Acceptance criteria

- [ ] Every transition in `docs/contracts.md` §1 reachable through Phase 1 is
      reachable through the service.
- [ ] A persisted run records its resolved profile and contains no secret value.
- [ ] Registering a new driver requires adding a file, not editing a switch.
- [ ] Adding an API resource requires adding a path file, not editing a shared
      document; `make api-check` still passes and the generated output is
      byte-identical across the split.
- [ ] Published events reach a subscribed WebSocket in sequence order.
- [ ] Startup reconciliation produces a truthful state for a run interrupted by
      a daemon kill.

## Deferrals

| Deferred | Owning task |
|---|---|
| Workspace aggregate; until then a single implicit workspace under one approved root | P2-055 |
| Credential references resolved at launch | P1-054 |
| Bounding concurrent and queued runs | P2-059 |
| Retrying a failed run | P2-057 |

## Traces to

`docs/architecture.md` §4 (API and application services, run supervisor);
`docs/code-structure.md` §2, §6; `docs/contracts.md` §1, §3;
`docs/roadmap.md` Phase 1
