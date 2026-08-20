# The workplan model

`docs/workplan/` converts a design baseline into implementation-sized tasks that
an autonomous agent or a contributor can pick up one at a time, each in its own
pull request.

This document defines the terms that plan is written in: its layout, the
invariants that hold across it, the shapes a task takes, the anatomy of a brief,
and the rules its tracker enforces. It is the reference the plan as a whole and
any individual task are both checked against.

## The four properties

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

## Layout

| Path | Purpose |
|---|---|
| `docs/workplan/README.md` | Narrative index, per-phase dependency tables, phase exit rule, how to use the plan |
| `docs/workplan/phase-N/<id>-<slug>.md` | One brief per task |
| `docs/workplan/tasks.json` | Machine-readable tracker — status, ownership, dependencies |
| `docs/workplan/tasks.schema.json` | Schema for the tracker |
| `docs/workplan/post-mortems/` | Records of plan failures — defects whose root cause is a seam between briefs rather than an implementation |

An existing plan's conventions win: match its ID scheme, directory layout, and
brief structure rather than the illustrative ones here.

## The seven invariants

These realize the four properties. They hold of the plan as a whole rather than
of any one task — a plan can violate one while every individual brief reads
well — and every task is responsible for leaving them intact at HEAD.

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

Every dependency edge is justified in writing, and every task declares both the
files it writes and the contracts it defines, so that concurrency is derived
rather than asserted.

Width needs two independent things, and both must hold: a task must be
*available* — its dependencies are complete — and *non-colliding*. A plan can
satisfy the first and fail the second. Ten available tasks that all edit the
composition root are, in practice, one task.

The two kinds of collision are not the same defect and do not get the same
response:

- a **contract collision** — two tasks that would change the meaning of the same
  externally significant contract — is resolved in the graph. One of them
  depends on the other. Two concurrently available tasks sharing a contract
  surface is a plan defect, not a scheduling inconvenience.
- a **write collision** — two available tasks whose concrete write sets overlap
  — is a warning about merge contention. It creates no edge. It is reported so
  that whoever dispatches both knows they will rebase.

*Dependency edges*, *Write set*, and *Contract surfaces* below are the
mechanism.

## Task shapes

Layer-ordered work is correct and the plan preserves it. The constraint is not
*what order layers come in* — it is that the seam reaches a running system in
the same task that introduces it.

Every task is one of four shapes, stated in the brief header.

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

## Dependency edges

Each dependency records its reason beside the ID in the brief header. Four
reasons are legitimate, and the fourth is not like the other three:

- **compile** — the task cannot build without a type, interface, or generated
  artifact the other produces;
- **contract** — the two tasks' declared **contract surfaces** overlap: the same
  API operation, schema component, port signature, persisted aggregate,
  configuration key, registry namespace, or migration slot. Every contract
  collision becomes an edge; *Contract surfaces* below decides which task comes
  first;
- **semantic** — this task's demonstration is impossible until the other's
  behavior exists;
- **revision** — this task's object *is* implementation the other task creates
  or substantially rewrites, so doing it first would revise code that does not
  survive. Found rather than planned; see below.

An edge whose reason cannot be written in one clause is not a real edge. In
particular, edges resting on "it logically comes after" (ordering intuition),
"same subsystem" (proximity), or "easier to review together" (split the pull
requests instead) are not edges.

One more illegitimate reason deserves its own paragraph, being the most common
and the most expensive: **"it is nicer to develop against the real thing."**
This is what I3 exists for. A controllable fake profile dissolves these edges
and is the single largest source of width in a layered plan — an adapter that
depends on a real driver because the driver is more convenient to test against
should depend on the fake instead, and be swapped later.

Dependencies are completion prerequisites, not suggested reading.

### Revision edges are found, not planned

The first three reasons are properties of the plan and are derivable while
writing it. `revision` is not. It records something learned after work started:
a defect found while implementing, an audit that failed, a write collision whose
reconciliation turned out to need more than a rebase.

A revision edge in a freshly derived phase is therefore a defect in the
derivation, not a well-ordered plan. If the planner already knows an
implementation will want hardening, restructuring, or better tests, that work
belongs in the task that writes the implementation. Planning a revision ahead of
time is planning to do the work twice.

Revision tasks are added in extend mode, after the fact, and the brief says what
was found. The edge holds only when the dependent names the implementation or
test surfaces the prerequisite produced and its objective is specifically to
revise those surfaces.

This is also what closes a plan defect. When a task cannot finish because the
implementation it builds on is wrong, the remedy is neither absorbing the fix
silently nor leaving a TODO: it is a revision task, whose ID is what makes the
deferral owned under I6.

`revision` launders none of the illegitimate reasons. "Easier to do later",
"avoids rework in general", "same subsystem", "nicer development order", and
"the prerequisite might influence the design" are still not edges. The edge says
the object of this task does not meaningfully exist in its intended form before
that task — not that this order is preferable.

