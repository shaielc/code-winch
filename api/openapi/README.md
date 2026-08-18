# Run HTTP contract

`code-winch.yaml` is the single source of truth for the public HTTP API. It
defines the initial `/api/v1` run lifecycle contract. Mutating operations use
`Idempotency-Key`; state transitions additionally require the latest strong
`ETag` in `If-Match`. Errors use `application/problem+json` with stable `code`
values and a safe correlation `requestId`.

## Layout

`code-winch.yaml` holds `info`, `servers`, `security`, and the components every
resource shares. Each resource's operations live under `paths/` and its object
schemas under `components/`, assembled into the document by `$ref`:

```
api/openapi/
├── code-winch.yaml          # shared components, and one $ref per resource
├── components/
│   ├── event.yaml           # Event, EventPage
│   ├── health.yaml          # HealthResponse
│   ├── run.yaml             # Run, CreateRunRequest, StopRunRequest
│   └── run-input.yaml       # RunInputRequest, InputAccepted
└── paths/
    ├── health.yaml
    ├── runs.yaml            # /runs
    ├── run.yaml             # /runs/{runId}
    ├── run-start.yaml
    ├── run-stop.yaml
    ├── run-events.yaml
    ├── run-events-stream.yaml
    └── run-input.yaml
```

**Adding a resource means adding a file under `paths/`, a file under
`components/`, and one `$ref` line each**, rather than editing a document every
other task is also editing. A path file reaches shared components as
`../code-winch.yaml#/components/...` and resource schemas as
`../components/<file>.yaml#/components/schemas/...`. The one shared file a new
resource still edits is `oapi-codegen.yaml`, which needs an `import-mapping`
entry per external document; that map has no wildcard form.

### What stays in the root, and why

`RunId`, `RunState`, `Problem`, and `FieldError` stay in `code-winch.yaml`
because the root's own `parameters` and `responses` sections reference them.
Two separate constraints both point the same way:

- oapi-codegen generates each components file in its own pass, and a pass
  cannot dedup against another's output. The root declares a `RunId`
  *parameter*, which generates `type RunId`; a `RunId` *schema* in
  `components/run.yaml` generates a second one, and the package no longer
  compiles.
- openapi-typescript namespace-prefixes any root component whose subtree
  reaches an external file — `components["parameters"]["parameters-RunId"]`
  instead of `["RunId"]`. Keeping the referenced schema in the root keeps the
  browser key names stable.

One prefixed key survives on purpose: the `RunAccepted` response references
`Run`, and `Run` belongs in `components/run.yaml` because that is the schema
tasks extend. Nothing outside the generated file reads
`components["responses"][...]`, so the cost is a name in generated output.

### Generation is two passes

`api-generate-go` runs oapi-codegen once over `code-winch.yaml`, then once per
`components/*.yaml` with `--config oapi-codegen-components.yaml`, each emitting
models into the same `httpapi` package. The `import-mapping` value `-` means
"this reference resolves in this package", so the root pass emits no import and
no duplicate declaration — but it also emits no type, which is why the second
pass is not optional. The components list is discovered from the directory, so
a new file is picked up without editing the Makefile.

Run `make api-generate` after changing the source. This updates the Go server
models in `internal/adapters/transport/httpapi` and the TypeScript declarations
in `web/src/api`. Do not edit either generated file directly.

`make api-check` validates the document, regenerates both outputs, requires a
clean generated diff, and compares it with the frozen v1 compatibility fixture.
Changing `test/contract/openapi/v1.yaml` requires an explicit versioning review;
it must not be updated merely to make a breaking-change check pass.

Implementations should log the operation, `requestId`, and `runId` when known.
They must not log request bodies, event payloads, workspace content, credentials,
or execution handles.

## Local authentication

The phase-one HTTP adapter accepts either a local bearer token or the
`winch_session` secure, HTTP-only, same-site cookie. Deployments must generate
independent authentication and CSRF secrets of at least 32 bytes and configure
the exact browser origin. Every mutation, including bearer-authenticated
requests, supplies that origin and `X-CSRF-Token`; this intentionally uses one
strict browser-facing policy rather than allowing an accidental CSRF bypass.
The session cookie is scoped to `/api/v1` and must only be served over TLS.

Request bodies default to 128 KiB and event pages are capped at 200 records.
Oversized requests receive a stable `payload_too_large` problem. Audit logs
contain operation, request, actor, and run identifiers only: input text,
terminal bytes, stop reasons, workspace content, event payloads, credentials,
and authorization tokens are excluded.

## Resumable event stream

`GET /api/v1/runs/{runId}/events/stream?after_sequence=N` upgrades to a
WebSocket. Authentication stays in the `Authorization` header or secure session
cookie; credentials are never accepted in the URL. The browser must send the
configured exact `Origin`. The server first replays durable events after `N`,
sends `caught_up`, then sends live deltas and periodic heartbeats. Consumers
must persist `lastSequence` and reconnect after it. A `disconnect` message (or
close reason `slow_consumer`) is resumable from that sequence. Authorization is
checked again periodically; `authorization_revoked` requires a fresh session.

Composition roots publish each event to `EventStream.Publish` only after the
event's durable transaction commits. The broadcaster is notification-only and
has a bounded queue per browser, so browser backpressure cannot delay a run.
