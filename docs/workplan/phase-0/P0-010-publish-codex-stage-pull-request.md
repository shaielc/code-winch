# P0-010: Publish a finished Codex stage as a pull request

**Phase:** 0 — Plan the implementation phases
**Shape:** capability
**Dependencies:** P0-009 (contract: publishes onto the `<stage>/<ID>-<N>` branch and extends the stage attempt record that P0-009 defines)

Second of four control-panel automation tasks (P0-009 to P0-012). This is a
Phase 0 meta task because the control panel is repository tooling, not the
product; see P0-009. The shape is the nearest fit: it adds observable behavior to
the panel, reached by its API, UI and a background check.

## Objective

When a Refine or Implement Codex task finishes, its changes are on
`<stage>/<ID>-<N>` and a pull request into `task/<ID>` is open, without anyone
opening Codex's web UI.

## Scope

Use the stage attempt's recorded branch from P0-009 for publication and retries;
do not reconstruct a fixed branch name or allocate another number.

`codex cloud exec` returns only a task URL. The CLI can report a task's status
and print its unified diff, but it cannot read the agent's messages or session
history, so the pull request is described from what the panel itself knows
(`workplan-control-panel/codex-interfaces.md`). Proposed approach, for the
refiner to confirm against code at dispatch:

- Notice completion without an operator action by reading the task's status
  from `codex cloud list --json` (`pending`, `ready`, `applied`, `error`).
  `ready` or `applied` publishes; `error` records a failure; a finished task
  with an empty diff records that no changes were made.
- Fetch the task's diff with `codex cloud diff`, preserving its trailing
  newline, and apply it in a temporary worktree at the stage branch's tip. The
  panel's own clone is never the apply target.
- Make one commit authored by the panel, with subject `<ID>: <Refine|Implement>
  <task title>` and body `Task: <ID>`, push it to `<stage>/<ID>-<N>`, then open the
  pull request with `gh pr create --base task/<ID> --head <stage>/<ID>-<N> --fill`.
  A pull request already open for that head is reused, not duplicated.
- Record on the stage attempt the Codex task id, outcome, published tip, and
  pull request number and URL. Show the stage as running, published, failed or
  no-changes, with the reason, in the stage cell and tree view.
- Add a route that retries publishing for a stage whose task finished but whose
  push or pull request failed.

## Required constraints

- No slow external call (`codex`, `git push`, `gh`) runs while the state lock
  is held; state is re-read before it is written.
- `GH_TOKEN` stays out of every `codex` call. Pushing and `gh` go through
  `github_run` and nothing else, and a test shows the new calls respect that.
- Agent output is untrusted. A Refine diff that changes anything but
  `docs/workplan/<brief>` is not published and the reason is recorded. A diff
  that does not apply cleanly is a recorded failure, never a partial push.
- The pull request's title, body and branch name name exactly one task ID, so
  the task-status gate (`scripts/stamp_task_completion.py --check-only`) passes.
- Publishing the same finished task twice yields one commit and one pull
  request.
- HEAD has no Codex status reader. If one has landed by dispatch time, use it;
  do not add a second.

## Non-goals

The agent's own summary text in the pull request (not readable through the
Codex CLI), best-of-N attempts (`--attempt`), creating the stage branch
(P0-009), starting reviewers, and returning reports.

## Runtime reachability

The panel process started by `python3 -m control_panel` or
`workplan-control-panel/compose.yml` checks dispatched tasks while it runs; the
retry route is part of its HTTP API and the UI.

## Write set

- `workplan-control-panel/control_panel/task_scheduler.py`
- `workplan-control-panel/control_panel/api.py`
- `workplan-control-panel/control_panel/integrations/codex.py`
- `workplan-control-panel/control_panel/integrations/repository.py`
- `workplan-control-panel/control_panel/integrations/process.py` (only if a diff cannot pass through `run()` unmodified)
- `workplan-control-panel/control_panel/state.py`
- `workplan-control-panel/control_panel/ui.py`
- `workplan-control-panel/tests/`
- `workplan-control-panel/README.md`

## Contract surfaces

The stage attempt's completion and publication fields; the publish-retry route;
the commit-message convention (`<ID>: …`, body `Task: <ID>`); the identity of a
stage pull request (head `<stage>/<ID>-<N>`, base `task/<ID>`). P0-011 and P0-012
consume these.

## Demonstration

Against a local bare remote, a fake `codex` on `PATH` that reports `pending`
then `ready` and prints a fixed diff, and a stubbed `gh`:

    $ POST /api/tasks/<ID>/refine, then wait for the fake to report ready
    → expect: refine/<ID>-<N> on origin has one new commit on top of task/<ID> with
      subject "<ID>: Refine <title>"; `gh pr create --base task/<ID> --head
      refine/<ID>-<N> --fill` was called once; the stage cell shows published
    $ the fake reports error
    → expect: nothing pushed; the stage cell shows failed with a reason
    $ the fake's diff also edits tasks.json (Refine)
    → expect: not published; the reason names the path
    $ POST the retry route twice for a published stage
    → expect: still one commit and one pull request

Optionally, with a real Codex environment, confirm `codex cloud list --json` and
`codex cloud diff` output on the installed CLI match what the panel reads.

## Verification

- `cd workplan-control-panel && python3 -m unittest discover -s tests`.
- Negative cases: `error` status, empty diff, diff that does not apply, Refine
  diff outside its brief, push rejected, `gh` failing after the push (retry
  then opens the pull request without a second commit), and `GH_TOKEN` absent
  from the `codex` child.
- Say what could not run here, including any real Codex or GitHub use.

## Acceptance criteria

- [ ] A finished Codex task is published to `<stage>/<ID>-<N>` as one panel-authored
      commit and a pull request into `task/<ID>` opened with `--fill`, with no
      operator action.
- [ ] `error`, empty-diff, non-applying and out-of-bounds outcomes are recorded
      with a reason and push nothing.
- [ ] Publishing or retrying twice produces one commit and one pull request.
- [ ] The pull request passes the task-status check-only gate.
- [ ] No `codex` call receives `GH_TOKEN`, and no external call runs under the
      state lock.
- [ ] The stage cell shows running, published, failed and no-changes states.

## Deferrals

| Deferred | Owning task |
|---|---|
| Starting the review-refinement or audit session on the published pull request | P0-011 |
| Reading the report and passing it back to the implementing agent | P0-012 |

## Traces to

`workplan-control-panel/codex-interfaces.md`;
`workplan-control-panel/codex-cloud-cli.md`;
`workplan-control-panel/README.md`; ADR-0004; `docs/testing.md` (Control panel
gate); `scripts/stamp_task_completion.py`; P0-009.
