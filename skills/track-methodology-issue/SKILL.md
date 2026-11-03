---
name: track-methodology-issue
description: Find or create tracking issues for evidence-backed development-methodology observations, and record recurrences, amendments, and whether they worked. Use during reflect-task or when updating an existing methodology tracking issue.
---

# Track methodology issue

Track **how the work is done**, not product defects or audit finding
dispositions. Read the source task and PR discussions, the applicable skills,
and the issue records before writing. An observation may concern planning,
refinement, review, implementation, audit, or a handoff across operations.

## Find the fitting issue

1. Identify the operation or handoff, the observed failure, and the suspected
   practice behind it. Keep evidence separate from the suspected cause.
2. Search **open and closed** GitHub issues for the operation or skill name,
   the failure's terms, related task and PR numbers, and known reflection or
   amendment references. Read candidate issue bodies **and comments**; follow
   links to earlier occurrences and amendments before judging the fit.
3. Reuse an issue that tracks the same failure mode or practice across tasks,
   even when the current task and symptoms differ. If observations involve
   distinct operations or causes, choose a fitting issue for each; cross-link
   related issues without copying the same record into all of them.
4. Open a new tracking issue only if none fits, or combining distinct problems
   would obscure how either can be assessed. Do not create one issue per task or
   another issue just because the earlier one is closed. Use a descriptive title
   such as `Track: <operation> — <failure mode>`. Explain its scope, initial
   evidence, known related issues, and what later work should check.

## Record observations

For every observation, add a comment to its fitting issue (or include the
initial observation in a new issue). Record:

- The task, source PR, review or decision, and direct evidence links; what
  happened, the expected practice, and the observed effect or human cost.
- The plausible methodology cause, uncertainty or competing explanation, and
  the instruction or handoff involved. Do not present inference as fact.
- Whether this is new, a recurrence, or already addressed by an amendment;
  distinguish a missing instruction from a rule that existed but was not used.
- Its disposition (**open**, **watch**, **fixed by amendment**, or another
  explicitly explained state) and a concrete check on the next comparable task.

Keep observations independently identifiable within the issue so a later
occurrence can refer to the same one. Do not silently overwrite earlier
evidence or mix an unrelated concern into a convenient issue.

## Record amendments and outcomes

After an amendment PR exists, record its link on each affected tracking issue,
the exact practice it changes, which observations it addresses, and the expected
effect and assessment check. Mark those observations **fixed by amendment**,
stating whether the PR is open or merged. This records the attempted remedy,
**not verification** that it works; leave unaddressed observations open or on
watch as appropriate.

On the next comparable task, compare what happened with that check. Record the
result and its evidence in the same issue. If the failure recurs, say the
amendment did not work and record the recurrence rather than opening a
duplicate. Do not close tracking issues or declare the methodology effective
solely because an amendment was submitted or merged.

## Done

Every observation has a fitting issue and evidence, each amendment is linked to
the observations it addresses, and each untested amendment names the later
comparison that will assess it. Report issue and amendment links to the
reflection author.
