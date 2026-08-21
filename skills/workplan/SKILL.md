---
name: workplan
description: Write and review the implementation workplan in docs/workplan/ — its task briefs, its dependency graph, and its tracker. Use when deriving a plan from the design set, adding tasks to a plan already in flight, correcting status or ownership, widening a plan that has become serial, auditing the plan for coverage gaps, unjustified dependency edges, and unreachable code, or closing a finished plan into a statement of what the system now is. Assumes a design-document set exists (at minimum docs/architecture.md and docs/decisions/).
---

# Workplan

This skill covers the plan as an artifact: the briefs, the tracker, and the
index that ties them together.

Read `skills/shared/workplan-model.md` first. It defines the layout, the four
properties, the seven invariants, the four task shapes, dependency-edge reasons,
write sets and contract surfaces, brief anatomy, and the frozen tracker schema.
Everything below is how to apply that model when the plan changes or is judged.

## Modes

Establish which mode applies before doing anything else.

1. **Create** — no plan exists. Read the full design set, derive phases, write
   every brief and the tracker. Where a previous plan was closed, read
   `docs/state.md` first: it says what already exists, what broke last time, and
   what is still missing. Creating against the design set alone re-plans work
   that is already done. See *Creating a plan over existing code*.
2. **Extend** — add tasks to a plan in flight. Append IDs; never renumber. This
   is where revision tasks arrive: a defect found while implementing, a failed
   task audit, a write collision that needed more than a rebase.
3. **Update** — status, ownership, or blocked reason changed. See *Keeping
   status true*.
4. **Audit** — report coverage gaps, invariant violations, and unreachable code
   without changing the plan.
5. **Close** — the plan is finished or is being abandoned. Record what the
   system actually is in `docs/state.md` and remove `docs/workplan/`. See *Close
   procedure*.

There is no re-derivation mode. A plan is not rewritten against the code it
produced; it is closed, and the next one is created against the design set and
the record the close leaves behind. Rewriting a plan in place preserves its
decomposition long after the decomposition is what went wrong.

Whichever mode applies, check the seven invariants explicitly. They hold of the
plan as a whole, so a plan can violate one while every individual brief reads
well.

## Writing briefs

Briefs are drafts until their phase opens. Writing all of them up front is
correct — it is what makes the whole-system map useful — but a brief written
before its phase's dependencies exist is a hypothesis, and contact with the code
is where it gets tested against reality.

Every brief follows the anatomy in the shared model. The three sections that
carry the most weight, and that a draft most often leaves thin:

- **Runtime reachability** — never empty. A task that cannot name the
  composition root, profile, and command that reach its code fits none of the
  four shapes; merge it into the task that first reaches it.
- **Write set and Contract surfaces** — the two together are what makes
  concurrency derivable, and they are not interchangeable: an overlapping write
  set is a merge warning, an overlapping contract surface is a missing
  dependency edge. A brief that declares neither is asserting width instead of
  showing it.
- **Demonstration** — commands a person can actually run, and what they should
  see. Not a formality, and not a restatement of the test names.

Deferrals name an ID that exists in `tasks.json` at the moment the deferral is
written. Writing a deferral therefore sometimes means creating the owning task
first.

### The split test

When a proposed task looks like it should be two, inspect the dependency between
the halves.

The split is invalid if the second half's objective is substantially *make the
capability the first half introduces usable through the maintained operator
surface*. That is the first half's completion condition wearing a task ID. Merge
the operator path back in.

The split is valid when the second half introduces a distinct observable
property: a new capability, a new substrate, failure or recovery guarantees,
performance or security guarantees, a revision of an implementation that already
works, or operator ergonomics beyond what is needed to drive the existing
capability.

### The overstuffed-seam test

After drafting a seam brief, classify every scope bullet as one of: required
behavior, required wiring, required operator reachability, required invariant
preservation, future enablement, hardening, or refactor.

A seam normally contains only the first four. Extract the rest unless removing
it would make the objective false or break an invariant at completion.

## Creating a plan over existing code

A plan is rarely written against an empty repository. After a close, and any
time one is derived for a system that already runs, the code arrives with some
invariants held and others not.

Check I1–I7 against that code before deriving phases. Start the system. Run the
fake profile by hand. Run the standing scenario suite, or find that there is not
one. `docs/state.md` says what the last plan believed; the repository says what
is true.

**Every invariant that fails becomes a task in the new plan.** Not a note in the
README, not a caveat inside some brief, not an assumption that a task will get
to it on the way past — a task with an ID, in a phase, with acceptance criteria
that prove the invariant holds again. A plan that inherits a system which does
not deploy, and does not contain the task that makes it deploy, is built on a
claim it never checked.

I1 and I2 failures go in the first phase. Every other task's demonstration
assumes a system that starts and deploys, so a plan that schedules those repairs
late has no demonstrable task before them.

