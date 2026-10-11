---
name: audit-phase
description: Audit a phase and its derived tasks against the architecture, roadmap, and repository. Use to judge increment order, boundaries, coverage, dependencies, concurrency, or phase completion.
---

# Audit phase

Act as an independent phase auditor. Read applicable `AGENTS.md` files, design
sources, phase, index, tracker, briefs, and implementation. Use
[the model](../shared/workplan-model.md) and
[the phase format](../shared/phase-design.md).

Judge the architecture → phase → task chain against the requested commit.
Report findings; change none of the artifacts being judged.

## Check

- Starting point: completed behavior exists; assumptions distinguish source
  evidence from runtime observations.
- Design: numbered increments add observable properties in a valid order;
  boundaries are decided and excluded concerns have destinations.
- Task fit: each task belongs to an increment, fits a shape, and keeps its
  minimum operator path. Independently observable work is not bundled.
- Coverage: responsibilities are accounted for; underived increments and legacy
  briefs awaiting reconciliation are labelled explicitly.
- Fakes: their current behavior or interface is not a product constraint; they
  can change to exercise the selected increments.
- Graph: references resolve, dependencies have legitimate reasons, no cycles
  exist, and contract collisions have an owner and edge. Report concurrent
  write collisions separately.
- Records: briefs, tracker, mappings, index, and collision report agree.
- Completion: criteria, reachability, demonstrations, and invariants hold at
  the audited commit, rather than merely having completed statuses.

Give each finding a command and observed output or `file:line`, its effect,
and the artifact or responsibility needing correction. Distinguish incomplete
task derivation from an undecided phase boundary. Derive width and critical
path from actual task coverage, never unnamed future tasks.

## Report

State which decisions and completion claims hold, each discrepancy, and what
remains unverified. Route accepted partition changes to `revise-phase`, added
coverage to `extend-phase`, and brief derivation to a separate task author.
Record a plan defect in `docs/workplan/post-mortems/` through a separate change
when the cause is a seam between briefs or an unowned invariant.

An audit neither changes statuses nor closes the plan. An explicitly requested
close uses `skills/close-workplan/SKILL.md`.
