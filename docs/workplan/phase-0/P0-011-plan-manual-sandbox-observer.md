# P0-011: Plan Phase 4: observe manual sandboxes

**Phase:** 0 — Planning and issue-resolution intake
**Shape:** planning
**Increment:** `plan-observer`
**Dependencies:** None.

## Objective

A contributor can review a decided Phase 4 design for a durable product view of sandboxes started by hand, preserving direct attach.

## Scope

- Assign create-phase to a separate phase author for docs/workplan/phase-4/README.md. Distinguish the contributor control panel from the absent product winchd; treat prior session capabilities as predicted prerequisites.
- Decide registration/discovery, durable observation of existing executions and product listing/routing as independently observable increments, including the first reachable daemon/deployment/operator slice and required migration/parity evidence.
- Resolve D6 at Stage 3 opening; D4 when registration/first control-plane command defines the boundary; D7 when sandbox records are first persisted; D2 only when the product needs its first already-existing attach component. Decide the narrow runner/product/attach split without lifecycle ownership.

## Non-goals

No daemon, protocol, store or product UI implementation or target briefs. No product start/stop, leases, scheduling or multi-user access; these belong to roadmap Stages 4–5.

## Planning outputs and handoff

Produce `docs/workplan/phase-4/README.md` using
[`create-phase`](../../../skills/create-phase/SKILL.md), and update its
index entry and Phase 0 handoff evidence. The assigned phase author decides
implementation increments; this brief does not decide them in advance.
Have a different author review that artifact using
[`audit-phase`](../../../skills/audit-phase/SKILL.md), resolve findings
within the selected boundary, then hand its next uncovered increment to a
different `derive-task` author through Phase 0 intake. No implementation
brief or speculative tracker owner is allocated here.

## Write set

- `docs/workplan/phase-4/README.md` — new decided phase artifact.
- `docs/workplan/README.md` — target artifact link and planning graph notes.
- `docs/workplan/phase-0/README.md` — this increment's artifact-review/handoff evidence.
- `docs/workplan/tasks.json` — approval automation's status stamp only; the planning author leaves status fields unchanged.

## Contract surfaces

- Planning artifact namespace `phase-4`: its outcome, increment boundaries, prerequisite distinctions and exit meaning.
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
cat docs/workplan/phase-4/README.md
sed -n '/## Phases to plan/,/## Using the plan/p' docs/workplan/README.md
bash scripts/list-available-tasks.sh
```

An auditor can locate the increment that makes two manually started sandboxes appear durably in one product place and confirm attach remains usable when the daemon is unavailable. The index resolves to the produced artifact; coverage identifies
the next separate-author handoff. Availability reflects actual tracker status
rather than the target product outcome.

## Verification

- Inspect current evidence: `cmd/winch-sandbox/main.go; cmd/winch/main.go; internal/adapters/transport/attach/server.go; workplan-control-panel/README.md`; distinguish source, observations and predictions.
- A separate `audit-phase` author reviews the new artifact against the traces below and finds no unresolved boundary, coverage or invariant defect.
- Apply the [shared planning checks](../README.md#planning-author-checks); report unavailable runtime gates without installing missing dependencies or claiming they passed.

## Acceptance criteria

- [ ] The design observes existing execution rather than creating it; durable run observations introduce no start/stop, lease or scheduler capability.
- [ ] Each triggered D2/D4/D6/D7 decision has an explicit condition/disposition, and any trigger not reached remains deferred with its roadmap trigger.
- [ ] The first daemon/store increment includes reachable runtime/deployment/operator and same-scenario substrate evidence obligations; current absence is not hidden by the contributor panel.
- [ ] The target artifact follows the phase format, has decided ordered increments, coverage marked not yet derived, an observable exit and a named separate-author handoff; references resolve.
- [ ] The index/Phase 0 records agree with the artifact, runtime source is unchanged, and no retired ID or speculative implementation owner is introduced.

## Deferrals

None. Target product work is a non-goal of this artifact-producing task;
the roadmap owns its delivery stage, not an invented implementation deferral.

## Traces to

`docs/workplan/phase-0/README.md` — `plan-observer`;
docs/roadmap.md §3 Stage 3, §5–6 D2/D4/D6/D7; docs/architecture.md §5 Topology B, §6; docs/contracts.md §2, §4–5, §8; ADR-0005; ADR-0006; `skills/shared/phase-design.md`;
`skills/shared/workplan-model.md` — planning shape and I1–I7.
