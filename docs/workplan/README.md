# Implementation workplan

**No plan is in flight.** The previous plan was closed and stripped, and the
next one has not been derived. `tasks.json` is an empty tracker, so the
availability query answers truthfully with nothing available:

    $ ./scripts/list-available-tasks.sh
    → []

Two documents hold what a plan would otherwise be read for.
[`docs/state.md`](../state.md) says what exists at HEAD, what went wrong getting
here, and what has no working code behind it.
[`docs/roadmap.md`](../roadmap.md) says the order the gaps get closed in, stage
by stage, and carries the deferred decisions a brief may name as an owner.

## Deriving the next plan

`skills/workplan/SKILL.md` in Create mode, against `docs/state.md` and the
stage of `docs/roadmap.md` being opened.
[`skills/shared/workplan-model.md`](../../skills/shared/workplan-model.md) is
the definition it works from — layout, the seven invariants, the four task
shapes, dependency edge reasons, write sets, contract surfaces, brief anatomy,
and the frozen tracker schema.

Two things in `docs/state.md` bind that derivation. The quality gates name paths
that no longer exist, which is an I1 and I2 failure, so the repair belongs in the
first phase rather than in a brief that assumes a system which builds. And the
roadmap has seven stages, so a plan covering all of them needs the phase numbers
the tracker schema now permits.

## Once a plan exists

- **`tasks.json` is the graph.** This file restates it as per-phase tables;
  drift between them is a defect. A task is available when its status is
  `pending` and every ID in `depends_on` is `completed`.
- **One task, one pull request.** Include `Task: <ID>` in the body. Automation
  stamps `completed` when the pull request is approved — do not edit status
  fields by hand.
- **A deferral names an owner that already exists** — a task ID in
  `tasks.json`, or a `docs/roadmap.md` deferred decision with a trigger. A phase
  name that resolves to no file is not an owner.
- **Write collisions are not edges.** Two available tasks touching one file
  rebase, and the pairs are listed per phase so the cost is visible. Two
  available tasks sharing a *contract surface* is a defect, not a cost.
- **Plan failures get post-mortems.** When a defect's root cause is a seam
  between briefs rather than an implementation, it goes in `post-mortems/`.
  Ordinary bugs do not.
