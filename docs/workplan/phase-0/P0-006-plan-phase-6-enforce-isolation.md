# P0-006: Plan phase-6: enforced isolation

**Phase:** 0 — Plan the implementation phases
**Shape:** planning
**Dependencies:** None

## Objective

Produce a reviewed `docs/workplan/phase-6/README.md` with decided ordered increments for roadmap Stage 5, together with a concrete tracked Phase 0 follow-up assigning derivation of the next uncovered implementation increment to a separate task author.

## Scope

- Begin `docs/workplan/phase-6/README.md` from the source-supported baseline at
  HEAD: `GET /api/session` reports `container-standard` with
  `network-egress` unenforced, static serving rejects resolved-path and symlink
  escapes, Compose publishes the attach surface on host loopback, and the image
  declares a non-root user. Cite the exact sources named in `docs/state.md`
  §*Sandbox page, posture, and operator status*, including its image and Compose
  evidence. Label all four as narrow source evidence, not runtime verification
  or closure of a launch blocker. The phase artifact does not yet exist at HEAD;
  that absence is the expected pre-task observation.
- Partition roadmap Stage 5 into ordered, independently demonstrable hardening
  increments. Give one increment ownership of each included concern: shared
  authentication/object authorization and request/stream defenses; profile
  enforcement and truthful posture; disposable workspace plus traversal and
  extraction defenses; resource limits, forced stop, crash recovery and orphan
  cleanup; deny-by-default egress and allowlisting; opaque credential references,
  scoped temporary injection and pre-persistence redaction; bounded parsers and
  safe renderers; sensitivity-class retention/export/deletion/cache/backup
  lifecycle; image/dependency provenance and scanning; incident response and
  identity rotation/revocation/recovery; and host-local attach reachability.
  Split concerns when each has a useful negative observation; combine them only
  when the artifact records why no independently observable intermediate state
  is safe or useful.
- Add an explicit LB01–LB11 accountability matrix to the phase artifact. For
  every launch blocker, name exactly one owning increment and the attack or
  failure attempted at phase exit. The observation must show refusal or safe
  cleanup at the real boundary (including the deployment network for LB04 and
  an off-host vantage point for LB11), rather than treating configuration,
  contract tests, or an in-process fake as generalized enforcement evidence.
- Make shared-deployment eligibility a phase-exit conjunction: all LB01–LB11
  evidence is recorded and release/security sign-off or a compliant,
  time-bounded exception exists as required by `docs/security.md` §11. The
  phase artifact must keep the posture at single-user loopback while any blocker
  is open and must not present `local-trusted` as isolated.
- Record concern ownership, prerequisites, exclusions, observable changes, task
  coverage, and the next uncovered increment using the exact phase format in
  `skills/shared/phase-design.md`. Earlier roadmap stages are planned
  prerequisites only: recheck the required control-plane, lifecycle, adapter,
  record, and renderer surfaces against HEAD when an implementation brief is
  dispatched; do not claim that a planning artifact supplies them.

## Required constraints

Only narrow static traversal checks, loopback publication source, non-root image
source and posture reporting exist; LB01–LB11 are not certified closed
(`docs/state.md` §*Shared-deployment security and data lifecycle*). Docker and
the runtime checks were unavailable at the recorded closure, so source
configuration is not observed isolation.

Existing security and lifecycle design contracts suffice to partition enforcement independently; earlier implementation phases are predicted baselines and must be rechecked at dispatch. No pending artifact is required to author this bounded security design.

Use `skills/create-phase/SKILL.md` as the phase-author operation, with
`skills/shared/phase-design.md` and the workplan model. The required behavior is
a decided phase artifact, an independently authored tracked follow-up planning
brief, consistent tracker/index/Phase 0 coverage, and an independent
`audit-phase` review. Product behavior and commands described there are
predictions. Preserve Phase 0's decided boundaries and every existing tracker
status. The active tracker has no dependency for this task; the `P0-006` row in
`docs/state.md` is explicitly a retired task from the closed prior plan and is
neither a deferral nor evidence for this task.

## Non-goals

- Implementing, configuring, or certifying any security control. This planning
  task changes workplan artifacts only and does not make the current deployment
  safe to share.
