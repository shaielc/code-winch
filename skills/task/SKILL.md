---
name: task
description: Refine, implement, or audit one task of the implementation workplan in docs/workplan/. Use when a prompt is about a single task — dispatching one, naming its ID, working on its brief, building or finishing the change that implements it, reviewing that pull request, or judging its claim to be complete. Keeps attention on that one brief — its shape, its declared surfaces, its demonstration, and its verification. Plan-level work belongs to the workplan skill.
---

# Task

Everything here is scoped to one task ID. The brief at
`docs/workplan/<phase>/<id>-*.md` is authoritative for scope, acceptance
criteria, and required verification; `docs/workplan/tasks.json` is authoritative
for status and dependencies.

Read `skills/shared/workplan-model.md` for the invariants a finished task
preserves, the four shapes and what each one must demonstrate, and the tracker
rules. Read the root `AGENTS.md` and any `AGENTS.md` in a subtree you touch;
they carry the repository's boundaries and its verification commands.

## Modes

1. **Refine the brief** — before the task is implemented, check its brief
   against HEAD and the plan, make it concrete enough to implement from, and
   commit it to the task's branch before any implementation.
2. **Implement** — build the change the brief describes, demonstrate it, and
   open the pull request.
3. **Audit** — judge one task against HEAD and report whether its acceptance
   criteria hold.

The three are separate steps. Whoever refines a brief does not also implement
or audit that task.

## Orient first, in any mode

1. **Read the brief end to end**, including **Non-goals**. Note its **Shape** —
   the shape determines what counts as a demonstration.
2. **Read the design sections under Traces to.** They are the authority the
   brief is derived from; where the brief and the design set disagree, that is a
   finding, not a choice.
3. **Check the dependencies.** Every ID under **Dependencies** is `completed` in
   `tasks.json`. A `completed` status is a claim: if the type, contract, or
   behavior the edge names is missing at HEAD, the dependency is not done. Stop
   and report that rather than building the missing piece inside this task.
4. **Read the Write set and Contract surfaces.** The write set forecasts which
   files the change touches, so tasks in flight can predict a rebase; it is not
   a boundary. The contract surfaces are what only this task may redefine.
5. **Run the demonstration commands as they stand.** Seeing the current behavior
   before the change is what makes the change observable afterwards.

## Refining a brief

A brief is written before the code it builds on exists, so much of it is a
prediction. Refine it when the task is dispatched, before any implementation.
By then the code is at HEAD, and this step has context the planner did not.

Check, with evidence — a command with its observed output, or a `file:line`:

1. **Assumptions about the existing code** — the type, contract, or behavior
   each dependency edge names exists at HEAD. So do the interfaces, files, and
   behavior the Scope relies on without an edge, in the form the brief assumes.
2. **Deferrals into this task** — search every brief's Deferrals table, every
   phase's *Deferrals in* register, and the tree for this task's ID. Carry each
   hit into Scope, or add a Scope item that gives it a different owner.
3. **Objective against criteria** — every guarantee the Objective states, such
   as durability, fencing, recovery, or no orphaned processes, ends one of three
   ways. Acceptance criteria name the failures that would break it and what a
   person observes in each. Or it is narrowed out of the Objective. Or a Scope
   item gives it to a hardening task.
4. **Demonstration** — the tools and existing commands it uses are present where
   it will run, and each expectation can be observed without false matches.

Then make the brief concrete, in the sections it already has:

- **Scope** — implementation instructions written against the code at HEAD:
  where the code lives, how it connects to what exists, and the order where
  order matters, with a short reason for each judgment call. Work beyond this
  task's own code is a Scope item too, for the implementation to carry out. That
  includes a hardening task to add, and a deferral to re-own in another brief.
- **Non-goals** — what the refinement explicitly excludes.
- **Write set** — the forecast, updated from the concrete Scope.
- **Acceptance criteria**, **Verification**, and **Demonstration** — the
  results of checks 3 and 4, and the test that will show each criterion.
- **Deferrals** — rows only for owners that already exist in `tasks.json`. An
  owner the implementation creates gets its row in that change.

Refine toward what the design set and the code at HEAD require, never toward an
implementation that already exists.

The artifact changes only the brief, and is committed to the task's branch
before any implementation. Implementation starts from it.

## Implementing

**Leave the system runnable and deployable.** Every commit starts, and deploys,
and the standing scenario suite stays green. This binds during the task, not
only at the end of it.

**Make the code reachable in this task.** A seam is not finished by defining a
port and an implementation: register it in the composition root and make it
reachable, so the **Runtime reachability** section describes something true.
Code no runtime configuration reaches is not finished, however well it is
tested. Where the capability is operator-visible, reachable means through the
maintained CLI: `curl` and a database client prove the behavior exists, not that
it is operable (I5).

**Deliver what the shape promises.**

- *Seam* — the standing scenario passes with the new seam in the path, and a
  person drives the new behavior by hand through the maintained CLI.
- *Swap* — the standing scenario passes unchanged against the new substrate,
  and the fake profile still passes too.
- *Capability* — a new scenario, added to the standing suite.
- *Hardening* — the attack or failure is attempted and visibly refused.
  Negative and adversarial tests, not happy paths.

**Keep fakes shipped-quality.** In-memory and fake implementations are a
supported way to run the product. Keep them controllable — scripted transcripts,
injectable latency, failure, malformed output — and state in their documentation
what they do not prove.

