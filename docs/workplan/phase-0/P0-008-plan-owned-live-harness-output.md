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

The derived brief must inspect HEAD, name concrete files and contracts,
preserve the existing startup/health/posture/page/status and loopback deployment,
ensure the page change is reachable in a deployed configuration, and provide
the minimum maintained operator path appropriate for observing the result.
State exact manual actions and expected observations separately from
automated tests, negative checks and repository gates. Choose the required
implementation shape, write set, and contract ownership from the actual source.
Allocate the next real plan-wide implementation ID only when creating the
complete brief. This planning task does not implement the UI.

## Required constraints

P0-001's reviewed Phase 1 artifact owns the partition; this task author must
not revise its increments. Route a defective boundary to a separate phase
author through `skills/revise-phase/SKILL.md` before continuing. The current
sandbox already serves the static attach page and has a browser-local composer;
the fake binary is shipped but is not started by the sandbox. Recheck these
claims at dispatch.

The `Unknown` state means the page has not established child status; it is not
a claim that the child is running or accepting input. Increment 2 owns
sandbox-backed `Down`, increment 3 owns actual child spawning/status, and
later increments own output and input behavior. No unsupported or speculative
implementation task IDs may be reserved. If a new completion deferral is
necessary, derive its real owner first or return the scope issue through
Phase 0. Shared planning file edits are collision warnings, not ordering edges.

## Non-goals

No Phase 1 product implementation in this planning task; the derived increment
itself adds no child supervisor, live-output transport, input delivery,
persistence, harness-specific availability, vendor semantics, or broad
restructuring. Do not derive subsequent increments or implementation briefs
as part of this task. Other increments retain their not-yet-derived coverage
and get separate Phase 0 planning handoffs.

## Planning outputs and handoff

- A different task author derives
  `docs/workplan/phase-1/<allocated-ID>-child-status-presentation.md` with
  one concrete tracker entry created at the same time.
- Reconcile `docs/workplan/README.md`,
  `docs/workplan/phase-0/README.md`, `docs/workplan/phase-1/README.md`,
  tracker dependencies and coverage, graph metrics, and write/contract collisions.
- Hand the finished implementation brief to a separate refiner using
  `skills/refine-task/SKILL.md`. Continue the Phase 0 handoff chain for the
  next uncovered increment without reserving future IDs.

## Write set

Forecast for the assigned operation:
`docs/workplan/phase-1/<allocated-ID>-child-status-presentation.md`,
`docs/workplan/tasks.json` (one new entry; statuses unchanged),
`docs/workplan/README.md`, `docs/workplan/phase-0/README.md`,
`docs/workplan/phase-1/README.md`. The placeholder is replaced with the
allocated ID by the separate task author.

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
- [ ] The brief goes to a separate refiner and the next handoff remains
      independently assigned without speculative IDs.

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
