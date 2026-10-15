# P0-004: Plan phase-4: observe existing sandboxes

**Phase:** 0 — Plan the implementation phases
**Shape:** planning
**Dependencies:** P0-003 (semantic: this artifact must preserve the decided ordinal/projection boundary in phase-3/README.md)

## Objective

Produce a reviewed `docs/workplan/phase-4/README.md` with decided ordered increments for roadmap Stage 3, and a single planning-shape task in Phase 4 that derives its first increment, created with `skills/phase-increment-handoff/SKILL.md`.

## Scope

- Partition discovery, registration/routing, durable observation of already-running sessions, and read-only multi-sandbox product visibility.
- Resolve D6 at Stage 3 opening, D4 as registration/first protocol command is defined, and D2 only at its first shared-component trigger.
- Distinguish a durable record of an existing execution from authority to launch it; preserve direct attach and independently startable sandbox behavior.
- Decide D7 if this phase first persists sandbox-produced records; otherwise name its exact roadmap trigger without inventing canonical event persistence.

## Required constraints

No winchd/product multi-sandbox surface or remote runner protocol exists at the baseline; the contributor control panel is unrelated product functionality.

phase-3/README.md's local record/projection/resume contract is preserved by the planned routing and control-plane observation artifact.

Use `skills/create-phase/SKILL.md` as the phase-author operation, with `skills/shared/phase-design.md` and the workplan model. The required behavior is a decided phase artifact, an independent `audit-phase` review, the single planning-shape task from `skills/phase-increment-handoff/SKILL.md`, and consistent tracker/index/Phase 0 and phase coverage. Product behavior and commands described there are predictions. Preserve Phase 0's decided boundaries and all task status fields.

## Non-goals

Starting/stopping sandboxes, leases, scheduling and multi-user access (Stage 3 exclusions; lifecycle Stage 4).

This task derives no implementation brief and allocates no implementation ID; the planning task it creates assigns that derivation. Product implementation and completion claims based only on a design remain excluded.

## Planning outputs and handoff

The phase author creates `docs/workplan/phase-4/README.md`, records its source-supported current baseline and planned prerequisites, chooses ordered increments and concern owners, states exit observations, and marks implementation task coverage **not yet derived**. Update this phase's row in the index and Phase 0 coverage. Proposed approach: inspect current source and applicable AGENTS first, compare against roadmap Stage 3, then write and review one bounded artifact. No product protocol or substrate implementation is prescribed by this brief.

A separate phase auditor uses `skills/audit-phase/SKILL.md` against the decided design. Resolve blocking artifact findings.

Handoff: complete `skills/phase-increment-handoff/SKILL.md` in this task. It creates a single planning-shape task, `docs/workplan/phase-4/<allocated-ID>-plan-<increment-slug>.md` numbered in Phase 4, that derives the phase's first increment with a **semantic dependency on P0-004**, because it consumes the reviewed `docs/workplan/phase-4/README.md` produced here. That task completes the same handoff for the next increment. Dispatch this phase-design task through a separate refiner using `skills/refine-task/SKILL.md` against current HEAD.

## Write set

Forecast:

- `docs/workplan/phase-4/README.md` (new phase artifact).
- `docs/workplan/README.md` (phase 4 design row, the planning task's dependency row, affected coverage and graph/collision report).
- `docs/workplan/phase-0/README.md` (increment order 4, Plan Phase 4, design and planning-task coverage).
- `docs/workplan/phase-4/<allocated-ID>-plan-<increment-slug>.md` (the planning-shape task from `phase-increment-handoff`).
- `docs/workplan/tasks.json` (that task's pending entry and its semantic dependency on P0-004; preserve all existing task statuses and metadata).

No existing `tasks.json` status write is assigned. A design-set correction discovered to be necessary must be declared with its concrete file and contract before proceeding through Phase 0's phase-revision routing; do not silently broaden this forecast.

## Contract surfaces

None changed in product code or wire schemas. This task decides only phase-4's delivery boundaries; it consumes the cited design contracts. Shared tracker additions, index and Phase 0 coverage edits are write collisions, not product contract collisions. A proposed change to shared product meaning requires Phase 0 disposition and a reconciled graph before assignment.

## Demonstration

After creating the design and its planning task, a reviewer runs from the repository root:

```sh
sed -n '1,240p' docs/workplan/phase-4/README.md
rg -n 'phase-4|P0-004|Plan Phase 4' docs/workplan/README.md docs/workplan/phase-0/README.md
bash scripts/list-available-tasks.sh
```

Expected artifact observations: Two manually started sandboxes can be the predicted phase-exit demonstration, while launch/stop/lease authority remains explicitly absent. Both index and Phase 0 locate the same design and its planning task. Inspect that task's tracker entry and brief: they agree on its real Phase 4 ID/path and semantic dependency on P0-004, name the phase's first increment, and leave implementation-brief coverage not yet derived. Availability reflects the tracker dependencies, not a claimed runtime milestone. Obtain the independent auditor's findings against the reviewed commit; artifact creation alone is not an audit verdict.

## Verification

- Independent `audit-phase` review checks baseline evidence, decided boundaries, task fit, source links and exit observations; reject absent prerequisites presented as existing behavior.
- Validate tracker against its unchanged schema, resolve dependency IDs/brief paths, and compare tracker/index/Phase 0 coverage; `bash scripts/list-available-tasks.sh` must preserve availability semantics. Confirm every `phase-increment-handoff` action holds and no implementation ID was allocated.
- `git diff --check`, completion-script tests (`python3 -m unittest discover -s test`) and `make check` per AGENTS. Report unavailable toolchains explicitly; do not claim a Go/runtime gate passed when it did not run. No runtime capability is added by this task.

## Acceptance criteria

- [ ] The target artifact is reachable from the index, has the phase-format sections, number-only Order values and descriptive Increment titles, and has no candidate boundary left for the task author to decide.
- [ ] Two manually started sandboxes can be the predicted phase-exit demonstration, while launch/stop/lease authority remains explicitly absent.
- [ ] Source evidence and planned prerequisites are labelled separately; injecting a missing prerequisite or a premature completion assertion yields an audit finding, not a pass.
- [ ] Every included concern has one owner, excluded roadmap concerns have their source destination, and a real deferred decision retains its explicit trigger.
- [ ] Independent phase audit has no unresolved blocking findings; index and Phase 0 match the reviewed artifact.
- [ ] Every `phase-increment-handoff` action holds: exactly one planning-shape task in Phase 4 derives the phase's first increment, with a semantic dependency on P0-004 and matching tracker, index, and Phase 0 and phase coverage. A missing task blocks completion.
- [ ] Unproduced implementation briefs are labelled not yet derived and have no speculative IDs.
- [ ] Existing runtime/source behavior, frozen tracker schema, and task status fields remain intact.

## Deferrals

None. Creating the planning task is part of completion. Implementation-brief derivation belongs to that task, while product implementation remains outside this planning task.

## Traces to

`docs/workplan/phase-0/README.md` ordered planning increment 4 (Plan Phase 4); docs/roadmap.md §3 Stage 3 and §6 D2/D4/D6/D7; docs/architecture.md §5 topology B; docs/contracts.md §4/8; ADR-0006.

