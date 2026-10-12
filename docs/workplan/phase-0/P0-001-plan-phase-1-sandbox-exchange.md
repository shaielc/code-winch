# P0-001: Plan phase-1: sandbox exchange

**Phase:** 0 — Plan the implementation phases
**Shape:** planning
**Dependencies:** None

## Objective

Produce a reviewed `docs/workplan/phase-1/README.md` that decides the observable increments completing roadmap Stage 0, and arrange the concrete, independently authored tracked Phase 0 follow-up required by the existing planning process.

## Scope

Use `skills/create-phase/SKILL.md` and `skills/shared/phase-design.md` to determine the ordered increments, boundaries, prerequisites, concern owners and observable phase exit from the roadmap, design set and current source. The phase author makes these decisions in the **phase artifact**, not by adopting a decomposition prescribed here.

Account for the remaining Stage 0 exchange: sandbox-resident child ownership and cleanup; live harness output and its retention/reload; browser input persistence and refusal when unavailable; actual delivery to the harness and its reply; malformed or excessive output, slow-reader isolation, and bounded delivery. Preserve the existing page and local composer where appropriate. Distinguish the existing local echo from persisted input and a delivered exchange.

Select coherent, independently observable increments with the wiring and minimum maintained operator path needed for each result. Determine dependencies from actual consumed behavior and contracts: output and input need not be artificially serialized. Include fake-profile scenario growth and inherited run/deploy gate preservation where the introduced behavior requires them. The fake harness is changeable, not a product-protocol constraint. Leave API design, substrate choices, fixture controls and implementation commands to subsequent task derivation and implementation.

## Required constraints

At refinement, `cmd/winch-sandbox/main.go` constructs only `attach.New`; `internal/adapters/transport/attach/server.go` serves GET/HEAD health, posture and static paths; `web/src/attach/App.tsx` keeps composer messages in browser component state rather than sending them; `cmd/winch/main.go` exposes only status; and the shipped fake harness is not started by the sandbox. These are source observations, not proof of runtime behavior. Recheck at implementation-branch HEAD.

Use current source, roadmap Stage 0, Phase 0 increment 0.1, contracts §§8–9 and ADR-0005/0006 as design authorities. The relevant workplan model and author/audit skills govern artifact format, handoffs, tracking, status preservation and independent review. Do not treat pending product behavior as already implemented.

## Non-goals

No product implementation, new product contracts or implementation briefs in this task. Roadmap Stages 1–5 own real-vendor and exit semantics, gap-free ordinal resume and structured views, control-plane and run lifecycle, canonical run sequencing, and shared-deployment security. Stage 0's reloadable history is not Stage 2's seamless resume.

The phase author does not write the independent follow-up planning brief.

## Planning outputs and handoff

- Create `docs/workplan/phase-1/README.md` with a sourced starting point, decided increment boundaries and prerequisites, concern ownership, exit observations, and implementation-task coverage marked **not yet derived**.
- Update `docs/workplan/README.md` and Phase 0 coverage so they point to the decided artifact.
- Obtain independent `audit-phase` review and resolve blocking findings.
- Arrange for a different task author to derive **at least one real, tracked Phase 0 planning follow-up** assigning derivation of the next uncovered Phase 1 implementation increment. Its dependencies and index/tracker/coverage records must be consistent. This concrete follow-up is required for P0-001 completion, not merely a verbal handoff. Remaining briefs are derived separately through the established process; allocate no speculative implementation IDs.

## Write set

Forecast: `docs/workplan/phase-1/README.md` (new), `docs/workplan/README.md`, `docs/workplan/phase-0/README.md`, `docs/workplan/tasks.json` (new follow-up entry only), and the independently authored `docs/workplan/phase-0/<allocated-ID>-<target-increment-slug>.md`. The separate author allocates its actual ID when deriving that complete follow-up. Preserve existing task status and metadata.

## Contract surfaces

No product code or wire-contract changes. Phase 1 delivery boundaries are the design output; shared planning-record edits are write collisions, not product contract ownership.

## Demonstration

Review the Phase 1 artifact, its observable boundaries, source references, coverage and next handoff. Confirm that Phase 0 and the index agree, that a concrete independently authored follow-up exists in the tracker, and that the independent phase audit has no unresolved blocking findings. This is an **artifact review**, not a runtime demonstration of the predicted sandbox exchange.

## Verification

Use `skills/audit-phase/SKILL.md`, the workplan model's planning-artifact checks, and the applicable repository gates in `AGENTS.md`. Verify planning-record consistency and report any unavailable gates accurately; no runtime capability is introduced by this task.

## Acceptance criteria

- [ ] The Phase 1 artifact has a source-supported baseline, decided and justified observable increments, prerequisites, one owner per concern, and a testable phase exit; implementation coverage is marked not yet derived.
- [ ] The design accounts for Stage 0 live output, retention, input, delivery, failure/reader guarantees and minimum operability without prescribing a decomposition or absorbing later roadmap stages.
- [ ] Existing behavior, planned prerequisites and unimplemented runtime promises remain distinguishable; fake harness behavior is not mistaken for a protocol constraint.
- [ ] Independent phase audit has no unresolved blocking findings; the index and Phase 0 agree with the reviewed design.
- [ ] A complete, independently authored Phase 0 follow-up planning brief is tracked with a real ID, appropriate dependency on P0-001 and matching coverage, while unproduced implementation briefs retain no placeholder IDs.
- [ ] Existing task statuses, tracker schema, and product behavior are preserved.

## Deferrals

Implementation-brief derivation belongs to the concrete tracked follow-up; product implementation remains outside P0-001. No unnamed deferral is introduced.

## Traces to

`docs/workplan/phase-0/README.md` ordered planning increment 1 (legacy 0.1); `docs/roadmap.md` §3 Stage 0 and §4; `docs/contracts.md` §§8–9; ADR-0005; ADR-0006.
