# Code Winch workplan

This is a new planning workplan. The previous plan was closed; no former pending task is silently restored. The roadmap defines product Stages 0–6, while Phase 0 holds meta tasks — planning phases, converting issues into phase tasks, and updating the architecture and design set — and implementation begins at Phase 1.

**Phase design and revision, design-set updates, and the conversion of implementation-scope issues into phase tasks must enter through [Phase 0](phase-0/README.md).** A design-document update that travels with an implementation task's own contract change stays in that task. Deriving briefs from an increment a phase has already decided is that phase's own planning task, numbered in that phase. Consult the active phase and `docs/state.md`, record the disposition through the applicable skill, and update Phase 0 coverage and this index. An implementation finding must not be treated as authorization to skip phase revision, task derivation or separate-author review.

## Phases

| Phase | Roadmap milestone | Design | Brief coverage |
|---|---|---|---|
| [0](phase-0/README.md) | Meta tasks: phase planning, issue conversion, design-set updates | Open | Seven phase-design planning briefs |
| [1](phase-1/README.md) | Stage 0 — sandbox exchange | Repartitioned into 30 experience-led increments; independent re-audit pending | Two derivation tasks (P1-001, P1-003) produced the child-status presentation brief P1-002 and the sandbox-backed status brief P1-004; P1-005 assigns increment 3's derivation |
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

`P0-001`–`P0-007` are unchanged. The next IDs are `P0-009` and `P1-006`. A
former `P0-008` or `P0-010` in older material still names the retired
derivation task above; the current `P0-008` is the issue-routing task below.

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
| [P0-008](phase-0/P0-008-route-go-gate-parity-issue.md) | Phase disposition for issue #149 (Go gate option parity) | None | Existing `Makefile`, `docs/testing.md` and HEAD are sufficient; no pending artifact is consumed |
| [P1-001](phase-1/P1-001-plan-owned-live-harness-output.md) | Phase 1 increment 1 implementation brief — child status presentation | P0-001 | P0-001 (semantic: the assigned brief must consume the reviewed increment boundary, prerequisite, and concern ownership decided in phase-1/README.md) |
| [P1-002](phase-1/P1-002-child-status-presentation.md) | Phase 1 increment 1 — present truthful unknown child status | None | Existing deployed attach page is sufficient; no planned product behavior is consumed |
| [P1-003](phase-1/P1-003-plan-sandbox-backed-child-status.md) | Phase 1 increment 2 implementation brief — sandbox-backed child status | P1-001 | P1-001 (semantic: derivation must consume the independently authored presentation boundary and preserve its accessible label/value contract) |
| [P1-004](phase-1/P1-004-sandbox-backed-child-status.md) | Phase 1 increment 2 — report sandbox-owned no-child status | P1-002 | P1-002 (semantic: implementation replaces its accessible `Unknown` value while preserving the child-status label/value identity) |
| [P1-005](phase-1/P1-005-plan-actual-child-status.md) | Phase 1 increment 3 implementation brief — actual child status | P1-003 | P1-003 (semantic: derivation must consume the `childStatus` contract and `Down` presentation defined in P1-004's brief and extend rather than redefine them) |

**Partial derivation:** the seven phase-design tasks assign phase design. P1-001
produced P1-002 for Phase 1 increment 1, and P1-003 produced P1-004 for
increment 2. P1-005 assigns the derivation of increment 3. It is pipelined: it
consumes P1-004's brief, not its implementation, so it is available while
P1-004 is being implemented. The brief it derives will take a `contract` edge on
P1-004. No implementation ID for increment 3 or later has been reserved.
P0-008 routes issue #149 to a phase author; it assigns no
implementation brief yet. All later specific-brief coverage remains not yet derived. Completion of each
phase-design task still requires its reviewed design and the planning task for
increment 1. Each phase-design and derivation task repeats the one-at-a-time
tracked handoff by completing `phase-increment-handoff`, which adds the
planning task for the next increment, allocating an ID only with a complete
brief. Product implementation
is never a dependency of a phase-design task. Recheck planned runtime
prerequisites against HEAD when dispatching; Stage 6's proven first adapter is
runtime execution evidence, not proof supplied by a completed planning task.

Graph: 13 nodes, 9 edges; critical path **5 tasks**: P0-002 → P0-003 →
P0-004 → P0-005 → P0-007. P0-002 → P0-007 is also a consumed-artifact
edge; P0-001 → P1-001 → P1-003 → P1-005 is the Phase 1 design and serial derivation
chain. P1-002 → P1-004 orders the sandbox-backed implementation after the
presentation identity it replaces. Longest-path layer widths are 5, 3, 2, 2, 1,
so average layer width is **2.6** (13/5); root width is **5**: P0-001, P0-002,
P0-006, P0-008 and P1-002. These are unweighted graph measures, not duration estimates.
Currently available: P0-002, P0-006, P0-008, P1-004 and P1-005.

Concurrent write warnings: every phase-design brief forecasts edits to the
shared index, Phase 0 coverage, and tracker for its independently derived
follow-up; P1-001, P1-003 and P1-005 also forecast Phase 1 coverage. The prior
11-node graph had 41 incomparable pairs. P1-005 is ordered only with its
ancestors P0-001, P1-001 and P1-003 and is incomparable with the other eight
tasks, for **49** incomparable pairs among the 12 earlier tasks. P0-008 has no edges and is incomparable with all 12, for **61 total**. P1-002 and P1-004 forecast
overlapping React files, but their dependency orders those writes. Among
currently available tasks, P0-002, P0-006, P0-008 and P1-005 share a planning-record
write collision (index, tracker, Phase 0 coverage); P1-004's forecast product
files do not overlap theirs, so P1-005's derivation and P1-004's implementation
proceed concurrently. Rebase/reconcile shared coverage rows rather than adding ordering-only
edges. No task edits existing tracker status fields.

Contract collisions: **none** in this derived graph. P1-002 owns the accessible
static `Child status: Unknown` presentation; P1-004 consumes and preserves that
identity while uniquely owning the additive `/api/session` `childStatus: "Down"`
contract and its validated replacement behavior. Their semantic edge prevents
concurrent redefinition. Phase 0 meta tasks change no running product behavior;
a meta task that revises the design set's meaning of a contract surface declares
that surface, so implementation tasks consuming it are ordered by the usual
contract rule. Required changes to shared product meaning must return through
Phase 0 for phase-author disposition and graph reconciliation before assignment.
