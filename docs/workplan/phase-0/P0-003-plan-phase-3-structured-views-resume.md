# P0-003: Plan phase-3: structured views and resume

**Phase:** 0 — Plan the implementation phases
**Shape:** planning
**Dependencies:** P0-002 (semantic: this artifact must consume the decided first-adapter output and native-exit boundary from phase-2/README.md)

## Objective

Produce a reviewed `docs/workplan/phase-3/README.md` with decided ordered increments for roadmap Stage 2, together with a concrete tracked Phase 0 follow-up assigning derivation of the next uncovered implementation increment to a separate task author.

## Scope

- Partition structured messages/tool calls/file changes from raw output, then terminal/conversation projections over the same immutable runner-local records.
- Select the gap-free ordinal resume boundary, including catch-up, gap repair and slow-reader observations required for its closed loop.
- Preserve a standalone sandbox and shared projection inputs that require no control-plane-only types; distinguish existing retention from the additional seamless reconnect guarantee.

## Required constraints

No structured record/projection/resume implementation exists at the baseline. The phase-2 design supplies planned semantics, not a completed real adapter.

phase-2/README.md's decided adapter output and exit boundary is the actual prerequisite artifact; product implementation may remain pending.

Use `skills/create-phase/SKILL.md` as the phase-author operation, with `skills/shared/phase-design.md` and the workplan model. The required behavior is a decided phase artifact, an independently authored tracked follow-up planning brief, consistent tracker/index/Phase 0 coverage, and an independent `audit-phase` review. Product behavior and commands described there are predictions. Preserve Phase 0's decided boundaries and all task status fields.

## Non-goals

Cross-session history, search, export, and retention policy (roadmap Stage 2 exclusions); control plane (Stage 3).

The assigned phase author does not derive implementation briefs or author the follow-up planning brief. Product implementation and completion claims based only on a design remain excluded. Arranging independent derivation of the tracked follow-up is required for this task to finish; a verbal handoff is insufficient.

## Planning outputs and handoff

The phase author creates `docs/workplan/phase-3/README.md`, records its source-supported current baseline and planned prerequisites, chooses ordered increments and concern owners, states exit observations, and marks implementation task coverage **not yet derived**. Update this phase's row in the index and Phase 0 coverage. Proposed approach: inspect current source and applicable AGENTS first, compare against roadmap Stage 2, then write and review one bounded artifact. No product protocol or substrate implementation is prescribed by this brief.

A separate phase auditor uses `skills/audit-phase/SKILL.md` against the decided design. Resolve blocking artifact findings. Then arrange for a **different task author** to use `skills/derive-task/SKILL.md` to produce at least one concrete Phase 0 planning follow-up brief assigning derivation of the next uncovered implementation increment in phase-3. The phase author cannot author that follow-up. This independent derivation is an expected artifact of P0-003's completion, not merely an invitation to do more work.

Allocate the follow-up's next plan-wide ID only when its complete brief is actually derived; create its real file and matching pending tracker entry, index row and Phase 0 coverage together. The follow-up records a **semantic dependency on P0-003** because it consumes the reviewed `docs/workplan/phase-3/README.md` produced here. Name the target increment until a real implementation brief ID exists; do not allocate a placeholder ID for the implementation task. Recompute the actual partial graph and collision report after each addition, preserving all existing statuses.

The follow-up itself assigns a separate task author to derive one implementation brief at a time against the decided phase artifact, reconcile its tracker/index/coverage records, and arrange the next concrete tracked planning follow-up when actionable coverage remains. Repeat that handoff until the phase's actionable increments are covered. Do not bundle all increments into one derivation or depend on product implementation simply to plan its brief. Dispatch this phase-design task through a separate refiner using `skills/refine-task/SKILL.md` against current HEAD.

## Write set

Forecast:

- `docs/workplan/phase-3/README.md` (new phase artifact).
- `docs/workplan/README.md` (phase 3 design row, actual follow-up dependency row, affected coverage and graph/collision report).
- `docs/workplan/phase-0/README.md` (increment order 3, legacy reference 0.3, design and follow-up coverage/handoff).
- `docs/workplan/phase-0/<allocated-ID>-<target-increment-slug>.md` (new follow-up planning brief, independently authored after the design is decided; concrete ID/title/path recorded then).
- `docs/workplan/tasks.json` (add the actual follow-up entry and its semantic dependency on P0-003; preserve all existing task statuses and metadata).

