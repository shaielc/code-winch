---
name: workplan
description: Derive an implementation workplan from the architecture and roadmap by designing a phase, deriving one task at a time, and revising the phase when derivation exposes a better boundary. Use to create or extend docs/workplan/, revise phase designs or task ownership, audit decomposition and dependencies, update plan records, or close a plan into docs/state.md.
---

# Workplan

Own the phase designs, task briefs, tracker, and plan index. Read
`skills/shared/workplan-model.md` for the invariants, task shapes, brief anatomy,
dependency reasons, and frozen tracker schema. Read applicable `AGENTS.md` files.
Use [Phase design](references/phase-design.md) for the phase artifact and
[Closing a workplan](references/close.md) when closing a plan.

## Choose the operation

- **Create** — establish the next phase from the design set and current system,
  then derive its tasks through the loop below.
- **Extend** — incorporate a new requirement or discovery into an existing
  phase, revise its increments if needed, then derive the affected tasks.
- **Update** — reconcile recorded ownership, dependency evidence, or write-set
  forecasts with agreed changes. Keep status changes under repository automation.
- **Audit** — report discrepancies with evidence; do not change the artifacts
  being judged.
- **Close** — when explicitly requested, settle the plan and replace it with
  `docs/state.md` using the close reference. An audit never closes a plan itself.

## Derivation loop

Follow this order:

1. Read the architecture, contracts, decisions, and roadmap for the phase.
2. Establish what already works in the repository.
3. Create the phase design with explicit increments and boundary decisions.
4. Derive one specific task from one increment.
5. If that derivation exposes a phase-level issue, revise the phase design and
   reconcile affected derived tasks.
6. Derive the next task. Repeat until the phase's increments are covered.

The phase is the map before detailed task derivation. It approximates task
splitting; it is not a collection of possible scopes or a summary written after
all briefs exist. An increment may split further into several observable tasks.
Writing one task does not require implementing it before deriving another.
Future phases remain phase-level designs until their dependencies justify more
detail; do not write every future brief in advance.

### Establish the starting point

Read `docs/state.md`, completed briefs, and the code they describe. Treat the
state report as a dated input and completed statuses as claims. Check the exact
behavior each planned change relies on. Run existing entrypoints and gates where
the environment allows; otherwise state the source evidence and runtime limits.

Separate existing behavior, required revisions, and missing behavior. A shipped
executable is not a running child; graceful server shutdown is not child
supervision. Reuse completed work rather than assigning it again.

Check I1–I7 against that starting point. Allocate explicit repair work for a
failed invariant, with I1/I2 repairs first. A capability that has not yet been
introduced is missing work, not an invariant failure that requires building all
of it immediately.

### Create the phase

Write `docs/workplan/phase-N/README.md` using the phase reference. Decide:

- the observable outcome and the minimum properties the phase needs;
- the concerns included and the concerns assigned to another roadmap stage;
- the increments, their before/after observations, and prerequisites;
- the boundary of each increment: included work and explicitly excluded work;
- why any two concerns must arrive together.

Challenge upstream requirements before accepting scope. "Required by this
objective" is circular when the objective itself bundles several concerns.
"Required by the contract" requires checking the contract's intended stage;
correct a contradictory or premature contract in the same change. Shared files,
an existing fixture, or convenience are not boundary justifications.

Choose the smallest closed loops that reach the phase outcome. Do not infer a
stop API, an exit-outcome taxonomy, or a lifecycle event model merely from owning
a child process. Safe child cleanup is required wiring; the other concerns need
their own product reason and stage.

Make these decisions in the phase artifact. A list of concerns followed by
"decide during task refinement" has not designed the phase.

### Derive one task

Select the next uncovered part of an increment whose prerequisites are known.
Inspect the implementation and design surfaces needed to derive it, then
write the brief using the shared anatomy.

1. State one observable change from the current or prerequisite-produced state.
   Remove already completed work from its scope.
2. Apply the split test below. The task must fit inside its increment; it may
   cover less than the increment when a smaller useful loop exists.
3. Declare required behavior and wiring. Distinguish product constraints from
   proposed implementation approaches. A filename forecast or fixture protocol
   does not make an approach binding.
4. Provide the minimum maintained operator path in the same task. Add the
   product UI behavior the task promises, with the substrate and wiring it needs.
5. Define acceptance criteria for that change. Select verification that actually
   exercises it, using the current system and an adaptable fake.
6. Allocate the next plan-wide task ID, declare its surfaces and dependency
   reasons, and add the brief, tracker entry, index row, and increment mapping
   together. Preserve completed IDs and never edit status fields by hand.

Keep the fake's product-visible behavior supported. Its invented dialect,
flags, record fields, and signal handling may change with its consumers,
scenarios, and documentation. Require compatibility only for a documented
consumer. Do not derive a record model from the fake or force production code to
adapt to a fixture detail without a product reason; contracts §9 governs this.
Expose only the fake controls needed for the selected behavior or already used
by a supported product path. A fixture's standalone flags are not a checklist
of new sandbox configuration keys. Preserve their existing consumers when
changing the fake; do not forward every flag into the product merely because
the executable already has it.

