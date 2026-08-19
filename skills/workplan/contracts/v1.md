---
name: workplan
description: Create, extend, re-derive at a phase gate, update, or audit an implementation workplan — a directory of task briefs plus a machine-readable tracker — derived from a project's design documents. Use when asked to plan implementation work across phases, add tasks to a plan already in flight, replan the next phase against the code that now exists, correct task status, widen a plan that has become serial, or check a plan for
coverage gaps and unreachable code. Assumes a design-document set exists (at minimum docs/architecture.md and docs/decisions/).
---

# Workplan

A workplan converts a design baseline into implementation-sized tasks that an
autonomous agent or a contributor can pick up one at a time, each in its own
pull request.

A good workplan holds four properties at once.

**It always describes a running system.** At every point in the plan's history
there is a configuration that starts, deploys, and can be driven by hand. Tasks
make that system more real; they do not accumulate parts to be assembled later.

**Every task changes observable behavior.** The change may be small — the same
scenario passing against a real database rather than an in-memory one — but it
is visible from outside the code, and the brief says how to see it.

**It is wide.** Many tasks are available at once, so several contributors or
dispatch workers proceed without waiting on each other. Width is a design
property, achieved deliberately or not at all.

**Every claim in it is checkable.** Status, completion, coverage, and
reachability are verifiable against the repository rather than asserted.

Faithful coverage of the design documents is necessary and not sufficient: a
plan can describe the architecture completely, pass its own contract tests, and
still produce a system nobody can start.

## Deliverables

| Path | Purpose |
|---|---|
| `docs/workplan/README.md` | Narrative index, per-phase dependency tables, phase exit rule, how to use the plan |
| `docs/workplan/phase-N/<id>-<slug>.md` | One brief per task |
| `docs/workplan/tasks.json` | Machine-readable tracker — status, ownership, dependencies |
| `docs/workplan/tasks.schema.json` | Schema for the tracker |

Resolve conventions before writing: if a plan already exists, match its ID
scheme, directory layout, and brief structure. Otherwise use the ones here.

## Modes

Establish which mode applies before doing anything else.

1. **Create** — no plan exists. Read the full design set, derive phases, write
   every brief and the tracker.
2. **Extend** — add tasks to a plan in flight. Append IDs; never renumber.
3. **Re-derive** — a phase gate just closed. Rewrite the next phase's briefs
   against the code that now exists. See *Gate procedure*.
4. **Update** — status, ownership, or blocked reason changed. See *Status
   truthfulness*.
5. **Audit** — report coverage gaps, invariant violations, and unreachable code
   without changing the plan.

Briefs are drafts until their phase opens. Writing all of them up front is
correct — it is what makes the whole-system map useful — but a brief written
before its phase's dependencies exist is a hypothesis, and the gate is where it
gets tested against reality.

## The seven invariants

These realize the four properties. They hold of the plan as a whole rather than
of any one task — a plan can violate one while every individual brief reads
well — so every mode checks them explicitly.

### I1 — The system runs from the first task

A composition root exists and produces a startable process before any domain
code is written. No task may leave the system unrunnable. The first task's
demonstration is something like `compose up` serving a health endpoint; it does
almost nothing, but it runs, and every later task keeps it running.

The consequence for task ordering is the whole point: a task can no longer
*add a part to something not yet assembled*. It either wires a new seam into a
system that already starts, or swaps an implementation behind a seam that is
already wired.

### I2 — The system deploys from the first task

Container definition, compose or equivalent, and configuration loading exist
before domain code. Deployment is an invariant, never a task. Migrations run at
startup from the moment there is a database. Deployment appearing as its own
task in a draft is the signal to fold that content into the first task instead.

### I3 — The fake configuration is a first-class, controllable runtime profile

Fakes and in-memory implementations are a **supported way to run the product**,
not test-only doubles. They are documented, exercised end to end in CI, and
runnable by hand at any commit in the plan's history.

Controllable means a person can drive the fake, not just replay it:

- scripted transcripts or scenario files selected at runtime;
- injectable latency, failure, malformed output, and early exit;
- deterministic seeding when determinism is wanted, and the ability to turn it
  off when exploration is wanted.

A fake has real drawbacks — it is not the provider, the database, or the
kernel — and the brief that introduces it says what it does not prove. Stating
those limits is what makes the profile safe to rely on for everything else.

### I4 — Substrate swaps prove parity against a standing scenario suite

One end-to-end scenario suite is written in the first phase against the
all-fake profile, driving the system through its real entrypoint. Every task
that makes a substrate real re-runs that same suite against its substrate.

```
scenario: create → start → stream → input → stop

  all in-memory        ✔
  + real store         ✔ same scenario
  + real process I/O   ✔ same scenario
  + real provider      ✔ same scenario
  + container sandbox  ✔ same scenario
```

"Same scenario, more real" is the acceptance criterion for a swap task. The
suite grows only when a task introduces genuinely new observable behavior.

Automated parity is the floor, not the ceiling. Every brief also carries a
**manual check**: exact commands a person runs and what they should see. A task
whose behavior cannot be observed by hand is a task whose behavior cannot be
observed.

### I5 — A maintained hands-on surface exists early

