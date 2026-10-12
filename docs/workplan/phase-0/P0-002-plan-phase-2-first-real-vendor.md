# P0-002: Plan phase-2: first real vendor

**Phase:** 0 — Plan the implementation phases
**Shape:** planning
**Dependencies:** None

## Objective

Produce a reviewed `docs/workplan/phase-2/README.md` with decided ordered increments for roadmap Stage 1, together with a concrete tracked Phase 0 follow-up assigning derivation of the next uncovered implementation increment to a separate task author.

## Scope

- Create the phase design with `skills/create-phase/SKILL.md`, using the current
  sandbox composition root in `cmd/winch-sandbox/main.go`, attach transport in
  `internal/adapters/transport/attach/`, fake executable in
  `cmd/fake-harness/main.go`, deployment files, and standing e2e scenario as the
  source baseline. Record explicitly that the composition root currently wires
  only the HTTP attach server, the fake is shipped but not started, and no
  harness adapter package, runner, integrated fake profile, or phase-2 artifact
  exists. These are observed facts, not Stage 1 prerequisites supplied by this
  task.
- Resolve roadmap D1 in the phase artifact. Compare currently supported,
  non-preview CLI behavior for plausible vendors using dated primary vendor
  documentation and a sanitized local probe when a CLI can be installed or is
  already present. Record the chosen CLI and version/range, invocation and
  login modes, stdin/stdout/stderr behavior, structured-output support, exit
  behavior, and why the evidence supports the choice; record unavailable probes
  as verification limits rather than inventing results. Do not place tokens,
  account names, home-directory paths, prompt/output content, or other secrets
  in the repository.
- Apply D5's trigger to the chosen CLI evidence. If normal operation requires
  TTY detection, raw mode, resize, or another terminal semantic, decide the PTY
  boundary and assign it to an ordered increment. Otherwise state that pipes
  suffice for the evidenced mode and preserve D5 for its existing trigger; do
  not add a speculative PTY abstraction.
- Partition Stage 1 into ordered, demonstrable increments that first establish
  one explicit harness-adapter seam and keep a reachable fake runtime profile,
  then swap the selected real CLI behind the same wrapper and attach page. The
  design must assign capability descriptor, configuration validation,
  launch-spec construction, incremental native decoding across arbitrary byte
  chunks, input/output mapping, login/operator setup, explicit composition-root
  registration, and native exit result plus harness-specific interpretation.
  Every introduced seam must name the fake implementation and runtime wiring
  that makes it reachable in the same increment; no descriptor, codec, or real
  adapter may be planned as unwired library code.
- Define phase exit around the standing Stage 0 exchange scenario: it passes
  through the shipped fake profile and, with separately supplied vendor login,
  through the real profile; the existing loopback page and maintained operator
  path expose the exchange and native-exit observation. State exactly what the
  fake proves and does not prove. Keep login at the single-user Stage 1 depth;
  do not claim Stage 5 credential scoping, secret isolation, or egress control.

## Required constraints

No real adapter, descriptor, codec, native exit interpretation, or PTY
implementation exists at the refinement baseline. `git ls-files cmd internal
test/e2e deployments web/src` shows only the fake executable, HTTP sandbox and
attach transport, status CLI, deployment files, attach UI, and startup/posture
scenario. `cmd/winch-sandbox/main.go` constructs only `attach.New`; the fake's
package comment says it is not started and no integrated fake adapter/profile
exists. `docs/state.md` also records that Stage 0 runner, retained output, and
server-input delivery are absent. Stage 0 exchange is therefore a **planned
prerequisite** to validate against HEAD when the phase is dispatched, never a
capability this planning task may mark implemented.

Existing architecture and roadmap specify the wrapper boundary sufficiently to design this artifact; recheck the actual Stage 0 source at dispatch rather than adding an ordering-only dependency.

Use `skills/create-phase/SKILL.md` as the phase-author operation, with
`skills/shared/phase-design.md` and the workplan model. The artifact must
separate **observed at HEAD**, **planned prerequisite from Phase 1/Stage 0**, and
**predicted Phase 2 result** in every baseline and demonstration claim. The
required result is a decided phase artifact, an independently authored tracked
follow-up planning brief, consistent tracker/index/Phase 0 coverage, and an
independent `audit-phase` review. Preserve Phase 0's decided boundaries and all
existing task status fields.

## Non-goals

