# Workplan V2 implementation review — rejected

Status: **Rejected pending correction of the clean-slate boundary.**

The current branch is substantially closer to the intended Workplan V2 design. The move to versioned generations, generation-qualified persistent task identity, separate `workplan` and `task` skills, completion closure, write-versus-contract collisions, the `revision` dependency reason, and the isolated-repository re-derivation approach are all directionally correct.

The implementation should still not be merged. The remaining problems are concentrated in the re-derivation boundary itself: the deriving agent is still told that sanitization occurred, the sanitizer cannot operate successfully against the repository's real completed briefs, the advertised harvest validation is stronger than the code actually performs, and preparation is not safely recoverable when sanitization fails.

## Blocking findings

### 1. The deriving agent is still made aware of the sanitization procedure

The purpose of the clean-slate workflow is not merely to hide old task files. The deriving agent must not be told that an unfinished plan was removed, that task IDs were deliberately made reusable, that another historical generation exists, or that a sanitizer/harvester prepared its repository.

The isolated checkout currently retains `skills/workplan/SKILL.md` and `skills/workplan/contracts/v2.md`, and those files explicitly describe the clean-slate re-derivation process. They tell the agent that:

- a previous unfinished plan existed;
- unfinished tasks are deliberately excluded;
- old unfinished IDs may be reused;
- `scripts/prepare_workplan_rederivation.py` prepared the environment;
- `harvest` later copies the derived generation back;
- an older generation remains available elsewhere for history.

That defeats the intended context boundary. The orchestration mechanism is an operator/tooling concern, not part of the planner's instructions.

The workplan skill visible inside the isolated checkout should behave as though the repository naturally contains one active generation representing completed implementation history plus whatever remaining work the agent is about to derive. It should contain the normal V2 planning rules but no explanation of how that repository state was manufactured.

The clean-slate preparation and harvest procedure should instead be documented with the script or in operator-only documentation that is removed from the isolated checkout.

### 2. The sanitizer fails on legitimate completed history

`prepare_workplan_rederivation.py` currently derives a forbidden set from every non-completed task ID in the previous generation and rejects the isolated checkout if any of those IDs appear in any text file.

That rule is too strong for the repository's actual history.

Completed task briefs can legitimately contain historical references to work that was unfinished at the time the brief was written. For example, completed P1-049 names P1-050, P3-029, P5-041, and P5-042 in its non-goals, runtime notes, and deferrals. Because P1-049 is completed, its brief is copied into the new generation; because those referenced IDs are unfinished in V1, the leak detector then rejects the checkout.

The V2 skill itself also contains example task IDs that may collide with unfinished V1 IDs.

Therefore the current `prepare` path is expected to fail against the real repository even though the synthetic tests pass.

This reveals the correct distinction:

- **completed implementation facts** are valid derivation input;
- **historical future-plan assertions embedded in completed briefs** are not necessarily valid derivation input.

Do not solve this by weakening the leak detector until it silently permits contamination. Instead, define a clean completed-history representation for the new generation.

A reasonable approach is to generate a completed-work snapshot from each completed task that preserves implementation facts such as objective, implemented scope, demonstration, acceptance evidence, provenance, and task identity while omitting historical deferrals and references whose only purpose was to describe then-future work. The immutable original brief remains in the archived generation.

The exact representation may differ, but `prepare` must be proven against the real repository, including completed briefs that mention unfinished IDs.

### 3. `harvest` does not provide the fail-closed validation claimed by the V2 contract

The PR description and V2 workplan instructions state that harvest validates the complete resulting generation before copying it back.

The current `validate_generation()` implementation checks only a subset of those promises:

- tracker has a task list;
- task IDs are unique;
- completed records carried forward are unchanged;
- referenced brief files exist;
- dependency IDs exist.

It does **not** currently enforce all of the advertised boundary conditions, including:

- validation of `tasks.json` against the generation's `tasks.schema.json`;
- dependency-cycle detection;
- validity and reciprocity of `supersedes` / `superseded_by` relationships;
- terminal-status constraints encoded by the V2 schema.

The implementation and documentation must agree. Prefer strengthening `harvest` to perform the full validation already promised by the V2 contract rather than weakening the contract.

At minimum harvest must fail closed on:

1. JSON Schema validation;
2. duplicate task IDs;
3. missing brief files;
4. unknown dependency IDs;
5. dependency cycles;
6. invalid terminal and replacement relationships;
7. mutation or removal of completed historical records that are defined as immutable;
8. an invalid or missing active-generation marker if the generated workplan requires one.

Add regression tests for each category.

### 4. `prepare` mutates and commits the source repository before proving isolation succeeded

The preparation sequence currently archives V1, seeds V2, updates `CURRENT`, and commits those changes to the source repository before it exports and validates the isolated checkout.

If export, sanitization, leak detection, or isolated-repository initialization then fails, the source repository has already advanced into a partially prepared state.

