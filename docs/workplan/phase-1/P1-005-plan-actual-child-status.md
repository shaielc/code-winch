# P1-005: Plan Phase 1 increment 3: actual child status

**Phase:** 1 — Sandbox exchange
**Shape:** planning
**Dependencies:** P1-003 (semantic: consume the P1-004 brief produced by P1-003 and extend its sandbox session `childStatus` contract and `Down` presentation)

## Objective

Produce one complete implementation brief, P1-006, for launching the shipped
fake and showing its actual child status on the attach page: `Running` after
successful launch while no exit has been observed, and `Exited` after the
sandbox observes the child's exit. Reconcile the brief with the tracker, workplan
index, and Phase 0/1 coverage. Demonstrating `Exited` and implementing harness
input or commands are outside this task's scope.

## Scope

Use `skills/derive-task/SKILL.md`, the Phase 1 design, P1-004's brief, and code
at HEAD to derive P1-006.

At this PR's HEAD, `cmd/winch-sandbox/main.go` constructs the attach HTTP
handler and server but starts no child. The attach server stores only posture,
the React app fetches the session once, and the image ships
`/usr/local/bin/fake-harness`. P1-004's `Down` response, validation, and
presentation are predictions from its pending brief, not implemented behavior
at this baseline. Recheck these observations when deriving the brief.

P1-006 must make the sandbox launch the shipped fake and connect the owned
process's start and exit observations to the session response and page.
A successful start permits `Running` until the sandbox observes the child's
exit; the recorded status then becomes `Exited`. A failed start never permits
`Running`. These values describe process observations, not health or readiness.
The brief must specify the process bookkeeping and focused tests needed to
keep those observations accurate, without browser inference or process-table
scans.

Plan against P1-004's brief without waiting for its implementation. Label
predicted behavior and give P1-006 a `contract` dependency on P1-004 because it
extends that task's exact wire and presentation contract. Keep the existing
one-shot page fetch: the derived hand check waits for the session endpoint to
report `Running`, then loads or reloads the page to fetch that state.

Allocate P1-006 only with its complete brief and matching tracker entry.
Reconcile dependency reasons, coverage, graph metrics, and collision analysis.
Then complete `skills/phase-increment-handoff/SKILL.md` by creating P1-007,
the planning brief for increment 4, **Harness output area**. It consumes only
the Phase 1 design and depends on P0-001.

## Required constraints

- Distinguish source observations from predictions in P1-004 and P1-006.
- Preserve loopback-only serving, posture disclosure, the local composer,
  deployment, `Unknown` fallback, `Down` presentation, and existing tests.
- Document the `childStatus` contract extension in the implementation change.
  Require focused checks that a failed launch cannot be presented as `Running`
  and that missing or invalid wire values retain the `Unknown` fallback.
- Keep the derived capability focused on launch and actual child-status wiring.
  Necessary exit observation, reaping, and process bookkeeping remain in scope.

## Non-goals

This task writes planning artifacts only. The derived task excludes demonstrating
`Exited` and implementing harness input or commands, including an exit/shutdown
command. Output transport or display, persistence, restart/retry policy, provider
semantics, and broad page restructuring are also outside the increment.
Choose process bookkeeping and its tests according to the launch/status
implementation; this brief imposes no separate harness-stop policy.
Do not derive a second implementation brief or reserve later implementation IDs.

## Planning outputs and handoff

Create `docs/workplan/phase-1/P1-006-actual-child-status.md` and its tracker
entry. Reconcile `docs/workplan/README.md`, `docs/workplan/phase-0/README.md`,
and `docs/workplan/phase-1/README.md`. Record the actual child-status contract,
launch surface, write forecast, and the boundary between status behavior and
its manual demonstration.
Preserve all existing statuses.

Create the complete planning-shaped
`docs/workplan/phase-1/P1-007-plan-harness-output-area.md` and its pending
tracker/index/coverage records as the handoff skill requires. A different
refiner then takes P1-006 with `skills/refine-task/SKILL.md`.

## Write set

- `docs/workplan/phase-1/P1-006-actual-child-status.md`
- `docs/workplan/phase-1/P1-007-plan-harness-output-area.md`
- `docs/workplan/tasks.json` (new entries only; existing statuses unchanged)
- `docs/workplan/README.md`
- `docs/workplan/phase-0/README.md`
- `docs/workplan/phase-1/README.md`

## Contract surfaces

This planning task changes no product contract. P1-006 will extend P1-004's
`GET /api/session` `childStatus` member with `Running` and `Exited`, based on
the owned child's start and observed exit. Record the collision and `contract`
edge on P1-004. Status behavior includes observed exit; its hand demonstration
does not. The contract conveys no output, input availability, or readiness.

## Demonstration

For P1-005, review the derived brief and planning records for completeness and
consistency. Check that the brief defines the start/exit observation path and
contains the following manual check for the future P1-006 implementation.

For P1-006, the predicted manual check starts the per-branch environment with
`make test-env`, waits for `GET /api/session` through the branch proxy to report
`Running`, and loads or reloads the attach page to show the same state. It also
checks existing posture and local composer behavior, then uses
`make test-env-down` for environment teardown. Use the proxy URL from
`deployments/README.md`; never replace or tear down the operator's default
`code-winch` deployment. The manual check does not make the fake exit or
require observing `Exited`.

## Verification

Run completion-script and control-panel tests, tracker JSON validation,
`scripts/list-available-tasks.sh`, an `rg` consistency search for
P1-005/P1-006/P1-007, `git diff --check`, `make check`, and `make test-cycle`.
Report unavailable Docker/toolchains as limitations.

The derived brief must name focused launch/status tests, attach contract tests,
React validation/presentation tests, the standing e2e scenario, and Go/web/build
gates for its forecast implementation surfaces. Keep those checks within the
launch and actual child-status capability.

## Acceptance criteria

- [ ] One complete brief, P1-006, defines launch, failed-start handling, and
      truthful `Running`/`Exited` observations from the owned process handle,
      takes a `contract` dependency on P1-004, and labels predictions.
- [ ] The brief includes focused status-observation tests and a repeatable
      `Running` page demonstration, with `Exited` demonstration and harness
      input/commands outside scope.
- [ ] It covers failed-launch status, wire validation, contract documentation,
      existing page behavior, and safe per-branch deployment checks.
- [ ] Tracker, index, Phase 0/1 coverage, graph metrics, and collision warnings
      agree, preserve statuses, and distinguish status coverage from
      demonstration.
- [ ] P1-007 is the sole complete derivation task for increment 4, depends only
      on P0-001, and has matching pending tracker, index, and coverage records.
      No later implementation ID is reserved.

## Deferrals

None.

## Traces to

`docs/workplan/phase-1/README.md` increment 3;
Phase 0 increment Plan Phase 1; `docs/roadmap.md` Stage 0;
`docs/contracts.md` §§8–9; ADR-0005/0006;
`skills/shared/workplan-model.md` planning shape and I1–I7.
