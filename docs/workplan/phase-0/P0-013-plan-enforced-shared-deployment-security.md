# P0-013: Plan Phase 6: enforced security

**Phase:** 0 — Planning and issue-resolution intake
**Shape:** planning
**Increment:** `plan-security`
**Dependencies:** None.

## Objective

A contributor can review a decided Phase 6 design mapping shared-deployment security blockers to observable enforcement and attack evidence.

## Scope

- Assign create-phase to a separate phase author for docs/workplan/phase-6/README.md. Preserve current posture as source evidence only; do not infer enforcement from profile labels, non-root image source or loopback compose declarations.
- Partition security §11 LB01–LB11 without losing credentials/scoped injection and pre-persistence redaction, egress/DNS/metadata controls, approved workspace/traversal/extraction, resources/stop escalation/orphan cleanup, identity/authorization/stream controls and malicious renderer/parser bounds.
- Give data lifecycle, provenance/SBOM/scans, incident/revocation/recovery and off-host attach reachability their explicit blocker coverage and evidence responsibilities. Distinguish fake negative evidence from real deployment-network/substrate/off-host proof; preserve honest UI/operator posture until each claim is proven.

## Non-goals

No enforcement, attack execution against external systems or blocker sign-off in this planning task; no target briefs. Hostile public tenancy requires a stronger isolation design and remains outside the container roadmap; workflows and second vendor belong to Stage 6.

## Planning outputs and handoff

Produce `docs/workplan/phase-6/README.md` using
[`create-phase`](../../../skills/create-phase/SKILL.md), and update its
index entry and Phase 0 handoff evidence. The assigned phase author decides
implementation increments; this brief does not decide them in advance.
Have a different author review that artifact using
[`audit-phase`](../../../skills/audit-phase/SKILL.md), resolve findings
within the selected boundary, then hand its next uncovered increment to a
different `derive-task` author through Phase 0 intake. No implementation
brief or speculative tracker owner is allocated here.

## Write set

- `docs/workplan/phase-6/README.md` — new decided phase artifact.
- `docs/workplan/README.md` — target artifact link and planning graph notes.
- `docs/workplan/phase-0/README.md` — this increment's artifact-review/handoff evidence.
- `docs/workplan/tasks.json` — approval automation's status stamp only; the planning author leaves status fields unchanged.

## Contract surfaces

- Planning artifact namespace `phase-6`: its outcome, increment boundaries, prerequisite distinctions and exit meaning.
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
cat docs/workplan/phase-6/README.md
sed -n '/## Phases to plan/,/## Using the plan/p' docs/workplan/README.md
bash scripts/list-available-tasks.sh
```

An auditor maps every LB01–LB11 item to an increment with an attempted violation and visible refusal/recovery evidence, and finds no shared-deployment claim made solely by configuration or planning. The index resolves to the produced artifact; coverage identifies
the next separate-author handoff. Availability reflects actual tracker status
rather than the target product outcome.

## Verification

- Inspect current evidence: `deployments/Dockerfile; deployments/compose.yml; internal/adapters/transport/attach/session.go; internal/adapters/transport/attach/static.go; web/src/attach/Posture.tsx`; distinguish source, observations and predictions.
- A separate `audit-phase` author reviews the new artifact against the traces below and finds no unresolved boundary, coverage or invariant defect.
- Apply the [shared planning checks](../README.md#planning-author-checks); report unavailable runtime gates without installing missing dependencies or claiming they passed.

## Acceptance criteria

- [ ] All eleven blockers have specific coverage and accountable evidence obligations; excluded hostile tenancy and remaining verification limits are explicit.
- [ ] Attack observations include egress/DNS/metadata, traversal/extraction, credential canaries, forced stop/orphans and off-host attach reachability, with the real substrate required where a fake cannot establish isolation.
- [ ] The phase exit requires reviewed blocker evidence/sign-off and retains honest unenforced posture before that evidence; creating the artifact closes no blocker.
- [ ] The target artifact follows the phase format, has decided ordered increments, coverage marked not yet derived, an observable exit and a named separate-author handoff; references resolve.
- [ ] The index/Phase 0 records agree with the artifact, runtime source is unchanged, and no retired ID or speculative implementation owner is introduced.

## Deferrals

None. Target product work is a non-goal of this artifact-producing task;
the roadmap owns its delivery stage, not an invented implementation deferral.

## Traces to

`docs/workplan/phase-0/README.md` — `plan-security`;
docs/roadmap.md §3 Stage 5, §5; docs/security.md §2–12, especially §11 LB01–LB11; docs/contracts.md §2, §4–6, §8; docs/state.md; ADR-0003; ADR-0006; `skills/shared/phase-design.md`;
`skills/shared/workplan-model.md` — planning shape and I1–I7.
