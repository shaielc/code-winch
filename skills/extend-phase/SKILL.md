---
name: extend-phase
description: Extend an existing phase with a new requirement or discovery. Use when adding phase coverage while preserving existing boundaries and handing task derivation to a separate author.
---

# Extend phase

Act as the phase author. Add coverage to an existing phase, then hand task
derivation to a different author using `skills/derive-task/SKILL.md`.

Read applicable `AGENTS.md` files, the requirement or discovery, its design
sources, phase README, index, tracker, relevant briefs, and code. Use
[the model](../shared/workplan-model.md) and
[the phase format](../shared/phase-design.md).

## Procedure

1. State what changed: a new requirement, missing coverage, or discovered defect.
   Name the source and existing behavior it affects.
2. Decide whether it belongs in this phase. Assign out-of-phase work to its
   roadmap stage or triggered decision rather than expanding the phase silently.
3. Add ordered increments with outcomes, boundaries, and prerequisites. Preserve
   completed work and existing labels. If the change splits, combines, or
   redefines an existing increment, use `skills/revise-phase/SKILL.md` instead.
4. Update the outcome, exit, and index where needed. Mark added increments as
   not yet derived and identify the next derivation handoff.
5. Report affected existing tasks. Rework of completed implementation may need
   a revision task; the separate task author derives it and records the edge.

Keep tracker statuses unchanged. Create no placeholder IDs or brief skeletons.
Use malleable fakes to avoid coupling added coverage to an unnecessary real
substrate or the fake's current protocol.

## Done

Added coverage is ordered and bounded, existing work is accounted for, and a
different task author has the next increment and its prerequisites.
