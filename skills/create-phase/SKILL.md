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
   path. Split independently observable concerns.
4. Decide each increment's prerequisites, responsibility, and excluded concerns.
   Do not leave candidate scopes for the task author to resolve.
5. Check inherited invariants. Put required repairs in the partition, with
   run/deploy repairs first. A capability not yet introduced does not require
   building its eventual guarantees in the first increment.
6. Write the phase and update the index. Mark task coverage as not yet derived.
   Identify the next increment for handoff without allocating IDs or briefs.

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
