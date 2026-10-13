# P0-002: Plan phase-2: first real vendor

**Phase:** 0 — Plan the implementation phases
**Shape:** planning
**Dependencies:** None

## Objective

Produce a reviewed `docs/workplan/phase-2/README.md` with decided ordered increments for roadmap Stage 1, together with a concrete tracked Phase 0 follow-up assigning derivation of the next uncovered implementation increment to a separate task author.

## Scope

- Inspect the Stage 0 sources that exist when this task is dispatched, beginning with `cmd/winch-sandbox`, `cmd/fake-harness`, `internal/runner`, the attach transport and its end-to-end scenarios. Record source-supported capabilities separately from the Stage 0 behavior Phase 2 will require; do not turn a missing Stage 0 prerequisite into Phase 2 scope.
- Resolve D1 in the phase artifact. Name one vendor CLI and cite sanitized, reproducible evidence for its supported version, non-interactive launch, input/output behavior, native exit behavior, and login prerequisite. Evidence may be checked-in sanitized command output or authoritative public documentation; it must not contain credentials, prompts, workspace content, or account identifiers.
- Resolve D5 only when that selected CLI's evidence shows resize, raw-mode, or TTY-detection behavior. State either the concrete PTY requirement and the evidence that triggered it, or that pipes remain the planned transport and D5 remains deferred under its roadmap trigger; do not prescribe both as candidates for a later task author.
- Partition Stage 1 from the page inward into ordered, independently observable increments. Each increment must name its consumed Stage 0 behavior, its visible result through the existing attach page or maintained operator path, and the concern it owns. Together the increments introduce one genuine adapter seam, a versioned capability descriptor, launch/input/output and login boundaries, vendor decoding, and native exit result plus harness-specific interpretation.
- Keep the fake as a selectable shipped runtime profile and assign parity evidence to the increment that first changes the wrapper seam. Plan required vendor login only at the single-user Stage 1 depth; report its prerequisite truthfully and do not claim the scoped-credential enforcement owned by Stage 5.

## Required constraints

At refinement HEAD, `docs/workplan/phase-2/README.md` does not exist. `docs/state.md` records that the sandbox composition root starts no harness, the shipped fake is standalone, and no real adapter, descriptor, codec, native exit interpretation, provider login, or PTY implementation exists. Stage 0 exchange is therefore a planned prerequisite, never current behavior. The phase artifact must cite the rechecked source rather than copying this dated observation as proof.

Existing architecture and roadmap specify the wrapper boundary sufficiently to design this artifact; recheck the actual Stage 0 source at dispatch rather than adding an ordering-only dependency.

Use `skills/create-phase/SKILL.md` as the phase-author operation, with `skills/shared/phase-design.md` and the workplan model. The required result is a decided phase artifact, an independently authored tracked follow-up planning brief, consistent tracker/index/Phase 0 coverage, and an independent `audit-phase` review. Product behavior and commands described there are predictions. Preserve Phase 0's decided boundaries and all existing task status fields.

## Non-goals

Second vendor/workflows (Stage 6), run start/stop lifecycle (Stage 4), stop escalation and orphan recovery (Stage 5).

The assigned phase author does not derive implementation briefs or author the follow-up planning brief. Product implementation and completion claims based only on a design remain excluded. Arranging independent derivation of the tracked follow-up is required for this task to finish; a verbal handoff is insufficient.

## Planning outputs and handoff

The phase author creates `docs/workplan/phase-2/README.md` in the exact format of `skills/shared/phase-design.md`. Record the source-supported baseline and runtime-verification limits; give every concern one owner; use number-only `Order` cells and descriptive increment names; separate planned prerequisites from existing behavior; state observable exit evidence; and identify the first uncovered increment while marking its implementation brief **not yet derived**. Update the Phase 2 row in the index and the Phase 0 `0.2` coverage row to agree with the reviewed artifact. No product protocol, command syntax, file layout, or substrate implementation is prescribed by this brief.

A separate phase auditor uses `skills/audit-phase/SKILL.md` against the commit containing the decided design. Record the reviewed commit and resolve every blocking artifact finding before handoff. Then arrange for a **different task author** to use `skills/derive-task/SKILL.md` to produce one concrete Phase 0 planning follow-up brief assigning derivation of the first uncovered implementation increment in Phase 2. The phase author cannot author that follow-up. This independent derivation is an expected artifact of P0-002's completion, not merely an invitation to do more work.

Allocate the follow-up's next plan-wide ID only when its complete brief is actually derived; create its real file and matching pending tracker entry, index row and Phase 0 coverage together. The follow-up records a **semantic dependency on P0-002** because it consumes the reviewed `docs/workplan/phase-2/README.md` produced here. Name the target increment until a real implementation brief ID exists; do not allocate a placeholder ID for the implementation task. Recompute the actual partial graph and collision report after each addition, preserving all existing statuses.

The follow-up itself assigns a separate task author to derive one implementation brief at a time against the decided phase artifact, reconcile its tracker/index/coverage records, and arrange the next concrete tracked planning follow-up when actionable coverage remains. Repeat that handoff until the phase's actionable increments are covered. Do not bundle all increments into one derivation or depend on product implementation simply to plan its brief. Dispatch this phase-design task through a separate refiner using `skills/refine-task/SKILL.md` against current HEAD.

## Write set

Forecast:

