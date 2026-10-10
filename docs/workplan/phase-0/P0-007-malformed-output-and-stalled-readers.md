# P0-007: Malformed output and a stalled reader cannot silence the harness

**Phase:** 0 — A sandbox you can talk to
**Shape:** hardening
**Dependencies:** P0-003 (contract: the record store the diagnostic records land in, and
the ordinal a disconnected reader is told to resume from)

## Objective

A harness that emits a line the decoder cannot parse, and a browser that stops reading the
stream, are both visibly contained: the unparseable line becomes a diagnostic record and
the records around it still arrive, and the stalled reader is disconnected with its last
ordinal rather than slowing the harness down.

## Scope

- `record kind: diagnostic` for a line the decoder rejects. The raw bytes are preserved in
  the record, bounded, and classified `confidential` — `docs/security.md` §5 lists
  suspected credential-bearing output under that class, and an unparseable line from an
  agent is exactly that. The decoder resynchronizes at the next newline and keeps going.
- A bound on line length, so a harness that emits a gigabyte with no newline produces a
  diagnostic and a resynchronization rather than growing the runner's memory without
  limit.
- Bounded per-subscriber buffering on `GET /api/session/stream`. A reader that stops
  consuming fills its buffer and is then disconnected with its last ordinal in the close
  reason, so it knows where to refetch from. The harness is never blocked waiting for a
  browser — `docs/contracts.md` §8: "A slow reader is disconnected with its last ordinal
  and never backpressures the harness." P0-002's `Session.append` sends to each
  subscriber while holding the session lock (`internal/runner/session.go:20-27`), so one
  full buffer blocks the harness's output pump, every new subscriber, and the stalled
  reader's own unsubscribe. No subscriber send may block while that lock is held, and
  dropping a subscriber never waits on it.
- A reader that has gone away is released without waiting for another record. P0-002's
  stream handler never reads the socket (`internal/adapters/transport/attach/stream.go`),
  so a client close is noticed only at the next write — and after `session.terminated`
  there is none.
- The disconnect is itself observable: an operational diagnostic record naming that a
  subscriber was dropped and at which ordinal, with no content.

## Non-goals

- Reconnecting automatically and closing the gap. A disconnected reader is told its last
  ordinal and refetches by hand here; automatic no-gap resume is roadmap Stage 2's
  "reconnect mid-session and lose nothing".
- Fuzzing the decoder, schema validation of payloads, and adversarial ANSI, Markdown, or
  URL handling in the page. `docs/security.md` §11 LB06 owns those and roadmap Stage 5 is
  where the security posture becomes real; this task's guarantee is that a malformed line
  is contained, not that every hostile payload is.
- Rate limiting, authentication, and request size limits on the surface as a whole.
  Roadmap Stage 5.
- Anything on the harness's stdin side. P0-006 bounds its own write.

## Runtime reachability

`internal/runner/`'s decoder and the attach surface's stream handler, both reached by
`docker compose -f deployments/compose.yml up --build` under the fake profile. The fake
harness's `-malformed-line` and `-delay` controls drive the first half by hand; a reader
that opens the stream and stops reading drives the second. `winch records --after` shows
what was recorded in both cases.

## Write set

