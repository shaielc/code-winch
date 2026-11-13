# P0-002: Plan phase-2: first real vendor

**Phase:** 0 — Plan the implementation phases
**Shape:** planning
**Dependencies:** None

## Objective

Produce a reviewed `docs/workplan/phase-2/README.md` that selects the first real vendor and decides the ordered, independently observable increments for roadmap Stage 1, then create the single Phase 2 planning-shape task that derives its first increment through `skills/phase-increment-handoff/SKILL.md`. The artifact must make the real CLI usable behind the same standalone-sandbox attach surface while preserving the fake profile, and must give native exit reporting its own observable result rather than treating process exit as proof of a run lifecycle.

## Scope

- Start the phase artifact from the checkout that exists at dispatch. Cite the current composition root and attach transport (`cmd/winch-sandbox/main.go`, `internal/adapters/transport/attach/`), shipped fake (`cmd/fake-harness/`), attach page (`web/src/attach/`), deployment profiles, and standing scenario. Label the page/status/standalone fake as source-supported; label harness launch, retained exchange, adapter packages, descriptor/codec, vendor login, and exit interpretation as missing unless HEAD then proves otherwise. Treat completed Phase 1 briefs as source claims to verify and pending Phase 1 briefs as planned prerequisites, not as implemented Stage 0 behavior.
- Resolve roadmap decision D1 in the Phase 2 README, not in a later task. Record a dated, sanitized evidence table for the candidate CLIs covering supported non-interactive launch, input and output modes, authentication/login, exit behavior, licensing/redistribution implications, and automation suitability; name one vendor and explain the rejection of the alternatives. Evidence may be an installed CLI's content-free help/version output, an official vendor source, or a sanitized recorded transcript. Do not require a live provider account or place credentials, prompt/output content, or machine-specific paths in the repository.
- Resolve D5 only from evidence that the selected CLI requires resize, raw mode, or TTY detection. If that trigger is absent, record that pipes remain the planned Stage 1 transport and leave D5 deferred with its roadmap trigger; preference, colored output, or an interactive default alone is not evidence that a PTY is required.
- Partition Stage 1 from the shipped attach surface into the smallest useful ordered observations. The design must separately account for (1) selecting/configuring and launching the chosen CLI through an explicit descriptor/launch boundary with the login prerequisite documented, (2) preserving Stage 0 input/output behavior through the real adapter and fake profiles at the same page/API/operator surface, and (3) reporting the native exit result and the chosen adapter's documented interpretation through that surface. If the phase combines any two, it must explain why the earlier observation cannot be useful or truthful on its own.
- Define the adapter seam at the responsibility level fixed by `docs/code-structure.md` §3–4 and ADR-0003: explicit composition-root registration, stable descriptor/capabilities, launch construction, chunk-boundary-tolerant input/output translation, and native-result interpretation. Leave Go signatures, package names, configuration key names, endpoint shapes, and fixture dialect to the implementation brief unless the existing attach contract or a separately routed design correction fixes them.
- Retain the fake as a shipped, selectable configuration and require the same standing scenario at the fake and selected-real-adapter depths where a real substrate is introduced. State what sanitized/opt-in real-vendor evidence proves and does not prove. Plan only the login/setup needed for a single-user local deployment; credential references, scoped injection, shared-deployment authorization, and enforced secret handling remain Stage 5 work.

## Required constraints

At refinement HEAD, `git ls-files internal cmd test/e2e deployments web/src/attach` shows no harness-adapter package, descriptor, codec, native-exit interpretation, or PTY implementation. `cmd/winch-sandbox/main.go` registers only the attach server; `cmd/fake-harness/main.go` is a standalone fixture; and the only standing scenario checks sandbox startup. Stage 0 exchange remains a planned Phase 1 prerequisite and must not be marked implemented merely because its design or brief exists. Recheck these observations when P0-002 is implemented because Phase 1 work is in flight.

P0-002 has no dependency because phase design can identify planned prerequisites without waiting for their implementation. Existing architecture, roadmap, and attach contract specify the responsibility boundary sufficiently to design the artifact. Do not add an ordering-only edge to Phase 1 planning or implementation; instead, name each Stage 0 behavior the Phase 2 design consumes, its current source or existing task owner, and whether it is present or planned.

