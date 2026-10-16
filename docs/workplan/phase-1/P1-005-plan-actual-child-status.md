# P1-005: Plan Phase 1 increment 3: actual child status

**Phase:** 1 — Sandbox exchange
**Shape:** planning
**Dependencies:** P1-003 (semantic: consume the P1-004 brief produced by P1-003, which defines the sandbox session `childStatus` contract and `Down` presentation)

## Objective

Produce one complete, independently authored implementation brief, P1-006,
for Phase 1 increment 3, **Actual child status**, together with its tracker
entry and the next planning handoff.

## Source baseline

At this PR's HEAD, `cmd/winch-sandbox/main.go` constructs the attach HTTP
handler and server but starts no child. The attach server stores only posture,
the React app fetches the session once, and the image ships
`/usr/local/bin/fake-harness`.

P1-004's session response, validation, and `Down` presentation are predicted
by its pending brief. Recheck the source when deriving P1-006 and distinguish
observations at that revision from behavior predicted by either brief.

## Scope

Use `skills/derive-task/SKILL.md`, the Phase 1 design, P1-004's brief, and
code at HEAD to derive P1-006. The implementation must launch the shipped fake
from the sandbox and connect observations of that owned process to the session
endpoint and existing attach page.

Specify the launch surface, process bookkeeping, and status wiring needed for
the contract below. Preserve the existing one-shot page fetch. Plan against
P1-004's brief without waiting for implementation; give P1-006 a `contract`
dependency on P1-004 because it extends that task's wire and presentation
contract.

## Contract surfaces and constraints

This planning task changes no product contract. The derived implementation
extends P1-004's `GET /api/session` `childStatus` member:

- `Running`: the owned child started successfully and the sandbox has not
  observed its exit.
- `Exited`: the sandbox has observed that child's exit.

A failed start cannot produce `Running`. Status comes from the owned process,
not browser timing, output, or host process-table scans, and conveys no health,
readiness, or input-availability claim. Include necessary exit observation and
process bookkeeping in the implementation forecast.

The implementation must document the wire extension and preserve P1-004's
`Down` presentation and `Unknown` fallback for missing or invalid values.
It must also preserve loopback-only serving, posture disclosure, the local
composer, deployment, and existing tests.

## Non-goals

This task writes planning artifacts only. The target increment excludes the
manual `Exited` demonstration, harness input or commands (including an
exit/shutdown command), output transport or display, persistence, restart/retry
policy, provider semantics, and broad page restructuring.

Do not derive a second implementation brief or reserve later implementation IDs.

## Planning outputs and handoff

Create `docs/workplan/phase-1/P1-006-actual-child-status.md` and its matching
tracker entry together. The brief must contain a concrete implementation write
forecast, contract documentation changes, tests, and the manual check described
below.

Reconcile dependency reasons, coverage, graph metrics, and collision analysis
in `docs/workplan/README.md`, `docs/workplan/phase-0/README.md`, and
`docs/workplan/phase-1/README.md`. Record P1-006 against increment 3 and its
contract collision with P1-004. Preserve all existing tracker statuses.

Complete `skills/phase-increment-handoff/SKILL.md` by creating the complete
planning-shaped `docs/workplan/phase-1/P1-007-plan-harness-output-area.md`
and its pending tracker, index, and coverage records. P1-007 plans increment 4,
**Harness output area**, consumes only the Phase 1 design, and depends on
P0-001. A different refiner then takes P1-006 with
`skills/refine-task/SKILL.md`.

## Write set

- `docs/workplan/phase-1/P1-006-actual-child-status.md`
- `docs/workplan/phase-1/P1-007-plan-harness-output-area.md`
- `docs/workplan/tasks.json` (new entries only; existing statuses unchanged)
- `docs/workplan/README.md`
- `docs/workplan/phase-0/README.md`
- `docs/workplan/phase-1/README.md`

## Demonstration

For P1-005, review the generated artifacts against the acceptance criteria.
Runtime behavior is a prediction to be checked during P1-006 implementation.

P1-006's brief must provide a repeatable manual check using `make test-env`,
the branch proxy URL from `deployments/README.md`, and
`make test-env-down`. Explain how the chosen fake configuration keeps the
child running long enough for endpoint and page observation without harness
input or commands. Wait for the session endpoint to report `Running`, then
load or reload the attach page and confirm the same status, posture disclosure,
and local composer behavior. Include environment teardown. Never replace or
tear down the operator's default `code-winch` deployment.

## Verification

For the planning change, run completion-script and control-panel tests, tracker
JSON validation, `scripts/list-available-tasks.sh`, an `rg` consistency search
for P1-005/P1-006/P1-007, `git diff --check`, `make check`, and
`make test-cycle`. Report unavailable Docker/toolchains as limitations.

P1-006's brief must name focused tests for successful and failed launch,
observed child exit, session wire values, and React validation/presentation.
Include checks for the preservation constraints above, the standing e2e
scenario, and Go/web/build gates for every forecast implementation surface.

## Acceptance criteria

- [ ] P1-006 exists as a complete implementation brief with a concrete write
      forecast and contract documentation changes. Its status definitions and
      preservation requirements match this brief's contract section.
- [ ] P1-006 names tests for every case listed under Verification and gives
      executable verification commands for its forecast surfaces.
- [ ] P1-006 contains the per-branch manual check, identifies the fake
      configuration that keeps `Running` observable, and explains why it works.
      Its demonstration and implementation requirements respect Non-goals.
- [ ] P1-006's tracker entry has a `contract` dependency on P1-004. The index,
      Phase 0/1 coverage, dependency reasons, metrics, and collision records
      agree with the tracker; all pre-existing statuses remain unchanged.
- [ ] P1-007 is the sole complete planning task for increment 4, has a pending
      tracker entry depending only on P0-001, and appears in the index and
      Phase 0/1 coverage. No later implementation ID is reserved.

## Deferrals

None.

## Traces to

`docs/workplan/phase-1/README.md` increment 3; Phase 0 increment Plan Phase 1;
`docs/roadmap.md` Stage 0; `docs/contracts.md` §§8–9; ADR-0005/0006;
`skills/shared/workplan-model.md` planning shape and I1–I7.
