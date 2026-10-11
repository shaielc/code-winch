# Phase 0 — Plan the implementation phases

## Outcome and starting point

At exit, a contributor can locate a reviewed, bounded phase design and an assignable specific implementation brief for each planned implementation phase, with status, dependencies, and coverage recorded consistently. Phase 0 produces planning artifacts only; it does not implement roadmap capabilities.

The prior workplan was explicitly closed and `docs/workplan/tasks.json` is empty (`docs/state.md`, `AGENTS.md`). The sandbox page and local composer exist; the fake harness is shipped but is not launched by the sandbox. Harness output, retained history, persisted input, and delivery are missing or unwired (`docs/state.md`). Runtime checks are limited by the verification caveats there. `docs/roadmap.md` fixes Stages 0–6; workplan phase numbering is distinct.

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
| 1 | 0.1 Stage 0 | `phase-1/README.md` and concrete, separately authored Stage 0 briefs | Distinguish shipped page/composer from unwired runner; split live output, history, recorded input, delivery, malformed-output limits and stalled readers; running/deployable and operator paths | New runtime wiring itself; Stage 1 semantics | This phase, roadmap Stage 0 and HEAD |
| 2 | 0.2 Stage 1 | `phase-2/README.md` and specific Stage 1 briefs | Real-vendor selection trigger D1, genuine adapter seam, native exit reporting, fake parity | Second vendor or workflows | Roadmap Stage 1, current implementation state |
| 3 | 0.3 Stage 2 | `phase-3/README.md` and specific Stage 2 briefs | Structured records, dual projections, no-gap resume | Cross-session history/search/export | Roadmap Stage 2; Stage 1 contract as planned |
| 4 | 0.4 Stage 3 | `phase-4/README.md` and specific Stage 3 briefs | Discovery, registration/routing, read-only control-plane visibility; resolve D2/D4/D6 when triggered | Launch/stop control | Roadmap Stage 3; Stage 2 boundary |
| 5 | 0.5 Stage 4 | `phase-5/README.md` and specific Stage 4 briefs | Run creation, stop, lifecycle, leases, idempotency, reconciliation | Shared deployment/security guarantees | Roadmap Stage 4; Stage 3 boundary |
| 6 | 0.6 Stage 5 | `phase-6/README.md` and specific Stage 5 briefs | Enforced security posture and negative demonstrations | Second vendor and workflows | Roadmap Stage 5; security launch blockers |
| 7 | 0.7 Stage 6 | `phase-7/README.md` and specific Stage 6 briefs | Second real adapter and cross-run workflows | Unspecified future product expansion | Roadmap Stage 6 and proven first adapter |

The seven rows are a **planning handoff order**, not invented implementation dependency edges. Each phase design must be created by a phase author using `create-phase`; a different author then uses `derive-task` to derive **one specific brief at a time** and reconcile tracker, index, coverage, dependencies and collisions. Planning a brief depends on the target phase design, not on implementation of the preceding phase. Later phases must be rechecked against HEAD when their roadmap triggers arrive.

## Task coverage

| Increment | Phase design | Specific-brief coverage |
|---|---|---|
| 0.1 | Not yet authored | **Not yet derived** |
| 0.2 | Not yet authored | **Not yet derived** |
| 0.3 | Not yet authored | **Not yet derived** |
| 0.4 | Not yet authored | **Not yet derived** |
| 0.5 | Not yet authored | **Not yet derived** |
| 0.6 | Not yet authored | **Not yet derived** |
| 0.7 | Not yet authored | **Not yet derived** |

Next handoff: a **different task author** derives a Phase 0 planning task for increment 0.1 under `skills/derive-task/SKILL.md`, assigning `phase-1/README.md` to the phase author; deriving that brief does not itself create Phase 1. No IDs are allocated here and no pending runtime tasks are declared complete.

## Phase exit

All seven phase designs have explicit baseline, ordered increments, prerequisites, boundaries, exit observations and coverage; separately authored specific briefs cover their actionable increments; `tasks.json`, the index, write/contract collision analysis and dependency graph agree. The reviewed artifacts, rather than runtime demonstrations of unimplemented product features, are the acceptance evidence. Phase 0 remains the front door for future planning and scope/issue resolution.
