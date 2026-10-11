# P0-016: Plan composer validation and keyboard evidence

**Phase:** 0 — Planning and issue-resolution intake
**Shape:** planning
**Increment:** `plan-composer-issues` — #81
**Dependencies:** P0-008 (contract: its decided Phase 1 increment boundaries govern the specific repair brief).

## Objective

A contributor can dispatch an evidence-linked Phase 1 brief for three nonduplicate Composer/App regression guarantees from issue #81.

## Scope

- At dispatch, read the exact current #81 body/state and cited PR #78 audit evidence, current Composer/App source and tests, the decided Phase 1 artifact from P0-008, tracker and real overlapping briefs. Record issue URL, revision/time and current source evidence.
- Assign a different derive-task author to derive one independently completable Phase 1 regression brief for Enter preventDefault, successful submission clearing a prior validation error, and App rendering preserving leading/trailing whitespace plus embedded newlines. Protect existing product behavior rather than changing its contract or adding features for test convenience.
- Record #81's fourth item, whitespace-button refusal, as duplicate intake evidence within #80's focus-refusal repair scope. The three-guarantee target brief excludes that duplicate and needs no future #80 implementation ID to finish; it contains no deferral to P0-015. Keep full-issue closure assessment separate from completion of this narrower regression brief.

## Non-goals

No product or test fixes, button-focus repair, issue comment/closure, or certification of all #81 items. Target scope is the three nonduplicate regressions; Enter newline browser behavior cannot be inferred from manually injecting a newline.

## Planning outputs and handoff

Use [`derive-task`](../../../skills/derive-task/SKILL.md) with a task author
separate from the Phase 1 author. Allocate the next real plan-wide ID only when
deriving the concrete target brief; create its complete file/tracker/index/Phase 1
coverage records together. No target file or tracker placeholder exists merely
because this planning brief names its output.

The output is a specific Phase 1 brief (forecast filename suffix
`-composer-validation-keyboard-evidence.md`) plus an evidence/duplicate disposition in this
planning brief. Hand the resulting target to a separate
[`refine-task`](../../../skills/refine-task/SKILL.md) author when dispatched
against current source. Any newly discovered uncovered boundary returns through
Phase 0 intake to the phase author; this author cannot silently revise it.

## Write set

- `docs/workplan/phase-1/<allocated-real-ID>-composer-validation-keyboard-evidence.md` — one new complete brief; resolve the ID from the tracker at dispatch, never preallocate a placeholder.
- `docs/workplan/tasks.json` — the real derived target entry/dependencies; approval automation alone stamps this planning task's status.
- `docs/workplan/phase-1/README.md` — derived target coverage only, preserving decided boundaries.
- `docs/workplan/phase-0/P0-016-plan-composer-validation-keyboard-evidence.md` — exact issue/source evidence and intake disposition.
- `docs/workplan/phase-0/README.md` and `docs/workplan/README.md` — matching handoff/coverage/graph notes.

## Contract surfaces

- Consumes planning artifact namespace `phase-1` decided by P0-008.
- Defines issue-planning namespace `composer-issue-81`: target brief scope/evidence attribution and handoff.
- Changes no product contract in Phase 0. Other issue-planning tasks consume the same read-only Phase 1 design; their own artifact meanings are distinct. New target implementation contract declarations and any real overlap are derived at dispatch.

## Dependency and concurrency analysis

P0-008 defines the phase boundary this task consumes, so the contract edge is
required. No Phase 1 product implementation or other issue-planning output is a
prerequisite. Shared index/tracker/coverage files create write warnings, never
extra edges. The [index](../README.md#phase-0-graph-and-handoff) reports the
derived graph's critical path, average width, available pairs and contract
collisions; #81's scope does not consume another issue's not-yet-existing
implementation owner.

The current graph has 10 tasks, critical-path length 2 and average width 5.0;
7 tasks are available now (maximum possible frontier 9). Current available
write collisions are the 21 unordered pairs among P0-008–P0-014; their
shared index/coverage/status writes are warnings, with no available contract
collision. The index explains the distinct target artifact namespaces.

## Demonstration

Read-only intake snapshot on 2026-10-11: [#81](https://github.com/shaielc/code-winch/issues/81)
is open, originates in PR #73 and cites audit PR #78. Composer.tsx:36–40 prevents
default; :30–33 clears validation on successful submit. Composer.test.tsx tests
submission and edit-time alert clearing but lacks direct default-prevention and
success-after-error assertions; App.test.tsx uses a simple payload.
Re-read exact evidence at dispatch; this snapshot is neither a runtime
observation nor final target acceptance.

Run after deriving the output from the repository root:

```sh
git diff -- docs/workplan/phase-1 docs/workplan/tasks.json docs/workplan/README.md
bash scripts/list-available-tasks.sh
```

Open the new brief named by its real `tasks.json` entry and its Phase 1 coverage
row. A reviewer sees exactly three target regression guarantees, current evidence for each, and a separate intake note for the button-refusal duplicate; no not-yet-derived owner is required to dispatch that brief. No issue mutation or repaired product behavior is demonstrated
by this review.

## Verification

- Review exact issue/audit/source evidence and the decided Phase 1 artifact; inspect scope closure, claimed observations and new target contract/write declarations.
- Verify the new real target's tracker entry, dependency reasons, index row and phase coverage agree, with no retired ID, ownerless deferral, cycle or concurrent contract conflict.
- Apply the [shared planning checks](../README.md#planning-author-checks); all runtime commands in the target remain predictions until that target's implementation verifies them.

## Acceptance criteria

- [ ] The target brief covers precisely default prevention, success-after-error clearing and App exact whitespace/newline rendering, with observable assertions that would fail when each guarantee breaks.
- [ ] Duplicate button-refusal evidence is explicitly accounted for at intake and excluded from this target brief, without a speculative owner or a completion deferral.
- [ ] The artifact states that closing #81 requires separate assessment of its full evidence, including duplicate button refusal; this planning completion and the three-guarantee target task alone cannot certify it closed.
- [ ] The concrete target brief is independently dispatchable under a decided Phase 1 increment, has valid complete graph/coverage records, and goes to a separate refiner.
- [ ] Exact current evidence and issue disposition are reviewable; source/tests are unchanged by Phase 0 and no issue is fixed, commented on or closed.

## Deferrals

None. This task produces a brief and evidence disposition; implementation and
full-issue closure are explicit non-goals, not completion conditions delegated
to hypothetical owners.

## Traces to

`docs/workplan/phase-0/README.md` — `plan-composer-issues`;
[#81](https://github.com/shaielc/code-winch/issues/81);
[PR #78](https://github.com/shaielc/code-winch/pull/78);
`docs/roadmap.md` §3 Stage 0, §4; `docs/contracts.md` §8–9;
`docs/state.md` — Browser-local composition and verification limits;
`skills/shared/workplan-model.md` — planning shape, completion closure and dependencies.
