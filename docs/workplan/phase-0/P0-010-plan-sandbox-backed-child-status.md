# P0-010: Plan Phase 1 increment 2: sandbox-backed child status

**Phase:** 0 — Plan the implementation phases
**Shape:** planning
**Dependencies:** P0-008 (semantic: the assigned brief must consume the independently derived child-status presentation boundary and preserve its label/value contract while replacing the browser constant)

## Objective

Produce one complete, independently authored implementation brief for Phase 1
increment 2, **Sandbox-backed child status**, and reconcile it with the tracker,
workplan index, and Phase 0/1 coverage. The derived capability must replace the
browser-only constant with sandbox-owned observation that truthfully reports
`Down` when the sandbox has no child, without starting or supervising one.

## Scope

Use `skills/derive-task/SKILL.md` against the reviewed Phase 1 design, the
completed P0-008 planning output, and code at HEAD. Derive exactly increment 2.
Inspect the existing attach server and page before deciding the minimum status
transport and maintained hand path. Preserve increment 1's accessible child
status identity while moving authority to the sandbox. Do not combine increment
3's child launch, ownership, cleanup, or `Running`/`Exited` behavior.

Allocate the next plan-wide implementation ID only when its complete brief and
matching tracker entry exist. Reconcile dependency reasons, coverage, graph
metrics, and collision analysis in the index and both phase records. Arrange
handoff to a different refiner using `skills/refine-task/SKILL.md`.

## Required constraints

- Distinguish source-observed behavior from predicted implementation results.
- Treat `Down` as sandbox-owned knowledge that no child is owned, not a browser
  inference from request failure or absence of output.
- Require negative tests for request failure/invalid status and prohibit false
  `Running`, readiness, fake availability, or exit-result claims.
- Preserve loopback-only serving, posture disclosure, local composer, deployment,
  and existing tests. Any API surface must be documented in the same change.
- State what a fake-free `Down` demonstration proves and does not prove.

## Non-goals

This task writes planning artifacts only. Its target increment excludes child
startup/supervision, output transport, input delivery, persistence, provider
semantics, and broad page restructuring. It does not derive a second
implementation brief or reserve IDs for later increments.

## Planning outputs and handoff

Create `docs/workplan/phase-1/<allocated-ID>-sandbox-backed-child-status.md` and
its matching tracker entry, then reconcile `docs/workplan/README.md`,
`docs/workplan/phase-0/README.md`, and `docs/workplan/phase-1/README.md`. Record
the concrete transport contract and its documentation/write forecast rather
than assuming increment 1 created a backend seam. Preserve every pre-existing
status. Hand the finished brief to a different refiner; any subsequent Phase 1
planning handoff requires its own separately authored tracked planning task.

## Write set

- `docs/workplan/phase-1/<allocated-ID>-sandbox-backed-child-status.md`
- `docs/workplan/tasks.json` (one new entry; existing statuses unchanged)
- `docs/workplan/README.md`
- `docs/workplan/phase-0/README.md`
- `docs/workplan/phase-1/README.md`

## Contract surfaces

This planning task changes no product contract. Its derived brief will own the
minimum sandbox-to-browser child-status transport and the `Down` presentation,
while consuming increment 1's accessible status identity. It must not own child
lifecycle, readiness, output, input, or provider interpretation. Reconcile any
shared product meaning as a contract collision before assignment.

## Demonstration

Review the derived brief and matching tracker/index/coverage records. Confirm
that its predicted compose-based hand check observes `Down` from a sandbox with
no child, preserves posture and composer behavior, and makes no claim that a
child was started or that input is ready. Confirm all records use one real ID
and dependency set and no later implementation ID has been reserved.

## Verification

Run completion-script and control-panel tests, tracker JSON validation,
`scripts/list-available-tasks.sh`, an `rg` consistency search for both P0-010
and the allocated implementation ID, `git diff --check`, `make check`, and
`make test-cycle`. Report unavailable Docker/toolchains as limitations. The
derived brief must additionally name focused frontend/server contract tests and
the gates for every forecast implementation surface.

## Acceptance criteria

- [ ] Exactly one complete Phase 1 brief covers increment 2 with one real ID and
      an explicit dependency on the behavior it actually consumes.
- [ ] The brief defines truthful sandbox-owned `Down`, failure behavior,
      transport documentation, tests, deployment reachability, and evidence
      limits without absorbing increment 3.
- [ ] Tracker, index, Phase 0/1 coverage, graph metrics, and collision warnings
      agree and preserve all existing statuses.
- [ ] A different refiner receives the finished implementation brief; no
      speculative later implementation ID exists.

## Deferrals

None. Child launch and ownership are excluded by the decided Phase 1 boundary,
not unfinished completion work in the sandbox-backed observation increment.

## Traces to

`docs/workplan/phase-1/README.md` increment 2; Phase 0 increment 0.1;
`docs/roadmap.md` Stage 0; `docs/contracts.md` §§8–9; ADR-0005/0006.
