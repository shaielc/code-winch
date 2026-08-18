---
name: task
description: Implement, verify, or audit one workplan task. Use when picking up a task ID, driving a brief to completion, checking whether a task is done, reviewing the change that claims to satisfy it, or deciding what a task should and should not absorb. Covers locating a brief from an ID, which brief sections bind, the boundary around a single task, and how to report findings against it without enlarging it. To change the plan itself — add, split, re-derive, or re-scope tasks — use the workplan skill instead.
---

# Task

Your subject is one brief and the change that claims to satisfy it. The rest of
the repository is context you may read and may not take on.

Read [`skills/workplan-model.md`](../workplan-model.md) first — what a workplan
is, the seven invariants, the four shapes, the vocabulary. Read `AGENTS.md` for
the rules every change in this repository obeys. Neither is repeated here.

Given work on the plan — add or split tasks, re-derive a phase, re-check the
graph or a phase's width — use
[`skills/workplan/SKILL.md`](../workplan/SKILL.md) instead.

Two modes, sharing one boundary:

1. **Implement** — do the task.
2. **Audit** — judge whether a task is done, changing nothing.

The boundary is the part that goes wrong. An audit that reports another task's
debt, or an implementation that quietly absorbs it, both fail the same way.

## Locate it

```sh
ID=P1-050
jq -r --arg id "$ID" '.tasks[] | select(.id == $id)' docs/workplan/tasks.json
```

The `brief` field resolves under `docs/workplan/`. Before reading anything else,
establish what sits on either side of the task:

```sh
# dependencies — completion prerequisites, not suggested reading
jq -r --arg id "$ID" '.tasks[] | select(.id == $id) | .depends_on[]' \
  docs/workplan/tasks.json

# dependents — the work that waits on this task, and that you must not do
jq -r --arg id "$ID" '.tasks[]
  | select(.depends_on | index($id)) | "\(.id) \(.status) \(.title)"' \
  docs/workplan/tasks.json
```

Tasks whose dependencies are all satisfied: `./scripts/list-available-tasks.sh`.

For an audit, get the change under judgment rather than the whole tree —
`git diff main...HEAD` for a branch in flight, or the merged commits for a task
already landed (`gh pr list --search "$ID" --state all`, since `Task: <ID>`
lives in pull request bodies). Judging a task against the tree's total state is
how audits end up reporting other tasks' work.

## What binds you

| Section | Force |
|---|---|
| **Scope** | Work you owe. |
| **Acceptance criteria** | Individually checkable; all must hold. |
| **Verification** | Minimum evidence expected, not a ceiling. |
| **Demonstration** | Run it. A demonstration nobody ran is a criterion nobody checked. |
| **Owned surfaces** | Where you may write. |
| **Non-goals** | Work you owe *not* to do. |
| **Deferrals** | Named debt with another owner. Not yours to discharge. |
| **Runtime reachability** | The configuration your work must be reachable from when you are done. |
| Objective, Traces to | Orientation. |

Non-goals and Deferrals bind as hard as Scope. They record decisions made
before you arrived, usually because another task owns the work.

## The boundary

A task is answerable for the change it makes and the claims its brief makes.
Not for everything visible from where it stands.

Before treating anything as this task's problem, check who owns it:

```sh
grep -rn "<symbol, path, or capability>" docs/workplan/
```

Then apply, in order:

1. **Another task owns it** — not yours. Name the ID and stop. An adjacent
   half-built thing whose owner is scheduled later is the plan working, not a
   defect.
2. **Your brief's Non-goals or Deferrals name it** — not yours. Already decided.
3. **It predates your change** — not yours. The diff, or `git log -S`, shows who
   introduced it. Incompleteness you noticed while reading is not debt you
   inherited by reading it.
4. **Nothing owns it and your change introduced it** — yours. Finish it, or add
   a task and cite the ID (I6).
5. **Nothing owns it and your change did not introduce it** — the plan has a
   hole. Report it; do not fill it. Filling it silently turns one task's change
   into three tasks' work with no record of where it went.

Plan-shape questions — design-set coverage, dependency graph, phase width,
reachability across the tree — are not task questions, and do not become task
questions because a task touched a file near them. Report them as observations
and leave them with the plan.

## Implementing

- `AGENTS.md` binds: inward dependencies, runnability, fakes as a shipped
  configuration, adversarial tests where it matters, verification, pull request
  conventions.
- Write inside your owned surfaces. An edit outside them collides with another
  branch and usually signals scope creep; where it is unavoidable, say why in
  the pull request.
- Run the Demonstration by hand and report what you observed.
- Every stub, TODO, or unimplemented branch you add carries an existing task ID.
- Report verification honestly: what ran, what failed, what you could not run.
- Do not edit status fields in `tasks.json`.

