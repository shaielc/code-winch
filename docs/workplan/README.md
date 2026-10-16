# Code Winch workplan

This is a new planning workplan. The previous plan was closed; no former pending task is silently restored. The roadmap defines product Stages 0–6, while Phase 0 holds meta tasks — planning phases, converting issues into phase tasks, and updating the architecture and design set — and implementation begins at Phase 1.

**Phase design and revision, design-set updates, and the conversion of implementation-scope issues into phase tasks must enter through [Phase 0](phase-0/README.md).** A design-document update that travels with an implementation task's own contract change stays in that task. Deriving briefs from an increment a phase has already decided is that phase's own planning task, numbered in that phase. Consult the active phase and `docs/state.md`, record the disposition through the applicable skill, and update Phase 0 coverage and this index. An implementation finding must not be treated as authorization to skip phase revision, task derivation or separate-author review.

## Phases

| Phase | Roadmap milestone | Design | Brief coverage |
|---|---|---|---|
| [0](phase-0/README.md) | Meta tasks: phase planning, issue conversion, design-set updates | Open | Seven phase-design planning briefs |
| [1](phase-1/README.md) | Stage 0 — sandbox exchange | Repartitioned into 30 experience-led increments; independent re-audit pending | Three derivation tasks produced P1-002, P1-004, and P1-006 for child-status increments 1–3; P1-007 assigns the independent harness-output-area derivation |
| 2 | Stage 1 — first real vendor | Scheduled via Plan Phase 2 | Not yet derived |
| 3 | Stage 2 — structured views and resume | Scheduled via Plan Phase 3 | Not yet derived |
| 4 | Stage 3 — observe existing sandboxes | Scheduled via Plan Phase 4 | Not yet derived |
| 5 | Stage 4 — run lifecycle control | Scheduled via Plan Phase 5 | Not yet derived |
| 6 | Stage 5 — enforce isolation | Scheduled via Plan Phase 6 | Not yet derived |
| 7 | Stage 6 — second adapter and workflows | Scheduled via Plan Phase 7 | Not yet derived |

## Planning and execution

1. Use `skills/create-phase/SKILL.md` to decide one implementation phase's boundaries and artifacts, including the index.
2. Hand its increments to a **different task author** using `skills/derive-task/SKILL.md`; derive one complete task at a time and reconcile the brief, tracker, dependency reasons, coverage and index. Do not confuse a planning brief with its assigned planning operation.
3. Use the refinement, implementation, audit and phase-revision skills with their prescribed independent authors/PR boundaries. Route scope defects and issue resolution back to Phase 0 rather than bypassing the plan.
4. Keep status changes under repository automation; do not claim completion from an empty tracker.

## Task IDs

IDs are `P<phase>-<NNN>` and **each phase counts on its own**: `P0-123` and
`P1-123` are different tasks and both valid. The next ID in a phase is one past
the highest number carrying that prefix; gaps are never filled. Meta work in
Phase 0 and planning in Phase 1 therefore never compete for a number, and a
phase's planning tasks and the briefs they derive share that phase's sequence.
The brief an issue conversion produces takes its ID from the phase that owns
the work. The rule is defined in [the workplan model](../../skills/shared/workplan-model.md).

The IDs were first numbered from one plan-wide counter and have since been
re-enumerated so each phase counts on its own. The two derivation tasks that
turn a decided Phase 1 increment into a brief are Phase 1 planning tasks, so
they moved out of Phase 0 and number beside the briefs they derive. Pull
requests, briefs and dispatch history written before the re-enumeration cite
the former IDs; read them through this table:

| Former ID | Current ID | Task |
|---|---|---|
| P0-008 | P1-001 | Plan Phase 1 increment 1: child status presentation |
| P1-009 | P1-002 | Present unknown child status |
| P0-010 | P1-003 | Plan Phase 1 increment 2: sandbox-backed child status |
| P1-011 | P1-004 | Report sandbox-backed child status |

`P0-001`–`P0-007` are unchanged. The next IDs are `P0-008` and `P1-008`. A
former `P0-008` or `P0-010` in older material still names the retired
derivation task above, not a future Phase 0 task that takes that number.

## Phase dependency and exit rules