**Stay inside the declared scope and contract surfaces.** Work the brief's
Scope does not ask for, or its Non-goals exclude, is scope creep. Changing a
contract the brief does not declare is the more serious deviation — another
task owns that surface and is working against the meaning you changed. Report
it rather than absorbing it. Files written outside the write set are neither:
list them in the pull request so the plan's write-collision report can be
updated.

**Update the design set with contract changes.** An API path, event or protocol
schema, port signature, or migration that changes takes its design document
update — or a new or superseding ADR — in the same change.

**Defer nothing without an owner.** No TODO, stub, unimplemented branch, or
"handled later" without a task ID that exists in `tasks.json` at the moment you
write it. If no such task exists, either finish the work or add the task and say
so in the pull request. Acceptance criteria are not satisfied by code that
defers them.

Where what you found is that an implementation another task produced needs
reworking, the task you add takes a `revision` edge to that task and its brief
says what was found. That is the reason `revision` exists — it is written when
the problem is discovered, not when the plan is drawn.

### Verify

Run, in this order:

1. **The repository's standard gates**, as `AGENTS.md` specifies — including the
   integration and web gates when the change touches what they cover. If a gate
   could not run in the environment, say which and why rather than reporting a
   pass.
2. **Everything under the brief's Verification.** That is the minimum evidence
   expected in the pull request, not a ceiling.
3. **The standing scenario suite**, unchanged, against the profile or substrate
   the brief names.
4. **The Demonstration, by hand, exactly as written.** Record the commands and
   what you actually saw, not what the brief predicted.

If the demonstration cannot produce the stated result, either the change is
incomplete or the brief no longer matches the system. Correct the brief in the
same change and say what changed and why; do not quietly reword it to match
whatever happened.

### Finish

- The pull request body contains `Task: <ID>` and no other task ID.
- Report what you ran, what the demonstration showed, every deferral with its
  owning task ID, any change to an undeclared contract surface with its reason,
  and the files written outside the write set.
- Leave status fields in `docs/workplan/tasks.json` alone. Automation stamps
  `completed` when the pull request is approved.

## Auditing one task

Judge the task against HEAD, and against the code — not against the pull request
description, the commit messages, or the brief's own optimism. Statuses have
been found wrong in practice, marked complete when acceptance criteria were not
met.

The audit judges the task against its brief as written. Whether the brief asks
for the right things was settled when the brief was refined. The audit does not
re-derive the objective's guarantees or grade the task against them.

Produce evidence for each line below. Evidence is a command with its observed
output, or a `file:line`. A claim with neither is unmet.

1. **Acceptance criteria** — one at a time, in order. No summarizing several
   into one verdict.
2. **Demonstration** — run it as written. If it does not produce the stated
   result, the task is not complete whatever the test suite says.
3. **Runtime reachability** — find the registration in the composition root and
   the profile or command that reaches the code. A port whose only
   implementations are test doubles, or an implementation nothing constructs, is
   an orphan.
4. **Verification** — every check the brief lists exists and passes now, not
   only when the pull request merged.
5. **Contract surfaces** — the change touched no contract the brief does not
   declare, or names and justifies each one it did. List the files written
   outside the write set for the plan's write-collision report; they are not a
   deviation to judge.
6. **Deferrals** — each TODO, stub, unimplemented branch, and unreachable error
   path the task introduced carries an ID that exists in `tasks.json`. Search
   the diff for `TODO`, `unimplemented`, "for now", "later", and "subsequent
   task".
7. **Shape** — the task delivered what its shape promises, per *Implementing*
   above.
8. **Invariants** — the seven hold at HEAD with this task in. Start the system,
   run the fake profile by hand, run the standing scenario suite.

The verdict is **complete** or **not complete**; there is no partial. Not
complete lists the specific criteria that fail and what would satisfy each.

Report defects found outside the lines above as separate findings, each with
its evidence. Examples are a failure path no criterion names, or a gap between
the objective and the criteria. They do not change the verdict. A person decides
whether each one blocks the task, becomes its own task, or corrects the brief.

When the cause is the brief rather than the code — a seam between two briefs, an
assumption one brief made about another's output, an invariant no task owns —
name it as a plan defect and say which briefs are involved.
`docs/workplan/post-mortems/` is where those are recorded. When the cause is
instead the code of an already-completed task, the remedy is a revision task
against it, which belongs to the `workplan` skill rather than to this one.

The audit reports; it does not act on its findings. It does not create tasks,
add deferral rows, or edit the brief it judges in the same change as its
verdict. Those are plan changes, made separately once a person has accepted the
finding.

## Review rounds

When a pull request goes through rounds with a reviewer, post each request on
the pull request, as a review or a comment, and answer it there. Commits record
what a round changed. Only the requests record why, and a later audit or
post-mortem that cannot see them is left guessing.

If the same concern changes direction twice, stop the rounds and escalate to a
person with both positions. An example is one refusal passing through three
different codes. Another round will not settle a disagreement the brief does
not settle.

## Before finishing

- The shape's demonstration was produced, not just its tests.
- The demonstration was run by hand and the observed output recorded.
- **Runtime reachability** names a configuration that truly reaches the code.
- The repository gates and every check under **Verification** pass, or the
  exceptions are stated.
- Every deferral names a task ID that exists in `tasks.json`.
- No undeclared contract surface changed; files written outside the write set
  are listed in the pull request.
- A refinement changes only the brief, and lands before any implementation
  commit on the task's branch.
- Contract changes carry their design document or ADR update.
- The system starts and deploys at HEAD.
- `Task: <ID>` is in the pull request body; no status field was edited.
