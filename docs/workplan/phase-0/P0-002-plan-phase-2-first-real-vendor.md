# P0-002: Plan phase-2: first real vendor

**Phase:** 0 — Plan the implementation phases
**Shape:** planning
**Dependencies:** None

## Objective

Produce an independently audited `docs/workplan/phase-2/README.md` that decides
the ordered, observable increments for roadmap Stage 1, including the first real
vendor and native-exit boundary, and hand its first increment to exactly one
tracked Phase 2 planning task created with
`skills/phase-increment-handoff/SKILL.md`.

## Scope

- Use `skills/create-phase/SKILL.md` and `skills/shared/phase-design.md` to author
  `docs/workplan/phase-2/README.md`. Start from the attach page, standalone fake
  binary, and absent runner/adapter inventory recorded in `docs/state.md` and
  verify those claims against HEAD. Treat the unfinished Phase 1 increments as
  planned prerequisites, citing `docs/workplan/phase-1/README.md` and their
  coverage rather than presenting Stage 0 exchange as implemented.
- Resolve D1 when Phase 2 opens: name one vendor CLI, cite dated first-party
  documentation and a sanitized, reproducible observation (or explicitly
  labelled source-only evidence) for its launch, login, input/output, exit, and
  non-interactive/terminal behavior. Record why that evidence is sufficient for
  the first adapter without turning the observed vendor dialect into the generic
  product contract.
- Apply D5's trigger to that evidence. If the selected CLI requires TTY
  detection, raw mode, resize, or another terminal semantic, decide the PTY
  boundary and the observable evidence for it; otherwise record that the trigger
  did not occur and keep pipes as the planned Stage 1 requirement. Do not leave
  PTY-versus-pipes for the later task author or implementer to choose.
- Partition Stage 1 from the shipped interaction surface into the smallest
  independently observable increments. Assign one owner to the selected
  vendor's descriptor/capabilities, launch and login, input/output translation,
  same-page exchange, fake-profile parity, and native-exit reporting. Separate
  native exit facts from their stable Code Winch interpretation, and state what
  a person observes for success, non-zero exit, signal termination, and an
  unlaunchable or unauthenticated CLI. Combine concerns only where a smaller
  increment would have no useful observation, and record that rationale.
- Preserve the existing Stage 0 attach contract and malleable fake profile as
  prerequisites to extend, not constraints imposed by the fake's private
  JSON-line dialect. Keep required vendor login at Stage 1's single-user depth;
  route scoped credential injection and enforcement to roadmap Stage 5 and a
  second adapter/workflows to Stage 6.
- After the phase artifact is decided, obtain an independent `audit-phase`
  report against its commit and resolve every blocking artifact finding. Then
  complete every `phase-increment-handoff` action in this task: allocate the next
  Phase 2 ID, create one planning-shape brief for the phase's lowest uncovered
  increment, add its pending tracker entry with the semantic edge on P0-002, and
  reconcile the index plus Phase 0 and Phase 2 coverage without changing any
  existing status field.

## Required constraints

No real adapter, descriptor, codec, native-exit interpretation, or PTY
implementation exists at HEAD. `cmd/winch-sandbox/main.go` composes only the
attach server; `cmd/fake-harness/main.go` is a standalone fixture; and
`internal/adapters/harness/` does not exist. Stage 0 exchange remains a planned
prerequisite, never an implemented baseline claim.

Existing architecture and roadmap specify the wrapper boundary sufficiently to design this artifact; recheck the actual Stage 0 source at dispatch rather than adding an ordering-only dependency.

The phase artifact records outcomes and evidence limits, not implementation
commands, fixture protocols, or file forecasts. Product behavior described by
the phase is predicted. Preserve Phase 0's decided boundary, the frozen tracker
schema, and every existing task status and metadata field.

## Non-goals

- Implementing the adapter, descriptor, codec, PTY, exit mapping, login flow, or
  any other product behavior.
- A second vendor or workflows (Stage 6); product-owned run start/stop lifecycle
  (Stage 4); and stop escalation, orphan recovery, scoped credential injection,
  or egress enforcement (Stage 5).
- Deriving an implementation brief or reserving an implementation ID. The one
  Phase 2 task created here owns derivation of the first decided increment.

Product completion claims based only on a phase design or vendor-source evidence
remain excluded.

## Planning outputs and handoff

The phase author creates `docs/workplan/phase-2/README.md` in the phase format.
It distinguishes the source-supported HEAD baseline from unfinished Phase 1
prerequisites; decides the D1 vendor and D5 disposition; gives every Stage 1
concern one owner; orders independently observable increments; states exit
observations and their evidence limits; and initially marks all implementation
brief coverage **not yet derived**. The workplan index and Phase 0 coverage point
to the same artifact.

A separate phase auditor uses `skills/audit-phase/SKILL.md` against the decided
artifact commit. The task change records the reviewed commit and the
disposition of each finding; unresolved blocking findings prevent completion.

