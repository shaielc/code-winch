---
name: task
description: Implement or audit exactly one active workplan task against its authored brief. Use for `/task implement <ID>` and `/task audit <ID>`. Do not redesign the workplan graph; report plan defects back to the workplan skill.
---

# Task

This skill owns one task and its implementation. It never owns plan decomposition.

The ownership boundary is strict:

> **`task` evaluates work against the brief. `workplan` evaluates whether the brief and its place in the graph are correct.**

Resolve the active workplan directory first. If `docs/workplan/CURRENT` exists, read the generation it names under `docs/workplan/<generation>/`; otherwise use the legacy unversioned `docs/workplan/` directory.

Read the selected tracker entry, its brief, completed dependencies, applicable `AGENTS.md`, and relevant design documents before changing or judging implementation.

## Modes

### Implement

`/task implement <ID>` asks: implement this task exactly as specified and prove it is complete.

1. Confirm the task exists in the active tracker and has an executable status (`pending`, `in_progress`, or `blocked`). `superseded` and `removed` tasks are planning history, not executable work.
2. Verify every declared dependency is `completed`. Do not treat `superseded` or `removed` as successful execution prerequisites.
3. Inspect the current implementation before editing it.
4. Implement the objective and acceptance criteria in the brief.
5. Stay inside the declared **Write set** and **Contract surfaces** unless repository reality makes a deviation necessary. If it does, identify the deviation explicitly.
6. Preserve repository invariants that already hold.
7. Run every required automated verification and the brief's manual demonstration.
8. Report exactly what was observed.

If implementation reveals that the task is impossible, incorrectly decomposed, missing part of its completion condition, or overstuffed with unrelated future enablement, stop redesigning and report a **plan defect**. The workplan skill owns the repair.

### Audit

`/task audit <ID>` asks: does the implementation actually satisfy this task's authored brief?

Judge planning-contract conformance using the task's `workplan_version`. A completed V1 task is not malformed merely because V2 later required new brief fields.

Check at least:

1. **Objective** — the named observable behavior exists.
2. **Scope** — required work is present and unrelated work has not been absorbed.
3. **Runtime reachability** — the implementation is reachable exactly as the brief claims.
4. **Demonstration** — the manual demonstration works and produces the stated result.
5. **Acceptance criteria** — every criterion holds against repository state.
6. **Verification** — required automated, negative, adversarial, integration, and scenario evidence exists.
7. **Deferrals** — no hidden TODO, stub, or unowned deferral escaped the brief.
8. **Surfaces** — implementation stayed within declared write and contract surfaces, or deviations are explicit.
9. **Invariant preservation** — the task did not break repository invariants that already held.
10. **Status truthfulness** — `completed` is supported by implementation evidence.

Return `PASS` or `FAIL`, with blocking findings for a failure.

A task audit can pass while a workplan audit still fails. For example, an implementation can satisfy an authored HTTP-only V1 brief even when the current plan should have required the maintained CLI path in the same seam. That is a plan decomposition defect, not retroactive task non-conformance.

## V2 completion closure

For a V2 operator-visible capability, the task is incomplete unless the minimal maintained hands-on path needed to exercise it ships in the same task. Raw HTTP, database inspection, test harnesses, and temporary development commands do not substitute for the maintained operator surface.

Do not require unrelated future hardening or refactoring merely because it touches the same seam. Completion work belongs in the task; future enablement and independent revision work do not.

## Before finishing

- Dependencies were verified from the active tracker.
- The authored workplan version was respected.
- The manual demonstration was actually run for implementation or checked for audit.
- Required verification was run.
- Surface deviations, if any, are named.
- Any plan defect is reported rather than silently redesigned.
