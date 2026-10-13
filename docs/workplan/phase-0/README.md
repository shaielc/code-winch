# Phase 0 — Plan the implementation phases

## Outcome and starting point

At exit, a contributor can locate a reviewed, bounded phase design and an assignable specific implementation brief for each planned implementation phase, with status, dependencies, and coverage recorded consistently. Phase 0 produces planning artifacts only; it does not implement roadmap capabilities.

The prior workplan was explicitly closed and `docs/workplan/tasks.json` was empty when Phase 0 opened (`docs/state.md`, `AGENTS.md`). The sandbox page and local composer exist; the fake harness is shipped but is not launched by the sandbox. Harness output, retained history, persisted input, and delivery are missing or unwired (`docs/state.md`). Runtime checks are limited by the verification caveats there. `docs/roadmap.md` fixes Stages 0–6; workplan phase numbering is distinct.

Sources: [roadmap](../../roadmap.md), [state](../../state.md), [architecture](../../architecture.md), [contracts](../../contracts.md), [security](../../security.md), [code structure](../../code-structure.md), and [workplan model](../../../skills/shared/workplan-model.md).

## Concern boundaries

| Concern | Owner |
|---|---|
| Plan scope, index and subsequent issue-resolution routing | Phase 0; this artifact and the workplan index |
| Establish executable Stage 0 remainder, reusable completed slices and operator verification | 0.1 → implementation Phase 1 |
| First real CLI adapter and fake parity | 0.2 → implementation Phase 2 / roadmap Stage 1 |
| Legible records, views, ordinal resume | 0.3 → implementation Phase 3 / roadmap Stage 2 |
| Sandbox discovery and a read-only control plane | 0.4 → implementation Phase 4 / roadmap Stage 3 |
| Control-plane ownership of lifecycle | 0.5 → implementation Phase 5 / roadmap Stage 4 |
| Enforced isolation, secrets and egress | 0.6 → implementation Phase 6 / roadmap Stage 5 |
| Second adapter and cross-run workflows | 0.7 → implementation Phase 7 / roadmap Stage 6 |
| Product implementation and code changes | Respective implementation phases, never Phase 0 |
| Revisions and issue-resolution follow-ups | Route through Phase 0 planning/coverage and the relevant phase author; do not silently change scope |

## Ordered planning increments

| Order | Increment | Observable planning output | Includes | Excludes | Prerequisites |
|---|---|---|---|---|---|
| 1 | Plan Stage 0: sandbox exchange | `phase-1/README.md` and concrete, separately authored Stage 0 briefs | Distinguish shipped page/composer from unwired runner; split live output, history, recorded input, delivery, malformed-output limits and stalled readers; running/deployable and operator paths | New runtime wiring itself; Stage 1 semantics | This phase, roadmap Stage 0 and HEAD |
| 2 | Plan Stage 1: first real vendor | `phase-2/README.md` and specific Stage 1 briefs | Real-vendor selection trigger D1, genuine adapter seam, native exit reporting, fake parity | Second vendor or workflows | Roadmap Stage 1, current implementation state |
| 3 | Plan Stage 2: structured views and resume | `phase-3/README.md` and specific Stage 2 briefs | Structured records, dual projections, no-gap resume | Cross-session history/search/export | Roadmap Stage 2; Stage 1 contract as planned |
| 4 | Plan Stage 3: observe existing sandboxes | `phase-4/README.md` and specific Stage 3 briefs | Discovery, registration/routing, read-only control-plane visibility; resolve D2/D4/D6 when triggered | Launch/stop control | Roadmap Stage 3; Stage 2 boundary |
| 5 | Plan Stage 4: run lifecycle control | `phase-5/README.md` and specific Stage 4 briefs | Run creation, stop, lifecycle, leases, idempotency, reconciliation | Shared deployment/security guarantees | Roadmap Stage 4; Stage 3 boundary |
| 6 | Plan Stage 5: enforced isolation | `phase-6/README.md` and specific Stage 5 briefs | Enforced security posture and negative demonstrations | Second vendor and workflows | Roadmap Stage 5; security launch blockers |
| 7 | Plan Stage 6: second adapter and workflows | `phase-7/README.md` and specific Stage 6 briefs | Second real adapter and cross-run workflows | Unspecified future product expansion | Roadmap Stage 6 and proven first adapter |

The seven rows are a **planning handoff order**, not invented implementation dependency edges. Each phase design must be created by a phase author using `create-phase`; a different author then uses `derive-task` to derive **one specific brief at a time** and reconcile tracker, index, coverage, dependencies and collisions. Planning a brief depends on the target phase design, not on implementation of the preceding phase. Later phases must be rechecked against HEAD when their roadmap triggers arrive.

The stable legacy references `0.1`–`0.7` below identify these unchanged planning boundaries; increment titles carry no numbering.

## Task coverage

| Increment | Phase design | Specific-brief coverage |
|---|---|---|
| 0.1 (order 1) | [P0-001](P0-001-plan-phase-1-sandbox-exchange.md) assigns the authored and independently audited [Phase 1 sandbox-exchange design](../phase-1/README.md) plus a tracked follow-up | [P0-008](P0-008-plan-owned-live-harness-output.md) independently derived [P1-009](../phase-1/P1-009-child-status-presentation.md) for increment 1; [P0-010](P0-010-plan-sandbox-backed-child-status.md) assigns independent derivation of increment 2 without reserving its implementation ID |
| 0.2 (order 2) | [P0-002](P0-002-plan-phase-2-first-real-vendor.md) assigns design plus tracked follow-up for phase-2 design; **not yet authored** | **Not yet derived** |
| 0.3 (order 3) | [P0-003](P0-003-plan-phase-3-structured-views-resume.md) assigns design plus tracked follow-up for phase-3 design; **not yet authored** | **Not yet derived** |
| 0.4 (order 4) | [P0-004](P0-004-plan-phase-4-observe-sandboxes.md) assigns design plus tracked follow-up for phase-4 design; **not yet authored** | **Not yet derived** |
| 0.5 (order 5) | [P0-005](P0-005-plan-phase-5-run-lifecycle.md) assigns design plus tracked follow-up for phase-5 design; **not yet authored** | **Not yet derived** |
| 0.6 (order 6) | [P0-006](P0-006-plan-phase-6-enforce-isolation.md) assigns design plus tracked follow-up for phase-6 design; **not yet authored** | **Not yet derived** |
| 0.7 (order 7) | [P0-007](P0-007-plan-phase-7-second-adapter-workflows.md) assigns design plus tracked follow-up for phase-7 design; **not yet authored** | **Not yet derived** |

Coverage remains partial: P1-009 is the first derived Phase 1 implementation
brief and awaits separate refinement and implementation. P0-010 is the concrete
planning handoff for the second brief; that brief and all later Phase 1 briefs
are not yet derived. The other phase designs and specific-brief coverage likewise
remain pending.

The actual dependency graph and concurrent write warnings are recorded in the workplan index. The task author has updated coverage only; the seven planning boundaries and Phase 0 exit condition are unchanged.

## Phase exit

All seven phase designs have explicit baseline, ordered increments, prerequisites, boundaries, exit observations and coverage; separately authored specific briefs cover their actionable increments; `tasks.json`, the index, write/contract collision analysis and dependency graph agree. The reviewed artifacts, rather than runtime demonstrations of unimplemented product features, are the acceptance evidence. Phase 0 remains the front door for future planning and scope/issue resolution.
