# Phase design artifact

Write `docs/workplan/phase-N/README.md` before deriving its task briefs. Use this
reference with the derivation loop in `skills/workplan/SKILL.md`.

The artifact selects a coarse task partition. It specifies increments and
boundary decisions; it does not leave them as candidates for a refiner to
choose. Individual tasks may divide an increment further. Keep exact commands,
test fixtures, transport implementations, and file forecasts in the derived
briefs unless a documented product constraint requires a phase-level choice.

## Format

Use the following sections. Keep the phase short enough to inspect as a whole.

### Outcome and starting point

State what a person can do at phase exit and why that outcome is needed. Cite
its architecture/roadmap sources and existing product constraints. Record the
current capabilities that can be reused, with source or command evidence and
runtime-verification limits. Separate established behavior from missing work.

### Concern boundaries

List included concerns, the increments that own them, and excluded concerns
with their roadmap stage or triggered decision owner. Distinguish minimum
resource cleanup from lifecycle reporting, stop controls, and recovery.
Explain why concerns grouped into an increment cannot sensibly arrive in
separate observable changes. An objective that already includes them is not
that explanation.

### Increments

Give each increment a stable phase-local label, such as `output-live`, and:

| Field | Decision it records |
|---|---|
| Outcome | The before/after behavior a person observes |
| Includes | The responsibility this increment adds and its minimal operator path |
| Excludes | Neighboring responsibilities explicitly owned elsewhere |
| Prerequisites | Existing behavior or earlier increments it needs |

Use one row or short subsection per increment. The result should approximate
task splitting: one increment is usually one task, and task derivation may find
smaller closed loops within it. Do not combine several increments in one task
merely because they share code. If that proves necessary, revise and justify the
phase boundary before deriving the combined task.

State the selected boundaries, for example: ownership is observable through
status; live output follows without persistence; durable history follows without
seamless reconnect. This is a decision, not advice to make that decision later.

### Task derivation record

Map each increment to the task IDs derived so far, or mark it **not yet
derived**. For a phase replacing an existing partition, identify legacy briefs
that require rederivation separately from completed work being reused. Those
IDs document existing coverage; they do not certify conformity to the new
partition. Phase-local labels are not valid deferral owners in task briefs.

Coverage may be partial while the derivation loop runs. Record which increment
comes next, and record a phase revision with the changed boundary and affected
task mappings. Do not add speculative tasks to the tracker before their briefs
have been derived. Keep executable dependencies and statuses in `tasks.json`.

### Phase exit

Describe the end-to-end behavior proving the selected phase outcome and the
properties it includes. Give observations at phase level; derive the commands
and fixtures task by task. Do not import lifecycle tests into an output phase,
or require startup-transcript replay to demonstrate live output.

## Before using the phase to derive a task

- The increments and boundaries are decided, not a list of alternatives.
- Included and excluded concerns have destinations.
- The next task's increment and prerequisites are identifiable.
- The baseline distinguishes completed work from missing behavior.
- Phase constraints are supported by the product requirement and its stage.
- Existing fixtures can change to exercise the chosen behavior.

An architectural decision that truly remains open belongs in the roadmap with
a trigger. Do not call an unspecified increment boundary a deferred
architectural decision.