Handoff: complete `skills/phase-increment-handoff/SKILL.md` in this task. It
creates `docs/workplan/phase-2/<allocated-ID>-plan-<increment-slug>.md` for the
lowest-order uncovered increment. The allocated ID is the next Phase 2 ID at
implementation time. Its tracker and brief carry a **semantic dependency on
P0-002** because it consumes the reviewed Phase 2 artifact. It allocates no
implementation ID and, in turn, requires the same handoff after it derives the
increment's implementation brief.

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

After creating the design and its planning task, a reviewer runs from the repository root:

```sh
test -f docs/workplan/phase-2/README.md && sed -n '1,260p' docs/workplan/phase-2/README.md
rg -n 'phase-2|P0-002|Plan Phase 2' docs/workplan/README.md docs/workplan/phase-0/README.md docs/workplan/phase-2/README.md
python3 - <<'PY'
import json
from pathlib import Path

tracker = json.loads(Path("docs/workplan/tasks.json").read_text())
tasks = tracker["tasks"]
handoffs = [task for task in tasks if task["phase"] == 2 and task["depends_on"] == ["P0-002"]]
assert len(handoffs) == 1, handoffs
task = handoffs[0]
brief = Path("docs/workplan") / task["brief"]
assert task["status"] == "pending" and brief.is_file(), task
text = brief.read_text()
assert "**Shape:** planning" in text and "P0-002 (semantic:" in text
print(task["id"], task["title"], task["brief"])
PY
bash scripts/list-available-tasks.sh
```

Expected artifact observations: the phase names one vendor and separates dated
vendor evidence, source-supported HEAD behavior, and planned prerequisites. It
records a decided D5 disposition, retains fake parity, and gives native success,
failure, signal, and launch/login failures observable interpretations. The index
and Phase 0 point to that design and to the same single Phase 2 planning task;
the Python check prints its real ID/title/path and proves its pending semantic
edge and planning brief. Availability reflects tracker dependencies, not a
runtime milestone. The independent audit report identifies the reviewed commit
and has no unresolved blocking finding.

## Verification

- Independent `audit-phase` review checks the baseline/prerequisite distinction,
  D1 evidence, D5 disposition, increment partition, concern ownership, task fit,
  exit observations, and evidence limits; absent prerequisites presented as
  existing behavior are blocking findings.
- Run `python3 -m jsonschema -i docs/workplan/tasks.json
  docs/workplan/tasks.schema.json`; verify every brief path and dependency ID
  resolves; and compare tracker, index, Phase 0, and Phase 2 coverage.
- Run the Demonstration block and `bash scripts/list-available-tasks.sh`; confirm
  every `phase-increment-handoff` action holds, exactly one Phase 2 planning task
  was allocated, no implementation ID was allocated, and existing availability
  semantics did not change except for the new task being blocked on P0-002.
- Run the repository gates required by `docs/testing.md` and `git diff --check`.
  No Go, web, container, or real-vendor runtime pass is evidence for this
  planning task unless that command actually ran; this task adds no runtime
  capability.

## Acceptance criteria

- [ ] The Phase 2 artifact is reachable from the index, has every phase-format
  section, uses number-only `Order` values and descriptive increment titles, and
  leaves no vendor, PTY, increment, ownership, or evidence-boundary candidate
  for the later task author to decide.
- [ ] The design names one first vendor and supplies dated, sanitized evidence
  for launch, login, I/O, exit, and terminal behavior; missing or contradictory
  evidence is visible to the auditor and blocks acceptance.
- [ ] D5 is either triggered with a decided PTY boundary and observable terminal
  evidence, or explicitly not triggered with evidence that pipes satisfy the
  selected CLI; it is not decided from the fake fixture's present behavior.
- [ ] Native success, non-zero exit, signal termination, and launch/login failure
  each have an observable native fact and planned stable interpretation, while
  fake-profile parity remains a shipped Stage 1 outcome.
- [ ] HEAD evidence, source-only vendor evidence, and unfinished Phase 1
  prerequisites are labelled separately; a missing prerequisite or premature
  implementation claim is a blocking audit finding.
- [ ] Every included concern has exactly one increment owner; combined concerns
  explain why separation has no useful observation; and excluded concerns point
  to roadmap Stages 4, 5, or 6 or retain D5's explicit trigger as applicable.
- [ ] Independent phase audit has no unresolved blocking findings; index and Phase 0 match the reviewed artifact.
- [ ] Every `phase-increment-handoff` action holds: exactly one planning-shape
  task in Phase 2 derives the lowest-order uncovered increment, with a semantic
  dependency on P0-002 and matching tracker, brief, index, Phase 0, and Phase 2
  coverage. A missing or duplicate task blocks completion.
- [ ] Unproduced implementation briefs are labelled not yet derived and have no speculative IDs.
- [ ] Existing runtime/source behavior, frozen tracker schema, and task status fields remain intact.

## Deferrals

None. Creating the planning task is part of completion. Implementation-brief derivation belongs to that task, while product implementation remains outside this planning task.

## Traces to

`docs/workplan/phase-0/README.md` ordered planning increment 2 (Plan Phase 2); docs/roadmap.md §3 Stage 1 and §6 D1/D5; docs/contracts.md §8–9; docs/code-structure.md §4; ADR-0003 (Proposed).

