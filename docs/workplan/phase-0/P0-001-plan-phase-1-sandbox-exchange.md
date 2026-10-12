# P0-001: Plan phase-1: sandbox exchange

**Phase:** 0 — Plan the implementation phases
**Shape:** planning
**Dependencies:** None

## Objective

Produce a reviewed `docs/workplan/phase-1/README.md` with decided ordered increments for roadmap Stage 0, together with a concrete tracked Phase 0 follow-up assigning derivation of the next uncovered implementation increment to a separate task author.

## Scope

- Create `docs/workplan/phase-1/README.md` with the exact sections in `skills/shared/phase-design.md`: source-supported outcome and starting point, concern owners, ordered increments, task coverage, and phase exit. Use number-only `Order` cells and descriptive increment titles; mark every implementation increment **not yet derived** and identify the next one without allocating its ID.
- Decide separate observable increments for (1) child ownership plus ephemeral live output, (2) retained ordinal history and reload, (3) persisted input acceptance/refusal, (4) delivery to the child and correlated reply, and (5) malformed/oversized output plus stalled-reader isolation. A row may include the wiring needed to make its own observation real, but may not hide a later observation behind that wiring.
- Preserve the roadmap's two branches: live output precedes output retention, and recorded input precedes delivery; neither branch depends on the other merely for ordering. State the actual join required by the private record store or delivered exchange rather than inventing a product dependency.
- Assign child launch/reaping and shutdown cleanup to the first increment that starts the harness, input-availability/refusal to the delivery increment, and payload/output bounds and bounded fan-out to their observable hardening increment. Assign each maintained page or `winch` operator path, all-fake scenario growth, and inherited run/deploy gate repair to the same increment as the behavior it proves.
- Treat `cmd/fake-harness/main.go` as a changeable fixture, not a protocol constraint. Require each future implementation brief to document what its fake demonstration does not prove.

## Required constraints

Refinement evidence at task-branch HEAD:

- `cmd/winch-sandbox/main.go:20-25` constructs only `attach.New`; no runner or child is registered.
- `internal/adapters/transport/attach/server.go:22-40` rejects every method except GET/HEAD and exposes only health, posture, and static-page routing.
- `web/src/attach/App.tsx:7-16,26-44` keeps submissions in component state and sends nothing to the server; reload loses them.
- `cmd/winch/main.go:8-17` exposes only `winch status`, while `deployments/Dockerfile:14-24` ships `fake-harness` but starts only `winch-sandbox`.
- `git ls-files 'internal/runner/*' 'cmd/winch/stream.go' 'cmd/winch/records.go'` produces no paths. The existing end-to-end scenario checks startup/status only; it does not exercise harness I/O (`test/e2e/scenario_sandbox_starts_test.go:10-30`).

These are source observations, not runtime verification. The refinement demonstration also confirms that `docs/workplan/phase-1/README.md` does not yet exist, the index says Phase 1 is not yet derived, and `bash scripts/list-available-tasks.sh` omits this in-progress task while listing the other currently available tasks.

This task has no dependency edge and consumes no pending product behavior. Its authorities are current source, roadmap Stage 0, Phase 0 increment 0.1, contracts §8–9, and ADR-0005/0006. The historical use of the same ID in `docs/state.md` describes a retired plan and is not a deferral or prerequisite for this task.

Use `skills/create-phase/SKILL.md` as the phase-author operation, with `skills/shared/phase-design.md` and the workplan model. The required behavior is a decided phase artifact, an independently authored tracked follow-up planning brief, consistent tracker/index/Phase 0 coverage, and an independent `audit-phase` review. Product behavior and commands described there are predictions. Preserve Phase 0's decided boundaries and all task status fields.

## Non-goals

Product code, runtime wiring, implementation briefs, and changes to product contracts are excluded. So are run lifecycle, the control plane, vendor semantics and exit interpretation, canonical sequence numbers, Stage 2 seamless ordinal resume, and shared-deployment security (roadmap Stages 1–5). Stage 0 may plan reloadable history and slow-reader isolation, but must not relabel either as Stage 2's gap-free reconnect semantics.

