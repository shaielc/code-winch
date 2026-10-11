# P0-008: Plan Phase 1: standalone fake sandbox

**Phase:** 0 — Planning and issue-resolution intake
**Shape:** planning
**Increment:** `plan-sandbox`
**Dependencies:** None.

## Objective

A contributor can review a decided Phase 1 design for finishing the standalone fake sandbox loop from the existing page and local composer.

## Scope

- Assign a separate phase author to create-phase for docs/workplan/phase-1/README.md. Reuse the mounted composer, posture/status entrypoints and deploy sources; inspect the root and gates rather than treating retired completion as proof.
- Decide independently observable output and input increments: raw live output with child ownership/minimum cleanup; local UI repairs where exact evidence requires them; retention/reload; recorded server input; delivery/reply; bounded output and reader isolation at their own depth. Decide shared record/input/diagnostic responsibility without forcing semantic decoding or Stage 2 resume into Stage 0.
- Place inherited run/deploy/gate repairs where required for the first runnable increment, and require each new operator capability's minimal maintained path in the increment that reaches it. Preserve the shipped fake as malleable and require fake/end-to-end evidence obligations.

## Non-goals

No harness/store/input implementation or implementation briefs; no retired brief reactivation. Runs, vendor semantics, native exit taxonomy, stop/retry, multi-view resume and shared-deployment enforcement remain roadmap Stages 1–5.

## Planning outputs and handoff

Produce `docs/workplan/phase-1/README.md` using
[`create-phase`](../../../skills/create-phase/SKILL.md), and update its
index entry and Phase 0 handoff evidence. The assigned phase author decides
implementation increments; this brief does not decide them in advance.
Have a different author review that artifact using
[`audit-phase`](../../../skills/audit-phase/SKILL.md), resolve findings
within the selected boundary, then hand its next uncovered increment to a
different `derive-task` author through Phase 0 intake. No implementation
brief or speculative tracker owner is allocated here.

## Write set

- `docs/workplan/phase-1/README.md` — new decided phase artifact.
- `docs/workplan/README.md` — target artifact link and planning graph notes.
- `docs/workplan/phase-0/README.md` — this increment's artifact-review/handoff evidence.
- `docs/workplan/tasks.json` — approval automation's status stamp only; the planning author leaves status fields unchanged.

## Contract surfaces

- Planning artifact namespace `phase-1`: its outcome, increment boundaries, prerequisite distinctions and exit meaning.
- No product API, port, event, registry, configuration or migration meaning is defined by this task. Shared index/tracker/Phase 0 coverage rows carry independent references, not a shared target-phase meaning.

## Dependency and concurrency analysis

No compile, contract or semantic output from another planning task is needed to
design this phase. Earlier product capabilities are predictions to identify and
revalidate, not planning completion edges. Other design tasks own distinct
target-phase namespaces. Shared index/coverage writes and automatic tracker
stamps are merge warnings; the [index](../README.md#phase-0-graph-and-handoff)
reports the current graph, available pairs and width. P0-015, P0-016 and P0-017
consume this task's decided Phase 1 namespace and take contract edges; none waits
for Phase 1 implementation.

The current graph has 10 tasks, critical-path length 2 and average width 5.0;
7 tasks are available now (maximum possible frontier 9). Current available
write collisions are the 21 unordered pairs among P0-008–P0-014; their
shared index/coverage/status writes are warnings, with no available contract
collision. The index explains the distinct target artifact namespaces.

## Demonstration

Run from the repository root after executing this planning task:

```sh
cat docs/workplan/phase-1/README.md
sed -n '/## Phases to plan/,/## Using the plan/p' docs/workplan/README.md
bash scripts/list-available-tasks.sh
```

An auditor follows the two independent paths from current local UI/raw-output gaps to page-to-fake-to-page exchange surviving reload, and can identify which increment owns each included guarantee. The index resolves to the produced artifact; coverage identifies
the next separate-author handoff. Availability reflects actual tracker status
rather than the target product outcome.

## Verification

- Inspect current evidence: `cmd/winch-sandbox/main.go; internal/adapters/transport/attach/server.go; web/src/attach/App.tsx; web/src/attach/Composer.tsx; Makefile; deployments/compose.yml`; distinguish source, observations and predictions.
- A separate `audit-phase` author reviews the new artifact against the traces below and finds no unresolved boundary, coverage or invariant defect.
- Apply the [shared planning checks](../README.md#planning-author-checks); report unavailable runtime gates without installing missing dependencies or claiming they passed.

## Acceptance criteria

- [ ] The baseline distinguishes source-supported page/composer/status from unobserved startup and absent runner/store/server input; run/deploy repairs are explicit rather than assumed passed.
- [ ] Input and output depth is decided, every concern has an increment or roadmap destination, and no seam postpones its basic operability to a separate task.
- [ ] The exit proves the Stage 0 exchange/reload outcome without requiring vendor decoding, terminal run states, Stage 2 resume or shared-deployment certification.
- [ ] The target artifact follows the phase format, has decided ordered increments, coverage marked not yet derived, an observable exit and a named separate-author handoff; references resolve.
- [ ] The index/Phase 0 records agree with the artifact, runtime source is unchanged, and no retired ID or speculative implementation owner is introduced.

## Deferrals

None. Target product work is a non-goal of this artifact-producing task;
the roadmap owns its delivery stage, not an invented implementation deferral.

## Traces to

`docs/workplan/phase-0/README.md` — `plan-sandbox`;
docs/roadmap.md §3 Stage 0, §4–6; docs/architecture.md §4–6; docs/contracts.md §8–9; docs/state.md; ADR-0005; ADR-0006; `skills/shared/phase-design.md`;
`skills/shared/workplan-model.md` — planning shape and I1–I7.
