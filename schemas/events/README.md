# Canonical event schemas

`v1/canonical-event.schema.json` is the source of truth for the version 1 event
envelope and its provider-neutral payload families. JSON Schema draft 2020-12
consumers validate known kinds with the matching conditional payload schema.
Unknown kinds and additive fields remain valid for forward compatibility.

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
