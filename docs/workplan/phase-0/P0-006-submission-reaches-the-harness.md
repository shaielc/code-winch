# P0-006: A submission reaches the harness and the harness's reply comes back

**Phase:** 0 — A sandbox you can talk to
**Shape:** capability
**Dependencies:** P0-005 (contract: the accepted input command and its typed payload),
P0-002 (contract: harness process I/O, which this adds the stdin side of)

## Objective

Type a line, and the fake harness answers it — the submission is written to the harness's
stdin and its reply arrives as the next records in the same ordered list. This is the
stage's hand check.

## Scope

- Encode an accepted submission into the frame the fake harness reads from stdin:
  `{"id": "<submission id>", "text": "<text>"}` (`cmd/fake-harness/main.go:39-42`),
  newline-delimited.
- Deliver after acceptance, not instead of it: the submission is recorded first, then
  written. A write that fails after acceptance produces a diagnostic record naming the
  submission, so an accepted-but-undelivered submission is visible rather than silent.
- The write is bounded: a harness that stops reading its stdin causes a deadline to
  expire and a diagnostic record, not a handler that blocks forever. This belongs here
  rather than in P0-007 because it is a property of the pipe this task opens.
- Check input-capability before delivering. The harness is input-capable while its process
  is running and its stdin is open; once `session.terminated` has been emitted it is not,
  and a submission is refused with `INPUT_UNSUPPORTED` rather than written to a closed
  pipe. `docs/contracts.md` §8 — "the session accepts no further input", and "there is no
  retry and no new attempt: restarting means starting another sandbox".
- The reply correlates to the submission: the harness echoes the submission id, and the
  resulting `stream.raw` record keeps it, so a reader can tell which line was answered.

## Non-goals

- A run state machine, attempts, and retry. `docs/contracts.md` §8 is explicit that
  harness exit is a record and not a state machine; the run aggregate is roadmap
  Stage 3 and its lifecycle Stage 4.
- Restarting the harness inside a live sandbox. Restarting means starting another
  sandbox, which is the property that keeps the run aggregate the control plane's job.
- Terminal semantics — a PTY, raw mode, resize, TTY detection. The fake speaks JSON lines
  on pipes and needs none; roadmap deferred decision D5 owns the question and its trigger
  is the first real harness requiring them.
- Malformed harness *output*, and a browser that stops reading the stream. Those are the
  output side of the same concern and belong to P0-007, which does not depend on this
  task and may land before it.

## Runtime reachability

`internal/runner/` holds the harness's stdin from the moment it spawns the process in
P0-002; this task writes to it. `docker compose -f deployments/compose.yml up --build`
reaches it; `winch input` and the page's composer are the two hands-on paths, and
`winch stream` or the page shows the reply.

## Write set

- `internal/runner/harness.go`, `internal/runner/input.go`, `internal/runner/session.go`
- `test/e2e/scenario_harness_echo_test.go`
- `test/contract/attach/input_refusal_test.go` (the after-exit refusal)
- `cmd/winch/input.go` (surface the correlation identifier in its output)

## Contract surfaces

- the harness input-frame encoding: an accepted submission → the harness's stdin frame
- the input-capable predicate, and `INPUT_UNSUPPORTED` as the refusal once the session
  has terminated
- the correlation from a submission identifier to the records that answer it

## Demonstration

    $ docker compose -f deployments/compose.yml up --build -d
    $ ./bin/winch input "ping"
    → expect: accepted, with an ordinal and a submission id

    $ ./bin/winch records --after 0
    → expect: an input.submitted record carrying "ping", then stream.raw records from the
      harness answering it and citing the same submission id

    $ xdg-open http://127.0.0.1:8080
    → type a line and send
    → expect: the fake harness's reply appears in the page without a reload.
      This is the Stage 0 hand check in docs/roadmap.md §2.

    # after the harness has gone, a submission is refused rather than lost
    $ WINCH_HARNESS_EARLY_EXIT=1 docker compose -f deployments/compose.yml up -d --force-recreate
    $ sleep 2 && ./bin/winch input "too late"
    → expect: INPUT_UNSUPPORTED, and no new record

## Verification

- Standing scenario suite passes with `scenario_harness_echo_test.go` added: submit, wait
  for the reply, assert the reply cites the submission and arrives at a later ordinal
  than the submission.
- `test/contract/attach/input_refusal_test.go` gains the after-exit case.
- Unit test: a submission whose stdin write fails produces a diagnostic record naming the
  submission id and not its text.
- The fake profile still passes every earlier scenario, including with `-delay` set, so
  delivery does not depend on the harness replying promptly.
- `make check`, `make e2e`, `make test-cycle`, `cd web && npm test`.

## Acceptance criteria

- [ ] A submission made through the page produces the harness's reply in the page with no
      reload and no manual step.
- [ ] The reply is correlatable to the submission by identifier, so two submissions in
      flight are distinguishable.
- [ ] The submission is recorded before it is written. Inject a stdin write failure (close
      the harness's stdin, or kill the process between acceptance and delivery) and
      observe the `input.submitted` record present **and** a diagnostic record naming the
      submission — never an accepted submission with no trace.
- [ ] After `session.terminated`, every submission is refused with `INPUT_UNSUPPORTED` and
      nothing is written to a closed pipe. Inject with `-early-exit` and with a SIGTERM to
      the harness, and observe the same refusal in both cases.
- [ ] No submission is written to the harness twice, including when its idempotency key is
      replayed. Replay the key and count the harness's replies: expect one.
- [ ] The submitted text appears in no log line or diagnostic. Canary as in P0-005.
- [ ] `cmd/fake-harness/` is unchanged — the stdin dialect is the fixture's and this task
      speaks it rather than altering it.
- [ ] A harness that stops reading stdin does not block the input handler. Inject by
      stopping the harness process with SIGSTOP and submitting; expect the deadline to
      expire, a diagnostic record, and a surface that still answers `winch status`.
- [ ] Every earlier task's demonstration still passes.

## Deferrals

None.

## Traces to

`docs/roadmap.md` §2 (Stage 0 hand check), §1 rule 4 (*The same submission reaching the
harness is observable by the harness's reply*), §3 Stage 0, §4 input path, §6 D5;
`docs/contracts.md` §3, §8 (*Input is accepted or refused, never queued silently*,
*Harness exit is a record, not a state machine*); `docs/architecture.md` §4 *Harness
adapters*; `docs/code-structure.md` §3; ADR-0003, ADR-0005.
