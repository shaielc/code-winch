---
name: workplan
description: Create, extend, re-derive, update, or audit an implementation workplan as a versioned task graph. Use for `/workplan create`, `/workplan extend`, `/workplan rederive`, `/workplan update`, and `/workplan audit`. Owns plan decomposition, invariant coverage, generations, dependency reasons, collisions, and re-derivation dispositions; does not audit one implementation against one task brief.
---

# Workplan

A workplan is a versioned implementation graph derived from the design baseline and the repository at HEAD. It must remain truthful as implementation accumulates.

The ownership boundary is strict:

> **`workplan` evaluates whether the plan and each brief's place in the graph are correct. `task` evaluates one implementation against its authored brief.**

## Active generation

If `docs/workplan/CURRENT` exists, it contains the active generation name such as `v2`; all active tracker, README, schema, and brief paths live under `docs/workplan/<generation>/`. Otherwise use legacy unversioned `docs/workplan/`.

`schema_version` says how to parse the tracker. `workplan_version` says which planning contract a task was authored under. They are different concepts.

Completed historical tasks may remain on an older workplan version. Unfinished work materially rewritten under V2 uses `workplan_version: 2`.

## Modes

1. **Create** — derive the first complete plan from design documents and HEAD.
2. **Extend** — add genuinely new work to an active plan; append IDs and preserve history.
3. **Re-derive** — reconcile every unfinished task against HEAD and the current planning contract.
4. **Update** — change status, ownership, blocking information, or planning disposition truthfully.
5. **Audit** — inspect the plan as a system without modifying it.

## Two layers of conformance

### Repository invariants at HEAD

Current repository invariants cannot be grandfathered by historical task versions. At minimum check:

- **I1 — Runnable:** a supported configuration starts.
- **I2 — Deployable:** deployment/configuration remains a maintained path.
- **I3 — Controllable fake profile:** fakes are runnable product configurations and can be driven intentionally.
- **I4 — Standing scenario parity:** substrate swaps prove the same standing scenarios.
- **I5 — Immediate operator reachability:** every new operator-visible capability is reachable through the maintained hands-on surface in the same task that introduces it.
- **I6 — Owned deferrals:** no TODO, stub, or postponed behavior exists without an owning task or explicit architectural decision trigger.
- **I7 — Real width:** dependency and collision structure permits meaningful parallel work rather than only nominal availability.

If any invariant fails at HEAD, the current plan must contain explicit current-version repair work. A legacy annotation explains old task conformance; it never excuses a current invariant gap.

### Planning-contract conformance

Judge each task brief using its `workplan_version`. Do not retroactively fail a completed V1 task because V2 later introduced fields or dependency vocabulary.

V2 tasks must follow the V2 brief and tracker rules below.

## Task shapes and completion closure

Every executable task is one of:

- **Seam** — introduces a boundary plus the implementation, wiring, public interaction surface, and minimal maintained operator/debug path needed to exercise it.
- **Swap** — replaces an implementation behind an existing seam while preserving scenario parity.
- **Capability** — adds a distinct observable behavior and its complete operator-visible loop.
- **Hardening** — adds an independently observable reliability, security, recovery, performance, or correctness guarantee.

A task may not delegate part of its own completion condition to a later task. If task B's main purpose is to make capability A usable through the maintained CLI, B belongs in A.

The opposite error is also a defect. Do not pull future refactors, extension scaffolding, unrelated observability, collision-avoidance restructuring, or independent hardening into a seam merely because it is nearby.

Classify scope bullets as required behavior, required wiring, required operator reachability, required invariant preservation, future enablement, hardening, or revision. A seam normally contains only the first four.

## Dependencies

Every dependency edge names exactly one reason and a one-clause justification:

- **compile** — the dependent cannot build without an artifact produced by the prerequisite.
- **contract** — the dependent needs a contract definition established by the prerequisite.
- **semantic** — the dependent's demonstration is impossible until the prerequisite behavior exists.
- **revision** — the dependent intentionally revises, hardens, restructures, optimizes, or tests implementation substantially created or rewritten by the prerequisite.

`revision` is not an ordering escape hatch. It is valid only when the dependent names the implementation/test surfaces it revises and those surfaces do not meaningfully exist in their intended form before the prerequisite.

Invalid reasons include nicer development order, possible rework, same subsystem, review convenience, and general design influence.

## Write collisions and contract collisions

V2 separates two concurrency risks.

**Write set** lists concrete files or deliberately broad paths the task expects to modify. Two concurrently available tasks have a write collision when their declared write sets overlap in a way likely to require merge/rebase coordination.

**Contract surfaces** list externally significant semantic namespaces the task may change even when files differ, for example API operations, OpenAPI components, persisted aggregates/state transitions, configuration keys, registry namespaces, and migration ownership boundaries.

Report write collisions and contract collisions separately. A collision does not automatically create a dependency; narrow ownership or deliberately serialize only when simultaneous work is unsafe.

## V2 brief template

