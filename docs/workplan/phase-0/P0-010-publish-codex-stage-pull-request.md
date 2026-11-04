# P0-010: Publish a finished Codex stage as a pull request

**Phase:** 0 — Plan the implementation phases
**Shape:** capability
**Dependencies:** P0-009 (contract: publish onto the recorded `<stage>/<ID>-<N>` branch and extend P0-009's stage-attempt record)

Second of four control-panel automation tasks (P0-009 to P0-012). The panel is
repository tooling, so Phase 0 owns this capability. P0-009 is complete at HEAD:
Refine and Implement attempts record `branch`, `base`, `head`, status and the
Codex task URL, and the numbered branch exists on origin before dispatch.

## Objective

When a Refine or Implement Codex task finishes, the running panel publishes its
changes as one commit on the recorded stage branch and opens one pull request
from that branch into `task/<ID>`, without operator action.

## Scope

- Add a bounded background reconciliation loop for live Refine and Implement
  attempts. Start and stop it with the HTTP server, check promptly after startup,
  and isolate per-attempt failures. Parse the Codex task ID from `task_url` and
  follow every `codex cloud list --json` cursor before treating it as absent.
- Map Codex `pending` to running, `ready` and `applied` to publishable, and
  `error` to terminal failure. Older attempt records without publication fields
  must continue to load and render.
- Fetch `codex cloud diff <task-id>` for a publishable attempt. Preserve bytes
  significant to `git apply`; extend the shared process wrapper rather than
  bypassing it if its stripped-text return cannot do that.
- In a temporary worktree at the recorded remote stage-branch tip, apply the
  entire diff and always remove the worktree. A failed apply records failure and
  leaves the remote unchanged. For Refine, inspect the applied result and refuse
  it unless the only changed path is the tracker-selected exact brief,
  `docs/workplan/<task.brief>`; never trust paths claimed by agent output.
- Record a successfully applied diff with no resulting changes as terminal
  `no-changes`. Otherwise create one commit with subject
  `<ID>: <Refine|Implement> <task title>` and body `Task: <ID>`, then push it to
  the recorded branch through the GitHub-credentialled command path.
- Reuse an open pull request whose head is the recorded branch and whose base is
  `task/<ID>`. Otherwise run `gh pr create --base task/<ID> --head
  <stage>/<ID>-<N> --fill`, validate its result, and record its number and URL.
  The commit message intentionally supplies the `--fill` title and body.
- Make commit, push and pull-request creation restart-safe and idempotent. After
  a lost push reply or a `gh` failure, recognize the remote published tip and
  resume without applying or committing again. Never force-push a branch that
  moved unexpectedly.
- Extend the attempt with Codex task ID, last observed status, publication
  outcome, published tip, pull-request number and URL, and a safe failure reason.
  Re-read under the state lock before every write and update only the same live
  attempt; no Codex, Git, or GitHub subprocess runs while that lock is held.
- Add `POST /api/tasks/<ID>/<refine|implement>/publish` to retry a finished
  attempt whose publication failed or was interrupted. Refuse unknown tasks,
  Audit, and stages without a publishable attempt. Use the same idempotent path
  as background reconciliation.
- Show running, published, failed, and no-changes in both stage-cell renderings.
  Published links to the pull request; failed exposes its reason and retry.
  Document the automatic check, states, and route in the panel README.

## Required constraints

- `GH_TOKEN` is absent from every `codex` child. Push and `gh` use `github_run`;
  local worktree, apply, and commit commands do not.
- Refine changes outside the exact brief and diffs that do not apply completely
  are recorded failures and push nothing.
- Empty diff, Codex error, and validation failure do not dispatch another task.
- Publication uses P0-009's recorded branch; it neither allocates nor rewrites
  a stage branch.
- State locks protect only reads and writes. A stale worker cannot overwrite an
  expired or superseded attempt.

## Non-goals

- Agent summaries or session history, which the Codex CLI does not expose.
- Best-of-N attempts or `codex cloud diff --attempt` selection.
- Stage-branch creation, new Codex dispatch, review/audit startup, or reports.
- Audit-stage publication or the final task pull request into `main`.

## Runtime reachability

`python3 -m control_panel` and the `control-panel` Compose service start the
completion reconciler with the HTTP server. The scheduler path is also reachable
through the retry route, and the table and dependency-tree views show its result.

## Write set

- `workplan-control-panel/control_panel/api.py`
- `workplan-control-panel/control_panel/task_scheduler.py`
- `workplan-control-panel/control_panel/state.py`
- `workplan-control-panel/control_panel/ui.py`
- `workplan-control-panel/control_panel/integrations/codex.py`
- `workplan-control-panel/control_panel/integrations/process.py`
- `workplan-control-panel/control_panel/integrations/repository.py`
- `workplan-control-panel/tests/test_task_scheduler.py`
- `workplan-control-panel/tests/test_ui.py`
- `workplan-control-panel/README.md`

## Contract surfaces

- Persisted stage-attempt completion and publication fields.
- API: `POST /api/tasks/<ID>/<refine|implement>/publish`.
- Stage pull request: head `<stage>/<ID>-<N>`, base `task/<ID>`.
- Commit: subject `<ID>: <Refine|Implement> <task title>`, body `Task: <ID>`.

P0-011 and P0-012 consume these surfaces.

## Demonstration

Run the panel against a local bare origin with a fake `codex` on `PATH` whose
status changes from `pending` to `ready` and whose diff is fixed, plus a stub
`gh` that records arguments:

    $ curl -sS -X POST -H 'Content-Type: application/json' \
        -H "Authorization: Bearer $PANEL_TOKEN" -d '{}' \
        http://127.0.0.1:8765/api/tasks/<ID>/refine
    $ git --git-dir=/path/to/origin.git log -1 --format='%s%n%b' refine/<ID>-1
    → expect after ready: one commit above task/<ID>, subject
      "<ID>: Refine <title>", body "Task: <ID>"; `gh` records one
      `pr create --base task/<ID> --head refine/<ID>-1 --fill`; both views show
      published and link to the pull request

    $ curl -sS -X POST -H 'Content-Type: application/json' \
        -H "Authorization: Bearer $PANEL_TOKEN" -d '{}' \
        http://127.0.0.1:8765/api/tasks/<ID>/refine/publish
    $ curl -sS -X POST -H 'Content-Type: application/json' \
        -H "Authorization: Bearer $PANEL_TOKEN" -d '{}' \
        http://127.0.0.1:8765/api/tasks/<ID>/refine/publish
    → expect: still one publication commit and one pull-request creation

Repeat with Codex `error`, then a Refine diff that also edits
`docs/workplan/tasks.json`:

    → expect: the branch tip does not move; both views show failed and a safe
      reason identifies the Codex error or out-of-bounds path

Real Codex/GitHub use is optional. If available, compare the installed CLI's
paged list and diff output with what the panel accepts; do not publish an
artificial task solely for this check.

## Verification

- `cd workplan-control-panel && python3 -m unittest discover -s tests`
- Focused integration coverage uses a local bare remote, fake `codex`, and
  stubbed `gh` for pending-to-ready publication and the retry route.
- Negative coverage: paged/missing/malformed status, Codex error, empty and
  non-applying diffs, out-of-bounds Refine diff, moved branch, rejected push,
  `gh` failure after push, and expiry/supersession during an external call.
- Process-boundary coverage proves diff content is preserved, `GH_TOKEN` is
  absent from Codex and present only for `github_run`, and external commands do
  not hold the state lock.
- Report unavailable checks, including real Codex or GitHub use.

## Acceptance criteria

- [ ] Without operator action, pending-to-ready/applied produces one commit on
      the recorded branch and one pull request into `task/<ID>`; both views show
      and link the result.
- [ ] Codex error, empty result, non-applying diff, out-of-bounds Refine diff,
      and moved branch produce their visible terminal state and push no agent
      changes.
- [ ] Reconciliation, restart, and retries cannot create a second commit or pull
      request, including after a successful push followed by failure.
- [ ] The `--fill` title and body name exactly one task, and the body passes
      `scripts/stamp_task_completion.py --check-only` metadata validation.
- [ ] Codex never receives `GH_TOKEN`; only GitHub operations do; external calls
      neither hold the state lock nor overwrite a superseded attempt.
- [ ] The retry API and both views distinguish running, published, failed, and
      no-changes with safe, actionable output.

## Deferrals

| Deferred | Owning task |
|---|---|
| Starting review-refinement or audit on the published pull request | P0-011 |
| Reading the report and passing it to the implementing agent | P0-012 |

## Traces to

`workplan-control-panel/codex-interfaces.md` (Codex Cloud and GitHub);
`workplan-control-panel/codex-cloud-cli.md` (status, paging, and diff);
`workplan-control-panel/README.md` (scheduler, API, credentials, runtime);
ADR-0004; `docs/testing.md` (Control panel gate);
`scripts/stamp_task_completion.py`; P0-009.
