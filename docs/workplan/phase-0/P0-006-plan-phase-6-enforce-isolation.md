# P0-006: Plan phase-6: enforced isolation

**Phase:** 0 — Plan the implementation phases
**Shape:** planning
**Dependencies:** None

## Objective

Produce a reviewed `docs/workplan/phase-6/README.md` with decided ordered increments for roadmap Stage 5, together with a concrete tracked Phase 6 planning follow-up assigning derivation of the next uncovered implementation increment to a separate task author.

## Scope

- Start from the shipped attach surface and partition enforcement of named profiles, disposable workspace/path boundaries, resource/stop/orphan controls, deny-egress/allowlist and credential reference/injection/redaction guarantees into independently observable hardening increments. Do not group controls merely because they share a security owner or implementation substrate; where two guarantees must remain together, record why neither produces a useful refusal by itself.
- Give every `security.md` §11 LB01–LB11 blocker one concern-boundary owner and one phase-exit negative observation. An item implemented in an earlier predicted phase remains a prerequisite to recheck, not evidence available at HEAD; an item excluded from Phase 6 must cite its roadmap destination and must continue to block shared deployment.
- Include sensitivity-class data lifecycle, parser/renderer bounds, identities/revocation, supply-chain provenance and incident exercises where the launch-blocker checklist requires them.
- Keep the current narrow controls reusable but distinguish instance-specific evidence (for example, one container driver's negative tests) from a general profile guarantee. Explicitly gate any posture beyond single-user loopback on audited blocker evidence, including an off-host LB11 reachability scan.

## Required constraints

HEAD source contains only narrow static traversal checks (`internal/adapters/transport/attach/static.go`), loopback publication (`deployments/compose.yml`), a non-root image (`deployments/Dockerfile`), and posture reporting (`internal/adapters/transport/attach/session.go`). `docs/state.md` records that LB01–LB11 are not certified closed and that the control plane, sandbox drivers, identities, credentials, network enforcement, lifecycle policy, and shared-deployment evidence do not exist. Treat those as planned prerequisites rather than current capabilities.

The dependency edge is genuinely empty: `tasks.json` and the workplan index make no pending artifact a prerequisite, and the security and lifecycle design contracts suffice to partition enforcement. Phase 2–5 artifacts and behavior are predicted inputs only. Name the contract or behavior each Phase 6 increment will eventually consume, label it **planned prerequisite**, and require its task author to recheck it at dispatch rather than inventing dependencies on the pending phase-design tasks.

The refinement search `rg -n --glob '*.md' 'P0-006' docs skills` found no incoming Deferrals row. The `docs/state.md` occurrence names a retired Stage 0 task from the closure record and is not work owned by this re-enumerated planning task; do not absorb harness delivery into Phase 6.

Use `skills/create-phase/SKILL.md` as the phase-author operation, with `skills/shared/phase-design.md` and the workplan model. The required artifacts are a decided phase design, an independently authored tracked **Phase 6** follow-up planning brief, consistent tracker/index/Phase 0 and Phase 6 coverage, and an independent `audit-phase` review. Product behavior and commands described there are predictions. Preserve Phase 0's decided boundaries and every existing task status field.

## Non-goals

Second vendor/workflows (Stage 6), hostile public/mutually untrusted container tenancy (security.md §11), and implementation of the controls in this planning task.

The assigned phase author does not derive implementation briefs or author the follow-up planning brief. Product implementation, security exceptions, launch sign-off, and completion claims based only on a design remain excluded. Arranging independent derivation of the tracked follow-up is required for this task to finish; a verbal handoff is insufficient.

## Planning outputs and handoff

The phase author creates `docs/workplan/phase-6/README.md` in the exact `skills/shared/phase-design.md` format. First record the four narrow source-supported controls above and the runtime-verification limits from `docs/state.md`; then list Phase 2–5 capabilities only as planned prerequisites. Build the concern-boundary table before ordering increments so every Stage 5 guarantee and LB01–LB11 has exactly one owner or cited exclusion. Order independently observable hardening increments from the shipped interaction surface, identify parallel branches explicitly, and give each increment a visible attack/refusal outcome, included wiring/operator path, exclusions, and prerequisites. Mark all implementation coverage **not yet derived** and identify exactly one next increment without allocating its implementation ID.

Update the Phase 6 row and partial graph/collision report in `docs/workplan/README.md`, and the Plan Phase 6 coverage row in `docs/workplan/phase-0/README.md`. Do not change product contracts, roadmap meaning, or Phase 0's ordered boundary. If the partition exposes a design contradiction rather than a delivery choice, stop and route it through Phase 0 phase revision instead of resolving it inside this task.

A separate phase auditor uses `skills/audit-phase/SKILL.md` against the decided design and records evidence against the reviewed commit. Resolve blocking artifact findings. Then arrange for a **different task author** to use `skills/derive-task/SKILL.md` to produce one concrete Phase 6 planning follow-up brief assigning derivation of the named next implementation increment. The phase author cannot author that follow-up. This independent derivation is an expected artifact of P0-006's completion, not merely an invitation to do more work.

Allocate the follow-up's next `P6-<NNN>` ID only when its complete brief is actually derived; create its real `docs/workplan/phase-6/<ID>-<slug>.md` file and matching pending tracker entry, index row, Phase 0 handoff record, and Phase 6 coverage together. The follow-up records a **semantic dependency on P0-006** because it consumes the reviewed `docs/workplan/phase-6/README.md` produced here. Name the target increment until a real implementation brief ID exists; do not allocate a placeholder ID for the implementation task. Recompute the actual partial graph and collision report after each addition, preserving all existing statuses.

The follow-up itself assigns a separate task author to derive one implementation brief at a time against the decided phase artifact, reconcile its tracker/index/coverage records, and arrange the next concrete tracked planning follow-up when actionable coverage remains. Repeat that handoff until the phase's actionable increments are covered. Do not bundle all increments into one derivation or depend on product implementation simply to plan its brief. Dispatch this phase-design task through a separate refiner using `skills/refine-task/SKILL.md` against current HEAD.

## Write set

Forecast:

- `docs/workplan/phase-6/README.md` (new phase artifact).
- `docs/workplan/README.md` (phase 6 design row, actual follow-up dependency row, affected coverage and graph/collision report).
- `docs/workplan/phase-0/README.md` (increment order 6, Plan Phase 6, design and follow-up coverage/handoff).
- `docs/workplan/phase-6/<allocated-P6-ID>-<target-increment-slug>.md` (new follow-up planning brief, independently authored after the design is decided; concrete ID/title/path recorded then).
- `docs/workplan/tasks.json` (add the actual follow-up entry and its semantic dependency on P0-006; preserve all existing task statuses and metadata).

No existing `tasks.json` status write is assigned; the new independently derived follow-up begins pending with the model's normal unassigned metadata. A design-set correction discovered to be necessary must be declared with its concrete file and contract before proceeding through Phase 0's phase-revision routing; do not silently broaden this forecast.

## Contract surfaces

None changed in product code or wire schemas. This task decides only phase-6's delivery boundaries; it consumes the cited design contracts. Shared tracker additions, index and Phase 0 coverage edits are write collisions, not product contract collisions. A proposed change to shared product meaning requires Phase 0 disposition and a reconciled graph before assignment.

## Demonstration

After creating the design and independently authored follow-up, a reviewer runs from the repository root:

```sh
sed -n '1,240p' docs/workplan/phase-6/README.md
rg -n 'phase-6|P0-006|0\.6' docs/workplan/README.md docs/workplan/phase-0/README.md
python3 - <<'PY'
import json

tasks = json.load(open('docs/workplan/tasks.json'))['tasks']
followups = [
    task for task in tasks
    if task['phase'] == 6 and task['status'] == 'pending'
    and 'P0-006' in task['depends_on']
]
assert followups, 'missing pending Phase 6 follow-up'
for task in followups:
    assert task['brief'].startswith('phase-6/')
    print(task['id'], task['brief'], task['depends_on'])
PY
bash scripts/list-available-tasks.sh
```

Expected artifact observations: every LB01–LB11 has a concrete concern owner and negative phase-exit observation; unenforced egress or an off-host attach port cannot be represented as safe shared deployment. The index, Phase 0, and Phase 6 locate the same design and actual Phase 6 follow-up brief. The Python check prints its real ID/path and dependency on P0-006. The follow-up names the next uncovered implementation increment while that implementation coverage remains not yet derived. Availability reflects tracker dependencies, not a claimed runtime milestone. Obtain the independent auditor's findings against the reviewed commit; artifact creation alone is not an audit verdict.

## Verification

- Independent `audit-phase` review checks baseline evidence, decided boundaries, task fit, source links and exit observations; reject absent prerequisites presented as existing behavior.
- Validate tracker against its unchanged schema, resolve dependency IDs/brief paths, and compare tracker/index/Phase 0/Phase 6 coverage; run the Demonstration's Python assertion and `bash scripts/list-available-tasks.sh`. Confirm the real Phase 6 follow-up and tracker entry agree, its semantic dependency consumes the producing phase-design artifact, and no placeholder implementation ID was allocated.
- `git diff --check`, completion-script tests (`python3 -m unittest discover -s test`) and `make check` per AGENTS. Report unavailable toolchains explicitly; do not claim a Go/runtime gate passed when it did not run. No runtime capability is added by this task.

## Acceptance criteria

- [ ] The target artifact is reachable from the index, has the phase-format sections, number-only Order values and descriptive Increment titles, and has no candidate boundary left for the task author to decide.
- [ ] Every LB01–LB11 has a concrete concern owner and negative phase-exit observation; unenforced egress or an off-host attach port cannot be represented as safe shared deployment.
- [ ] Source evidence and planned prerequisites are labelled separately; injecting a missing prerequisite or a premature completion assertion yields an audit finding, not a pass.
- [ ] Every included concern has one owner, excluded roadmap concerns have their source destination, and a real deferred decision retains its explicit trigger.
- [ ] Independent phase audit has no unresolved blocking findings; index and Phase 0 match the reviewed artifact.
- [ ] One complete independently authored Phase 6 follow-up brief exists in the tracker and index, assigns the named next uncovered Phase 6 implementation increment, records its semantic dependency on P0-006, and has matching Phase 0 and Phase 6 coverage. A missing follow-up blocks completion.
- [ ] The follow-up specifies repeated concrete handoffs for remaining actionable coverage; unproduced implementation briefs are labelled not yet derived and have no speculative IDs.
- [ ] Existing runtime/source behavior, frozen tracker schema, task status fields, and separate-author handoff remain intact.

## Deferrals

None. Producing the independently authored tracked follow-up is part of completion. Actual implementation-brief derivation is assigned to that real follow-up, while product implementation remains outside this planning task.

## Traces to

`docs/workplan/phase-0/README.md` ordered planning increment 6 (Plan Phase 6); docs/roadmap.md §3 Stage 5 and §5; docs/security.md §2–12, especially §11 LB01–LB11; docs/contracts.md §1–5/8.
