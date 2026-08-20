# Workplan V2 implementation review — second rejection

Status: **Rejected pending correction of historical truth and contract-boundary issues.**

This review supersedes the unresolved conclusions of `workplan-v2-review-rejection.md`. The earlier implementation defects around clean-slate isolation, transactional preparation, harvest validation, source drift, and CI coverage have largely been corrected. The current design direction remains sound.

The implementation should still not be merged. The remaining blockers are now concentrated in the semantic boundary between historical work, current planning, and the sanitized derivation repository.

## What is now correct

Preserve these choices:

- one active `skills/workplan/SKILL.md` rather than separate executable V1/V2 planner skills;
- immutable planning contracts under `skills/workplan/contracts/v1.md` and `v2.md`;
- separate `workplan` and `task` skills;
- versioned workplan generations;
- generation-scoped short task IDs;
- generation-qualified persistent automation identities such as `v2/P1-050`;
- completion closure and immediate operator reachability;
- write collisions separated from contract collisions;
- `revision` as a legitimate dependency reason;
- isolated derivation in a separate local Git repository with one root commit and no remote;
- transactional preparation and retry-safe rollback;
- harvest of only the newly derived generation;
- source-drift rejection before harvest;
- schema, dependency, replacement, and completed-history validation during harvest;
- dedicated CI coverage for the workplan tooling.

The contract versioning model is also correct: **version the planning contract, not the planner implementation.** `SKILL.md` describes how the planner operates today; `contracts/v1.md` and `contracts/v2.md` define the rules under which historical and current tasks are judged.

## Blocking findings

### 1. Completed implementation records are derived from plans, then treated as facts

The preparation script creates each carried-forward completed record by extracting sections from the historical authored brief, including `Objective`, `Scope`, `Runtime reachability`, `Demonstration`, `Verification`, `Acceptance criteria`, and `Traces to`.

The resulting document is then labelled a completed implementation record, and the Workplan skill tells the re-deriving agent to treat completed implementation records as established facts.

That inference is invalid.

An authored brief records what a task intended to accomplish. It is not evidence that every scope item, verification claim, demonstration, or acceptance criterion was actually satisfied by the implementation at HEAD. This repository has already shown why that distinction matters: completed tasks have later required fixes after verification against current repository state.

The new generation must not convert historical planning assertions into implementation truth merely because the task status is `completed`.

Required correction:

- treat the carried-forward record as **historical implementation context or provenance**, not authoritative proof of repository behavior;
- make HEAD the authority for current behavior and invariant truth;
- word the Workplan skill accordingly: completed records may guide inspection, but current claims must be checked against HEAD;
- avoid language such as "established facts" when those records are synthesized from the authored brief.

A useful distinction is:

- **historical task record** — what work was intended/completed under an older generation;
- **repository fact at HEAD** — what inspection, tests, runtime behavior, and current artifacts prove now.

The planner may use the former to understand provenance. It must derive the latter from the repository.

### 2. Historical task audit loses the actual authored brief

The `task` skill correctly says that `/task audit <ID>` evaluates implementation against the task's authored brief and resolves the task's historical `workplan_version` to the corresponding immutable contract.

But after V1 completed tasks are carried into V2, their `brief` field points at the synthesized completed-history record rather than the actual V1 brief stored in the archived generation.

That means the active-generation task entry no longer identifies the authored brief that the task skill claims to audit.

This breaks the historical audit model.

Required correction:

- preserve explicit provenance for carried-forward completed tasks, for example:
  - `origin_generation: "v1"`
  - `origin_brief: "phase-1/p1-049-implement-the-local-runner.md"`
- or define an equivalent machine-readable resolution mechanism;
- make `/task audit` resolve historical tasks to the original archived authored brief and original contract;
- do not substitute a synthesized completed-history summary for the authored brief during task audit.

The long-term identity should be unambiguous as `<generation>/<ID>`.

A historical audit such as `v1/P1-049` should remain possible even after V2 becomes active.

### 3. Legacy unversioned tasks have no explicit `workplan_version` compatibility rule

The current V1 tracker on `main` predates Workplan V2 and therefore does not contain a `workplan_version` field on each task.

The new Workplan and Task skills, however, instruct agents to resolve `workplan_version: N` to `skills/workplan/contracts/vN.md`.

This creates an ambiguous compatibility interval after the V2 skill/tooling merge but before the first generation conversion.

Required correction:

Add one explicit compatibility rule everywhere this is resolved:

> A task in the legacy unversioned `docs/workplan/tasks.json` tracker with no `workplan_version` field is authored under Workplan V1.

Do not require an immediate data migration merely to make historical tasks auditable.

### 4. The isolated checkout rewrites repository/design context instead of only removing planning contamination

The preparation script currently scans text outside the workplan and rewrites unfinished task IDs to phrases such as `later work` in Markdown/RST and Go comments.

This is too broad.

The re-deriving agent is supposed to inspect HEAD and the design baseline. Rewriting design prose, operational docs, architecture notes, or source comments creates a repository that is no longer simply HEAD with the obsolete pending plan removed.

Worse, replacing `P2-059` with `later work` can preserve the semantic contamination while hiding the identifier. A sentence such as:

> admission policy is handled by P2-059

can become:

> admission policy is handled by later work

The old decomposition is still influencing the planner.

Required correction:

