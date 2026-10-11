# Phase artifact format

Use this format for `docs/workplan/phase-N/README.md`. The phase specifies the
order and boundaries of observable increments before task briefs are derived.
An increment usually becomes one task and can split into smaller tasks.

## Outcome and starting point

State what a person can do at phase exit and cite its design/roadmap sources.
Record reusable capabilities and missing behavior. Support the baseline with
source references or observed commands, and state runtime-verification limits.

## Concern boundaries

Give each concern one row. Name the increment that introduces it, or the roadmap
stage or triggered decision owning it outside the phase. Keep different concerns
separate even when an increment needs both as required wiring.

## Ordered increments

Number increments in delivery order. Put the **number only** in `Order` and a **descriptive title** in `Increment`:

| Order | Increment | Observable change | Includes | Excludes | Prerequisites |
|---|---|---|---|---|---|
| 1 | Live harness output | Before/after behavior | Responsibility and minimum operator path | Neighboring concerns owned elsewhere | Existing behavior or earlier increments |

`Order` is the increment's sequence number; `Increment` names the work in plain language. Do not embed the order number in the increment title. Keep numbers separate from titles so the order can change without renaming the increment or breaking coverage references. Order guides derivation; task edges still need the model's
reasons. State independent branches explicitly rather than treating neighboring
rows as dependencies.

Select boundaries in the rows. Avoid candidates and instructions to decide
later. Give a concise rationale only for an otherwise ambiguous grouping.
Keep procedure, commands, fixtures, protocol implementations, and file forecasts
out of the phase artifact.

## Task coverage

Map increments to task IDs derived so far or mark **not yet derived**. Identify
legacy briefs needing rederivation separately from completed reusable work.
Coverage records derivation, not implementation. Identify the next increment.
Keep statuses and executable dependencies in `tasks.json`; allocate no
speculative IDs in the phase.

## Phase exit

Describe observations proving the outcome and included guarantees. Leave
commands and setup to task briefs. Add no tests for excluded concerns.
