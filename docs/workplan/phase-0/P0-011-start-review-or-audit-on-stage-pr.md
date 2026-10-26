# P0-011: Start the review or audit session when a stage pull request opens

**Phase:** 0 — Plan the implementation phases
**Shape:** capability
**Dependencies:** P0-010 (contract: starts from the stage pull request and published head that P0-010 records on the stage attempt)

Third of four control-panel automation tasks (P0-009 to P0-012). This is a
Phase 0 meta task because the control panel is repository tooling, not the
product; see P0-009. The shape is the nearest fit: it adds observable behavior to
the panel, reached when a stage pull request is published.

## Objective

A published Refine pull request starts a review-refinement Claude session, and a
published Implement pull request starts an audit session, each at the pull
request's current head, with no button press.

## Scope

Audit is today a button that fires a Claude routine with a prompt naming one
pull request. Review-refinement has a skill (`skills/review-refinement/SKILL.md`)
and no panel stage. Proposed approach, for the refiner to confirm against code
at dispatch:

- When P0-010 publishes a stage pull request, or pushes a new head to one,
  start the matching session once for that head: `review-refinement` for
  `refine/<ID>`, `audit` for `implement/<ID>`. A head that already has a live
  session is not started again.
- Add the review-refinement stage with its prompt, record, route and UI
  control, beside Audit. Pick a stage name that does not collide with the
  existing tests that expect `/api/tasks/<ID>/review` to return 404.
- Decide how the review-refinement request reaches Claude. The existing
  routine's own prompt accepts only audit requests, so either configure a second
  routine (new URL and token settings in `compose.yml`, `.env.example`, `api.py`
  and the README) or widen the existing routine's prompt, and say which.
- Make the pull request selection explicit. The automatic start passes the pull
  request it just published. The manual Audit control prefers the
  `implement/<ID>` pull request, because an open Refine pull request on the same
  base would otherwise defeat its "exactly one candidate" rule.
- Both prompts require the report's first line to be
  `Report: <review-refinement|audit> <ID> <head-sha>`; an audit report also states
  its audit state on the next line (`Not Complete`, `Input Required` or
  `Complete`, as in `skills/audit-task/SKILL.md`). P0-012 relies on this header.

## Required constraints

- A reviewer or auditor remains comment-only: it posts a review comment and
  does not approve, request changes, commit or push, as `prompts/audit.md`
  already requires.
- A routine that refuses the request (bad token, rate limit) records no session
  and is surfaced to the operator; the panel does not retry in a loop. The
  routine allows 30 fires per hour.
- Text sent to the routine is data, not instruction, as the routine's own
  prompt already treats it; the new prompt names only the task, pull request
  and head.
- Starting is idempotent per pull request head, as Audit is today, and the
  persist-before-act rule keeps an uncertain outcome visible rather than
  repeating it.

## Non-goals

Reading or acting on the report (P0-012); polling a Claude session's status,
which uses an undocumented endpoint and is unnecessary because completion is the
report appearing on the pull request; reviews run by Codex; changing what
`review-refinement` or `audit-task` ask the reviewer to judge.

## Runtime reachability

The publish path from P0-010 in the running panel, the new stage's route and UI
control, and the Claude routine(s) configured on claude.ai/code/routines for the
URLs and tokens the panel is given.

## Write set

- `workplan-control-panel/control_panel/task_scheduler.py`
- `workplan-control-panel/control_panel/api.py`
- `workplan-control-panel/control_panel/integrations/claude.py`
- `workplan-control-panel/control_panel/integrations/repository.py`
- `workplan-control-panel/control_panel/prompts/` (a new `review-refinement` prompt; `audit.md`)
- `workplan-control-panel/control_panel/ui.py`
- `workplan-control-panel/tests/`
- `workplan-control-panel/README.md`, `claude-cloud-cli.md`, `compose.yml`, `.env.example`

## Contract surfaces

The `Report:` header format; the review-refinement stage's name, route and
prompt; the routine URL and token settings; the rule for which pull request an
audit targets. P0-012 consumes the header and the sessions started here.

## Demonstration

Against a local HTTP stub for the routine (as `tests/test_claude.py` does) and a
published stage pull request recorded by the P0-010 path:

    $ publish a refine/<ID> pull request
    → expect: one routine request naming <ID>, the pull request and its head;
      the stage cell links the session
    $ publish an implement/<ID> pull request while refine/<ID> is open
    → expect: the audit request names the implement pull request, not the refine one
    $ push a new head to the same pull request
    → expect: one new session for the new head, none for the old
    $ the routine answers 429
    → expect: no session recorded; the operator sees the refusal; no retry loop

Optionally, with real routines: the resulting review begins with the `Report:`
header and lands as a comment-only review on the pull request.

## Verification

- `cd workplan-control-panel && python3 -m unittest discover -s tests`.
- Negative cases: refusal and uncertain outcomes record nothing started; a
  second publish at the same head starts nothing; a missing routine setting
  gives a clear error rather than a crash.
- Say what could not run here, including any real routine use.

## Acceptance criteria

- [ ] Publishing a stage pull request starts exactly one matching session for
      that head without operator action.
- [ ] A new head starts a new session; an unchanged head does not.
- [ ] The audit targets the Implement pull request even when a Refine pull
      request is open on the same base.
- [ ] Both prompts require the `Report:` header, and the review-refinement and
      audit prompts remain comment-only.
- [ ] A refused request is surfaced and not retried automatically.
- [ ] The routine settings for the new stage are documented and wired through
      `compose.yml` and `.env.example`.

## Deferrals

| Deferred | Owning task |
|---|---|
| Recognising the finished report on the pull request and acting on it | P0-012 |

## Traces to

`workplan-control-panel/claude-cloud-cli.md`;
`workplan-control-panel/README.md`; `skills/review-refinement/SKILL.md`;
`skills/audit-task/SKILL.md`; ADR-0004; `docs/testing.md` (Control panel gate);
P0-010.
