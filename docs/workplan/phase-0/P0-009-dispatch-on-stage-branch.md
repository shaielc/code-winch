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

Launching Refine or Implement for a task pushes `refine/<ID>-<N>` or `implement/<ID>-<N>`
from the current `task/<ID>` tip and runs the Codex Cloud task on that branch,
so every stage's work has a branch the panel owns.

## Scope

Today the stage prompts tell the agent to create its own branch and open its own
pull request, which in Codex Cloud means a person clicks "Create PR" in
Codex's web UI. Proposed approach, for the refiner to confirm against code at
dispatch:

- Before dispatch, create and push `<stage>/<ID>-<N>` at the task branch's tip,
  following the existing task-branch helper in `integrations/repository.py`,
  and pass it to `codex cloud exec --branch`.
- Record the stage branch and the task-branch tip it was created from on the
  stage attempt, beside the `head` the attempt is already scoped to.
- Rewrite `prompts/refine.md` so the agent refines the brief and commits it on
  the prepared `refine/<ID>-<N>` branch, without creating a branch or opening a
  pull request. Rewrite `prompts/implement.md` so the agent changes files only:
  no branch, commit or pull request. Update the prompt tests that assert the
  old wording.
- Number branches per task and stage, starting at 1. Every fresh submission,
  including after the task tip moves or the stage expires, increments the
  number and starts from the current task tip, regardless of open pull
  requests. Preserve earlier stage branches and their pull requests. A duplicate
  request with a live attempt at the same task tip returns that attempt instead.
- Expiring a task deletes all its numbered stage branches, as it already
  deletes the task branch. Rebase keeps carrying stage branches that have open pull requests.
- Show the stage branch in the stage cell and tree view.

## Required constraints

- No Codex task is submitted when the branch cannot be created or pushed, and
  the failure names the task and operation.
- The persist-before-act idempotence, the branch-revision scoping, and the
  rule that `GH_TOKEN` never enters the `codex` environment keep holding; the
  existing credential-boundary test covers the new dispatch path.
- A stage branch with an open pull request is never rewritten except by Rebase.
- The Audit stage, its routine, and the task-to-main pull request flow are
  unchanged.

## Non-goals

Fetching or publishing the agent's changes, opening stage pull requests,
starting reviewers, and returning reports; best-of-N attempts; any change to the
Claude routine. These belong to P0-010, P0-011 and P0-012 (see Deferrals).

## Runtime reachability

`POST /api/tasks/<ID>/refine` and `.../implement`, and the UI's Refine and
Implement controls, in the panel started by `python3 -m control_panel` or
`workplan-control-panel/compose.yml`.

## Write set

- `workplan-control-panel/control_panel/task_scheduler.py`
- `workplan-control-panel/control_panel/integrations/repository.py`
- `workplan-control-panel/control_panel/integrations/codex.py`
- `workplan-control-panel/control_panel/state.py` (only if the attempt record needs a schema change)
- `workplan-control-panel/control_panel/prompts/refine.md`, `implement.md`
- `workplan-control-panel/control_panel/ui.py`
- `workplan-control-panel/tests/`
- `workplan-control-panel/README.md`

## Contract surfaces

The stage-branch name `<stage>/<ID>-<N>`, where N increments per task and stage;
the recorded branch is the handoff to publication, review and report returns;
the stage attempt record's branch and base fields; the Refine and Implement prompts' instructions about branches and pull
requests; the panel README's description of the flow. P0-010, P0-011 and P0-012
consume these.

## Demonstration

Against a local bare remote with a fake `codex` on `PATH`, as the panel tests do:

    $ POST /api/tasks/<ID>/refine
    → expect: origin has refine/<ID>-1 at the same commit as task/<ID>; the fake
      codex received `--branch refine/<ID>-1`; the prompt has no instruction to
      create a branch or open a pull request
    $ POST /api/tasks/<ID>/refine        (same tip)
    → expect: the existing submission is returned; no second branch or task
    $ move task/<ID>, then POST /api/tasks/<ID>/refine
    → expect: refine/<ID>-2 starts at the new task tip; refine/<ID>-1 and any
      open pull request are unchanged
    $ expire Refine, then POST /api/tasks/<ID>/refine at the same task tip
    → expect: refine/<ID>-3 starts at that tip; earlier branches are unchanged
    $ make the branch push fail, then POST /api/tasks/<ID>/implement
    → expect: an error naming the task and operation; no codex task recorded

Optionally, with a real Codex environment: the dispatched task opens on
`refine/<ID>-<N>` and its diff is against that branch's tip.

## Verification

- `cd workplan-control-panel && python3 -m unittest discover -s tests`, per
  `docs/testing.md`.
- Negative cases: branch creation failing submits nothing; `GH_TOKEN` is absent
  from the `codex` child environment; a second launch at the same tip does not
  create a second branch.
- Say what could not run here, including any real Codex or GitHub use.

## Acceptance criteria

- [ ] Refine and Implement each push `<stage>/<ID>-<N>` from the current task tip
      before dispatch and pass it to `codex cloud exec --branch`.
- [ ] The stage attempt records the branch and the task tip it started from.
- [ ] The Refine prompt requires committing the refined brief on the prepared
      `refine/<ID>-<N>` branch, without creating a branch or opening a pull request.
      The Implement prompt no longer asks the agent to branch, commit or open
      a pull request.
- [ ] A relaunch after the task tip moved, or after stage expiry at the same
      tip, uses the next numbered branch from the current task tip, whether or
      not an earlier branch has an open pull request. Earlier branches and pull
      requests are preserved. A duplicate request with a live attempt at the
      same tip returns it without creating another branch.
- [ ] Expiring the task removes its stage branches; Rebase still replays open
      stage pull requests onto the rewritten task branch.
- [ ] A failed branch push leaves no submission recorded.
- [ ] Stage attempts recorded before this change still load and render in the
      stage cell and tree view without the new branch and base fields.

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
