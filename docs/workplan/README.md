# Implementation workplan

Phase 0 repairs the gaps [`docs/state.md`](../state.md) records under *What went
wrong* — invariants the previous plan never established, and quality gates that
broke when it closed. Later phases implement the capabilities listed under *What
is not implemented*.

Pick up work with `scripts/list-available-tasks.sh`. Each task is one pull
request; include `Task: <ID>` in the body. Automation stamps `completed` in
`tasks.json` when the pull request is approved — do not edit status fields by
hand.

## How the plan is put together

[`skills/shared/workplan-model.md`](../../skills/shared/workplan-model.md) is
the definition — layout, the seven invariants, the four task shapes, dependency
edge reasons, write sets, contract surfaces, brief anatomy, and the frozen
tracker schema. What follows is only what a reader needs to use this plan.

- **`tasks.json` is the graph.** Phase briefs restate it as tables; drift
  between them is a defect. A task is available when its status is `pending` and
  every ID in `depends_on` is `completed`.
- **A deferral names an owner that already exists** — a task ID in
  `tasks.json`, a `docs/roadmap.md` deferred decision with a trigger, or a phase
  whose brief lists the deferral in its *Deferrals in* register. A phase name
  that resolves to no file is not an owner.
- **Write collisions are not edges.** Two available tasks touching one file
  rebase; the pairs are listed in the phase brief so the cost is visible. Two
  available tasks sharing a *contract surface* is a defect, not a cost.
- **Plan failures get post-mortems.** When a defect's root cause is a seam
  between briefs rather than an implementation, it goes in
  [`post-mortems/`](post-mortems/). Ordinary bugs do not.
