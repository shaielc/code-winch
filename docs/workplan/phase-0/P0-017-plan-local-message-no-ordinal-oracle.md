# P0-017: Plan local-message no-ordinal oracle repair

**Phase:** 0 — Planning and issue-resolution intake
**Shape:** planning
**Increment:** `plan-composer-issues` — #82
**Dependencies:** P0-008 (contract: its decided Phase 1 increment boundaries govern the specific repair brief).

## Objective

A contributor can dispatch an evidence-linked Phase 1 brief strengthening the observable local-only message oracle from issue #82.

## Scope

- At dispatch, read exact current #82 body/state and cited PR #78 audit evidence, current App/Composer source and tests, the decided Phase 1 artifact from P0-008, tracker and relevant real briefs. Record issue URL, revision/time and source observations rather than copying the intake wording into acceptance criteria.
- Assign a different derive-task author to derive one concrete Phase 1 test-oracle repair brief under the local-input increment. Current source renders a Local label and exact payload in each list item; it has no runner-backed ordinal structure. Resolve the issue's requested oracle against that observable local-only DOM contract without inventing an ordinal UI.
- Require an oracle that rejects accidentally exposed numeric-only ordinal metadata while preserving numeric user content and embedded whitespace/newlines. State credible negative/mutation evidence; do not merely search for the literal word ordinal or ban all digits. If the decided Phase 1 boundary cannot own this oracle, return the finding to the phase author before allocating an incompatible brief.

## Non-goals

No test or UI fixes, invented runner ordinal DOM/API, server-input implementation, issue comment/closure or completion claim. No component internals/React keys as a substitute for observable message semantics.

## Planning outputs and handoff

Use [`derive-task`](../../../skills/derive-task/SKILL.md) with a task author
separate from the Phase 1 author. Allocate the next real plan-wide ID only when
deriving the concrete target brief; create its complete file/tracker/index/Phase 1
coverage records together. No target file or tracker placeholder exists merely
because this planning brief names its output.

The output is a specific Phase 1 brief (forecast filename suffix
`-local-message-no-ordinal-oracle.md`) plus an evidence/duplicate disposition in this
planning brief. Hand the resulting target to a separate
[`refine-task`](../../../skills/refine-task/SKILL.md) author when dispatched
against current source. Any newly discovered uncovered boundary returns through
Phase 0 intake to the phase author; this author cannot silently revise it.

## Write set

- `docs/workplan/phase-1/<allocated-real-ID>-local-message-no-ordinal-oracle.md` — one new complete brief; resolve the ID from the tracker at dispatch, never preallocate a placeholder.
- `docs/workplan/tasks.json` — the real derived target entry/dependencies; approval automation alone stamps this planning task's status.
- `docs/workplan/phase-1/README.md` — derived target coverage only, preserving decided boundaries.
- `docs/workplan/phase-0/P0-017-plan-local-message-no-ordinal-oracle.md` — exact issue/source evidence and intake disposition.
- `docs/workplan/phase-0/README.md` and `docs/workplan/README.md` — matching handoff/coverage/graph notes.

## Contract surfaces

- Consumes planning artifact namespace `phase-1` decided by P0-008.
- Defines issue-planning namespace `composer-issue-82`: target brief scope/evidence attribution and handoff.
- Changes no product contract in Phase 0. Other issue-planning tasks consume the same read-only Phase 1 design; their own artifact meanings are distinct. New target implementation contract declarations and any real overlap are derived at dispatch.

## Dependency and concurrency analysis

P0-008 defines the phase boundary this task consumes, so the contract edge is
required. No Phase 1 product implementation or other issue-planning output is a
prerequisite. Shared index/tracker/coverage files create write warnings, never
extra edges. The [index](../README.md#phase-0-graph-and-handoff) reports the
derived graph's critical path, average width, available pairs and contract
collisions; #82's scope does not consume another issue's not-yet-existing
implementation owner.

The current graph has 10 tasks, critical-path length 2 and average width 5.0;
7 tasks are available now (maximum possible frontier 9). Current available
write collisions are the 21 unordered pairs among P0-008–P0-014; their
shared index/coverage/status writes are warnings, with no available contract
collision. The index explains the distinct target artifact namespaces.

## Demonstration

Read-only intake snapshot on 2026-10-11: [#82](https://github.com/shaielc/code-winch/issues/82)
is open, originates in PR #73 and cites audit PR #78. App.tsx:26–37 renders local list items as Local plus payload, with no ordinal UI. App.test.tsx:19–32 only asserts absence of /ordinal/i and would miss numeric metadata.
Re-read exact evidence at dispatch; this snapshot is neither a runtime
observation nor final target acceptance.

Run after deriving the output from the repository root:

```sh
git diff -- docs/workplan/phase-1 docs/workplan/tasks.json docs/workplan/README.md
bash scripts/list-available-tasks.sh
```

Open the new brief named by its real `tasks.json` entry and its Phase 1 coverage
row. A reviewer can identify the proposed observable local-message contract, why the current literal-word check is weak, and how a numeric ordinal mutation fails while a legitimate numeric payload passes. No issue mutation or repaired product behavior is demonstrated
by this review.

## Verification

- Review exact issue/audit/source evidence and the decided Phase 1 artifact; inspect scope closure, claimed observations and new target contract/write declarations.
- Verify the new real target's tracker entry, dependency reasons, index row and phase coverage agree, with no retired ID, ownerless deferral, cycle or concurrent contract conflict.
- Apply the [shared planning checks](../README.md#planning-author-checks); all runtime commands in the target remain predictions until that target's implementation verifies them.

## Acceptance criteria

- [ ] The target brief records that no runner-backed ordinal structure exists at current source, and chooses an observable oracle consistent with the decided Phase 1 local-only contract.
- [ ] Its acceptance and demonstration distinguish accidental numeric metadata from user text containing digits; strengthening the oracle requires no new runner-backed UI or implementation-only selector.
- [ ] Any boundary mismatch is returned to the phase author; the exact issue request is reconciled explicitly rather than claiming a nonexistent ordinal structure was tested.
- [ ] The concrete target brief is independently dispatchable under a decided Phase 1 increment, has valid complete graph/coverage records, and goes to a separate refiner.
- [ ] Exact current evidence and issue disposition are reviewable; source/tests are unchanged by Phase 0 and no issue is fixed, commented on or closed.

## Deferrals

None. This task produces a brief and evidence disposition; implementation and
full-issue closure are explicit non-goals, not completion conditions delegated
to hypothetical owners.

## Traces to

`docs/workplan/phase-0/README.md` — `plan-composer-issues`;
[#82](https://github.com/shaielc/code-winch/issues/82);
[PR #78](https://github.com/shaielc/code-winch/pull/78);
`docs/roadmap.md` §3 Stage 0, §4; `docs/contracts.md` §8–9;
`docs/state.md` — Browser-local composition and verification limits;
`skills/shared/workplan-model.md` — planning shape, completion closure and dependencies.
