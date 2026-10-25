# Phase 0 — Plan the implementation phases

## Outcome and starting point

At exit, a contributor can locate a reviewed, bounded phase design and an assignable specific implementation brief for each planned implementation phase, with status, dependencies, and coverage recorded consistently. Phase 0 holds meta tasks. Its outputs are phase designs and their revisions, the briefs that convert an issue into tasks of the phase that owns the work, and revisions to the design set (architecture, contracts, security, code structure, roadmap, ADRs), and changes to the repository toolchain and gates (Makefile, toolchain image, CI, testing documentation). It does not derive briefs from an increment a phase has already decided — that is the phase's own planning task — and it does not implement roadmap capabilities.

The prior workplan was explicitly closed and `docs/workplan/tasks.json` was empty when Phase 0 opened (`docs/state.md`, `AGENTS.md`). The sandbox page and local composer exist; the fake harness is shipped but is not launched by the sandbox. Harness output, retained history, persisted input, and delivery are missing or unwired (`docs/state.md`). Runtime checks are limited by the verification caveats there. `docs/roadmap.md` fixes Stages 0–6; workplan phase numbering is distinct.

Sources: [roadmap](../../roadmap.md), [state](../../state.md), [architecture](../../architecture.md), [contracts](../../contracts.md), [security](../../security.md), [code structure](../../code-structure.md), and [workplan model](../../../skills/shared/workplan-model.md).

## Concern boundaries

| Concern | Owner |
|---|---|
| Plan scope, index and subsequent issue-resolution routing | Phase 0; this artifact and the workplan index |
| Establish executable Stage 0 remainder, reusable completed slices and operator verification | Plan Phase 1, then implementation Phase 1 |
| First real CLI adapter and fake parity | Plan Phase 2, then implementation Phase 2 / roadmap Stage 1 |
| Legible records, views, ordinal resume | Plan Phase 3, then implementation Phase 3 / roadmap Stage 2 |
| Sandbox discovery and a read-only control plane | Plan Phase 4, then implementation Phase 4 / roadmap Stage 3 |
| Control-plane ownership of lifecycle | Plan Phase 5, then implementation Phase 5 / roadmap Stage 4 |
| Enforced isolation, secrets and egress | Plan Phase 6, then implementation Phase 6 / roadmap Stage 5 |
| Second adapter and cross-run workflows | Plan Phase 7, then implementation Phase 7 / roadmap Stage 6 |
| Converting an implementation-scope issue into phase tasks | Phase 0 meta task assigning derivation; the brief lands in the owning phase and takes its ID from that phase's own counter |
| Architecture and design-set updates (`docs/architecture.md`, `contracts.md`, `security.md`, `code-structure.md`, `roadmap.md`, ADRs) | Phase 0 meta task, with the affected phase designs and briefs reconciled; an update that travels with an implementation task's own contract change stays in that task |
| Product implementation and product code changes | Respective implementation phases, never Phase 0 |
| Repository toolchain and gate changes (Makefile, toolchain image, CI, `docs/testing.md`, the workplan control panel in `workplan-control-panel/`) | Phase 0 meta task; the brief lands here and takes a Phase 0 ID. Issue #149 is [P0-008](P0-008-align-go-gate-options.md); the control panel's stage-branch, pull-request, review/audit and return automation is [P0-009](P0-009-dispatch-on-stage-branch.md) to [P0-012](P0-012-return-reports-to-implementing-agent.md) |
| Revisions and issue-resolution follow-ups | Route through Phase 0 planning/coverage and the relevant phase author; do not silently change scope |

