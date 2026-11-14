# P1-009: Plan Phase 1 increment 5: first live child output

**Phase:** 1 — Sandbox exchange
**Shape:** planning
**Dependencies:** P0-001 (semantic: consume the reviewed Phase 1 design),
P1-005 (semantic: consume the P1-006 brief that defines owned fake-child
launch and process observation), P1-007 (semantic: consume the P1-008 brief
that defines the harness-output presentation identity)

## Objective

Produce one complete, independently authored implementation brief for Phase 1
increment 5, **First live child output**, and reconcile it with the tracker,
workplan index, and Phase 0/1 coverage. The derived capability must make raw
bytes emitted by the owned fake visible in the existing harness-output area.

## Scope

Use `skills/derive-task/SKILL.md` against the reviewed Phase 1 design, the
P1-006 and P1-008 briefs, and code at HEAD. Derive exactly increment 5. Recheck
which prerequisite briefs are implemented at the derivation revision, inspect
the child launch and pipe ownership, attach transport, browser presentation,
fake controls, CLI/operator surface, tests, and deployed route, and distinguish
source observations from predicted prerequisite behavior.

The derived task must consume the owned-child lifecycle and labelled output
area through explicit dependency edges to their implementation tasks. It must
include child-byte reading, minimum live transport, raw browser display, and a
maintained operator observation at this increment's depth. Exclude increment
6's guarantee that one piece becomes visible before a delayed second piece.

## Required constraints

- Treat child output as raw, untrusted bytes and do not infer JSON framing,
  provider events, semantic messages, prompt/reply correlation, readiness, or
  terminal meaning from the current fake dialect.
- Preserve owned-child status, the truthful output-area identity, one-shot
  posture fetch unless the minimum output transport requires a distinct path,
  local composer, loopback-only serving, deployment, and existing tests.
- Select a bounded minimum representation and transport. State precisely what
  the first slice does and does not prove about chunk boundaries, later bytes,
  malformed output, slow readers, replay, ordering, and persistence.
- Make the capability reachable through the running fake profile and the
  maintained attach surface; do not add a backend seam that no runtime
  configuration reaches.

## Non-goals

This task writes planning artifacts only. Its target increment excludes
incremental-before-completion delivery, retained history, input, semantic
decoding, provider adapters, terminal rendering, comprehensive malformed-output
recovery, slow-reader isolation, and durable records. Derive one implementation
brief only and reserve no later implementation ID.

## Planning outputs and handoff

Create `docs/workplan/phase-1/<allocated-ID>-first-live-child-output.md` and its
matching tracker entry, then reconcile `docs/workplan/README.md`,
`docs/workplan/phase-0/README.md`, and `docs/workplan/phase-1/README.md`. Record
the concrete process/output ownership, transport boundary, browser behavior,
operator path, test forecast, evidence limits, dependencies, and collisions.

Complete `skills/phase-increment-handoff/SKILL.md` by adding the sole complete
planning-shaped task for Phase 1 increment 6 and its matching tracker/index/
coverage records; do not reserve increment 6's implementation ID. A different
refiner then takes the derived increment 5 brief with
`skills/refine-task/SKILL.md`.

## Write set

- `docs/workplan/phase-1/<allocated-ID>-first-live-child-output.md`
- `docs/workplan/phase-1/<next-ID>-plan-incremental-output.md`
- `docs/workplan/tasks.json` (new entries only; existing statuses unchanged)
- `docs/workplan/README.md`
- `docs/workplan/phase-0/README.md`
- `docs/workplan/phase-1/README.md`

## Contract surfaces

This planning task changes no product contract. Its derived brief will consume
the owned-child lifecycle and harness-output presentation, and will own the
minimum live raw-output transport and presentation behavior it defines. It must
declare contract dependencies wherever it extends an existing wire operation
or transport surface.

## Demonstration

Review the derived brief and matching tracker/index/coverage records. Confirm
that its predicted branch-isolated hand check uses `make test-env`, the proxy
URL documented in `deployments/README.md`, and `make test-env-down`; makes
controllable raw fake output visible in the labelled region; preserves child
status, posture, and composer; and reports what the observation cannot prove.
Confirm all records use one real implementation ID and dependency set and no
later implementation ID has been reserved.

## Verification

Run completion-script and control-panel tests, tracker JSON validation,
`scripts/list-available-tasks.sh`, an `rg` consistency search for this planning
task and both newly allocated IDs, `git diff --check`, `make check`, and
`make test-cycle`. Report unavailable Docker/toolchains as limitations.

The derived brief must name focused tests for child-byte reading, transport,
raw browser rendering, bounds selected for this slice, child-exit/error paths,
and preservation of status, posture, and composer behavior; all affected web
and Go gates; the standing scenario; and a deployed-page observation distinct
from static HTML inspection.

## Acceptance criteria

- [ ] Exactly one complete Phase 1 implementation brief covers increment 5
      with one real ID and dependency edges to the implementation tasks whose
      child-ownership and output-area behavior it consumes.
- [ ] The brief defines reachable, bounded first raw output, preservation
      requirements, focused tests, a maintained hand path, and evidence limits
      without absorbing incremental delivery or later hardening.
- [ ] Tracker, index, Phase 0/1 coverage, graph metrics, and collision warnings
      agree and preserve all existing statuses.
- [ ] The sole complete planning task for increment 6 is added from actual
      consumed planning artifacts; no increment 6 implementation ID or later ID
      is reserved.

## Deferrals

None.

## Traces to

`docs/workplan/phase-1/README.md` increment 5; Phase 0 increment Plan Phase 1;
`docs/roadmap.md` Stage 0; `docs/contracts.md` §§8–9; ADR-0005/0006;
`skills/shared/workplan-model.md` planning shape and I1–I7.
