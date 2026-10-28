# P0-002: Plan phase-2: first real vendor

**Phase:** 0 — Plan the implementation phases
**Shape:** planning
**Dependencies:** None

## Objective

Produce a reviewed `docs/workplan/phase-2/README.md` with decided ordered increments for roadmap Stage 1, and a single planning-shape task in Phase 2 that derives its first increment, created with `skills/phase-increment-handoff/SKILL.md`.

## Scope

- Inspect the Stage 0 implementation and briefs that exist at the implementation commit before treating exchange, input delivery, persistence, or child ownership as prerequisites. Record each consumed capability as either observed at HEAD or planned, with the owning Phase 1 brief and dependency that a later derived implementation brief must take. Phase design does not depend on unfinished Phase 1 implementation merely to be authored.
- Resolve D1 in the phase artifact rather than passing a shortlist to the task author. Compare currently supported, documented CLI behavior using dated vendor documentation plus a sanitized local probe or fixture evidence where available: non-interactive launch, input channel, incremental output, native exit status, authentication mode, licensing/automation constraints, and whether behavior changes when stdin/stdout are pipes. Name one vendor, the evidence and evidence limits, and why the rejected candidates do not change the generic seam.
- Resolve D5 only from that selected CLI's evidence. State pipes as the decided Phase 2 transport when resize, raw mode, or TTY detection is not required; otherwise state the concrete terminal semantic that triggers PTY work and give it its own observable increment before behavior that consumes it. Do not leave pipes-versus-PTY to implementation.
- Partition Stage 1 into the smallest useful, ordered observations available through the existing attach page and wrapper: truthful selected-profile/capability presentation; adapter registration and launch construction; selected-vendor login/preflight at single-user depth; compatible input/output through the wrapper; and native exit result plus harness-specific interpretation. Combine rows only when the phase artifact explains why the earlier observation would not be independently useful.
- Keep descriptor, launch, input, incremental decoding, and native-exit mapping behind the capability-based harness boundary in `docs/decisions/0003-capability-based-adapters.md` and `docs/code-structure.md` §4. Assign explicit composition-root registration and an operator-visible path to the increment that first makes the real profile usable; do not prescribe provider branches in generic attach/application code.
- Retain fake-profile parity as a shipped configuration in every affected boundary and scenario. The phase may change the fake's private protocol and controls; it must state that fake evidence proves wrapper behavior, not vendor compatibility, login, terminal semantics, or provider availability.

## Required constraints

At refinement HEAD, `cmd/winch-sandbox/main.go` constructs only `attach.New`; `cmd/fake-harness/main.go` is a separately runnable fixture; and `git ls-files 'internal/adapters/harness/*' 'internal/runner/*'` is empty. No real adapter, descriptor, codec, native exit interpretation, integrated fake profile, or PTY implementation exists. Stage 0 exchange remains a planned prerequisite, never an existing capability merely because Phase 1 briefs describe it. `docs/state.md` is the closure record; where its retired task labels differ from the active tracker, current source and active workplan records govern status.

Existing architecture and roadmap specify the wrapper boundary sufficiently to design this artifact; recheck the actual Stage 0 source at dispatch rather than adding an ordering-only dependency.

Use `skills/create-phase/SKILL.md` as the phase-author operation, with `skills/shared/phase-design.md` and the workplan model. The required behavior is a decided phase artifact, an independent `audit-phase` review of the authored commit, the single planning-shape task from `skills/phase-increment-handoff/SKILL.md`, and consistent tracker/index/Phase 0 and phase coverage. Product behavior and commands described there are predictions. Preserve Phase 0's decided boundaries, the tracker schema, and all existing task status fields.

## Non-goals

Second vendor/workflows (roadmap Stage 6); control-plane run start/stop lifecycle (Stage 4); stop escalation, orphan recovery, scoped credential injection and enforced egress/isolation (Stage 5); structured views and resumable history (Stage 2). Vendor-required login/preflight at the existing single-user attach depth is in scope, but generalized credential storage or shared-deployment authorization is not.

