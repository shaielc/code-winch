# P0-020: Keep run output when an event append fails

**Phase:** 0 — Foundation repair
**Shape:** hardening
**Dependencies:** P0-008 (revision: its object is the observation consumer and terminal transition P0-008's start use case produced, which drop an observation whose append fails)

## Objective

A storage failure on one harness output event neither loses that output nor
lets the run report `completed` without it.

## Scope

- Keep an observation whose append fails and retry it, in runner order, before
  any later observation of the same execution is appended. Today the consumer
  logs the failure and moves on (`cmd/winchd/main.go:122-125`), so every
  sequence that was committed stays contiguous while the output is gone.
- Write the successful terminal transition only once every observation of the
  execution is stored. When retries are exhausted, clean up the execution so no
  harness outlives it, and record the run `failed` with a stable reason code in
  its terminal `run.lifecycle` event if the store accepts that write; if it does
  not, the run stays non-terminal rather than claiming an outcome.
- Do not release the run's lease or forget its execution before the terminal
  event and transition are stored or given up on. Today `finish` does both
  whatever the append and transition returned (`internal/application/start.go:242-260`).
- Regression tests for a failed output append and a failed terminal append,
  against PostgreSQL with an injected failure.

## Non-goals

- Keeping undelivered observations across a daemon restart.
- A start interrupted before it holds a lease — P0-021.
- Writes by an execution that has lost its lease — P0-022.
- Outbox delivery or live streaming — P0-009, P0-010.

## Runtime reachability

- **Composition root:** `cmd/winchd` (the observation consumer).
- **Profile:** `harnessProfile=fake`, `sandboxProfile=local`, PostgreSQL storage.
- **Command:** `winch run start`, `winch run events`, `winch run get`.

## Write set

- `internal/application/start.go`
- `internal/application/start_test.go`
- `cmd/winchd/main.go` (observation consumer)
- `test/e2e/append_failure_test.go`
- `docs/contracts.md` §1 (the reason code)

## Contract surfaces

- persisted aggregate: a run's `completed` transition — written only after
  every observation of its execution is stored
- event: the `run.lifecycle` `reasonCode` for a run failed by exhausted
  persistence (`reasonCode` is a free string in
  `schemas/events/v1/canonical-event.schema.json`, so the schema does not change)

## Demonstration

Against a daemon started with a transcript that prints a canary:

    $ printf 'echo persist-canary\nexit\n' > /tmp/canary.txt
    $ WINCH_FAKE_HARNESS_TRANSCRIPT=/tmp/canary.txt winchd &
    $ psql "$WINCH_DATABASE_URL" <<'SQL'
    CREATE SEQUENCE demo_append_failures;
    CREATE FUNCTION demo_fail_append() RETURNS trigger LANGUAGE plpgsql AS $$
    BEGIN
      IF NEW.payload->>'data' LIKE '%persist-canary%' THEN
        IF nextval('demo_append_failures') = 1 THEN
          RAISE EXCEPTION 'injected append failure';
        END IF;
      END IF;
      RETURN NEW;
    END $$;
    CREATE TRIGGER demo_fail_append BEFORE INSERT ON run_events
      FOR EACH ROW EXECUTE FUNCTION demo_fail_append();
    SQL
    $ RUN_ID=$(winch run create --workspace /tmp/ws --harness fake --sandbox local)
    $ winch run start "$RUN_ID" && sleep 2
    $ winch run events "$RUN_ID"
    → expect: a `stream.raw` event containing `persist-canary`, sequences gap-free
    $ winch run get "$RUN_ID"
    → expect: `completed`

Make the failure permanent — drop the inner `IF` so every append of the canary
raises — and repeat with a new run:

    $ winch run get "$RUN_ID"
    → expect: never `completed`; `failed`, with the persistence reason code on
      the last `run.lifecycle` event from `winch run events`
    $ pgrep -c -x fake-harness
    → expect: 0

Remove the injection afterwards with `DROP TRIGGER demo_fail_append ON run_events;
DROP FUNCTION demo_fail_append(); DROP SEQUENCE demo_append_failures;`.

## Verification

- `make check` and `make test-integration` pass.
- `make e2e` passes, including `test/e2e/append_failure_test.go`, and the
  standing scenarios pass unchanged.
- Contract suites for fake harness and local sandbox still pass.

## Acceptance criteria

- [ ] With one injected failure on an output append, `winch run events` includes
  that output with gap-free sequences, and the run is `completed`.
- [ ] With every append of that output failing, the run never reaches
  `completed`: it ends `failed` with a stable reason code documented in
  `docs/contracts.md` §1, and no `fake-harness` process survives.
- [ ] A failed terminal `run.lifecycle` append does not release the run's lease
  or forget its execution before the terminal event and transition are stored
  or given up on.
- [ ] Each of the three has a test that fails without the change.
- [ ] I1, I2, and I3 still hold.

## Deferrals

| Deferred | Owning task |
|---|---|
| Observations still pending when the daemon stops | Phase 1 — the *Truthful terminal state for a run the daemon was stopped under* row in [`../phase-1/README.md`](../phase-1/README.md) |

## Traces to

- `docs/contracts.md` §1 (terminal states) and §2 (event envelope)
- `docs/architecture.md` §4 (the supervisor sequences events before persistence)
- [`../post-mortems/2026-09-16-a-plan-defect-read-from-one-implementation.md`](../post-mortems/2026-09-16-a-plan-defect-read-from-one-implementation.md)
  — the audit finding and its reproduction
- P0-008 — the implementation this task revises
