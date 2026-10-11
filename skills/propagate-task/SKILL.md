---
name: propagate-task
description: Route task audit findings to existing tasks or issues and open a propagation PR from a new branch into the source task branch. Use after a task audit or agreed discovery, reconstructing its merge stack and keeping finding dispositions explicit.
---

# Propagate task

Carry task findings to their follow-up owners. Read applicable `AGENTS.md`,
[the model](../shared/workplan-model.md), the source brief and implementation PR,
its audit and human decisions, the tracker, relevant task briefs, and issues.
Use [stack reconstruction](references/task-merge-stack.md) before making changes.

## Establish the source

Identify the source task branch from the implementation PR's actual base.
Reconstruct its refinement, implementation, and propagation PRs, including
merged changes, open proposals, audited commits, and recorded dispositions.
Do not assume an implementation is present just because its PR is open.

Give each finding a stable reference: source PR, audit, finding label, and
observed behavior. Preserve its original classification. Read recorded human
verdicts for Defect and Deferrable findings; tracking a finding does not decide
whether it may be deferred or accepted. Missing verdicts belong to
`skills/task-implement-merge/SKILL.md`.

## Route every finding

| Classification | Required action |
|---|---|
| Blocking | Report the source implementation's merge blocker; propagation does not clear it |
| Deferrable | Check for an appropriate existing task; otherwise open an issue |
| Defect | Check for an appropriate existing task; otherwise open an issue |
| Harden / hardening | Open an issue for each finding |

Check task suitability against its objective, non-goals, phase, contract ownership,
status, and prerequisite behavior. A nearby filename or similar title is not
ownership. A completed task cannot absorb new work without a separately derived
revision task. Do not create placeholder tracker entries.

For a suitable existing task, update its brief with the agreed follow-up,
acceptance observation, and source/disposition links. Keep scope, verification,
write forecast, and deferral records consistent. A change to its objective or
phase boundary needs the appropriate phase/task author, not silent propagation.
For work the verdict requires fixed in the source implementation, record that
current task as the owner; an issue is not evidence that the fix is complete.

When no suitable task exists, open an issue describing the finding, impact,
expected result, source evidence, original classification, human verdict (or
that it remains unresolved), and any later planning trigger. An issue tracks the
follow-up; it does not invent a task ID or satisfy the source task's completion.
For hardening, create an issue rather than extending a task by default.
Keep issue-only follow-ups in the ledger and PR discussion, not as invented
task owners in brief Deferrals tables or implementation TODOs.

Before creating an issue, check whether this exact audit finding was already
propagated. Reuse its recorded issue on reruns; do not duplicate it. This check
is not permission to skip a new hardening finding because a task looks related.

## Open the propagation PR

1. Refresh the source task branch and create a **new propagation branch from its
   current head**. Keep unmerged implementation changes separate; read them at
   their PR head rather than pulling them into the propagation branch.
2. Commit only propagation changes: affected briefs and a finding ledger. Store
   the ledger at `docs/workplan/propagations/<source-task>/<source-PR>-<audit>.md`,
   matching existing conventions. Record each finding's task or issue, verdict,
   source link, and the reconstructed stack. The ledger provides a reviewable
   change even when every finding routes to an issue.
3. Push the new branch and open a PR **into the source task branch**, not into
   the repository's default branch or a downstream task branch. Include the
   source task marker required by repository automation and links to its
   implementation, audit, decisions, destinations, and issues. Other task IDs
   belong in the ledger if the PR body allows only one task marker.
4. Verify the PR's actual head/base and diff, and report its URL with the routing
   ledger. Leave tracker statuses alone. Do not merge the propagation PR or the
   task branch as part of this operation.

## Done

Every Defect and Deferrable finding has a checked existing-task destination or
an issue; every hardening finding has an issue. A new-branch propagation PR
into the source task branch carries the ledger and any agreed brief changes.
Unresolved verdicts and blockers remain visible; no finding was silently
reclassified or treated as fixed merely because it was tracked.
