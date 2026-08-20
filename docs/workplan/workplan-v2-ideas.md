# Workplan V2 ideas

Status: design notes for a future revision of `skills/workplan/SKILL.md`. These ideas are intentionally non-normative until the skill and tracker schema are updated.

## Goals

Workplan V2 should make re-derivation and audit behavior explicit when a repository contains work authored under older planning rules. It should distinguish historical task conformance from current repository invariants, represent legitimate revision/refactor ordering, and separate merge contention from semantic contract contention.

The resulting plan should remain truthful at HEAD. Historical deficiencies must not be hidden behind versioning: if older work leaves a current invariant unsatisfied, the re-derived plan must contain explicit tasks that repair that invariant.

## 1. Version the workplan contract per task

`schema_version` only describes how to parse the tracker. It does not identify which workplan rules a task was authored under.

Add a separate per-task workplan version, for example:

```json
{
  "id": "P1-049",
  "workplan_version": 2
}
```

A mixed-generation plan is expected after re-derivation. Completed historical tasks may remain at an older workplan version while pending tasks are rewritten to the current version.

### Audit rule

Evaluate a task brief against the workplan specification recorded for that task. Do not retroactively fail an immutable completed task because a later workplan version introduced new required brief fields or dependency vocabulary.

This versioning applies to task/brief conformance only. It does not waive repository-level invariants at HEAD.

For example, a completed V1 task may legitimately lack V2 `Write set`, `Contract surfaces`, or `Runtime reachability` fields. But if the repository at HEAD is not deployable, lacks the standing scenario suite, or otherwise fails a current invariant, the current plan is still incomplete.

### Historical invariant repair

Re-derivation must explicitly compare current HEAD against all current workplan invariants.

For every invariant that is unsatisfied because historical work did not establish it, no longer establishes it, or established only part of it, the re-derived plan must add one or more current-version tasks whose objective is to repair that invariant.

A historical-version annotation is never sufficient by itself. Versioning explains why an old task is not malformed; a repair task explains how the current repository will become conformant.

Each invariant-repair task should identify:

- the failed invariant;
- the evidence at HEAD showing the gap;
- the runtime or operator behavior that will become observable;
- the current-version acceptance criteria that prove the invariant is restored;
- dependencies required to make the repair meaningful.

This gives the agent a useful distinction:

- **legacy-conformant task**: correctly represents work completed under an older workplan contract;
- **current invariant gap**: repository behavior that still requires new work;
- **repair task**: explicit current-version work that closes that gap.

## 2. Separate write collisions from contract collisions

V1 overloads `Owned surfaces` and collision handling. File or directory overlap is useful for predicting merge contention, but it is not a complete definition of architectural collision.

V2 should model at least two kinds of concurrency risk.

### Write collision

Two concurrently available tasks have a write collision when their declared concrete write sets overlap in a way likely to require merge or rebase coordination.

Examples:

- the same source file;
- a directory declaration that intentionally covers the same files as another task;
- the same generated document or schema file.

The declaration should be explicit and narrow enough to be useful:

```text
Write set:
- cmd/winch/credential.go
- internal/application/credential.go
```

A broad directory should not be used merely as shorthand when the actual expected files are known.

### Contract collision

Two tasks have a contract collision when they may independently change the meaning of the same externally significant contract or namespace, even if they edit different files.

Examples:

- the same HTTP resource or operation;
- the same OpenAPI component;
- the same persisted aggregate or state transition;
- the same configuration key;
- the same driver or registry namespace;
- the same migration slot or schema ownership boundary.

Example:

```text
Contract surfaces:
- API: /api/v1/credentials
- RunSpec.credential_refs
- migration namespace: credentials
```

Gate reporting should list write collisions and contract collisions separately. Neither kind automatically creates a dependency edge. The planner may instead narrow surfaces, coordinate ownership, or deliberately serialize tasks when simultaneous work would be unsafe.

The important distinction is:

- write sets model likely code/document merge contention;
- contract surfaces model semantic coordination risk.

## 3. Add a revision/refactor dependency reason

