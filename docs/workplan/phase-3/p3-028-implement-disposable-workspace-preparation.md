# P3-028: Implement disposable workspace preparation

**Phase:** 3 — Docker isolation
**Shape:** hardening
**Dependencies:** P2-055 (semantic: preparation copies a registered workspace beneath its approved root), P2-023 (semantic: changes come back as artifacts rather than mutating the user's checkout)

## Objective

A run works in a disposable copy of the workspace, and its changes return as
reviewable diffs instead of touching the user's checkout.

## Scope

- Per-run preparation producing a disposable worktree or copy, named
  deterministically so an orphan is recognizable after a crash.
- Path resolution beneath the workspace's approved root, with traversal,
  symlink, submodule, and hardlink escapes refused before anything is mounted.
- Safe archive extraction where a workspace source is an archive: bounded size,
  bounded entry count, no absolute or parent paths, no device or link entries.
- Change capture: a diff against the preparation baseline, emitted as
  file-change events and an artifact.
- Cleanup on terminal state and on daemon restart, with a sweep that removes
  orphaned preparations and reports what it removed by ID.

## Non-goals

- Container creation, mounts, and limits — P3-029 consumes what this prepares.
- Pushing changes to a git remote or signing commits; both are separate explicit
  capabilities per `docs/security.md` §7.

## Runtime reachability

Every run on the compose stack once the workspace policy selects a disposable
preparation; visible as the preparation path in the run's resolved
configuration.

## Owned surfaces

`internal/application/workspace_prepare.go`, `internal/adapters/worktree/`,
`test/integration/worktree/`.

## Demonstration

    $ ID=$(winch run create --workspace <id> --harness fake --scenario writes-files --json | jq -r .id)
    $ winch run start $ID && git -C /srv/code/project status --porcelain
    → expect: empty; the user's checkout is untouched

    $ winch run artifacts $ID
    → expect: a diff artifact containing the run's changes

    $ winch run stop $ID && ls "$WINCH_PREPARE_ROOT"
    → expect: no leftover directory for that run

    $ docker compose kill winchd   # mid-run, then restart
    → expect: the startup sweep removes the orphaned preparation and logs its
      run ID and nothing else

    $ winch run create --workspace <id-with-symlink-escape> --harness fake
    → expect: refused before preparation, naming the escape class

## Verification

- Standing scenario suite passes with disposable preparation enabled.
- Traversal and extraction test table: relative, absolute, symlink, hardlink,
  submodule, and archive bomb cases.
- Orphan sweep test after a killed daemon.
- Large-repository preparation timing test with a bounded budget.

## Acceptance criteria

- [ ] No run path can read or write outside its prepared copy and approved root.
- [ ] The user's primary checkout is never mutated.
- [ ] Every preparation is removed on terminal state or by the restart sweep.
- [ ] Refusals name a stable escape class and no filesystem path.

## Deferrals

| Deferred | Owning task |
|---|---|
| Mounting the preparation into a container | P3-029 |
| Read-only preparation for the review profile | P3-030 |

## Traces to

`docs/security.md` §6 (workspace paths), T07, T08, LB03;
`docs/architecture.md` §4; `docs/roadmap.md` Phase 3
