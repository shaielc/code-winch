# P1-007: Plan Phase 1 increment 4: harness output area

**Phase:** 1 — Sandbox exchange
**Shape:** planning
**Dependencies:** P0-001 (semantic: the assigned brief consumes the reviewed
increment boundary and concern ownership in the Phase 1 design produced by
P0-001)

## Objective

Produce one complete, independently authored implementation brief for Phase 1
increment 4, **Harness output area**, and reconcile it with the tracker,
workplan index, and Phase 0/1 coverage. The derived capability must add a
truthful, independently identifiable empty output region to the existing attach
page without generating, transporting, or implying harness output.

## Scope

Use `skills/derive-task/SKILL.md` against the reviewed Phase 1 design and code
at HEAD. Derive exactly increment 4. Inspect the existing attach-page
composition, accessibility semantics, styles, component tests, web build, and
deployed static route before selecting the smallest output-region presentation
and maintained hand path.

This increment is independent of increments 1–3: consume the Phase 1 design,
not child-status implementation or process behavior. Preserve whatever truthful
child-status presentation exists at the derivation revision, but do not make
the empty state conditional on child state, output, browser timing, or an API
response. Exclude increment 5's child-byte reading and live transport and
increment 6's incremental delivery.

Allocate the next Phase 1 implementation ID only when its complete brief and
matching tracker entry exist. Reconcile dependency reasons, coverage, graph
metrics, and collision analysis in the workplan index and Phase 0/1 records.
Arrange handoff to a different refiner using `skills/refine-task/SKILL.md`.

## Required constraints

- The derived brief must define accessible label/region semantics and exact
  truthful empty-state language that cannot be mistaken for proof that a child
  emitted no output, is ready, is running, accepts input, or has exited.
- Preserve the existing one-shot session fetch, child-status identity, posture
  disclosure, local composer, loopback-only serving, deployment, and existing
  tests. This presentation task must not add or change an HTTP contract.
- Distinguish source-observed behavior from predicted implementation results.
  The current Go e2e scenario's static HTML inspection does not prove the React
  DOM; require a semantic component test and deployed-page hand observation.
- Keep the output region bounded to presentation and empty-state behavior. Do
  not create speculative records, fake output, transport seams, polling,
  streaming, persistence, ordinals, or provider semantics.

## Non-goals

This task writes planning artifacts only. Its target increment excludes child
output generation or reading, live/incremental transport, output records,
history, persistence, input, provider decoding, terminal rendering, and broad
page restructuring. Derive one implementation brief only and reserve no later
implementation ID.

## Planning outputs and handoff

Create `docs/workplan/phase-1/<allocated-ID>-harness-output-area.md` and its
matching tracker entry, then reconcile `docs/workplan/README.md`,
`docs/workplan/phase-0/README.md`, and `docs/workplan/phase-1/README.md`. Record
the concrete accessibility/presentation boundary, test forecast, deployed hand
path, and evidence limits rather than assuming output transport exists.

Complete `skills/phase-increment-handoff/SKILL.md` by adding the sole complete
planning-shaped task for Phase 1 increment 5 and its matching tracker/index/
coverage records; do not reserve increment 5's implementation ID. A different
refiner then takes the derived increment 4 brief with
`skills/refine-task/SKILL.md`.

## Write set

- `docs/workplan/phase-1/<allocated-ID>-harness-output-area.md`
- `docs/workplan/phase-1/<next-ID>-plan-first-live-child-output.md`
- `docs/workplan/tasks.json` (new entries only; existing statuses unchanged)
- `docs/workplan/README.md`
- `docs/workplan/phase-0/README.md`
- `docs/workplan/phase-1/README.md`

## Contract surfaces

This planning task changes no product contract. Its derived brief will own the
attach page's accessible harness-output region and truthful empty-state
presentation. It consumes only the existing attach page and Phase 1 design; it
must not own or change the session wire contract, child lifecycle/status,
output transport or records, input, persistence, or provider interpretation.

## Demonstration

Review the derived brief and matching tracker/index/coverage records. Confirm
that its predicted branch-isolated hand check uses `make test-env`, the proxy
URL documented in `deployments/README.md`, and `make test-env-down`; observes
an independently labeled harness-output region with the exact truthful empty
state; preserves child status, posture, and the usable local composer; and does
not claim that any child bytes were read or transported. Confirm all records
use one real implementation ID and dependency set and no later implementation
ID has been reserved.

## Verification

Run completion-script and control-panel tests, tracker JSON validation,
`scripts/list-available-tasks.sh`, an `rg` consistency search for this planning
task and both newly allocated IDs, `git diff --check`, `make check`, and
`make test-cycle`. Report unavailable Docker/toolchains as limitations.

The derived brief must name a focused React semantic test for the labeled
region, exact empty state, and absence of false output/readiness/lifecycle
claims; preservation checks for child status, posture, and composer; all web
format/lint/typecheck/test/build gates; the standard repository gates and
standing e2e scenario; and a deployed-page observation that does not confuse
static HTML inspection with a rendered React assertion.

## Acceptance criteria

- [ ] Exactly one complete Phase 1 implementation brief covers increment 4
      with one real ID and no dependency on increments 1–3.
- [ ] The brief defines an accessible, truthful empty output area, preservation
      requirements, frontend tests, deployed reachability, and evidence limits
      without absorbing live or incremental output.
- [ ] Tracker, index, Phase 0/1 coverage, graph metrics, and collision warnings
      agree and preserve all existing statuses.
- [ ] The sole complete planning task for increment 5 is added with its
      planning dependency derived from actual consumed artifacts; no increment
      5 implementation ID or later ID is reserved.
- [ ] A different refiner receives the finished increment 4 brief, and artifact
      review demonstrates every output of this planning task.

## Deferrals

None. Live child output and incremental transport are excluded by the decided
increments 5 and 6, not unfinished completion work for the output-area
presentation.

## Traces to

`docs/workplan/phase-1/README.md` increment 4; Phase 0 increment Plan Phase 1;
`docs/roadmap.md` Stage 0; `docs/contracts.md` §§8–9; ADR-0005/0006;
`skills/shared/workplan-model.md` planning shape and I1–I7.