The current dependency vocabulary does not cleanly represent work whose purpose is to revise an implementation produced by another task.

Add a fourth legitimate dependency reason. `revision` is preferable to the narrower name `refactor` because it also covers hardening, cleanup, performance work, test restructuring, and other deliberate follow-on changes.

Proposed dependency reasons:

- `compile`
- `contract`
- `semantic`
- `revision`

### Revision edge

A revision dependency means: the dependent task intentionally changes, hardens, restructures, optimizes, or tests implementation that is itself created or substantially rewritten by the prerequisite task. Performing the revision first would target the wrong or transient implementation.

Example:

```text
P1-061 -> P1-049
revision: hardens process-state handling and tests in the local runner implementation substantially introduced by P1-049.
```

This must not become a generic ordering escape hatch.

A revision edge is valid only when the dependent task names the implementation or test surfaces produced or substantially changed by the prerequisite and its objective is specifically to revise those surfaces.

Invalid rationale remains:

- easier to do later;
- avoids possible rework in general;
- same subsystem;
- nicer development order;
- prerequisite might influence the design.

The edge says "the object of this task does not meaningfully exist in its intended form before that task," not merely "we prefer this order."

## 4. Make re-derivation disposition every unfinished task

Re-derivation should not simply append gap-closure tasks around an existing pending plan. Every non-completed task from the previous plan must receive an explicit disposition under the new workplan version.

For each task whose status is not `completed`, the re-derivation must choose exactly one outcome:

1. **Rewrite** — the work is still required, but its brief, dependencies, surfaces, demonstration, acceptance criteria, or decomposition must be updated to the new plan.
2. **Split** — the old task is too broad or combines concerns that should now be independently schedulable. Replace it with narrower current-version tasks and record the replacement relationship.
3. **Merge/replace** — the work is still required but is better represented by another task. Record which task supersedes it.
4. **Remove** — the work is no longer required because the architecture, implementation, or evidence at HEAD makes it obsolete. Record a reason.

A re-derivation is incomplete while any pre-existing non-completed task has no disposition.

### Tracker semantics

Because task IDs are append-only, removal should not mean silently deleting historical task identity from the tracker.

V2 should add an explicit terminal planning disposition such as `superseded`/`removed`, or separate execution status from planning disposition. The exact schema needs design, but the historical record must remain machine-readable.

For example, a superseded task should be able to point at its replacements:

```json
{
  "id": "P2-023",
  "status": "superseded",
  "superseded_by": ["P2-062", "P2-063"],
  "workplan_version": 1
}
```

A removed task should record why no replacement is required.

Reusing an old task ID with materially different scope should be avoided when it would make historical completion or review evidence ambiguous. Minor brief refinement can retain the ID; semantic replacement should create new IDs and preserve the supersession relationship.

## 5. Re-derivation procedure in V2

A V2 re-derivation should be an explicit reconciliation procedure rather than a free-form rewrite.

Suggested sequence:

1. Record the current workplan specification version.
2. Read HEAD and determine which repository-level invariants currently pass or fail.
3. Add explicit current-version repair tasks for every failed invariant not already owned by adequate unfinished work.
4. Inventory every existing non-completed task.
5. Give each unfinished task exactly one disposition: rewrite, split, replace/merge, or remove.
6. Rewrite all surviving unfinished briefs to the current workplan version.
7. Re-evaluate dependency edges using the current reason vocabulary, including `revision`.
8. Declare concrete write sets and semantic contract surfaces separately.
9. Recompute availability, critical paths, width, write collisions, and contract collisions from the resulting graph.
10. Verify every deferral has an owning task or an explicit architectural decision trigger.
11. Validate the tracker schema and ensure README summaries agree with machine-readable state.
12. Record a re-derivation changelog containing invariant repairs and the disposition of every inherited unfinished task.

### Required re-derivation report

The changelog/report should answer at least:

