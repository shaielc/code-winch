# P2-057: Retry a failed run

**Phase:** 2 — Structured experience and second harness
**Shape:** capability
**Dependencies:** P1-050 (contract: retry is a new run operation on the existing run API and use cases)

## Objective

A user retries a failed run and gets a new attempt linked to the original,
making the `Failed → Queued` transition the domain has always encoded reachable
for the first time.

## Scope

- A `RetryRun` use case creating a linked attempt: new attempt ID, incremented
  attempt number, same run ID, same resolved profile unless explicitly
  overridden, and a lifecycle event citing the attempt it follows.
- History is never rewritten. The failed attempt's events keep their sequence
  numbers; the new attempt appends.
- Retry is refused for a run in a non-retryable terminal state, and is
  idempotent under a repeated idempotency key.
- API, CLI, and a run-page control offering retry only when the state permits it.
- Attempt lineage is visible: the run detail shows every attempt with its
  outcome, and event history can be filtered to one attempt.

## Non-goals

- Automatic retry. This is a user-initiated command; workflow-level retry
  policy is P4-037's concern.
- Resuming a harness session. A retry is a new execution, not a continuation.

## Runtime reachability

`POST /api/v1/runs/{id}/retry`; `winch run retry`; the retry control on the run
page.

## Owned surfaces

`internal/application/run_retry.go`,
`api/openapi/paths/run-retry.yaml`,
`cmd/winch/run_retry.go`, `web/src/features/runs/AttemptList.tsx`.

## Demonstration

    $ ID=$(winch run create --harness fake --scenario exits-nonzero --json | jq -r .id)
    $ winch run start $ID && winch run get $ID --json | jq -r .state
    → expect: failed

    $ winch run retry $ID && winch run get $ID --json | jq '.attempts | length'
    → expect: 2, with the second queued and linked to the first

    $ winch run watch $ID --attempt 1 | wc -l
    → expect: the original attempt's history, unchanged and still ordered

    $ winch run retry $ID   # while the retry is running
    → expect: refused with a stable state-conflict code

## Verification

- Standing scenario suite gains one scenario: fail → retry → succeed.
- State machine tests for retry from every terminal and non-terminal state.
- Idempotency test for a repeated retry key.
- A test asserting the first attempt's events are byte-identical after retry.

## Acceptance criteria

- [ ] `Failed → Queued` is reachable from the API, the CLI, and the browser.
- [ ] A retry creates a linked attempt and rewrites no history.
- [ ] Retry from a non-retryable state is refused with a stable code.
- [ ] Attempt lineage is visible and filterable in the UI.

## Deferrals

| Deferred | Owning task |
|---|---|
| Workflow-driven retry policy with backoff | P4-037 |
| Admission control applied to retried runs | P2-059 |

## Traces to

`docs/contracts.md` §1 (`Failed → Queued`); `docs/architecture.md` §6;
`docs/roadmap.md` Phase 2
