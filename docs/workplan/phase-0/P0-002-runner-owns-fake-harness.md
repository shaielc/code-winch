# P0-002: The runner owns the fake harness and the page shows its output live

**Phase:** 0 — A sandbox you can talk to
**Shape:** seam
**Dependencies:** P0-001 (compile: the composition root, config loader, and page shell
this wires the runner into)

## Objective

The sandbox starts a harness process of its own, and a person watching the page or
running `winch stream` sees each line the harness emits as an ordered record, including
the record that says the harness has ended and how.

## Scope

- `internal/runner/` — the runner owns the harness process: spawn, stdout pump,
  incremental line decoding tolerant of arbitrary chunk boundaries, and process reaping.
- The session record, which is the stage's central shape:
  `{ordinal, kind, occurredAt, sensitivity, payload}`. The ordinal is runner-local,
  monotonic, and gap-free within the session. No `sequence`, no `eventId`, no
  `schemaVersion` — `docs/contracts.md` §8 says the control plane adds those, and
  roadmap Stage 0 lists them as deliberately absent.
- Decode the fake harness's dialect: one JSON object per line as
  `{kind, payload, sensitivity?}` (`cmd/fake-harness/main.go:23-36`). The runner adds
  the ordinal and `occurredAt`; the harness's own `kind` and `sensitivity` pass through.
- `record kind: session.terminated` — when the harness process exits, the runner emits a
  terminal record carrying the native exit and its mapped outcome: clean exit →
  `completed`, nonzero exit → `failed`, terminating signal → `stopped`. A runner that
  owns a process and does not report its end leaves the page silently frozen, which is
  why this is not a later task.
- `GET /api/session/stream` — WebSocket, ordinal-ordered, fanning out to more than one
  reader. Origin validated and payloads bounded, per `docs/security.md` §2.
- The fake harness as a controllable **profile**, not a test double (I3). Sandbox
  configuration selects it and passes through its controls: `WINCH_HARNESS_PROFILE`,
  `WINCH_HARNESS_TRANSCRIPT`, `WINCH_HARNESS_DELAY`, and the injectable failures
  `-force-failure`, `-early-exit`, `-malformed-line`. `deployments/README.md` documents
  how to drive it by hand and what it does not prove — it is not a coding agent, it has
  no tool calls, approvals, usage reporting, login, or terminal semantics, and its
  dialect is invented, so an event model built only against it would be modelling the
  fixture.
- Pass the sandbox's session identifier as the fixture's required `-run-id`. Stage 0 has
  no runs; the flag name belongs to the fixture and the fixture is left unchanged.
- `test/contract/attach/` — a golden fixture pinning the record wire format, replacing
  the deleted `schemas/events/v1/fixtures/raw-stream.json` that
  `cmd/fake-harness/main.go:29` still cites, and removing that dangling comment.
- `winch stream` and the page's record list.

## Non-goals

- Persistence. Records live in a bounded in-memory buffer and are gone when the
  container stops; reload shows nothing. P0-003.
- Input of any kind. P0-004, P0-005, P0-006.
- `HarnessDriver`, `HarnessCodec`, capability descriptors, and
  `internal/adapters/harness/`. `docs/roadmap.md` §1 — "a port with one implementation
  is a stage that came too early" — and §3 Stage 1 makes a real vendor CLI the second
  case that forces the seam. Harness ownership stays inside `internal/runner/`.
- More than one record projection. One raw-stream view; roadmap Stage 2 adds the second.
- Malformed-output and slow-reader guarantees. P0-007.

## Runtime reachability

`cmd/winch-sandbox` constructs the runner and starts the harness at boot under the
profile `WINCH_HARNESS_PROFILE` names. The `sandbox` service of
`deployments/compose.yml` reaches it; `winch stream` and the page's record list are the
two hands-on paths.

## Write set

- `internal/runner/runner.go`, `internal/runner/harness.go`, `internal/runner/record.go`,
  `internal/runner/session.go`
- `internal/adapters/transport/attach/stream.go`, `.../server.go`
- `cmd/winch-sandbox/main.go`, `cmd/winch-sandbox/config.go`
- `cmd/winch/stream.go`
- `web/src/attach/App.tsx`, `web/src/attach/RecordList.tsx`,
  `web/src/attach/useSessionStream.ts`