The assigned phase author does not derive implementation briefs or author the follow-up planning brief. Product implementation and completion claims based only on a design remain excluded. Arranging independent derivation of the tracked follow-up is required for this task to finish; a verbal handoff is insufficient.

## Planning outputs and handoff

The phase author uses `skills/create-phase/SKILL.md` to create `docs/workplan/phase-1/README.md`, records the source-supported baseline above separately from predicted prerequisites, chooses the boundaries in Scope without leaving alternatives to a task author, states exit observations, and marks implementation task coverage **not yet derived**. Update Phase 1's index row and Phase 0 coverage in the same commit. The artifact decides responsibility and order, not API shapes, fixture flags, commands, file forecasts, or substrate choices.

A separate phase auditor uses `skills/audit-phase/SKILL.md` against the reviewed design commit. Record the audited commit and resolve every blocking artifact finding before handoff. Then arrange for a **different task author** to use `skills/derive-task/SKILL.md` to produce at least one concrete Phase 0 planning follow-up brief assigning derivation of the next uncovered Phase 1 implementation increment. The phase author cannot author that follow-up. This independent derivation is an expected artifact of P0-001's completion, not merely an invitation to do more work.

Allocate the follow-up's next plan-wide ID only when its complete brief is actually derived; create its real file and matching pending tracker entry, index row and Phase 0 coverage together. The follow-up records a **semantic dependency on P0-001** because it consumes the reviewed `docs/workplan/phase-1/README.md` produced here. Name the target increment until a real implementation brief ID exists; do not allocate a placeholder ID for the implementation task. Recompute the actual partial graph and collision report after each addition, preserving all existing statuses.

The follow-up itself assigns a separate task author to derive one implementation brief at a time against the decided phase artifact, reconcile its tracker/index/coverage records, and arrange the next concrete tracked planning follow-up when actionable coverage remains. Repeat that handoff until the phase's actionable increments are covered. Do not bundle all increments into one derivation or depend on product implementation simply to plan its brief. Dispatch this phase-design task through a separate refiner using `skills/refine-task/SKILL.md` against current HEAD.

## Write set

Forecast:

- `docs/workplan/phase-1/README.md` (new phase artifact).
- `docs/workplan/README.md` (phase 1 design row, actual follow-up dependency row, affected coverage and graph/collision report).
- `docs/workplan/phase-0/README.md` (increment order 1, legacy reference 0.1, design and follow-up coverage/handoff).
- `docs/workplan/phase-0/<allocated-ID>-<target-increment-slug>.md` (new follow-up planning brief, independently authored after the design is decided; concrete ID/title/path recorded then).
- `docs/workplan/tasks.json` (add the actual follow-up entry and its semantic dependency on P0-001; preserve all existing task statuses and metadata).

No existing `tasks.json` status write is assigned; the new independently derived follow-up begins pending with the model's normal unassigned metadata. A design-set correction discovered to be necessary must be declared with its concrete file and contract before proceeding through Phase 0's phase-revision routing; do not silently broaden this forecast.

## Contract surfaces

None changed in product code or wire schemas. This task decides only phase-1's delivery boundaries; it consumes the cited design contracts. Shared tracker additions, index and Phase 0 coverage edits are write collisions, not product contract collisions. A proposed change to shared product meaning requires Phase 0 disposition and a reconciled graph before assignment.

## Demonstration

The refinement established the before-state from the repository root: the first command fails because the phase artifact is absent, the search finds only scheduled/not-derived coverage, and availability excludes this in-progress task. After creating the design and independently authored follow-up, a reviewer runs:

