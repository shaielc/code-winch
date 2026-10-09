---
name: reflect
description: Optionally improve task-development instructions, skills, tools, environment, and verification using evidence from completed work.
---

# Reflect

Use this optional side operation when a task reveals a repeatable source of
error, wasted effort, or ambiguity in the development system itself. Reflect
is not part of the mandatory task delivery loop (refine → implement → audit →
merge) and must not become a routine ceremony.

## Trigger

Ask: **What did this task reveal about our instructions, skills, tools,
environment, or workflow that could make similar future tasks more reliable
or cheaper?**

Reflect only when the evidence suggests a recurring problem or a broadly
reusable clarification. One-off product changes belong in tasks or
propagation, not in this skill.

## Procedure

1. **Reconstruct the actual sequence.** Read the task brief, implementation,
   PR discussion and review rounds, and related issues created from audit
   findings. Before saying a finding was dropped, check whether an issue
   or downstream task already owns it.
2. **Collect evidence.** Record the observed failure, ambiguity, or unnecessary
   cost and cite a command with output, a `file:line`, an issue, or review.
   Distinguish what was observed from a proposed explanation.
3. **Choose the narrowest durable fix.** Prefer a sentence or example in an
   existing instruction, skill, or verification document to new mandatory
   phases, gates, or templates. Do not change product contracts under the
   guise of workflow improvement.
4. **Preserve gate accuracy.** Specify what a command actually checks and what
   it does not. In particular, do not describe Go checks as covering web UI,
   or a Docker path as covering host-only lint, unless verified.
5. **Show the difference.** Record the evidence, exact amendment, how it
   prevents recurrence, and any tradeoff. If no small reusable change is
   justified, report that and leave the process untouched.

## Examples from interactive UI work

- Audits use Blocking, Defect, Deferrable and Harden as *descriptions* with
  explicit human dispositions, not as automatic product decisions.
- When reconstructing whether audit findings were addressed, inspect issues
  opened from those findings as well as the source PR.
- For behavior with multiple input methods or submission paths, consider a
  behavior/action-path matrix (e.g. keyboard submit, button submit, invalid
  submit, focus behavior). Use it when it reveals path-specific gaps; do not
  require it on every task.
- Verification docs should distinguish the Go host gate, Docker integration
  path, and web-specific checks by their actual coverage. A green Go gate
  alone cannot establish that web interaction behavior was verified.

## Boundary

Reflection proposes or makes small *development-system* amendments separately
from the task audit's verdict and from downstream product propagation. It
does not revise the audited brief to justify code, decide defect dispositions,
or add merge requirements without an explicit process decision.