A repair brief fills the ordinary anatomy, with three sections carrying the
weight:

- **Objective** — the invariant restored, stated as observable behavior.
- **Acceptance criteria** — checkable against the repository, including the
  invariant's own evidence: the command and the output that show it holding.
- **Traces to** — the invariant, and the section of `docs/state.md` that
  recorded the gap.

Repair tasks take the ordinary four shapes. The shape follows from what becomes
observable, not from the fact that something is being restored: a repair that
makes the system deploy again is demonstrated the way the first task of a
greenfield plan would demonstrate it.

The gaps `docs/state.md` records under *what is not implemented* are a different
input and get ordinary treatment. Those are missing capabilities, derived from
the design set like any other. An invariant failure is not a missing capability
— it is the plan's own floor giving way, which is why it goes first.

## Width is designed, not hoped for

Concurrency is derived from dependency edges, write sets, and contract surfaces,
in the same way availability is derived from status. Those declarations are
defined in the shared model; what follows is how to shape a plan so they produce
width.

### Design the spine so it does not collide

I1 and I4 create a shared spine — composition root, scenario suite, CLI — that
every task has reason to touch. Left alone, that spine serializes the plan. The
remedy is architectural and belongs in the design documents, not only in the
plan:

- wiring is **append-only** — implementations self-register from their own
  files and the composition root reads a registry, rather than every task
  editing one central switch;
- profiles, scenarios, and CLI commands are **one file each**;
- migration identifiers are allocated so two open branches cannot claim the
  same one.

Where the architecture cannot offer this, say so in the brief and accept the
serialization knowingly.

### Shape and width

Swap tasks are naturally parallel: different substrates, disjoint surfaces.
Seam tasks contend for the composition root, so define seams early and in as few
tasks as the boundaries allow. Capability tasks contend for the scenario suite.
A phase that opens with two or three independent roots stays wide; a phase that
opens with one is a chain whatever its task count.

### Report it

Whenever the plan is created, extended, or audited, compute and record per
phase: the **critical path** (longest
dependency chain), the **average width** (task count over critical path length),
the **write-collision set** (pairs of concurrently-available tasks with
overlapping write sets), and the **contract-collision set** (pairs sharing a
contract surface).

Report the two collision sets separately, because they mean different things. A
write collision is a cost to acknowledge — those two will rebase. A contract
collision between concurrently-available tasks is a defect to fix before the
phase opens: the edge is missing, or one task's surface needs narrowing, or the
two are one task.

Report the numbers and justify any chain that is long relative to its phase; do
not adopt a target width in the abstract.

## Close procedure

Close ends a plan. It reads what happened, writes the record to `docs/state.md`,
and removes `docs/workplan/`. It writes no briefs and no tracker: the next plan
is created against the design set and this record.

Close happens because someone asked for it, never as a consequence of an audit
finding the plan in bad shape. An audit that finds problems reports them; only a
decision to end the plan runs this.

1. **Settle every task.** Status, brief, and whether the acceptance criteria
   hold at HEAD. A `completed` status is a claim about the repository — check it
   against the repository, not against the pull request that closed it. Sample
   nothing; the point of a close is that the whole tracker gets resolved.
2. **Read every post-mortem.** Each one names a plan rule that would have caught
   its defect, and those rules are the most valuable thing the plan produced.
   They must be restated in the report, because the directory is about to go.
3. **Read the code that exists**, not the briefs' account of it. What runs, what
   is registered in the composition root, what the standing suite actually
   covers, what a person can reach by hand.
4. **Check the invariants at HEAD.** I1–I7 against the real repository. Start
   the system. Run the fake profile by hand. Run the standing scenario suite.
5. **Find what no task claimed.** Stubs, TODOs, unimplemented branches, and
   unreachable error paths whether or not a brief named them; promises in the
   design set that never became a task at all.
6. **Write `docs/state.md`.** See below.
7. **Remove `docs/workplan/` in the same commit** that adds the report, so the
   two replace each other in one step. Git keeps the old plan; the working tree
   should not go on presenting a plan nobody is executing.

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
directory being removed. Tasks marked complete that were not. Seams that leaked.
Invariants that were established and then lost, and what lost them.

**What is not implemented.** Every design-set promise with no working code
behind it, every stub and unreachable branch still in the tree, and every
deferral whose owning task never landed. State each as a gap in the system
rather than as a task: naming the work is the next plan's job, and a state
document that reads like a backlog will be treated as one.

Every claim carries a command with its output or a `file:line`. A close report
whose claims cannot be rechecked is worth less than the tracker it replaced.

## Keeping status true

Statuses are claims about the repository, and claims decay.

- Mark `completed` only after verifying the acceptance criteria against HEAD,
  not against the pull request description.
