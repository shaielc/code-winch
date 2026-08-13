# Canonical event schemas

`v1/canonical-event.schema.json` is the source of truth for the version 1 event
envelope and its provider-neutral payload families. JSON Schema draft 2020-12
consumers validate known kinds with the matching conditional payload schema.
Unknown kinds and additive fields remain valid for forward compatibility.
The `fixtures` and `fixtures-invalid` directories contain a passing and failing
example for every provider-neutral event family.

Extensions must use a `domain/vN` namespace. Provider-specific data belongs in
an extension, never a core payload. Consumers must preserve unknown envelope,
source, payload, and extension data when forwarding an event, and must order a
run by `sequence` rather than `occurredAt` or an identifier.

Validate the checked-in schemas and fixtures, including compatibility round
trips, with:

```sh
go test ./pkg/protocol -run 'Test(Schema|Fixtures|Compatibility)'
```

Validation failures exposed by the Go envelope use stable codes and include
only event/run IDs and field names. They never include payload or extension
content. A missing or unrecognized sensitivity must be handled as
`confidential` by policy code; schema validation rejects it at ingestion so a
producer must make the classification explicit.

Provider events pass through `protocol.NormalizeEvent` before persistence. The
default encoded-event limit is 256 KiB. Invalid known payloads, sensitivity, or
extension namespaces degrade to a confidential `diagnostic.emitted` event so a
malformed provider record cannot terminate the run. The diagnostic contains
only a stable code, byte count, and SHA-256 digest. Raw evidence is retained in
the non-authoritative `code-winch.invalid/v1` extension when it fits the limit;
oversized evidence is represented by its digest and must remain available in
the bounded provider-output capture. Logs must use validation codes and
event/run IDs, never payload or extension content.