Give demonstrations a reliable observation path. Inspect existing mechanisms,
exercise what is available, and identify predictions for new code. An early
output task can observe newly produced output after attachment; it need not gain
transcript playback, startup replay, or durable history just to make its test
convenient. Keep exact commands and test setup in the task brief.

A brief cannot defer to an increment label: a deferral owner must be a real task
in `tasks.json`, or a roadmap deferred decision with a trigger. If required
completion work has no derived owner, include it in this task or derive its
owner before recording the deferral. Do not create placeholder tracker entries
or skeletal briefs just to provide an ID.

### Revise the phase when needed

Revise the phase when deriving a task shows a missing increment, redundant work,
an incompatible contract, an unjustified bundle, or a prerequisite that changes
the partition. A different filename or local implementation choice alone does
not require a phase revision.

Record the changed increment and boundary decision in the phase README. Update
affected pending briefs, their mappings, dependencies, index, and collision
report in the same plan change. Preserve completed work and IDs; newly found
rework of completed implementation becomes a revision task with evidence.
Then continue task derivation from the revised phase.

When revising an already-open phase, identify which legacy briefs still need
rederivation. Do not present those briefs as conforming to the new partition.
Reconcile a task before assigning it under the revised phase. An unchanged
tracker status is not evidence that this reconciliation happened.

## Decide task boundaries

Split when the next task adds a distinct observable property: ownership,
delivery, persistence, an input action, recovery, performance, or security.
Each part must close its own loop at its own depth. Keep the basic maintained
operator path with the behavior it exposes; a second task whose only purpose is
to make the first task operable is an invalid split.

For each scope item, ask whether it is needed for the selected change, its
operator path, or invariant preservation. Extract future enablement, standalone
hardening, and refactoring. If several concerns appear indispensable because
the objective already combines them, revisit the objective and the phase
partition first. Do not use scope classification to rubber-stamp a bundle.

## Derive and report concurrency

Dependencies name compile, contract, semantic, or revision in one clause.
Shared files create write-collision warnings, not dependency edges. Shared
contract meaning requires ownership and an edge; if both tasks define it, narrow
their scopes or combine the relevant work. Revision edges describe discoveries
about existing implementation, never planned rewrites of code yet to be built.

Avoid unnecessary central edits when the existing architecture permits it, but
do not build a registry solely to enable later tasks. A claimed one-file command
still needs registration in the actual composition root; include that wiring in
the forecast and collision report.

After each task derivation or phase revision, report the graph currently derived:
critical path, average width, concurrent write collisions, and concurrent
contract collisions. Label partial coverage explicitly. Do not claim an
underived phase is wide because its unnamed increments look independent.
At complete phase derivation, every included increment has real task coverage,
the graph is acyclic, and no concurrently available tasks redefine one contract.

## Audit

Judge the architecture-to-phase-to-task chain against HEAD. Report, with commands
and observed output or `file:line` evidence:

- phase outcome, explicit increments, boundaries, and excluded concern ownership;
- completed work incorrectly planned again, or assumptions about a prerequisite;
- task scope that crosses its increment or combines independently observable work;
- missing phase coverage, and tasks implementing requirements outside the phase;
- fake interfaces elevated into product requirements without justification;
- runtime reachability, invariant preservation, and demonstration credibility;
- unknown dependency references, cycles, unjustified edges, and surface collisions;
- agreement among increment mappings, briefs, index, and tracker;
- completed statuses whose acceptance criteria no longer hold.

Report incomplete task derivation separately from an unresolved phase decision.
The former means a decided increment needs a brief; the latter means the phase
has not yet selected a boundary. An audit reports; it does not silently fix,
re-own, or close the plan it judges.

Record a plan failure in `docs/workplan/post-mortems/` when its cause is a seam
between briefs, an unowned invariant, or a faulty derivation assumption. Name
the involved briefs and the rule that would have caught it. Close preserves
those rules in `docs/state.md`.

## Before finishing

- Every designed phase has an explicit partition, with outcomes, boundaries,
  prerequisites, and concern ownership; detail follows the phase reference.
- Every derived task maps to an increment and fits a shared task shape.
- Each new task leaves a runnable, deployable system and its basic operator path.
- Every brief has the shared anatomy, evidenced prerequisites, credible
  demonstrations, declared write sets, and contract surfaces.
- Each deferral has a valid task or triggered decision owner.
- Tracker/schema conventions remain intact; IDs are append-only and status
  fields were not edited by the planning change.
- Index, tracker, briefs, increment mapping, and current collision report agree.
- Partial derivation and legacy briefs awaiting reconciliation are stated
  explicitly; completion is not claimed for uncovered increments.
- Contract or roadmap changes required by the selected boundaries accompany the
  phase change.
- Close follows the close reference, leaves a supported empty tracker, and
  records the system rather than another backlog.
