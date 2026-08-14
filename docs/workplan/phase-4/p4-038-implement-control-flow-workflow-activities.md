# P4-038: Implement control-flow workflow activities

**Phase:** 4 — Top-level workflows
**Shape:** capability
**Dependencies:** P4-036 (compile: the activity registry), P2-024 (semantic: `approval.wait` needs bound, expiring approvals), P2-023 (semantic: `artifact.publish` needs an artifact store)

## Objective

A workflow branches on a condition, fans out with bounded concurrency, waits for
a human approval, and publishes an artifact.

## Scope

- `condition` with typed predicates over prior step outputs, no embedded code.
- `parallel` and `foreach` with bounded concurrency declared in the definition
  and enforced by the coordinator, not by the host's goroutine count.
- `approval.wait`: binds to an approval record from P2-024, honours its expiry,
  and resolves through the same input path a user uses.
- `artifact.publish` writing through the artifact store with its sensitivity
  class, so retention and export treat it identically to a run's artifacts.
- Compensation where a definition declares it, invoked on branch failure.
- A cancelled or failed branch leaves siblings in a defined state rather than
  orphaned.

## Non-goals

- Arbitrary user code in a definition. Custom behavior uses a registered,
  policy-controlled activity per `docs/contracts.md` §7.
- Unbounded fan-out.

## Runtime reachability

`winch workflow start` with definitions exercising each step type; approvals
answered through `winch run approve` or the browser.

## Owned surfaces

`internal/workflow/activities/control.go`,
`internal/workflow/activities/approval.go`,
`internal/workflow/activities/artifact.go`,
`test/scenarios/workflow-control-*.json`.

## Demonstration

    $ winch workflow start --definition test/scenarios/workflow-foreach.json
    $ winch workflow get $WID --json | jq '[.steps[] | select(.state=="running")] | length'
    → expect: never more than the declared concurrency, sampled repeatedly

    $ winch workflow start --definition test/scenarios/workflow-approval.json
    $ winch workflow get $WID --json | jq -r '.steps[] | select(.type=="approval.wait") | .state'
    → expect: waiting, with an approval ID
    $ winch run approve $RID --approval $A --decision allow
    → expect: the step completes and the instance advances

    # the same instance, left unanswered past the approval's expiry:
    → expect: the step fails on expiry, and compensation runs if declared

    $ winch workflow get $WID --json | jq -r '.steps[] | select(.type=="artifact.publish") | .output.artifactId'
    $ winch run artifacts --artifact $AID
    → expect: the artifact exists with its declared sensitivity class

## Verification

- Standing scenario suite passes with a control-flow workflow in the path.
- Concurrency bound test under a fast-completing `foreach`.
- Approval expiry and double-resolution tests with a fake clock.
- Compensation test on branch failure, including compensation itself failing.

## Acceptance criteria

- [ ] Declared concurrency is never exceeded.
- [ ] `approval.wait` uses the same approval records and input path a user does.
- [ ] Published artifacts carry a sensitivity class and are subject to retention.
- [ ] A failed branch leaves siblings in a defined, inspectable state.
- [ ] No definition can execute arbitrary code.

## Deferrals

| Deferred | Owning task |
|---|---|
| Exposing these steps over HTTP and in the UI | P4-039, P4-040 |

## Traces to

`docs/contracts.md` §7; `docs/security.md` §10; `docs/roadmap.md` Phase 4
