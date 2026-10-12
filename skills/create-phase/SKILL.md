---
name: create-phase
description: Create a phase design from the architecture, roadmap, and existing system. Use when opening an implementation phase and selecting ordered increments and boundaries before task briefs are derived.
---

# Create phase

Act as the phase author. Produce `docs/workplan/phase-N/README.md` and its
entry in the plan index. Hand the phase to a different task author using
`skills/derive-task/SKILL.md`; phase creation ends before brief derivation.

Read applicable `AGENTS.md` files, the design set, roadmap, `docs/state.md`,
and the active plan. Use [the model](../shared/workplan-model.md) and
[the phase format](../shared/phase-design.md).

Phase 0 partitions planning artifacts and handoffs. For that phase, use the
model's planning shape rather than requiring new runtime wiring or operator
commands. Later phases partition product implementation.
When opening a new workplan without an explicitly selected target phase, create
Phase 0 first to schedule phase design and subsequent specific-brief derivation.
An explicit request to design a named implementation phase selects that target.

## Procedure

1. Establish what works now. Inspect code and completed briefs; use existing
   entrypoints where available. Distinguish observed behavior, source evidence,
   and behavior still to be built. Reuse completed work.
2. Select the phase outcome from the design and roadmap. State what a person
   can do at exit, and what remains outside the phase.
3. Partition the outcome into ordered increments. Each adds one observable
   property at its own depth, including its required wiring and minimum operator
   path. Split independently observable concerns. Begin from the shipped
   interaction surface and progressively deepen what a person can observe.
   Challenge every boundary for a smaller useful outcome, including truthful
   presentation before authoritative backend behavior. When combining such
   outcomes, state why separation would not produce a useful increment;
   increment count is not a target.
4. Decide each increment's prerequisites, responsibility, and excluded concerns.
   Do not leave candidate scopes for the task author to resolve. Derive edges
   from consumed behavior or contracts, not adjacent increment numbers. Separate
   abstract, generalized constraints from demonstrations tied to a particular
   implementation or instance. State the limits of what each demonstration
   establishes; do not infer general guarantees from instance-specific behavior.
5. Check inherited invariants. Put required repairs in the partition, with
   run/deploy repairs first. A capability not yet introduced does not require
   building its eventual guarantees in the first increment.
6. Write the phase and update the index. Mark task coverage as not yet derived.
   Identify the next increment for handoff without allocating IDs or briefs.
   After repartitioning, reconcile existing tracker and brief titles, coverage,
   mappings, and dependency reasons with the decided boundaries, preserving IDs
   and statuses. Keep reusable procedure in skills and task-specific outcomes
   and constraints in artifacts.

Challenge upstream scope rather than accepting a bundled objective as its own
justification. Correct a contract or roadmap that contradicts the selected
boundary. Keep product meaning in the design set and delivery order in the
roadmap and phase.

Treat fakes as malleable: adapt them to exercise an increment rather than
designing the increment around their current implementation. Leave commands,
fixtures, protocol choices, and file forecasts to the task author.

## Done

The phase has a known starting point, numbered increments, explicit boundaries,
prerequisites, and an observable exit. The index points to it. A different task
author has a clear next increment.
