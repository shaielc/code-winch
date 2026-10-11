# Code Winch workplan

This is a new planning workplan. The previous plan was closed; no former pending task is silently restored. The roadmap defines product Stages 0–6, while Phase 0 controls planning and implementation begins at Phase 1.

**All future planning, planning revisions, and resolution of implementation-scope issues must enter through [Phase 0](phase-0/README.md).** Consult the active phase and `docs/state.md`, record the disposition through the applicable skill, and update Phase 0 coverage and this index. An implementation finding must not be treated as authorization to skip phase revision, task derivation or separate-author review.

## Phases

| Phase | Roadmap milestone | Design | Brief coverage |
|---|---|---|---|
| [0](phase-0/README.md) | Planning and issue-resolution coordination | Open | Planning briefs not yet derived |
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

These are phase outcome conditions, **not** task-graph dependency edges. The tracker currently contains zero tasks: critical path 0, average available width 0, concurrent write collisions none, contract collisions none. Once briefs exist, report the actual partial graph and avoid inventing edges from stage sequence alone.