Second vendor/workflows (roadmap Stage 6), run start/stop lifecycle (Stage 4),
stop escalation and orphan recovery (Stage 5), Stage 2 structured views and
ordinal resume, and implementing the missing Stage 0 exchange. Do not redesign
the attach API, introduce control-plane runs/canonical sequence numbers, select
a durable store, or promise network/credential isolation.

The assigned phase author does not derive implementation briefs or author the follow-up planning brief. Product implementation and completion claims based only on a design remain excluded. Arranging independent derivation of the tracked follow-up is required for this task to finish; a verbal handoff is insufficient.

## Planning outputs and handoff

The phase author creates `docs/workplan/phase-2/README.md` in the exact format
required by `skills/shared/phase-design.md`: outcome and starting point, concern
boundaries, ordered increments with number-only `Order` cells and descriptive
titles, task coverage, and phase exit. Before choosing increments, inventory the
current composition root, attach transport, fake, configuration, deployment,
CLI, web surface, and e2e scenario; then resolve D1 and conditionally D5 from
the evidence described in Scope. For every increment, identify its shape,
observable result, prerequisites, owner for each included concern, exclusions
with roadmap destinations, write/contract collision implications, fake parity,
operator demonstration, and verification. Do not allow a later increment merely
to wire or expose an earlier seam.

Update the workplan index's phase-2 design/coverage row, the P0-002 dependency
row if its artifact description changes, and the graph/collision report from
the actual newly derived follow-up. Update Phase 0's 0.2 coverage to point to
the reviewed design and real follow-up. Product behavior and commands in the
new design remain predictions until their implementation tasks land.

A separate phase auditor uses `skills/audit-phase/SKILL.md` against the decided design. Resolve blocking artifact findings. Then arrange for a **different task author** to use `skills/derive-task/SKILL.md` to produce at least one concrete Phase 0 planning follow-up brief assigning derivation of the next uncovered implementation increment in phase-2. The phase author cannot author that follow-up. This independent derivation is an expected artifact of P0-002's completion, not merely an invitation to do more work.

Allocate the follow-up's next plan-wide ID only when its complete brief is actually derived; create its real file and matching pending tracker entry, index row and Phase 0 coverage together. The follow-up records a **semantic dependency on P0-002** because it consumes the reviewed `docs/workplan/phase-2/README.md` produced here. Name the target increment until a real implementation brief ID exists; do not allocate a placeholder ID for the implementation task. Recompute the actual partial graph and collision report after each addition, preserving all existing statuses.

The follow-up itself assigns a separate task author to derive one implementation
brief at a time against the decided phase artifact, reconcile its
tracker/index/coverage records, and arrange the next concrete tracked planning
follow-up when actionable coverage remains. Repeat that handoff until the
phase's actionable increments are covered. Do not bundle all increments into
one derivation or depend on product implementation simply to plan its brief.

## Write set

Forecast:

- `docs/workplan/phase-2/README.md` (new phase artifact).
- `docs/workplan/README.md` (phase 2 design/coverage row, actual follow-up dependency row, and recomputed graph/collision report).
- `docs/workplan/phase-0/README.md` (increment order 2 / legacy reference 0.2 design and follow-up coverage/handoff).
- `docs/workplan/phase-0/<allocated-ID>-<target-increment-slug>.md` (new follow-up planning brief, independently authored after the design is decided; concrete ID/title/path recorded then).
- `docs/workplan/tasks.json` (add the actual follow-up entry and its semantic dependency on P0-002; preserve all existing task statuses and metadata).

No existing `tasks.json` status write is assigned; the new independently derived follow-up begins pending with the model's normal unassigned metadata. A design-set correction discovered to be necessary must be declared with its concrete file and contract before proceeding through Phase 0's phase-revision routing; do not silently broaden this forecast.

## Contract surfaces

None changed in product code or wire schemas. This task decides only phase-2's delivery boundaries; it consumes the cited design contracts. Shared tracker additions, index and Phase 0 coverage edits are write collisions, not product contract collisions. A proposed change to shared product meaning requires Phase 0 disposition and a reconciled graph before assignment.

## Demonstration

After creating the design and independently authored follow-up, a reviewer runs from the repository root:

```sh
test -f docs/workplan/phase-2/README.md
sed -n '1,240p' docs/workplan/phase-2/README.md
rg -n '^\| 2 \||phase-2/README\.md|P0-002|0\.2' docs/workplan/README.md docs/workplan/phase-0/README.md
python3 -m json.tool docs/workplan/tasks.json >/dev/null
bash scripts/list-available-tasks.sh
```

