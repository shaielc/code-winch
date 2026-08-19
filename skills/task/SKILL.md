---
name: task
description: Implement or audit exactly one active workplan task against its authored brief. Use for `/task implement <ID>` and `/task audit <ID>`. Do not redesign the workplan graph; report plan defects back to the workplan skill.
---

# Task

This skill owns one task and its implementation. It never owns plan decomposition.

> **`task` evaluates work against the brief. `workplan` evaluates whether the brief and its place in the graph are correct.**

Resolve the active workplan first. If `docs/workplan/CURRENT` exists, read the generation it names under `docs/workplan/<generation>/`; otherwise use legacy unversioned `docs/workplan/`.

Read the selected tracker entry, its brief, completed dependencies, applicable `AGENTS.md`, and relevant design documents.

For planning-contract rules, resolve `workplan_version: N` to the immutable `skills/workplan/contracts/vN.md`. If that contract file is missing, the task cannot be audited truthfully; report the missing contract instead of substituting the current rules.

## Implement

`/task implement <ID>` means implement this task exactly as specified and prove it complete.

Status handling is strict:

- `pending` — may be started after dependencies are verified.
- `in_progress` — may be continued when this is the appropriate active implementation.
- `blocked` — report `blocked_reason` and stop. Do not implement until the blocker is resolved.
- `completed` — reject as an implementation target.
- `superseded` — reject as planning history.
- `removed` — reject as planning history.

Then:

1. Verify every dependency is `completed`; `superseded` and `removed` do not satisfy execution dependencies.
2. Inspect the current implementation before editing.
3. Implement the objective and acceptance criteria.
4. Stay inside declared **Write set** and **Contract surfaces** unless repository reality requires a deviation; name deviations explicitly.
5. Preserve repository invariants that already hold.
6. Run required automated verification and the brief's manual demonstration.
7. Report exactly what was observed.

If implementation reveals the task is impossible, incorrectly decomposed, missing part of its completion condition, or overstuffed with unrelated future enablement, report a **plan defect**. Do not silently redesign the workplan.

## Audit

`/task audit <ID>` asks whether the implementation satisfies this task's authored brief.

Judge task/brief conformance using the task's historical workplan contract. Separately verify that the implementation did not violate repository invariants that should hold at HEAD.

Check at least objective, scope, runtime reachability, demonstration, acceptance criteria, required verification, deferrals, declared write and contract surfaces, invariant preservation, and status truthfulness.

Return `PASS` or `FAIL`, with blocking findings for a failure.

A task audit may pass while a workplan audit fails. That is expected when the authored task was conformant to its historical contract but the current plan decomposition or repository invariants need repair.

## V2 completion closure

For a V2 operator-visible capability, the task is incomplete unless the minimal maintained hands-on path needed to exercise it ships in the same task. Raw HTTP, database inspection, test harnesses, and temporary development commands do not substitute for the maintained operator surface.

Do not require unrelated future hardening or refactoring merely because it touches the same seam.

## Before finishing

- Active generation and short task ID were resolved unambiguously.
- Status permits the requested mode.
- Dependencies were verified.
- The authored workplan contract was loaded.
- Demonstration and required verification were actually checked.
- Surface deviations are named.
- Plan defects are reported rather than silently redesigned.