- In update mode, re-verify any task whose completion is a dependency of the
  work being planned.
- In audit mode, sample completed tasks and report any whose criteria do not
  hold.
- `blocked` requires a reason naming what would unblock it, ideally a task ID.

## Audit mode

Report, without changing the plan:

- **Coverage** — every section of the design set maps to at least one task;
  every task traces back to a documented requirement. Name the gaps and the
  scope inflation separately.
- **Reachability** — every port, interface, or boundary in the codebase has at
  least one implementation registered in a composition root and reachable at
  runtime. List the orphans.
- **Invariants** — I1 through I7, each with the evidence checked.
- **Graph** — no unknown references, no forward references, no cycles. Every
  edge carries one of the four reasons; list the unjustified ones, since each is
  width the plan is giving away. List `revision` edges in phases that have not
  opened separately: a revision planned before the implementation exists is work
  scheduled to be done twice.
- **Width** — critical path, average width, and the write- and
  contract-collision sets per phase. A contract collision between two
  concurrently-available tasks is a blocking finding; a write collision is
  reported as a cost.
- **Completion closure** — a task whose main purpose is to expose through the
  maintained CLI a capability an earlier task introduced, and which depends
  directly on that task. Report the pair: this is the earlier task's completion
  condition filed under a separate ID. Report the mirror case too — scope
  bullets that exist mainly to prepare later tasks.
- **Round trips** — every user-facing interaction that is modelled or rendered
  is also actionable. A payload kind that can be displayed but not submitted, or
  a state transition the documentation promises but no API exposes, is a round
  trip that stops halfway.
- **Status** — sampled completed tasks that do not hold.

## Plan failures get post-mortems

When a defect's root cause is the plan rather than an implementation — a seam
between two briefs, an invariant with no owning task, an assumption one brief
made about another's output — record it in `docs/workplan/post-mortems/`. Each
record names the briefs involved, explains why no single one is wrong on its
face, and states what plan rule would have caught it. Ordinary implementation
bugs do not belong there.

Those rules outlive the plan that produced them. Close restates every one of
them in `docs/state.md` before the directory goes, which is the only reason
removing it is safe.

## Smells

Concrete patterns that have produced broken plans. Treat each as a defect to
fix, not a style preference.

- A port whose only implementations are test doubles.
- An empty composition root past the first task.
- A roadmap line — "minimal local authentication", "credential-reference
  storage" — that maps to no task. Every promise in the design set has an owner
  or an explicit deferral.
- Telemetry, configuration, or logging foundations scheduled as the final task,
  while every earlier brief asserts logging requirements. Retrofitting the whole
  plan's ad-hoc logging at the end is the expensive path; split a small
  foundation task into the first phase and leave dashboards and runbooks late.
- A capability that is modelled and rendered in one phase but only actionable in
  a much later one. Ship the round trip in one task.
- A task whose purpose is to make an earlier task's capability reachable by hand.
- Scope bullets that exist mainly to prepare later tasks: pre-emptive file
  restructuring, extension registries, permissive policy hooks, unrelated
  metrics, or failure-mode hardening that could be demonstrated on its own.
- Identical prose sections across every brief.
- A first phase whose tasks are all infrastructure and whose only user-visible
  result is at the end.
- An aggregate that many tasks assume as a precondition and no task builds.
- A phase that opens with one root task, so everything in it inherits one chain.
- A dependency edge whose only justification is ordering intuition.
- A central switch, registry, or specification file every task must edit to make
  its own work reachable.
- A plan created over a system that does not start, with no task that makes it
  start.

## Before finishing

In close mode, only the last item applies. In every other mode:

- Every brief has a non-empty **Runtime reachability** section and declares its
  **write set** and its **contract surfaces**.
- Every brief has a **Demonstration** with commands a person can actually run.
- Every dependency edge names compile, contract, semantic, or revision in one
  clause, and no `revision` edge was written before the implementation it
  revises exists.
- Every deferral names an ID that exists in `tasks.json`.
- `tasks.json` validates against the frozen schema; `README.md` tables agree
  with it; every `brief` path resolves.
- The dependency graph has no unknown references, forward references, or cycles.
- Critical path, average width, and both collision sets are recorded per phase.
- No two concurrently-available tasks share a contract surface.
- The first task produces a system that starts and deploys.
- Every invariant failing in the code the plan was created over has a task that
  restores it, and any I1 or I2 repair is in the first phase.
- No task defers part of its own completion condition; every operator-visible
  capability is drivable by hand in the task that introduces it.
- Every task fits one of the four shapes.
- A close left `docs/state.md` in place of `docs/workplan/`: every post-mortem
  rule restated, every claim carrying a command or a `file:line`, no section
  written as a backlog, and both changes in one commit.
