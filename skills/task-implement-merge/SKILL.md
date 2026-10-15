---
name: task-implement-merge
description: Merge an open implementation-branch PR into its task branch after checking a current audit and obtaining explicit human verdicts for every Defect and Deferrable finding. Use for implementation PR merge readiness, not for merging a task branch into the default branch.
---

# Task implement merge

Take an open **implementation branch → task branch** PR. Read applicable
`AGENTS.md`, its task brief, current diff, reviews, audit reports, and decisions.
Use [stack reconstruction](../propagate-task/references/task-merge-stack.md) to
identify the actual branches, merged refinement, implementation head, task head,
and audit lineage. This operation merges only that implementation PR.

## Check the audit

1. Verify the PR is open, belongs to the specified task, and targets its task
   branch. Do not select the task-branch → default-branch PR or a refinement or
   propagation PR. If selection is ambiguous, obtain the intended PR first.
2. Find an independent audit that explicitly covers this implementation PR's
   **current head SHA**. Read its full findings and later reports or decisions.
   An approval, resolved thread, passing CI, or audit of an older head is not
   a current audit. **If there is no current audit, refuse to merge.** A
   partial audit covers only the checks it ran; each unrun check needs a later
   audit at the same head or a person's recorded acceptance of the gap.
3. **If the current audit has any Blocking finding, refuse to merge.** Report
   each blocker and the condition for a new audit. Do not override or downgrade
   it because a person accepted a different finding or a follow-up issue exists.
4. Inventory every Deferrable, Defect, and Harden finding individually. Preserve
   classification and audit state; use human decisions to establish merge
   readiness, not to rewrite the audit.

## Obtain finding verdicts

For **each Deferrable, Defect, and Harden finding**, summarize the observed
problem, effect, audit evidence, and the decision needed. Use the
available question tool (`request_user_input` when exposed) to ask for a verdict.
Give every finding its own question; batch only up to the tool's question limit.

- Deferrable: **Handle before merge**, **Defer to a task/issue**, or **Reject**.
- Defect: **Fix before merge**, **Accept and track**, or **Reject**.
- Harden: **Track as issue** or **Reject**.

**Reject** means the finding is invalid: the person disputes it, or the model,
skills, or `AGENTS.md` already settle it. Record the person's reason. A rejected
finding needs no action before merge and gets no follow-up; the audit record
keeps it unchanged.

Show concrete scope/impact in the option descriptions. The tool is for product
tradeoffs; invocation of this skill already authorizes the merge once its
conditions hold. Do not ask a second generic merge-permission question.

Reuse an explicit prior human verdict for the same finding and audited revision;
show its source instead of asking it again. A general approval, silence, elapsed
time, issue creation, or an auditor's recommendation is not a verdict. If the
question tool is unavailable, ask the finding-specific question in conversation.
Without a verdict, continue independent checks and report the unresolved
finding; do not merge.

Record each verdict with its finding, audit/head, decision, person, and task or
issue destination in the PR discussion. If the decision is to fix/handle before
merge, wait for the actual change and a current audit; do not treat a promised
fix as done. For accepted/deferred follow-ups, use `propagate-task` to select an
appropriate existing task or open an issue and produce its propagation PR.
Link that concrete destination before treating the disposition as recorded.
An issue does not clear a Blocking finding or fulfill missing task scope.

Harden findings never block this merge, but each needs a verdict before
propagation. `propagate-task` opens issues only for those tracked.

## Merge

1. Confirm every Defect, Deferrable, and Harden finding has a recorded verdict
   and every required pre-merge action is satisfied. Report audit state separately from merge readiness.
2. Refresh the PR head, task branch, audit, decisions, and required checks.
   Require the repository's merge checks and branch rules to pass. New head
   commits require a new current audit; task-base changes require rechecking
   the combined diff and integration checks. Resolve conflicts through the
   implementation workflow, then obtain an audit covering the resulting head.
3. Merge the implementation PR into its verified task branch, using the
   repository's supported merge method and an expected-head-SHA check. Do not
   force a protected merge, enable a later unaudited merge, or merge the task
   branch into the default branch.
4. Verify GitHub reports the PR merged and the task branch contains the merge
   result. Report the PR, merge commit, audit, and verdict/destination links.
   Include the verdict record and the findings awaiting propagation.
   Leave tracker statuses unchanged.

## Done

The audited implementation PR is merged into its task branch with a recorded
verdict for every finding and all required actions done. Otherwise report precisely why it remains
unmerged: missing/stale audit, Blocking findings, unresolved dispositions,
incomplete required fixes, or failed repository merge checks.
