# P0-014: Plan Phase 7: second provider and workflows

**Phase:** 0 — Planning and issue-resolution intake
**Shape:** planning
**Increment:** `plan-workflows`
**Dependencies:** None.

## Objective

A contributor can review a decided Phase 7 design separating second-provider seam proof from durable workflows spanning ordinary runs.

## Scope

- Assign create-phase to a separate phase author for docs/workplan/phase-7/README.md. Mark one-provider, lifecycle and security capabilities as predicted prerequisites and inspect current source before calling any of them reusable code.
- Decide second-provider proof first, including descriptor/launch/codec/native exit and the same real/fake standing scenario without provider-specific generic-core branches.
- Partition version-pinned workflow definitions, durable instances/steps, normal command/event coordination, deterministic idempotency, waits/deadlines, bounded branching and retry/compensation at independently observable depth. Require a two-run demonstration and reachable product/operator path for each workflow capability; do not build extension seams solely for hypothetical future engines.

## Non-goals

No provider or workflow implementation, provider account operation or target briefs. No process control bypass from workflows, arbitrary embedded workflow code, unbounded branching or beyond-roadmap scaling topology.

## Planning outputs and handoff

Produce `docs/workplan/phase-7/README.md` using
[`create-phase`](../../../skills/create-phase/SKILL.md), and update its
index entry and Phase 0 handoff evidence. The assigned phase author decides
implementation increments; this brief does not decide them in advance.
Have a different author review that artifact using
[`audit-phase`](../../../skills/audit-phase/SKILL.md), resolve findings
within the selected boundary, then hand its next uncovered increment to a
different `derive-task` author through Phase 0 intake. No implementation
brief or speculative tracker owner is allocated here.

## Write set

- `docs/workplan/phase-7/README.md` — new decided phase artifact.
- `docs/workplan/README.md` — target artifact link and planning graph notes.
- `docs/workplan/phase-0/README.md` — this increment's artifact-review/handoff evidence.
- `docs/workplan/tasks.json` — approval automation's status stamp only; the planning author leaves status fields unchanged.

## Contract surfaces

- Planning artifact namespace `phase-7`: its outcome, increment boundaries, prerequisite distinctions and exit meaning.
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
cat docs/workplan/phase-7/README.md
sed -n '/## Phases to plan/,/## Using the plan/p' docs/workplan/README.md
bash scripts/list-available-tasks.sh
```

An auditor can separate the two-provider same-scenario proof from the two-run workflow demonstration, and trace crash replay/deduplication to durable ordinary-command orchestration. The index resolves to the produced artifact; coverage identifies
the next separate-author handoff. Availability reflects actual tracker status
rather than the target product outcome.

## Verification

- Inspect current evidence: `cmd/fake-harness/main.go; cmd/winch-sandbox/main.go; cmd/winch/main.go; docs/state.md`; distinguish source, observations and predictions.
- A separate `audit-phase` author reviews the new artifact against the traces below and finds no unresolved boundary, coverage or invariant defect.
- Apply the [shared planning checks](../README.md#planning-author-checks); report unavailable runtime gates without installing missing dependencies or claiming they passed.

## Acceptance criteria

- [ ] The second provider tests the existing seam with parity evidence before workflow outcomes are claimed; generic code contains no planned vendor branches.
- [ ] Workflows issue ordinary application commands, pin definition/profile versions and give durable replay/idempotency/retry/wait responsibilities decided increment coverage.
- [ ] The exit includes two real adapters plus a workflow spanning two runs, retains fake support and names no scaling work beyond the roadmap.
- [ ] The target artifact follows the phase format, has decided ordered increments, coverage marked not yet derived, an observable exit and a named separate-author handoff; references resolve.
- [ ] The index/Phase 0 records agree with the artifact, runtime source is unchanged, and no retired ID or speculative implementation owner is introduced.

## Deferrals

None. Target product work is a non-goal of this artifact-producing task;
the roadmap owns its delivery stage, not an invented implementation deferral.

## Traces to

`docs/workplan/phase-0/README.md` — `plan-workflows`;
docs/roadmap.md §3 Stage 6; docs/contracts.md §7, §9; docs/architecture.md §4 Harness adapters and Workflow coordinator, §5 Topology D; docs/code-structure.md §3–4, §7; ADR-0003; `skills/shared/phase-design.md`;
`skills/shared/workplan-model.md` — planning shape and I1–I7.
