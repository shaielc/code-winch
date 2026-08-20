---
name: workplan
description: Write and review the implementation workplan in docs/workplan/ — its task briefs, its dependency graph, and its tracker. Use when deriving a plan from the design set, adding tasks to a plan already in flight, re-deriving the next phase at a gate, correcting status or ownership, widening a plan that has become serial, or auditing the plan for coverage gaps, unjustified dependency edges, and unreachable code. Assumes a design-document set exists (at minimum docs/architecture.md and docs/decisions/).
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
   every brief and the tracker.
2. **Extend** — add tasks to a plan in flight. Append IDs; never renumber. This
   is where revision tasks arrive: a defect found while implementing, a failed
   task audit, a write collision that needed more than a rebase.
3. **Re-derive** — a phase gate just closed. Rewrite the next phase's briefs
   against the code that now exists. See *Gate procedure*.
4. **Update** — status, ownership, or blocked reason changed. See *Keeping
   status true*.
5. **Audit** — report coverage gaps, invariant violations, and unreachable code
   without changing the plan.

Whichever mode applies, check the seven invariants explicitly. They hold of the
plan as a whole, so a plan can violate one while every individual brief reads
well.

## Writing briefs

Briefs are drafts until their phase opens. Writing all of them up front is
correct — it is what makes the whole-system map useful — but a brief written
before its phase's dependencies exist is a hypothesis, and the gate is where it
gets tested against reality.

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

At every gate, compute and record per phase: the **critical path** (longest
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

## Gate procedure

When a phase closes, before opening the next:

1. **Prove the exit statement.** The phase's exit condition from the roadmap is
   demonstrated by an automated end-to-end or integration scenario, not by a
   happy-path UI walkthrough.
2. **Check the invariants at HEAD.** I1–I7, against the real repository. Start
   the system. Run the fake profile by hand. Run the standing scenario suite.
3. **Audit the statuses.** See below.
4. **Re-read the code that now exists.** Not the design docs — the code. The
   next phase's briefs were written against a hypothesis about what this phase
   would produce.
5. **Rewrite, split, or drop the next phase's briefs** to match. Re-check every
   dependency edge against the four reasons — an edge defensible on paper is
   often unnecessary once the code exists, and a missing one is usually obvious
   by then. Recompute critical path, width, and collisions for the phase.
6. **Record the diff and its reason** in a short changelog section of
   `README.md`. A brief that changed silently teaches nobody anything.

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
- Identical prose sections across every brief.
- A first phase whose tasks are all infrastructure and whose only user-visible
  result is at the end.
- An aggregate that many tasks assume as a precondition and no task builds.
- A phase that opens with one root task, so everything in it inherits one chain.
- A dependency edge whose only justification is ordering intuition.
- A central switch, registry, or specification file every task must edit to make
  its own work reachable.

## Before finishing

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
- Every task fits one of the four shapes.
