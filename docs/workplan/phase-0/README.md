# Phase 0 — Planning and issue-resolution intake

## Outcome and starting point

At exit, a contributor can review a decided design for each implementation
phase in the table below and dispatch separately derived, concrete briefs for
the existing composer issues. Future phase planning, specific-brief planning,
and issue resolution enter the workplan through Phase 0. This phase produces
planning artifacts and handoffs; it implements no product capability and does
not certify a roadmap stage complete.

The previous plan is retired. Its historical P0-001 through P0-007 IDs are not
live owners and must not be reused; the next task author allocates beyond 007.
[The state report](../../state.md) preserves its source evidence and verification
limits. The current root wires only the attach HTTP server
(`cmd/winch-sandbox/main.go:15-36`); the transport serves health, posture and
static assets, rejecting POST (`internal/adapters/transport/attach/server.go:22-40`).
The browser mounts a local composer and transient messages
(`web/src/attach/App.tsx:7-44`); `winch status` is the only CLI command
(`cmd/winch/main.go:8-17`). The shipped fake is not launched by the root.
Harness ownership, live output, durable records and server input remain absent.
Go, Docker and web dependencies are unavailable here; runtime and browser
verification are unperformed, rather than passed or evidence of regression.

The [roadmap](../../roadmap.md) §§2–5 supplies phase coverage and product delivery
order. Workplan numbering deliberately differs: Phase 1 covers unfinished
roadmap Stage 0, and Phase 7 covers Stage 6. The architecture and contracts
describe the destination, not capabilities already available. These planning
increments select target outcomes and boundaries; the target phase author
decides its implementation increments using `create-phase`.

## Concern boundaries

| Concern | Planning owner | Boundary |
|---|---|---|
| Complete the standalone fake sandbox loop | `plan-sandbox` | Plan Phase 1 / Stage 0 from the existing page and local composer; retain separate output/input paths, required run/deploy repairs, raw output, local ordinals, retention and server input. |
| First real harness and native exit interpretation | `plan-first-vendor` | Plan Phase 2 / Stage 1; keep the fake shipped; revisit D1 and D5 only at their roadmap triggers. |
| Structured session records, two views, no-gap ordinal resume | `plan-readable-session` | Plan Phase 3 / Stage 2; no cross-session history, search, export or retention policy. |
| Observe manually started sandboxes | `plan-observer` | Plan Phase 4 / Stage 3; durable observation and product listing, without start/stop, leases or scheduling; preserve direct attach. D2, D4, D6 and D7 remain triggered decisions. |
| Control-plane lifecycle | `plan-run-lifecycle` | Plan Phase 5 / Stage 4; start/stop, idempotency, outbox, fencing, canonical sequence and recovery for loopback single-user operation. |
| Enforced shared-deployment security | `plan-security` | Plan Phase 6 / Stage 5 against security §11, including credentials, egress, traversal, escalation and orphan cleanup; no certification without attack evidence. |
| Second harness and cross-run workflows | `plan-workflows` | Plan Phase 7 / Stage 6; second-provider seam proof followed by durable workflows; no beyond-roadmap scaling topology. |
| Existing composer issue-resolution briefs | `plan-composer-issues` | Review #80, #81 and #82 against source, then derive specific Phase 1 repair briefs under its decided boundaries; no fixes or issue closure in Phase 0. |
| New planning requests and issue discoveries | Phase 0 intake below | Revise or extend Phase 0 before derivation when the concern is not covered; record real planning tasks, never placeholder implementation owners. |

## Ordered increments

Order selects the next uncovered increment for derivation. Rows 1–7 are
independent planning branches: product milestone order is not a completion
dependency between their phase-design tasks. A later phase design must identify
which prerequisites are only predicted and be checked against current source
before its implementation briefs are dispatched. Row 8 consumes the decided
Phase 1 artifact, not Phase 1 implementation.

| Order | Increment | Observable change | Includes | Excludes | Prerequisites |
|---|---|---|---|---|---|
| 1 | `plan-sandbox` | A decided Phase 1 design replaces the absence of an active Stage 0 implementation plan. | Existing-slice reuse, inherited invariant repairs, independently observable input/output loops, owner boundaries and separate task-author handoff. | Implementing harness I/O/store/input; reactivating retired briefs. | Current source, state report, roadmap Stage 0 and §4. |
| 2 | `plan-first-vendor` | A decided Phase 2 design names the first real-harness outcome and fake parity obligations. | Stage 1 scope, native exit reporting, triggered vendor/terminal decisions and predicted Stage 0 prerequisites. | Vendor implementation, second vendor, run stop or security escalation. | Design baseline and roadmap Stage 1; no earlier product implementation required to plan. |
| 3 | `plan-readable-session` | A decided Phase 3 design separates readable projections and resumable session guarantees. | Stage 2 record/view/resume boundaries and predicted earlier prerequisites. | Building projections, canonical run sequence, cross-session history. | Design baseline and roadmap Stage 2. |
| 4 | `plan-observer` | A decided Phase 4 design makes observation of manually started sandboxes the control-plane boundary. | Stage 3 observation, narrow product/attach split, triggered decisions and predicted earlier prerequisites. | Product start/stop, leases, scheduling, multi-user access. | Design baseline and roadmap Stage 3. |
| 5 | `plan-run-lifecycle` | A decided Phase 5 design partitions the single-user create/start/stop loop and its correctness guarantees. | Stage 4 lifecycle, minimal reachable operator paths, durable ordering and predicted earlier prerequisites. | Implementing lifecycle, shared-deployment certification, workflows. | Design baseline and roadmap Stage 4. |
| 6 | `plan-security` | A decided Phase 6 design maps the security launch blockers to observable adversarial increments. | Stage 5 scope, honest posture, attack evidence obligations and predicted earlier prerequisites. | Implementing controls, declaring blockers closed, hostile public tenancy. | Design baseline, security §11 and roadmap Stage 5. |
| 7 | `plan-workflows` | A decided Phase 7 design separates second-provider proof from workflow capabilities. | Stage 6 adapter parity, durable workflow boundaries and predicted earlier prerequisites. | Implementing providers/workflows, scale beyond the roadmap. | Design baseline and roadmap Stage 6. |
| 8 | `plan-composer-issues` | Specific repair briefs trace #80/#81/#82 to a decided Phase 1 increment and current evidence. | Exact issue/source review, separate task-author derivation, dependencies, coverage, collision report and handoff. | Fixing behavior/tests, closing or commenting on issues, completion claims. | The decided Phase 1 artifact from `plan-sandbox`; its implementation is not required. |

