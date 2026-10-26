# P0-006: Plan phase-6: enforced isolation

**Phase:** 0 — Plan the implementation phases
**Shape:** planning
**Dependencies:** None

## Objective

Produce a reviewed `docs/workplan/phase-6/README.md` with decided ordered increments for roadmap Stage 5, and a single planning-shape task in Phase 6 that derives its first increment, created with `skills/phase-increment-handoff/SKILL.md`.

## Scope

- Create the missing `docs/workplan/phase-6/README.md` in the format required by `skills/shared/phase-design.md`. Start from the source-supported controls and verification limits in `docs/state.md`; do not treat the retired task table there as the active tracker. Record planned Phase 2–5 outputs as prerequisites to recheck at dispatch, not capabilities present at HEAD.
- Partition enforcement of named profiles, disposable workspace/path boundaries, resource/stop/orphan controls, deny-egress/allowlist and credential reference/injection/redaction guarantees into ordered, independently observable increments. Put number-only values in `Order`, descriptive names in `Increment`, and decide every boundary rather than leaving candidates for a later task author.
- Add a complete LB01–LB11 evidence matrix to the phase artifact. For each blocker, name exactly one owning increment (or a source-supported exclusion), the prerequisite it consumes, and a negative phase-exit observation that fails while the control is absent or bypassed. An exclusion must continue to prohibit shared deployment; it cannot operate as a waiver.
- Assign the checklist's coupled concerns without losing their separate evidence: sensitivity-class retention/export/deletion/cache/backup behavior (LB07), parser and renderer bounds (LB06), browser and runner identities plus revocation/recovery (LB01/LB09/LB10), supply-chain provenance/SBOM/scanning (LB08), host-local attach reachability (LB11), and incident exercises (LB09).
- Keep the implemented static-root containment tests, loopback publication configuration, non-root image configuration, posture reporting, and content-free telemetry conventions reusable, but label their runtime verification limits. They are narrower than shared-deployment assurance and do not close LB01–LB11. Gate any shared-deployment posture on recorded, independently audited blocker evidence.
- Update the workplan index and Phase 0 coverage to point to the decided artifact and its one Phase 6 planning handoff. Recompute dependency and write-collision prose from the resulting tracker rather than copying the current 12-node/9-edge summary.

## Required constraints

Only narrow static traversal checks, loopback publication source, non-root image source and posture reporting exist; LB01–LB11 are not certified closed (docs/state.md).

At refinement HEAD, `docs/workplan/phase-6/README.md` is absent; the index and Phase 0 coverage both say Phase 6 is not yet authored/derived. `P0-006` has no tracker dependencies, no active brief defers work into it, and `bash scripts/list-available-tasks.sh` omits it because its status is already `in_progress`. The retired `P0-006 — harness delivery/reply` row in `docs/state.md` is a historical citation, not this task or an owner. Preserve that distinction in the new artifact.

Existing security and lifecycle design contracts suffice to partition enforcement independently; earlier implementation phases are predicted baselines and must be rechecked at dispatch. No pending artifact is required to author this bounded security design.

Use `skills/create-phase/SKILL.md` as the phase-author operation, with `skills/shared/phase-design.md` and the workplan model. Work in this order: inventory HEAD and predicted prerequisites; decide concern owners and increment order; write the phase artifact; reconcile index and Phase 0 coverage; obtain and resolve an independent `audit-phase` review; then complete `phase-increment-handoff` from the reviewed artifact and reconcile the graph again. The required result is a decided phase artifact, an independent audit with no unresolved blocking finding, one planning-shape handoff task, and consistent tracker/index/Phase 0/Phase 6 coverage. Product behavior and commands in the new phase are predictions. Preserve Phase 0's decided boundaries and every existing task status and metadata field.

## Non-goals

Second vendor/workflows (roadmap Stage 6), hostile public or mutually untrusted container tenancy (the stronger-driver launch blocker in `docs/security.md` §11), and implementation of any security control in this planning task.

