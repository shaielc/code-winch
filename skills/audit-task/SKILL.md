---
name: audit-task
description: Audit one task implementation against its accepted brief and the current code. Use when reviewing a task pull request or judging its acceptance, reachability, verification, findings, and merge readiness.
---

# Audit task

Work on one task ID. Read its brief end to end, the phase README, index,
tracker, design sections under Traces to, [the model](../shared/workplan-model.md),
and applicable `AGENTS.md` files. The brief owns scope and acceptance; the tracker
owns status and dependencies. Use a separate author for each task operation:
derive, refine, implement, and audit.

Audit the requested commit independently of its refiner and implementer.

For a planning shape, judge artifact acceptance and handoff using the model's
planning criteria and the assigned author operation's requirements. Runtime
reachability and target-capability scenarios are implementation criteria;
planning completion does not certify the predicted behavior.

Before recording any finding, Harden included, check whether the model, the
skills, or `AGENTS.md` already settle the concern, and cite the rule if they do.
A criterion about an event outside the task — a later dispatch, another author's
operation — is a plan defect in the brief, not a finding against the work.

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

## Judge the task

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
2. **Demonstration** — run it as written. Record any discrepancy as a finding
   and classify it; passing tests do not replace an observed demonstration.
   If following it literally would change a deployment you do not own, run the
   same check where the project's deployment and test instructions say
   demonstrations run, record the substitution, and report the brief's command
   as a plan finding.
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
7. **Shape** — the task delivered what its shape promises, as defined by the
   shared model.
8. **Invariants** — the seven hold at HEAD with this task in. Start the system,
   run the fake profile by hand, run the standing scenario suite.

For every audit check above, report what holds and what does not, with
evidence. **Classify every finding**, including deviations from acceptance
criteria and demonstrations as well as findings outside the brief:

- **Blocking** — must be resolved before merge. State the condition to satisfy.
- **Defect** — real deviation from intended behavior. Requires explicit human
  disposition before merge: fix now or accept and track. The audit does not
  choose between these options.
- **Deferrable** — valid work that could reasonably move elsewhere or later.
  Requires human disposition before merge: handle now or assign a destination
  and owner. The auditor may suggest one but does not decide to defer it.
- **Harden** — recommended robustness, test-quality, or maintainability
  follow-up; no pre-merge action required.

Keep observed acceptance results distinct from classification. A failed
criterion is not automatically Blocking or an automatic not-complete verdict.
For example, rejected-button focus behavior may be a non-blocking Defect
requiring human disposition; test coverage weaknesses may be Harden; stable
identity for persisted submissions may be Deferrable if a person assigns it
downstream.

A check you could not run is a limit of this audit, not a finding against the
implementation. Under that check, say what did not run and why, and mark the
audit **partial**. Merge readiness then waits on a later audit at the same head
that runs the check, or on a person's recorded acceptance of the gap. Classify
only what you observed.

### Audit state

Assign one **audit state** from the classified findings, using this precedence:

- **Not Complete** — at least one Blocking finding exists.
- **Input Required** — no Blocking finding exists, but at least one Defect or
  Deferrable finding exists. A person must decide its disposition.
- **Complete** — no Blocking, Defect, or Deferrable findings exist. Harden
  findings may still be recorded.

Blocking takes precedence over Defect and Deferrable. **Harden never changes
the audit state.** Report the state and the evidence for each finding, without
conflating whether an acceptance criterion held with its classification.

### Merge readiness

**Audit state is not merge readiness.** Not Complete means Blocking findings
must be resolved before merge. Input Required means the audit cannot decide
how to handle a Defect or Deferrable finding; a person must record a disposition
for each before merge. They may fix a Defect now or accept and track it, may
handle Deferrable work now or assign a destination and owner, and may reject
either as invalid. Harden requires no pre-merge fix; a person tracks or rejects
it in `task-implement-merge` before `propagate-task` opens its issue.

After the human dispositions are recorded, a pull request may be merge-ready
even if its original audit state is Input Required. Keep that audit state and
the original findings as an accurate record; do not silently erase, downgrade,
or reclassify them because someone accepted the risk or assigned the work.
A subsequent audit may report a new state when the observed findings change.
Report the state and any outstanding merge-readiness conditions separately.
The audit reports; a person makes the product and process tradeoffs.

When the cause is the brief rather than the code — a seam between two briefs, an
assumption one brief made about another's output, an invariant no task owns —
name it as a plan defect and say which briefs are involved.
`docs/workplan/post-mortems/` is where those are recorded. When the cause is
instead the code of an already-completed task, report the need for a revision
task. A separate phase author decides any
boundary change and a task author derives that task.

The audit reports; it does not act on its findings. It does not create tasks,
add deferral rows, or edit the brief it judges in the same change as its
verdict. Those are plan changes, made separately once a person has accepted the
finding.

## Report

Give acceptance results, classified findings, audit state, and outstanding
merge-readiness conditions separately. Include observed commands or source
references. Make no implementation, brief, tracker, or phase edits as part of
this verdict. `propagate-task` routes findings and `reflect-task` examines the
methodology as separate operations. `task-implement-merge` obtains required
human verdicts and decides merge readiness without changing this audit.
