---
name: reflect-task
description: Find and improve flaws in the development methodology using evidence from a task's planning, execution, handoffs, reviews, and decisions. Use to explain recurring process failures or wasted effort, including nontechnical causes.
---

# Reflect task

Examine **how the work was planned, communicated, decided, and reviewed**.
Find methodology problems that made the task harder, less reliable, or more
expensive. Product defects are evidence to inspect, not the output of reflection.

Read applicable `AGENTS.md`, the task and phase artifacts, PRs and discussions,
audit findings, decisions, and follow-up tasks or issues. Establish what happened
before attributing it to the process. Keep reflection separate from the audit
verdict and from finding propagation.

## Look for methodology failures

- Goals or acceptance expectations were interpreted differently by participants.
- Planning combined concerns, omitted ownership, or assumed prerequisite work.
- A handoff lost a decision, changed its meaning, or reached the wrong owner.
- Review or approval rules caused repeated rounds without resolving disagreement.
- Information arrived too late, feedback had no destination, or an accepted
  tradeoff was forgotten.
- Instructions or incentives encouraged unsupported claims, redundant work, or
  a ritual that did not improve the outcome.

Consider planning, collaboration, prioritization, communication, and decision
making before narrowing to tools, code, or tests. A technical symptom may have
a nontechnical cause. A single task can expose a reusable methodology defect;
repeat occurrence strengthens the evidence but is not a prerequisite.

## Procedure

1. Reconstruct the sequence: expectations, responsibilities, handoffs, decisions,
   changes, review findings, and their dispositions. Link the actual records;
   check existing follow-ups before claiming something was dropped.
2. Describe an observed cost or failure and the development practice that
   plausibly caused it. Separate evidence from inference and consider competing
   explanations. A bad outcome alone does not prove a bad methodology.
3. Identify the rule, role, artifact, or feedback path that should change. Explain
   how that change would have altered the observed sequence.
4. Propose or make the smallest reusable amendment in the relevant instructions
   or workflow. Include a way to assess whether it improves the next comparable
   task and any added cost. Avoid adding a mandatory ceremony without a reason.
5. Report the observation, process cause, exact change, expected effect, and
   remaining uncertainty. If no methodology problem is supported, say so and
   leave the process unchanged.

## Example

A contributor and reviewer repeatedly disagree because the expected audience
was never stated in the brief. Reflection traces the disagreement to the
planning handoff and proposes recording the audience before work begins. It
checks whether the next comparable task avoids that repeated clarification.
The output is an improvement to the development method, not a redesign of the
product being reviewed.

## Boundary

Do not revise an audited brief to justify its implementation, choose human
finding dispositions, redefine product contracts, or close issues as a result
of reflection. Route product follow-ups through `propagate-task`. Reflection
may improve technical or nontechnical practices, but its conclusion must be
about the methodology that produced the work.