- `docs/workplan/phase-2/README.md` (new phase artifact).
- `docs/workplan/README.md` (phase 2 design row, actual follow-up dependency row, affected coverage and graph/collision report).
- `docs/workplan/phase-0/README.md` (increment order 2, legacy reference 0.2, design and follow-up coverage/handoff).
- `docs/workplan/phase-0/<allocated-ID>-<target-increment-slug>.md` (new follow-up planning brief, independently authored after the design is decided; concrete ID/title/path recorded then).
- `docs/workplan/tasks.json` (add the actual follow-up entry and its semantic dependency on P0-002; preserve all existing task statuses and metadata).

No existing `tasks.json` status write is assigned; the new independently derived follow-up begins pending with the model's normal unassigned metadata. A design-set correction discovered to be necessary must be declared with its concrete file and contract before proceeding through Phase 0's phase-revision routing; do not silently broaden this forecast.

## Contract surfaces

None changed in product code or wire schemas. This task decides only phase-2's delivery boundaries; it consumes the cited design contracts. Shared tracker additions, index and Phase 0 coverage edits are write collisions, not product contract collisions. A proposed change to shared product meaning requires Phase 0 disposition and a reconciled graph before assignment.

## Demonstration

The refinement baseline produces a missing-file error for the first command, while the index and Phase 0 describe Phase 2 as not yet derived. After creating the design and independently authored follow-up, a reviewer runs from the repository root:

```sh
test -f docs/workplan/phase-2/README.md && sed -n '1,260p' docs/workplan/phase-2/README.md
rg -n 'phase-2|P0-002|0\.2|first real vendor' docs/workplan/README.md docs/workplan/phase-0/README.md
python3 - <<'PY'
import json
from pathlib import Path

tasks = json.loads(Path("docs/workplan/tasks.json").read_text())["tasks"]
children = [task for task in tasks if "P0-002" in task["depends_on"]]
assert children, "P0-002 has no tracked follow-up"
for task in children:
    brief = Path("docs/workplan") / task["brief"]
    assert brief.is_file(), brief
    print(task["id"], task["title"], brief)
PY
bash scripts/list-available-tasks.sh
```

Expected artifact observations: The first command succeeds and shows a complete phase artifact. It identifies one first vendor, cites sanitized evidence, makes a single evidence-backed PTY-or-pipes decision, and assigns an observable native exit result plus interpretation while retaining fake parity. The index and Phase 0 locate the same design and actual follow-up brief. The Python check prints at least one real follow-up whose tracker entry resolves to a file and records the semantic dependency on P0-002; inspection of that brief names the first uncovered increment and preserves not-yet-derived implementation coverage. Availability reflects tracker dependencies, not a claimed runtime milestone. The independent audit names the reviewed commit and has no unresolved blocking findings; artifact creation alone is not an audit verdict.

## Verification

- Independent `audit-phase` review of the named design commit checks baseline evidence, vendor and terminal decisions, increment boundaries, task fit, source links and exit observations; reject absent prerequisites presented as existing behavior, unsanitized vendor evidence, or a PTY choice left to task derivation.
- Run the Demonstration commands. Validate `docs/workplan/tasks.json` against its unchanged schema, resolve every dependency ID and brief path, and compare tracker/index/Phase 0/Phase 2 coverage. Confirm at least one real follow-up planning brief and tracker entry exist, its semantic dependency consumes the producing phase-design artifact, and no placeholder implementation ID was allocated.
- `git diff --check`, completion-script tests (`python3 -m unittest discover -s test`) and `make check` per AGENTS. Report unavailable toolchains explicitly; do not claim a Go/runtime gate passed when it did not run. No runtime capability is added by this task.

## Acceptance criteria

- [ ] The target artifact is reachable from the index, has the phase-format sections, number-only Order values and descriptive Increment titles, and has no candidate boundary left for the task author to decide.
- [ ] The design names one vendor CLI and supported version, cites sanitized evidence for launch, login, input/output and native exit behavior, and records one evidence-backed PTY-or-pipes decision under D5 rather than leaving alternatives for derivation.
- [ ] Ordered increments assign the adapter descriptor, launch/input/output and login boundaries, vendor decoding, fake-profile parity, and observable native exit result plus interpretation exactly once; each names a visible result and its Stage 0 prerequisite.
- [ ] Source evidence and planned prerequisites are labelled separately; injecting a missing prerequisite or a premature completion assertion yields an audit finding, not a pass.
- [ ] Every included concern has one owner, excluded roadmap concerns have their source destination, and a real deferred decision retains its explicit trigger.
- [ ] Independent phase audit names the reviewed commit and has no unresolved blocking findings; index and Phase 0 match that reviewed artifact.
- [ ] At least one complete independently authored Phase 0 follow-up brief exists in the tracker and index, assigns the next uncovered phase-2 implementation increment, records its semantic dependency on P0-002, and has matching Phase 0 coverage. A missing follow-up blocks completion.
- [ ] The follow-up specifies repeated concrete handoffs for remaining actionable coverage; unproduced implementation briefs are labelled not yet derived and have no speculative IDs.
- [ ] Existing runtime/source behavior, frozen tracker schema, task status fields, and separate-author handoff remain intact.

## Deferrals

None. Producing the independently authored tracked follow-up is part of completion. Actual implementation-brief derivation is assigned to that real follow-up, while product implementation remains outside this planning task.

## Traces to

`docs/workplan/phase-0/README.md` ordered planning increment 2 (legacy reference 0.2); docs/roadmap.md §3 Stage 1 and §6 D1/D5; docs/contracts.md §8–9; docs/code-structure.md §4; ADR-0003 (Proposed).