```markdown
# <ID>: <Title>

**Phase:** <N> — <phase name>
**Shape:** seam | swap | capability | hardening
**Workplan version:** 2
**Dependencies:** <ID> (compile | contract | semantic | revision: <one clause>), … or None

## Objective

One observable sentence about the running system.

## Scope

- Required work only.

## Non-goals

- Explicit future enablement, hardening, or revision that is not required for completion.

## Runtime reachability

Composition root, profile, public interaction surface, and maintained operator/debug command that exercise the behavior.

## Write set

- Concrete files or intentionally broad paths expected to change.

## Contract surfaces

- `API: ...`
- `configuration: ...`
- `migration namespace: ...`

Or: `None.`

## Demonstration

    $ <command>
    → expect: <observable result>

## Verification

- Standing scenario suite against the applicable profile/substrate.
- Focused automated and negative/adversarial tests where required.
- Repository checks affected by the change.

## Acceptance criteria

- [ ] Observable checkable statements.
- [ ] Existing repository invariants remain true.

## Deferrals

| Deferred | Owning task |
|---|---|
| <thing> | <existing ID> |

Or: `None.`

## Traces to

`docs/<file>.md` §<section>, ADR-<nnnn>
```

## V2 tracker contract

New V2 generations use `skills/workplan/tasks.schema.json` copied into the generation as `tasks.schema.json`.

Each task contains the legacy execution fields plus:

- `workplan_version` — integer planning-contract version.
- `supersedes` — historical task IDs this task replaces.
- `superseded_by` — replacement IDs when this task is terminally superseded.
- `removal_reason` — non-empty only for terminally removed work.

V2 status values are:

`pending`, `in_progress`, `blocked`, `completed`, `superseded`, `removed`.

`superseded` and `removed` are planning-history terminal states, not successful dependency completion. Executable dependencies must become `completed`.

IDs remain append-only within a generation. Preserve an inherited task ID for a genuine rewrite of the same work; create replacement IDs when scope changes enough that historical evidence would become ambiguous.

## Re-derivation procedure

V2 re-derivation is reconciliation, not a free-form clean slate.

1. Record the current workplan contract version and repository baseline SHA.
2. Evaluate every repository invariant at HEAD.
3. Add explicit current-version repair tasks for every failed invariant not already adequately owned.
4. Inventory every inherited task whose status is not `completed`.
5. Give each inherited unfinished task exactly one disposition:
   - **Rewrite** — retain the ID, rewrite it to V2, and keep it executable.
   - **Split** — mark the inherited task `superseded` and point `superseded_by` at two or more replacement tasks.
   - **Merge/replace** — mark it `superseded` and point at the replacement task(s).
   - **Remove** — mark it `removed` with a concrete `removal_reason`.
6. Rewrite every surviving unfinished brief to V2.
7. Re-evaluate every dependency using compile/contract/semantic/revision.
8. Declare Write set and Contract surfaces separately.
9. Recompute availability, critical path, average width, write collisions, and contract collisions.
10. Verify every deferral has an owner or architectural decision trigger.
11. Validate the tracker against the generation schema and align README tables with it.
12. Record a re-derivation report/changelog.

A re-derivation is incomplete while any inherited unfinished task lacks a machine-readable disposition.

Use `scripts/prepare_workplan_rederivation.py prepare` to archive the old generation and seed the next one. The helper deliberately carries all inherited tasks into V2 so none can disappear silently. It stamps completed history as legacy version 1 and unfinished tasks as needing reconciliation. `harvest` refuses to copy the new generation back until every inherited unfinished task is rewritten, superseded, or removed.

## Required re-derivation report

Record at least:

- HEAD invariants failing before re-derivation and the task(s) repairing each one;
- inherited unfinished tasks rewritten, split, superseded/merged, or removed and why;
- dependency-reason changes;
- critical path and average width per phase;
- write collisions;
- contract collisions;
- legacy-version tasks relevant to current invariant gaps.

## Audit mode

`/workplan audit` asks: **is this a truthful, coherent, correctly decomposed implementation plan for the repository at HEAD?**

Report:

- **Invariant coverage** — I1–I7 with checked evidence and explicit repair ownership for failures.
- **Design coverage** — every documented requirement maps to work or a deliberate deferral; every task traces to a requirement.
- **Completion closure** — flag operator-visible capabilities whose maintained hands-on path was postponed.
- **Overstuffing** — flag future enablement, unrelated hardening, or pre-emptive refactors folded into seams without being required for completion.
- **Graph** — no unknown refs or cycles; every edge has a valid V2 reason for V2 tasks.
- **Width** — critical path and average width by phase.
- **Write collisions** — concurrently available pairs with overlapping write sets.
- **Contract collisions** — concurrently available pairs that may change the same semantic namespace.
- **Deferrals** — every implementation deferral has an owning task; design deferrals have revisit triggers.
- **Re-derivation dispositions** — every inherited unfinished task is accounted for.
- **Status truthfulness** — sampled completion claims hold against HEAD.

Do not substitute this graph audit for `/task audit <ID>`, which requires implementation evidence against one authored brief.

## Before finishing

- The active generation is explicit and machine-readable.
- Current invariant gaps have explicit repair work.
- Every V2 executable brief has runtime reachability, Write set, Contract surfaces, demonstration, verification, acceptance criteria, deferrals, and traceability.
- Every V2 dependency names compile, contract, semantic, or revision with a one-clause reason.
- Every inherited unfinished task in a re-derivation has exactly one disposition.
- `tasks.json` validates against the active generation schema.
- README and tracker describe the same graph.
- Critical path, average width, write collisions, and contract collisions are recorded.
- The maintained operator surface closes each new operator-visible capability in the same task.
