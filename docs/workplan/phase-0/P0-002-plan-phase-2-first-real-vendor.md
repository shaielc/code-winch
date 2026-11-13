# P0-002: Plan phase-2: first real vendor

**Phase:** 0 — Plan the implementation phases
**Shape:** planning
**Dependencies:** None

## Objective

Produce `docs/workplan/phase-2/README.md` with decided ordered increments for roadmap Stage 1 and create the single Phase 2 planning task that derives its first increment, with the design, tracker, index, and coverage records agreeing on that handoff.

## Scope

- Inspect the CLI of the selected real vendor and record sanitized, source-linked evidence for its supported launch, login, input, output, exit, and terminal behavior. Resolve D1 to one named vendor. Resolve D5 only when that evidence shows that resize, raw mode, or TTY detection is required; otherwise retain D5 with its roadmap trigger rather than deciding PTY support speculatively.
- Write `docs/workplan/phase-2/README.md` with the sections required by `skills/shared/phase-design.md`. Separate the source-supported HEAD baseline from planned prerequisites, then partition Stage 1 into the smallest ordered increments that independently expose: the real-vendor selection and capability/launch boundary, native input/output behind the existing attach surface, and the native exit result with its vendor-specific interpretation. If those observations cannot be usefully separated, state the concrete coupling in the phase design rather than leaving the task author to choose the boundary.
- Keep provider-specific launch, codec, login, and exit meaning in the harness adapter and keep generic attach/page behavior provider-neutral. Retain the shipped fake profile and require the standing sandbox scenario to remain runnable against it; distinguish what sanitized fixtures prove from an opt-in real-CLI check that may require local vendor login.
- Plan only the vendor login needed for the loopback, single-user Stage 1 demonstration. Do not represent credentials in run records or logs and do not claim the scoped credential resolution or enforcement assigned to roadmap Stage 5.

## Required constraints

At HEAD, `cmd/winch-sandbox/main.go` registers only the attach server; the tracked product tree contains no harness-adapter package, descriptor, codec, vendor launch path, or native-exit interpretation. `cmd/fake-harness/main.go` is a separately runnable fixture, not a sandbox-owned child or vendor-compatibility proof. The attach contract currently exposes only posture and the Phase 1 child-status work is incomplete: P1-002 and P1-004 are completed, P1-006 is pending, and output, retained input/delivery, and a page-to-harness exchange remain planned Stage 0 prerequisites. Record those facts in the phase baseline and do not mark them implemented or absorb their implementation into Stage 1.

Existing architecture and roadmap specify the wrapper boundary sufficiently to design this artifact; recheck the actual Stage 0 source at dispatch rather than adding an ordering-only dependency.

Use `skills/create-phase/SKILL.md` as the phase-author operation, with `skills/shared/phase-design.md` and the workplan model. After deciding the artifact, complete `skills/phase-increment-handoff/SKILL.md` in the same change so the single planning-shape task, tracker/index entries, and Phase 0 and phase coverage agree. Product behavior and commands in the new phase are predictions. Preserve Phase 0's decided boundaries, the frozen tracker schema, and every existing task status and metadata field.

## Non-goals

- A second vendor and cross-run workflows (roadmap Stage 6).
- Run creation, intentional stop, retry, and lifecycle state (roadmap Stage 4).
- Stop escalation, orphan recovery, scoped credential enforcement, and enforced isolation/egress (roadmap Stage 5).
- Implementing or repairing the unfinished Stage 0 exchange, or implementing any Stage 1 adapter, codec, descriptor, PTY, attach, page, or exit behavior.

This task derives no implementation brief and allocates no implementation ID; the planning task it creates assigns that derivation. Product implementation and completion claims based only on a design remain excluded.

## Planning outputs and handoff

Create `docs/workplan/phase-2/README.md` from the current product tree, `docs/state.md`, roadmap Stage 1 and D1/D5, `docs/contracts.md` §§8–9, `docs/code-structure.md` §4, and ADR-0003. Record the incomplete Stage 0 exchange as planned prerequisites, not a dependency on adjacent planning work. Name the selected vendor and cite stable first-party CLI documentation or checked-in sanitized evidence for each vendor-specific planning fact; do not put credentials, prompts, transcript content, or local account state in the repository.

Choose numbered increments with one owner per included concern, explicit prerequisites, includes/excludes, and observations a person can later make through the existing loopback attach page and operator path. The phase exit must cover the real adapter through the same surface as the fake, required single-user login, fake parity, and visible native exit result plus documented interpretation. Keep implementation coverage **not yet derived** except for the first increment's planning task created by this change. Update the Phase 2 row and dependency/coverage/graph records in `docs/workplan/README.md` and the Plan Phase 2 coverage row in `docs/workplan/phase-0/README.md`.

Complete `skills/phase-increment-handoff/SKILL.md`: allocate the next Phase 2 ID from the tracker, create exactly one `docs/workplan/phase-2/<allocated-ID>-plan-<increment-slug>.md`, and add its pending tracker entry. Give it a semantic dependency on P0-002 because it consumes the phase design produced here. Its brief derives exactly the lowest-order uncovered increment and completes the same handoff for the next increment; it does not reserve or create an implementation ID.

## Write set

Forecast:

