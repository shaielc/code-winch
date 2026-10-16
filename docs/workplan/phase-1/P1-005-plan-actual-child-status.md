# P1-005: Plan Phase 1 increment 3: actual child status

**Phase:** 1 — Sandbox exchange
**Shape:** planning
**Dependencies:** P1-003 (semantic: consume the P1-004 brief produced by P1-003 and extend its sandbox session `childStatus` contract and `Down` presentation)

## Objective

Produce one complete implementation brief, P1-006, for launching the shipped
fake and showing truthful sandbox-owned `Running` status on the attach page.
Reconcile the brief with the tracker, workplan index, and Phase 0/1 coverage.

The review narrows this task from the phase increment's original
`Running`/`Exited` objective. Exit observation and an `Exited` demonstration
are outside this task's completion criteria; record that scope difference in
the planning records rather than claiming the full original increment is covered.

## Scope

Use `skills/derive-task/SKILL.md`, the Phase 1 design, P1-004's brief, and code
at HEAD to derive P1-006.

At this PR's HEAD, `cmd/winch-sandbox/main.go` constructs the attach HTTP
handler and server but starts no child. The attach server stores only posture,
the React app fetches the session once, and the image ships
`/usr/local/bin/fake-harness`. P1-004's `Down` response, validation, and
presentation are predictions from its pending brief, not implemented behavior
at this baseline. Recheck these observations when deriving the brief.

P1-006 must make the sandbox launch the shipped fake and connect its own
knowledge of that child to the existing session response and attach page.
Report `Running` only from sandbox-owned observation, never from browser
timing, output, or a host process-table scan. The brief must explain how it
avoids claiming a successful launch when spawning fails. Do not prescribe
HTTP startup ordering or require sandbox startup to fail when the child cannot
be launched.

Plan against P1-004's brief without waiting for its implementation. Label
predicted behavior and give P1-006 a `contract` dependency on P1-004 because it
extends that task's exact wire and presentation contract. Keep the existing
one-shot page fetch; this capability needs no polling or streaming.

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
- Keep the derived capability limited to launch and running-status wiring.
  It must not require harness stopping, an exit control, or an exit demonstration.

## Non-goals

This task writes planning artifacts only. The derived task excludes harness-stop
and shutdown policies, exit observation and demonstration of `Exited`, an
exit/shutdown command, output transport or display, input delivery, persistence,
restart/retry policy, provider semantics, and broad page restructuring.
Do not add lifecycle requirements or tests for termination, reaping, signal
forwarding, or bounded child shutdown as conditions for this capability.
Do not derive a second implementation brief or reserve later implementation IDs.

## Planning outputs and handoff

Create `docs/workplan/phase-1/P1-006-actual-child-status.md` and its tracker
entry. Reconcile `docs/workplan/README.md`, `docs/workplan/phase-0/README.md`,
and `docs/workplan/phase-1/README.md`. Record the running-status contract,
launch surface, write forecast, and the reduced coverage of increment 3.
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
`GET /api/session` `childStatus` member with `Running` and connect it to
sandbox-owned child observation. Record the collision and `contract` edge
on P1-004. Do not add an `Exited` requirement or output/input/readiness meaning.

## Demonstration

Review the derived brief and planning records. Its predicted hand check must
show `Running` on the attach page after the sandbox launches the fake and
preserve the existing posture and local composer behavior. It must not require
making the fake exit or stopping it to complete the demonstration.

Use the per-branch `make test-env`, proxy URL, and `make test-env-down` path
from `deployments/README.md`; never replace or tear down the operator's default
`code-winch` deployment. These are predicted implementation checks. This
planning task demonstrates the completeness and consistency of the artifacts,
including the explicit limit on increment-3 coverage.

## Verification

Run completion-script and control-panel tests, tracker JSON validation,
`scripts/list-available-tasks.sh`, an `rg` consistency search for
P1-005/P1-006/P1-007, `git diff --check`, `make check`, and `make test-cycle`.
Report unavailable Docker/toolchains as limitations.

The derived brief must name focused launch/status tests, attach contract tests,
React validation/presentation tests, the standing e2e scenario, and Go/web/build
gates for its forecast implementation surfaces. Keep those checks within the
launch and running-status capability.

## Acceptance criteria

- [ ] One complete brief, P1-006, covers launch and truthful running status,
      takes a `contract` dependency on P1-004, and labels predictions.
- [ ] Its criteria and hand check require no harness-stop implementation,
      startup-ordering policy, fatal sandbox-startup behavior, or `Exited`
      demonstration.
- [ ] It covers failed-launch status, wire validation, contract documentation,
      existing page behavior, and safe per-branch deployment checks.
- [ ] Tracker, index, Phase 0/1 coverage, graph metrics, and collision warnings
      agree, preserve statuses, and disclose the narrowed increment-3 scope.
- [ ] P1-007 is the sole complete derivation task for increment 4, depends only
      on P0-001, and has matching pending tracker, index, and coverage records.
      No later implementation ID is reserved.

## Deferrals

None assigned by this task. The excluded lifecycle and exit work is not an
unfinished acceptance criterion of P1-006; the planning records must not claim
that the original increment's exit capability has been delivered.

## Traces to

`docs/workplan/phase-1/README.md` increment 3 (scope narrowed by PR #136 review);
Phase 0 increment Plan Phase 1; `docs/roadmap.md` Stage 0;
`docs/contracts.md` §§8–9; ADR-0005/0006;
`skills/shared/workplan-model.md` planning shape and I1–I7.