This task derives no implementation brief and allocates no implementation ID; the planning task it creates assigns that derivation. Product implementation and completion claims based only on a design remain excluded.

## Planning outputs and handoff

The phase author creates `docs/workplan/phase-2/README.md` in the shared phase format. It must record the source-supported baseline separately from planned Phase 1 prerequisites; a dated D1 evidence table and one selected vendor; the D5 result and supporting terminal evidence; ordered, number-only increment rows with one observable change apiece; concern owners and excluded destinations; exit observations; and implementation-task coverage marked **not yet derived**. Update Phase 2's row in the index and Phase 0 coverage. Inspect current source and applicable instructions first, then compare the reachable attach experience and the latest Phase 1 briefs against roadmap Stage 1. No product protocol or substrate implementation is prescribed by this brief.

A separate phase auditor uses `skills/audit-phase/SKILL.md` against the authored commit. Record the reviewed commit and resolve every blocking artifact finding before completion; do not treat the phase author's self-review as the independent verdict.

Handoff: after the audit is resolved, complete `skills/phase-increment-handoff/SKILL.md` in this task. Allocate the next Phase 2 ID from the tracker and create exactly one planning-shape brief, `docs/workplan/phase-2/<allocated-ID>-plan-<increment-slug>.md`, for the lowest-order uncovered increment. Its tracker edge to P0-002 is **semantic** because it consumes the reviewed Phase 2 design; add no edge merely for implementation order. That task completes the same handoff for the next uncovered increment. Leave every implementation ID unallocated and hand the new planning brief to a different refiner.

## Write set

Forecast:

