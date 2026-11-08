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
(`workplan-control-panel/codex-interfaces.md`). Implement against the existing
stage-attempt record and repository helpers:

- Notice completion without an operator action through the Codex CLI status
  listing (`pending`, `ready`, `applied`, `error`). Locate the recorded task
  across pages before deciding that it is absent; `ready` or `applied`
  publishes, `error` records failure, and an empty finished diff records
  no changes.
- Fetch the task's diff with `codex cloud diff` without changing bytes needed
  by `git apply`. `integrations/process.py` currently strips captured output;
  ensure this diff path does not. Apply in a temporary worktree at the recorded
  stage branch's tip; the panel's own clone is never the apply target.
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

- Publication performs no slow external call (`codex`, `git push`, `gh`) while
  holding the state lock, and rechecks the live attempt before writing state.
- `GH_TOKEN` stays out of every `codex` call. Pushing and `gh` go through
  `github_run` and nothing else, and a test shows the new calls respect that.
- Agent output is untrusted. A Refine diff that changes anything but
  `docs/workplan/<brief>` is not published and the reason is recorded. A diff
  that does not apply cleanly is a recorded failure, never a partial push.
- The pull request's title, body and branch name name exactly one task ID, so
  the task-status gate (`scripts/stamp_task_completion.py --check-only`) passes.
- Publishing the same finished task twice yields one commit and one pull
  request.
- Reuse existing Codex status or repository helpers if present at implementation
  time rather than adding competing readers or PR operations.

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

Use the maintained panel UI in an **isolated, configured** control-panel
deployment (startup and sign-in: `workplan-control-panel/README.md`). Use a
real, eligible workplan task already being processed there; do not create a
synthetic task or build a fake Codex/GitHub environment for this manual check.
The task must have a Refine or Implement stage that actually produces a diff.
This check requires working Codex Cloud and GitHub credentials; report the
specific unavailable prerequisite if it cannot run.

1. Open the panel, sign in, and use **Prepare** if the selected task is not
   prepared. Select **Refine** (or **Implement**) in its stage cell and observe
   the recorded conversation and numbered stage branch.
2. After the Codex task finishes, reload the panel. Observe **published** and
   open the stage's pull-request link. On GitHub, verify that its head is the
   recorded stage branch, its base is `task/<ID>`, and its publication commit
   has subject `<ID>: <Refine|Implement> <title>` and body `Task: <ID>`.
   Confirm that the branch contains the task's expected changed files.
3. From a terminal with access to that isolated panel, set `PANEL_URL` to
   its base URL and `TASK_ID` to the selected task ID. With `PANEL_TOKEN` set
   to that panel's token, run the same retry twice:

       curl -fsS -X POST -H "Authorization: Bearer $PANEL_TOKEN" \\
         -H 'Content-Type: application/json' -d '{}' \\
         "$PANEL_URL/api/tasks/$TASK_ID/refine/publish"

   (Use `implement/publish` instead when demonstrating Implement.) Reload
   the stage and PR. The same publication commit and PR must remain, with no
   additional publication commit.

Keep the manual demonstration distinct from fake-backed negative coverage in
Verification. Record what was actually observed, not predicted output.

## Verification

- `cd workplan-control-panel && python3 -m unittest discover -s tests`.
- Focused checks: a recorded task beyond the first status page is found;
  diff content needed by `git apply` is preserved (the shared runner currently
  strips it); `error`, empty diff, non-applying diff, out-of-brief Refine diff,
  and push rejection do not publish agent changes.
- A `gh` failure after a successful push is retryable without another commit
  or pull request. Verify duplicate publication attempts and `GH_TOKEN` absent
  from Codex, with no new publication subprocess under the state lock.
- Say what could not run here, including any real Codex or GitHub use.

## Acceptance criteria

- [ ] A finished Codex task is published to `<stage>/<ID>-<N>` as one panel-authored
      commit and a pull request into `task/<ID>` opened with `--fill`, with no
      operator action.
- [ ] `error`, empty-diff, non-applying and out-of-bounds outcomes are recorded
      with a reason and push nothing.
- [ ] Publishing or retrying twice, including after a push succeeded but PR
      creation failed, produces one commit and one pull request.
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
