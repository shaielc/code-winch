---
name: close-workplan
description: Close an implementation workplan into docs/state.md and a supported empty tracker. Use only when explicitly asked to end a plan and preserve the system record and lessons.
---

# Close workplan

Read applicable `AGENTS.md` files and [the model](../shared/workplan-model.md).
Run this operation only for an explicit request to end the plan.

## Procedure

Close ends a plan. It reads what happened, writes the record to `docs/state.md`,
and strips `docs/workplan/` back to an empty tracker. It writes no briefs: the
next plan is created against the design set and this record.

Close happens because someone asked for it, never as a consequence of an audit
finding the plan in bad shape. An audit that finds problems reports them; only a
decision to end the plan runs this.

1. **Settle every task.** Status, brief, and whether the acceptance criteria
   hold at HEAD. A `completed` status is a claim about the repository — check it
   against the repository, not against the pull request that closed it. Sample
   nothing; the point of a close is that the whole tracker gets resolved.
2. **Read every post-mortem.** Each one names a plan rule that would have caught
   its defect, and those rules are the most valuable thing the plan produced.
   They must be restated in the report, because `post-mortems/` is about to go.
   Read propagation ledgers too; carry unresolved accepted findings and their
   task/issue destinations into the state report before removing those ledgers.
3. **Read the code that exists**, not the briefs' account of it. What runs, what
   is registered in the composition root, what the standing suite actually
   covers, what a person can reach by hand.
4. **Check the invariants at HEAD.** I1–I7 against the real repository. Start
   the system. Run the fake profile by hand. Run the standing scenario suite.
5. **Find what no task claimed.** Stubs, TODOs, unimplemented branches, and
   unreachable error paths whether or not a brief named them; promises in the
   design set that never became a task at all.
6. **Write `docs/state.md`.** See below.
7. **Empty the plan in the same commit** that adds the report, so the two replace
   each other in one step. Git keeps the old plan; the working tree should not go
   on presenting a plan nobody is executing. Remove the briefs, the phase
   directories, `README.md`, `post-mortems/`, and `propagations/`. Keep two files:

   - `docs/workplan/tasks.schema.json`, unchanged. It is the definition of the ID
     scheme and the tracker's shape, and it is what the next plan is written
     against. A close that deletes it deletes the format its own report cites.
   - `docs/workplan/tasks.json`, emptied to the smallest body its own schema
     accepts — `schema_version`, `status_values`, and an empty `tasks` array, all
     three of which `tasks.schema.json` requires:

         {
           "schema_version": 1,
           "status_values": ["pending", "in_progress", "blocked", "completed"],
           "tasks": []
         }

     The dispatch automation opens this file unconditionally and validates it, so
     an absent or shapeless tracker fails every pull request rather than
     reporting no available tasks. Validate it before committing the close; an
     example that omits a required key is how this went wrong once already.

   The empty tracker is the closed state, not a stub to fill in. `create-phase`
   designs the next phase; a separate `derive-task` author writes its tasks.
8. **Sanitize the references.** See below.

### Sanitizing the references

The plan is cited from outside itself: by the automation that dispatches it, by
the contributor documentation that explains how to pick a task up, by code
comments that name an owning task, and by deferrals that name the ID meant to
resolve them. Once the briefs are gone every one of those is dangling — nothing
can resolve a task ID any more, and nothing can open a brief path. A close that
leaves them behind hands the next plan a tree that describes a plan that no
longer exists.

Sweep the whole tree, not only `docs/`. Two searches find nearly all of it: the
ID pattern, and the string `docs/workplan/`. Read the surrounding sentence in
each hit rather than pattern-replacing, and check at least contributor
documentation (`AGENTS.md`, `README.md`), CI workflows, dispatch scripts and
their prompts, decision records, and comments in source.

Each hit resolves one of three ways.

- **It carried something.** A deferral naming an owning task, a code comment
  naming the task that removes a stub, a workflow-slot comment naming who
  allocated it. What the reference was holding is a gap in the system, so it
  belongs in *What is not implemented* under a description of the gap. Then the
  reference goes.
- **It points at the plan as a place.** A brief path, a link into
  `docs/workplan/README.md`, a "start here" that routes a contributor to the
  tracker. Repoint it at `docs/state.md` or delete it, whichever leaves a true
  sentence.
- **It is machinery.** Scripts, workflows, and control panels that read the
  tracker keep working against the empty one and report nothing available; that
  is the whole reason the tracker survives. Leave them alone, and confirm the
  claim rather than assuming it — run the availability query and the status gate
  against the emptied tracker and see them succeed with an empty result.

The check is that no task ID and no `docs/workplan/` brief path survives outside
git history, and that every document describing how to pick up work either
describes the empty state truthfully or is gone.

### The report

`docs/state.md` is a statement about the system, not a plan. It has three parts,
and someone writing the next plan should need nothing else from the one that
closed.

**What was done.** The capabilities that exist and are reachable, each with the
evidence: the command that drives it and the profile it runs under. Group by
capability rather than by task ID — the next plan does not inherit this one's
decomposition, and organising the record around it smuggles that decomposition
forward.

**What went wrong.** Every post-mortem's rule, restated so it survives the
briefs being removed. Tasks marked complete that were not. Seams that leaked.
Invariants that were established and then lost, and what lost them.

**What is not implemented.** Every design-set promise with no working code
behind it, every stub and unreachable branch still in the tree, and every
deferral whose owning task never landed. State each as a gap in the system
rather than as a task: naming the work is the next plan's job, and a state
document that reads like a backlog will be treated as one.

Every claim carries a command with its output or a `file:line`. A close report
whose claims cannot be rechecked is worth less than the tracker it replaced.

