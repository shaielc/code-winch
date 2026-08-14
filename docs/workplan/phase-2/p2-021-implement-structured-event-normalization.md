# P2-021: Implement structured event normalization

**Phase:** 2 — Structured experience and second harness
**Shape:** capability
**Dependencies:** P1-050 (semantic: normalized events are observable only once events reach the run API), P1-052 (semantic: the scenario file format is how structured provider output is produced on demand)

## Objective

A harness that emits a tool call, a message delta, or a usage record produces
typed canonical events with explicit sensitivity, instead of raw stream bytes.

## Scope

- Domain validators for the remaining event families in `docs/contracts.md` §2:
  user/assistant/system message and delta, tool call and result, approval
  request and resolution, file change, artifact, usage, workflow link.
- Adapter mapping helpers shared by harness packages, so a second adapter maps
  rather than reimplements.
- Event-size bounds and sensitivity assignment at the point of construction;
  a producer that omits sensitivity yields `confidential`, never `public`.
- Unknown kinds and unknown extension namespaces are preserved and surfaced as
  a diagnostic event, never fatal to the run.
- Extend the fake harness scenario steps to emit each structured family, so the
  new behavior is driveable by hand.

## Non-goals

- Rendering. Projections are P2-022 and P2-023.
- Making extensions authoritative for generic workflow or policy decisions.
- Approval semantics beyond emitting the event — binding, expiry, and policy are
  P2-024.

## Runtime reachability

`GET /api/v1/runs/{id}/events` and `winch run watch` on the compose stack with
a structured scenario selected.

## Owned surfaces

`pkg/protocol/event.go`, `schemas/events/v1/` (payload schemas and fixtures),
`internal/adapters/harness/normalize/`, `internal/adapters/harness/fake/`,
`test/scenarios/structured-*.json`.

## Demonstration

    $ winch run create --harness fake --scenario structured-tools --json | jq -r .id
    $ winch run start $ID && winch run watch $ID --json | jq -r '.kind'
    → expect: message deltas, a tool call, a tool result, and a usage record —
      typed kinds, not raw.stream

    $ winch run watch $ID --json | jq -r 'select(.kind=="tool.call") | .sensitivity'
    → expect: an explicit class, never null

    $ winch run create --harness fake --scenario malformed --json
    → expect: the run survives; a diagnostic event records the failure with a
      stable code and no payload content

## Verification

- Standing scenario suite passes, extended with one scenario asserting typed
  events rather than raw output.
- Valid and invalid fixtures for every family; old fixtures still decode.
- Fuzz the incremental parsers and extension decoding across chunk boundaries.
- Event-size boundary tests at, just under, and just over the limit.

## Acceptance criteria

- [ ] Every family in `docs/contracts.md` §2 has valid and invalid fixtures.
- [ ] Malformed provider data degrades to a diagnostic without terminating a run.
- [ ] Sensitivity is explicit on every persisted event.
- [ ] An unknown extension namespace round-trips unchanged.

## Deferrals

| Deferred | Owning task |
|---|---|
| Rendering these families in the browser | P2-022, P2-023 |
| Approval binding, expiry, and policy evaluation | P2-024 |

## Traces to

`docs/contracts.md` §2; `docs/architecture.md` §4 (event pipeline);
`docs/security.md` §5, T11; `docs/decisions/0002-canonical-events-and-renderers.md`
