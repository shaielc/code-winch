# P0-010: Plan Phase 3: readable and rejoinable session

**Phase:** 0 — Planning and issue-resolution intake
**Shape:** planning
**Increment:** `plan-readable-session`
**Dependencies:** None.

## Objective

A contributor can review a decided Phase 3 design for structured session records, two projections and gap-free resume by runner-local ordinal.

## Scope

- Assign create-phase to a separate phase author for docs/workplan/phase-3/README.md. Identify raw/session retention and one-vendor capabilities as predicted Stage 0–1 inputs, separately from current local-only page state.
- Decide structured record production and terminal/conversation projections over a shared session-compatible record shape, with incremental/bounded projection obligations and safe fallback for unknown records.
- Partition ordinal resume, snapshot/live handover, caught-up/heartbeat/gap repair and slow-reader guarantees at observable depth; require reconnect and view-switch evidence through the maintained surface without canonical run sequencing.

## Non-goals

No renderer, parser or reconnect implementation or target briefs. Cross-session history/search/export/retention policy is excluded by Stage 2; canonical per-run sequence belongs to the control-plane stages.

## Planning outputs and handoff

Produce `docs/workplan/phase-3/README.md` using
[`create-phase`](../../../skills/create-phase/SKILL.md), and update its
index entry and Phase 0 handoff evidence. The assigned phase author decides
implementation increments; this brief does not decide them in advance.
Have a different author review that artifact using
[`audit-phase`](../../../skills/audit-phase/SKILL.md), resolve findings
within the selected boundary, then hand its next uncovered increment to a
different `derive-task` author through Phase 0 intake. No implementation
brief or speculative tracker owner is allocated here.

## Write set

- `docs/workplan/phase-3/README.md` — new decided phase artifact.
- `docs/workplan/README.md` — target artifact link and planning graph notes.
- `docs/workplan/phase-0/README.md` — this increment's artifact-review/handoff evidence.
- `docs/workplan/tasks.json` — approval automation's status stamp only; the planning author leaves status fields unchanged.

## Contract surfaces

- Planning artifact namespace `phase-3`: its outcome, increment boundaries, prerequisite distinctions and exit meaning.
- No product API, port, event, registry, configuration or migration meaning is defined by this task. Shared index/tracker/Phase 0 coverage rows carry independent references, not a shared target-phase meaning.

## Dependency and concurrency analysis

No compile, contract or semantic output from another planning task is needed to
design this phase. Earlier product capabilities are predictions to identify and
revalidate, not planning completion edges. Other design tasks own distinct
target-phase namespaces. Shared index/coverage writes and automatic tracker
stamps are merge warnings; the [index](../README.md#phase-0-graph-and-handoff)
reports the current graph, available pairs and width. There is no available target-artifact contract collision.

The current graph has 10 tasks, critical-path length 2 and average width 5.0;
7 tasks are available now (maximum possible frontier 9). Current available
write collisions are the 21 unordered pairs among P0-008–P0-014; their
shared index/coverage/status writes are warnings, with no available contract
collision. The index explains the distinct target artifact namespaces.

## Demonstration

Run from the repository root after executing this planning task:

```sh
cat docs/workplan/phase-3/README.md
sed -n '/## Phases to plan/,/## Using the plan/p' docs/workplan/README.md
bash scripts/list-available-tasks.sh
```

An auditor traces the reconnect-with-no-loss and terminal/conversation switch observations to decided increments, while the record shape remains usable without a control plane. The index resolves to the produced artifact; coverage identifies
the next separate-author handoff. Availability reflects actual tracker status
rather than the target product outcome.

## Verification

- Inspect current evidence: `web/src/attach/App.tsx; internal/adapters/transport/attach/server.go; test/e2e/scenario_sandbox_starts_test.go`; distinguish source, observations and predictions.
- A separate `audit-phase` author reviews the new artifact against the traces below and finds no unresolved boundary, coverage or invariant defect.
- Apply the [shared planning checks](../README.md#planning-author-checks); report unavailable runtime gates without installing missing dependencies or claiming they passed.

## Acceptance criteria

- [ ] The two views project the same retained session record meaning rather than maintaining independent histories or taking a control-plane-only input.
- [ ] Resume guarantees explicitly cover the snapshot/live boundary and ordinal gaps, with failures/slow readers observable at the increment that claims them.
- [ ] Planned prerequisites and current gaps are labelled; no cross-session retention policy or canonical sequence is smuggled into the phase.
- [ ] The target artifact follows the phase format, has decided ordered increments, coverage marked not yet derived, an observable exit and a named separate-author handoff; references resolve.
- [ ] The index/Phase 0 records agree with the artifact, runtime source is unchanged, and no retired ID or speculative implementation owner is introduced.

## Deferrals

None. Target product work is a non-goal of this artifact-producing task;
the roadmap owns its delivery stage, not an invented implementation deferral.

## Traces to

`docs/workplan/phase-0/README.md` — `plan-readable-session`;
docs/roadmap.md §3 Stage 2, §5–6; docs/contracts.md §6, §8; docs/architecture.md §4 Event pipeline and renderers; ADR-0002; ADR-0006; `skills/shared/phase-design.md`;
`skills/shared/workplan-model.md` — planning shape and I1–I7.
