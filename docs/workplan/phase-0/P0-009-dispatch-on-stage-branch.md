# P0-009: Dispatch Refine and Implement on a panel-created stage branch

**Phase:** 0 — Plan the implementation phases
**Shape:** capability
**Dependencies:** None

First of four tasks that move the Refine → pull request → review/audit → return
loop from manual steps into the workplan control panel; P0-010, P0-011 and
P0-012 follow it in that order. This is a Phase 0 meta task because the control
panel is repository tooling, not the product, and it is the same kind of change
as P0-008's toolchain work. The shape is the nearest fit: it adds observable
behavior to the panel, reached by its HTTP API and UI.

## Objective

Launching Refine or Implement for a task pushes `refine/<ID>` or `implement/<ID>`
from the current `task/<ID>` tip and runs the Codex Cloud task on that branch,
so every submitted stage has a recorded branch and base revision that the panel
owns without rewriting a branch that already carries an open pull request.

## Scope

Today `TaskScheduler.stage` fetches the `task/<ID>` tip, uses that tip as the
attempt's `head`, persists `submitting`, and calls `codex.submit` with the task
branch. The two Codex prompt templates then ask the agent to create a branch,
commit, and open a pull request. Make that flow concrete in this order:

1. In `integrations/repository.py`, add a single stage-branch naming helper for
   Refine and Implement and a preparation helper that fetches `task/<ID>`, lists
   open pull requests into that task branch, and creates or replaces
   `<stage>/<ID>` at the fetched task tip. Use the existing `github_run`
   credential boundary for remote reads and pushes. Push a new branch normally;
   replace an existing branch only with a lease on its fetched tip so concurrent
   movement fails rather than being overwritten. If an open pull request names
   that stage branch as its head, return a conflict instead of rewriting it;
   this preserves the branch that review is using. A branch without an open
   pull request is stale and may be replaced, including after stage expiry.
2. In `TaskScheduler.stage`, keep Audit on its current Claude/pull-request path.
   For Refine and Implement, prepare the stage branch before appending the
   persist-before-act `submitting` attempt, then call `codex.submit` with that
   branch. Continue using the fetched task tip as `head`, because
   `state.live_attempt` already makes that the idempotence revision, and add
   `branch` plus `base` (the same task-tip commit) to every new Codex attempt.
   A second launch at that base returns the existing attempt before doing branch
   work. If the task tip moved, create a new attempt from the new tip only when
   the fixed stage branch can safely be replaced; report a conflict while its
   pull request is open. Translate preparation failures into a `TaskError` that
   names the task and stage, without appending an attempt or invoking Codex.
3. Change `prompts/refine.md` and `prompts/implement.md` to say that the agent is
   already on the panel-created stage branch and must change files and commit
   there, but must not create or switch branches and must not open a pull
   request. A commit remains required: P0-010 will fetch the resulting diff and
   apply it to this branch. Update the prompt rendering assertions.
4. Extend whole-task expiry to delete `refine/<ID>` and `implement/<ID>` as well
   as `task/<ID>` before releasing the reservation. Treat absent branches as
   idempotent, as `delete_task_branch` does. Preserve the reservation and stage
   records if any deletion fails. Stage-only expiry continues to retire only
   the attempt; its next launch replaces the unreviewed stage branch from the
   current task tip.
5. Leave `rebase_task_branch`'s generic open-PR replay in charge of stage
   branches: `task_pull_requests` already supplies every open PR into the task
   branch, so no stage-specific rebase path is needed. Add coverage proving an
   open `refine/<ID>` head is replayed and never overwritten by dispatch.
6. Render each attempt's recorded branch in both the table stage cell and the
   dependency-tree stage details. Older state records have no `branch`, so the
   UI must retain a harmless fallback rather than requiring a state migration.
   Update the README flow and expiry descriptions to match the panel-owned
   branch lifecycle.

## Required constraints

- No Codex task is submitted when the branch cannot be created or pushed, and
  no stage attempt is appended; the failure names the task and stage operation.
- The persist-before-act idempotence, the branch-revision scoping, and the
  rule that `GH_TOKEN` never enters the `codex` environment keep holding; the
  existing credential-boundary test covers the new dispatch path.
- A stage branch with an open pull request is never rewritten by dispatch or
  expiry; only the existing Rebase operation may replay it.
- The Audit stage, its routine, and the task-to-main pull request flow are
  unchanged.

## Non-goals

