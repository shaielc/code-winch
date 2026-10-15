# P0-002: Plan phase-2: first real vendor

**Phase:** 0 — Plan the implementation phases
**Shape:** planning
**Dependencies:** None

## Objective

Produce a reviewed `docs/workplan/phase-2/README.md` with decided ordered increments for roadmap Stage 1, together with a concrete tracked Phase 2 planning follow-up assigning derivation of the next uncovered implementation increment to a separate task author.

## Scope

- Decide the Stage 1 first-vendor choice under D1 when this phase opens. Record dated, sanitized evidence from the vendor's supported documentation and a disposable local probe of the installed CLI (version, non-interactive invocation, stdin/stdout/stderr behavior, exit status, login prerequisite, and whether operation refuses a pipe or requires a TTY); do not record credentials, prompts containing secrets, or user content. Decide D5 only if that evidence reaches its terminal-semantics trigger.
- Partition one genuine adapter seam behind the existing attach wrapper/page: capability description, process launch, command input, raw output decoding, and independently observable native exit reporting. Keep the vendor command, flags, environment and native decoder inside its adapter; keep generic launch/input/output and stable exit meaning vendor-neutral.
- Retain fake-profile parity as a shipped configuration; plan required vendor login at the single-user stage depth without claiming Stage 5 scoped-credential enforcement.

## Required constraints

At refinement HEAD, `cmd/winch-sandbox/main.go` constructs only the attach HTTP handler, `internal/adapters/transport/attach/server.go` exposes health/session/static routes, and `cmd/winch/main.go` accepts only `status`. `cmd/fake-harness/main.go` is a standalone JSON-lines fixture that accepts pipes; it is shipped by the image but is not launched by the sandbox. There is no `internal/runner` or real adapter, descriptor, native codec/exit interpretation, or PTY implementation (`git ls-files cmd internal test/e2e deployments web/src`). These are source observations, not runtime verification. The older closure record in `docs/state.md` also says Stage 0 exchange was incomplete; since the active tracker now contains later Phase 1 work, inspect HEAD and record completed reusable slices plus missing behavior in the new phase artifact rather than treating the closure record as current implementation proof.

Existing architecture and roadmap specify the wrapper boundary sufficiently to design this artifact; recheck the actual Stage 0 source at dispatch rather than adding an ordering-only dependency.

Use `skills/create-phase/SKILL.md` as the phase-author operation, with `skills/shared/phase-design.md` and the workplan model. The required behavior is a decided phase artifact, an independently authored tracked follow-up planning brief in Phase 2, consistent tracker/index/Phase 0 coverage, and an independent `audit-phase` review. Product behavior and commands described there are predictions. Preserve Phase 0's decided boundaries and all existing task status fields.

## Non-goals

Second vendor/workflows (Stage 6), run start/stop lifecycle (Stage 4), stop escalation and orphan recovery (Stage 5).

The assigned phase author does not derive implementation briefs or author the follow-up planning brief. Product implementation, vendor installation, account mutation, credential capture, and completion claims based only on a design remain excluded. Arranging independent derivation of the tracked follow-up is required for this task to finish; a verbal handoff is insufficient.

## Planning outputs and handoff

The phase author creates `docs/workplan/phase-2/README.md` in the exact section/table form required by `skills/shared/phase-design.md`. First record the source-supported HEAD baseline and runtime-verification limits; then record missing Stage 0 behavior as planned prerequisites rather than silently implementing it. Select one vendor using the evidence fields in Scope, decide whether D5 fired, and choose ordered increments with one owner per concern. At minimum, the partition must distinguish truthful availability/login presentation, vendor-neutral capability and launch boundaries, vendor-specific input/output adaptation, native-exit observation and stable interpretation, and fake/real profile parity unless the artifact explains why any adjacent split has no independently useful observation. Each increment states its before/after observation, minimum operator path, prerequisites, exclusions, and limits of proof. Mark all implementation task coverage **not yet derived**, identify exactly one next increment by title, and update the Phase 2 row in the index plus Phase 0 coverage. Do not prescribe or implement product protocol or substrate code in this planning task.

A separate phase auditor uses `skills/audit-phase/SKILL.md` against the decided design. Resolve blocking artifact findings in the phase-design commit. Then arrange for a **different task author** to use `skills/derive-task/SKILL.md` to produce one concrete Phase 2 planning follow-up brief assigning derivation of the named next uncovered implementation increment. The phase author cannot author that follow-up, and the follow-up author cannot derive the implementation brief in the same operation. This independent derivation is an expected artifact of P0-002's completion, not merely an invitation to do more work.

Allocate the follow-up's next Phase 2 ID only when its complete brief is actually derived; create its real file under `docs/workplan/phase-2/` and matching pending tracker entry, index row and Phase 0 coverage together. The follow-up records a **semantic dependency on P0-002** because it consumes the reviewed `docs/workplan/phase-2/README.md` produced here. Name the target increment until a real implementation brief ID exists; do not allocate a placeholder ID for the implementation task. Recompute the actual partial graph, critical path/layers/width, and write/contract collision report from `tasks.json` after the addition, preserving all existing statuses.

The follow-up itself assigns a separate task author to derive one implementation brief at a time against the decided phase artifact, reconcile its tracker/index/coverage records, and arrange the next concrete tracked Phase 2 planning follow-up when actionable coverage remains. Repeat that handoff until the phase's actionable increments are covered. Do not bundle all increments into one derivation or depend on product implementation simply to plan its brief. This refined brief is the required separate-refiner artifact; implementation begins only from its committed task branch.

## Write set

Forecast:

