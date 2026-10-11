---
name: revise-phase
description: Revise an existing phase's increments, order, or boundary decisions. Use when derivation, implementation, or an accepted audit finding exposes a flawed partition or requires reassigning phase responsibility.
---

# Revise phase

Act as the phase author. Decide the revised partition and hand affected task
derivation to a different author using `skills/derive-task/SKILL.md`.

Read applicable `AGENTS.md` files, the change request or finding, design sources,
phase README, index, tracker, affected briefs, and code. Use
[the model](../shared/workplan-model.md) and
[the phase format](../shared/phase-design.md).

## Procedure

1. Identify the failed boundary: bundled concerns, missing coverage, repeated
   completed work, a conflicting contract, or a changed prerequisite. A filename
   forecast or local implementation choice alone does not revise a phase.
2. Decide the new ordered partition. State each changed increment's observable
   outcome, inclusions, exclusions, and prerequisites. Update the roadmap or
   contract if the decision changes their meaning.
3. Preserve completed IDs and record reusable behavior. Rework of completed
   implementation needs a separately derived revision task, never a rewritten
   account of what the original task promised.
4. Record affected pending briefs as requiring rederivation. Update mappings
   and index so legacy briefs are not presented as conforming to the new phase.
   Identify the next derivation handoff.
5. Have the separate task author reconcile affected briefs, dependency reasons,
   tracker graph, index, and collision report before assignment. Keep incomplete
   reconciliation explicit until that handoff is done.

Decide phase boundaries here. A task author can propose a revision but does not
silently combine increments or rewrite the phase while deriving a task.
Keep status fields unchanged and allocate no placeholder IDs.

## Done

The revised order and boundaries are explicit. Completed work and affected
briefs are accounted for. The index states remaining reconciliation, and the
next task handoff is clear.