- Which HEAD invariants were failing before re-derivation?
- Which tasks now repair each failure?
- Which inherited unfinished tasks were rewritten?
- Which were split, superseded, merged, or removed, and why?
- Which dependency reasons changed?
- What are the new critical paths and average widths?
- What write collisions exist?
- What contract collisions exist?
- Are any legacy-version tasks relevant to current invariant gaps?

This makes re-derivation auditable instead of allowing old pending tasks to survive accidentally.

## 6. Two layers of conformance

V2 should explicitly distinguish two layers.

### Repository invariants at HEAD

These are current truths about the system and cannot be grandfathered by a historical task version. Examples include running/deployable composition, controllable fake profile, standing scenarios, maintained operator/debug surface, and owned deferrals.

If one fails, the plan must contain an explicit path to restore it.

### Planning-contract conformance

These are rules for how tasks and briefs are represented: brief shape, dependency vocabulary, declared surfaces, demonstrations, acceptance criteria, and tracker fields.

A task is judged using the workplan version under which it was authored or last materially rewritten.

This prevents two opposite errors:

- falsely declaring old completed work malformed because the planning language evolved;
- using legacy status to excuse a repository that still violates current invariants.

## 7. Enforce completion closure without overstuffing seams

V2 should prevent two opposite planning failures:

1. **under-scoping a capability** by postponing part of its completion condition to a follow-up task;
2. **over-scoping a seam** by pulling future refactors, hardening, and extension scaffolding into the task merely because they are nearby.

### Completion closure

A task may not delegate part of its own completion condition to a later task.

For an operator-visible capability, completion includes the minimal maintained hands-on path needed to exercise that capability against the running deployment.

Do not split a capability into:

- one task that implements the internal or HTTP behavior; and
- a later task whose main purpose is to make that same behavior operable through the maintained CLI.

If the CLI command is the ordinary maintained hands-on way to exercise the capability, it belongs in the same task that introduces the capability.

Raw HTTP calls, direct database inspection, internal test harnesses, or temporary development commands may support verification, but they do not justify deferring the maintained operator path.

A later CLI task is legitimate only when it adds genuinely new operator capability or hardens an already-complete operator surface, for example richer output modes, scripting guarantees, diagnostics, or commands for behavior introduced by that later task.

### Strengthen the seam definition

A V2 seam should leave a complete vertical slice through the running system. It should contain:

- the port or boundary;
- the implementation needed for the current runtime profile;
- composition-root wiring;
- the public interaction surface;
- the minimal maintained operator/debug command needed to drive it by hand.

Demonstration: a person drives the new behavior through the maintained operator surface against the running deployment.

A seam is not complete if a later task is required merely to expose or operate behavior already implemented by the seam.

### Strengthen the hands-on invariant

I5 should become stronger than "a CLI exists early." It should require immediate reachability of new operator-visible capabilities:

> **I5 — Every new operator-visible capability is immediately reachable through the maintained hands-on surface.**
>
> The maintained CLI is not a follow-up integration layer. When a task introduces a new operator-visible capability, that same task adds the minimal CLI path required to exercise it. Later CLI tasks may improve ergonomics or expose genuinely new capabilities, but may not postpone basic operability of an already-completed seam.

### Keep completion work separate from future enablement

A seam task contains everything necessary for its objective to be true, but nothing whose primary purpose is to make later work easier.

Before adding a scope item to a seam, ask:

1. Is this required for the task's observable demonstration?
2. Is it required to preserve a repository invariant at task completion?
3. Would the capability be incomplete or misleading without it?

If all three answers are no, the item normally does not belong in the seam.

In particular, extract work whose primary purpose is:

- reducing future merge collisions;
- restructuring files for future parallelism;
- introducing extension points for later tasks;
- adding policy hooks with permissive or no-op behavior for future policies;
- hardening restart or failure behavior beyond the task's stated capability;
- adding observability not required to prove the capability;
- refactoring a working implementation into a more extensible form.

Such work should become a separate capability, hardening, or revision task as appropriate.

### Split test

When considering splitting one proposed task into two, inspect the dependency between them.

The split is invalid if the second task's objective is substantially:

> make the capability introduced by the first task actually usable through the maintained operator surface.

