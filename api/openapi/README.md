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
