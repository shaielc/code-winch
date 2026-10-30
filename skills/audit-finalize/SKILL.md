---
name: audit-finalize
description: Finalize the audit of an implementation-branch PR into its task branch by checking the audit is current, obtaining a person's verdict for every finding, and recording the verdicts and any fix requests on the PR. Use after an audit and before the PR is merged or its findings propagated. It does not merge.
---

# Audit finalize

Take an open **implementation branch → task branch** PR. Read applicable
`AGENTS.md`, its task brief, current diff, reviews, audit reports, earlier
verdict records, and decisions. Use
[stack reconstruction](../propagate-task/references/task-merge-stack.md) to
identify the actual branches, merged refinement, implementation head, task head,
and audit lineage. This operation decides and records; it does not merge.

## Check the audit

1. Verify the PR is open, belongs to the specified task, and targets its task
   branch. Do not select the task-branch → default-branch PR or a refinement or
   propagation PR. If selection is ambiguous, obtain the intended PR first.
2. Find an independent audit that explicitly covers this implementation PR's
   **current head SHA**. Read its full findings and later reports or decisions.
   An approval, resolved thread, passing CI, or audit of an older head is not
   a current audit. **If there is no current audit, stop and report it.** A
   partial audit covers only the checks it ran; each unrun check needs a later
   audit at the same head or a person's recorded acceptance of the gap.
3. **A Blocking finding is not cleared by a verdict.** Report each blocker and
   the condition for a new audit. Do not downgrade it because a person accepted
   a different finding or a follow-up issue exists.
4. Inventory every Deferrable, Defect, and Harden finding individually. Preserve
   classification and audit state; human decisions establish merge readiness,
   not a rewritten audit.

## Obtain finding verdicts

For **each Deferrable, Defect, and Harden finding**, summarize the observed
problem, effect, audit evidence, and the decision needed. Use the
available question tool (`request_user_input` when exposed) to ask for a verdict.
Give every finding its own question; batch only up to the tool's question limit.

- Deferrable: **Handle now**, **Defer to a task/issue**, or **Reject**.
- Defect: **Fix now**, **Accept and track**, or **Reject**.
- Harden: **Track as issue** or **Reject**.

**Reject** means the finding is invalid: the person disputes it, or the model,
skills, or `AGENTS.md` already settle it. Ask for the reason; a Reject without
one is not yet a verdict. A rejected finding needs no action and gets no
follow-up; the audit record keeps it unchanged.

Show concrete scope/impact in the option descriptions. Reuse an explicit prior
human verdict for the same finding and audited revision; show its source instead
of asking it again. A general approval, silence, elapsed time, issue creation,
or an auditor's recommendation is not a verdict. If the question tool is
unavailable, ask the finding-specific question in conversation. Without a
verdict, continue independent checks and report the unresolved finding.

## Record

Post one verdict record in the PR discussion: each finding with its audit and
head, decision, person, reason for any Reject, and task or issue destination.

For each **Fix now** or **Handle now**, post the fix request on the PR as the
person's decision: the finding, what the person decided, and any mechanism or
constraint they named. The implementer works from that comment. The fix is a
new head, so it needs a new audit and a new finalization; a promised fix is not
done.

Accepted and deferred follow-ups, and tracked Harden findings, go to
`propagate-task`, which opens issues only for those tracked.

## Merge readiness

The PR is ready to merge when every Defect, Deferrable, and Harden finding has a
recorded verdict, no Fix now or Handle now is outstanding, no Blocking finding
stands, every unrun check has a later audit or a recorded acceptance, and the
repository's merge checks pass. Report audit state separately from merge
readiness.

A person merges it, by hand or by explicitly asking an agent to. An agent asked
to merge first confirms these conditions at the current head, then merges into
the verified task branch with the repository's supported method and an
expected-head-SHA check. It does not force a protected merge or merge the task
branch into the default branch.

## Done

Report the verdict record, any fix requests, the findings awaiting propagation,
and whether the PR is ready to merge or what remains: a missing or stale audit,
Blocking findings, unresolved verdicts, outstanding fixes, or failed merge
checks. Leave tracker statuses unchanged.
