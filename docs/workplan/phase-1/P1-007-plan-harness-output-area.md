# P1-007: Plan Phase 1 increment 4: harness output area

**Phase:** 1 — Sandbox exchange
**Shape:** planning
**Dependencies:** P0-001 (semantic: the assigned brief consumes the reviewed
increment boundary and concern ownership in the Phase 1 design produced by
P0-001)

## Objective

Produce one complete implementation brief for Phase 1 increment 4, **Harness
output area**, and reconcile it with the tracker, workplan index, and Phase 0/1
coverage. The brief must specify an independently identifiable output region on
the existing attach page whose empty presentation is truthful without claiming
that harness output has been generated, read, transported, or interpreted.

## Scope

Use `skills/derive-task/SKILL.md` against the reviewed Phase 1 design and the
code at HEAD, and derive exactly increment 4. Allocate `P1-008` to the
implementation brief only after the brief and its tracker entry are complete.

Write the derived brief against these observed attach-page surfaces:

- `web/src/attach/App.tsx` is the page composition root. It currently renders
  posture when the one-shot `GET api/session` fetch supplies it, an always
  present `Child status` region, an always present browser-local `Local
  messages` region, and `Composer`. Direct the implementation to add the output
  region here, alongside those existing page sections, without new state,
  effects, fetches, or child-status conditions. Keeping the region static is
  what makes this presentation increment independent of increments 1–3 and of
  network timing.
- Require the accessible region name **Harness output**, supplied by a visible
  heading and `aria-labelledby`, and require the exact empty-state text
  **Harness output has not been connected yet.** This describes the product
  capability that is absent; it does not claim that a child produced nothing,
  that the child is ready or running, or that the browser completed a read.
- Reuse the existing `section` and `.empty-state` presentation in
  `web/src/attach/style.css`. Forecast a style-file change only if the derived
  implementation needs output-area-specific layout that the current shared
  rules cannot supply; visual redesign is not a reason to change it.
- Extend `web/src/attach/App.test.tsx`, where page-region semantics and the
  existing child-status, posture, and local-composer preservation behavior are
  already tested. The focused test must find the `Harness output` region by
  role and accessible name, assert the exact empty-state sentence within that
  region while the session fetch remains pending, and reject false lifecycle,
  readiness, emitted-output, and successful-read claims. It must also show the
  region remains unchanged after a successful session response, a rejected
  response, and a local composer submission; existing focused tests remain the
  preservation evidence for the details of those features.
- Preserve relative-path deployment through the existing Vite build and
  branch-isolated proxy. The derived brief's hand check starts the isolated
  environment with `make test-env`, opens the proxy URL described by
  `deployments/README.md`, inspects the rendered React page, and always runs
  `make test-env-down`. Static HTML inspection in the Go scenario is not
  evidence for a React-rendered region.

Exclude increment 5's child-byte reading and first live transport and increment
6's incremental delivery. Do not make the fixture emit content to demonstrate
an empty presentation; no fake-harness change is needed. Preserve the existing
one-shot session fetch, child-status identity and values, posture disclosure,
browser-local composer, loopback-only serving, deployment, and standing tests.
The derived task has no dependency on the increment 1–3 tasks because it
consumes only behavior already present at HEAD and defines a distinct
presentation surface.

After deriving increment 4, complete
`skills/phase-increment-handoff/SKILL.md`: allocate `P1-009` to the sole planning
task for increment 5, **First live child output**, and add its complete brief
and matching planning records. Give `P1-009` semantic edges to `P0-001`,
`P1-005`, and `P1-007`: it consumes the phase design and the derivation
artifacts for prerequisite increments 3 and 4, never the implementation tasks.
Do not reserve increment 5's implementation ID.

## Required constraints

- The derived brief owns only the accessible name **Harness output** and the
  exact empty-state sentence **Harness output has not been connected yet.** It
  must not describe the empty state as “no output,” “waiting,” “connected,” or
  similar evidence about a child or transport that this increment lacks.
- The implementation forecast is `web/src/attach/App.tsx` and
  `web/src/attach/App.test.tsx`, plus `web/src/attach/style.css` only if the
  existing shared section/empty-state rules prove insufficient during
  implementation.
- The brief changes no HTTP route, session payload, Go source, fake-harness
  behavior, deployment configuration, child lifecycle, or composer behavior.
- Separate observations at refinement HEAD from predicted implementation and
  runtime results. `App.tsx` currently has no harness-output region, and the
  current Go e2e scenario verifies served static assets rather than the
  rendered React DOM; neither is evidence that the predicted region exists.
- Verification in the derived brief names focused checks only and points to
  `docs/testing.md` for repository gates rather than copying its gate list.

## Non-goals

This task changes planning artifacts only. Its derived implementation brief
does not generate or read child output; add polling, streaming, or another HTTP
contract; define output records, ordinals, persistence, history, or provider
semantics; alter child lifecycle or input; introduce terminal rendering; or
restructure the page. It derives one implementation brief and reserves no
implementation ID for increment 5 or any later increment.

