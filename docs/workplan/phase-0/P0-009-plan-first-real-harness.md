# P0-009: Plan Phase 2: first real harness

**Phase:** 0 — Planning and issue-resolution intake
**Shape:** planning
**Increment:** `plan-first-vendor`
**Dependencies:** None.

## Objective

A contributor can review a decided Phase 2 design for one real vendor behind the existing standalone attach surface, with the fake still supported.

## Scope

- Assign create-phase to a separate phase author for docs/workplan/phase-2/README.md. Inspect current source; identify launch/I/O/retention/input capability predicted from roadmap Stage 0 without claiming it exists or waiting for its implementation.
- Decide the first-vendor swap, capability/launch/codec responsibilities, supported login needs and native exit interpretation at independently observable depth. Revisit D1 at Stage 1 opening; revisit D5 only if the selected real harness requires terminal semantics, recording evidence rather than mandating a PTY in advance.
- Require the same standing interaction scenario against the real substrate and the shipped fake, maintained attach/operator paths, source-safe vendor fixtures, and honest limits on account-free CI versus opt-in real-vendor evidence.

## Non-goals

No vendor implementation, provider account operation or target briefs. Second-provider proof belongs to Stage 6; intentional run stop/retry to Stage 4; credential/shared-deployment hardening to Stage 5.

## Planning outputs and handoff

Produce `docs/workplan/phase-2/README.md` using
[`create-phase`](../../../skills/create-phase/SKILL.md), and update its
index entry and Phase 0 handoff evidence. The assigned phase author decides
implementation increments; this brief does not decide them in advance.
Have a different author review that artifact using
[`audit-phase`](../../../skills/audit-phase/SKILL.md), resolve findings
within the selected boundary, then hand its next uncovered increment to a
different `derive-task` author through Phase 0 intake. No implementation
brief or speculative tracker owner is allocated here.

## Write set

- `docs/workplan/phase-2/README.md` — new decided phase artifact.
- `docs/workplan/README.md` — target artifact link and planning graph notes.
- `docs/workplan/phase-0/README.md` — this increment's artifact-review/handoff evidence.
- `docs/workplan/tasks.json` — approval automation's status stamp only; the planning author leaves status fields unchanged.

## Contract surfaces

- Planning artifact namespace `phase-2`: its outcome, increment boundaries, prerequisite distinctions and exit meaning.
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
cat docs/workplan/phase-2/README.md
sed -n '/## Phases to plan/,/## Using the plan/p' docs/workplan/README.md
bash scripts/list-available-tasks.sh
```

An auditor can distinguish the planned Stage 0 substrate from current code, find a decided first vendor/trigger disposition, and trace real/fake parity and attach-visible native exit to concrete increments. The index resolves to the produced artifact; coverage identifies
the next separate-author handoff. Availability reflects actual tracker status
rather than the target product outcome.

## Verification

- Inspect current evidence: `cmd/fake-harness/main.go; cmd/winch-sandbox/main.go; deployments/Dockerfile; test/e2e/scenario_sandbox_starts_test.go`; distinguish source, observations and predictions.
- A separate `audit-phase` author reviews the new artifact against the traces below and finds no unresolved boundary, coverage or invariant defect.
- Apply the [shared planning checks](../README.md#planning-author-checks); report unavailable runtime gates without installing missing dependencies or claiming they passed.

## Acceptance criteria

- [ ] The design chooses one vendor at D1's trigger and records the D5 disposition with its condition; no unresolved candidate scope is left to the task author.
- [ ] Native exit result/interpretation is observable without borrowing the Stage 4 run lifecycle or Stage 5 escalation guarantee.
- [ ] A fake-compatible standing scenario and separate real-vendor evidence obligation are assigned; absent real execution is not described as parity.
- [ ] The target artifact follows the phase format, has decided ordered increments, coverage marked not yet derived, an observable exit and a named separate-author handoff; references resolve.
- [ ] The index/Phase 0 records agree with the artifact, runtime source is unchanged, and no retired ID or speculative implementation owner is introduced.

## Deferrals

None. Target product work is a non-goal of this artifact-producing task;
the roadmap owns its delivery stage, not an invented implementation deferral.

## Traces to

`docs/workplan/phase-0/README.md` — `plan-first-vendor`;
docs/roadmap.md §3 Stage 1, §4–6 D1/D5; docs/architecture.md §4 Harness adapters; docs/code-structure.md §3–4, §7; docs/contracts.md §8–9; ADR-0003; ADR-0005; `skills/shared/phase-design.md`;
`skills/shared/workplan-model.md` — planning shape and I1–I7.
