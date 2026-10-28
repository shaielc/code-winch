# P1-007: Plan Phase 1 increment 4: harness output area

**Phase:** 1 — Sandbox exchange
**Shape:** planning
**Dependencies:** P0-001 (semantic: the assigned brief consumes the reviewed
increment boundary and concern ownership in the Phase 1 design produced by
P0-001)

## Objective

Produce one complete implementation brief, P1-008, for Phase 1 increment 4,
**Harness output area**, together with its tracker and workplan records. The
brief must add a separately identifiable, truthfully empty output region to the
existing attach page without generating, reading, transporting, or implying
harness output.

## Source baseline

At this branch's HEAD, `web/src/attach/App.tsx` renders posture, a child-status
section, browser-local messages, and the composer in one component. It fetches
`GET /api/session` once, but has no output state or output region. Its tests in
`App.test.tsx` already query the child-status and local-message sections by
accessible region name and exercise posture, status, and local submission.
`style.css` supplies the shared section treatment and an `.empty-state` style,
so this increment does not require a new style or component solely to create
the semantic region.

The attach page is built into the sandbox image and is reachable in an isolated
branch environment through the proxy path documented in
`deployments/README.md`. The current standing Go scenario inspects the static
HTML shell and HTTP behavior; it does not render React and cannot establish the
new region's DOM semantics. P1-006 is pending child-process work, but increment
4 consumes neither its brief nor its predicted behavior. P0-001 is completed,
and its Phase 1 design contains the increment boundary this planning task uses.

## Scope

Use `skills/derive-task/SKILL.md`, the reviewed Phase 1 design, and the source
baseline above to derive exactly P1-008. In that brief:

- require `App.tsx` to render a `section` named by a visible `Harness output`
  heading, following the page's existing `aria-labelledby` section pattern;
- fix the initial and only text in that region to **`Harness output will appear
  here when it is received.`** This describes the presentation destination and
  a future receipt condition; it does not assert that a child exists, emitted
  nothing, is running or ready, accepts input, or exited;
- keep the region present from the initial render, independently of session
  fetch success, child status, local messages, and browser timing;
- add a focused semantic case to `App.test.tsx` that finds the region by its
  accessible name, checks the exact empty text, and rejects lifecycle,
  readiness, input-availability, and received-output claims; and
- preserve the one-shot session fetch, child-status identity and values,
  posture and error disclosure, local composer behavior, loopback-only
  deployment, HTTP surface, and existing tests.

Forecast only the existing `App.tsx` and `App.test.tsx` files unless inspection
at derivation time shows that styling is necessary for the output area to be
usable. Reuse the shared section and empty-state presentation rather than
requiring a speculative output model or component seam. The implementer may
choose incidental markup details that do not change these observable
semantics.

## Non-goals

This task writes planning artifacts only. P1-008 must not start or inspect a
child, manufacture sample output, read child bytes, add browser output state,
poll or stream, change an HTTP contract, or introduce records, ordinals,
history, persistence, input delivery, provider decoding, terminal rendering,
or a broad page restructure. Those are outside increment 4; they are not
completion work deferred by this brief.

Derive one implementation brief only. Do not reserve an implementation ID for
increment 5 or any later increment.

## Planning outputs and handoff

Create `docs/workplan/phase-1/P1-008-harness-output-area.md` and its matching
tracker entry. Reconcile its dependency reason, Phase 1 coverage, Phase 0
coverage, index row, graph measures, write collisions, and contract-collision
analysis in `docs/workplan/README.md`, `docs/workplan/phase-0/README.md`, and
`docs/workplan/phase-1/README.md`. P1-008 has no product dependency: the attach
page exists at HEAD and increment 4 is independent of increments 1–3.

In the same planning change, complete
`skills/phase-increment-handoff/SKILL.md`: create P1-009 as the sole complete
planning-shaped task for increment 5, **First live child output**, with its
tracker, index, and coverage records. Derive P1-009's planning dependencies
from the planning tasks whose artifacts it consumes; do not reserve increment
5's implementation ID. The workflow then hands P1-008 to a different refiner
using `skills/refine-task/SKILL.md`.

## Write set

- `docs/workplan/phase-1/P1-008-harness-output-area.md`
- `docs/workplan/phase-1/P1-009-plan-first-live-child-output.md`
- `docs/workplan/tasks.json` (new entries only; existing statuses unchanged)
- `docs/workplan/README.md`
- `docs/workplan/phase-0/README.md`
- `docs/workplan/phase-1/README.md`

## Contract surfaces

This planning task changes no product contract. P1-008 will own the attach
page's accessible `Harness output` region and the exact empty-state sentence.
It consumes only the existing page and must declare no HTTP, child-lifecycle,
output-transport, record, input, persistence, or provider contract.

## Demonstration

Review P1-008 and its matching tracker, index, and coverage records. Confirm
that the predicted manual check starts the isolated branch deployment with
`make test-env`, opens the exact branch proxy URL formed as documented in
`deployments/README.md`, and finishes with `make test-env-down`. It must direct
a person to observe the independently named `Harness output` region and exact
sentence, then confirm that child status, posture, and the usable local composer
remain present. It must explicitly say this observation proves presentation
only, not that a child emitted no bytes or that output was read or transported.

Confirm P1-008 uses its one real ID consistently, has no dependency, and owns
only increment 4. Confirm P1-009 is the single next planning handoff, its
dependencies name the planning artifacts it consumes, and no increment 5
implementation ID or later ID is present.

## Verification

- Parse `docs/workplan/tasks.json`, validate it against
  `docs/workplan/tasks.schema.json`, and assert that every brief path resolves.
- Run `bash scripts/list-available-tasks.sh` and confirm P1-008's availability
  follows from its empty dependency set while P1-009 follows its recorded
  planning dependencies.
- Search the tracker, index, Phase 0/1 coverage, P1-008, and P1-009 for the
  allocated IDs and confirm their titles, paths, dependency reasons, and
  increment ownership agree without another reserved ID.
- Recompute and review the graph measures, incomparable pairs, available tasks,
  write collisions, and contract collisions recorded in the index.

P1-008's brief must name one focused React semantic test for the region's
accessible name, exact empty sentence, unconditional initial presence, and
absence of false lifecycle/readiness/input/output claims. It must also name the
existing `App.test.tsx` cases as preservation evidence and point to the Web and
Go gates in `docs/testing.md` rather than duplicate their command lists.

## Acceptance criteria

- [ ] P1-008 is one complete implementation brief for increment 4, with no
      dependency and a write forecast limited to the existing React page and
      its semantic component test unless the brief evidences a necessary style
      edit.
- [ ] P1-008 fixes the accessible region name and exact empty sentence above,
      requires unconditional initial presentation, and names observable test
      failures for every false claim excluded by the objective.
- [ ] P1-008 preserves the current page, HTTP, deployment, and test behavior;
      its verification and isolated deployed-page demonstration distinguish a
      rendered React observation from static-shell inspection.
- [ ] Tracker, index, Phase 0/1 coverage, graph measures, and collision records
      agree, preserve existing statuses, and contain P1-009 as the sole complete
      planning handoff for increment 5 without reserving its implementation ID
      or any later ID.

## Deferrals

None.

## Traces to

`docs/workplan/phase-1/README.md` increment 4; Phase 0 increment Plan Phase 1;
`docs/roadmap.md` Stage 0; `docs/contracts.md` §§8–9; ADR-0005/0006;
`skills/shared/workplan-model.md` planning shape and I1–I7.
