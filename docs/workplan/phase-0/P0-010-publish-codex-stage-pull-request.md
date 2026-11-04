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

Use the Refine or Implement attempt's recorded `branch` from P0-009 for every
publication attempt; do not reconstruct a branch name or allocate another
number. Extract and persist the Codex task id from the recorded `task_url`, and
fail that attempt with a reason if the URL does not contain one.

`codex cloud exec` returns only a task URL. The CLI can report a task's status
and print its unified diff, but it cannot read the agent's messages or session
history, so derive the commit and pull-request text from the tracker task and
stage record (`workplan-control-panel/codex-interfaces.md`). Implement the flow
against the code at HEAD in this order:

- Extend `integrations/codex.py` with a paginated `codex cloud list --json`
  reader that locates the recorded task id, and a `codex cloud diff <task-id>`
  reader. Treat `pending` as still running, `ready` and `applied` as ready to
  publish, and `error` as terminal failure. Do not treat a task omitted from one
  20-item page as terminal; follow the returned cursor. Preserve the diff's
  bytes exactly, including its trailing newline: the shared `run()` currently
  calls `strip()`, so give this call an explicit non-stripping result path in
  `integrations/process.py` rather than changing every existing caller.
- Add a scheduler-owned background reconciliation loop started with the HTTP
  server. It periodically snapshots submitted Refine and Implement attempts,
  performs Codex and repository calls outside `locked()`, then reacquires the
  lock, reloads state, and writes a result only if the same attempt is still
  current. One failed or malformed task lookup must not stop later checks. The
  existing submission route remains prompt: publication is not added to the
  request that dispatches Codex.
- For a ready attempt, ask `integrations/repository.py` to fetch the recorded
  stage branch and task branch and create a detached temporary worktree at the
  fetched stage tip. Apply the unified diff there as one atomic operation. A
  Refine diff is publishable only when every changed path is exactly the brief
  named by the tracker task; reject renames, deletions, binary patches, and any
  second path. An Implement diff is not subject to that Refine-only path rule.
  An empty diff becomes `no-changes`. An apply failure becomes `failed`; neither
  outcome commits or pushes anything. Always remove the temporary worktree.
- Commit a publishable worktree as the configured panel author with subject
  `<ID>: <Refine|Implement> <task title>` and body `Task: <ID>`. Push the new tip
  to the recorded stage branch with a force-with-lease against the fetched tip,
  so a concurrently moved branch is a recorded failure rather than an
  overwrite. Then use `github_run` to find an existing open pull request for
  that exact head and base; reuse it, or run `gh pr create --base task/<ID>
  --head <stage>/<ID>-<N> --fill`. Do not use the task-to-main pull-request
  helpers, whose base and state record have different meanings.
- Extend the attempt record in the existing state schema with the Codex task
  id, publication status (`running`, `published`, `failed`, or `no-changes`),
  update time, failure reason where applicable, published commit, and pull
  request number and URL where applicable. Render those states and reasons in
  both the stage cell and dependency-tree view while preserving the conversation
  link and compatibility with attempts that predate these fields.
- Add `POST /api/tasks/<ID>/<refine|implement>/publish` and the corresponding
  stage-menu action. It retries status/diff/application, push, or pull-request
  creation for the recorded attempt without creating a branch, commit, or pull
  request that already exists. Refuse Audit and attempts that have not been
  submitted. The background loop and this route call the same idempotent
  scheduler operation.

## Required constraints

- No slow external call (`codex`, `git push`, `gh`) runs while the state lock
  is held; state is re-read and the attempt identity revalidated before it is
  written.
- `GH_TOKEN` stays out of every `codex` call. Pushing and `gh` go through
  `github_run` and nothing else, and a test shows the new calls respect that.
- Agent output is untrusted. A Refine diff that changes anything but the exact
  tracker brief path under `docs/workplan/` is not published and the reason is
  recorded. A diff that does not apply cleanly is a recorded failure, never a
  partial push.
- The pull request's title, body and branch name name exactly one task ID, so
  the task-status gate (`scripts/stamp_task_completion.py --check-only`) passes.
- Publishing the same finished task twice yields one commit and one pull
  request.
- Keep list pagination bounded against a repeated cursor and reject malformed
  JSON or an unknown status as a failure of that attempt, not of the polling
  thread.

## Non-goals