- `docs/workplan/phase-2/README.md` (new phase artifact).
- `docs/workplan/README.md` (phase 2 design row, actual follow-up dependency row, affected coverage and graph/collision report).
- `docs/workplan/phase-0/README.md` (increment order 2, Plan Phase 2, design and follow-up coverage/handoff).
- `docs/workplan/phase-2/<allocated-ID>-plan-<target-increment-slug>.md` (new Phase 2 follow-up planning brief, independently authored after the design is decided; concrete ID/title/path recorded then).
- `docs/workplan/tasks.json` (add the actual follow-up entry and its semantic dependency on P0-002; preserve all existing task statuses and metadata).

No existing `tasks.json` status write is assigned; the new independently derived follow-up begins pending with the model's normal unassigned metadata. A design-set correction discovered to be necessary must be declared with its concrete file and contract before proceeding through Phase 0's phase-revision routing; do not silently broaden this forecast.

## Contract surfaces

None changed in product code or wire schemas. This task decides only phase-2's delivery boundaries; it consumes the cited design contracts. Shared tracker additions, index and Phase 0 coverage edits are write collisions, not product contract collisions. A proposed change to shared product meaning requires Phase 0 disposition and a reconciled graph before assignment.

## Demonstration

Before the change, `docs/workplan/phase-2/README.md` is absent, the two existing index/Phase 0 hits say Phase 2 is scheduled or not yet authored, and `bash scripts/list-available-tasks.sh` lists `P0-006` and `P1-004` but not in-progress `P0-002`. After creating the design and independently authored follow-up, a reviewer runs from the repository root:

```sh
sed -n '1,240p' docs/workplan/phase-2/README.md
rg -n 'phase-2|P0-002|0\.2' docs/workplan/README.md docs/workplan/phase-0/README.md
python3 - <<'PYTHON'
import json
from pathlib import Path

tracker = json.loads(Path("docs/workplan/tasks.json").read_text())
followups = [task for task in tracker["tasks"] if task["phase"] == 2 and task["depends_on"] == ["P0-002"]]
assert len(followups) >= 1
for task in followups:
    assert (Path("docs/workplan") / task["brief"]).is_file()
print([(task["id"], task["brief"]) for task in followups])
PYTHON
bash scripts/list-available-tasks.sh
```

Expected artifact observations: The design identifies one first vendor, evidence for any PTY requirement, and an observable native exit result plus interpretation while retaining fake parity. Both index and Phase 0 locate the same design and an actual follow-up brief. Inspect its tracker entry and brief: they agree on its real ID/path and semantic dependency on P0-002, name the next uncovered implementation increment, and preserve pending implementation-brief coverage. Availability reflects the tracker dependencies, not a claimed runtime milestone. Obtain the independent auditor's findings against the reviewed commit; artifact creation alone is not an audit verdict.

## Verification

- Independent `audit-phase` review checks baseline evidence, decided boundaries, task fit, source links and exit observations; reject absent prerequisites presented as existing behavior.
- Validate the tracker against `docs/workplan/tasks.schema.json` using the repository's available validator or an explicit equivalent check of required keys, IDs, dependency references and brief paths. Compare tracker/index/Phase 0 coverage; `bash scripts/list-available-tasks.sh` must preserve availability semantics. Confirm at least one real Phase 2 follow-up planning brief and tracker entry exist, their semantic dependency consumes the producing phase-design artifact, and no placeholder implementation ID was allocated.
- `git diff --check`, completion-script tests (`python3 -m unittest discover -s test`) and `make check` per AGENTS. Report unavailable toolchains explicitly; do not claim a Go/runtime gate passed when it did not run. No runtime capability is added by this task.

## Acceptance criteria

- [ ] The target artifact is reachable from the Phase 2 index row, has the phase-format sections, number-only Order values and descriptive Increment titles, identifies exactly one next increment, and leaves no vendor or candidate boundary for the task author to decide.
- [ ] The design names one first vendor and cites dated, sanitized supported-documentation evidence plus the result or explicit environment limitation of the disposable CLI probe. It records whether D5 fired and why; if it fired, PTY ownership and an observable terminal-specific check are assigned, and if not, PTY work remains deferred to D5's roadmap trigger.
- [ ] The ordered increments assign the descriptor, launch, input, decoding, native-exit observation/stable interpretation, login presentation, and fake/real parity concerns exactly once. Exit observations distinguish native exit facts from Code Winch's interpretation and state what the single tested vendor does not prove about a second adapter.
- [ ] Source evidence and planned prerequisites are labelled separately; injecting a missing prerequisite or a premature completion assertion yields an audit finding, not a pass.
- [ ] Every included concern has one owner, excluded roadmap concerns have their source destination, and a real deferred decision retains its explicit trigger.
- [ ] Independent phase audit has no unresolved blocking findings; index and Phase 0 match the reviewed artifact.
- [ ] At least one complete independently authored Phase 2 planning follow-up brief exists in the tracker and index, assigns the next uncovered phase-2 implementation increment, records its semantic dependency on P0-002, and has matching Phase 0 coverage. A missing follow-up blocks completion.
- [ ] The follow-up specifies repeated concrete handoffs for remaining actionable coverage; unproduced implementation briefs are labelled not yet derived and have no speculative IDs.
- [ ] Existing runtime/source behavior, frozen tracker schema, task status fields, and separate-author handoff remain intact.

## Deferrals

None. Producing the independently authored tracked Phase 2 follow-up is part of completion. Actual implementation-brief derivation is assigned to that real follow-up, while product implementation remains outside this planning task. PTY work may be excluded only by retaining roadmap D5 and its existing trigger; it is not an ownerless deferral created here.

## Traces to

`docs/workplan/phase-0/README.md` ordered planning increment 2 (Plan Phase 2); docs/roadmap.md §3 Stage 1 and §6 D1/D5; docs/contracts.md §8–9; docs/code-structure.md §4; ADR-0003 (Proposed).
