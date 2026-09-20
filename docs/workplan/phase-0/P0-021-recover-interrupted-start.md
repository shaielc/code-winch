# P0-021: Recover a start interrupted before launch

**Phase:** 0 — Foundation repair
**Shape:** hardening
**Dependencies:** P0-008 (revision: its object is the start use case P0-008 produced, which commits `queued` and then abandons the run when lease acquisition fails or the request is cancelled)

## Objective

A start whose request is cancelled, or whose lease acquisition fails, never
leaves the run permanently `queued`: the run launches, or it reaches a terminal
state a person can read back.

## Scope

- Once `created → queued` is committed, finish lease acquisition and launch
  under a context the daemon owns rather than the HTTP request's. Today both
  run under the request context (`internal/application/start.go:121-133`), so a
  caller that disconnects mid-acquisition leaves the run `queued`, and a retry
  is refused with `409 run_state_conflict`.
- When lease acquisition itself fails, move the attempt to a terminal state
  instead of returning with it `queued`. The domain lets `queued` leave only
  through `acquire_lease` or `cancel` (`internal/domain/run.go:188-194`). If a
  truthful exit needs a `queued → failed` transition, add it to the state
  machine and to the lifecycle diagram in `docs/contracts.md` §1 in this task.
- State in `docs/contracts.md` §1 what an accepted start does after its caller
  disconnects.
- Tests: a request cancelled during a delayed acquisition (PostgreSQL, e2e),
  and an acquisition that returns an error (unit).

## Non-goals

- Making `start` idempotent. It stays conditional, as `docs/contracts.md` §1
  describes.
- Retrying a failed run.
- A run left `queued` by a daemon that stopped mid-start.
- An event append failing after launch — P0-020.
- Writes by an execution that has lost its lease — P0-022.

## Runtime reachability

- **Composition root:** `cmd/winchd`.
- **Profile:** `harnessProfile=fake`, `sandboxProfile=local`, PostgreSQL storage.
- **Command:** `winch run start` and `winch run get`; a raw HTTP start that
  disconnects early.

## Write set

- `internal/application/start.go`
- `internal/application/start_test.go`
- `test/e2e/start_cancel_test.go`
- `docs/contracts.md` §1
- `internal/domain/run.go` and `internal/domain/run_test.go` — only if a
  `queued → failed` transition is added

## Contract surfaces

- API: `POST /api/v1/runs/{runId}/start` — what an accepted start does when its
  caller disconnects
- persisted aggregate: how a run attempt leaves `queued` when its lease cannot
  be acquired, including any new domain transition

## Demonstration

Delay lease acquisition in the database, then disconnect before it finishes:

    $ psql "$WINCH_DATABASE_URL" <<'SQL'
    CREATE FUNCTION demo_slow_acquire() RETURNS trigger LANGUAGE plpgsql AS $$
    BEGIN
      IF OLD.supervisor_lease_token IS NULL AND NEW.supervisor_lease_token IS NOT NULL THEN
        PERFORM pg_sleep(2);
      END IF;
      RETURN NEW;
    END $$;
    CREATE TRIGGER demo_slow_acquire BEFORE UPDATE ON runs
      FOR EACH ROW EXECUTE FUNCTION demo_slow_acquire();
    SQL
    $ RUN_ID=$(winch run create --workspace /tmp/ws --harness fake --sandbox local)
    $ curl --max-time 0.2 -X POST "$WINCH_API_URL/api/v1/runs/$RUN_ID/start" \
        -H "Authorization: Bearer $WINCH_TOKEN" -H "X-CSRF-Token: $WINCH_CSRF_TOKEN" \
        -H "Origin: $WINCH_API_URL" -H 'Idempotency-Key: demo-cancel' -H 'If-Match: "1"'
    → expect: curl gives up after 0.2 s
    $ sleep 3 && winch run get "$RUN_ID"
    → expect: `running` or a terminal state — never `queued`

Remove the injection afterwards with `DROP TRIGGER demo_slow_acquire ON runs;
DROP FUNCTION demo_slow_acquire();`.

## Verification

- `make check` and `make test-integration` pass.
- `make e2e` passes, including `test/e2e/start_cancel_test.go`, and the
  standing scenarios pass unchanged.
- Contract suites for fake harness and local sandbox still pass.

## Acceptance criteria

- [ ] A start whose HTTP request is cancelled during lease acquisition leaves the
  run `running` or terminal once acquisition completes, never `queued`, and
  `winch run get` shows it.
- [ ] A start whose lease acquisition returns an error leaves the run in a
  terminal state and launches no harness process.
- [ ] `docs/contracts.md` §1 states what a start does after its caller
  disconnects, and its lifecycle diagram shows any transition this task adds.
- [ ] Each of the three has a test that fails without the change.
- [ ] I1 and I2 still hold.

## Deferrals

| Deferred | Owning task |
|---|---|
| A run left `queued` by a daemon that stopped mid-start | Phase 1 — the *Truthful terminal state for a run the daemon was stopped under* row in [`../phase-1/README.md`](../phase-1/README.md) |

## Traces to

- `docs/contracts.md` §1 (run lifecycle; `start` is conditional)
- `internal/domain/run.go` — the attempt state machine
- [`../post-mortems/2026-09-16-a-plan-defect-read-from-one-implementation.md`](../post-mortems/2026-09-16-a-plan-defect-read-from-one-implementation.md)
  — the audit finding and its reproduction
- P0-008 — the implementation this task revises