- Second-vendor/workflow behavior (roadmap Stage 6), or treating the container
  driver as a boundary for hostile public or mutually untrusted tenants. The
  latter remains prohibited by `docs/security.md` §11 until the stronger-driver
  launch condition there is satisfied; this task does not invent an owner or a
  speculative implementation ID for it.
- Revising product contracts, security policy, or roadmap stage boundaries. A
  contradiction discovered while authoring the phase returns through Phase 0
  revision rather than being silently decided in this artifact.

The assigned phase author does not derive implementation briefs or author the follow-up planning brief. Product implementation and completion claims based only on a design remain excluded. Arranging independent derivation of the tracked follow-up is required for this task to finish; a verbal handoff is insufficient.

## Planning outputs and handoff

The phase author creates `docs/workplan/phase-6/README.md`, records the concrete
HEAD baseline above separately from planned prerequisites, chooses ordered
increments and one owner for every concern and LB row, states negative exit
observations and their evidence limits, and marks every implementation task as
**not yet derived**. The artifact identifies one next uncovered increment by
title without allocating an ID. In the same phase-author change, update the
phase-6 row in `docs/workplan/README.md` and the 0.6 coverage row in
`docs/workplan/phase-0/README.md`; do not derive a brief in that operation.

A separate phase auditor uses `skills/audit-phase/SKILL.md` against the decided
design and records a verdict against the reviewed commit. Resolve every Blocking
finding before handoff. Then a **different task author** uses
`skills/derive-task/SKILL.md` to produce one concrete Phase 0 planning follow-up
brief assigning derivation of the named next uncovered phase-6 increment. The
phase author cannot author that follow-up, and the auditor cannot substitute an
artifact review with a claim that predicted controls work. This independent
derivation is an expected artifact of P0-006's completion, not merely an
invitation to do more work.

Allocate the follow-up's next plan-wide ID only when its complete brief is actually derived; create its real file and matching pending tracker entry, index row and Phase 0 coverage together. The follow-up records a **semantic dependency on P0-006** because it consumes the reviewed `docs/workplan/phase-6/README.md` produced here. Name the target increment until a real implementation brief ID exists; do not allocate a placeholder ID for the implementation task. Recompute the actual partial graph and collision report after each addition, preserving all existing statuses.

The follow-up itself assigns a separate task author to derive one implementation
brief at a time against the decided phase artifact, reconcile its
tracker/index/coverage records, and arrange the next concrete tracked planning
follow-up when actionable coverage remains. Repeat that handoff until the
phase's actionable increments are covered. Do not bundle all increments into
one derivation or depend on product implementation simply to plan its brief.

## Write set

Forecast:

- `docs/workplan/phase-6/README.md` (new phase artifact).
- `docs/workplan/README.md` (phase 6 design row, actual follow-up dependency row, affected coverage and graph/collision report).
- `docs/workplan/phase-0/README.md` (increment order 6, legacy reference 0.6, design and follow-up coverage/handoff).
- `docs/workplan/phase-0/<allocated-ID>-<target-increment-slug>.md` (new follow-up planning brief, independently authored after the design is decided; concrete ID/title/path recorded then).
- `docs/workplan/tasks.json` (add the actual follow-up entry and its semantic dependency on P0-006; preserve all existing task statuses and metadata).

No existing `tasks.json` status write is assigned; the new independently derived follow-up begins pending with the model's normal unassigned metadata. A design-set correction discovered to be necessary must be declared with its concrete file and contract before proceeding through Phase 0's phase-revision routing; do not silently broaden this forecast.

## Contract surfaces

None changed in product code or wire schemas. This task decides only phase-6's delivery boundaries; it consumes the cited design contracts. Shared tracker additions, index and Phase 0 coverage edits are write collisions, not product contract collisions. A proposed change to shared product meaning requires Phase 0 disposition and a reconciled graph before assignment.

## Demonstration

After creating the design and independently authored follow-up, a reviewer runs from the repository root:

