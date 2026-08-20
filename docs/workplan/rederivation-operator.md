# Workplan re-derivation operator procedure

This document is for the person or automation preparing a planning checkout. It is not part of the planner's instruction set.

## Prepare

Run:

```sh
python3 scripts/prepare_workplan_rederivation.py prepare --checkout <empty-directory>
```

Preparation is transactional with respect to the source repository. The tool constructs the prospective next generation in temporary storage, creates the isolated local Git repository, removes operator-only planning history, and validates that no unfinished task IDs remain visible before it commits the archived/new-generation state in the source repository. If isolation fails, the source branch is restored to the exact pre-prepare commit and the incomplete checkout is removed.

The isolated repository has one root commit and no remote. Its active workplan contains clean completed implementation records rather than copies of old briefs: objective, implemented scope, runtime reachability, demonstration, verification, acceptance, and trace facts are retained, while obsolete future-plan assertions such as historical non-goals and deferrals are omitted.

## Derive

Run the planning agent inside the isolated repository. The agent uses `skills/workplan/SKILL.md` normally.

## Harvest

After the derived repository is clean and committed, return to the source repository and run:

```sh
python3 scripts/prepare_workplan_rederivation.py harvest
```

Harvest refuses to proceed unless the source branch HEAD still equals the archive commit recorded by `prepare`. It validates the isolated active generation against its JSON Schema and additionally checks unique IDs, brief existence, dependency references and cycles, replacement reciprocity, terminal relationships, the active-generation marker, and immutable completed implementation records. Only the active generation directory is copied back; the isolated repository is never merged.

If the source branch changed after preparation, derive again from the newer repository state. Supporting concurrent source changes requires an explicit future revalidation workflow rather than implicit acceptance of drift.
