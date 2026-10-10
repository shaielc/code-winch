# P0-005: A submission is accepted or refused, recorded with an ordinal, and survives a reload

**Phase:** 0 — A sandbox you can talk to
**Shape:** seam
**Dependencies:** P0-003 (contract: the record store and ordinal allocation a submission
is appended to), P0-004 (semantic: the page submission this wires has no control to
submit from until P0-004 adds one)

## Objective

Submit a line and the runner either records it with an ordinal of its own — visible in
the record list, and still there after a reload — or refuses it with a code that says
why; it is never accepted silently and then dropped.

## Scope

- `POST /api/session/input` — one typed payload, `text` for this stage, plus a
  client-generated idempotency key scoped to the session. There is no run, so the key is
  session-scoped rather than run-scoped, per `docs/contracts.md` §8.
- `record kind: input.submitted` — an accepted submission becomes a record in the same
  ordered store as harness output, so the exchange reads as one sequence.
- Acceptance returns the record's ordinal and the submission's identifier. A repeated key
  returns the first submission's identifier and kind, and appends nothing.
- Refusal uses the §3 codes that have meaning without a run — `INPUT_INVALID`,
  `INPUT_UNSUPPORTED`, `INPUT_UNAUTHORIZED` — with a diagnostic naming the ordinal and
  the input kind and never the payload content.
- Bounded payload size, rejected as `INPUT_INVALID` rather than read into memory.
- `winch input "<text>"`, and the page's composer submits instead of echoing locally.
- P0-004 deliberately used positional React keys for its transient append-only local
  echo. Do not carry that implementation detail into persisted submissions: when the
  local list is replaced, render submitted records with stable runner-backed identity
  (the submission identifier or other stable record identity), not `key={index}`.

## Non-goals

- Delivery to the harness. An accepted submission is recorded and goes no further; the
  harness never sees it and does not reply. P0-006.
- `interrupt`, `terminal_bytes`, and `resize`. `docs/contracts.md` §3 names them in the
  first delivery slice of the control plane's contract; Stage 0's hand check is a line of
  text, and raw terminal input is a separately authorized capability whose authorization
  story does not exist until roadmap Stage 5.
- A bearer token, and therefore any real source of `INPUT_UNAUTHORIZED` beyond a
  malformed or absent session scope. Roadmap Stage 0 puts authentication beyond a
  loopback binding outside the stage; `docs/security.md` §11 LB11 owns the rest.
- An outbox, a transactional commit spanning two stores, and `INPUT_STALE_STATE` /
  `INPUT_RUN_NOT_FOUND`. Those are the control plane's, roadmap Stage 4.

## Runtime reachability

`cmd/winch-sandbox` wires the input handler to the runner, which appends to the store
P0-003 built. `docker compose -f deployments/compose.yml up --build` reaches it;
`winch input` and the page's composer are the two hands-on paths.

## Write set

- `internal/runner/input.go`, `internal/runner/session.go`, `internal/runner/record.go`
- `internal/adapters/transport/attach/input.go`, `.../server.go`
- `cmd/winch-sandbox/main.go`
- `cmd/winch/input.go`
- `web/src/attach/App.tsx`, `web/src/attach/Composer.tsx`, `web/src/attach/useSubmitInput.ts`
- `test/e2e/scenario_input_recorded_test.go`
- `test/contract/attach/input_refusal_test.go`

## Contract surfaces

- API: `POST /api/session/input`
- `record kind: input.submitted`
- the session-scoped input idempotency-key namespace
- the refusal codes `INPUT_INVALID`, `INPUT_UNSUPPORTED`, `INPUT_UNAUTHORIZED` as this
  surface uses them
- CLI: `winch input`

## Demonstration

    $ docker compose -f deployments/compose.yml up --build -d
    $ ./bin/winch input "hello"
    → expect: accepted, with the ordinal it was recorded at

    $ ./bin/winch records --after 0
    → expect: an input.submitted record carrying "hello", interleaved with the harness's
      own records in ordinal order

    # the same key twice appends once
    $ curl -fsS -XPOST http://127.0.0.1:8080/api/session/input \
        -H 'content-type: application/json' \
        -d '{"key":"k1","kind":"text","text":"twice"}'
    $ curl -fsS -XPOST http://127.0.0.1:8080/api/session/input \
        -H 'content-type: application/json' \
        -d '{"key":"k1","kind":"text","text":"twice"}'
    $ ./bin/winch records --after 0 | grep -c twice
    → expect: 1

    # an unsupported kind is refused by name
    $ curl -sS -XPOST http://127.0.0.1:8080/api/session/input \
        -H 'content-type: application/json' \
        -d '{"key":"k2","kind":"resize","cols":80,"rows":24}'
    → expect: INPUT_UNSUPPORTED, and no new record

    $ xdg-open http://127.0.0.1:8080     # type a line, send, then reload
    → expect: the line is still there after the reload, now carrying an ordinal

## Verification

- Standing scenario suite passes with `scenario_input_recorded_test.go` added: submit,
  read the records, assert the submission is present with an ordinal and in order.
- `test/contract/attach/input_refusal_test.go` pins each refusal code to its cause and
  asserts no diagnostic contains the payload text.
- Idempotency unit tests: same key twice, same key with a different payload, different
  keys with the same payload.
- Bounded-size test: a payload over the limit is refused as `INPUT_INVALID` and the
  process memory does not grow by the payload's size.
- `make check`, `make docker-e2e` (against `make test-env`), `make test-cycle`,
  `cd web && npm test`.

## Acceptance criteria

- [ ] An accepted submission is a record with an ordinal, in the same ordered store as
      harness output, and is present after a reload.
- [ ] Every refusal path returns one of the three codes and appends no record. The
      failures a person can inject: an empty payload and an unknown field
      (`INPUT_INVALID`), `kind: "resize"` (`INPUT_UNSUPPORTED`), a malformed session
      scope (`INPUT_UNAUTHORIZED`). Each is observable as the code plus an unchanged
      record list.
- [ ] No submission is accepted and then dropped: for every 2xx response there is a
      record at the ordinal the response named. Checkable by submitting and refetching.
- [ ] A repeated idempotency key returns the first submission's identifier and kind and
      appends nothing — the record count is unchanged.
- [ ] A payload over the bound is refused rather than buffered.
- [ ] No diagnostic, log line, or error body contains the submitted text. Inject a
      canary string, submit it in a way that is refused, and grep the container logs and
      the response for it; expect zero occurrences (`docs/security.md` §5).
- [ ] The page's composer now submits; the local-echo behaviour P0-004 shipped is
      replaced rather than left beside it.
- [ ] The P0-004 local list's `key={index}` does not survive this replacement.
      Persisted submitted records use stable runner-backed identity as their React key.
- [ ] P0-001 through P0-004's demonstrations still pass, except P0-004's reload step,
      which this task deliberately changes — the brief records that the local echo
      becomes a persisted submission here.

## Deferrals

| Deferred | Owning task |
|---|---|
| The submission reaching the harness and the harness replying | P0-006 |
| Refusing a submission once the harness has exited | P0-006 |

## Traces to

`docs/roadmap.md` §1 rule 4 (*The same control persisting what you submitted is
observable by reloading*), §3 Stage 0, §4 input path; `docs/contracts.md` §3, §8
(*Input is accepted or refused, never queued silently*); `docs/security.md` §5
(telemetry, errors, and audit fields), T10, T11; `docs/architecture.md` §6 (a session is
not a run); ADR-0005, ADR-0006.
