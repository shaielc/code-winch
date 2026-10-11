# P0-003: Plan phase-3: structured views and resume

**Phase:** 0 — Plan the implementation phases
**Shape:** planning
**Dependencies:** P0-002 (semantic: this artifact must consume the decided first-adapter output and native-exit boundary from phase-2/README.md)

## Objective

Produce a reviewed `docs/workplan/phase-3/README.md` with decided ordered increments for roadmap Stage 2, ready for a separate task author to derive implementation briefs.

## Scope

- Partition structured messages/tool calls/file changes from raw output, then terminal/conversation projections over the same immutable runner-local records.
- Select the gap-free ordinal resume boundary, including catch-up, gap repair and slow-reader observations required for its closed loop.
- Preserve a standalone sandbox and shared projection inputs that require no control-plane-only types; distinguish existing retention from the additional seamless reconnect guarantee.

## Required constraints

No structured record/projection/resume implementation exists at the baseline. The phase-2 design supplies planned semantics, not a completed real adapter.

phase-2/README.md's decided adapter output and exit boundary is the actual prerequisite artifact; product implementation may remain pending.

Use `skills/create-phase/SKILL.md` as the phase-author operation, with `skills/shared/phase-design.md` and the workplan model. The required behavior is a decided phase artifact, consistent index/Phase 0 coverage, and an independent `audit-phase` review. Product behavior and commands described there are predictions. Preserve Phase 0's decided boundaries and all task status fields.

## Non-goals

Cross-session history, search, export, and retention policy (roadmap Stage 2 exclusions); control plane (Stage 3).

Deriving implementation briefs, implementing runtime capability, or claiming a phase completed merely because its design exists. Specific-brief derivation remains uncovered in Phase 0; it is a separate operation, not a completion deferral of this phase-design task.

## Planning outputs and handoff

The phase author creates `docs/workplan/phase-3/README.md`, records its source-supported current baseline and planned prerequisites, chooses ordered increments and concern owners, states exit observations, and marks implementation task coverage **not yet derived**. Update this phase's row in the index and Phase 0 coverage. Proposed approach: inspect current source and applicable AGENTS first, compare against roadmap Stage 2, then write and review one bounded artifact. No product protocol or substrate implementation is prescribed by this brief.

A separate phase auditor uses `skills/audit-phase/SKILL.md` against the resulting commit. Resolve blocking artifact findings before handing the decided increments to a different task author through `skills/derive-task/SKILL.md`, one brief at a time. This phase author must not derive its own implementation briefs. Dispatch this planning task through a separate refiner using `skills/refine-task/SKILL.md` against current HEAD.

## Write set

Forecast:

- `docs/workplan/phase-3/README.md` (new phase artifact).
- `docs/workplan/README.md` (only phase 3 design/index row and affected coverage).
- `docs/workplan/phase-0/README.md` (only increment order 3, legacy reference 0.3, design coverage/handoff).

No `tasks.json` status write is assigned. A design-set correction discovered to be necessary must be declared with its concrete file and contract before proceeding through Phase 0's phase-revision routing; do not silently broaden this forecast.

## Contract surfaces

None changed in product code or wire schemas. This task decides only phase-3's delivery boundaries; it consumes the cited design contracts. Shared index and Phase 0 coverage edits are write collisions, not product contract collisions. A proposed change to shared product meaning requires Phase 0 disposition and a reconciled graph before assignment.

## Demonstration

After creating the artifact, a reviewer runs from the repository root:

```sh
sed -n '1,240p' docs/workplan/phase-3/README.md
rg -n 'phase-3|P0-003|0\.3' docs/workplan/README.md docs/workplan/phase-0/README.md
bash scripts/list-available-tasks.sh
```

Expected artifact observations: The design gives two views over one record shape and an explicit no-gap reconnect observation, without introducing canonical control-plane sequences. Both index and Phase 0 locate the same design and distinguish pending implementation-brief coverage. Availability reflects the tracker dependencies, not a claimed runtime milestone. Obtain the independent auditor's findings against the reviewed commit; artifact creation alone is not an audit verdict.

## Verification

- Independent `audit-phase` review checks baseline evidence, decided boundaries, task fit, source links and exit observations; reject absent prerequisites presented as existing behavior.
- Validate tracker against its unchanged schema, resolve dependency IDs/brief paths, and compare tracker/index/Phase 0 coverage; `bash scripts/list-available-tasks.sh` must preserve availability semantics.
- `git diff --check`, completion-script tests (`python3 -m unittest discover -s test`) and `make check` per AGENTS. Report unavailable toolchains explicitly; do not claim a Go/runtime gate passed when it did not run. No runtime capability is added by this task.

## Acceptance criteria

- [ ] The target artifact is reachable from the index, has the phase-format sections, number-only Order values and descriptive Increment titles, and has no candidate boundary left for the task author to decide.
- [ ] The design gives two views over one record shape and an explicit no-gap reconnect observation, without introducing canonical control-plane sequences.
- [ ] Source evidence and planned prerequisites are labelled separately; injecting a missing prerequisite or a premature completion assertion yields an audit finding, not a pass.
- [ ] Every included concern has one owner, excluded roadmap concerns have their source destination, and a real deferred decision retains its explicit trigger.
- [ ] Independent phase audit has no unresolved blocking findings; index and Phase 0 match the reviewed artifact and show implementation briefs as not yet derived.
- [ ] Existing runtime/source behavior, frozen tracker schema, task status fields, and separate-author handoff remain intact.

## Deferrals

None. Product implementation and specific-brief derivation are non-goals of this artifact-producing task, not unfinished portions of its completion condition.

## Traces to

`docs/workplan/phase-0/README.md` ordered planning increment 3 (legacy reference 0.3); docs/roadmap.md §3 Stage 2; docs/contracts.md §6/8; ADR-0002 (Proposed); ADR-0006.