- `internal/runner/harness.go`, `internal/runner/record.go`, `internal/runner/session.go`
- `internal/adapters/transport/attach/stream.go`
- `internal/runner/runner_test.go`, `internal/adapters/transport/attach/stream_test.go`
- `test/e2e/scenario_malformed_output_test.go`, `test/e2e/scenario_stalled_reader_test.go`
- `test/contract/attach/record_golden_test.go` (the diagnostic record's shape)

## Contract surfaces

- `record kind: diagnostic`, and its `confidential` sensitivity
- the decoder's line-length bound and its resynchronization behaviour
- the stream's slow-subscriber disconnect semantics, including the last ordinal reported
  in the close reason

## Demonstration

    # an unparseable line is recorded, and the records around it survive
    $ WINCH_HARNESS_MALFORMED_LINE=1 docker compose -f deployments/compose.yml up --build -d
    $ ./bin/winch records --after 0
    → expect: the transcript's stream.raw records, a diagnostic record carrying the
      bad line, and contiguous ordinals across it — no gap where the bad line was

    # the surface is still alive afterwards
    $ ./bin/winch status
    → expect: the posture, as before

    # a reader that stops reading is dropped, not tolerated.
    # Piping into a process that is not reading fills the pipe, then the client's
    # receive buffer, then the server's send buffer for that subscriber.
    $ WINCH_HARNESS_DELAY=10ms docker compose -f deployments/compose.yml up -d --force-recreate
    $ ./bin/winch stream | (sleep 60; cat) &
    $ ./bin/winch stream        # a second, healthy reader
    → expect: the healthy reader keeps receiving records throughout, and the stalled one
      is closed with a reason naming the last ordinal it received

    $ ./bin/winch records --after 0
    → expect: the harness kept producing records while that reader was stuck — the
      ordinals did not stall — and a diagnostic record names the dropped subscriber

`scenario_stalled_reader_test.go` is the precise check, because it controls exactly when
the reader stops reading; the command above is the hands-on version of the same thing.

## Verification

- Standing scenario suite passes with both scenarios added. `scenario_malformed_output_test.go`
  asserts contiguous ordinals across the bad line and the diagnostic's presence;
  `scenario_stalled_reader_test.go` asserts the harness's ordinal count keeps rising while
  a subscriber is not reading, and that the subscriber is closed.
- Decoder unit tests: a line over the bound, a line that is valid JSON but not a record, a
  truncated line at EOF, and a bad line between two good ones.
- Session unit tests, run with `-race`: a stalled subscriber's unsubscribe and a new
  subscribe both return while its buffer is full; a closed reader's subscription is
  released while no record is being published.
- The golden fixture in `test/contract/attach/` gains the diagnostic record's shape.
- Every earlier scenario still passes unchanged, including with `-delay` set.
- `make check`, `make docker-e2e` (against `make test-env`), `make test-cycle`.

## Acceptance criteria

- [ ] `-malformed-line` produces exactly one `diagnostic` record preserving the rejected
      bytes, and the records before and after it are both present with contiguous
      ordinals. No record is lost and the runner does not exit.
- [ ] A line longer than the bound produces a diagnostic and a resynchronization at the
      next newline. Inject by writing a long unterminated line into the harness's output;
      observe bounded memory and a surface that still answers.
- [ ] A subscriber that stops reading is disconnected, and the close reason names the last
      ordinal it received, so refetching from that ordinal loses nothing.
- [ ] While that subscriber is stuck, the harness keeps producing records and other
      subscribers keep receiving them. This is the guarantee the objective states: inject
      a stalled reader alongside a healthy one and assert the healthy one is unaffected.
- [ ] A stalled subscriber's unsubscribe completes while its buffer is full, and a new
      subscriber can attach meanwhile. Inject a reader that never reads, fill its buffer,
      then cancel it and attach another; neither call blocks.
- [ ] A reader that disconnects while the session is idle is released with no further
      record published. Inject by attaching readers after `session.terminated`, closing
      them, and observing their subscriptions gone.
- [ ] A dropped subscriber leaves an operational diagnostic record naming the ordinal and
      containing no content.
- [ ] The diagnostic record's preserved bytes are classified `confidential`, and appear in
      no log line, no telemetry, and no error body. Inject a canary inside a malformed
      line and grep the container logs for it; expect zero occurrences.
- [ ] Every earlier task's demonstration still passes.

## Deferrals

None.

## Traces to

`docs/roadmap.md` §3 Stage 0; `docs/contracts.md` §8 (*Reading is snapshot plus stream* —
the slow-reader clause); `docs/security.md` §5 (sensitivity classes and the telemetry
field rules), T11, T12, §11 LB06; `docs/architecture.md` §4 *Event pipeline and
renderers*; ADR-0002 (*a malformed parser must degrade to diagnostic/raw events rather
than losing output*).
