# P2-056: Action approvals and structured answers

**Phase:** 2 — Structured experience and second harness
**Shape:** capability
**Dependencies:** P2-024 (semantic: there is no bound approval to answer until binding and expiry exist), P2-022 (contract: the approval card is the UI surface the control belongs on)

## Objective

A user answers an approval and a structured question from the browser or the
CLI, and the run continues — closing the round trip that Phase 2 otherwise
displays but cannot complete.

## Scope

- The two input payload kinds `docs/contracts.md` §3 lists that P1-016
  deliberately left out: `approval` and `structured answer`. Both go through the
  same idempotent input command path as `text`, with the same acceptance,
  outbox, and delivery guarantees.
- An approval input cites the approval ID and the operation digest; a mismatch,
  a replay, or an expired approval is refused with a stable code.
- Adapter encoding: each harness codec encodes both kinds into its native form,
  and the shared harness contract suite checks that a driver declaring the
  capability actually round-trips them.
- API and CLI: answer an approval, submit a structured answer, both authorized
  as run control rather than run read.
- UI: the approval card gains approve and deny controls with the scope shown,
  and a structured question renders its choices as a form. A card whose approval
  has expired is visibly dead, not merely unresponsive.

## Non-goals

- Policy evaluation and auto-resolution — P2-024 owns those.
- Approval steps inside a workflow — P4-038.
- Free-form terminal input as an approval substitute; raw terminal input keeps
  its separate authorization.

## Runtime reachability

`POST /api/v1/runs/{id}/input` with `kind=approval` or `kind=structured_answer`;
`winch run approve|answer`; the approval card in the run page.

## Owned surfaces

`internal/application/input.go` (two new kinds),
`api/openapi/components/run-input.yaml`,
`cmd/winch/approve.go`, `web/src/features/runs/ApprovalCard.tsx`,
`internal/adapters/harness/*/` (codec encode paths).

## Demonstration

    $ ID=$(winch run create --harness fake --scenario asks-approval --json | jq -r .id)
    $ winch run start $ID
    $ A=$(winch run approvals $ID --json | jq -r '.approvals[0].id')
    $ winch run approve $ID --approval $A --decision allow
    → expect: 202, an approval resolution event, and the harness continuing

    $ winch run approve $ID --approval $A --decision allow
    → expect: refused as already resolved, with the same stable code as a replay

    # in the browser, with a second scenario asking a structured question:
    → expect: a form with the declared choices; submitting it produces the same
      event a CLI answer would

    # after expiry:
    → expect: the card is shown as expired and its controls are disabled

## Verification

- Standing scenario suite gains one scenario: request → answer → run continues.
- Idempotency tests for both kinds, including repeated key with changed body.
- Harness contract suite covers approval and structured-answer encoding for
  every adapter declaring support.
- UI tests for the expired, denied, and unsupported-capability states.

## Acceptance criteria

- [ ] Every input payload kind in `docs/contracts.md` §3 is submittable through
      the API, the CLI, and the browser.
- [ ] An expired or replayed approval is refused with a stable code.
- [ ] An adapter that does not support approvals disables the control with a
      stated reason rather than failing at delivery time.
- [ ] Answering is authorized as run control and audited.

## Deferrals

| Deferred | Owning task |
|---|---|
| Waiting on an approval from a workflow step | P4-038 |

## Traces to

`docs/contracts.md` §3; `docs/security.md` §10; `docs/architecture.md` §4;
`docs/roadmap.md` Phase 2