### When the brief itself is wrong

Briefs are hypotheses written before the code existed, and some do not survive
contact. Raise these; do not absorb them.

- **A criterion you cannot witness.** No command in the repository produces the
  evidence, or the only surface that could is a task downstream of yours. That
  is a plan defect, and the fix is a plan change, not an unowned surface
  invented to paper over it. See
  `docs/workplan/post-mortems/2026-08-17-unverifiable-stream-criterion.md` for
  what it costs.
- **Scope spanning several shapes.** A brief that is a seam, a capability, and a
  telemetry foundation at once is several tasks sharing a header. Splitting it
  is a gate decision, not something to perform mid-implementation.
- **A contract change the brief did not anticipate.** Update the design document
  that describes the surface, or add or supersede an ADR, in the same change.
- **A criterion that over-promises.** The brief claims something the design set
  does not require and the code does not do. The correction is usually to the
  brief; see *The smallest true correction* below. Say so rather than building
  to the wording.

## Auditing

Change nothing — not the plan, not the code. Read the brief, then the change,
then check in this order. Every finding cites a path and line.

1. **Acceptance criteria** — each one, against HEAD, true or false. Not "mostly".
2. **Demonstration** — run it. The expected result appears or it does not.
3. **Verification** — the listed evidence exists and passes.
4. **Scope** — every bullet has corresponding work in the change.
5. **Owned surfaces** — the change writes what the brief declared, and declared
   what it wrote. Undeclared surfaces are a width defect worth naming.
6. **Deferrals and stubs** — every deferral *this brief writes* names a task
   that exists, and every stub *this change introduces* carries an ID.

That list is the audit. Its boundary is the one above.

### Findings and observations

Report in two sections, and the split matters more than the contents.

**Findings** — inside the boundary. Each names the criterion or invariant it
violates, the evidence at a path and line, and the smallest correction. These
block completion.

**Observations** — outside it. What you noticed elsewhere, each with the owning
task ID where one exists, and an explicit note that it does not block this task.
Do not suppress them; file them where someone can act on them.

"Worth resolving before the task closes" is neither. It is a problem handed to
whoever is holding the task, and it is the phrasing by which an audit becomes
scope creep.

### The smallest true correction

When the brief and the repository disagree, one of them is wrong, and the audit
says which.

**The default is the brief.** It was written as a hypothesis. A criterion that
over-promises against what the design set actually requires is corrected by
rewriting the criterion; growing the code to satisfy the over-promise adds scope
nobody asked for and no document requires.

Grow the code only when the design set — `docs/architecture.md`,
`docs/contracts.md`, `docs/security.md`, `docs/roadmap.md`, the ADRs — requires
the missing behavior. Then the remedy is a new or existing owning task, named.
Never recommend that the task in flight absorb work it did not open.

### Do not move scope out of it either

The opposite correction is the more tempting one, and it fails the same way. A
brief too large to review is a decomposition defect: report it, and let the gate
re-derive it. Moving scope, a criterion, or a criterion's witness to another
task changes the graph the whole phase was planned against, and a review sees
one brief and one diff — not the tasks that declare `depends_on` this one.

Two failure modes, both recorded in
`docs/workplan/post-mortems/2026-08-17-unverifiable-stream-criterion.md`:
whatever is easiest to detach is rarely what should move, and detaching a
criterion's witness leaves a criterion nobody can check.

Review effort belongs to the pull request. Splitting the *pull request* is
always available, changes no brief, and is the correct answer to "this is too
large to read".

### Verdict

End with the tracker status the evidence supports, and nothing else:

- `completed` — every acceptance criterion holds at HEAD.
- `in_progress` — work has started and some criteria do not hold.
- `blocked` — with what would unblock it, ideally a task ID.
- `pending` — not started.

A task whose criteria hold is complete even if you found plan-level problems
while reading it. Do not invent an `owner`: it is set by whoever picks the task
up. Do not edit `tasks.json` — recommend the status and let the planner or
automation stamp it.

## Before you call it done

Implementing:

- [ ] Every acceptance criterion holds, checked against the working tree.
- [ ] The Demonstration was run, and what it showed is in the pull request.
- [ ] `make check` — or `make test-cycle` with its stated gap — plus everything
      the brief's Verification section lists.
- [ ] Nothing written outside owned surfaces without saying why.
- [ ] Every deferral and stub names a task ID that exists in `tasks.json`.
- [ ] `Task: <ID>` in the pull request body, and no other task ID.

Auditing:

- [ ] Every acceptance criterion has a verdict and evidence at a path and line.
- [ ] Findings are inside the boundary; everything else is an observation with
      an owner and a note that it does not block.
- [ ] Every finding names a correction, and no correction enlarges this task.
- [ ] The verdict is a status the evidence supports.
