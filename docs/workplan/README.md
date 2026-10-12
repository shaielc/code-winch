# Code Winch workplan

This is a new planning workplan. The previous plan was closed; no former pending task is silently restored. The roadmap defines product Stages 0–6, while Phase 0 controls planning and implementation begins at Phase 1.

**All future planning, planning revisions, and resolution of implementation-scope issues must enter through [Phase 0](phase-0/README.md).** Consult the active phase and `docs/state.md`, record the disposition through the applicable skill, and update Phase 0 coverage and this index. An implementation finding must not be treated as authorization to skip phase revision, task derivation or separate-author review.

## Phases

| Phase | Roadmap milestone | Design | Brief coverage |
|---|---|---|---|
| [0](phase-0/README.md) | Planning and issue-resolution coordination | Open | Seven phase-design planning briefs and the first specific-brief planning handoff derived |
| [1](phase-1/README.md) | Stage 0 — sandbox exchange | Designed and independently audited; first tracked handoff assigned by P0-008 | Increment 1 derivation assigned; implementation brief not yet derived |
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
| [P0-001](phase-0/P0-001-plan-phase-1-sandbox-exchange.md) | [phase-1/README.md](phase-1/README.md) — sandbox exchange | None | Existing design/HEAD is sufficient; no pending artifact is consumed |
| [P0-002](phase-0/P0-002-plan-phase-2-first-real-vendor.md) | phase-2/README.md — first real vendor | None | Existing design/HEAD is sufficient; no pending artifact is consumed |
| [P0-003](phase-0/P0-003-plan-phase-3-structured-views-resume.md) | phase-3/README.md — structured views and resume | P0-002 | P0-002 (semantic: this artifact must consume the decided first-adapter output and native-exit boundary from phase-2/README.md) |
| [P0-004](phase-0/P0-004-plan-phase-4-observe-sandboxes.md) | phase-4/README.md — observe existing sandboxes | P0-003 | P0-003 (semantic: this artifact must preserve the decided ordinal/projection boundary in phase-3/README.md) |
| [P0-005](phase-0/P0-005-plan-phase-5-run-lifecycle.md) | phase-5/README.md — run lifecycle control | P0-004 | P0-004 (semantic: this artifact must extend the decided observe-only run/registration boundary in phase-4/README.md) |
| [P0-006](phase-0/P0-006-plan-phase-6-enforce-isolation.md) | phase-6/README.md — enforced isolation | None | Existing design/HEAD is sufficient; no pending artifact is consumed |
| [P0-007](phase-0/P0-007-plan-phase-7-second-adapter-workflows.md) | phase-7/README.md — second adapter and workflows | P0-002, P0-005 | P0-002 (semantic: this artifact must consume the first-adapter seam decided in phase-2/README.md); P0-005 (semantic: workflow boundaries require the ordinary run commands and event ownership decided in phase-5/README.md) |
| [P0-008](phase-0/P0-008-plan-owned-live-harness-output.md) | Phase 1 increment 1 implementation brief — owned live harness output | P0-001 | P0-001 (semantic: the assigned brief must consume the reviewed increment boundary, prerequisite, and concern ownership decided in phase-1/README.md) |

**Partial derivation:** the seven pending phase-design tasks assign phase design, and P0-008 is the first concrete independently authored follow-up assigning implementation-brief derivation. Phase 1 increment 1 now has planning-handoff coverage, but its implementation brief remains not yet derived; all other specific-brief coverage is also not yet derived. Completion of each phase-design task still requires its reviewed design and at least one real follow-up. Repeat one-at-a-time tracked handoffs for remaining actionable coverage, allocating an ID only with a complete planning brief. Product implementation is never a dependency of a phase-design task. Recheck planned runtime prerequisites against HEAD when dispatching; Stage 6's proven first adapter is runtime execution evidence, not proof supplied by a completed planning task.

Graph: 8 nodes, 6 edges; critical path **5 tasks**: P0-002 → P0-003 → P0-004 → P0-005 → P0-007. P0-002 → P0-007 is also a consumed-artifact edge, and P0-001 → P0-008 is the independently decided Phase 1 design-to-brief edge. Longest-path layer widths are 3, 2, 1, 1, 1, so average layer width is **1.6** (8/5); root width is **3**: P0-001, P0-002 and P0-006. These are unweighted graph measures, not duration estimates. P0-008 is not available until P0-001 completes.

Concurrent write warnings: every phase-design brief forecasts edits to `docs/workplan/README.md`, `docs/workplan/phase-0/README.md`, and `docs/workplan/tasks.json` for its independently derived follow-up addition; P0-008 forecasts those same records plus the Phase 1 coverage record for the assigned derivation. The potentially concurrent pairs among P0-001–P0-007 remain 11; P0-008 is incomparable with P0-002 through P0-007, adding 6 pairs for **17 total**. P0-001 and P0-008 are ordered by a semantic edge and are not a concurrent pair. Currently available collisions remain (P0-001, P0-002), (P0-001, P0-006), and (P0-002, P0-006). Rebase/reconcile shared coverage rows rather than adding ordering-only edges. No task edits tracker status fields.

Contract collisions: **none** in this derived assignment graph. Planning tasks consume product contracts without changing them; P0-008 assigns a future brief that must declare its owned and consumed surfaces before implementation dispatch. Required changes to shared product meaning must return through Phase 0 for phase-author disposition and graph reconciliation before assignment.
