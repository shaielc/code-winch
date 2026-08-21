---
name: task
description: Carry one workplan task from its brief to a finished, demonstrated change, or judge whether a task that claims completion actually holds. Use when picking up a task ID from docs/workplan/, implementing or finishing the change for one, reviewing the pull request that implements one, or auditing a single task's acceptance criteria against HEAD. Keeps attention on that one brief — its shape, its declared surfaces, its demonstration, and its verification.
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

1. **Implement** — build the change the brief describes, demonstrate it, and
   open the pull request.
2. **Audit** — judge one task against HEAD and report whether its acceptance
   criteria hold.

## Orient first, in either mode

1. **Read the brief end to end**, including **Non-goals**. Note its **Shape** —
   the shape determines what counts as a demonstration.
2. **Read the design sections under Traces to.** They are the authority the
   brief is derived from; where the brief and the design set disagree, that is a
   finding, not a choice.
3. **Check the dependencies.** Every ID under **Dependencies** is `completed` in
   `tasks.json`. A `completed` status is a claim: if the type, contract, or
   behavior the edge names is missing at HEAD, the dependency is not done. Stop
   and report that rather than building the missing piece inside this task.
4. **Read the Write set and Contract surfaces.** The write set is this change's
   boundary, and other tasks are in flight against the same tree. The contract
   surfaces are what only this task may redefine.
5. **Run the demonstration commands as they stand.** Seeing the current behavior
   before the change is what makes the change observable afterwards.

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

**Stay inside the declared surfaces.** A write outside the write set is usually
scope creep; where one is genuinely unavoidable, make it minimal and say why in
the pull request. Changing a contract the brief does not declare is the more
serious deviation — another task owns that surface and is working against the
meaning you changed. Report it rather than absorbing it.

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
  owning task ID, and any write outside the write set or change to an undeclared
  contract surface, with its reason.
- Leave status fields in `docs/workplan/tasks.json` alone. Automation stamps
  `completed` when the pull request is approved.

## Auditing one task

Judge the task against HEAD, and against the code — not against the pull request
description, the commit messages, or the brief's own optimism. Statuses have
been found wrong in practice, marked complete when acceptance criteria were not
met.

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
5. **Write set and contract surfaces** — the change stayed inside the declared
   write set and touched no contract the brief does not declare. Any deviation
   is named and justified.
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

When the cause is the brief rather than the code — a seam between two briefs, an
assumption one brief made about another's output, an invariant no task owns —
name it as a plan defect and say which briefs are involved.
`docs/workplan/post-mortems/` is where those are recorded. When the cause is
instead the code of an already-completed task, the remedy is a revision task
against it, which belongs to the `workplan` skill rather than to this one.

## Before finishing

- The shape's demonstration was produced, not just its tests.
- The demonstration was run by hand and the observed output recorded.
- **Runtime reachability** names a configuration that truly reaches the code.
- The repository gates and every check under **Verification** pass, or the
  exceptions are stated.
- Every deferral names a task ID that exists in `tasks.json`.
- Writes stayed inside the write set; no undeclared contract surface changed.
- Contract changes carry their design document or ADR update.
- The system starts and deploys at HEAD.
- `Task: <ID>` is in the pull request body; no status field was edited.
