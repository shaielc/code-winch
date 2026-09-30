# Post-mortem: Phase 0 workplan creation (2026-08-23)

## Briefs involved

This post-mortem covers the **create-mode derivation** of Phase 0, not a single
implementation task. The briefs that were written, revised, or removed during the
session:

| ID | Title | Outcome |
|---|---|---|
| P0-001–P0-005 | Foundation repair (gates, CLI ship, fake profile, dev-run CI, doc sanitize) | Written in first pass; dependencies mostly stable |
| P0-006 (v1) | Bind run round trip through daemon | Monolithic seam; **removed** |
| P0-007 (v1) | Standing end-to-end scenario suite | Terminal capstone task; **removed** |
| P0-006 (v2) | Create and read runs | Split from v1 |
| P0-008–P0-011 | Start, input+outbox, WebSocket stream, stop | Split from v1; dependencies revised several times |

No single brief was wrong on its face at the moment it was drafted. Each
satisfied the brief anatomy and traced to `docs/state.md`. The defect is in
**how create mode decomposed a capability gap into tasks and edges** before the
first review.

## What went wrong

### 1. Many review rounds before the plan stabilized

The first pass produced five repair tasks plus a monolithic run-binding seam and
a terminal e2e task. Five subsequent reviews were needed to:

- add e2e to Phase 0;
- split the run seam;
- separate outbox worker from WebSocket;
- split input from stop;
- fix dependencies (remove policy edges, add P0-004 → P0-003, remove P0-007).

The model and skill describe how to **audit** a plan and how to **split** an
overstuffed draft, but create mode has no **derivation procedure** — no step
that says how to go from a `state.md` gap to a chain of operator-sized tasks
before the first publish.

### 2. Bias toward large tasks

The first run-binding brief bundled application use cases, `postgres.Store`,
supervisor, local runner, outbox worker, `EventStream.Publish`, driver
registries, and HTTP demonstration — with CLI deferred to Phase 1.

That followed a literal reading of:

- **Seam shape** (`workplan-model.md` §Task shapes): one task owns boundary,
  implementation, composition-root wiring, public surface, and minimal CLI.
- **Merge when reachability is unclear** (`workplan-model.md` lines 207–211):
  merge into the task that first reaches it, even at the cost of a larger task.
- **As few seam tasks as the boundaries allow** (`skills/workplan/SKILL.md`
  §Width, lines 159–160).

Interpreting “run round trip” as **one capability** made the seam definition
prescribe one large task. Deferring CLI appeared to resolve the size problem
but violated **I5** and **completion closure** — caught only in review, not by
a create-mode check.

### 3. Dependencies misplaced

| Edge (as first written) | Problem | Legitimate reason? |
|---|---|---|
| P0-006 → P0-001 | “CI must gate storage before seam merges” | **No** — policy preference, not compile/contract/semantic |
| P0-007 → P0-001, P0-011 | e2e capstone waits on CI + full chain | Partially; P0-007 itself was unnecessary |
| P0-011 → P0-010 | “full observe path before stop” | **No** — ordering intuition (`workplan-model.md` lines 276–278) |
| P0-004 missing → P0-003 | CI exercises fake profile | **Yes** — semantic; added only after review |

Create mode mandates checking I1–I7 and recording width metrics, but does not
require walking **every** `depends_on` entry through the four edge reasons and
rejecting policy edges before the plan is published. Audit mode catches
unjustified edges **after** the fact.

### 4. Deferrals pointed at a phase instead of a task

Two briefs carry a deferral row whose owning task is `Phase 1 (not yet planned)`:

| Brief | Deferred | Named owner |
|---|---|---|
| P0-003 | Controllable in-memory store profile | Phase 1 (not yet planned) |
| P0-011 | Browser `winch_session` establishment | Phase 1 (not yet planned) |

I6 requires an ID **that exists in `tasks.json` at the moment the deferral is
written**. A phase name is not an ID, and phase 1 was deliberately not derived,
so there was no ID to name. Neither brief is wrong on its face — each is honest
that the work is not being done, and "Phase 1" reads as informative. The rule
only breaks when you ask whether `tasks.json` can resolve the string, which is
invisible from inside a single brief.