An operator/debug CLI is built in the first phase and kept working through
every phase. It is the guaranteed way to drive the system by hand, independent
of API churn and of whatever state the product UI is in. Each phase's briefs
state which CLI commands reach the new behavior.

This is not a substitute for the product's real interaction surface; it is what
guarantees interaction exists at all while that surface is being built.

### I6 — Nothing is deferred without an owning task

A brief may not say "later", "a subsequent task", or "out of scope for now"
without naming a task ID **that exists in `tasks.json` at the moment the
deferral is written**. Any stub, TODO, unimplemented branch, or unreachable
error path introduced by a task carries the ID of the task that removes it.

If the deferral is a genuine architectural open question rather than
implementation work, it does not belong in a brief at all — it belongs in the
design set as a deferred decision with an explicit trigger to revisit, and the
brief cites it.

The test is that a reader can always answer "when does this get finished?" with
an ID rather than a shrug.

### I7 — The plan stays wide

Every dependency edge is justified in writing, and every task declares the
surfaces it writes, so that concurrency is derived rather than asserted.

Width needs two independent things, and both must hold: a task must be
*available* — its dependencies are complete — and *non-colliding* — no other
available task writes the same files or contracts. A plan can satisfy the first
and fail the second. Ten available tasks that all edit the composition root are,
in practice, one task.

The mechanism is in *Parallelism* below.

## Task shapes

Layer-ordered work is correct and this skill preserves it. The constraint is
not *what order layers come in* — it is that the seam reaches a running system
in the same task that introduces it.

Every task is one of four shapes. State the shape in the brief header.

**Seam** — introduces a port or boundary, its fake/in-memory implementation,
its registration in the composition root, and its reachability from the CLI or
API. Demonstration: the standing scenario passes with the new seam in the path.

**Swap** — replaces one implementation behind an existing seam with a real one.
Demonstration: the standing scenario passes unchanged against the new
substrate; the fake profile still passes too.

**Capability** — adds genuinely new observable behavior. Demonstration: a new
scenario, added to the standing suite.

**Hardening** — adds adversarial or negative guarantees. Demonstration: the
attack or failure is attempted and visibly refused. Security-sensitive work
requires negative tests, not happy paths.

Every task fits one of the four. A unit of work that fits none of them produces
code no runtime configuration reaches; merge it into the task that first reaches
it, even at the cost of a larger task. Small tasks are what keep a plan legible
and that value survives the merge — reported progress on unreachable code does
not.

## Parallelism

Concurrency is derived from two declarations, in the same way availability is
derived from status rather than stored as a flag.

### Dependency edges carry reasons

Record the reason beside each ID in the brief header. Three reasons are
legitimate:

- **compile** — the task cannot build without a type, interface, or generated
  artifact the other produces;
- **contract** — both tasks would otherwise define the same contract surface:
  the same API path, schema, port, or migration;
- **semantic** — this task's demonstration is impossible until the other's
  behavior exists.

An edge whose reason cannot be written in one clause is not a real edge. In
particular, drop edges resting on "it logically comes after" (ordering
intuition), "same subsystem" (proximity), or "easier to review together" (split
the pull requests instead).

One more illegitimate reason deserves its own paragraph, being the most common
and the most expensive: **"it is nicer to develop against the real thing."** This is
what I3 exists for. A controllable fake profile dissolves these edges and is the
single largest source of width in a layered plan — an adapter that depends on a
real driver because the driver is more convenient to test against should depend
on the fake instead, and be swapped later.

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

## tasks.json rules

The tracker's schema is **frozen**. It has exactly these fields per task, and
nothing may be added — dispatch automation and availability queries read it:

`id`, `title`, `phase`, `phase_name`, `status`, `depends_on`, `owner`,
`blocked_reason`, `brief`

Consequences to respect:

- Everything this skill introduces — shape, reachability, demonstration,
  deferrals, traceability — lives in the **briefs**, not the tracker.
- Availability is derived, never stored: a task is next available when its
  status is `pending` and every ID in `depends_on` is `completed`. Do not add a
  `ready` flag; it goes stale.
- `blocked_reason` is a non-empty string when status is `blocked`, and `null`
  otherwise. The schema enforces this.
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

## Status truthfulness

A status is a claim about the repository, and claims decay. Task statuses have
been found wrong in practice — marked complete when acceptance criteria were
not met.

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
  edge carries one of the three reasons; list the unjustified ones, since each
  is width the plan is giving away.
- **Width** — critical path, average width, and collision set per phase, as in
  *Parallelism*.
- **Round trips** — every user-facing interaction that is modelled or rendered
  is also actionable. A payload kind that can be displayed but not submitted, or
  a state transition the documentation promises but no API exposes, is a round
  trip that stops halfway.
- **Status** — sampled completed tasks that do not hold.

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
  **owned surfaces**.
- Every brief has a **Demonstration** with commands a person can actually run.
- Every dependency edge names compile, contract, or semantic in one clause.
- Every deferral names an ID that exists in `tasks.json`.
- `tasks.json` validates against the frozen schema; `README.md` tables agree
  with it; every `brief` path resolves.
- The dependency graph has no unknown references, forward references, or cycles.
- Critical path, average width, and the collision set are recorded per phase.
- The first task produces a system that starts and deploys.
- Every task fits one of the four shapes.
