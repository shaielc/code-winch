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

P0-010 records `branch`, `published_tip`, `pull_request_number`, and
`pull_request_url` on the exact numbered Refine or Implement attempt before
returning it from `_publish_attempt`. Extend that publication path rather than
rediscovering a pull request from the task base: after publication is recorded,
start the reviewer for that attempt's recorded pull request and published tip.
Multiple numbered attempts may remain open, so neither the automatic path nor
its idempotence may be keyed only by task ID or base branch.

- Add `review-refinement` as a panel stage beside `audit`, including a bundled
  prompt, attempt record, API route, UI control and conversation link. Use the
  route/stage slug `review-refinement`; keep `/api/tasks/<ID>/review` unknown as
  current API tests require. The existing `audit` stage and manual control stay
  supported.
- Use a separate Claude routine for review-refinement, configured by
  `CLAUDE_REVIEW_REFINEMENT_ROUTINE_URL` and
  `CLAUDE_REVIEW_REFINEMENT_ROUTINE_TOKEN`. The existing routine prompt accepts
  only audit requests, so sharing it would weaken the routine-side request
  boundary. Thread both stage configurations from `api.py` into the scheduler,
  through `compose.yml` and `.env.example`, and document their setup.
- Generalise the Claude integration's configuration errors so they identify the
  selected routine instead of always naming the audit variables. Preserve its
  distinction between an HTTP refusal, which proves no session started, and an
  accepted or transport failure, whose outcome is uncertain.
- Factor reviewer start so the automatic publication path can supply the exact
  stage attempt, pull-request URL/number and `published_tip` without calling
  `repository.task_pull_requests`. A Refine publication starts
  `review-refinement`; an Implement publication starts `audit`. Record reviewer
  attempts against both the pull request and head so a repeated publication
  check reuses the one session for that pair.
- Keep manual starts explicit. Manual review-refinement targets an open
  `refine/<ID>-<N>` stage pull request and manual audit targets an open
  `implement/<ID>-<N>` stage pull request. Filter by the recorded head-ref prefix
  before applying an optional pull-request number; if more than one matching
  candidate remains, return the candidates for UI selection. An open pull
  request for the other stage must never create ambiguity.
- Persist the `submitting` reviewer attempt before firing its routine. On a
  `RoutineError`, remove that attempt and surface the refusal without an
  automatic retry. On a timeout, malformed accepted response or other uncertain
  failure, retain the attempt so the operator must inspect Claude and explicitly
  expire it before trying again. Do not hold the state lock during the routine
  HTTP request; recheck the attempt identity under the lock before recording its
  returned session URL.
- Both prompts require the report's first line to be
  `Report: <review-refinement|audit> <ID> <head-sha>`. The audit prompt also
  requires the next line to state `Not Complete`, `Input Required` or `Complete`
  as defined by `skills/audit-task/SKILL.md`. Both prompts require a comment-only
  pull-request review. P0-012 consumes this header.

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
- Starting is idempotent per reviewer stage, pull request and head. The
  persist-before-act rule keeps an uncertain outcome visible rather than
  repeating it, including when publication polling overlaps or the panel
  restarts.

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
- `workplan-control-panel/control_panel/prompts/review-refinement.md`
- `workplan-control-panel/control_panel/prompts/audit.md`
- `workplan-control-panel/control_panel/ui.py`
- `workplan-control-panel/tests/test_claude.py`
- `workplan-control-panel/tests/test_prompts.py`
- `workplan-control-panel/tests/test_publication.py`
- `workplan-control-panel/tests/test_task_scheduler.py`
- `workplan-control-panel/README.md`, `claude-cloud-cli.md`, `compose.yml`, `.env.example`

## Contract surfaces

The `Report:` header format; the review-refinement stage's name, route and
prompt; the routine URL and token settings; the rule for which pull request an
audit targets. P0-012 consumes the header and the sessions started here.

## Demonstration

Run the focused local panel scenarios, whose routine HTTP server and local Git
repositories require no live Claude or GitHub account:

    $ cd workplan-control-panel
    $ python3 -m unittest tests.test_publication tests.test_task_scheduler tests.test_claude
    → expect: publishing a refine/<ID>-<N> pull request fires one
      review-refinement request for its recorded URL and published tip, and the
      stage cell links the returned session
    → expect: publishing an implement/<ID>-<N> pull request while a Refine
      pull request is open fires one audit request for the Implement pull request
    → expect: repeated or overlapping publication checks at the same head
      fire no duplicate; a distinct recorded head can start one new session
    → expect: a 429 leaves no reviewer attempt or session, exposes the
      refusal, and causes no automatic retry; an uncertain response leaves one
      visible attempt and also causes no automatic retry

For the maintained hand check in an isolated configured panel deployment,
publish one real Refine or Implement attempt as P0-010 describes. Observe the
matching reviewer stage appear without pressing its button, follow its Claude
session link, and verify that its eventual pull-request review starts with the
required `Report:` line and is comment-only. Record if live routine credentials
or GitHub access make this check unavailable.

## Verification

- `cd workplan-control-panel && python3 -m unittest discover -s tests`.
- Focused cases: each publication stage selects its matching reviewer and exact
  recorded pull request/head; other-stage and multiple numbered pull requests
  cannot misdirect it; repeated and overlapping publication checks start one
  session; a changed head starts one additional session.
- Negative cases: an HTTP refusal removes the provisional attempt and starts no
  retry; an uncertain outcome remains visible and blocks another fire; a
  missing stage-specific routine setting gives a clear error rather than a
  crash. Assert that routine I/O occurs outside the state lock and stale or
  expired attempts are not overwritten when the call returns.
- Say what could not run here, including any real routine use.

## Acceptance criteria

- [ ] Publishing a stage pull request starts exactly one matching session for
      its recorded pull request and published head without operator action;
      concurrent or repeated publication checks do not duplicate it.
- [ ] A different recorded head may start one new session; an unchanged head
      does not.
- [ ] The audit targets the Implement pull request even when a Refine pull
      request is open on the same base, and review-refinement applies the
      inverse rule.
- [ ] Both prompts require the `Report:` header, and the review-refinement and
      audit prompts remain comment-only.
- [ ] A refused request records no attempt or session and is surfaced without
      automatic retry; an uncertain outcome remains visible and also is not
      fired again automatically.
- [ ] Both stages' routine settings are documented and wired through
      `compose.yml` and `.env.example`, with stage-specific configuration errors.
- [ ] No routine HTTP request runs under the state lock, and a returning request
      cannot overwrite an attempt expired or replaced while it was in flight.

## Deferrals

| Deferred | Owning task |
|---|---|
| Recognising the finished report on the pull request and acting on it | P0-012 |

## Traces to

`workplan-control-panel/claude-cloud-cli.md`;
`workplan-control-panel/README.md`; `skills/review-refinement/SKILL.md`;
`skills/audit-task/SKILL.md`; ADR-0004; `docs/testing.md` (Control panel gate);
P0-010.
