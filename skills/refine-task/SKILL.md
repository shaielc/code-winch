---
name: refine-task
description: Refine one derived task brief against current code before implementation. Use when a task is dispatched and its assumptions, scope, criteria, and demonstration need to become concrete.
---

# Refine task

Work on one task ID. Read its brief end to end, the phase README, index,
tracker, design sections under Traces to, [the model](../shared/workplan-model.md),
and applicable `AGENTS.md` files. The brief owns scope and acceptance; the tracker
owns status and dependencies. Use a separate author for each task operation:
derive, refine, implement, and audit.

For a planning shape, refine the concrete outputs and handoff against current
design artifacts. Treat target runtime behavior as a prediction; apply the
model's artifact criteria instead of requiring that behavior to exist now.

## Orient

1. **Read the brief end to end**, including **Non-goals**. Note its **Shape** —
   the shape determines what counts as a demonstration.
2. **Read the design sections under Traces to.** They are the authority the
   brief is derived from; where the brief and the design set disagree, that is a
   finding, not a choice.
3. **Check the dependencies.** Every ID under **Dependencies** is `completed` in
   `tasks.json`. A `completed` status is a claim: if the type, contract, or
   behavior the edge names is missing at HEAD, the dependency is not done. Stop
   and report that rather than building the missing piece inside this task.
4. **Read the Write set and Contract surfaces.** The write set forecasts which
   files the change touches, so tasks in flight can predict a rebase; it is not
   a boundary. The contract surfaces are what only this task may redefine.
5. **Run the demonstration commands as they stand.** Seeing the current behavior
   before the change is what makes the change observable afterwards.

## Refine the brief

A brief is written before the code it builds on exists, so much of it is a
prediction. Refine it when the task is dispatched, before any implementation.
By then the code is at HEAD, and this step has context the planner did not.

Check, with evidence — a command with its observed output, or a `file:line`:

1. **Assumptions about the existing code** — the type, contract, or behavior
   each dependency edge names exists at HEAD. So do the interfaces, files, and
   behavior the Scope relies on without an edge, in the form the brief assumes.
2. **Deferrals into this task** — search every brief's Deferrals table, the phase
   coverage records, and the tree for this task's ID. Check each hit against this
   task's increment. Include its owned work or report
   an ownership conflict to the phase/task authors.
3. **Objective against criteria** — every guarantee the Objective states ends
   one of three ways. Acceptance criteria name the failures that would break it
   and what a person observes in each. Or it is narrowed out of the Objective.
   Or report a boundary change for separate phase revision and task derivation.
4. **Phase fit** — the task fits its decided increment and has no unresolved
   rederivation marker. Report a legacy or bundled brief to the task author;
   refinement does not substitute for phase revision or task derivation.
5. **Demonstration** — the tools and existing commands it uses are present where
   it will run, and each expectation can be observed without false matches.
   It runs where the project's deployment and test instructions say a
   demonstration runs, and leaves deployments it does not own alone.

Then make the brief concrete, in the sections it already has:

- **Scope** — implementation instructions written against the code at HEAD:
  where the code lives, how it connects to what exists, and the order where
  order matters, with a short reason for each judgment call. Keep these
  instructions inside the decided increment. A needed boundary or
  ownership change goes to the phase author and task author before refinement
  resumes; do not ask implementation to invent another task.
- **Non-goals** — what the refinement explicitly excludes.
- **Write set** — the forecast, updated from the concrete Scope.
- **Acceptance criteria**, **Verification**, and **Demonstration** — the
  results of the criteria and demonstration checks, and the test that will show
  each criterion. Verification lists only checks CI does not run; the repository
  gates are CI's.
- **Deferrals** — rows only for owners that already exist in `tasks.json`. Have
  a separate task author derive a missing owner before adding its row.

Refine toward what the design set and the code at HEAD require, never toward an
implementation that already exists.

The artifact changes only the brief, and is committed to the task's branch
before any implementation. Implementation starts from it.

Treat the fake as malleable. Change proposed fixture behavior rather than
expanding product scope to fit its current flags, protocol, or startup timing.
Label predicted results for new code separately from observations at HEAD.

## Done

Change only this brief, commit it on the task branch before implementation, and
hand it to a different implementer. Preserve tracker statuses. Report unresolved
phase boundaries or missing prerequisite behavior rather than absorbing them.
