---
name: workplan
description: Create, extend, re-derive at a phase gate, update, or audit an implementation workplan — a directory of task briefs plus a machine-readable tracker — derived from a project's design documents. Use when asked to plan implementation work across phases, add tasks to a plan already in flight, replan the next phase against the code that now exists, correct task status, widen a plan that has become serial, or check the plan as a whole for coverage gaps, unjustified dependency edges, and unreachable code. Assumes a design-document set exists (at minimum docs/architecture.md and docs/decisions/). To implement one task, or to judge whether one task is done, use the task skill instead.
---

# Workplan

The subject is the plan: its phases, its briefs, its dependency graph, its
tracker. You are answerable for coverage of the design set, for width, and for
every claim in the plan being checkable against the repository.

Read [`skills/workplan-model.md`](../workplan-model.md) first. It defines what a
workplan is, the four properties, the seven invariants, the four task shapes,
the vocabulary, and the tracker's frozen shape. Everything below assumes it.

Given work on one task — implement this brief, is this task done, review the
change that claims to satisfy it — use
[`skills/task/SKILL.md`](../task/SKILL.md) instead.

## Modes

Establish which mode applies before doing anything else.

1. **Create** — no plan exists. Read the full design set, derive phases, write
   every brief and the tracker.
2. **Extend** — add tasks to a plan in flight. Append IDs; never renumber.
3. **Re-derive** — a phase gate just closed. Rewrite the next phase's briefs
   against the code that now exists. See *Gate procedure*.
4. **Update** — status, ownership, or blocked reason changed. See *Status
   truthfulness*.
5. **Audit** — report on the plan as a whole, changing nothing. See *Audit
   mode*.

Resolve conventions before writing: if a plan already exists, match its ID
scheme, directory layout, and brief structure. Otherwise use the ones here.

Briefs are drafts until their phase opens. Writing all of them up front is
correct — it is what makes the whole-system map useful — but a brief written
before its phase's dependencies exist is a hypothesis, and the gate is where it
gets tested against reality.

Every mode checks the seven invariants explicitly, because a plan can violate
one while every individual brief reads well.

## Parallelism

Concurrency is derived from two declarations, in the same way availability is
derived from status rather than stored as a flag.

### Dependency edges carry reasons

Record the reason — compile, contract, or semantic — beside each ID in the brief
header. An edge whose reason cannot be written in one clause is not a real edge.

Drop edges resting on "it logically comes after" (ordering intuition), "same
subsystem" (proximity), or "easier to review together" (split the pull requests
instead). Review effort belongs to pull requests; it is never a reason to move
scope between briefs, in either direction.

One more illegitimate reason deserves its own paragraph, being the most common
and the most expensive: **"it is nicer to develop against the real thing."** This
is what I3 exists for. A controllable fake profile dissolves these edges and is
the single largest source of width in a layered plan — an adapter that depends
on a real driver because the driver is more convenient to test against should
depend on the fake instead, and be swapped later. Written as a semantic edge,
this fallacy passes review; check it explicitly.

### Tasks declare owned surfaces

Each brief lists the files, directories, and contracts it writes. Two tasks with
disjoint owned surfaces may run concurrently whatever phase they sit in; two
tasks with overlapping surfaces are serialized in practice regardless of what
the graph says.

Surfaces that attract every task at once, and so deserve attention early: the
composition root, the standing scenario suite, ordered migration numbers, a
single API specification file, and any central registry, switch, or factory.

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
and the **collision set** (pairs of concurrently-available tasks with
overlapping owned surfaces). Report the numbers and justify any chain that is
long relative to its phase; do not adopt a target width in the abstract.

## Brief template

Keep briefs short. If a sentence would be identical across every brief, it is
boilerplate: put it in `README.md` once and delete it from the briefs. Identical
"architectural context" and "deliverables" sections across a whole plan are a
reliable sign that up-front depth was padding.

```markdown
# <ID>: <Title>

**Phase:** <N> — <phase name>
**Shape:** seam | swap | capability | hardening
**Dependencies:** <ID> (compile | contract | semantic: <one clause>), … or None

## Objective

One sentence, stated as an observable change to the running system.

## Scope

- Bullets. What this task builds.

## Non-goals

- Bullets. What it deliberately does not build.

## Runtime reachability

Which composition root, profile, and command reach this code once the task is
done. Never empty — a task that cannot fill this in fits none of the four
shapes and should be merged into the one that reaches it.

## Owned surfaces

Files, directories, and contracts this task writes. Concurrency is derived from
this list, not asserted: a task with no overlap here can run alongside any other
available task.

## Demonstration

Exact commands and the expected observable result.

    $ <command>
    $ <command>
    → expect: <what a person sees>

## Verification

- Standing scenario suite passes against <profile/substrate>.
- <Focused automated tests specific to this task.>
- <Repository format, lint, unit-test, and build checks affected.>

## Acceptance criteria

- [ ] Observable, checkable statements. No restatement of the scope.
- [ ] Invariants that must still hold at completion.

## Deferrals

| Deferred | Owning task |
|---|---|
| <thing not done here> | <existing ID> |

Or: `None.`

## Traces to

`docs/<file>.md` §<section>, ADR-<nnnn>
```

