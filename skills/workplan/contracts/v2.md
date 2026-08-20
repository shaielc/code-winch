---
name: workplan
description: Create, extend, re-derive, update, or audit an implementation workplan as a versioned task graph. Owns plan decomposition, invariant coverage, dependency reasons, collisions, and generation-scoped planning; does not audit one implementation against one task brief.
---

# Workplan

A workplan is a versioned implementation graph derived from the design baseline and the repository at HEAD. It must remain truthful as implementation accumulates.

> **`workplan` evaluates whether the plan and each brief's place in the graph are correct. `task` evaluates one implementation against its authored brief.**

## Active generation and contracts

If `docs/workplan/CURRENT` exists, read the generation it names under `docs/workplan/<generation>/`; otherwise use the legacy unversioned `docs/workplan/` layout.

`schema_version` describes the tracker format. `workplan_version` identifies the planning contract under which a task was authored. Normative contracts are immutable:

- `skills/workplan/contracts/v1.md`
- `skills/workplan/contracts/v2.md`

A task audit resolves `workplan_version: N` to `skills/workplan/contracts/vN.md`.

## Modes

1. **Create** — derive a complete plan from design documents and HEAD.
2. **Extend** — add genuinely new work to the active generation.
3. **Re-derive** — derive the remaining implementation graph from HEAD, the design baseline, and the completed implementation facts present in the active generation.
4. **Update** — change status, ownership, blocking information, or within-generation planning disposition truthfully.
5. **Audit** — inspect the plan as a system without modifying it.

For re-derivation, treat completed implementation records as facts about what exists. Do not infer unimplemented work from absent future-plan assertions. Re-check the design set and repository invariants directly, then derive the remaining task boundaries from current reality.

## Two layers of conformance

### Repository invariants at HEAD

Current repository invariants cannot be grandfathered by historical task versions. At minimum check:

- **I1 — Runnable:** a supported configuration starts.
- **I2 — Deployable:** deployment/configuration remains a maintained path.
- **I3 — Controllable fake profile:** fakes are runnable product configurations and can be driven intentionally.
- **I4 — Standing scenario parity:** substrate swaps prove the same standing scenarios.
- **I5 — Immediate operator reachability:** every new operator-visible capability is reachable through the maintained hands-on surface in the same task that introduces it.
- **I6 — Owned deferrals:** no TODO, stub, or postponed behavior exists without an owning task or explicit architectural decision trigger.
- **I7 — Real width:** dependency and collision structure permits meaningful parallel work rather than nominal availability only.

If an invariant fails at HEAD, the active plan contains explicit current-version repair work.

### Planning-contract conformance

Judge each task using the contract recorded in `workplan_version`. A completed V1 task is not malformed merely because V2 later introduced different representation rules.

## Task shapes and completion closure

Every executable V2 task is one of:

- **Seam** — introduces a boundary plus implementation, wiring, public interaction surface, and minimal maintained operator/debug path.
- **Swap** — replaces an implementation behind an existing seam while preserving scenario parity.
- **Capability** — adds a distinct observable behavior and its complete operator-visible loop.
- **Hardening** — adds an independently observable reliability, security, recovery, performance, or correctness guarantee.

A task may not delegate part of its own completion condition to a later task. If another task's main purpose would be to make this capability usable through the maintained operator surface, that operator path belongs here.

Do not overstuff seams with future refactors, unrelated hardening, extension scaffolding, or collision-avoidance restructuring. A seam normally contains required behavior, required wiring, required operator reachability, and required invariant preservation only.

## Dependencies

Every V2 dependency edge names exactly one reason with a one-clause justification:

- **compile** — the dependent cannot build without an artifact produced by the prerequisite.
- **contract** — the dependent needs a contract definition established by the prerequisite.
- **semantic** — the dependent's demonstration is impossible until the prerequisite behavior exists.
- **revision** — the dependent intentionally revises, hardens, restructures, optimizes, or tests implementation substantially created or rewritten by the prerequisite.

`revision` is not a generic ordering escape hatch. Invalid reasons include nicer development order, possible rework, same subsystem, review convenience, and general design influence.

## Write collisions and contract collisions

V2 separates two concurrency risks.

**Write set** lists concrete files or deliberately broad paths expected to change. A write collision is likely merge/rebase contention.

**Contract surfaces** list externally significant semantic namespaces the task may change even when files differ: API operations, OpenAPI components, persisted aggregates/state transitions, configuration keys, registry namespaces, migration ownership boundaries, and similar contracts.

Report write collisions and contract collisions separately. A collision does not automatically create a dependency.

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

- Explicit future enablement, hardening, or revision not required for completion.

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

V2 generations use the tracker schema in their own `tasks.schema.json`.

Each V2 task contains the execution fields plus `workplan_version`, `supersedes`, `superseded_by`, and `removal_reason`.

V2 status values are `pending`, `in_progress`, `blocked`, `completed`, `superseded`, and `removed`. `superseded` and `removed` are terminal planning-history states, not successful dependency completion.

Task IDs are scoped to the active workplan generation. Human-facing commands may use the short ID after resolving the active generation; persistent automation identity uses `<generation>/<ID>`.

## Re-derivation requirements

When deriving the remaining graph:

1. Read HEAD and the full design baseline.
2. Treat the completed implementation records in the active tracker as established facts.
3. Check I1–I7 directly against repository state and add explicit repair tasks for failures.
4. Derive all remaining task boundaries from current architecture and behavior.
5. Declare dependency reasons, Write sets, Contract surfaces, demonstrations, verification, acceptance criteria, and owned deferrals.
6. Recompute critical path, average width, write collisions, and contract collisions.
7. Validate the tracker against the active generation schema, verify every brief path and dependency, and ensure README summaries match the graph.

## Audit mode

`/workplan audit` asks: **is this a truthful, coherent, correctly decomposed implementation plan for the repository at HEAD?**

Report invariant coverage, design coverage and deferrals, completion closure, overstuffed seams, graph validity and dependency reasons, critical path and average width, write collisions, contract collisions, status truthfulness, and contract-version resolvability.

Do not substitute this graph audit for `/task audit <ID>`, which requires implementation evidence against one authored brief.

## Before finishing

- The active generation is explicit and machine-readable.
- Every task's `workplan_version` resolves to an immutable contract file.
- Current invariant gaps have explicit repair work.
- Every V2 executable brief has runtime reachability, Write set, Contract surfaces, demonstration, verification, acceptance criteria, deferrals, and traceability.
- Every V2 dependency names compile, contract, semantic, or revision with a one-clause reason.
- `tasks.json` validates against the active generation schema.
- README and tracker describe the same graph.
- Critical path, average width, write collisions, and contract collisions are recorded.
- The maintained operator surface closes each new operator-visible capability in the same task.
