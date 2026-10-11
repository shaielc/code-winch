# P0-015: Plan composer rejected-submission focus repair

**Phase:** 0 — Planning and issue-resolution intake
**Shape:** planning
**Increment:** `plan-composer-issues` — #80
**Dependencies:** P0-008 (contract: its decided Phase 1 increment boundaries govern the specific repair brief).

## Objective

A contributor can dispatch an evidence-linked Phase 1 repair brief for textarea focus after rejected Enter or button submission.

## Scope

- At dispatch, read the exact current #80 body/state and its cited PR #78 audit evidence, current Composer/App source and tests, the decided Phase 1 artifact from P0-008, tracker and any already derived overlapping brief. Record issue URL, revision/time and source references; today's intake summary is not a substitute.
- Assign a different task author to derive-task, one concrete brief at a time, for the Phase 1 increment owning the inherited local composer. The target brief covers textarea focus after whitespace refusal through Enter and the button, draft/alert preservation and the button-refusal regression; do not revive the retired P0-004 assignment.
- Inspect the actual event/focus path and credible browser/component observation. A button click must begin with button focus so a synthetic event that leaves the textarea focused cannot falsely prove correction continuity. Reconcile any changed or duplicate issue evidence with real existing tracker owners before deriving.

## Non-goals

No focus fix, component-test change, issue comment/closure or completion claim. Recorded input, harness delivery and unrelated composer improvements are outside this repair-brief scope.

## Planning outputs and handoff

Use [`derive-task`](../../../skills/derive-task/SKILL.md) with a task author
separate from the Phase 1 author. Allocate the next real plan-wide ID only when
deriving the concrete target brief; create its complete file/tracker/index/Phase 1
coverage records together. No target file or tracker placeholder exists merely
because this planning brief names its output.

The output is a specific Phase 1 brief (forecast filename suffix
`-composer-rejected-submission-focus.md`) plus an evidence/duplicate disposition in this
planning brief. Hand the resulting target to a separate
[`refine-task`](../../../skills/refine-task/SKILL.md) author when dispatched
against current source. Any newly discovered uncovered boundary returns through
Phase 0 intake to the phase author; this author cannot silently revise it.

## Write set

- `docs/workplan/phase-1/<allocated-real-ID>-composer-rejected-submission-focus.md` — one new complete brief; resolve the ID from the tracker at dispatch, never preallocate a placeholder.
- `docs/workplan/tasks.json` — the real derived target entry/dependencies; approval automation alone stamps this planning task's status.
- `docs/workplan/phase-1/README.md` — derived target coverage only, preserving decided boundaries.
- `docs/workplan/phase-0/P0-015-plan-composer-refusal-focus.md` — exact issue/source evidence and intake disposition.
- `docs/workplan/phase-0/README.md` and `docs/workplan/README.md` — matching handoff/coverage/graph notes.

## Contract surfaces

- Consumes planning artifact namespace `phase-1` decided by P0-008.
- Defines issue-planning namespace `composer-issue-80`: target brief scope/evidence attribution and handoff.
- Changes no product contract in Phase 0. Other issue-planning tasks consume the same read-only Phase 1 design; their own artifact meanings are distinct. New target implementation contract declarations and any real overlap are derived at dispatch.

## Dependency and concurrency analysis

P0-008 defines the phase boundary this task consumes, so the contract edge is
required. No Phase 1 product implementation or other issue-planning output is a
prerequisite. Shared index/tracker/coverage files create write warnings, never
extra edges. The [index](../README.md#phase-0-graph-and-handoff) reports the
derived graph's critical path, average width, available pairs and contract
collisions; #80's scope does not consume another issue's not-yet-existing
implementation owner.

The current graph has 10 tasks, critical-path length 2 and average width 5.0;
7 tasks are available now (maximum possible frontier 9). Current available
write collisions are the 21 unordered pairs among P0-008–P0-014; their
shared index/coverage/status writes are warnings, with no available contract
collision. The index explains the distinct target artifact namespaces.

## Demonstration

Read-only intake snapshot on 2026-10-11: [#80](https://github.com/shaielc/code-winch/issues/80)
is open, originates in PR #73 and cites audit PR #78. web/src/attach/Composer.tsx:21–30 returns after setting the validation error without refocusing; Composer.test.tsx:46–62 covers Enter refusal with textarea already focused, not button refusal.
Re-read exact evidence at dispatch; this snapshot is neither a runtime
observation nor final target acceptance.

Run after deriving the output from the repository root:

```sh
git diff -- docs/workplan/phase-1 docs/workplan/tasks.json docs/workplan/README.md
bash scripts/list-available-tasks.sh
```

Open the new brief named by its real `tasks.json` entry and its Phase 1 coverage
row. A reviewer traces #80 to the decided Phase 1 local-input increment and the new brief's button/Enter focus evidence, then sees a real pending implementation owner and a separate-refiner handoff. No issue mutation or repaired product behavior is demonstrated
by this review.

## Verification

- Review exact issue/audit/source evidence and the decided Phase 1 artifact; inspect scope closure, claimed observations and new target contract/write declarations.
- Verify the new real target's tracker entry, dependency reasons, index row and phase coverage agree, with no retired ID, ownerless deferral, cycle or concurrent contract conflict.
- Apply the [shared planning checks](../README.md#planning-author-checks); all runtime commands in the target remain predictions until that target's implementation verifies them.

## Acceptance criteria

- [ ] The derived brief has one independently completable focus-repair outcome with Enter/button refusal, preserved draft/alert, an honest observation path and focused regression evidence.
- [ ] Source/issue evidence explains the missing button path without claiming a synthetic unfocused click or unrun browser test demonstrated it.
- [ ] Any #81 button-refusal overlap is recorded as intake attribution to this requested scope, with no issue closure implied; any actual contract consumption found in real target briefs gains a justified edge.
- [ ] The concrete target brief is independently dispatchable under a decided Phase 1 increment, has valid complete graph/coverage records, and goes to a separate refiner.
- [ ] Exact current evidence and issue disposition are reviewable; source/tests are unchanged by Phase 0 and no issue is fixed, commented on or closed.

## Deferrals

None. This task produces a brief and evidence disposition; implementation and
full-issue closure are explicit non-goals, not completion conditions delegated
to hypothetical owners.

## Traces to

`docs/workplan/phase-0/README.md` — `plan-composer-issues`;
[#80](https://github.com/shaielc/code-winch/issues/80);
[PR #78](https://github.com/shaielc/code-winch/pull/78);
`docs/roadmap.md` §3 Stage 0, §4; `docs/contracts.md` §8–9;
`docs/state.md` — Browser-local composition and verification limits;
`skills/shared/workplan-model.md` — planning shape, completion closure and dependencies.
