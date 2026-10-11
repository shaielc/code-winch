# Reconstruct the task merge stack

Use repository evidence to identify the task integration branch and the PRs
feeding it. Branch names alone do not establish a PR's role.

1. Resolve the repository, task ID, source implementation PR, and its actual
   base branch. Check the tracker entry and task marker. The implementation
   PR's base is the task branch; confirm the task branch's outgoing PR and
   intended integration target. Do not substitute the default branch.
2. Fetch current refs or read their commit objects through the repository API.
   Record the implementation head SHA and task-branch head SHA. Read the brief
   from those revisions, not an unrelated local checkout.
3. List **all states** of PRs targeting the task branch, including refinement,
   implementation, and earlier propagation PRs. Read their bodies, commits,
   reviews, discussions, and merge results. Follow linked prerequisite PRs.
4. Reconstruct the actual lineage. Check first-parent history and ancestry for
   merge commits; use the recorded merge result and corresponding commit/diff
   for squash or rebase merges. An open PR is not on the task branch yet. A
   closed, unmerged PR is not completed work. Do not merge or cherry-pick a PR
   merely to reconstruct the stack.
5. Associate each audit with its implementation PR and audited head SHA. Collect
   findings and human dispositions, with links and dates. Superseding reports
   must explicitly cover the same finding; thread resolution or PR approval
   alone does not establish a disposition.
6. Produce this record before deciding what to propagate or merge:

   | Role | PR | Head branch / SHA | Base branch / SHA | State / merge result | Audit and disposition links |
   |---|---|---|---|---|---|
   | Refinement, implementation, or propagation | URL | Actual ref and commit | Actual ref and commit | Open, merged, or closed unmerged | Report, audited SHA, and decisions |

Record the outgoing task-branch PR separately. Distinguish code already on the
branch from unmerged proposals. If the source or task branch is ambiguous,
report the candidates and obtain the missing selection before any write.

Immediately before creating a propagation branch or merging an implementation,
refresh both refs and the relevant PR states. New implementation commits need
an audit covering the new head. Changes on the task branch require rechecking
the combined diff, collisions, and required integration checks.