- `docs/workplan/phase-2/README.md` (new phase artifact).
- `docs/workplan/README.md` (phase 2 design row, planning-task dependency row, affected coverage, and recalculated graph/collision report).
- `docs/workplan/phase-0/README.md` (increment order 2, Plan Phase 2, design and planning-task coverage).
- `docs/workplan/phase-2/<allocated-ID>-plan-<increment-slug>.md` (the planning-shape task from `phase-increment-handoff`).
- `docs/workplan/tasks.json` (that task's pending entry and its semantic dependency on P0-002; preserve all existing task statuses and metadata).

No existing `tasks.json` status write is assigned. A design-set correction discovered to be necessary must be declared with its concrete file and contract before proceeding through Phase 0's phase-revision routing; do not silently broaden this forecast.

## Contract surfaces

None changed in product code or wire schemas. This task decides only phase-2's delivery boundaries; it consumes the cited design contracts. Shared tracker additions, index and Phase 0 coverage edits are write collisions, not product contract collisions. A proposed change to shared product meaning requires Phase 0 disposition and a reconciled graph before assignment.

## Demonstration

After creating the design and its planning task, run from the repository root:

```sh
sed -n '1,240p' docs/workplan/phase-2/README.md
rg -n 'phase-2|P0-002|Plan Phase 2' docs/workplan/README.md docs/workplan/phase-0/README.md
bash scripts/list-available-tasks.sh
python3 - <<'PY'
import json
from pathlib import Path

tracker = json.loads(Path("docs/workplan/tasks.json").read_text())
task = next(t for t in tracker["tasks"] if t["phase"] == 2 and t["brief"].startswith("phase-2/"))
print(task["id"], task["status"], task["depends_on"], task["brief"])
print(Path("docs/workplan", task["brief"]).is_file())
PY
```

Expected artifact observations: the design names one first vendor; labels current source evidence separately from planned prerequisites; records whether D5 remains open or cites evidence that triggered it; and gives independently observable native exit reporting while retaining fake parity. Both index and Phase 0 locate the same design and planning task. The Python check prints one pending Phase 2 planning task with `['P0-002']` and `True`; its brief and coverage name the same first increment. Availability reflects tracker dependencies and makes no runtime claim. No Phase 2 implementation ID or product source appears in the diff.

## Verification

- Check every phase-format section, numbered increment, boundary, prerequisite, source link, exit observation, and concern owner against `skills/shared/phase-design.md`; fail if absent Stage 0 behavior is presented as current or if D1 remains undecided.
- Validate `docs/workplan/tasks.json` against its unchanged schema; resolve every dependency ID and brief path; compare tracker, index, Phase 0, and Phase 2 coverage; run `bash scripts/list-available-tasks.sh`; and confirm every `phase-increment-handoff` action holds with no implementation ID allocated.
- Run `git diff --check` and the local Markdown-link check used by the repository. The repository-wide Go and web gates are governed by `docs/testing.md`; this planning-only change adds no duplicate runtime gate.

## Acceptance criteria

- [ ] The target artifact is reachable from the index, has every phase-format section, number-only Order values and descriptive Increment titles, and leaves no candidate boundary for the task author to decide; a missing section, nonnumeric order, or unresolved split is visible in artifact review.
- [ ] The design identifies one first vendor and cites sanitized evidence for its launch, login, input, output, exit, and terminal behavior; an unnamed vendor or unsupported CLI claim fails artifact review.
- [ ] D5 is either retained with its exact roadmap trigger or resolved from cited evidence that the selected CLI requires resize, raw mode, or TTY detection; preference or an uncited assumption does not resolve it.
- [ ] Ordered increments retain the fake as a shipped profile and end with a real CLI behind the same loopback attach/page surface, with its native exit result and vendor-specific interpretation visible; an increment that retires the fake, forks the generic UI by provider, or reports only a generic child status fails review.
- [ ] Source evidence and planned prerequisites are labelled separately; injecting a missing prerequisite or a premature completion assertion yields an audit finding, not a pass.
- [ ] Every included concern has one owner, excluded roadmap concerns have their source destination, and a real deferred decision retains its explicit trigger.
- [ ] Index and Phase 0 agree with the created phase artifact on its first increment, coverage, dependency graph, and collision report; differing names, paths, counts, or edges fail the consistency check.
- [ ] Every `phase-increment-handoff` action holds: exactly one planning-shape task in Phase 2 derives the phase's first increment, with a semantic dependency on P0-002 and matching tracker, index, Phase 0, and phase coverage. A missing or second handoff task, a missing edge, or an implementation ID fails the check.
- [ ] Unproduced implementation briefs are labelled not yet derived and have no speculative IDs.
- [ ] Existing runtime/source behavior, frozen tracker schema, and task status fields remain intact.

## Deferrals

None. Creating the first-increment planning task is part of this task. Its future implementation-brief derivation is workflow handoff rather than deferred work owned by this brief; later roadmap stages remain Non-goals with their roadmap destinations.

## Traces to

`docs/workplan/phase-0/README.md` ordered planning increment 2 (Plan Phase 2); `docs/roadmap.md` §3 Stage 1 and §6 D1/D5; `docs/contracts.md` §§8–9; `docs/code-structure.md` §4; `docs/decisions/0003-capability-based-adapters.md` (Proposed).

