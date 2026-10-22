---
name: implement-task
description: Implement, demonstrate, verify, and open the pull request for one refined task brief. Use when building or finishing a specific workplan task under its accepted scope.
---

# Implement task

For a planning shape, produce the artifacts using its assigned phase/task-author
skill and the model's planning criteria. Verify outputs and handoff; runtime
reachability and new-capability demonstrations belong to the target
implementation task, not this planning task.

Work on one task ID. Read its brief end to end, the phase README, index,
tracker, design sections under Traces to, [the model](../shared/workplan-model.md),
and applicable `AGENTS.md` files. The brief owns scope and acceptance; the tracker
owns status and dependencies. Use a separate author for each task operation:
derive, refine, implement, and audit.

## Orient

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

## Build the change

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
supported way to run the product. Treat them as malleable: change their behavior
or implementation to exercise the task, updating callers, tests, and
documentation together. State what the
fake does not prove. Its existing dialect and controls do not add product scope.

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
write it. If no such task exists, finish the work within scope or hand the missing
work to the phase/task authors for an explicit decision and derived owner. Do
not create a placeholder or broaden this task silently. Acceptance criteria
are not satisfied by code that defers them.

Where what you found is that an implementation another task produced needs
reworking, the separately derived task takes a `revision` edge to that task and
its brief says what was found. That is the reason `revision` exists — it is written when
the problem is discovered, not when the plan is drawn.

### Verify

Run, in this order:

1. **Everything under the brief's Verification.** That is the minimum evidence
   expected in the pull request, not a ceiling.
2. **The Demonstration, by hand, exactly as written.** Record the commands and
   what you actually saw, not what the brief predicted. If following it literally
   would change a deployment you do not own, run the same check where the
   project's deployment and test instructions say demonstrations run, record
   the substitution, and report the brief's command as a plan finding.
3. **The repository's gates**, as `docs/testing.md` specifies, including the
   standing scenario suite — unchanged, against the profile or substrate the
   brief names.

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

## Review rounds

When a pull request goes through rounds with a reviewer, post each request on
the pull request, as a review or a comment, and answer it there. Commits record
what a round changed. Only the requests record why, and a later audit or
post-mortem that cannot see them is left guessing.

If the same concern changes direction twice, stop the rounds and escalate to a
person with both positions. An example is one refusal passing through three
different codes. Another round will not settle a disagreement the brief does
not settle.

## Optional handoffs

Use `skills/propagate-task/SKILL.md` to route findings to tasks or issues through
a separate PR into the task branch, and `skills/reflect-task/SKILL.md` to examine
the development methodology.
These are optional operations, not delivery gates.

## Done

The shape's demonstration is observed by hand; the code is reachable; gates and
brief verification pass or their limits are reported. Contract changes have
matching design updates, deferrals have real owners, and the PR names only
`Task: <ID>`. Status fields remain untouched. A different auditor uses
`skills/audit-task/SKILL.md` to judge the result.
Use `skills/task-implement-merge/SKILL.md` to handle finding verdicts and merge
that audited implementation PR into the task branch.