Two checks before a brief is final:

- **Every criterion has a witness inside the task.** Compare the acceptance
  criteria against the Demonstration and Verification blocks line by line.
  Demonstrations get written against the Objective sentence and then criteria
  accrete above them. A criterion whose only evidence comes from a task that
  depends on this one cannot be judged, and the schedule is what is wrong.
- **The header names one shape, and the scope fits it.** Scope spanning several
  shapes means several tasks sharing a header.

## tasks.json rules

The tracker's schema is **frozen** — the fields are listed in
`workplan-model.md`, and nothing may be added, because dispatch automation and
availability queries read it. Consequences to respect:

- Everything this skill introduces — shape, reachability, demonstration,
  deferrals, traceability — lives in the **briefs**, not the tracker.
- Do not add a `ready` flag. Availability is derived and a stored flag goes
  stale.
- IDs are append-only. Merged pull requests, dispatch history, and cross-brief
  references all cite them; renumbering silently invalidates that record. A task
  inserted logically "between" two others still gets the next free number.
- The ID pattern permits phases 0–5 only. A seventh phase requires a schema
  change, which is out of bounds — fold the work into an existing phase or raise
  it as a decision.
- `brief` paths are `phase-N/<slug>.md` and must resolve to a real file.

Keep `README.md`'s dependency tables and `tasks.json` in agreement; they are two
views of one graph, and drift between them is a defect.

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
   dependency edge against the three reasons — an edge defensible on paper is
   often unnecessary once the code exists, and a missing one is usually obvious
   by then. Recompute critical path, width, and collisions for the phase.
6. **Record the diff and its reason** in a short changelog section of
   `README.md`. A brief that changed silently teaches nobody anything.

Re-scoping a task — splitting it along its shapes, moving a criterion, moving
its witness — is gate work, and this is the gate. It changes the phase's graph,
critical path, and width, which is why steps 5 and 6 exist: the recomputation
and the recorded reason.

Re-scopes reach you as reports from outside — a review found a brief too large,
an implementer found a criterion the task cannot witness. Re-derive the split
here rather than ratifying the one the reporter proposed. Review pressure
selects the most detachable scope, which is rarely the scope that should move.

## Status truthfulness

A status is a claim about the repository, and claims decay. Task statuses have
been found wrong in practice — marked complete when acceptance criteria were
not met.

- Mark `completed` only after verifying the acceptance criteria against HEAD,
  not against the pull request description.
- In update mode, re-verify any task whose completion is a dependency of the
  work being planned.
- In a plan audit, sample completed tasks and report any whose criteria do not
  hold.
- `blocked` requires a reason naming what would unblock it, ideally a task ID.

## Audit mode

The subject is the plan. Report, without changing it:

- **Coverage** — every section of the design set maps to at least one task;
  every task traces back to a documented requirement. Name the gaps and the
  scope inflation separately.
- **Reachability** — every port, interface, or boundary in the codebase has at
  least one implementation registered in a composition root and reachable at
  runtime. List the orphans.
- **Invariants** — I1 through I7, each with the evidence checked.
- **Graph** — no unknown references, no forward references, no cycles. Every
  edge carries one of the three reasons; list the unjustified ones, since each
  is width the plan is giving away.
- **Width** — critical path, average width, and collision set per phase, as in
  *Parallelism*.
- **Round trips** — every user-facing interaction that is modelled or rendered
  is also actionable. A payload kind that can be displayed but not submitted, or
  a state transition the documentation promises but no API exposes, is a round
  trip that stops halfway.
- **Witnesses** — every acceptance criterion is observable by a surface that
  exists no later than its own task. List criteria whose only witness is
  downstream.
- **Status** — sampled completed tasks that do not hold.

Findings here are attributed to the plan, or to the task that owns the gap where
one exists — never to whichever task happens to be in flight. A half-built
capability whose owner is scheduled later is not a finding at all; it is the
plan working.

Where the root cause is a seam between two briefs rather than one brief being
wrong, write it up in `docs/workplan/post-mortems/` following the convention
there.

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
- One brief carrying several shapes' work under a single shape header.
- A criterion whose only witness is a task that depends on the task claiming it.
- A hands-on surface scheduled after the API it exists to drive. Instruments
  come first.

## Before finishing

Applies to create, extend, and re-derive. An audit changes nothing, so it
finishes when its findings are attributed at the right scope instead.

- Every brief has a non-empty **Runtime reachability** section and declares its
  **owned surfaces**.
- Every brief has a **Demonstration** with commands a person can actually run.
- Every acceptance criterion has a witness no later than its own task.
- Every dependency edge names compile, contract, or semantic in one clause.
- Every deferral names an ID that exists in `tasks.json`.
- `tasks.json` validates against the frozen schema; `README.md` tables agree
  with it; every `brief` path resolves.
- The dependency graph has no unknown references, forward references, or cycles.
- Critical path, average width, and the collision set are recorded per phase.
- The first task produces a system that starts and deploys.
- Every task fits one of the four shapes, and only one.