## Planning outputs and handoff

Create all of the following as one reconciled artifact set:

- `docs/workplan/phase-1/P1-008-harness-output-area.md`, the complete
  implementation brief for increment 4;
- the matching `P1-008` tracker entry;
- reconciled dependency, coverage, graph-metric, write-collision, and
  contract-collision records in `docs/workplan/README.md`,
  `docs/workplan/phase-0/README.md`, and
  `docs/workplan/phase-1/README.md`; and
- `docs/workplan/phase-1/P1-009-plan-first-live-child-output.md` with its
  matching tracker/index/coverage records, as the sole complete planning task
  handed increment 5 by `skills/phase-increment-handoff/SKILL.md`.

The `P1-008` brief must record the concrete accessibility boundary, exact
language, implementation forecast, focused test, deployed hand path, and
evidence limits above. Handoff after this task is to a different refiner, who
applies `skills/refine-task/SKILL.md` to `P1-008`; that later operation is not
an output or acceptance criterion of this task.

## Write set

- `docs/workplan/phase-1/P1-008-harness-output-area.md`
- `docs/workplan/phase-1/P1-009-plan-first-live-child-output.md`
- `docs/workplan/tasks.json` (new entries only; existing statuses unchanged)
- `docs/workplan/README.md`
- `docs/workplan/phase-0/README.md`
- `docs/workplan/phase-1/README.md`

## Contract surfaces

This planning task changes no product contract. Its derived brief will own the
attach page's accessible **Harness output** region and exact empty-state
presentation. It consumes the existing attach-page composition only; it must
not own or change the session HTTP contract, child lifecycle/status, output
transport or records, input, persistence, or provider interpretation.

## Demonstration

Review the task diff and the resulting tracker/index/coverage records. Confirm
that `P1-008` is the sole implementation ID for increment 4; it has no task
dependency; its scope, tests, and predicted deployed-page hand check all use
the exact accessible name and empty-state sentence above; and the check clearly
distinguishes a rendered React observation from static HTML inspection.

Confirm the predicted hand check uses `make test-env`, the branch proxy URL
documented in `deployments/README.md`, and `make test-env-down`; preserves the
visible child-status, posture, local-message, and composer surfaces; and makes
no claim that child bytes were generated or transported. Confirm `P1-009` is
the sole complete planning task for increment 5 and that no increment 5
implementation ID or later ID was allocated.

At refinement HEAD these outputs do not yet exist, so this artifact review is
predicted for the planning implementation rather than reported as a passing
runtime observation.

## Verification

- Validate `docs/workplan/tasks.json` against
  `docs/workplan/tasks.schema.json`, including unique IDs, resolvable brief
  paths, valid dependency targets, and an acyclic dependency graph.
- Run `bash scripts/list-available-tasks.sh` and confirm its result agrees with
  the reconciled graph.
- Search for `P1-007`, `P1-008`, and `P1-009` across the tracker, index, Phase
  0/1 coverage, and briefs; confirm the current IDs are consistent, the index's
  explicitly labelled former `P1-009` alias remains distinguishable from the
  new current ID, and no later Phase 1 ID was allocated.
- Inspect the diff to confirm only the six forecast planning files changed and
  no existing tracker status changed.
- Use `docs/testing.md` for the repository gates; do not duplicate them here.

## Acceptance criteria

- [ ] One complete `P1-008` implementation brief and matching tracker entry
      cover increment 4, with no dependency on increments 1–3 and no product
      implementation in this planning change.
- [ ] The brief requires an always-present region named **Harness output** and
      the exact sentence **Harness output has not been connected yet.**, plus a
      focused semantic test that catches absence, misnaming, changed language,
      conditional rendering, and false output/readiness/lifecycle claims.
- [ ] The brief preserves the current session fetch, child-status, posture,
      local-message/composer, deployment, and test behavior, and excludes all
      output generation, reading, transport, record, and persistence work.
- [ ] The brief's predicted hand check observes the rendered page through the
      isolated proxy, uses the documented setup/teardown commands, and does not
      treat static HTML or the empty state as evidence about harness activity.
- [ ] Tracker, index, Phase 0/1 coverage, graph metrics, and collision records
      agree, and all pre-existing tracker statuses are byte-for-byte unchanged.
- [ ] `P1-009` is the sole complete planning task for increment 5, with
      semantic dependencies on `P0-001`, `P1-005`, and `P1-007`; no increment
      5 implementation ID or later Phase 1 ID is reserved.

## Deferrals

None. Live child output and incremental transport are outside this decided
increment, not unfinished completion work delegated by it.

## Traces to

`docs/workplan/phase-1/README.md` increment 4 and concern boundaries;
`docs/workplan/phase-0/README.md` Plan Phase 1 coverage; `docs/roadmap.md` §§1–3
Stage 0; `docs/contracts.md` §§8–9; ADR-0005; ADR-0006;
`skills/shared/workplan-model.md` planning shape and I1–I7.
