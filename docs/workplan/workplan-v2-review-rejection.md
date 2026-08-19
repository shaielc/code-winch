# Workplan V2 implementation review — rejected

Status: **Rejected pending redesign.**

The current branch contains useful Workplan V2 work, especially the split between `workplan` and `task`, completion closure, write-versus-contract collisions, and the `revision` dependency reason. However, the implementation currently violates core design decisions made for V2 and should not be merged in its present form.

## Blocking findings

### 1. Re-derivation is not clean-slate

The current implementation deliberately copies every unfinished task and its brief into the next workplan generation and requires the re-deriving agent to rewrite, supersede, or remove each inherited task.

That is explicitly not the intended design.

The purpose of generation-based re-derivation is to prevent the new planning agent from being anchored by mistakes in the previous unfinished plan. The re-deriving agent must not see abandoned unfinished tasks, their titles, their IDs, their briefs, their dependency graph, or a report explaining how they were dispositioned.

The preparation workflow should instead:

1. archive the current workplan as an immutable generation, for example `docs/workplan/v1/`;
2. create the next generation from completed implementation history only;
3. create a derivation branch containing only the new generation;
4. run re-derivation against HEAD, the design baseline, completed history, and the current workplan rules;
5. harvest only the newly derived generation back to the archive branch.

The previous unfinished plan remains available in the old generation for historical inspection, but it is not an input to the derivation agent.

`scripts/prepare_workplan_rederivation.py`, its tests, and `skills/workplan/SKILL.md` must be changed accordingly.

### 2. Generation-scoped task IDs are not reflected in persistent automation identity

The design allows task IDs to be scoped by workplan generation. An abandoned `v1/P1-050` may therefore be reused as a different `v2/P1-050`.

The current scheduler and completion tooling still persist and resolve tasks using the bare ID `P1-050`. This makes stale state from an older generation capable of affecting a new task with the same ID.

Persistent identities must include the generation, for example:

```text
v2/P1-050
```

This applies at minimum to:

- scheduler state keys;
- persisted dispatch metadata;
- PR/task completion identity;
- any automation record that survives a generation change.

Human-facing commands such as `/task implement P1-050` may continue to use the short form when operating inside the active generation, but persistent automation must not.

### 3. Harvest validation is incomplete

The re-derivation helper is intended to be the safety boundary before a newly derived generation is copied back, but its validator currently accepts structurally invalid plans.

For example, a replacement task can reference a brief path that does not exist and still pass the current validator.

Harvest must fail closed on at least:

- tracker validation against the generation JSON Schema;
- duplicate IDs;
- missing brief files;
- unknown dependencies;
- dependency cycles;
- invalid terminal/replacement relationships;
- mutation of completed historical task records where those records are intended to be immutable.

The validator should validate the actual resulting workplan, not only disposition metadata.

## Additional required changes

### 4. Preserve normative historical planning contracts

`workplan_version` only works if the corresponding historical instructions remain available.

Replacing `skills/workplan/SKILL.md` with V2 while retaining tasks marked `workplan_version: 1` leaves future `/task audit` operations without a normative V1 contract to evaluate those tasks against.

Preserve immutable contract versions, for example:

```text
skills/workplan/contracts/v1.md
skills/workplan/contracts/v2.md
```

or archive the applicable planning contract inside each workplan generation.

A task's `workplan_version` must resolve to actual instructions.

### 5. `blocked` tasks are not executable

`skills/task/SKILL.md` currently treats `pending`, `in_progress`, and `blocked` as executable statuses.

A blocked task should not be implemented until its blocker is resolved. `/task implement` should:

- accept `pending` tasks;
- allow continuation of an appropriate `in_progress` task;
- report the blocking reason and stop for `blocked` tasks;
- reject `completed`, `superseded`, and `removed` tasks as implementation targets.

This should agree with scheduler behavior, which already dispatches only `pending` tasks.

## What should remain

The following parts of the implementation are conceptually correct and should be retained while addressing the blockers:

- separate `skills/workplan/SKILL.md` and `skills/task/SKILL.md`;
- `/workplan audit` evaluates the graph, decomposition, coverage, collisions, and invariants;
- `/task audit <ID>` evaluates one implementation against its authored brief;
- `task` reports plan defects rather than silently redesigning the workplan;
- completion closure for operator-visible capabilities;
- separation of write collisions from contract collisions;
- the `revision` dependency reason;
- active workplan generation resolution through `docs/workplan/CURRENT`;
- harvesting only the newly derived generation rather than merging the sanitized derivation branch.

## Acceptance conditions for a new review

This implementation can be reconsidered when:

1. the clean derivation branch contains no unfinished task information from the previous generation;
2. new-generation derivation starts from completed history only;
3. persistent task automation identity is generation-qualified;
4. harvest validates the complete generated workplan and fails closed on invalid structure;
5. historical workplan contracts remain available for task-version-aware auditing;
6. task execution semantics correctly reject blocked tasks;
7. regression tests explicitly prove the clean-slate property and generation-ID isolation.

Until those conditions hold, the branch should remain unmerged.