| Phase | Planning prerequisite | Exit evidence |
|---|---|---|
| 0 | Existing design, roadmap, state and HEAD | Seven reviewed designs and concrete separately authored briefs, reconciled coverage/graph |
| 1 | Phase 0 increment Plan Phase 1: design and briefs | Stage 0 exchange can be observed and survives reload |
| 2 | Phase 0 increment Plan Phase 2: design and briefs | Real agent behind wrapper with fake parity |
| 3 | Phase 0 increment Plan Phase 3: design and briefs | Multiple views and gap-free resume |
| 4 | Phase 0 increment Plan Phase 4: design and briefs | Two manually started sandboxes visible centrally |
| 5 | Phase 0 increment Plan Phase 5: design and briefs | Create/stop run through product |
| 6 | Phase 0 increment Plan Phase 6: design and briefs | Security refusals demonstrated |
| 7 | Phase 0 increment Plan Phase 7: design and briefs | Two vendors and a cross-run workflow |

These are phase outcome conditions, **not** task-graph dependency edges. The planning briefs below assign future phase design plus an independently authored tracked follow-up; they do not themselves author implementation phases.

## Task dependencies

| Task | Phase-design output | Depends on | Reason |
|---|---|---|---|
| [P0-001](phase-0/P0-001-plan-phase-1-sandbox-exchange.md) | [phase-1/README.md](phase-1/README.md) — sandbox exchange | None | Existing design/HEAD is sufficient; no pending artifact is consumed |
| [P0-002](phase-0/P0-002-plan-phase-2-first-real-vendor.md) | phase-2/README.md — first real vendor | None | Existing design/HEAD is sufficient; no pending artifact is consumed |
| [P0-003](phase-0/P0-003-plan-phase-3-structured-views-resume.md) | phase-3/README.md — structured views and resume | P0-002 | P0-002 (semantic: this artifact must consume the decided first-adapter output and native-exit boundary from phase-2/README.md) |
| [P0-004](phase-0/P0-004-plan-phase-4-observe-sandboxes.md) | phase-4/README.md — observe existing sandboxes | P0-003 | P0-003 (semantic: this artifact must preserve the decided ordinal/projection boundary in phase-3/README.md) |
| [P0-005](phase-0/P0-005-plan-phase-5-run-lifecycle.md) | phase-5/README.md — run lifecycle control | P0-004 | P0-004 (semantic: this artifact must extend the decided observe-only run/registration boundary in phase-4/README.md) |
| [P0-006](phase-0/P0-006-plan-phase-6-enforce-isolation.md) | phase-6/README.md — enforced isolation | None | Existing design/HEAD is sufficient; no pending artifact is consumed |
| [P0-007](phase-0/P0-007-plan-phase-7-second-adapter-workflows.md) | phase-7/README.md — second adapter and workflows | P0-002, P0-005 | P0-002 (semantic: this artifact must consume the first-adapter seam decided in phase-2/README.md); P0-005 (semantic: workflow boundaries require the ordinary run commands and event ownership decided in phase-5/README.md) |
| [P1-001](phase-1/P1-001-plan-owned-live-harness-output.md) | Phase 1 increment 1 implementation brief — child status presentation | P0-001 | P0-001 (semantic: the assigned brief must consume the reviewed increment boundary, prerequisite, and concern ownership decided in phase-1/README.md) |
| [P1-002](phase-1/P1-002-child-status-presentation.md) | Phase 1 increment 1 — present truthful unknown child status | None | Existing deployed attach page is sufficient; no planned product behavior is consumed |
| [P1-003](phase-1/P1-003-plan-sandbox-backed-child-status.md) | Phase 1 increment 2 implementation brief — sandbox-backed child status | P1-001 | P1-001 (semantic: derivation must consume the independently authored presentation boundary and preserve its accessible label/value contract) |
| [P1-004](phase-1/P1-004-sandbox-backed-child-status.md) | Phase 1 increment 2 — report sandbox-owned no-child status | P1-002 | P1-002 (semantic: implementation replaces its accessible `Unknown` value while preserving the child-status label/value identity) |
| [P1-005](phase-1/P1-005-plan-actual-child-status.md) | Phase 1 increment 3 implementation brief — actual child status | P1-003 | P1-003 (semantic: derivation must consume the `childStatus` contract and `Down` presentation defined in P1-004's brief and extend rather than redefine them) |
| [P1-006](phase-1/P1-006-actual-child-status.md) | Phase 1 increment 3 — report the owned fake child's actual status | P1-004 | P1-004 (contract: implementation extends its additive `/api/session` `childStatus` values and validated `Down` presentation with owned-process `Running` and `Exited`) |
| [P1-007](phase-1/P1-007-plan-harness-output-area.md) | Phase 1 increment 4 implementation brief — harness output area | P0-001 | P0-001 (semantic: derivation consumes the independently reviewed increment boundary and concern ownership in the Phase 1 design) |

**Partial derivation:** the seven phase-design tasks assign phase design. P1-001
produced P1-002 for Phase 1 increment 1, P1-003 produced P1-004 for increment
2, and P1-005 produced P1-006 for increment 3. P1-006 takes a `contract` edge
on P1-004 because it extends that pending brief's wire and presentation
contract. P1-007 assigns increment 4's independent presentation derivation and
depends only on the Phase 1 design produced by P0-001. No implementation ID for
increment 4 or later has been reserved.
All later specific-brief coverage remains not yet derived. Completion of each
phase-design task still requires its reviewed design and the planning task for
increment 1. Each phase-design and derivation task repeats the one-at-a-time
tracked handoff by completing `phase-increment-handoff`, which adds the
planning task for the next increment, allocating an ID only with a complete
brief. Product implementation
is never a dependency of a phase-design task. Recheck planned runtime
prerequisites against HEAD when dispatching; Stage 6's proven first adapter is
runtime execution evidence, not proof supplied by a completed planning task.

Graph: 14 nodes, 11 edges; critical path **5 tasks**: P0-002 → P0-003 →
P0-004 → P0-005 → P0-007. P0-002 → P0-007 is also a consumed-artifact
edge; P0-001 → P1-001 → P1-003 → P1-005 is the Phase 1 design and serial derivation
chain. P1-002 → P1-004 orders the sandbox-backed implementation after the
presentation identity it replaces, and P1-004 → P1-006 orders the additive
contract extension. P0-001 → P1-007 records the independent increment 4 design
input. Longest-path layer widths are 4, 4, 3, 2, 1, so average layer width is
**2.8** (14/5); root width is **4**: P0-001, P0-002, P0-006 and P1-002. These
are unweighted graph measures, not duration estimates. Currently available:
P0-002, P0-006, P1-004, P1-005 and P1-007.

Concurrent write warnings: every phase-design brief forecasts edits to the
shared index, Phase 0 coverage, and tracker for its independently derived
follow-up; P1-001, P1-003, P1-005, and P1-007 also forecast Phase 1 coverage.
The 14-node graph has **71 incomparable pairs**. P1-006 forecasts overlapping
session and React files with P1-004, but their contract edge orders those
writes. P1-007's future implementation brief is presentation-only and may
forecast `App.tsx`, `App.test.tsx`, and styles that overlap P1-004/P1-006; it
has no semantic dependency on their behavior, so this is a write collision to
rebase rather than an ordering edge. Among currently available tasks, P0-002,
P0-006, P1-005, and P1-007 share planning-record writes (index, tracker, and
Phase 0 coverage; both Phase 1 planning tasks also share Phase 1 coverage).
P1-004's forecast product files do not overlap the planning artifacts. Rebase
and reconcile shared records rather than adding ordering-only edges. No task
edits existing tracker status fields.

Contract collisions: **one ordered collision** in this derived graph. P1-002 owns the accessible
static `Child status: Unknown` presentation; P1-004 consumes and preserves that
identity while uniquely owning the additive `/api/session` `childStatus: "Down"`
contract and its validated replacement behavior. P1-006 extends that same wire
and presentation contract with `Running` and `Exited`, and its contract edge on
P1-004 prevents concurrent redefinition. P1-007 owns only a distinct output
region and collides on files, not product meaning. Phase 0 meta tasks change no running product behavior;
a meta task that revises the design set's meaning of a contract surface declares
that surface, so implementation tasks consuming it are ordered by the usual
contract rule. Required changes to shared product meaning must return through
Phase 0 for phase-author disposition and graph reconciliation before assignment.