In that case, merge the operator path into the first task.

The split is valid when the second task introduces a distinct observable property, such as:

- a new capability;
- a new substrate;
- failure or recovery guarantees;
- performance or security guarantees;
- a revision of an already-working implementation;
- operator ergonomics beyond what is required to drive the existing capability.

### Overstuffed seam test

After drafting a seam brief, classify every scope bullet as one of:

- required behavior;
- required wiring;
- required operator reachability;
- required invariant preservation;
- future enablement;
- hardening;
- refactor/revision.

A seam should normally contain only the first four.

Future enablement, hardening, and revision work should be extracted unless removing it would make the seam's objective false or violate an invariant at task completion.

### Audit smells

An audit should flag a likely invalid split when:

- task A introduces an application/API capability;
- task B depends directly on A;
- task B's main purpose is to expose that same capability through the maintained CLI.

This usually means task B is part of task A's completion condition.

An audit should also flag a likely overstuffed seam when scope bullets exist mainly to prepare unrelated later tasks. Typical examples are pre-emptive API-file restructuring, extension registries, permissive policy hooks, unrelated metrics, or failure-mode hardening that can be demonstrated independently.

The combined rule is:

> Do not under-scope the seam by postponing basic operability. Do not over-scope the seam by pulling future refactors and hardening into it.

## 8. Split plan reasoning from task execution

The current workplan skill combines two different units of reasoning: the implementation plan as a graph and one executable task. V2 should split these responsibilities into two skills with an explicit ownership boundary.

### `workplan` skill

The `workplan` skill reasons about the plan as a system.

Representative modes:

```text
/workplan create
/workplan extend
/workplan rederive
/workplan audit
```

It owns:

- phase structure and task decomposition;
- task shapes and completion closure;
- dependency graph and dependency reasons;
- repository-invariant coverage;
- workplan generations;
- write and contract collision analysis;
- critical paths and parallel width;
- deferrals and task ownership;
- re-derivation and replacement of incorrect task boundaries;
- whether a task should be merged, split, superseded, rewritten, or removed.

`/workplan audit` asks:

> Is this a truthful, coherent, correctly decomposed implementation plan for the repository at HEAD?

It may inspect implementation evidence to answer that question, but it does not audit whether a particular implementation satisfies one task brief.

### `task` skill

The `task` skill reasons about exactly one task and its implementation.

Representative modes:

```text
/task implement P1-050
/task audit P1-050
```

It reads the active workplan generation, selected task brief, completed dependencies, relevant design baseline, and the implementation or pull request under review.

It does not redesign the task graph.

#### Task implementation

`/task implement <ID>` asks:

> Implement this task exactly as specified and prove it is complete.

It should:

- verify declared dependencies are complete;
- inspect the existing implementation before changing it;
- implement the task objective and acceptance criteria;
- stay within declared write and contract surfaces unless reality makes a deviation necessary;
- run required automated verification;
- run the task's manual demonstration;
- preserve repository invariants that were already true;
- report any mismatch between the brief and repository reality.

If implementation reveals that the task is impossible, incorrectly decomposed, overstuffed, missing part of its completion condition, or otherwise malformed, the task skill must report a **plan defect**. It must not silently redesign the workplan as implementation freedom.

Plan defects belong to the `workplan` skill.

#### Task audit

`/task audit <ID>` asks:

> Does the implementation actually satisfy this task's authored brief?

The audit should check at least:

1. **Objective** — the named observable behavior exists.
2. **Scope** — required work is present and unrelated scope has not been absorbed.
3. **Runtime reachability** — the implementation is reachable exactly as the brief claims.
4. **Demonstration** — the documented manual demonstration works and produces the stated observable result.
5. **Acceptance criteria** — every criterion is checked against repository state, not inferred from commit messages or intent.
6. **Verification** — required automated, negative, adversarial, integration, or scenario evidence exists as specified.
7. **Deferrals** — no hidden TODO, stub, or unowned deferral escaped the brief.
8. **Surfaces** — implementation stayed within declared write and contract surfaces, or deviations are identified explicitly.
9. **Invariant preservation** — the task did not break repository invariants that were already established.
10. **Status truthfulness** — a `completed` status is supported by implementation evidence.

