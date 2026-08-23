# P0-005: Sanitize post-close plan references

**Phase:** 0 — Foundation repair
**Shape:** hardening
**Dependencies:** None

## Objective

Every contributor-facing document describes the current plan or the empty state
truthfully; no task ID or brief path from the closed plan survives outside git
history.

## Scope

- Update `AGENTS.md` *Start here* and *Current state* to route contributors to
  `docs/workplan/README.md` and name what Phase 0 repairs versus what remains in
  `docs/state.md`.
- Replace stale `docs/workplan/` citations in:
  - `internal/adapters/postgres/migrate.go` (migration-slot comment)
  - `scripts/task-prompt.md`
  - `runner/README.md`
  - `docs/decisions/0004-task-status-authority.md`
- Confirm `scripts/list-available-tasks.sh` and
  `scripts/stamp_task_completion.py --check-only` succeed against the new
  tracker (empty available list is fine; missing file is not).

## Non-goals

- Changing dispatch automation behavior beyond what truthful references require.
- Writing Phase 1 briefs.
- Editing `docs/state.md` (it remains the close report until the next close).

## Runtime reachability

- **Composition root:** none — documentation and comments only.
- **Profile:** contributor onboarding and CI task-status gate.
- **Command:** `scripts/list-available-tasks.sh`,
  `scripts/stamp_task_completion.py --check-only`.

## Write set

- `AGENTS.md`
- `internal/adapters/postgres/migrate.go`
- `scripts/task-prompt.md`
- `runner/README.md`
- `docs/decisions/0004-task-status-authority.md`
- `README.md` (only if it still cites the removed plan index)

## Contract surfaces

None.

## Demonstration

    $ scripts/list-available-tasks.sh
    → expect: JSON array of available Phase 0 tasks (five entries at plan open)

    $ python3 scripts/stamp_task_completion.py --check-only
    → expect: exit 0 on a branch whose tracker matches the default branch shape

    $ rg 'docs/workplan/phase-[0-9]/P[0-9]+-' --glob '!docs/workplan/**' .
    → expect: no matches outside `docs/workplan/`

    $ rg 'P[0-5]-[0-9]{3}' --glob '!docs/workplan/**' --glob '!.git/**' .
    → expect: only machinery/tests that intentionally mention example IDs, not
      stale ownership deferrals

## Verification

- `make check` passes (no code behavior change expected).
- Manual read of updated docs: each sentence is true at HEAD.

## Acceptance criteria

- [ ] `AGENTS.md` points to the active plan and accurately summarizes Phase 0
  scope.
- [ ] No source comment or contributor doc cites a removed brief path.
- [ ] Migration-slot comment in `migrate.go` references the current plan or
  `docs/state.md`, not the closed plan index.
- [ ] Task-status gate and availability query succeed against `tasks.json`.
- [ ] Close-procedure sanitization rule from `skills/workplan/SKILL.md` is
  satisfied for references introduced by the close at `ccad757`.

## Deferrals

None.

## Traces to

- `docs/state.md` §*Automation left without its tracker* and §*Documentation that
  outran the code*
- `skills/workplan/SKILL.md` §*Sanitizing the references*
