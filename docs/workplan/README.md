# Code Winch workplan

This is a new planning workplan. The previous plan was closed; no former pending task is silently restored. The roadmap defines product Stages 0–6, while Phase 0 controls planning and implementation begins at Phase 1.

**All future planning, planning revisions, and resolution of implementation-scope issues must enter through [Phase 0](phase-0/README.md).** Consult the active phase and `docs/state.md`, record the disposition through the applicable skill, and update Phase 0 coverage and this index. An implementation finding must not be treated as authorization to skip phase revision, task derivation or separate-author review.

## Phases

| Phase | Roadmap milestone | Design | Brief coverage |
|---|---|---|---|
| [0](phase-0/README.md) | Planning and issue-resolution coordination | Open | Seven phase-design planning briefs derived; specific-brief planning coverage not yet derived |
| 1 | Stage 0 — sandbox exchange | Scheduled via 0.1 | Not yet derived |
| 2 | Stage 1 — first real vendor | Scheduled via 0.2 | Not yet derived |
| 3 | Stage 2 — structured views and resume | Scheduled via 0.3 | Not yet derived |
| 4 | Stage 3 — observe existing sandboxes | Scheduled via 0.4 | Not yet derived |
| 5 | Stage 4 — run lifecycle control | Scheduled via 0.5 | Not yet derived |
| 6 | Stage 5 — enforce isolation | Scheduled via 0.6 | Not yet derived |
| 7 | Stage 6 — second adapter and workflows | Scheduled via 0.7 | Not yet derived |

## Planning and execution

1. Use `skills/create-phase/SKILL.md` to decide one implementation phase's boundaries and artifacts, including the index.
2. Hand its increments to a **different task author** using `skills/derive-task/SKILL.md`; derive one complete task at a time and reconcile the brief, tracker, dependency reasons, coverage and index. Do not confuse a planning brief with its assigned planning operation.
3. Use the refinement, implementation, audit and phase-revision skills with their prescribed independent authors/PR boundaries. Route scope defects and issue resolution back to Phase 0 rather than bypassing the plan.
4. Keep status changes under repository automation; do not claim completion from an empty tracker.

## Phase dependency and exit rules

| Phase | Planning prerequisite | Exit evidence |
|---|---|---|
| 0 | Existing design, roadmap, state and HEAD | Seven reviewed designs and concrete separately authored briefs, reconciled coverage/graph |
| 1 | Phase 0 increment 0.1 design and briefs | Stage 0 exchange can be observed and survives reload |
| 2 | Phase 0 increment 0.2 design and briefs | Real agent behind wrapper with fake parity |
| 3 | Phase 0 increment 0.3 design and briefs | Multiple views and gap-free resume |
| 4 | Phase 0 increment 0.4 design and briefs | Two manually started sandboxes visible centrally |
| 5 | Phase 0 increment 0.5 design and briefs | Create/stop run through product |
| 6 | Phase 0 increment 0.6 design and briefs | Security refusals demonstrated |
| 7 | Phase 0 increment 0.7 design and briefs | Two vendors and a cross-run workflow |

These are phase outcome conditions, **not** task-graph dependency edges. The planning briefs below assign future phase design plus an independently authored tracked follow-up; they do not themselves author implementation phases.

## Phase 0 task dependencies

| Task | Phase-design output | Depends on | Reason |
|---|---|---|---|
| [P0-001](phase-0/P0-001-plan-phase-1-sandbox-exchange.md) | phase-1/README.md — sandbox exchange | None | Existing design/HEAD is sufficient; no pending artifact is consumed |
| [P0-002](phase-0/P0-002-plan-phase-2-first-real-vendor.md) | phase-2/README.md — first real vendor | None | Existing design/HEAD is sufficient; no pending artifact is consumed |
| [P0-003](phase-0/P0-003-plan-phase-3-structured-views-resume.md) | phase-3/README.md — structured views and resume | P0-002 | P0-002 (semantic: this artifact must consume the decided first-adapter output and native-exit boundary from phase-2/README.md) |
| [P0-004](phase-0/P0-004-plan-phase-4-observe-sandboxes.md) | phase-4/README.md — observe existing sandboxes | P0-003 | P0-003 (semantic: this artifact must preserve the decided ordinal/projection boundary in phase-3/README.md) |
| [P0-005](phase-0/P0-005-plan-phase-5-run-lifecycle.md) | phase-5/README.md — run lifecycle control | P0-004 | P0-004 (semantic: this artifact must extend the decided observe-only run/registration boundary in phase-4/README.md) |
| [P0-006](phase-0/P0-006-plan-phase-6-enforce-isolation.md) | phase-6/README.md — enforced isolation | None | Existing design/HEAD is sufficient; no pending artifact is consumed |
| [P0-007](phase-0/P0-007-plan-phase-7-second-adapter-workflows.md) | phase-7/README.md — second adapter and workflows | P0-002, P0-005 | P0-002 (semantic: this artifact must consume the first-adapter seam decided in phase-2/README.md); P0-005 (semantic: workflow boundaries require the ordinary run commands and event ownership decided in phase-5/README.md) |

**Partial derivation:** these seven pending planning tasks assign phase design together with at least one independently authored tracked follow-up assigning implementation-brief derivation for the next uncovered increment. Every Phase 0 increment still lacks actual specific-brief planning coverage, and implementation briefs remain not yet derived. Completion of each phase-design task requires a separate task author to derive a real follow-up planning brief, add its tracker/index/coverage records, and record a semantic dependency on the phase-design task whose artifact it consumes. Allocate its ID only when the complete follow-up is derived; name the implementation increment until a real target brief exists. Repeat one-at-a-time tracked handoffs for remaining actionable coverage. Recompute this partial graph and collision report after each actual addition; the figures below describe only the seven tasks already derived. Product implementation is never a dependency of a phase-design task. Recheck planned runtime prerequisites against HEAD when dispatching; Stage 6's proven first adapter is runtime execution evidence, not proof supplied by a completed planning task.

Graph: 7 nodes, 5 edges; critical path **5 tasks**: P0-002 → P0-003 → P0-004 → P0-005 → P0-007. P0-002 → P0-007 is also a consumed-artifact edge. Layer widths are 3, 1, 1, 1, 1, so average layer width is **1.4** (7/5); immediately available width is **3**: P0-001, P0-002 and P0-006. These are unweighted graph measures, not duration estimates.

Concurrent write warnings: each brief forecasts edits to `docs/workplan/README.md`, `docs/workplan/phase-0/README.md`, and `docs/workplan/tasks.json` for its independently derived follow-up addition. The potentially concurrent pairs are P0-001 with P0-002/003/004/005/006/007, and P0-006 with P0-002/003/004/005/007 (11 pairs). Currently available collisions are (P0-001, P0-002), (P0-001, P0-006), and (P0-002, P0-006). Each owns a distinct target phase README. Rebase/reconcile shared coverage rows rather than adding ordering-only edges. No task edits tracker status fields.

Contract collisions: **none** in this derived assignment graph. Tasks consume product contracts without changing them and decide distinct phase boundaries. Required changes to shared product meaning must return through Phase 0 for phase-author disposition and graph reconciliation before assignment.