## Write set

The concrete files and directories a task writes. This models merge contention
and nothing else.

```text
Write set:
- cmd/winch/credential.go
- internal/application/credential.go
- internal/adapters/postgres/migrations/007_*.sql
```

Name the files actually expected. A directory is a legitimate entry when the
task genuinely owns it — creates it, or rewrites most of it — but not as
shorthand for files the planner could have listed.

Two available tasks whose write sets overlap are a **write collision**. That is
a warning, not an edge: both proceed, and whoever takes the second one rebases.
Report the pairs so the cost is visible; do not add a dependency to make the
report clean. Where reconciling the overlap turns out to need real work rather
than a merge, that work is its own task, carrying a `revision` edge to what it
reconciles.

Surfaces that attract every task at once, and so deserve attention early: the
composition root, the standing scenario suite, ordered migration numbers, a
single API specification file, and any central registry, switch, or factory.

## Contract surfaces

The externally significant contracts and namespaces whose *meaning* the task
defines or changes, regardless of which files carry them.

```text
Contract surfaces:
- API: POST /api/v1/credentials
- schema: RunSpec.credential_refs
- migration namespace: credentials
```

The kinds that count:

- an HTTP resource or operation;
- an OpenAPI or event-schema component;
- a port or interface signature;
- a persisted aggregate or one of its state transitions;
- a configuration key;
- a driver, profile, or registry namespace;
- a migration slot or schema-ownership boundary.

Two tasks sharing a contract surface have a **contract collision**, and a
contract collision is a dependency edge. The direction is not a coin flip: the
task that **defines** the surface comes first, and the task that extends,
implements, or consumes it takes the edge with reason `contract`.

If both tasks would define the same surface there is no direction to pick, and
that is the plan being wrong rather than the rule failing. Either one task owns
the surface and the other narrows its own, or the two are one task.

What the two declarations model is the whole distinction: a write set predicts a
rebase, a contract surface predicts a disagreement about meaning. Two tasks can
share a file with no contract collision — appending independent handlers to one
registry — and can collide on a contract while sharing no file at all, which is
exactly the case a write set misses.

## Brief anatomy

Briefs are short. A sentence that would be identical across every brief is
boilerplate: it belongs in `README.md` once. Identical "architectural context"
and "deliverables" sections across a whole plan are a reliable sign that
up-front depth was padding.

```markdown
# <ID>: <Title>

**Phase:** <N> — <phase name>
**Shape:** seam | swap | capability | hardening
**Dependencies:** <ID> (compile | contract | semantic | revision: <one clause>), … or None

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

## Write set

Concrete files and directories this task writes. Overlap with another available
task is a merge warning, not an edge.

## Contract surfaces

Contracts and namespaces whose meaning this task defines or changes. Overlap
with another task is a dependency edge, not a warning. Or: `None.`

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

## The tracker

`docs/workplan/tasks.json`'s schema is **frozen**. It has exactly these fields
per task, and nothing may be added — dispatch automation and availability
queries read it:

`id`, `title`, `phase`, `phase_name`, `status`, `depends_on`, `owner`,
`blocked_reason`, `brief`

Consequences to respect:

- Shape, reachability, demonstration, deferrals, and traceability live in the
  **briefs**, not the tracker.
- Availability is derived, never stored: a task is next available when its
  status is `pending` and every ID in `depends_on` is `completed`.
  `scripts/list-available-tasks.sh` computes it. Do not add a `ready` flag; it
  goes stale.
- `blocked_reason` is a non-empty string when status is `blocked`, and `null`
  otherwise. The schema enforces this.
- IDs are append-only. Merged pull requests, dispatch history, and cross-brief
  references all cite them; renumbering silently invalidates that record. A task
  inserted logically "between" two others still gets the next free number.
- The ID pattern permits phases 0–5 only. A seventh phase requires a schema
  change, which is out of bounds — fold the work into an existing phase or raise
  it as a decision.
- `brief` paths are `phase-N/<slug>.md` and must resolve to a real file.
- Status fields are stamped by automation when a pull request is approved
  (`scripts/stamp_task_completion.py`). An implementing change does not edit
  them.

`README.md`'s dependency tables and `tasks.json` are two views of one graph.
Drift between them is a defect.

## Status truthfulness

A status is a claim about the repository, and claims decay. Task statuses have
been found wrong in practice — marked complete when acceptance criteria were
not met.

- `completed` means the acceptance criteria were verified against HEAD, not
  against the pull request description.
- `blocked` requires a reason naming what would unblock it, ideally a task ID.
- A dependency's `completed` status is itself a claim. If the work a task needs
  from a dependency is missing at HEAD, the dependency is not complete, whatever
  the tracker says.
