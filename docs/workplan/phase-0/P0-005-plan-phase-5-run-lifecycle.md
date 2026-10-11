# P0-005: Plan phase-5: run lifecycle control

**Phase:** 0 — Plan the implementation phases
**Shape:** planning
**Dependencies:** P0-004 (semantic: this artifact must extend the decided observe-only run/registration boundary in phase-4/README.md)

## Objective

Produce a reviewed `docs/workplan/phase-5/README.md` with decided ordered increments for roadmap Stage 4, ready for a separate task author to derive implementation briefs.

## Scope

- Partition product-driven creation, preparation/start, intentional stop and run-state observations at individually closed depths.
- Own lifecycle, leases/fencing, creation/input idempotency, transactional outbox, canonical gap-free ordering and restart reconciliation at the roadmap's Stage 4 depth.
- Preserve direct sandbox attach and the maintained operator path; keep the single-user loopback deployment posture explicit.
- Resolve D7 at first persistence of runner records if the Stage 3 design has not triggered it; consume an existing decision rather than deciding it twice.

## Required constraints

No run aggregate, supervisor, outbox, leases, canonical persistence or product launch/stop exists at the baseline.

phase-4/README.md's registration and observed-run record boundary is consumed as a decided planning artifact, not implemented runtime behavior.

Use `skills/create-phase/SKILL.md` as the phase-author operation, with `skills/shared/phase-design.md` and the workplan model. The required behavior is a decided phase artifact, consistent index/Phase 0 coverage, and an independent `audit-phase` review. Product behavior and commands described there are predictions. Preserve Phase 0's decided boundaries and all task status fields.

## Non-goals

Shared deployment and Stage 5 security guarantees, including forced-stop escalation/orphan hardening; workflows (Stage 6).

Deriving implementation briefs, implementing runtime capability, or claiming a phase completed merely because its design exists. Specific-brief derivation remains uncovered in Phase 0; it is a separate operation, not a completion deferral of this phase-design task.

## Planning outputs and handoff

The phase author creates `docs/workplan/phase-5/README.md`, records its source-supported current baseline and planned prerequisites, chooses ordered increments and concern owners, states exit observations, and marks implementation task coverage **not yet derived**. Update this phase's row in the index and Phase 0 coverage. Proposed approach: inspect current source and applicable AGENTS first, compare against roadmap Stage 4, then write and review one bounded artifact. No product protocol or substrate implementation is prescribed by this brief.

A separate phase auditor uses `skills/audit-phase/SKILL.md` against the resulting commit. Resolve blocking artifact findings before handing the decided increments to a different task author through `skills/derive-task/SKILL.md`, one brief at a time. This phase author must not derive its own implementation briefs. Dispatch this planning task through a separate refiner using `skills/refine-task/SKILL.md` against current HEAD.

## Write set

Forecast:

- `docs/workplan/phase-5/README.md` (new phase artifact).
- `docs/workplan/README.md` (only phase 5 design/index row and affected coverage).
- `docs/workplan/phase-0/README.md` (only increment order 5, legacy reference 0.5, design coverage/handoff).

No `tasks.json` status write is assigned. A design-set correction discovered to be necessary must be declared with its concrete file and contract before proceeding through Phase 0's phase-revision routing; do not silently broaden this forecast.

## Contract surfaces

None changed in product code or wire schemas. This task decides only phase-5's delivery boundaries; it consumes the cited design contracts. Shared index and Phase 0 coverage edits are write collisions, not product contract collisions. A proposed change to shared product meaning requires Phase 0 disposition and a reconciled graph before assignment.

## Demonstration

After creating the artifact, a reviewer runs from the repository root:

```sh
sed -n '1,240p' docs/workplan/phase-5/README.md
rg -n 'phase-5|P0-005|0\.5' docs/workplan/README.md docs/workplan/phase-0/README.md
bash scripts/list-available-tasks.sh
```

Expected artifact observations: The design identifies product-visible create/stop loops and separates persisted intent, observed execution and immutable attempts; it states observable duplicate/stale/restart checks. Both index and Phase 0 locate the same design and distinguish pending implementation-brief coverage. Availability reflects the tracker dependencies, not a claimed runtime milestone. Obtain the independent auditor's findings against the reviewed commit; artifact creation alone is not an audit verdict.

## Verification

- Independent `audit-phase` review checks baseline evidence, decided boundaries, task fit, source links and exit observations; reject absent prerequisites presented as existing behavior.
- Validate tracker against its unchanged schema, resolve dependency IDs/brief paths, and compare tracker/index/Phase 0 coverage; `bash scripts/list-available-tasks.sh` must preserve availability semantics.
- `git diff --check`, completion-script tests (`python3 -m unittest discover -s test`) and `make check` per AGENTS. Report unavailable toolchains explicitly; do not claim a Go/runtime gate passed when it did not run. No runtime capability is added by this task.

## Acceptance criteria

- [ ] The target artifact is reachable from the index, has the phase-format sections, number-only Order values and descriptive Increment titles, and has no candidate boundary left for the task author to decide.
- [ ] The design identifies product-visible create/stop loops and separates persisted intent, observed execution and immutable attempts; it states observable duplicate/stale/restart checks.
- [ ] Source evidence and planned prerequisites are labelled separately; injecting a missing prerequisite or a premature completion assertion yields an audit finding, not a pass.
- [ ] Every included concern has one owner, excluded roadmap concerns have their source destination, and a real deferred decision retains its explicit trigger.
- [ ] Independent phase audit has no unresolved blocking findings; index and Phase 0 match the reviewed artifact and show implementation briefs as not yet derived.
- [ ] Existing runtime/source behavior, frozen tracker schema, task status fields, and separate-author handoff remain intact.

## Deferrals

None. Product implementation and specific-brief derivation are non-goals of this artifact-producing task, not unfinished portions of its completion condition.

## Traces to

`docs/workplan/phase-0/README.md` ordered planning increment 5 (legacy reference 0.5); docs/roadmap.md §3 Stage 4 and §6 D7; docs/contracts.md §1–5; docs/architecture.md §4/7; ADR-0005/0006.

