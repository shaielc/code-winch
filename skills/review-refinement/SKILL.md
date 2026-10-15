---
name: review-refinement
description: Review a refined task brief against its original by answering four questions — did the refinement find problems, is it too verbose or too sparse, did it increase scope, and what value and cost does it add. Use when a refinement pull request is open and before implementation starts.
---

# Review refinement

Compare the brief at the refinement PR's base (the original) with the brief at
its head (the refined). Read applicable `AGENTS.md` files and
[`refine-task`](../refine-task/SKILL.md). Answer the four questions below, each
with evidence: a command with its observed output, or a `file:line`.

1. **Did it find problems with the original brief?** List each problem the
   refinement fixed or reported, and check it against the code at HEAD.
2. **Did it make the brief too verbose or too sparse?** Compare section sizes,
   original against refined. Name the passages that are padding and the places
   an implementer would still have to guess.
3. **Did it increase scope?** List every addition to Scope, Write set, Contract
   surfaces, Acceptance criteria, and Deferrals that the original Objective does
   not already imply.
4. **What value does it add, and what cost?** Value: what an implementer gains
   from the refined brief. Cost: everything the refinement adds beyond that,
   whether or not questions 1–3 found it — reading, work to build and verify,
   constraints on how to implement, and what can go wrong or go stale.