The deeper cause is structural, and the fault is in the rule rather than in the
briefs: **a plan derived for a subset of phases will always generate deferrals
pointing outside itself.** I6 offers exactly two escapes — an existing task ID,
or a deferred decision in the design set with a trigger — and neither is
reachable for ordinary implementation work in a phase nobody has derived yet.
Create mode permits a partial derivation, so it must also permit a deferral that
points at one. It does not, so both briefs wrote the only true thing available
and were wrong by construction.

What makes a phase deferral a shrug is not the phase name; it is that nothing
carries it forward. `post-mortems/` and the briefs are both stripped at close,
and no create-mode input requires reading the previous plan's outstanding
deferrals. So "Phase 1" is a promise with no reader. Legitimizing the form means
fixing that, not banning it.

A second flavor appeared in the same session. P0-001's non-goals named P0-007 as
the owner of the e2e suite; P0-007 was then removed later in the derivation, and
the reference was left dangling. I6 constrains the moment a deferral is written
and says nothing about what happens when the named task is deleted or its scope
moves, so nothing required a sweep.

### 5. Phase 0 went straight to PostgreSQL and never built the fake substrate

Every run task — P0-006, P0-008, P0-009, P0-010, P0-011 — declares
`PostgreSQL storage` under *Runtime reachability*. The plan therefore contains no
configuration in which the product runs without a database, and
`internal/adapters/memory` — which implements every port — stays a test double
through the whole phase. That is the *port whose only implementations are test
doubles* smell, and `docs/state.md` had already recorded it against the previous
plan; the new plan inherited it and scheduled no repair.

This inverts I3 and I4 rather than merely under-serving them. I4's ladder starts
at the all-fake profile and each swap task re-runs the same scenario against a
more real substrate. Starting at the real substrate leaves the ladder with no
first rung: there is no fake baseline to prove parity against, so every
"same scenario, more real" criterion in later phases has nothing to compare to.

**The phase could have been satisfied without the database entirely.** Create,
read, start, input, stream, and stop are all demonstrable against
`internal/adapters/memory` plus the fake harness and the local sandbox.
PostgreSQL then becomes a swap task that re-runs those same scenarios — which is
what I4 describes.

The cost was visible within one audit round:

- `make e2e` requires a live PostgreSQL, so it cannot run under a plain
  `go test`, and P0-006 acquired a dependency on P0-001 purely to get a database
  into CI. Against a memory profile that edge does not exist — width given away
  by a substrate choice rather than by a real prerequisite.
- P0-003 correctly identified the missing piece and deferred it, which is how
  this defect and defect 4 compound: the plan skipped the fake substrate and then
  deferred it to nowhere.

No single brief is wrong on its face here either. `docs/state.md` describes gaps
in terms of the deployed product, which is PostgreSQL-backed, so deriving tasks
against that substrate looks like faithful coverage. Nothing inside one brief
signals that the profile ladder was never built.

## What would have caught it

Rules that already exist but were applied too late or outweighed by other text:

1. **Overstuffed-seam test** (`skills/workplan/SKILL.md` lines 82–89) — classify
   scope bullets; extract everything beyond required behavior, wiring, operator
   reachability, and invariant preservation.
2. **Illegitimate dependency edges** (`workplan-model.md` lines 276–278) —
   ordering intuition and “same subsystem” are not edges.
3. **Completion closure / smell** (`skills/workplan/SKILL.md` lines 319–323,
   359–360) — a task must not defer CLI for a capability it introduces; a later
   task must not exist only to make an earlier task operable.
4. **Suite grows with behavior** (`workplan-model.md` line 112) — I4’s standing
   scenario should extend incrementally per task, not arrive as one terminal
   task.
5. **Split by distinct observable property** (`skills/workplan/SKILL.md` lines
   76–80) — create/start/input/stop/outbox/stream are distinct properties.
6. **I3 — the fake configuration is a first-class runtime profile**
   (`workplan-model.md` §I3) — fakes are a supported way to run the product, so
   an in-memory store is a profile the plan ships, not a double it inherits.