Do not certify an LB item from design prose, aggregate multiple blockers behind one untestable "security complete" observation, choose a sandbox driver or deployment product unless the existing design requires that choice, or revise product API/event/port contracts. A necessary contract or phase-boundary correction must use Phase 0's revision routing before this task resumes.

This task derives no implementation brief and allocates no implementation ID; the planning task it creates assigns that derivation. Product implementation and completion claims based only on a design remain excluded.

## Planning outputs and handoff

The phase author creates `docs/workplan/phase-6/README.md`, records the source-supported current baseline separately from planned prerequisites, chooses ordered increments and one owner for every included concern, states per-increment exit observations, includes the LB01–LB11 evidence matrix, and marks all implementation-task coverage **not yet derived**. The artifact must distinguish design-time predictions from evidence needed to claim a blocker closed. Update the Phase 6 row, dependency/graph summary, concurrent-write analysis, and coverage in `docs/workplan/README.md`, plus Plan Phase 6 coverage in `docs/workplan/phase-0/README.md`.

A separate phase auditor uses `skills/audit-phase/SKILL.md` against the decided design. Resolve blocking artifact findings.

Handoff: after the phase audit is resolved, complete every action in `skills/phase-increment-handoff/SKILL.md`. Allocate the next Phase 6 ID from the tracker (it is `P6-001` at refinement HEAD, but collision-check again immediately before writing) and create exactly one planning-shape brief named from the decided first increment. Its tracker entry and brief have a **semantic dependency on P0-006** because they consume the reviewed `docs/workplan/phase-6/README.md`; its objective is to derive one implementation brief for that first increment, not implement it, and it repeats the handoff for the next increment. Record that planning task in the Phase 6 artifact, index, and Phase 0 coverage while leaving its as-yet-unproduced implementation brief **not yet derived**. Dispatch implementation of this phase-design task to an author different from this refiner.

## Write set

Forecast:

- `docs/workplan/phase-6/README.md` (new phase artifact).
- `docs/workplan/README.md` (phase 6 design row, the planning task's dependency row, affected coverage and graph/collision report).
- `docs/workplan/phase-0/README.md` (increment order 6, Plan Phase 6, design and planning-task coverage).
- `docs/workplan/phase-6/P6-001-plan-<first-increment-slug>.md` (forecast from the current counter; use the collision-checked Phase 6 ID and decided slug at implementation time).
- `docs/workplan/tasks.json` (that task's pending entry and its semantic dependency on P0-006; preserve all existing task statuses and metadata).

No existing `tasks.json` status write is assigned. A design-set correction discovered to be necessary must be declared with its concrete file and contract before proceeding through Phase 0's phase-revision routing; do not silently broaden this forecast.

## Contract surfaces

None changed in product code or wire schemas. This task decides only phase-6's delivery boundaries; it consumes the cited design contracts. Shared tracker additions, index and Phase 0 coverage edits are write collisions, not product contract collisions. A proposed change to shared product meaning requires Phase 0 disposition and a reconciled graph before assignment.

## Demonstration

After creating the design and its planning task, a reviewer runs from the repository root:

```sh
cat docs/workplan/phase-6/README.md
rg -n '^\| 6 \||Plan Phase 6|P0-006|P6-001|phase-6/' docs/workplan/README.md docs/workplan/phase-0/README.md docs/workplan/phase-6/README.md
python3 - <<'PY'
import json
from pathlib import Path

tracker = json.loads(Path("docs/workplan/tasks.json").read_text())
phase_tasks = [task for task in tracker["tasks"] if task["phase"] == 6]
assert len(phase_tasks) == 1, phase_tasks
task = phase_tasks[0]
assert task["depends_on"] == ["P0-006"], task
assert task["status"] == "pending", task
brief = Path("docs/workplan") / task["brief"]
assert brief.is_file(), brief
text = brief.read_text()
assert "**Shape:** planning" in text
assert "P0-006" in text
print(task["id"], brief)
PY
bash scripts/list-available-tasks.sh
```

Before implementation, the first command fails because the Phase 6 artifact is absent; the index and Phase 0 report it as not authored/derived, and availability omits this already-in-progress task. Afterward, the artifact shows every LB01–LB11 exactly once in its evidence matrix with a concrete owner and independently observable negative exit proof; unenforced egress or an off-host attach port cannot be represented as safe shared deployment. Both index and Phase 0 locate the same design and planning task. The Python check proves that exactly one Phase 6 task exists, is pending, has the real brief path and exact semantic dependency, and is planning-shaped. Inspect its objective and coverage to confirm it names the decided first increment, derives rather than implements one brief, repeats the handoff, and leaves that implementation brief not yet derived. Availability must follow tracker dependencies rather than claim a runtime milestone. Record the independent auditor's reviewed commit and resolved findings; artifact creation alone is not an audit verdict.

## Verification

- Run an independent `audit-phase` review against the candidate Phase 6 artifact. It checks baseline evidence, decided boundaries, one concern owner, task fit, source links, prerequisite labels, all eleven blocker rows, and independently observable exit failures; reject absent prerequisites presented as existing behavior and resolve every blocking finding before handoff.
- Run the demonstration above. Validate `docs/workplan/tasks.json` against its unchanged schema, verify every dependency ID and brief path, and compare tracker/index/Phase 0/Phase 6 coverage. Run the collision and graph checks required by `create-phase` and `phase-increment-handoff`; confirm exactly one Phase 6 planning task was allocated and no implementation-task ID was allocated.
- Run `git diff --check`, the Markdown local-link/path check used by the repository, completion-script tests (`python3 -m unittest discover -s test`), control-panel tests (`cd workplan-control-panel && python3 -m unittest discover -s tests`), and `make check` per AGENTS. This documentation task does not require starting a deployment or claim runtime capability. Report unavailable toolchains explicitly; do not claim a Go/runtime gate passed when it did not run.

## Acceptance criteria

- [ ] The target artifact is reachable from the index, has the phase-format sections, number-only Order values and descriptive Increment titles, and has no candidate boundary left for the task author to decide.
- [ ] The source baseline names only controls present at HEAD and their verification limits; planned Phase 2–5 inputs are labelled prerequisites to recheck, and the retired historical P0-006 is not treated as the active task or an owner.
- [ ] Every LB01–LB11 appears exactly once in the evidence matrix with one owning increment or source-supported exclusion, its prerequisite, and an independently observable negative phase-exit proof. An exclusion still blocks shared deployment; unenforced egress or an off-host attach port cannot be represented as safe.
- [ ] Source evidence and planned prerequisites are labelled separately; injecting a missing prerequisite or a premature completion assertion yields an audit finding, not a pass.
- [ ] Every included concern has one owner, excluded roadmap concerns have their source destination, and a real deferred decision retains its explicit trigger.
- [ ] Independent phase audit identifies the reviewed commit and has no unresolved blocking findings; index and Phase 0 match that reviewed artifact.
- [ ] Every `phase-increment-handoff` action holds: exactly one pending planning-shape task in Phase 6 derives one brief for the decided first increment, repeats the next-increment handoff, has the exact semantic dependency `[P0-006]`, and matches tracker, index, Phase 0, and Phase 6 coverage. A missing, duplicate, implementation-shaped, or collision-prone task blocks completion.
- [ ] Unproduced implementation briefs are labelled not yet derived and have no speculative IDs.
- [ ] Tracker schema and all pre-existing task records/status fields remain byte-for-byte intact; only the one new Phase 6 planning record is appended. Existing runtime/source behavior remains intact, and the recalculated graph, availability, and write-collision records agree with the tracker.

## Deferrals

None. Creating the planning task is part of completion. Implementation-brief derivation belongs to that task, while product implementation remains outside this planning task.

## Traces to

`docs/workplan/phase-0/README.md` ordered planning increment 6 (Plan Phase 6); docs/roadmap.md §3 Stage 5 and §5; docs/security.md §2–12, especially §11 LB01–LB11; docs/contracts.md §1–5/8.