The agent's own summary text in the pull request (not readable through the
Codex CLI), best-of-N attempts (`--attempt`), creating or renumbering the stage
branch, changing the task-to-main pull-request flow, starting reviewers, and
returning reports.

## Runtime reachability

The panel process started by `python3 -m control_panel` or
`workplan-control-panel/compose.yml` starts the reconciliation loop and checks
submitted Refine and Implement attempts while it runs. The publish route is
part of its authenticated HTTP API and its action is reachable from each
stage's UI menu.

## Write set

- `workplan-control-panel/control_panel/task_scheduler.py`
- `workplan-control-panel/control_panel/api.py`
- `workplan-control-panel/control_panel/integrations/codex.py`
- `workplan-control-panel/control_panel/integrations/repository.py`
- `workplan-control-panel/control_panel/integrations/process.py`
- `workplan-control-panel/control_panel/ui.py`
- `workplan-control-panel/tests/`
- `workplan-control-panel/README.md`

## Contract surfaces

The stage attempt's Codex-task, completion, and publication fields; `POST
/api/tasks/<ID>/<refine|implement>/publish`;
the commit-message convention (`<ID>: …`, body `Task: <ID>`); the identity of a
stage pull request (head `<stage>/<ID>-<N>`, base `task/<ID>`). P0-011 and P0-012
consume these.

## Demonstration

In the focused HTTP scenario, use the existing local bare-remote fixture, a
fake `codex` on `PATH` whose paginated list reports `pending` then `ready` and
whose diff command prints a fixed patch, and the existing stubbed GitHub
boundary:

    $ POST /api/tasks/<ID>/refine, then wait for the panel to report published
    → expect: refine/<ID>-<N> on origin has one new commit on top of task/<ID> with
      subject "<ID>: Refine <title>"; `gh pr create --base task/<ID> --head
      refine/<ID>-<N> --fill` was called once; the stage cell shows published
    $ the fake reports error
    → expect: nothing pushed; the stage cell shows failed with a reason
    $ the fake's diff also edits tasks.json (Refine)
    → expect: not published; the reason names the path
    $ POST /api/tasks/<ID>/refine/publish twice for the published stage
    → expect: still one commit and one pull request

Optionally, with a real Codex environment, confirm `codex cloud list --json` and
`codex cloud diff` output on the installed CLI match what the panel reads.

## Verification

- `cd workplan-control-panel && python3 -m unittest discover -s tests`.
- Focused tests cover multi-page task lookup and repeated cursors; malformed
  list output and unknown status; `error` status; empty diff; exact trailing
  newline preservation; a diff that does not apply; Refine deletion, rename,
  binary, second-path, and outside-brief patches; a leased push rejection; and
  one failed lookup not stopping reconciliation of another task.
- The HTTP scenario covers automatic `pending` to `ready` publication and the
  manual publish route. It also covers `gh` failing after the push, followed by
  a retry that opens the pull request without a second commit or PR, and runs
  `scripts/stamp_task_completion.py --check-only` against the created PR event
  fixture.
- The credential-boundary test proves every new push and `gh` call uses
  `github_run` and that `GH_TOKEN` is absent from both new `codex` child
  commands. A lock assertion proves no Codex, push, or `gh` call runs while the
  state lock is held.
- Say what could not run here, including any real Codex or GitHub use.

## Acceptance criteria

- [ ] A `ready` or `applied` Codex task is automatically published to the
      attempt's recorded `<stage>/<ID>-<N>` as one panel-authored
      commit and a pull request into `task/<ID>` opened with `--fill`, with no
      operator action.
- [ ] `error`, malformed/unknown status, empty-diff, non-applying and
      Refine out-of-bounds outcomes are recorded
      with a reason and push nothing.
- [ ] Publishing or retrying twice, including retry after the push succeeded
      but pull-request creation failed, produces one commit and one pull request;
      a leased push refuses a concurrently moved branch.
- [ ] The pull request passes the task-status check-only gate.
- [ ] No `codex` call receives `GH_TOKEN`, and no external call runs under the
      state lock.
- [ ] Paginated lookup finds the recorded task beyond the first page and cannot
      loop forever on a repeated cursor; one bad task does not stop other tasks
      from being reconciled.
- [ ] Existing stage attempts still load and render, and the stage cell and tree
      view show running, published, failed and no-changes states with reasons and
      the existing conversation link.

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
