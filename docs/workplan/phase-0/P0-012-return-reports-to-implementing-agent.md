# P0-012: Return review and audit reports to the implementing agent

**Phase:** 0 — Plan the implementation phases
**Shape:** capability
**Dependencies:** P0-011 (contract: recognises the `Report:` header and the review and audit sessions that P0-011 defines)

Last of four control-panel automation tasks (P0-009 to P0-012). This is a Phase 0
meta task because the control panel is repository tooling, not the product; see
P0-009. The shape is the nearest fit: it adds observable behavior to the panel,
reached by pull-request events and the panel's API.

## Objective

When a review-refinement or audit report appears on a stage pull request, the
panel notices it and, if the report calls for changes, runs a new Codex task on
the same stage branch with that report, so the fixes arrive as new commits on
the same pull request.

## Scope

A report return continues the reviewed attempt on its recorded branch and pull
request. It does not allocate a new numbered branch; fresh stage submissions
and retries after expiry are the P0-009 path.

A finished Codex Cloud task cannot receive a follow-up turn (the CLI has no
command for it; upstream issue openai/codex#24777 is open), so "pass it back"
means a new task on `<stage>/<ID>-<N>`, a new conversation with the report in its
prompt. Proposed approach, for the refiner to confirm against code at dispatch:

- **Listen to pull-request changes.** Extend the Actions route that already
  tells the panel about merges (`.github/workflows/task-scheduler.yml`) to
  forward review and comment events on pull requests whose base is `task/*` to a
  new authenticated panel route. The refiner may choose polling instead and says
  why. Either way, the panel re-reads the pull request from GitHub and treats an
  event as a hint; a sync also reconciles, so a missed event does not strand a
  stage.
- **Recognise a report.** It opens with the `Report:` header from P0-011, names
  the pull request's current head, and comes from the expected author. Each
  report is handled once, by review id; a report for an older head is ignored.
- **Decide what returns.** An audit in state `Not Complete` returns its Blocking
  findings. `Input Required` and `Complete` stop, because
  `skills/audit-task/SKILL.md` leaves Defect and Deferrable dispositions to a
  person, and the stage shows that a person is needed. A review-refinement
  report is returned whole.
- **Send it.** Start `codex cloud exec --branch <stage>/<ID>-<N>` with a fixed
  instruction plus the report, framed as untrusted data and bounded in size. Its
  result is published by P0-010, which pushes a new head and so restarts the
  reviewer through P0-011.
- **Bound the loop.** Stop after a configurable number of automatic returns per
  stage per task (proposed default 2) and show that a person is needed.
- Show the loop state — waiting for report, returned (round n), needs a person —
  on the stage.

## Required constraints

- A report is accepted only from the identity the reviewer runs as, on the
  current head of a stage pull request the panel opened. Comments from anyone
  else are ignored, never forwarded to Codex.
- The report is attacker-influenced content: the prompt that carries it says it
  is data to address, not instructions to follow, and the agent's changes are
  published under P0-010's limits (a Refine return still may change only its
  brief).
- The new route authenticates like `/api/events/merge`. A duplicate or replayed
  event does nothing.
- The round limit holds across restarts: it is recorded in state, not in memory.
- No slow external call runs under the state lock.

## Non-goals

Acting on human comments; merging or approving; recording finding verdicts
(`audit-finalize`); returning Defect, Deferrable or Harden findings
without a person; continuing the original Codex conversation; `@codex` comments
through the GitHub integration, which are unverified for the panel's GitHub App.

## Runtime reachability

The new event route and the sync pass of the running panel, the Actions workflow
that forwards events to it, and the stage UI.

## Write set

- `.github/workflows/task-scheduler.yml` (or a new workflow beside it)
- `workplan-control-panel/control_panel/task_scheduler.py`
- `workplan-control-panel/control_panel/api.py`
- `workplan-control-panel/control_panel/integrations/repository.py`
- `workplan-control-panel/control_panel/prompts/` (a return prompt)
- `workplan-control-panel/control_panel/state.py`
- `workplan-control-panel/control_panel/ui.py`
- `workplan-control-panel/tests/`
- `workplan-control-panel/README.md`, `compose.yml`, `.env.example` (for the round limit)

## Contract surfaces

The pull-request event route and the workflow triggers that feed it; the
return-round fields on the stage attempt; the return prompt; the round-limit
setting.

## Demonstration

Against a local bare remote, a stubbed `gh`, and a fake `codex`, with a stage
pull request already published and a session started by the P0-011 path:

    $ deliver an audit report "Not Complete" for the current head
    → expect: one codex task on implement/<ID>-<N> whose prompt carries the Blocking
      findings; the stage shows returned (round 1)
    $ deliver the same event again
    → expect: nothing new
    $ deliver a report for an older head, or from another author
    → expect: ignored
    $ deliver an "Input Required" audit
    → expect: no codex task; the stage shows a person is needed
    $ repeat returns past the limit, then restart the panel
    → expect: the limit still holds; the stage shows a person is needed

Optionally, with real services: a real audit's report returns and the fixes
appear as a new commit on the same pull request.

## Verification

- `cd workplan-control-panel && python3 -m unittest discover -s tests`, and the
  workflow change validated against the existing `task-scheduler.yml` behavior.
- Adversarial cases: report text containing instructions, forged author, stale
  head, duplicate and replayed events, oversize report, and a missed event
  recovered by sync.
- Say what could not run here, including any real GitHub Actions, Codex or
  Claude use.

## Acceptance criteria

- [ ] A current-head `Not Complete` audit report from the expected identity
      starts exactly one Codex task on the stage branch carrying its Blocking
      findings, and its result lands as new commits on the same pull request.
- [ ] `Input Required` and `Complete` audits start nothing and show that a
      person is needed.
- [ ] A review-refinement report returns to the Refine stage the same way.
- [ ] Duplicate, replayed, stale-head and other-author reports are ignored.
- [ ] After the configured number of automatic returns, nothing more starts,
      including after a restart.
- [ ] A missed event is recovered by the next sync.

## Deferrals

None.

## Traces to

`workplan-control-panel/codex-interfaces.md`;
`workplan-control-panel/README.md`; `skills/audit-task/SKILL.md`;
`skills/review-refinement/SKILL.md`; `skills/audit-finalize/SKILL.md`;
ADR-0004; `docs/testing.md` (Control panel gate); P0-010; P0-011.
