---
name: derive-task
description: Derive one task brief from a decided phase increment. Use when translating a phase into a specific task or rederiving a pending brief, with tracker, dependency, index, and coverage updates.
---

# Derive task

Act as the task author, separate from the phase author. Read applicable
`AGENTS.md` files, [the model](../shared/workplan-model.md), the phase README,
its design sources, index, tracker, relevant briefs, and implementation.

Derive one task at a time: architecture → phase author → task author → optional
phase revision by the phase author → next task. Deriving another task does not
require implementing the previous one first.

Planning-shaped tasks use the model's planning shape and belong to one of two
places. Phase 0 holds only meta tasks: designing or revising a phase,
converting an issue into phase tasks, and updating the design set (see *Phase 0
— meta tasks* in [the model](../shared/workplan-model.md)). Deriving the briefs
for an increment a phase has already decided is that phase's own planning task,
not a Phase 0 task.

For either, name the target artifact, the assigned phase/task-author operation,
and the handoff. Deriving this brief does not itself perform the operation it
assigns. Every acceptance criterion must be checkable from the task's own
artifacts; name the next operation in the handoff, not as a criterion.

## Select the change

1. Choose the next uncovered part of an increment in the phase's stated order.
   Check prerequisites: distinguish existing behavior from behavior promised by
   a pending task, and identify the actual surface consumed.
2. State one observable change at the increment's depth. Remove completed work.
   A task can cover less than an increment when a smaller useful closed loop
   exists; it cannot silently combine increments.
3. Apply the model's scope test. Include required wiring, invariant preservation,
   and the minimum maintained operator path. Extract distinct capabilities,
   hardening, and future enablement.
4. If the boundary fails, report the proposed change to the phase author using
   `skills/revise-phase/SKILL.md`. Resume after that decision; do not revise
   the phase in this role.

## Write the brief

Use the shared anatomy and select a task shape. State required behavior and
product constraints separately from proposed implementation approaches.
Provide reachability, concrete write forecasts, owned contract surfaces,
acceptance criteria, verification, and a credible manual demonstration.
For a planning shape, use planning outputs and artifact review in place of
runtime reachability and demonstrations of the target product capability.

Treat the fake as malleable. Change its behavior, protocol, timing, or controls
to exercise this task; do not broaden product scope to accommodate its current
implementation. Update affected callers, tests, and documentation together.
Choose only controls the task needs. Standalone fixture flags are not a
checklist of product configuration keys.

Inspect observation paths. Run what exists; label new behavior and commands as
predictions. Demonstrate the chosen capability without adding another solely
for test convenience: a live reader can observe new output before durable replay.

Deferrals need an actual owner in `tasks.json`, or a genuine architectural
decision with a trigger in the design set. An increment label is not a task ID.
Include work needed for this task's completion, or have its real owner derived
before recording a deferral. Create no placeholders merely to supply an ID.

## Integrate the task

1. Allocate the next ID of the new task's own phase: one past the highest
   `P<phase>-` number in `tasks.json`, never filling a gap. A task that derives
   a decided increment is a task of that increment's phase, and the briefs it
   produces take their numbers from the same sequence. Only a Phase 0 meta task
   — an issue conversion, a phase design, a design-set update — takes a Phase 0
   number, and the briefs an issue conversion produces take their numbers from
   the phase that owns the work. Preserve IDs when rederiving pending briefs,
   and never renumber merged or completed tasks.
2. Add or reconcile the brief, tracker entry, dependency reasons, index row,
   and phase coverage mapping together. Updating coverage does not authorize
   changing the phase's boundaries.
3. Derive compile, contract, semantic, or discovered revision edges. Shared
   files create collision warnings rather than edges.
4. Report the current graph's critical path, average width, concurrent write
   collisions, and contract collisions. Label partial derivation explicitly.
5. Leave status fields unchanged. For legacy rederivation, reconcile affected
   graph references and remove the reconciliation marker only when the brief
   and graph match the decided increment.

## Done

One complete brief fits its increment and has credible observations, valid
ownership and dependencies, and matching tracker/index/coverage records.
Remaining increments are explicit. Hand the task to a separate refiner using
`skills/refine-task/SKILL.md` when it is dispatched against current code.
