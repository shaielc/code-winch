# Implementation workplan

Phase 0 repairs the gaps [`docs/state.md`](../state.md) records under *What went
wrong* — invariants the previous plan never established, and quality gates that
broke when it closed. Later phases implement the capabilities listed under *What
is not implemented*.

Pick up work with `scripts/list-available-tasks.sh`. Each task is one pull
request; include `Task: <ID>` in the body. Automation stamps `completed` in
`tasks.json` when the pull request is approved — do not edit status fields by
hand.

## Phases

Every phase has a brief, including phases with no tasks yet. The phase brief
carries that phase's objective, scope, task table, dependency graph, width, and
its register of deferrals from other phases.

| Phase | Name | Status | Brief |
|---|---|---|---|
| 0 | Foundation repair | derived — 11 tasks | [`phase-0/README.md`](phase-0/README.md) |
| 1 | Browser-reachable single-user product | not derived | [`phase-1/README.md`](phase-1/README.md) |

Phases 2 to 5 have no briefs yet. Nothing defers to them, and `docs/roadmap.md`
is the design-set statement of what they cover.

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

## Inherited state

At plan creation (`workplan/2026_08_21`), HEAD matched the close report at
`ccad757`:

- The daemon starts, migrates, and serves; `make check` passes.
- Run routes return 500/404 because `unavailableBackend` is still bound
  (`cmd/winchd/main.go:159-183`).
- `winch dev run` drives a fake harness locally; the operator CLI is not built by
  `make build` or installed in the deployment image.
- The fake harness is not controllable and is exercised only by hand.
- CI runs `make check` only; storage integration tests and a complete fake run
  are not gated.
- No `test/e2e` directory or `make e2e` target exists.