The result should be an explicit task-level verdict such as `PASS` or `FAIL`, with blocking findings when it fails.

### Ownership boundary

The central distinction should be explicit in both skills:

> **`task` evaluates work against the brief. `workplan` evaluates whether the brief and its place in the graph are correct.**

This means the two audits may legitimately reach different conclusions.

For example, `/task audit P1-050` may find that an implementation correctly satisfies the authored HTTP-only task brief, while `/workplan audit` may still find that P1-050 is incorrectly decomposed because a required operator completion condition was deferred into another task.

Neither skill should absorb the other's responsibility:

- `task` does not repair plan decomposition while implementing or auditing a task;
- `workplan` does not substitute a graph audit for implementation evidence that belongs to a task audit.

## 9. A failing invariant is a task, not a finding

Extracted from §1's *Historical invariant repair* and step 3 of §5, both of
which are otherwise set aside. This idea stands on its own: it needs no per-task
versioning and no re-derivation procedure.

The seven invariants hold of the plan at HEAD, and HEAD moves. An invariant
established once can stop being true — a composition root gains a seam nothing
registers into, a fake profile rots, the standing suite is quietly excluded from
CI, a deferral's owning task is closed without the work being done.

Today the gate procedure says to check I1–I7 against the real repository, and
then says nothing about what happens when one fails. A plan can record that the
system no longer deploys and open the next phase anyway.

The rule: whenever an invariant is unsatisfied — never established, no longer
established, or only partly established — the plan gains one or more tasks whose
objective is to restore it. Recording the failure in an audit, a changelog, or a
report does not discharge it. Only a task with an ID does. This is I6's
discipline applied to invariants rather than to deferrals.

Beyond the ordinary brief anatomy, a repair task would name:

- which invariant failed;
- the evidence at HEAD showing the gap — a command and its output, or a
  `file:line`, not a description of the shortfall;
- what becomes observable once it is restored;
- the acceptance criteria that prove restoration, checkable against the
  repository;
- the dependencies without which the repair would be hollow.

Where it would land if adopted:

- a subsection of the shared model closing *The seven invariants*;
- a gate-procedure step after the invariant check, so a noted failure does not
  close the gate;
- audit mode, where an unowned invariant failure becomes a blocking finding
  rather than an observation;
- the task skill, where an invariant failing for reasons older than the task
  under audit is reported as a plan defect rather than folded into that task's
  verdict.

### Open questions

- Do repair tasks take the ordinary four shapes, with the shape following from
  what becomes observable, or does repair want a fifth shape? Forcing "the
  system deploys again" into *capability* is defensible — the behavior is
  genuinely absent at HEAD — but it reads oddly beside a first-time capability.
- Does a repair task get an edge to whatever broke the invariant? A `revision`
  edge fits when one task's implementation caused the regression, but an
  invariant that was never established has nothing to point at.
- Is "unfinished work already owns this invariant" a judgment the gate makes, or
  is the repair task opened regardless and closed as a duplicate?
- Does a repair task differ from an ordinary hardening task in anything but the
  reason it was written? If not, the idea is a rule about the gate rather than a
  new kind of task.

## Open schema questions

Before making V2 normative, decide:

- whether `workplan_version` is an integer, semantic version, or named contract version;
- whether the tracker gets a new `status` value such as `superseded` or gains a separate `disposition` field;
- how replacement relationships (`supersedes`, `superseded_by`) are represented;
- whether write sets support directories/globs or require concrete paths;
- the controlled vocabulary for contract-surface kinds;
- whether `revision` is the final dependency-reason name;
- whether a materially rewritten pending task retains its ID or is superseded by a new ID, and where the threshold is defined.

The core rule should remain simple: version history explains old work, re-derivation reconciles all unfinished work, current invariant gaps always produce explicit repair work, and every completed capability closes its operator-visible loop without absorbing unrelated future work.