Expected artifact observations: the file exists and identifies exactly one first
vendor, its dated sanitized evidence, the D5 disposition, ordered ownership of
an observable native exit result and interpretation, and fake/real parity. The
baseline labels missing Stage 0 exchange separately from predicted Stage 1
results. Both indexes locate the same design and an actual follow-up brief.
Inspect that follow-up's tracker entry and brief: ID, title, path and semantic
dependency on P0-002 agree; it names the next uncovered implementation
increment; no implementation placeholder ID exists; prior statuses are
unchanged. Availability reflects tracker dependencies, not a claimed runtime
milestone. Obtain the independent auditor's findings against the reviewed
commit; artifact creation alone is not an audit verdict.

## Verification

- Independent `audit-phase` review checks the source inventory, D1 evidence and
  decision, D5 trigger/disposition, ordered boundaries, fake parity, concern
  ownership, source links, and exit observations; reject absent prerequisites
  presented as existing behavior, an unwired seam, or vendor claims unsupported
  by dated primary evidence.
- Validate `docs/workplan/tasks.json` against its unchanged
  `docs/workplan/tasks.schema.json`, resolve every dependency ID and brief path,
  and compare tracker/index/Phase 0 coverage. `bash
  scripts/list-available-tasks.sh` must preserve availability semantics. Confirm
  at least one real follow-up planning brief and tracker entry exist, their
  semantic dependency consumes the producing phase-design artifact, no
  placeholder implementation ID was allocated, and statuses of existing tasks
  are unchanged from the task branch.
- `git diff --check`, completion-script tests (`python3 -m unittest discover -s test`) and `make check` per AGENTS. Report unavailable toolchains explicitly; do not claim a Go/runtime gate passed when it did not run. No runtime capability is added by this task.

## Acceptance criteria

- [ ] The target artifact is reachable from both planning indexes, has every
  phase-format section, number-only `Order` values and descriptive `Increment`
  titles, and leaves no adapter, wiring, parity, exit-reporting, login, or
  operator-surface boundary for a later task author to invent.
- [ ] The design selects exactly one first vendor under D1 and cites dated,
  sanitized primary evidence for its supported invocation, login, I/O,
  structured-output and exit behavior; unavailable local probes are recorded as
  limits, not successful observations.
- [ ] D5 is either decided from evidence that the selected mode requires
  terminal semantics, with PTY ownership assigned, or explicitly remains
  deferred under its existing trigger because the evidenced mode works over
  pipes; the design includes no speculative PTY seam.
- [ ] Ordered increments assign descriptor, validation/launch, chunk-tolerant
  codec and mappings, input/output, login, explicit runtime registration,
  native exit result and harness-specific interpretation, fake parity, and a
  maintained hand path exactly once, with no unwired intermediate result.
- [ ] Source evidence, missing Stage 0 prerequisites, and predicted Stage 1
  results are labelled separately; injecting a missing prerequisite or a
  premature completion assertion yields an audit finding, not a pass.
- [ ] Every included concern has one owner; Stage 2 views/resume and roadmap
  Stages 4–6 concerns retain their stated destinations; and any still-open D5
  decision retains its exact roadmap trigger.
- [ ] Independent phase audit has no unresolved blocking findings; index and Phase 0 match the reviewed artifact.
- [ ] At least one complete independently authored Phase 0 follow-up brief exists in the tracker and index, assigns the next uncovered phase-2 implementation increment, records its semantic dependency on P0-002, and has matching Phase 0 coverage. A missing follow-up blocks completion.
- [ ] The follow-up specifies repeated concrete handoffs for remaining actionable coverage; unproduced implementation briefs are labelled not yet derived and have no speculative IDs.
- [ ] Existing runtime/source behavior, frozen tracker schema, task status fields, and separate-author handoff remain intact.

## Deferrals

None. Producing the independently authored tracked follow-up is part of completion. Actual implementation-brief derivation is assigned to that real follow-up, while product implementation remains outside this planning task.

## Traces to

`docs/workplan/phase-0/README.md` ordered planning increment 2 (legacy reference 0.2); docs/roadmap.md §3 Stage 1 and §6 D1/D5; docs/contracts.md §8–9; docs/code-structure.md §4; ADR-0003 (Proposed).