Fetching or applying Codex's diff, publishing it to the prepared branch,
opening stage pull requests, starting reviewers, and returning reports;
best-of-N attempts; changing Audit or the Claude routine; changing the fixed
`<stage>/<ID>` naming contract to support concurrent attempts. Publication,
review/audit, and report return belong to the existing owners in Deferrals.

## Runtime reachability

`POST /api/tasks/<ID>/refine` and `.../implement`, and the UI's Refine and
Implement controls, in the panel started by `python3 -m control_panel` or
`workplan-control-panel/compose.yml`.

## Write set

- `workplan-control-panel/control_panel/task_scheduler.py`
- `workplan-control-panel/control_panel/integrations/repository.py`
- `workplan-control-panel/control_panel/prompts/refine.md`, `implement.md`
- `workplan-control-panel/control_panel/ui.py`
- `workplan-control-panel/tests/`
- `workplan-control-panel/README.md`

## Contract surfaces

The stage-branch names `refine/<ID>` and `implement/<ID>`; the stage attempt
record's `branch` and `base` fields (with existing `head` retained as the
idempotence revision); the Refine and Implement prompts' branch, commit, and
pull-request instructions; the panel README's description of the flow. P0-010,
P0-011 and P0-012 consume these. The state format remains schema version 2
because attempt records are already extensible mappings and old records remain
readable.

## Demonstration

Run the panel test suite, whose scheduler fixture uses a local bare remote and a
fake `codex`, and inspect the captured dispatch and state:

    $ POST /api/tasks/<ID>/refine
    → expect: origin has refine/<ID> at the same commit as task/<ID>; the fake
      codex received `--branch refine/<ID>`; the prompt has no instruction to
      create or switch branches or open a pull request; the attempt records
      branch `refine/<ID>` and base equal to the task tip
    $ POST /api/tasks/<ID>/refine        (same tip)
    → expect: the existing submission is returned; no second branch or task
    $ make the branch push fail, then POST /api/tasks/<ID>/implement
    → expect: an error naming the task and Implement operation; no attempt is
      appended and the fake codex receives no call
    $ expose refine/<ID> as the head of an open PR, move task/<ID>, then POST
      /api/tasks/<ID>/refine
    → expect: conflict; refine/<ID> is unchanged and codex receives no call

Optionally, with a real Codex environment: the dispatched task opens on
`refine/<ID>` and its diff is against that branch's tip.

## Verification

- `cd workplan-control-panel && python3 -m unittest discover -s tests`, per
  `docs/testing.md`.
- Negative cases: branch creation failing submits nothing; `GH_TOKEN` is absent
  from the `codex` child environment; a second launch at the same tip neither
  pushes nor submits; an open-PR stage branch is not rewritten; a lease failure
  leaves the remote branch and attempt history unchanged; whole-task expiry
  preserves its reservation if stage-branch deletion fails.
- Say what could not run here, including any real Codex or GitHub use.

## Acceptance criteria

- [ ] Refine and Implement each push `<stage>/<ID>` from the current task tip
      before dispatch and pass it to `codex cloud exec --branch`.
- [ ] Each new Codex attempt records `branch` and `base`, with `base` and the
      attempt's idempotence `head` equal to the fetched task tip; records written
      before this change still load and render.
- [ ] The Refine and Implement prompts no longer ask the agent to branch,
      switch branches, or open a pull request, and still require a commit on the
      already-selected stage branch.
- [ ] A relaunch after the task tip moved replaces an unreviewed stage branch
      from the new tip; if that branch heads an open pull request, dispatch
      refuses without rewriting it or invoking Codex.
- [ ] Expiring the task removes its stage branches; Rebase still replays open
      stage pull requests onto the rewritten task branch; failure to delete any
      branch does not release the reservation or expire its attempts.
- [ ] A failed or lease-rejected stage-branch push leaves no new attempt and no
      Codex submission; the returned error identifies the task and stage.
- [ ] The table and dependency-tree stage details show the recorded branch for
      new attempts and remain usable for legacy attempts without that field.

## Deferrals

| Deferred | Owning task |
|---|---|
| Fetching the finished task's changes, pushing them, and opening the pull request | P0-010 |
| Starting the review-refinement or audit session on that pull request | P0-011 |
| Passing reports back to the implementing agent | P0-012 |

## Traces to

`workplan-control-panel/README.md` (flow and API);
`workplan-control-panel/codex-interfaces.md`;
`workplan-control-panel/codex-cloud-cli.md`; ADR-0004;
`docs/testing.md` (Control panel gate);
`skills/shared/workplan-model.md` (Phase 0 meta tasks).
