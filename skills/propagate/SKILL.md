---
name: propagate
description: Optionally carry a concrete discovery from one task into an already-planned downstream task without redesigning the plan.
---

# Propagate

Use this optional side operation when completing or auditing a task reveals a
fact or constraint that materially changes or clarifies work already planned
downstream. It is not a routine step or a new delivery gate. The core loop
remains refine → implement → audit → merge.

## Trigger

Ask: **Did this task reveal something that should materially change or
clarify an already-planned downstream task?**

If no downstream task exists, or the discovery does not affect its documented
objective, scope, or acceptance criteria, do not propagate. Record an audit
finding or use ordinary planning instead; do not invent speculative work.

## Procedure

1. **Identify the destination.** Read the relevant downstream brief and
   `docs/workplan/tasks.json`; confirm the task ID exists and is the owner.
2. **State the discovery.** Cite the observed behavior or source (`file:line`,
   command and output, review finding), and explain why the downstream brief
   needs clarification.
3. **Make the smallest change.** Update the relevant Scope, assumption,
   acceptance criterion, or demonstration in that brief. Preserve its non-goals,
   contract ownership, and objective; do not opportunistically redesign.
4. **Check consistency.** Keep the task's write set, tests, and other brief
   sections consistent where affected. Do not change task status fields.
5. **Report the handoff.** Name the source task, destination task, evidence,
   changed brief section, and what remains for the downstream implementation.

An audit reports findings and does not silently implement its own dispositions.
Get the human disposition/owner for a Deferrable finding before editing the
downstream brief; make propagation as a separate change.

## Example

P0-004's append-only transient echo uses positional React keys. P0-005
replaces that echo with persisted submissions: clarify P0-005's brief to
require stable runner-backed record identity, rather than carrying
`key={index}` into persisted records. Do not redesign P0-005's submission
protocol or implement it during propagation.

## Done when

The destination brief states the necessary fact or constraint, its owner is
explicit, and no unrelated surfaces or plans changed.
