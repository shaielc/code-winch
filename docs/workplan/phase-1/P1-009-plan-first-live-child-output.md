# P1-009: Plan Phase 1 increment 5: first live child output

**Phase:** 1 — Sandbox exchange
**Shape:** planning
**Dependencies:** P1-005 (semantic: consume the P1-006 brief defining owned
fake-child lifecycle), P1-007 (semantic: consume the P1-008 brief defining the
distinct harness-output region and truthful disconnected state)

## Objective

Produce one complete, independently authored implementation brief for Phase 1
increment 5, **First live child output**, and reconcile it with the tracker,
workplan index, and Phase 0/1 coverage. The derived capability must make raw
bytes emitted by the owned fake appear in the existing Harness output region.

## Scope

Use `skills/derive-task/SKILL.md` against the reviewed Phase 1 design, the
P1-006 and P1-008 briefs, and code at HEAD. Derive exactly increment 5. Recheck
whether those predicted capabilities have been implemented and distinguish
source observations from planned behavior.

Inspect fake launch and stdout/stderr behavior, process ownership and cleanup,
the attach transport, the output-area presentation, component and contract
tests, the standing scenario, web build, and deployed observation path. Select
the smallest child-byte read, live transport, and browser presentation loop
that lets a person observe fake-produced raw output.

The derived implementation depends on the actual-child-status implementation
and harness-output-area implementation whose contracts it consumes, even if
either is still pending. Exclude increment 6's proof that pieces arrive
incrementally and later history, input, interpretation, and hardening work.

## Required constraints

- Preserve truthful child-status meanings and ownership/cleanup, and replace
  P1-008's disconnected empty state only when the view is actually connected
  to the chosen minimum live transport.
- Raw presentation must not claim semantic messages, replies, terminal
  interpretation, retained history, readiness, or provider compatibility.
- Define minimum bounded child-reading and browser/publication behavior needed
  to keep the sandbox runnable when readers disconnect; do not absorb the
  advanced size, resynchronization, or multi-reader guarantees of increments
  26–29.
- Preserve posture disclosure, browser-local composer behavior, loopback-only
  serving, deployment, and existing tests. Declare and document every HTTP or
  stream contract the implementation adds.
- Require an observation of the rendered deployed page. Static HTML inspection
  cannot prove that live bytes reached React.

## Non-goals

This task writes planning artifacts only. Its target increment excludes proof
of incremental arrival, durable replay/history, input or reply correlation,
semantic/provider decoding, terminal emulation, output ordinals, persistence,
full malformed/oversized-output hardening, slow-reader guarantees, automatic
resume, and broad page restructuring. Derive one implementation brief only and
reserve no later implementation ID.

## Planning outputs and handoff

Create `docs/workplan/phase-1/<allocated-ID>-first-live-child-output.md` and its
matching tracker entry, then reconcile `docs/workplan/README.md`,
`docs/workplan/phase-0/README.md`, and `docs/workplan/phase-1/README.md`.
Record the child-byte boundary, minimum live transport, raw presentation,
cleanup/bounds, tests, hand path, and evidence limits.

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

This planning task changes no product contract. Its derived brief will extend
the owned fake-child lifecycle and the accessible Harness output presentation
with a declared minimum live raw-output transport. It must take contract edges
on the implementation briefs produced by P1-005 and P1-007 and must document
any new HTTP/stream surface. It owns no durable record, input, provider
interpretation, or terminal contract.

## Demonstration

Review the derived brief and matching tracker/index/coverage records. Confirm
its predicted branch-isolated hand check uses `make test-env`, the proxy URL in
`deployments/README.md`, and `make test-env-down`; causes deterministic raw
output from the owned fake; observes it in the rendered Harness output region;
preserves child status, posture and the local composer; and makes no claim of
incremental delivery, retention, interpretation, or reply correlation.

## Verification

Run completion-script and control-panel tests, tracker JSON validation,
`scripts/list-available-tasks.sh`, an `rg` consistency search for this planning
task and both newly allocated IDs, `git diff --check`, `make check`, and
`make test-cycle`. Report unavailable Docker/toolchains as limitations.

The derived brief must name focused child-read, transport disconnect/cleanup,
transport contract, and React raw-presentation tests; preservation checks for
child status, posture, and composer; all affected web and Go gates; the standing
scenario; and a deployed rendered-page observation with explicit evidence
limits.

## Acceptance criteria

- [ ] Exactly one complete Phase 1 implementation brief covers increment 5,
      with contract dependencies on the actual-child-status and output-area
      implementations derived by the named planning prerequisites.
- [ ] The brief defines the smallest observable raw child-byte path, required
      cleanup and minimum bounds, declared transport contract, browser tests,
      deployed reachability, and evidence limits without absorbing increment 6
      or later capabilities.
- [ ] Tracker, index, Phase 0/1 coverage, graph metrics, and collision warnings
      agree and preserve all existing statuses.
- [ ] The sole complete planning task for increment 6 is added with dependencies
      derived from consumed planning artifacts; no increment 6 implementation
      ID or later ID is reserved.
- [ ] A different refiner receives the finished increment 5 brief, and artifact
      review demonstrates every output of this planning task.

## Deferrals

None. Incremental arrival, retained history, input, interpretation, and output
hardening are exclusions owned by decided later phase increments, not unfinished
completion work for the first live-output slice.

## Traces to

`docs/workplan/phase-1/README.md` increment 5; P1-006; P1-008;
`docs/roadmap.md` Stage 0; `docs/contracts.md` §§8–9; ADR-0005/0006;
`skills/shared/workplan-model.md` planning shape and I1–I7.
