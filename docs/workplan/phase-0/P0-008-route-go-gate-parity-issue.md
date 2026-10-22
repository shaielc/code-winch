# P0-008: Route issue #149: Go gate option parity

**Phase:** 0 — Plan the implementation phases
**Shape:** planning
**Dependencies:** None

## Objective

Decide, as phase author, which phase owns making the two Go gate options run
the same checks (issue [#149](https://github.com/shaielc/code-winch/issues/149)),
and record that disposition so a task author can derive its implementation
brief.

## Scope

Issue #149 reports that the Go gate differs by option. `make test-cycle` runs
`make format-check vet test build e2e` in the toolchain and has no lint.
`make check` runs `format-check vet lint test build` and has no e2e; `make e2e`
needs a started environment (`make test-env`) because `test/e2e/sandbox.go`
fails when `WINCH_E2E_URL` is unset. `docs/testing.md` currently describes the
difference instead of treating the options as equivalent.

The issue's expected result is that either option runs the same checks, by
adding lint to `make test-cycle`, by letting `make check` (or `make e2e`) start
the test environment it needs, or by making the e2e runnable without one.
Choosing among these is an implementation decision for the derived brief; this
task decides ownership.

Use `skills/revise-phase/SKILL.md` (or `skills/create-phase/SKILL.md` if the
decision is a new phase). Inspect the `Makefile` targets named above,
`test/e2e/sandbox.go`, `.github/workflows/go.yml`, `docs/testing.md`,
`deployments/README.md`, the Phase 1 README, and the roadmap. The issue is
repository tooling: it advances no Stage 0 increment and no product capability.
Decide one of:

- add an increment, or a tooling concern, to an existing phase's design, with
  its observable outcome, inclusions, exclusions and prerequisites;
- open a new phase for repository gates, justified against the roadmap; or
- decline the issue with a recorded reason.

State whether the derived work is one brief. If it is, name it as an
independent, non-blocking item that does not reorder Phase 1 increments, and
add the planning-shape task that derives it.

## Required constraints

- Phase boundaries are decided here, not by the task author who derives the
  brief. Do not silently combine this with an existing Phase 1 increment.
- Preserve completed IDs and every existing status. Allocate an implementation
  ID only with a complete brief; reserve none.
- Keep `docs/testing.md` accurate: it names #149 as the current difference
  between the options, and the implementation brief owns updating it.

## Non-goals

This task writes planning artifacts only. It changes no `Makefile`, workflow,
test, or `docs/testing.md` content, and derives no implementation brief.

## Planning outputs and handoff

The decided disposition recorded in the owning phase README (or the new phase
README), the matching Phase 0 coverage row, the workplan index, and the
`tasks.json` entry for the derivation task. Handoff: complete
`skills/phase-increment-handoff/SKILL.md` for the decided phase, then a
different author takes the derivation task with `skills/derive-task/SKILL.md`.

## Write set

- `docs/workplan/phase-<N>/README.md` (the owning phase, or a new phase)
- `docs/workplan/phase-<N>/<ID>-plan-<slug>.md` (the derivation task)
- `docs/workplan/tasks.json` (new entries only; existing statuses unchanged)
- `docs/workplan/README.md`
- `docs/workplan/phase-0/README.md`

## Contract surfaces

None. The derived brief will own the Go gate's composition in the `Makefile`
and the corresponding description in `docs/testing.md`; no product contract
changes.

## Demonstration

Review the phase record and confirm that issue #149 has a stated disposition,
an owning phase with explicit boundaries, and either one derivation task with a
brief and tracker entry or a recorded reason for declining. Confirm Phase 0
coverage, the index and the tracker agree.

## Verification

Run completion-script and control-panel tests, tracker JSON validation,
`scripts/list-available-tasks.sh`, an `rg` search for `#149` and the allocated
IDs, and `git diff --check`. Run the Go gate and say which option ran.

## Acceptance criteria

- [ ] The owning phase, or a recorded decline, is stated with its reasoning.
- [ ] A chosen phase README states the outcome, inclusions, exclusions and
      prerequisites of the work, and does not reorder existing Phase 1
      increments.
- [ ] Exactly one derivation task exists for the work, with a brief, a tracker
      entry and index and coverage rows; no implementation ID is reserved.
- [ ] Tracker, index and Phase 0 coverage agree and preserve all existing
      statuses.

## Deferrals

None. The implementation of gate parity is owned by the brief this task
assigns for derivation, not unfinished work here.

## Traces to

Issue #149; review threads on #148; `docs/testing.md`; `Makefile`;
`skills/shared/workplan-model.md` (Phase 0 meta tasks).
