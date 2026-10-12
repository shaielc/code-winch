# P0-008: Plan Phase 1 increment 1: child status presentation

**Phase:** 0 — Plan the implementation phases
**Shape:** planning
**Dependencies:** P0-001

## Objective

Produce one complete, independently authored implementation brief for Phase 1
increment 1, **Child status presentation**, and reconcile the brief with the
tracker, index, Phase 0 coverage and Phase 1 coverage.

## Scope

Use `skills/derive-task/SKILL.md` against the reviewed
`docs/workplan/phase-1/README.md`. The first implementation increment adds
an honest child-status element to the already running attach page. Its manual
demonstration is to start the existing sandbox, open the attach page, and see
`Unknown` while no authoritative child status is available. This is a real
user-visible increment; do not invent child spawning, a new backend status
endpoint, or a fake `Running` assertion to make it look more complete.

Derive the single smallest observable increment at the decided boundary; leave
implementation details, evidence and acceptance checks to `derive-task`.
Allocate the implementation ID only with the complete brief.

## Required constraints

Keep the `Unknown` presentation truthful: it does not imply child execution or
input readiness. Increment 2 introduces sandbox-backed status and increment 3
introduces child execution. If source inspection invalidates this boundary,
return it to the phase author rather than silently broadening the task.

## Non-goals

No Phase 1 product implementation in this planning task; the derived increment
itself adds no child supervisor, live-output transport, input delivery,
persistence, harness-specific availability, vendor semantics, or broad
restructuring. Do not derive a second Phase 1 implementation brief in this task. The next
Phase 0 planning brief must be authored separately and registered; later Phase 1
implementation briefs retain their not-yet-derived coverage.

## Planning outputs and handoff

- A different task author derives
  `docs/workplan/phase-1/<allocated-ID>-child-status-presentation.md` with
  one concrete tracker entry created at the same time.
- Reconcile `docs/workplan/README.md`,
  `docs/workplan/phase-0/README.md`, `docs/workplan/phase-1/README.md`,
  tracker dependencies and coverage, graph metrics, and write/contract collisions.
- Hand the finished implementation brief to a separate refiner using
  `skills/refine-task/SKILL.md`.
- **Create the next concrete Phase 0 planning task** for increment 2,
  Sandbox-backed child status, with its own independently authored brief and
  tracker/index/coverage entries. This is an actual planning-task creation,
  not just a recommendation or an untracked handoff; do not allocate an
  implementation ID for increment 2.

## Write set

Forecast for the assigned operation:
`docs/workplan/phase-1/<allocated-ID>-child-status-presentation.md`,
`docs/workplan/tasks.json` (one new entry; statuses unchanged),
`docs/workplan/README.md`, `docs/workplan/phase-0/README.md`,
`docs/workplan/phase-1/README.md`, plus the new concrete Phase 0
planning brief and its tracker entry for the next increment. No later
implementation IDs are reserved.

## Contract surfaces

This Phase 0 planning task changes no product contract. The implementation
brief must state the attach-page presentation it owns and the existing
sandbox/static-page boundary it consumes, avoiding future child/process
contracts and documenting any shared writing collisions.

## Demonstration

Review the resulting brief: its observable behavior is specifically the
`Unknown` child-status presentation against the deployed page; the manual
steps, real runtime reachability, existing/inherited gates, scope exclusions,
files, contracts, tests and task dependencies are credible. Confirm tracker,
index and both phase coverage tables agree. This is an artifact review, not a
claim that the new UI has already been implemented.

## Verification

Validate tracker schema, brief existence, dependency references, status
preservation, links, graph/collision reporting and applicable planning tests.
Record any gates not run or unavailable.

## Acceptance criteria

- [ ] One complete Phase 1 implementation brief covers exactly the first,
      truthful child-status UI increment and has a matching real tracker ID.
- [ ] Its demonstration uses the deployed attach page and maintains operator
      reachability without child spawning or a newly invented protocol.
- [ ] Owned/consumed contracts, concrete files, expected observation,
      adversarial/automated verification and inherited invariants are explicit.
- [ ] Tracker, index, Phase 0/1 coverage, graph and collision warnings agree;
      preexisting status values are unchanged.
- [ ] The implementation brief goes to a separate refiner, and a separately
      authored concrete Phase 0 planning brief for increment 2 exists in the
      tracker and coverage records without speculative implementation IDs.

## Deferrals

Later Phase 1 increments own deeper child state, output, input, persistence
and hardening in the decided phase artifact. These increments are coverage
boundaries, not IDs that authorize an implementation brief to introduce an
unowned completion deferral. The assigned author must not leave a dangling
implementation promise.

## Traces to

`docs/workplan/phase-1/README.md` increment 1; Phase 0 increment 0.1;
`docs/roadmap.md` Stage 0; `docs/contracts.md` §§8–9; ADR-0005/0006;
`skills/shared/workplan-model.md` invariants I1–I7.