- `docs/workplan/phase-2/README.md` (new phase artifact).
- `docs/workplan/README.md` (phase 2 design row, the planning task's dependency row, affected coverage and graph/collision report).
- `docs/workplan/phase-0/README.md` (increment order 2, Plan Phase 2, design and planning-task coverage).
- `docs/workplan/phase-2/<allocated-ID>-plan-<increment-slug>.md` (the planning-shape task from `phase-increment-handoff`).
- `docs/workplan/tasks.json` (that task's pending entry and its semantic dependency on P0-002; preserve all existing task statuses and metadata).

No product source, design-set document, deployment file, Phase 1 brief, or tracker schema is forecast. Vendor research is evidence recorded in the Phase 2 artifact, not authorization to add a transcript containing prompts, credentials, workspace content, or other secrets.

No existing `tasks.json` status write is assigned. A design-set correction discovered to be necessary must be declared with its concrete file and contract before proceeding through Phase 0's phase-revision routing; do not silently broaden this forecast.

## Contract surfaces

None changed in product code or wire schemas. This task decides only phase-2's delivery boundaries; it consumes the cited design contracts. Shared tracker additions, index and Phase 0 coverage edits are write collisions, not product contract collisions. A proposed change to shared product meaning requires Phase 0 disposition and a reconciled graph before assignment.

## Demonstration

Before implementation, the missing Phase 2 artifact makes the first command fail and the two coverage records say **not yet authored/derived**; `P0-002` is absent from availability while its tracker status is `in_progress`. After creating the design and its handoff task, a reviewer runs from the repository root:

```sh
sed -n '1,240p' docs/workplan/phase-2/README.md
rg -n '^\| \[?2|Plan Phase 2|P0-002' docs/workplan/README.md docs/workplan/phase-0/README.md
python3 - <<'PYTHON'
import json
from pathlib import Path

tracker = json.loads(Path("docs/workplan/tasks.json").read_text())
phase_two = [task for task in tracker["tasks"] if task["phase"] == 2]
planning = [task for task in phase_two if task["title"].startswith("Plan Phase 2 increment 1:")]
assert len(planning) == 1, planning
task = planning[0]
assert task["status"] == "pending", task
assert task["depends_on"] == ["P0-002"], task
brief = Path("docs/workplan") / task["brief"]
assert brief.is_file(), brief
text = brief.read_text()
assert "**Shape:** planning" in text
assert "P0-002 (semantic:" in text
print(task["id"], task["brief"])
PYTHON
bash scripts/list-available-tasks.sh
```

Expected artifact observations: The design names one vendor and the dated evidence limits behind D1, states either pipes or the exact D5 PTY trigger, assigns an independently observable native exit result plus interpretation, and retains fake parity. Both index and Phase 0 locate the same design and planning task. The Python check proves there is exactly one pending Phase 2 increment-1 planning task, its path resolves, and its semantic dependency is recorded in both tracker and brief. Phase 2 implementation coverage remains not yet derived. Availability reflects tracker dependencies, not a claimed runtime milestone. The pull request records the independent auditor's reviewed commit and disposition; artifact creation alone is not an audit verdict.

## Verification

- Independent `audit-phase` review checks the dated vendor evidence, D1/D5 decisions, separation of HEAD observations from planned prerequisites, increment boundaries, task fit, source links, and exit observations; reject absent prerequisites presented as existing behavior or a vendor/transport choice delegated to task derivation.
- Run the demonstration, validate the tracker against its unchanged schema, resolve every dependency ID and brief path, and compare tracker/index/Phase 0/Phase 2 coverage. `bash scripts/list-available-tasks.sh` must preserve dependency semantics. Confirm every `phase-increment-handoff` action holds, exactly one Phase 2 increment-1 planning task was added, and no implementation ID was allocated.
- `git diff --check`, completion-script tests (`python3 -m unittest discover -s test`) and `make check` per AGENTS. Report unavailable toolchains explicitly; do not claim a Go/runtime gate passed when it did not run. No runtime capability is added by this task.

## Acceptance criteria

- [ ] The target artifact is reachable from the index, has the phase-format sections, number-only Order values and descriptive Increment titles, and has no candidate boundary left for the task author to decide.
- [ ] The design identifies one first vendor from a dated, sanitized comparison of launch, I/O, exit, login, automation and pipe behavior; records evidence limits; and leaves no vendor shortlist or selection criterion for the task author to decide.
- [ ] D5 is explicitly resolved to pipes or to a concrete observed terminal semantic with PTY work placed before its consumer; native exit result and harness-specific interpretation are independently observable while fake parity is retained.
- [ ] Source evidence and planned prerequisites are labelled separately; injecting a missing prerequisite or a premature completion assertion yields an audit finding, not a pass.
- [ ] Every included concern has one owner, excluded roadmap concerns have their source destination, and a real deferred decision retains its explicit trigger.
- [ ] Independent phase audit names the reviewed commit and has no unresolved blocking findings; index and Phase 0 match that reviewed artifact.
- [ ] Every `phase-increment-handoff` action holds: exactly one planning-shape task in Phase 2 derives the phase's first increment, with a semantic dependency on P0-002 and matching tracker, index, and Phase 0 and phase coverage. A missing task blocks completion.
- [ ] Unproduced implementation briefs are labelled not yet derived and have no speculative IDs.
- [ ] Existing runtime/source behavior, frozen tracker schema, and task status fields remain intact.

## Deferrals

None. Creating the planning task is part of completion. Implementation-brief derivation belongs to that tracked task. Stage 2, Stage 4, Stage 5, and Stage 6 exclusions are roadmap destinations, not unowned TODOs; product implementation remains outside this planning task.

## Traces to

`docs/workplan/phase-0/README.md` ordered planning increment 2 (Plan Phase 2); `docs/roadmap.md` §3 Stage 1 and §6 D1/D5; `docs/contracts.md` §8–9; `docs/code-structure.md` §4; `docs/decisions/0003-capability-based-adapters.md` (Proposed).
