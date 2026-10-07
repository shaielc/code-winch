# P0-002: The runner owns the fake harness and the page shows its output live

**Phase:** 0 — A sandbox you can talk to
**Shape:** seam
**Dependencies:** P0-001 (compile: the composition root, config loader, and page shell
this wires the runner into)

## Objective

The sandbox starts a harness process of its own, and a person watching the page or
running `winch stream` sees what the harness emits as ordered records, including the
record that says the harness has ended and how. The runner knows a harness only as a
process; what its bytes mean is the harness codec's business.

## Scope

- `internal/runner/` — the runner owns the harness *process* and nothing about its
  format: spawn, stdout pump, process reaping. It expects only what any harness
  provides — stdout, stderr, stdin, and an exit status or terminating signal — and it
  pumps stdout bytes, as read, into a codec. It never parses harness output
  (`docs/code-structure.md` §3; roadmap D5 leaves open that a real harness needs a PTY).
- A **fake harness codec** as its own unit, separate from the runner, holds everything
  that knows the fake's JSON-lines dialect (`cmd/fake-harness/main.go:23-36`):
  incremental, tolerant of arbitrary chunk boundaries, with a flush at end of output.
  It yields unsequenced items `{kind, sensitivity, payload}` and decides what an
  unparseable line becomes. Unknown fields, a default sensitivity, and the `stream.raw`
  kind the fake happens to emit (`cmd/fake-harness/main.go:76-77`) are the codec's
  concerns. It is not the `HarnessDriver` port (see Non-goals); the composition root
  hands the runner this one codec.
- The session record, which is the stage's central shape:
  `{ordinal, kind, occurredAt, sensitivity, payload}`. The runner allocates `ordinal`
  (runner-local, monotonic, gap-free within the session) and `occurredAt`. `kind`,
  `sensitivity`, and `payload` are codec output that the runner carries and does not
  interpret. No `sequence`, no `eventId`, no `schemaVersion` — `docs/contracts.md` §8
  says the control plane adds those, and roadmap Stage 0 lists them as deliberately
  absent.
- `record kind: session.terminated` — the runner emits it from the process exit status
  alone, never from output: clean exit → `completed`, nonzero exit → `failed`,
  terminating signal → `stopped`. A runner that owns a process and does not report its
  end leaves the page silently frozen, which is why this is not a later task.
- Signal repair in the fake: on SIGTERM it emits its last observation and then
  `os.Exit(0)` (`cmd/fake-harness/main.go:67-75`), so a stop is indistinguishable from a
  clean exit and `stopped` could never be observed. Make the fake end by the signal
  (restore the default action and re-raise after the final observation). This concerns
  exit status, not format.
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
- The `HarnessDriver` port, capability descriptors, and `internal/adapters/harness/`.
  `docs/roadmap.md` §1 — "a port with one implementation is a stage that came too
  early" — applies to the port; §3 Stage 1 makes a real vendor CLI the second case that
  forces it. The codec *unit* above is not that port.
- More than one record projection. One raw-stream view; roadmap Stage 2 adds the second.
- Malformed-output and slow-reader guarantees. P0-007. Until then the codec drops a
  line it cannot parse and the runner is unaffected.

## Runtime reachability

`cmd/winch-sandbox` constructs the runner and starts the harness at boot under the
profile `WINCH_HARNESS_PROFILE` names. The `sandbox` service of
`deployments/compose.yml` reaches it; `winch stream` and the page's record list are the
two hands-on paths.

## Write set

- `internal/runner/runner.go`, `internal/runner/harness.go`, `internal/runner/record.go`,
  `internal/runner/session.go`
- `internal/runner/fakecodec/codec.go` (the fake's dialect; nothing else imports its format)
- `internal/adapters/transport/attach/stream.go`, `.../server.go`
- `cmd/winch-sandbox/main.go`, `cmd/winch-sandbox/config.go`
- `cmd/winch/stream.go`
- `web/src/attach/App.tsx`, `web/src/attach/RecordList.tsx`,
  `web/src/attach/useSessionStream.ts`
- `test/e2e/scenario_harness_output_test.go`
- `test/contract/attach/record_golden_test.go`, `test/contract/attach/testdata/stream-raw.json`
- `cmd/fake-harness/main.go` (signal repair; remove the comment citing the deleted fixture)
- `deployments/README.md`

## Contract surfaces

- schema: the session record — `{ordinal, kind, occurredAt, sensitivity, payload}`;
  the runner owns `ordinal` and `occurredAt`, the codec owns the rest
- the runner-local ordinal namespace and its allocation
- registry namespace: record kinds, which this task defines
- `record kind: stream.raw` — the fake codec's output, not a runner guarantee
- `record kind: session.terminated`, and the native-exit → outcome mapping
- the codec unit's input/output shape (bytes in, unsequenced items out)
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
    → expect: records with ordinals 1, 2, 3, … as the transcript plays (the fake codec
      yields them as kind stream.raw)

    $ xdg-open http://127.0.0.1:8080
    → expect: the same lines appearing in the page as they arrive

    # a stop is distinguishable from a crash
    # signal the harness child, not the sandbox: killing the sandbox closes both clients
    $ docker compose -f deployments/compose.yml exec sandbox sh -c 'kill -TERM "$(pidof fake-harness)"'
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
- Unit tests for the fake codec across arbitrary chunk boundaries — a record split
  across two reads decodes once, not twice and not never.
- Runner tests with a stub codec and a non-JSON child: ordinals stay contiguous and the
  terminal mapping holds, which shows the runner does not depend on the fake's format.
- Two concurrent readers of `/api/session/stream` both receive every record.
- `make check`, `make e2e`, `make test-cycle`, `cd web && npm test`.

## Acceptance criteria

- [ ] Ordinals are monotonic and gap-free from 1 within a session, asserted by the
      standing scenario rather than by inspection.
- [ ] A record arriving split across two reads produces exactly one record. Inject by
      feeding the fake codec a transcript line in two chunks.
- [ ] The runner never reads harness output: with a stub codec and a child that prints
      non-JSON bytes, ordinals are contiguous and the terminal record is correct.
      Ordinal contiguity, terminal mapping, and fan-out hold regardless of the codec.
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
- [ ] `cmd/fake-harness/main.go` no longer cites a path absent from the tree, and
      SIGTERM now reaches the runner as a signal exit rather than exit 0.
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
