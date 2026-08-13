# Run HTTP contract

`code-winch.yaml` is the single source of truth for the public HTTP API. It
defines the initial `/api/v1` run lifecycle contract. Mutating operations use
`Idempotency-Key`; state transitions additionally require the latest strong
`ETag` in `If-Match`. Errors use `application/problem+json` with stable `code`
values and a safe correlation `requestId`.

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