No existing `tasks.json` status write is assigned; the new independently derived follow-up begins pending with the model's normal unassigned metadata. A design-set correction discovered to be necessary must be declared with its concrete file and contract before proceeding through Phase 0's phase-revision routing; do not silently broaden this forecast.

## Contract surfaces

None changed in product code or wire schemas. This task decides only phase-3's delivery boundaries; it consumes the cited design contracts. Shared tracker additions, index and Phase 0 coverage edits are write collisions, not product contract collisions. A proposed change to shared product meaning requires Phase 0 disposition and a reconciled graph before assignment.

## Demonstration

After creating the design and independently authored follow-up, a reviewer runs from the repository root:

```sh
sed -n '1,240p' docs/workplan/phase-3/README.md
rg -n 'phase-3|P0-003|0\.3' docs/workplan/README.md docs/workplan/phase-0/README.md
bash scripts/list-available-tasks.sh
```

Expected artifact observations: The design gives two views over one record shape and an explicit no-gap reconnect observation, without introducing canonical control-plane sequences. Both index and Phase 0 locate the same design and an actual follow-up brief. Inspect its tracker entry and brief: they agree on its real ID/path and semantic dependency on P0-003, name the next uncovered implementation increment, and preserve pending implementation-brief coverage. Availability reflects the tracker dependencies, not a claimed runtime milestone. Obtain the independent auditor's findings against the reviewed commit; artifact creation alone is not an audit verdict.

## Verification

- Independent `audit-phase` review checks baseline evidence, decided boundaries, task fit, source links and exit observations; reject absent prerequisites presented as existing behavior.
- Validate tracker against its unchanged schema, resolve dependency IDs/brief paths, and compare tracker/index/Phase 0 coverage; `bash scripts/list-available-tasks.sh` must preserve availability semantics. Confirm at least one real follow-up planning brief and tracker entry exist, their semantic dependency consumes the producing phase-design artifact, and no placeholder implementation ID was allocated.
- `git diff --check`, completion-script tests (`python3 -m unittest discover -s test`) and `make check` per AGENTS. Report unavailable toolchains explicitly; do not claim a Go/runtime gate passed when it did not run. No runtime capability is added by this task.

## Acceptance criteria

- [ ] The target artifact is reachable from the index, has the phase-format sections, number-only Order values and descriptive Increment titles, and has no candidate boundary left for the task author to decide.
- [ ] The design gives two views over one record shape and an explicit no-gap reconnect observation, without introducing canonical control-plane sequences.
- [ ] Source evidence and planned prerequisites are labelled separately; injecting a missing prerequisite or a premature completion assertion yields an audit finding, not a pass.
- [ ] Every included concern has one owner, excluded roadmap concerns have their source destination, and a real deferred decision retains its explicit trigger.
- [ ] Independent phase audit has no unresolved blocking findings; index and Phase 0 match the reviewed artifact.
- [ ] At least one complete independently authored Phase 0 follow-up brief exists in the tracker and index, assigns the next uncovered phase-3 implementation increment, records its semantic dependency on P0-003, and has matching Phase 0 coverage. A missing follow-up blocks completion.
- [ ] The follow-up specifies repeated concrete handoffs for remaining actionable coverage; unproduced implementation briefs are labelled not yet derived and have no speculative IDs.
- [ ] Existing runtime/source behavior, frozen tracker schema, task status fields, and separate-author handoff remain intact.

## Deferrals

None. Producing the independently authored tracked follow-up is part of completion. Actual implementation-brief derivation is assigned to that real follow-up, while product implementation remains outside this planning task.

## Traces to

`docs/workplan/phase-0/README.md` ordered planning increment 3 (legacy reference 0.3); docs/roadmap.md §3 Stage 2; docs/contracts.md §6/8; ADR-0002 (Proposed); ADR-0006.

