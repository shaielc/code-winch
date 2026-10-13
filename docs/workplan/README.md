# Code Winch workplan

This is a new planning workplan. The previous plan was closed; no former pending task is silently restored. The roadmap defines product Stages 0–6, while Phase 0 controls planning and implementation begins at Phase 1.

**All future planning, planning revisions, and resolution of implementation-scope issues must enter through [Phase 0](phase-0/README.md).** Consult the active phase and `docs/state.md`, record the disposition through the applicable skill, and update Phase 0 coverage and this index. An implementation finding must not be treated as authorization to skip phase revision, task derivation or separate-author review.

## Phases

| Phase | Roadmap milestone | Design | Brief coverage |
|---|---|---|---|
| [0](phase-0/README.md) | Planning and issue-resolution coordination | Open | Seven phase-design planning briefs; first implementation brief derived and second specific-brief handoff assigned |
| [1](phase-1/README.md) | Stage 0 — sandbox exchange | Repartitioned into 30 experience-led increments; independent re-audit pending | Child-status presentation brief P1-009 derived; sandbox-backed status derivation assigned by P0-010 |
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
| [P0-008](phase-0/P0-008-plan-owned-live-harness-output.md) | Phase 1 increment 1 implementation brief — child status presentation | P0-001 | P0-001 (semantic: the assigned brief must consume the reviewed increment boundary, prerequisite, and concern ownership decided in phase-1/README.md) |
| [P1-009](phase-1/P1-009-child-status-presentation.md) | Phase 1 increment 1 — present truthful unknown child status | None | Existing deployed attach page is sufficient; no planned product behavior is consumed |
| [P0-010](phase-0/P0-010-plan-sandbox-backed-child-status.md) | Phase 1 increment 2 implementation brief — sandbox-backed child status | P0-008 | P0-008 (semantic: derivation must consume the independently authored presentation boundary and preserve its accessible label/value contract) |

**Partial derivation:** the seven phase-design tasks assign phase design. P0-008
produced the independently authored P1-009 implementation brief for Phase 1
increment 1, and P0-010 is the next concrete planning handoff for increment 2;
it reserves no implementation ID. All later specific-brief coverage remains not
yet derived. Completion of each phase-design task still requires its reviewed
design and at least one real follow-up. Repeat one-at-a-time tracked handoffs for
remaining actionable coverage, allocating an ID only with a complete brief.
Product implementation is never a dependency of a phase-design task. Recheck
planned runtime prerequisites against HEAD when dispatching; Stage 6's proven
first adapter is runtime execution evidence, not proof supplied by a completed
planning task.

Graph: 10 nodes, 7 edges; critical path **5 tasks**: P0-002 → P0-003 →
P0-004 → P0-005 → P0-007. P0-002 → P0-007 is also a consumed-artifact
edge; P0-001 → P0-008 is the Phase 1 design-to-brief edge, and P0-008 →
P0-010 orders the next derivation behind the presentation boundary it consumes.
P1-009 is a root because its implementation consumes only behavior already at
HEAD. Longest-path layer widths are 4, 2, 2, 1, 1, so average layer width is
**2.0** (10/5); root width is **4**: P0-001, P0-002, P0-006 and P1-009.
These are unweighted graph measures, not duration estimates. P0-010 is not
available until P0-008 completes; P1-009 is available independently.

Concurrent write warnings: every phase-design brief forecasts edits to the shared
index, Phase 0 coverage, and tracker for its independently derived follow-up;
P0-008 and P0-010 also forecast Phase 1 coverage. The potentially concurrent
pairs among P0-001–P0-008 remain 17. P1-009 is incomparable with all eight,
and P0-010 is incomparable with P0-002 through P0-007 and P1-009, for **32
total** incomparable pairs. P0-010 is ordered with P0-001 and P0-008 and those
are not concurrent pairs. Among currently available tasks, P0-002 and P0-006
retain a shared planning-record write collision; P1-009's forecast React files
do not overlap theirs. Rebase/reconcile shared coverage rows rather than adding
ordering-only edges. No task edits existing tracker status fields.

Contract collisions: **none** in this derived assignment graph. P1-009 uniquely
owns the accessible static `Child status: Unknown` presentation; P0-010's
planning output must consume that boundary and declare its future transport
surface before implementation dispatch. Planning tasks do not change product
contracts. Required changes to shared product meaning must return through Phase
0 for phase-author disposition and graph reconciliation before assignment.