That is not safely retryable. A subsequent invocation may see the newly created generation as current and attempt to advance again, producing V3 rather than completing the failed V2 preparation.

Preparation needs transactional or explicitly resumable semantics.

Acceptable designs include:

- construct and validate the prospective archive/new-generation tree in a temporary workspace first, then apply and commit it only after isolation succeeds;
- record a durable preparation state before mutation and make retries idempotently resume the same generation;
- on failure, automatically restore the source repository to the exact pre-prepare commit and remove the incomplete checkout.

Whichever approach is chosen, a failed `prepare` must not leave the operator with a silently advanced generation that changes the meaning of the next retry.

Add a regression test that forces a sanitization failure after generation construction and proves that the source repository is left in a well-defined retryable state.

### 5. Harvest does not protect against source-branch drift

The preparation state records the source baseline and archive commit, but `harvest()` does not verify that the source repository is still at the repository state against which the new workplan was derived.

That permits this sequence:

1. prepare V2 from repository state A;
2. derive a complete workplan in the isolated repository;
3. merge unrelated implementation changes into the source branch, producing state B;
4. harvest the workplan derived from A onto B.

A Workplan's central claim is that it is truthful at HEAD, so silently accepting this drift is unsafe.

Harvest should fail if the source branch has moved beyond the expected archive commit unless an explicit revalidation procedure proves the derived workplan against the new HEAD.

The simple initial rule should be:

```text
source branch HEAD == recorded archive commit
```

If later workflows need to support concurrent implementation during re-derivation, add a separate explicit rebase/revalidation path rather than making drift implicit.

Add a regression test that advances the source branch after `prepare` and confirms that `harvest` refuses the stale derivation.

## Test and CI coverage gap

The current tests correctly exercise several useful properties:

- the derived checkout is a separate Git repository;
- it has one root commit and no remote;
- the archived generation is absent;
- the orchestration script is absent;
- abandoned IDs can be reused in the new generation;
- harvesting replaces only the new generation while preserving V1.

Those tests are worth keeping.

However, the fixture's completed brief is effectively empty and therefore misses the repository's real historical-reference problem. The suite also needs explicit coverage for:

- completed briefs or completed-history records that historically mention unfinished tasks;
- schema-invalid derived trackers;
- dependency cycles;
- invalid replacement relationships;
- failed preparation and retry behavior;
- source-branch drift before harvest.

The current GitHub workflows are green, but that is not evidence that this Python regression suite is being run by CI. `make check` currently covers the Go/OpenAPI/build checks and does not include these Python tests. Either wire the re-derivation tests into an existing quality target/workflow or add a dedicated tooling test job.

## What should remain

The following design choices are correct and should be preserved:

- versioned workplan generations;
- immutable archived historical generations;
- generation-scoped short task IDs;
- generation-qualified persistent automation identity;
- separate `skills/workplan/SKILL.md` and `skills/task/SKILL.md` responsibilities;
- completion closure and immediate operator reachability;
- write collisions distinct from contract collisions;
- `revision` as an explicit dependency reason;
- isolated derivation in a new local Git repository with one root commit and no remote;
- harvesting only the newly derived generation rather than merging the isolated repository;
- preserving V1 and V2 normative planning contracts for version-aware audit.

The move from a mere sanitized branch to a history-free local repository is particularly important and should not be reverted. The defect is that the contents of that isolated repository still disclose the orchestration concept and currently cannot pass their own leak rules against real completed history.

## Acceptance conditions for a new review

This implementation can be reconsidered when all of the following are true:

1. **No orchestration awareness:** the deriving agent's checkout contains no instruction explaining that a previous unfinished plan was hidden, that sanitization occurred, that old IDs were intentionally freed, or that a harvest step exists.
2. **Real-repository preparation succeeds:** `prepare` succeeds against the actual repository state, including completed historical work that originally referred to unfinished tasks, without exposing obsolete future-plan decomposition to the deriving agent.
3. **Completed history has an explicit clean representation:** the new generation carries implementation facts from completed work without blindly importing historical future-plan assertions that contaminate re-derivation.
4. **Harvest validation matches the V2 contract:** schema validation, cycles, brief references, dependency references, terminal/replacement semantics, uniqueness, and immutable completed history are all enforced and tested.
5. **Preparation is safely retryable:** a failure after preparation begins cannot silently leave the source repository advanced to a new generation in a way that changes the next invocation's behavior.
6. **Source drift is rejected or explicitly revalidated:** harvest cannot install a plan derived from an older repository state onto a changed HEAD without a deliberate validation path.
7. **The Python tooling tests run in CI:** green repository checks include the clean-slate preparation/harvest regression suite.
8. **Existing V2 improvements remain intact:** generation-qualified automation identity, historical contracts, blocked-task semantics, completion closure, collision semantics, and the isolated-repository architecture are not regressed while fixing these issues.

Until those conditions hold, PR #40 should remain draft and unmerged.