```sh
sed -n '1,240p' docs/workplan/phase-1/README.md
rg -n 'phase-1|P0-001|0\.1' docs/workplan/README.md docs/workplan/phase-0/README.md
bash scripts/list-available-tasks.sh
python3 - <<'PY'
import json
from pathlib import Path

tasks = json.loads(Path("docs/workplan/tasks.json").read_text())["tasks"]
task = next(task for task in tasks if task["id"] == "P0-001")
followups = [candidate for candidate in tasks if task["id"] in candidate["depends_on"]]
assert followups, "no tracked follow-up consumes the phase design"
for followup in followups:
    assert (Path("docs/workplan") / followup["brief"]).is_file()
print([(item["id"], item["brief"]) for item in followups])
PY
```

Expected artifact observations: the phase artifact renders completely; its ordered rows distinguish ephemeral live output, reloadable history, persisted submission, harness delivery/reply, and hardening without mistaking the existing local echo for delivery or reload for seamless resume. Both index and Phase 0 locate the same design and actual follow-up. The Python check prints at least one real ID/path pair. Inspect that brief and tracker entry to confirm the dependency reason is semantic consumption of this design, the next uncovered increment is named, implementation coverage remains pending, and no placeholder implementation ID exists. Availability reflects tracker dependencies, not a claimed runtime milestone. The independent audit identifies its reviewed commit and has no unresolved blocking finding.

## Verification

- Independent `audit-phase` review checks baseline evidence, decided boundaries, task fit, source links and exit observations; reject absent prerequisites presented as existing behavior.
- Run the four demonstration commands, validate `docs/workplan/tasks.json` against its unchanged schema, resolve every dependency ID and brief path, and compare tracker/index/Phase 0 coverage. Confirm at least one real follow-up and tracker entry exist, its dependency reason says it consumes the reviewed phase design, and no placeholder implementation ID was allocated.
- Run `python3 -m unittest discover -s test`, `cd workplan-control-panel && python3 -m unittest discover -s tests`, `git diff --check`, and `make check`. Because this task adds no runtime capability, `make test-cycle` is a preservation gate rather than evidence for the predicted exchange; run it when Docker is available and report an environment limitation otherwise.

## Acceptance criteria

- [ ] The target artifact is reachable from the index, has the phase-format sections, number-only Order values and descriptive Increment titles, and has no candidate boundary left for the task author to decide.
- [ ] A reviewer can distinguish ephemeral live output, reloadable history, persisted submission, and harness reply without mistaking the local echo for delivery.
- [ ] Harness ownership/cleanup, input availability, output bounds, stalled readers, maintained operator access, scenario growth, and gate repairs each have exactly one increment owner; the phase does not prematurely prescribe their API, fixture, or storage mechanisms.
- [ ] Source evidence and planned prerequisites are labelled separately; injecting a missing prerequisite or a premature completion assertion yields an audit finding, not a pass.
- [ ] Stage 1 vendor behavior, Stage 2 seamless resume, Stages 3–4 control-plane/run behavior, and Stage 5 shared-deployment security remain excluded with their roadmap destinations; any cited deferred decision retains its roadmap §6 trigger.
- [ ] Independent phase audit has no unresolved blocking findings; index and Phase 0 match the reviewed artifact.
- [ ] At least one complete independently authored Phase 0 follow-up brief exists in the tracker and index, assigns the next uncovered phase-1 implementation increment, records its semantic dependency on P0-001, and has matching Phase 0 coverage. A missing follow-up blocks completion.
- [ ] The follow-up specifies repeated concrete handoffs for remaining actionable coverage; unproduced implementation briefs are labelled not yet derived and have no speculative IDs.
- [ ] Existing runtime/source behavior, frozen tracker schema, task status fields, and separate-author handoff remain intact.

## Deferrals

None. Producing the independently authored tracked follow-up is part of completion. Actual implementation-brief derivation is assigned to that real follow-up, while product implementation remains outside this planning task.

## Traces to

`docs/workplan/phase-0/README.md` ordered planning increment 1 (legacy reference 0.1); docs/roadmap.md §3 Stage 0 and §4; docs/contracts.md §8–9; ADR-0005; ADR-0006.
