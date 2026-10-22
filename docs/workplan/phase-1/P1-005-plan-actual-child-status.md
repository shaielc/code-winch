# P1-005: Plan Phase 1 increment 3: actual child status

**Phase:** 1 — Sandbox exchange
**Shape:** planning
**Dependencies:** P1-003 (semantic: the assigned brief must consume the independently derived sandbox session `childStatus` contract and its `Down` presentation, which P1-003 produced in P1-004, and extend rather than redefine it)

## Objective

Produce one complete, independently authored implementation brief for Phase 1
increment 3, **Actual child status**, and reconcile it with the tracker,
workplan index, and Phase 0/1 coverage. The derived capability must make the
sandbox launch and own the shipped fake, and make the attach page report
`Running` while that child runs and `Exited` once it ends, from sandbox-owned
observation.

## Scope

Use `skills/derive-task/SKILL.md` against the reviewed Phase 1 design, the
completed P1-003 planning output (the P1-004 brief), and code at HEAD. Derive
exactly increment 3. Inspect the sandbox composition root, the shipped fake
binary, and the attach server and page before deciding the minimum spawn,
ownership, cleanup, and status-wiring path and the maintained hand path.

This task is pipelined: it derives increment 3 while P1-004 is still being
implemented. Plan against P1-004's brief as the consumed contract, label
behavior it predicts as a prediction, and record the derived brief's dependency
on P1-004's implementation as a graph edge rather than waiting for it here. Do
not combine increment 4–6 output display or transport, any input behavior, or
vendor-native exit interpretation.

Allocate the next Phase 1 implementation ID only when its complete brief and
matching tracker entry exist. Reconcile dependency reasons, coverage, graph
metrics, and collision analysis in the index and both phase records. Then
complete `skills/phase-increment-handoff/SKILL.md` in this same change.

## Required constraints

- Distinguish source-observed behavior from behavior predicted by P1-004's
  pending implementation and by the derived brief.
- Treat `Running` and `Exited` as sandbox-owned knowledge of a child the sandbox
  launched and owns, not a browser inference from output, request timing, or
  process-table scans.
- `Exited` reports only that the owned child ended. Exit codes, signals, or
  success/failure meaning are vendor-native interpretation (D5, Phase 2) unless
  the brief labels them as raw process observations.
- Require minimum cleanup: no orphaned child after sandbox shutdown, and
  observable behavior when spawn fails. Require negative tests for spawn failure
  and invalid status values, and prohibit false readiness or input-availability
  claims.
- Fake controls are chosen for this demonstration (for example, a way to make
  the fake exit) and may change the fake's current behavior; they do not define
  product configuration.
- Preserve loopback-only serving, posture disclosure, the local composer,
  deployment, the `Unknown` fallback and `Down` state, and existing tests. Any
  change to the `childStatus` wire contract is documented in the same change.

## Non-goals

This task writes planning artifacts only. Its target increment excludes reading
or presenting child output, input delivery, persistence, restart or retry
policy, provider semantics, and broad page restructuring. It does not derive a
second implementation brief or reserve implementation IDs for later increments.

## Planning outputs and handoff

Create `docs/workplan/phase-1/<allocated-ID>-actual-child-status.md` and its
matching tracker entry, then reconcile `docs/workplan/README.md`,
`docs/workplan/phase-0/README.md`, and `docs/workplan/phase-1/README.md`. Record
the extended `childStatus` values, the owning spawn/cleanup surface, and the
documentation/write forecast. Preserve every pre-existing status.

Handoff: complete `skills/phase-increment-handoff/SKILL.md` in this task; a
different refiner then takes the derived brief with
`skills/refine-task/SKILL.md`.

## Write set

- `docs/workplan/phase-1/<allocated-ID>-actual-child-status.md`
- `docs/workplan/phase-1/<next-ID>-plan-<next-increment>.md` (from the handoff)
- `docs/workplan/tasks.json` (new entries only; existing statuses unchanged)
- `docs/workplan/README.md`
- `docs/workplan/phase-0/README.md`
- `docs/workplan/phase-1/README.md`

## Contract surfaces

This planning task changes no product contract. Its derived brief will extend
P1-004's `GET /api/session` `childStatus` member with `Running` and `Exited` and
own the sandbox's child launch and ownership. That is a contract collision with
P1-004, which defines `childStatus`. The derived brief therefore takes a
`contract` edge on P1-004. It must not own output, input, readiness, or provider
interpretation.

## Demonstration

Review the derived brief and matching tracker/index/coverage records. Confirm
that its predicted compose-based hand check observes `Running` after the
sandbox starts the fake and `Exited` after the fake is made to end, preserves
`Down`, posture, and composer behavior, and claims no output or input
readiness. Confirm all records use one real ID and dependency set and that no
later implementation ID has been reserved. Confirm the next plan increment task
exists with its brief, tracker entry, and coverage.

## Verification

Run completion-script and control-panel tests, tracker JSON validation,
`scripts/list-available-tasks.sh`, an `rg` consistency search for both P1-005
and the allocated implementation ID, `git diff --check`, `make check`, and
`make test-cycle`. Report unavailable Docker/toolchains as limitations. The
derived brief must additionally name focused process-ownership, server-contract,
and frontend tests and the gates for every forecast implementation surface.

## Acceptance criteria

- [ ] Exactly one complete Phase 1 brief covers increment 3 with one real ID and
      a `contract` dependency on P1-004.
- [ ] The brief defines sandbox-owned `Running`/`Exited`, spawn-failure and
      cleanup behavior, fake controls, contract documentation, tests,
      deployment reachability, and evidence limits without absorbing
      increments 4–6.
- [ ] Tracker, index, Phase 0/1 coverage, graph metrics, and collision warnings
      agree and preserve all existing statuses.
- [ ] No speculative later implementation ID exists.
- [ ] Every `phase-increment-handoff` action holds: the next undelivered Phase 1
      increment has exactly one derivation task with a complete brief, an edge
      only to planning work it consumes, and matching tracker, index, and
      coverage records.

## Deferrals

None. Output, input, and vendor-native exit interpretation are excluded by the
decided Phase 1 and roadmap boundaries, not unfinished completion work in the
actual-child-status increment.

## Traces to

`docs/workplan/phase-1/README.md` increment 3; Phase 0 increment Plan Phase 1;
`docs/roadmap.md` Stage 0 and §6 D5; `docs/contracts.md` §§8–9; ADR-0005/0006.