7. **I4 — the standing suite is written against the all-fake profile**
   (`workplan-model.md` §I4, the ladder at lines 101–109) — the first rung is
   all-in-memory; real substrates are swap tasks above it.
8. **Smell: a port whose only implementations are test doubles**
   (`skills/workplan/SKILL.md` §Smells) — true of `internal/adapters/memory`
   before the plan was written and still true after it.

Rules that **do not** exist yet but would have prevented the failure:

9. **Operator-verb slicing (create mode)** — derive run-path tasks from one
   HTTP/CLI operation per seam, each with its minimal CLI command in the same
   task.
10. **Dependency derivation pass (create mode)** — for each edge, write
    compile | contract | semantic in one clause; delete the edge if the clause
    is policy (“CI should gate first”) or proximity (“same subsystem”).
11. **CI-exercises-profile edge** — a hardening task that runs a profile in CI
    has a **semantic** dependency on the task that defines that profile.
12. **Iterative e2e** — the task that introduces observable behavior extends
    `make e2e` in the same PR; no separate “standing suite” capstone unless it
    only wires CI for an already-growing scenario file.
13. **Phase deferral as a legitimate owner form** — I6 admits only a task ID or
    a design-set deferred decision, which a partial derivation cannot always
    supply. It needs a third form: a phase brief that exists before its tasks
    do, carries a register of the deferrals aimed at it, and survives a close.
14. **ID sweep on removal** — when a task is removed or its scope moves during
    derivation, every deferral and non-goal naming it dangles, and nothing
    requires re-resolving them before publish.
15. **Substrate ladder check (create mode)** — a phase in which every task
    names a real substrate has skipped I4's first rung; the fake profile is a
    task inside that phase, never a deferral out of it.

## Stabilized outcome

The derivation session ended with P0-007 removed, the e2e suite growing through
P0-006 to P0-011, and the CI gate in P0-011. A full audit on 2026-08-23 found
that shape unworkable and revised it; the section below is the state after that
audit.

```
P0-001 CI gates ──► P0-006 create/read ──► P0-008 start ──► P0-009 input+outbox ──► P0-010 stream ──┐
       └──────────────────────────────────────────────┐              │                              │
P0-003 fake profile ──────────────────────────────────┘              │                              ▼
       └──────────► P0-004 fake run in CI                            └──► P0-011 stop ────────► P0-007 finalize
P0-002 ship CLI ──► P0-004                                                                  ▲
P0-001 ──────────────────────────────────────────────────────────────────────────────────────┘
```

- **P0-001, P0-002, P0-003, P0-005** independent at open. P0-006 is no longer
  among them.
- **P0-007 reinstated** as *Finalize phase 0*: it adds the assembled
  `create → start → stream → input → stop` scenario, wires `make e2e` into CI,
  and closes I4. It is not the capstone the first derivation removed — every
  operation still contributes its own scenario in its own task.
- **Each of P0-006 and P0-008 to P0-011 owns one partial scenario in its own
  file**, and states in its non-goals which steps it does not add. Separate
  files removed the `test/e2e/` write collision between P0-009 and P0-011.
- **P0-006 → P0-001** was re-added, with a different clause than the one
  rejected during derivation: not "CI should gate storage first" but that
  P0-006's durability claim is unverifiable in CI until a database exists there.
  It costs width, and it is a symptom of defect 5 — against an in-memory profile
  the edge would not exist at all.
- **P0-008 hardcodes the `fake`/`local` pair** and rejects any other persisted
  profile, rather than building a driver registry. `docs/code-structure.md:119`
  keeps registration explicit in the composition root either way, and ADR-0003
  requires the rejection.

Critical path 6 (`P0-001 → P0-006 → P0-008 → P0-009 → P0-010 → P0-007`), average
width 11 ÷ 6 ≈ 1.8, four tasks available at open. The first derivation reported
5 and 2.0, both wrong: the chain it measured included a `P0-003 → P0-006` edge
that never existed.

## Plan rules to add (proposed)

Add to **create mode** in `skills/workplan/SKILL.md`, after *Creating a plan
over existing code*:

