# Adapter contract kits

The Go packages in this directory are reusable suites for adapter authors:

- `harness.Run` verifies descriptors, launch construction, JSON input framing,
  flush behavior, safe malformed-record diagnostics, and incremental decoding at
  **every byte boundary** of the supplied fixture.
- `sandbox.Run` verifies start/inspect/stop/cleanup behavior, lifecycle
  idempotency, and each capability advertised by the driver. `sandbox.Validate`
  exposes the same checks as an error-returning function, which is useful for
  negative tests of deliberately dishonest drivers.

Run the focused contracts and deterministic fake-harness process test with:

```sh
go test ./internal/adapters/harness/fake ./internal/adapters/sandbox/fake
```

The fake harness accepts and emits newline-delimited JSON. Output records have
`kind`, `payload`, and optional `sensitivity` fields. Tests control chunk
boundaries, delays, malformed bytes, and exit codes through `fake.Process` and
an injected `application.Clock`; it never sleeps or accesses the network.

Diagnostics deliberately report stable codes and resource identifiers only.
They must not include record contents, user input, environment values, or
resolved credentials.
