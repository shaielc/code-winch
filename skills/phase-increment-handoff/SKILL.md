---
name: phase-increment-handoff
description: Hand off a phase's increments for derivation by creating a single planning-shape task. Use at the end of a phase-design task, after its phase README is decided, or of an increment-derivation task, after its derived brief and tracker entry exist; the actions land in that task's change.
---

# Phase increment handoff

Act as the author of the planning task that is handing off:

- a **phase-design task** (Phase 0, "Plan Phase N"), which produced
  `docs/workplan/phase-N/README.md`;
- an **increment-derivation task** (Phase N, "Plan Phase N increment K"), which
  derived one implementation brief with `skills/derive-task/SKILL.md`.

These actions are part of that task and land in its change. Read applicable
`AGENTS.md` files, [the model](../shared/workplan-model.md), the phase README,
the workplan index, and the tracker.

Complete every action below. The task's output is handed off only when all of
them hold.

## Actions

### 1. Add the next plan increment task

Create a single planning-shape task that derives the phase's next increment, so
derivation continues while earlier briefs are refined and implemented.

1. **Select the increment.** Take the lowest-order increment in the phase table
   whose coverage is neither a derived brief nor a derivation task — increment 1
   after a phase-design task. If every increment is covered, record that in the
   phase coverage note and skip the rest of this action.
2. **Copy its boundary.** Read the row: observable change, includes, excludes,
   and prerequisites. The phase decides the boundary; a boundary that does not
   hold goes to the phase author through `skills/revise-phase/SKILL.md`.
3. **Derive the dependency.** Add a `semantic` edge to each planning task that
   produces an artifact the new task consumes, completed or not:
   - the phase-design task that produces the phase README;
   - the derivation task of each prerequisite increment, whose brief defines
     what the new derivation consumes.

   Existing behavior at HEAD creates no edge. Never depend on an implementation
   task, and never add an edge from adjacent increment numbers.
4. **Write the brief** at `docs/workplan/phase-N/<ID>-plan-<increment-slug>.md`
   in the model's planning shape, like the phase's existing derivation briefs:
   `# <ID>: Plan Phase <N> increment <K>: <increment>`, `Shape: planning`, and
   the dependencies with their reasons.
   - **Objective:** one complete implementation brief for this increment,
     reconciled with tracker, index, and Phase 0 and phase coverage, stating the
     row's observable change.
   - **Scope:** use `skills/derive-task/SKILL.md` against the phase, the consumed
     briefs, and HEAD; derive exactly this increment; name the surfaces to
     inspect and the excluded neighbouring increments. When a consumed brief is
     not yet implemented, plan against the brief and record the derived task's
     edge to that implementation.
   - **Required constraints:** the row's inclusions and exclusions, the
     applicable concern-boundary owners and evidence limits, and the inherited
     invariants the derived brief must preserve.
   - **Non-goals:** the row's exclusions; one implementation brief only; no
     reserved later implementation IDs.
   - **Planning outputs and handoff:** the derived brief at
     `<allocated-ID>-<slug>.md`, its tracker entry, and the records to
     reconcile. Handoff: complete `skills/phase-increment-handoff/SKILL.md` in
     this task; a different refiner then takes the derived brief with
     `skills/refine-task/SKILL.md`.
   - **Write set:** the derived brief, the next plan increment task's brief,
     `docs/workplan/tasks.json`, the index, and the Phase 0 and phase READMEs.
   - **Contract surfaces:** none for the planning task; name the surfaces the
     derived brief will own and consume, and the `contract` edge it takes
     wherever it extends a surface another brief defines.
   - **Demonstration, Verification, Acceptance criteria:** artifact review and
     the repository gates, checkable from the task's own diff, including that
     this handoff was completed.
   - **Deferrals** and **Traces to** as the model requires.
5. **Integrate it.** Allocate the next ID in Phase N; after a derivation task,
   that is after the derived brief's. Add the brief and a `pending` tracker
   entry together, mark the increment's coverage in the phase README and the
   Phase 0 row as a derivation task assigned, and add the index dependency row
   with its reason. Allocate no implementation ID and leave every existing
   status unchanged.

## Report

State the graph after all actions: nodes, edges, critical path, layer widths and
average width, incomparable pairs, available tasks, write collisions, and
contract collisions.

## Done

Every action holds in the task's change: the next undelivered increment has
exactly one planning-shape derivation task with a complete brief, edges only to
planning work it consumes, and matching tracker, index, and coverage records, or
the phase records that every increment is covered.
