# The workplan model

`docs/workplan/` converts the design baseline into implementation-sized tasks
that an agent or a contributor picks up one at a time, each in its own pull
request.

Two skills act on it. Both assume this document and neither repeats it.

| Skill | Subject | Use it to |
|---|---|---|
| [`workplan`](workplan/SKILL.md) | the plan | create, extend, re-derive at a gate, correct status, audit the plan |
| [`task`](task/SKILL.md) | one task | implement a brief, verify it, judge whether it is done |

Changing the plan and doing a task are different jobs with different
boundaries. Establish which one you are in before doing anything else.

## Artifacts

| Path | Purpose |
|---|---|
| `docs/workplan/README.md` | Narrative index, per-phase dependency tables, phase exit rule, how to use the plan |
| `docs/workplan/phase-N/<id>-<slug>.md` | One brief per task |
| `docs/workplan/tasks.json` | Machine-readable tracker — status, ownership, dependencies |
| `docs/workplan/tasks.schema.json` | Schema for the tracker |
| `docs/workplan/post-mortems/` | Records of *plan* failures, not implementation bugs |

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

## The seven invariants

These realize the four properties. They hold of the plan as a whole rather than
of any one task — a plan can violate one while every individual brief reads
well.

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

A CLI existing is not the same as new behavior being reachable from it.

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

The obligation is to own what you defer. It is not to enumerate everything the
system has not finished: a Deferrals table is not an inventory of remaining
work, and reading it as one finds infinite debt in every task.

### I7 — The plan stays wide

Every dependency edge is justified in writing, and every task declares the
surfaces it writes, so that concurrency is derived rather than asserted.

Width needs two independent things, and both must hold: a task must be
*available* — its dependencies are complete — and *non-colliding* — no other
available task writes the same files or contracts. A plan can satisfy the first
and fail the second. Ten available tasks that all edit the composition root are,
in practice, one task.

### What the invariants oblige of one task

The plan as a whole is answerable for all seven. A single task is answerable
for the ones its own change engages: it leaves the system runnable and
deployable (I1, I2), keeps the fake profile working and honest (I3), re-runs
the standing suite when it makes a substrate real (I4), states which commands
reach its new behavior (I5), owns what it defers and what it stubs (I6), and
stays inside its declared surfaces (I7).

It is not answerable for the plan's compliance overall. Debt a task did not
introduce belongs to whoever introduced it.

## Task shapes

Layer-ordered work is correct and this model preserves it. The constraint is
not *what order layers come in* — it is that the seam reaches a running system
in the same task that introduces it.

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
it, even at the cost of a larger task.

A brief that fits *several* shapes at once has the opposite problem: it is not
one big task but several tasks sharing a header, and the response is to split it
along the shapes.

## Vocabulary

**Dependency edge.** A completion prerequisite, never suggested reading. Each
edge names one of three reasons in one clause:

- **compile** — the task cannot build without a type, interface, or generated
  artifact the other produces;
- **contract** — both tasks would otherwise define the same contract surface:
  the same API path, schema, port, or migration;
- **semantic** — this task's demonstration is impossible until the other's
  behavior exists.

An edge whose reason cannot be written in one clause is not a real edge.

**Owned surfaces.** The files, directories, and contracts a task writes.
Concurrency is derived from this list rather than asserted: two tasks with
disjoint surfaces run concurrently whatever phase they sit in; two tasks with
overlapping surfaces serialize regardless of what the graph says.

**Demonstration.** Exact commands a person runs and the observable result. It
is a manual check, not a formality.

**Witness.** The surface that produces the evidence for a criterion. A
criterion whose only witness is a task downstream of it cannot be judged, and
that is a defect in the plan rather than in the implementation.

## The tracker

`tasks.json` has exactly these fields per task, and the schema is frozen:

`id`, `title`, `phase`, `phase_name`, `status`, `depends_on`, `owner`,
`blocked_reason`, `brief`

- Status is one of `pending`, `in_progress`, `blocked`, `completed`.
- **Availability is derived, never stored.** A task is next available when its
  status is `pending` and every ID in `depends_on` is `completed`.
- `blocked_reason` is a non-empty string when status is `blocked`, and `null`
  otherwise.
- Everything else the model introduces — shape, reachability, demonstration,
  deferrals, traceability — lives in the **briefs**.

Status is a claim about the repository, and claims decay. `completed` means the
acceptance criteria were verified against HEAD, not that a pull request said so.