Use `skills/create-phase/SKILL.md` as the phase-author operation, with `skills/shared/phase-design.md` and the workplan model. The required behavior is a decided phase artifact, an independent `audit-phase` review of that artifact, the single planning-shape task from `skills/phase-increment-handoff/SKILL.md`, and consistent tracker/index/Phase 0 and phase coverage. Product behavior and commands described there are predictions. Preserve Phase 0's decided boundaries, the frozen tracker schema, and every existing task status.

## Non-goals

Second vendor and workflows (Stage 6); run creation, intentional start/stop, retries, attempts, and lifecycle state (Stage 4); stop escalation, orphan recovery, shared-deployment credential enforcement, and egress/isolation guarantees (Stage 5); Stage 2 structured projections and resumable views.

This task derives no implementation brief and allocates no implementation ID; the planning task it creates assigns that derivation. Product implementation and completion claims based only on a design remain excluded.

## Planning outputs and handoff

The phase author creates `docs/workplan/phase-2/README.md` using the phase format, records the source-supported baseline separately from planned Phase 1 prerequisites, closes D1 with the evidence above, records the D5 disposition, chooses ordered increments and one concern owner for each included responsibility, states observable exit and evidence limits, and marks implementation-task coverage **not yet derived**. Update Phase 2's index row and Phase 0 coverage. Inspect current source and the active Phase 1 briefs before partitioning; do not copy the closure-era baseline forward. No product protocol or substrate implementation is prescribed by this brief.

A separate phase auditor uses `skills/audit-phase/SKILL.md` against the decided design and the commit containing it. Record the audit evidence in the task report and resolve blocking artifact findings before handoff is considered accepted; the audit does not edit the artifact or task statuses.

Handoff: after the design and audit are settled, complete every action in `skills/phase-increment-handoff/SKILL.md` in this task. Allocate the next unused Phase 2 ID and create exactly one planning-shape brief at `docs/workplan/phase-2/<allocated-ID>-plan-<increment-slug>.md` for the lowest-order uncovered increment. Give it a **semantic dependency on P0-002**, because it consumes the reviewed Phase 2 README; add other semantic dependencies only for planning artifacts that its increment actually consumes, never for implementation order. Its tracker entry, index row, Phase 0 coverage, Phase 2 coverage, graph measures, and collision report must agree. It performs the next handoff after deriving one implementation brief. Allocate no implementation ID here.

## Write set

Forecast:

