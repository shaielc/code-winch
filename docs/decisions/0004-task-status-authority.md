# ADR-0004: Task status authority and scheduler state

- **Status:** Proposed
- **Date:** 2026-08-11

## Context

The task scheduler dispatches available workplan tasks to Codex Cloud when a
task pull request merges. It read the task graph from `origin/main` but recorded
completion in a local state file inside a Docker volume, overlaying that state on
the tracker at dispatch time. The overlay was unconditional, so a local entry
also overrode a status already recorded in `docs/workplan/tasks.json`.

That made the volume the durable record of project progress while
`docs/workplan/tasks.json` claimed to be the source of truth, and the two could
disagree. Losing the volume would lose every completion and re-dispatch finished
work. A local `in_progress` entry masked a real `completed` status and blocked
dependent tasks, and nothing expired an entry whose task never resolved to a
pull request, so a missed match consumed a concurrency slot permanently.

Completion and in-flight tracking were conflated despite opposite durability
needs: completion must survive indefinitely and be reviewable, while in-flight
state matters only between dispatch and merge.

## Decision

`docs/workplan/tasks.json` on the default branch is the sole authority for
`completed`. The scheduler no longer derives completion from merge events.

A status gate enforces this on every pull request. A check-only job requires each
pull request to resolve to exactly one known task ID from its title, body, or
branch name; pull requests labelled `no-task` are exempt. On approval, a second
job stamps `completed` into the tracker and pushes that commit to the pull
request branch, so the status change is reviewed and merged with the work that
earned it. The stamp requires the branch to already contain the default branch,
which guarantees the tracker being written contains every other recorded
completion and cannot revert one.

The scheduler's local state is demoted to in-flight leases only: which tasks are
currently dispatched, for concurrency accounting. Leases do not expire on a
timer; they are released manually through a control panel served from the runner
directory, which also exposes where a lease disagrees with the tracker.

The stamping job runs on GitHub-hosted runners rather than the self-hosted
runner, so the token that can write to the repository is never present on the
host that executes agent-authored code.

## Consequences

Losing the scheduler state volume now costs at most a few duplicate dispatches
instead of the entire completion history. `tasks.json` means what the workplan
README says it means, and status changes carry review history in git.

Concurrent completions of non-adjacent tasks auto-merge, since task entries sit
about ten lines apart and outside git's default merge context; adjacent task IDs
completing simultaneously can still conflict.

The repository must disable "dismiss stale pull request approvals when new
commits are pushed", because the stamp commit would otherwise invalidate the
approval that triggered it and the pull request could never merge.

Pull requests from forks cannot be stamped automatically and require a manual
tracker edit. Every non-task pull request needs the `no-task` label.

Because leases never expire on their own, a task whose pull request is abandoned
holds its slot until someone releases it. This is deliberate: silent automatic
recovery previously hid the matching failures that caused the problem.

## Alternatives

- Scheduler pushes status commits directly to the default branch: rejected
  because it requires write access and protection bypass on the host that runs
  agent-authored code, and needs push-race handling.
- Derive completion by querying merged pull requests and store nothing: rejected
  because the tracker would stay permanently `pending`, so a reader of the
  repository still learns nothing about progress.
- Move task tracking to GitHub Issues or Projects: rejected for now because it
  moves the source of truth out of the repository, though it remains the better
  fit if concurrent status mutation becomes painful.
- Expire leases on a timer: rejected because automatic recovery masks the
  ID-matching failures that produce stale leases.

## Revisit when

Tracker merge conflicts become routine rather than occasional, fork
contributions need automated stamping, or manual lease release becomes frequent
enough that the underlying matching failure should be fixed instead.
