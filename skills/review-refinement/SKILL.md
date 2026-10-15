---
name: review-refinement
description: Review one refined task brief against its original by answering four questions — did the refinement find real problems, is it right-sized, did it grow scope, and is its value worth its cost. Use when a refinement pull request is open and before it merges into the task branch and implementation starts.
---

# Review refinement

Work on one refinement PR (refinement branch → task branch) for one task ID, as
someone other than its refiner. Read applicable `AGENTS.md` files,
[the model](../shared/workplan-model.md) (brief anatomy and shapes), and
[`refine-task`](../refine-task/SKILL.md), the rules the refiner worked under.

Judge the difference between two versions of one brief, `docs/workplan/<brief>`:
the **original** at the PR's base commit and the **refined** at its head. Check
the refiner's claims against the code at HEAD, not against the PR description.
Explain any complex read-only command before running it. The review changes no
file.

## Orient

1. Record the PR, its base and head SHAs, and the task ID and brief path.
2. Confirm the diff touches only the brief and no tracker status. Anything else
   is a Question 3 finding.
3. List what changed, by section: Scope, Non-goals, Write set, Acceptance
   criteria, Verification, Demonstration, Deferrals — and any change to
   Objective, Shape, Dependencies, Traces to, or Contract surfaces.
4. Read what the PR description says the refinement found.

## Questions

Answer each with evidence: a command with its observed output, or a `file:line`.
A claim with neither is not shown.

### 1. Did it find problems with the original brief?

For each problem the refinement fixes or reports — a wrong assumption, an
ownership conflict, a missing criterion, a demonstration that cannot run —
check it at HEAD and label it:

- **Confirmed** — the original was wrong and the evidence shows it.
- **Not shown** — no evidence; the change is a guess.
- **Preference** — the original was workable and the change restyles it.

Then spot-check the two or three claims in the original that would cost most if
wrong, such as a dependency edge or the demonstration commands, and say whether
the refinement left a real problem in place. A refinement that finds nothing
because the brief was sound is a valid result.

### 2. Is it too verbose or too sparse?

Count lines per section, original against refined, in a small table. Then read
for:

- **Verbose** — restates design documents or the model; narrates what was
  checked instead of stating the result; carries boilerplate that belongs in the
  phase README; dictates detail the implementer sees in the code anyway.
- **Sparse** — Scope bullets still read as predictions; no place where the code
  lives; a criterion without an observable result; a demonstration that cannot
  run where it will run; a judgment call with no reason.

Name the passages to cut and the sections to add.

### 3. Did it increase scope?

Trace each addition to the original: Scope bullets, Write set files, Contract
surfaces, Acceptance criteria, and Deferrals rows, including work moved in from
another brief's Deferrals.

- **Concretization** — implied by the original Objective, now made specific.
- **Coverage** — a criterion added for a failure that would break an Objective
  guarantee.
- **Growth** — neither; the task now builds or promises more.

Report growth as counts ("+2 Scope bullets, +3 files, +1 contract surface"). Also
report anything removed: a narrowed Objective or a dropped criterion needs the
reason the refinement states. Growth is a boundary change for the phase author
and task author, as `refine-task` says; the review reports it and does not
approve it.

### 4. What value does it add, and what cost?

- **Value** — the Confirmed problems an implementer would otherwise hit, the
  concreteness gained (locations, order, a runnable demonstration).
- **Cost** — lines the implementer must now read (Question 2), predictions
  written as instructions without evidence (Question 1), growth (Question 3),
  and constraints that narrow implementation choices without a stated reason.

End with one sentence on whether the value covers the cost.

## Verdict

State one:

- **Accept** — merge the refinement as it stands.
- **Accept with trims** — list the exact cuts and additions; the refiner or a
  maintainer applies them.
- **Return to refiner** — load-bearing changes are Not shown or Preference; say
  which.
- **Escalate** — growth or an ownership conflict needs the phase author and task
  author before the refinement can merge.
- **Use the original** — the cost exceeds the value; implement from the original
  brief.

## Report

Post one comment on the refinement PR: the SHAs reviewed, the four answers with
their evidence, and the verdict. Make no brief, tracker, or implementation
edits, and create no tasks or deferral rows.

## Done

The PR carries one review that answers all four questions with evidence and ends
in one verdict.
