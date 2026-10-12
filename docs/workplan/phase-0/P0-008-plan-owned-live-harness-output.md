# P0-008: Plan Phase 1 increment 1: owned live harness output

**Phase:** 0 — Plan the implementation phases
**Shape:** planning
**Dependencies:** P0-001

## Objective

Produce one complete, independently authored implementation brief for Phase 1
increment 1, **Owned live harness output**, and reconcile that brief with the
tracker, workplan index, Phase 0 coverage, and Phase 1 coverage.

## Scope

Use `skills/derive-task/SKILL.md` to derive the first Phase 1 implementation
task from the fixed boundaries in `docs/workplan/phase-1/README.md`. Allocate
the next real plan-wide implementation ID only when writing that complete
brief. The brief must turn the increment's single observable change into a
runnable vertical slice: starting the sandbox owns and reaps the fake child,
raw incremental output reaches the existing attach page, and a maintained
operator path can observe it.

The derived brief must include the runner-local session-record seam, identity
and ordinal allocation, append-before-publication behavior, bounded basic
cleanup, composition and image wiring, and fake-profile scenario growth that
the phase design assigns to increment 1. It must preserve the existing
startup, posture, status, page, loopback deployment, and repository gates.
Inspect HEAD and name concrete owned files and contract surfaces rather than
assuming the design destination already exists.

Apply the workplan scope test and brief anatomy. Separate required behavior
from proposed implementation, state reachability through the running sandbox,
forecast concrete writes, derive only semantic/compile/contract dependencies,
and include acceptance, adversarial verification, and a credible manual
demonstration. Treat the fake harness as a malleable shipped profile, not as a
product protocol constraint. This planning task assigns brief derivation; it
does not derive the brief or implement the runtime capability itself.

## Required constraints

P0-001's reviewed Phase 1 design is the authority for increment boundaries,
prerequisites, concern ownership, and exclusions. Do not revise that design in
the task-author role. If HEAD makes the boundary invalid, report the defect to
a separate Phase 1 phase author using `skills/revise-phase/SKILL.md` and resume
derivation only after disposition.

At dispatch, recheck the current source and distinguish existing behavior from
planned behavior. The present baseline has a sandbox composition root that
constructs the attach server, a static attach page with browser-local composer
state, a status-only host CLI, and a fake-harness binary shipped but not
started. The derived task must leave the daemon startable and deployable, wire
every seam it introduces, and keep the fake profile runnable by hand.

Allocate no IDs for later Phase 1 increments. If the derived brief needs a
completion deferral, its owner must already exist in `tasks.json`; otherwise
include the work required for increment 1 or return the scope issue through
Phase 0. Shared file forecasts are collision warnings, not ordering edges.

## Non-goals

Do not write the Phase 1 implementation brief or product code in this task.
Do not revise Phase 1 boundaries. Retained reload history, input recording,
input delivery and replies, malformed-output hardening, and isolated slow
readers remain the distinct later increments fixed by the Phase 1 artifact;
they receive separately authored planning handoffs rather than placeholder
IDs. Roadmap Stages 1–6 remain outside this Stage 0 brief.

## Planning outputs and handoff

- A separate task author uses `skills/derive-task/SKILL.md` to create
  `docs/workplan/phase-1/<allocated-ID>-owned-live-harness-output.md` as one
  complete implementation brief and allocates its real ID in
  `docs/workplan/tasks.json` at the same time.
- That author updates `docs/workplan/README.md`,
  `docs/workplan/phase-0/README.md`, and
  `docs/workplan/phase-1/README.md` so the dependency reason, coverage, graph
  metrics, and collision reports match the new brief and tracker entry.
- The completed implementation brief is handed to a different refiner using
  `skills/refine-task/SKILL.md`; implementation and audit remain separate
  operations.
- If actionable Phase 1 coverage remains, arrange the next concrete,
  separately authored Phase 0 planning handoff one increment at a time without
  reserving speculative IDs.

## Write set

Forecast for the assigned operation:
`docs/workplan/phase-1/<allocated-ID>-owned-live-harness-output.md`,
`docs/workplan/tasks.json` (one new entry without status edits),
`docs/workplan/README.md`, `docs/workplan/phase-0/README.md`, and
`docs/workplan/phase-1/README.md`. The task author must replace the path
placeholder with the allocated real ID and make no product-code changes while
deriving the brief.

## Contract surfaces

No product contract changes are made by this planning task. The assigned brief
must declare the session-record, child-process, live-publication, attach-page,
and operator-path surfaces it will own or consume, and must turn concurrent
meaning changes into dependency edges. Shared planning records are write
collisions only.

## Demonstration

Review the newly derived implementation brief and confirm that its observable
result is exactly Phase 1 increment 1; its runtime path begins at sandbox
startup and reaches both the attach page and maintained operator observation;
its files, contracts, dependencies, acceptance criteria, verification, and
manual commands are concrete; and the tracker, index, and both phase coverage
tables agree. This is an artifact review, not a runtime demonstration of live
harness output.

## Verification

Run the planning consistency tests required by the repository and validate
`docs/workplan/tasks.json` against its schema. Confirm that the allocated brief
exists, all dependency IDs exist, no task status changed, coverage is exact,
and graph/collision figures in the index reflect the actual tracker. Report
repository gates accurately; this task introduces no runtime behavior.

## Acceptance criteria

- [ ] Exactly one real Phase 1 implementation brief covers Owned live harness
      output at the fixed increment depth and has a matching tracker entry.
- [ ] The brief includes the complete reachable sandbox-to-child-to-live-reader
      slice, record seam, cleanup, composition, fake scenario, maintained
      operator path, and inherited gates without absorbing another increment.
- [ ] Existing behavior, predicted behavior, concrete writes, owned/consumed
      contracts, dependency reasons, adversarial checks, and manual evidence
      are explicit and source-supported.
- [ ] Tracker, index, Phase 0 coverage, Phase 1 coverage, graph metrics, and
      collision reports agree; existing statuses are unchanged.
- [ ] The implementation brief is handed to a separate refiner, and remaining
      coverage has no placeholder implementation IDs or unnamed deferrals.

## Deferrals

Product implementation belongs to the implementation task whose brief this
planning operation will allocate. Remaining Phase 1 increments retain their
explicit not-yet-derived coverage in the fixed phase design and require their
own concrete Phase 0 handoffs; no speculative owner is created here.

## Traces to

`docs/workplan/phase-1/README.md` increment 1; Phase 0 ordered planning
increment 1 (legacy 0.1); `docs/roadmap.md` Stage 0; `docs/contracts.md` §§8–9;
ADR-0005; ADR-0006; and `skills/shared/workplan-model.md` invariants I1–I7.