- `test/e2e/scenario_harness_output_test.go`
- `test/contract/attach/record_golden_test.go`, `test/contract/attach/testdata/stream-raw.json`
- `cmd/fake-harness/main.go` (remove the comment citing the deleted fixture)
- `deployments/README.md`

## Contract surfaces

- schema: the session record — `{ordinal, kind, occurredAt, sensitivity, payload}`
- the runner-local ordinal namespace and its allocation
- registry namespace: record kinds, which this task defines
- `record kind: stream.raw`
- `record kind: session.terminated`, and the native-exit → outcome mapping
- API: `GET /api/session/stream` (WebSocket)
- profile namespace: `fake`, and the config keys that drive it —
  `WINCH_HARNESS_PROFILE`, `WINCH_HARNESS_TRANSCRIPT`, `WINCH_HARNESS_DELAY`,
  `WINCH_HARNESS_FORCE_FAILURE`, `WINCH_HARNESS_EARLY_EXIT`,
  `WINCH_HARNESS_MALFORMED_LINE`. P0-006 and P0-007 drive the fake through these;
  they add no key of their own.
- CLI: `winch stream`

## Demonstration

    $ docker compose -f deployments/compose.yml up --build -d
    $ ./bin/winch stream
    → expect: stream.raw records with ordinals 1, 2, 3, … as the transcript plays

    $ xdg-open http://127.0.0.1:8080
    → expect: the same lines appearing in the page as they arrive

    # a stop is distinguishable from a crash
    $ docker compose -f deployments/compose.yml kill -s SIGTERM sandbox
    → expect: a session.terminated record with outcome "stopped"

    # an injected failure says so
    $ WINCH_HARNESS_FORCE_FAILURE=1 docker compose -f deployments/compose.yml up -d --force-recreate
    $ ./bin/winch stream
    → expect: a session.terminated record with outcome "failed" and the native exit code

## Verification

- Standing scenario suite passes against the all-fake profile, with
  `scenario_harness_output_test.go` added: start the sandbox, read the stream, assert
  ordinals are contiguous from 1 and the transcript's lines arrive in order.
- `test/contract/attach/` golden test pins the record wire format byte for byte.
- Unit tests for the decoder across arbitrary chunk boundaries — a record split across
  two reads decodes once, not twice and not never.
- Two concurrent readers of `/api/session/stream` both receive every record.
- `make check`, `make e2e`, `make test-cycle`, `cd web && npm test`.

## Acceptance criteria

- [ ] Ordinals are monotonic and gap-free from 1 within a session, asserted by the
      standing scenario rather than by inspection.
- [ ] A record arriving split across two reads produces exactly one record. Inject by
      feeding the decoder a transcript line in two chunks.
- [ ] `-force-failure` produces `session.terminated` with outcome `failed` and the
      native exit code; SIGTERM produces outcome `stopped`; `-early-exit` produces a
      terminal record rather than a stream that simply stops. All three observable
      through `winch stream`.
- [ ] A harness that exits never leaves the page with no terminal record. Kill the
      process inside the container and watch the record arrive.
- [ ] `-delay` changes the arrival rate of records and nothing else, so the fake is
      controllable at runtime rather than only replayable (I3).
- [ ] `deployments/README.md` states what the fake does not prove, naming at least the
      absence of tool calls, approvals, login, and terminal semantics, and that its
      dialect is invented.
- [ ] No record carries `sequence`, `eventId`, or `schemaVersion` — checkable against
      the golden fixture.
- [ ] `cmd/fake-harness/main.go` no longer cites a path absent from the tree.
- [ ] P0-001's demonstration still passes unchanged: the posture is still served and
      the gates are still green.

## Deferrals

| Deferred | Owning task |
|---|---|
| Records surviving the container, and reading them from an ordinal | P0-003 |
| Malformed harness output degrading to a diagnostic record | P0-007 |
| A reader that stops reading not backpressuring the harness | P0-007 |

## Traces to

`docs/roadmap.md` §3 Stage 0, §4 output path; `docs/contracts.md` §8 (*Records are
ordinal-ordered*, *Reading is snapshot plus stream*, *Harness exit is a record, not a
state machine*); `docs/architecture.md` §4 *Event pipeline and renderers*;
`docs/code-structure.md` §3; ADR-0002, ADR-0005; `docs/security.md` §2;
`docs/state.md` *A code comment cites a fixture that was deleted*.