```sh
test -f docs/workplan/phase-6/README.md
sed -n '1,240p' docs/workplan/phase-6/README.md
rg -n '^## (Outcome and starting point|Concern boundaries|Ordered increments|Task coverage|Phase exit)$|LB(0[1-9]|1[01])' docs/workplan/phase-6/README.md
rg -n 'phase-6|P0-006|0\.6' docs/workplan/README.md docs/workplan/phase-0/README.md
python3 - <<'PY'
import json
from pathlib import Path

tracker = json.loads(Path("docs/workplan/tasks.json").read_text())
tasks = {task["id"]: task for task in tracker["tasks"]}
assert tasks["P0-006"]["status"] == "in_progress"
followups = [task for task in tracker["tasks"] if task["depends_on"] == ["P0-006"]]
assert len(followups) == 1, followups
followup = followups[0]
brief = Path("docs/workplan") / followup["brief"]
assert brief.is_file(), brief
assert "**Dependencies:** P0-006 (semantic:" in brief.read_text()
print(followup["id"], brief)
PY
bash scripts/list-available-tasks.sh
```

Before the task, `test`/`sed` fail because `phase-6/README.md` does not exist;
the index and Phase 0 say it is not authored, and the availability command omits
this in-progress task. After the task, the artifact has all five required phase
sections; every LB01–LB11 has one concern owner and a negative phase-exit
observation; and unenforced egress or an off-host attach port prevents shared
deployment eligibility. Both indexes locate the same design and actual
follow-up. The Python check proves the unchanged task status, exactly one direct
follow-up, a resolving brief path, and the declared semantic dependency in that
brief without being confused by the unrelated retired-ID prose in
`docs/state.md`. Inspect the printed follow-up brief to confirm it names the next
uncovered increment and leaves its implementation brief not yet derived.
Availability reflects tracker dependencies, not a runtime milestone. Obtain the
independent auditor's verdict against the reviewed commit; artifact creation
alone is not an audit pass.

## Verification

- Independent `audit-phase` review checks baseline evidence, decided boundaries, task fit, source links and exit observations; reject absent prerequisites presented as existing behavior.
- Run the exact Demonstration block. Validate the tracker against
  `docs/workplan/tasks.schema.json`, resolve every dependency ID and brief path,
  and compare tracker/index/Phase 0/Phase 6 coverage. The availability output
  must follow the reconciled graph. Confirm exactly one direct follow-up from
  P0-006, its semantic dependency consumes the reviewed phase design, and no
  placeholder implementation ID was allocated.
- `git diff --check`, completion-script tests (`python3 -m unittest discover -s test`) and `make check` per AGENTS. Report unavailable toolchains explicitly; do not claim a Go/runtime gate passed when it did not run. No runtime capability is added by this task.

## Acceptance criteria

- [ ] The target artifact is reachable from the index, has the phase-format sections, number-only Order values and descriptive Increment titles, and has no candidate boundary left for the task author to decide.
- [ ] Every LB01–LB11 appears exactly once in the phase accountability matrix
      with one owning increment, an attack/failure observation at the applicable
      boundary, and the limit of that evidence; unenforced egress, an off-host
      attach port, or any other open blocker prevents shared-deployment
      eligibility.
- [ ] Source evidence and planned prerequisites are labelled separately; injecting a missing prerequisite or a premature completion assertion yields an audit finding, not a pass.
- [ ] Every included concern in Scope has one increment owner; Stage 6 concerns
      and hostile multi-tenant container use retain their cited destinations,
      and no new ownerless deferral or speculative ID is introduced.
- [ ] Independent phase audit has no unresolved blocking findings; index and Phase 0 match the reviewed artifact.
- [ ] At least one complete independently authored Phase 0 follow-up brief exists in the tracker and index, assigns the next uncovered phase-6 implementation increment, records its semantic dependency on P0-006, and has matching Phase 0 coverage. A missing follow-up blocks completion.
- [ ] The follow-up specifies repeated concrete handoffs for remaining actionable coverage; unproduced implementation briefs are labelled not yet derived and have no speculative IDs.
- [ ] Existing runtime/source behavior, frozen tracker schema, task status fields, and separate-author handoff remain intact.

## Deferrals

None. Producing the independently authored tracked follow-up is part of completion. Actual implementation-brief derivation is assigned to that real follow-up, while product implementation remains outside this planning task.

## Traces to

`docs/workplan/phase-0/README.md` ordered planning increment 6 (legacy reference 0.6); docs/roadmap.md §3 Stage 5 and §5; docs/security.md §2–12, especially §11 LB01–LB11; docs/contracts.md §1–5/8.