Phase 0 IDs count on their own (`P0-<NNN>`); see [Task IDs](../README.md#task-ids). Only meta tasks are Phase 0 tasks. A phase's own planning task, which derives briefs from an increment that phase has decided, belongs to that phase and shares its sequence with the briefs it derives.

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

The seven rows are a **planning handoff order**, not invented implementation dependency edges. Each phase design must be created by a phase author using `create-phase`, and the phase-design task ends with `phase-increment-handoff`, adding the planning task for increment 1. A different author then uses `derive-task` to derive **one specific brief at a time** and reconcile tracker, index, coverage, dependencies and collisions; each derivation task ends with `phase-increment-handoff` for the next increment. Planning a brief depends on the target phase design, not on implementation of the preceding phase. Later phases must be rechecked against HEAD when their roadmap triggers arrive.

Each increment is named for the phase it plans, "Plan Phase 1" to "Plan Phase 7", so a reference survives a change of order. Completed briefs P0-001, P1-001 and P1-003 cite the same increments in an earlier dotted form; read `0.N` there as "Plan Phase N".

## Task coverage

| Increment | Phase design | Specific-brief coverage |
|---|---|---|
| Plan Phase 1 (order 1) | [P0-001](P0-001-plan-phase-1-sandbox-exchange.md) assigns the authored and independently audited [Phase 1 sandbox-exchange design](../phase-1/README.md) plus a tracked follow-up | [P1-001](../phase-1/P1-001-plan-owned-live-harness-output.md) independently derived [P1-002](../phase-1/P1-002-child-status-presentation.md) for increment 1; [P1-003](../phase-1/P1-003-plan-sandbox-backed-child-status.md) independently derived [P1-004](../phase-1/P1-004-sandbox-backed-child-status.md) for increment 2; [P1-005](../phase-1/P1-005-plan-actual-child-status.md) independently derived [P1-006](../phase-1/P1-006-actual-child-status.md) for increment 3; [P1-007](../phase-1/P1-007-plan-harness-output-area.md) assigns increment 4's derivation |
| Plan Phase 2 (order 2) | [P0-002](P0-002-plan-phase-2-first-real-vendor.md) assigns design plus tracked follow-up for phase-2 design; **not yet authored** | **Not yet derived** |
| Plan Phase 3 (order 3) | [P0-003](P0-003-plan-phase-3-structured-views-resume.md) assigns design plus tracked follow-up for phase-3 design; **not yet authored** | **Not yet derived** |
| Plan Phase 4 (order 4) | [P0-004](P0-004-plan-phase-4-observe-sandboxes.md) assigns design plus tracked follow-up for phase-4 design; **not yet authored** | **Not yet derived** |
| Plan Phase 5 (order 5) | [P0-005](P0-005-plan-phase-5-run-lifecycle.md) assigns design plus tracked follow-up for phase-5 design; **not yet authored** | **Not yet derived** |
| Plan Phase 6 (order 6) | [P0-006](P0-006-plan-phase-6-enforce-isolation.md) assigns design plus tracked follow-up for phase-6 design; **not yet authored** | **Not yet derived** |
| Plan Phase 7 (order 7) | [P0-007](P0-007-plan-phase-7-second-adapter-workflows.md) assigns design plus tracked follow-up for phase-7 design; **not yet authored** | **Not yet derived** |

Issue [#149](https://github.com/shaielc/code-winch/issues/149) (Go gate option parity) is repository toolchain work, which Phase 0 owns; [P0-008](P0-008-align-go-gate-options.md) is its derived brief and is outside the seven-increment planning handoff order.

The workplan control panel is contributor tooling, not the product, so its changes are the same kind of Phase 0 meta task. Closing the manual Refine → pull request → review/audit → return loop is assigned to four serial briefs outside the seven-increment handoff order: [P0-009](P0-009-dispatch-on-stage-branch.md) creates and dispatches on the `<stage>/<ID>` branch; [P0-010](P0-010-publish-codex-stage-pull-request.md) publishes a finished Codex stage as a pull request; [P0-011](P0-011-start-review-or-audit-on-stage-pr.md) starts the review-refinement or audit session on it; [P0-012](P0-012-return-reports-to-implementing-agent.md) returns the report to the implementing agent. Each takes a `contract` edge on its predecessor. [`codex-interfaces.md`](../../../workplan-control-panel/codex-interfaces.md) records what the Codex CLI can and cannot supply, which bounds their design.

Coverage remains partial: P1-002, P1-004, and P1-006 are the first three derived
Phase 1 implementation briefs. P1-002 and P1-004 are implemented; P1-006 awaits
separate refinement and implementation. P1-004 consumes P1-002, and P1-006
extends P1-004's contract. P1-007 assigns
the independent derivation of increment 4; all later Phase 1 briefs remain not
yet derived, and no later implementation ID has been reserved. The other phase designs and specific-brief
coverage likewise remain pending.

The actual dependency graph and concurrent write warnings are recorded in the workplan index. The task author has updated coverage only; the seven planning boundaries and Phase 0 exit condition are unchanged.

## Phase exit

All seven phase designs have explicit baseline, ordered increments, prerequisites, boundaries, exit observations and coverage; separately authored specific briefs cover their actionable increments; `tasks.json`, the index, write/contract collision analysis and dependency graph agree. The reviewed artifacts, rather than runtime demonstrations of unimplemented product features, are the acceptance evidence. Phase 0 remains the front door for future planning, design-set updates and scope/issue resolution.