### Derive tasks from operator verbs

When `docs/state.md` records a multi-step interaction (for example
`create → start → stream → input → stop`), default to **one task per step**
unless two steps share a single observable demonstration. Each task includes the
minimal CLI for that step (I5). Do not map the whole interaction to one seam
because the gap is described as one capability.

### Extend the standing suite in the introducing task

When I4 applies, the task that introduces new observable behavior adds the
matching scenario steps and keeps `make e2e` green through steps introduced so
far. Do not create a terminal “e2e suite” task unless its only job is CI wiring
and the scenario file already exists.

### Dependency derivation pass (required before publish)

For every `depends_on` entry, record compile, contract, or semantic in the brief
header. Reject and remove edges whose clause is:

- CI or merge policy (“should gate in CI first”);
- subsystem proximity (“same area of the codebase”);
- capstone ordering (“full path before partial”).

Write collisions are not edges (`workplan-model.md` lines 334–337).

### Profile CI dependency

If a task runs automated tests against `harnessProfile=fake` (or any named
profile), it **semantically depends** on the task that defines controllable
behavior for that profile.

### Deferring to a phase that is not yet derived

Amend **I6** in `workplan-model.md` to admit a third owner form: a phase. What
makes a task ID an owner is that it resolves to a brief at the moment the
deferral is written, so a phase earns the same standing the same way — **every
phase gets a brief, including phases with no tasks yet.**

`docs/workplan/phase-N/README.md`, alongside that phase's task briefs. A phase
that has not been derived is a directory holding only its brief. It carries:

- **Objective** — what the phase makes true, stated as the phase exit condition.
- **Scope** — the capabilities it covers, traced to `docs/roadmap.md` and to
  `docs/state.md` gaps.
- **Deferrals in** — the register: every deferral from another phase that names
  this one, each with the brief it came from. This is what the phase brief
  exists for.
- **Status** — derived (its tasks are in `tasks.json`) or not derived yet.

The rule then reads symmetrically with the task rule. A deferral may name a
phase when that phase's brief exists **and lists the deferral in its register**.
Writing a phase deferral means editing two files, exactly as writing a task
deferral means the owning task already exists. A phase name that resolves to no
file is still a shrug.

This needs no tracker change: `phase` and `phase_name` are already fields, and
the frozen schema says nothing about phase documents.

Two obligations close the loop across a plan boundary:

- **Close** folds every underived phase brief's register into `docs/state.md`
  under *What is not implemented*, the way it restates post-mortem rules. Briefs
  do not survive a close; the register has to.
- **Create** treats those entries as a required input alongside `docs/state.md`
  gaps. Deriving a phase whose register has entries, and giving one of them no
  task, is a coverage defect the audit reports.

### Re-resolve references when a task is removed

Removing a task or moving its scope during derivation dangles every deferral,
non-goal, and acceptance criterion that names it. I6 constrains only the moment
a deferral is written. Before publish, re-resolve every task ID mentioned in
prose against `tasks.json` — the same sweep the close procedure runs over the
tree, applied to the plan's own text.

### Build the substrate ladder before the substrates

When I4 applies, the first phase's standing scenario runs against the all-fake
profile — including storage — before any task declares a real substrate. If
every task in a phase names a real database, queue, or provider, the plan has
skipped I4's first rung and there is nothing for later swap tasks to prove
parity against.

The check is mechanical: read the *Runtime reachability* profile line of every
task in the phase. If none of them is all-fake, the fake profile is a missing
task in that phase — not a deferral out of it, and not something the previous
plan's inherited test doubles already satisfy.

Add a cross-reference in **`workplan-model.md`** §Task shapes (Seam): a seam is
a complete vertical slice for **one operator operation**, not for every
operation behind a port or use-case family.

## Traces to

- `skills/workplan/SKILL.md` — Create mode, *The split test*, *The
  overstuffed-seam test*, *Width*, *Smells*, *Before finishing*
- `skills/shared/workplan-model.md` — I4, I5, Seam shape, Completion closure,
  Dependency edges, Write set
- `docs/state.md` — *What went wrong*, *What is not implemented* (run round trip)
