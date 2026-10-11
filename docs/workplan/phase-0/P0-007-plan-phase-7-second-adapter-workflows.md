# P0-007: Plan phase-7: second adapter and workflows

**Phase:** 0 — Plan the implementation phases
**Shape:** planning
**Dependencies:** P0-002 (semantic: this artifact must consume the first-adapter seam decided in phase-2/README.md); P0-005 (semantic: workflow boundaries require the ordinary run commands and event ownership decided in phase-5/README.md)

## Objective

Produce a reviewed `docs/workplan/phase-7/README.md` with decided ordered increments for roadmap Stage 6, ready for a separate task author to derive implementation briefs.

## Scope

- Partition second real-adapter parity from independently observable cross-run workflow behavior; use the second case to test the first seam rather than pre-building provider abstractions.
- Plan the versioned workflow graph, deterministic step command keys, durable replay, event waits and bounded branching through ordinary application commands.
- Keep provider-specific behavior out of generic workflow/application code; make effective policy no broader than its deployment/workspace/workflow/profile intersection.
- State the roadmap's proven-first-adapter requirement as execution evidence to recheck when Stage 6 opens; completion of a planning dependency does not prove vendor runtime parity.

## Required constraints

Neither a real vendor seam nor a workflow runtime exists at the baseline. First-adapter and ordinary run contracts are planned inputs, not certified capabilities.

phase-2/README.md's adapter boundary and phase-5/README.md's ordinary run commands/events are concrete prerequisite artifacts; no runtime implementation is required for artifact review.

Use `skills/create-phase/SKILL.md` as the phase-author operation, with `skills/shared/phase-design.md` and the workplan model. The required behavior is a decided phase artifact, consistent index/Phase 0 coverage, and an independent `audit-phase` review. Product behavior and commands described there are predictions. Preserve Phase 0's decided boundaries and all task status fields.

## Non-goals

Unspecified future product expansion, arbitrary server code in definitions, new sandbox backends solely for parity, or shared-policy weakening.

Deriving implementation briefs, implementing runtime capability, or claiming a phase completed merely because its design exists. Specific-brief derivation remains uncovered in Phase 0; it is a separate operation, not a completion deferral of this phase-design task.

## Planning outputs and handoff

The phase author creates `docs/workplan/phase-7/README.md`, records its source-supported current baseline and planned prerequisites, chooses ordered increments and concern owners, states exit observations, and marks implementation task coverage **not yet derived**. Update this phase's row in the index and Phase 0 coverage. Proposed approach: inspect current source and applicable AGENTS first, compare against roadmap Stage 6, then write and review one bounded artifact. No product protocol or substrate implementation is prescribed by this brief.

A separate phase auditor uses `skills/audit-phase/SKILL.md` against the resulting commit. Resolve blocking artifact findings before handing the decided increments to a different task author through `skills/derive-task/SKILL.md`, one brief at a time. This phase author must not derive its own implementation briefs. Dispatch this planning task through a separate refiner using `skills/refine-task/SKILL.md` against current HEAD.

## Write set

Forecast:

- `docs/workplan/phase-7/README.md` (new phase artifact).
- `docs/workplan/README.md` (only phase 7 design/index row and affected coverage).
- `docs/workplan/phase-0/README.md` (only increment order 7, legacy reference 0.7, design coverage/handoff).

No `tasks.json` status write is assigned. A design-set correction discovered to be necessary must be declared with its concrete file and contract before proceeding through Phase 0's phase-revision routing; do not silently broaden this forecast.

## Contract surfaces

None changed in product code or wire schemas. This task decides only phase-7's delivery boundaries; it consumes the cited design contracts. Shared index and Phase 0 coverage edits are write collisions, not product contract collisions. A proposed change to shared product meaning requires Phase 0 disposition and a reconciled graph before assignment.

## Demonstration

After creating the artifact, a reviewer runs from the repository root:

```sh
sed -n '1,240p' docs/workplan/phase-7/README.md
rg -n 'phase-7|P0-007|0\.7' docs/workplan/README.md docs/workplan/phase-0/README.md
bash scripts/list-available-tasks.sh
```

Expected artifact observations: The design has distinct second-adapter and cross-run workflow increments, with workflow replay not duplicating effects and no direct process control by the coordinator. Both index and Phase 0 locate the same design and distinguish pending implementation-brief coverage. Availability reflects the tracker dependencies, not a claimed runtime milestone. Obtain the independent auditor's findings against the reviewed commit; artifact creation alone is not an audit verdict.

## Verification

- Independent `audit-phase` review checks baseline evidence, decided boundaries, task fit, source links and exit observations; reject absent prerequisites presented as existing behavior.
- Validate tracker against its unchanged schema, resolve dependency IDs/brief paths, and compare tracker/index/Phase 0 coverage; `bash scripts/list-available-tasks.sh` must preserve availability semantics.
- `git diff --check`, completion-script tests (`python3 -m unittest discover -s test`) and `make check` per AGENTS. Report unavailable toolchains explicitly; do not claim a Go/runtime gate passed when it did not run. No runtime capability is added by this task.

## Acceptance criteria

- [ ] The target artifact is reachable from the index, has the phase-format sections, number-only Order values and descriptive Increment titles, and has no candidate boundary left for the task author to decide.
- [ ] The design has distinct second-adapter and cross-run workflow increments, with workflow replay not duplicating effects and no direct process control by the coordinator.
- [ ] Source evidence and planned prerequisites are labelled separately; injecting a missing prerequisite or a premature completion assertion yields an audit finding, not a pass.
- [ ] Every included concern has one owner, excluded roadmap concerns have their source destination, and a real deferred decision retains its explicit trigger.
- [ ] Independent phase audit has no unresolved blocking findings; index and Phase 0 match the reviewed artifact and show implementation briefs as not yet derived.
- [ ] Existing runtime/source behavior, frozen tracker schema, task status fields, and separate-author handoff remain intact.

## Deferrals

None. Product implementation and specific-brief derivation are non-goals of this artifact-producing task, not unfinished portions of its completion condition.

## Traces to

`docs/workplan/phase-0/README.md` ordered planning increment 7 (legacy reference 0.7); docs/roadmap.md §3 Stage 6; docs/contracts.md §7/9; docs/architecture.md §4 Workflow coordinator; docs/security.md §10; ADR-0003 (Proposed).