Issue #80 concerns textarea focus after rejected button submission, #81 concerns
Composer keyboard/validation evidence, and #82 concerns the local-message
no-ordinal assertion. These summaries are intake pointers, not final acceptance
criteria. Their planning author must read exact issue evidence and current code,
including overlap between #80 and #81, before deriving independently completable
repair briefs. If the Phase 1 boundary cannot own a finding, return it to the
phase author for revision rather than inventing coverage.

## Task coverage

| Increment | Derived tasks |
|---|---|
| `plan-sandbox` | [P0-008](P0-008-plan-standalone-fake-sandbox.md) |
| `plan-first-vendor` | [P0-009](P0-009-plan-first-real-harness.md) |
| `plan-readable-session` | [P0-010](P0-010-plan-readable-rejoinable-session.md) |
| `plan-observer` | [P0-011](P0-011-plan-manual-sandbox-observer.md) |
| `plan-run-lifecycle` | [P0-012](P0-012-plan-control-plane-run-lifecycle.md) |
| `plan-security` | [P0-013](P0-013-plan-enforced-shared-deployment-security.md) |
| `plan-workflows` | [P0-014](P0-014-plan-second-provider-and-workflows.md) |
| `plan-composer-issues` | [P0-015](P0-015-plan-composer-refusal-focus.md) (#80), [P0-016](P0-016-plan-composer-validation-keyboard-evidence.md) (#81), [P0-017](P0-017-plan-local-message-no-ordinal-oracle.md) (#82) |

All eight increments are covered by ten concrete planning briefs. Next handoff:
dispatch P0-008–P0-014 to separate phase authors using `create-phase`; the
three issue-planning tasks consume only P0-008's decided Phase 1 artifact and
then use separate `derive-task` authors. Derivation is complete; planning-task
execution, audits and product implementation have not occurred.

There are no active legacy briefs to rederive. Reusable code is baseline
evidence, not reused historical completion or assignment. The #81 button-refusal
item duplicates #80's requested scope; its three remaining regression guarantees
form an independent target brief. This intake attribution is not a product
deferral owner, and full-issue closure must assess all exact issue evidence.

## Future planning and issue-resolution intake

1. Start here for every new phase-design request, specific task brief, issue or
   accepted follow-up. Record its source and compare it with the roadmap, state,
   current source, active phase boundaries and existing tracker owners.
2. The Phase 0 author uses `extend-phase` for a new planning concern or
   `revise-phase` for a changed boundary; keep this table and the index aligned.
   An already-covered concern goes to its uncovered increment without a duplicate.
3. A different task author uses `derive-task` to register a real Phase 0 planning
   brief with the next append-only ID. Phase-design briefs assign `create-phase`;
   specific-brief and issue-resolution briefs assign `derive-task` and depend on
   the target's decided phase artifact. Planning IDs and issues do not substitute
   for an existing implementation task owner when a brief defers part of its
   implementation completion. Accepted audit follow-ups may route to issues
   under `propagate-task`; planning their resolution still enters here. No
   hypothetical implementation task supplies an owner.
4. The planning task produces the named phase artifact or specific brief(s),
   updates index/coverage/tracker consistently, and hands off to a separate
   author. Product fixes belong to Phase 1 or later. Issue evidence may show
   duplicates, no remaining defect, or a boundary mismatch; record the evidence
   and route the decision here rather than silently fixing or closing it.
5. Dispatch implementation only from a specific brief under a decided target
   phase, after refinement against current source. Planning completion proves
   the artifact ready for handoff; issue closure and product completion require
   their own verified implementation and authorized workflow.

## Phase exit

All seven target phase designs have explicit starting points, observable exits,
ordered increments, boundaries and prerequisite distinctions, reviewed using
`audit-phase`. The composer issue intake has evidence-linked specific repair
briefs with complete graph/coverage records and an implementation handoff.
The index and tracker match Phase 0 coverage, use no retired IDs, and leave the
frozen schema unchanged. Reviewers can follow the intake route for new work.
No target phase, repair or shared-deployment guarantee is declared implemented
by these observations.