- sanitize the **planning layer**, not arbitrary repository text;
- keep design documents and implementation artifacts byte-for-byte identical to the source baseline unless a specifically identified planning-only annotation must be removed;
- use a narrow explicit allowlist of operator/planning-only files that may be stripped;
- if an unfinished task identity leaks into an ordinary design or implementation artifact, fail preparation and require an explicit source correction instead of automatically rewriting the text;
- do not treat lexical removal of task IDs as sufficient proof that semantic planning contamination is gone.

The isolated repository should be truthful enough that "inspect HEAD and the design baseline" still means what it says.

### 5. Harvest validates against a schema that the deriving agent can modify

`validate_generation()` reads `tasks.schema.json` from the isolated generation and validates `tasks.json` against that file.

But that schema is part of the planner's writable checkout.

The deriving agent can accidentally or intentionally weaken the schema and then pass validation against its modified version. This defeats the purpose of fail-closed harvest validation.

Required correction:

- record the expected V2 schema hash during `prepare` and require the generation's `tasks.schema.json` to match exactly during `harvest`; or
- validate against the immutable source-branch V2 schema and separately require the copied generation schema to be byte-identical to it.

The planner may write the task graph. It must not be able to redefine the tracker contract that validates that graph.

## High-priority planning-contract findings

These are not separate architecture failures, but they should be corrected before merge because they encode the behavior Workplan V2 was created to protect.

### 6. The normative V2 contract compresses the completion-closure rules too far

The active V2 skill correctly states the core ideas:

- a task may not delegate part of its own completion condition;
- the maintained operator path belongs in the same operator-visible capability;
- seams should not absorb unrelated future refactors or hardening.

However, several concrete tests from the V2 design notes were dropped when making the contract normative. Those tests are valuable because they make the rule operational rather than aspirational.

Restore explicit guidance equivalent to:

#### Split test

A split is probably invalid when task B's main objective is:

> make the capability introduced by task A usable through the maintained operator surface.

In that case, B is normally part of A's completion condition.

#### Seam scope test

Before putting work into a seam, ask:

1. Is it required for the observable demonstration?
2. Is it required to preserve a repository invariant at task completion?
3. Would the capability be incomplete or misleading without it?

If all three are no, it normally does not belong in the seam.

#### Overstuffed-seam classification

Classify proposed scope as:

- required behavior;
- required wiring;
- required operator reachability;
- required invariant preservation;
- future enablement;
- hardening;
- refactor/revision.

A seam should normally contain only the first four.

Also retain the concrete audit smell that motivated this work:

- task A introduces the API/application capability;
- task B depends directly on A;
- task B mainly exposes the same capability through the maintained CLI.

That is usually one capability split incorrectly across two tasks.

### 7. Generation cutover needs an explicit quiescence rule

The automation correctly namespaces task identities by generation, but preparing a new active generation while old-generation implementation work is still in flight creates avoidable ambiguity.

For example, an existing V1 implementation pull request may still carry V1 identity while the repository has already switched `CURRENT` to V2.

Source-drift rejection protects harvest correctness, but it does not define the operational cutover semantics.

Required correction:

Document and, where practical, enforce a generation cutover rule:

- pause scheduling of the old generation;
- ensure no old-generation task is currently leased/in progress;
- ensure no implementation pull request expected to complete against the old active tracker is still in flight, or define an explicit exception path;
- only then run `prepare` and make the new generation active.

This belongs in operator tooling/procedure, not in the deriving planner's context.

## Documentation cleanup

Two branch documents are now stale enough to be misleading:

1. `workplan-v2-ideas.md` still contains parts of the earlier model in which every inherited unfinished task is dispositioned inside the new graph and IDs are globally append-only. The implemented generation model deliberately hides old unfinished work from derivation and scopes IDs by generation.
2. `workplan-v2-review-rejection.md` still says PR #40 must remain rejected for several concrete script defects that have since been fixed.

Keep historical review documents if useful, but mark them explicitly superseded by this review or move them into a clearly historical design/review area. Do not leave contradictory documents looking simultaneously normative.

## Acceptance conditions for another review

This implementation can be reconsidered when all of the following are true:

1. **Historical records are not asserted as implementation truth.** HEAD remains authoritative for current repository behavior and invariants.
2. **Historical authored briefs remain resolvable.** Carried-forward completed records preserve provenance to their original generation and brief, and task audit uses that original brief.
3. **Legacy compatibility is explicit.** Missing `workplan_version` in the unversioned legacy tracker resolves to Workplan V1.
4. **The isolated repository preserves HEAD/design truth.** Broad lexical rewriting of ordinary repository/design files is removed; sanitization is limited to planning/orchestration contamination.
5. **The V2 schema is immutable during derivation.** Harvest validates against a trusted schema or exact trusted hash, not a planner-modifiable contract.
6. **Completion-closure tests are normative.** The split test, seam-scope test, overstuffed-seam test, and API→CLI audit smell are restored to the V2 contract or skill.
7. **Generation cutover is quiescent.** The operator procedure defines how old-generation in-flight work is drained before activating the next generation.
8. **Existing improvements remain intact.** Do not regress transactional preparation, isolated one-root/no-remote derivation, source-drift rejection, fail-closed harvest validation, generation-qualified automation identity, separate skills, historical contracts, collision semantics, or CI coverage.

Until these conditions hold, PR #40 should remain draft and unmerged.