- `docs/workplan/phase-2/README.md` (new phase artifact).
- `docs/workplan/README.md` (phase 2 design row, the planning task's dependency row, affected coverage and graph/collision report).
- `docs/workplan/phase-0/README.md` (increment order 2, Plan Phase 2, design and planning-task coverage).
- `docs/workplan/phase-2/<allocated-ID>-plan-<increment-slug>.md` (the planning-shape task from `phase-increment-handoff`).
- `docs/workplan/tasks.json` (that task's pending entry and its semantic dependency on P0-002; preserve all existing task statuses and metadata).

No existing `tasks.json` status write is assigned. A design-set correction discovered to be necessary must be declared with its concrete file and contract before proceeding through Phase 0's phase-revision routing; do not silently broaden this forecast.

## Contract surfaces

None changed in product code or wire schemas. This task decides only phase-2's delivery boundaries; it consumes the cited design contracts. Shared tracker additions, index and Phase 0 coverage edits are write collisions, not product contract collisions. A proposed change to shared product meaning requires Phase 0 disposition and a reconciled graph before assignment.

## Demonstration

At refinement HEAD, the first command below exits 2 because `docs/workplan/phase-2/README.md` does not exist; the cross-reference search finds only scheduled/not-yet-authored coverage, and availability omits in-progress P0-002. These are baseline observations, not expected final results.

After creating the design and its planning task, a reviewer runs from the repository root:

```sh
sed -n '1,240p' docs/workplan/phase-2/README.md
rg -n 'phase-2|P0-002|Plan Phase 2' docs/workplan/README.md docs/workplan/phase-0/README.md
bash scripts/list-available-tasks.sh
python3 - <<'PY'
import json
from pathlib import Path

tracker = json.loads(Path("docs/workplan/tasks.json").read_text())
children = [
    task for task in tracker["tasks"]
    if task["phase"] == 2 and "P0-002" in task["depends_on"]
]
assert len(children) == 1, children
child = children[0]
brief = Path("docs/workplan") / child["brief"]
assert child["status"] == "pending" and brief.is_file(), child
text = brief.read_text()
assert "**Shape:** planning" in text and "P0-002" in text
print(child["id"], child["brief"])
PY
```

Expected artifact observations: the design names one first vendor and its sanitized evidence; records either the evidence-triggered PTY decision or an unchanged D5 trigger; partitions truthful launch, exchange/parity, and native-exit observations; and retains fake parity. Both index and Phase 0 locate the same design and planning task. The script prints the one real Phase 2 ID/path whose pending planning brief depends on P0-002. Phase 2 coverage names that task for increment 1 while implementation-brief coverage remains not yet derived. Availability follows tracker dependencies rather than claiming a runtime milestone. The independent auditor's findings identify the reviewed commit; artifact creation alone is not an audit verdict.

## Verification

- Independent `audit-phase` review checks the selected-vendor evidence, D5 disposition, baseline/prerequisite labels, increment boundaries, task fit, source links, evidence limits, and exit observations; reject absent prerequisites presented as existing behavior or an interactive preference presented as a PTY requirement.
- Run the demonstration, validate `docs/workplan/tasks.json` against `docs/workplan/tasks.schema.json`, resolve every task ID and brief path, and compare tracker/index/Phase 0/Phase 2 coverage. Confirm every `phase-increment-handoff` action holds, its graph/collision report is current, and no implementation ID was allocated.
- Run `git diff --check` for this Markdown-only planning change. Repository Go/web/runtime gates are unchanged and are not task-specific verification; do not claim any runtime capability or gate result from this artifact.

## Acceptance criteria

- [ ] The target artifact is reachable from the index, has every phase-format section, number-only Order values, descriptive Increment titles, and no candidate vendor or increment boundary left for the derivation-task author to decide.
- [ ] The design selects one vendor from dated, sanitized evidence covering launch, I/O, login, exit, licensing/redistribution, and automation; absence of a named choice or evidence for any claimed constraint fails review.
- [ ] D5 is either decided from selected-CLI evidence of resize, raw mode, or TTY detection, or remains deferred with its roadmap trigger. Preference or an interactive default does not satisfy the trigger.
- [ ] Ordered increments give truthful, independently observable outcomes for adapter launch/configuration, real/fake exchange parity, and native exit result plus adapter interpretation, or explicitly justify why a proposed split would have no useful truthful observation.
- [ ] Source-supported behavior and planned prerequisites are labelled separately, and every consumed Stage 0 behavior names its current source or existing task owner; a missing prerequisite or premature completion assertion is a blocking audit finding.
- [ ] Every included concern has one owner; fake parity remains shipped; demonstration limits are explicit; and excluded Stage 2/4/5/6 concerns retain their roadmap destinations.
- [ ] Independent phase audit has no unresolved blocking findings; index and Phase 0 match the reviewed artifact.
- [ ] Every `phase-increment-handoff` action holds: exactly one pending planning-shape task in Phase 2 covers the lowest-order uncovered increment, consumes P0-002 semantically, and matches its real ID/path/title/dependencies across brief, tracker, index, Phase 0 and Phase 2 coverage, graph, and collision report. A missing or duplicate task blocks completion.
- [ ] Unproduced implementation briefs are labelled not yet derived and have no speculative IDs.
- [ ] Existing runtime/source behavior, frozen tracker schema, and task status fields remain intact.

## Deferrals

None. Creating the planning task is part of completion. Implementation-brief derivation belongs to that task, while product implementation remains outside this planning task.

## Traces to

`docs/workplan/phase-0/README.md` ordered planning increment 2 (Plan Phase 2); docs/roadmap.md §3 Stage 1 and §6 D1/D5; docs/contracts.md §8–9; docs/code-structure.md §4; ADR-0003 (Proposed).

