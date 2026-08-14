# P2-055: Register workspaces and approved roots

**Phase:** 2 — Structured experience and second harness
**Shape:** seam
**Dependencies:** P1-050 (contract: run creation currently takes a free-form path and must take a workspace reference instead)

## Objective

A run names a registered workspace instead of an arbitrary path, and the
workspace is where source, ownership, and policy actually live.

## Scope

- The `Workspace` aggregate from `docs/architecture.md` §6: id, owner, source,
  and policy. Persisted, validated, and listed — the root policy boundary that
  P2-024, P2-026, P2-027, P3-028, and P3-032 all assume.
- Registration validates the source path: canonicalized beneath an
  administrator-approved root, with traversal and symlink escapes refused at
  registration time rather than at launch.
- Workspace policy fields that later tasks read: permitted sandbox profiles,
  permitted harness profiles, and whether `local-trusted` is allowed at all.
- API and CLI: register, list, read, update policy, delete. Deleting a workspace
  with active runs is refused with a stable code.
- Run creation takes a workspace ID. The single implicit workspace P1-050
  deferred is replaced by a real record; a migration registers the existing
  configured root as one workspace so in-flight deployments keep working.

## Non-goals

- Authorization of workspace access by actor — P2-026 builds the authorization
  and audit trail on top of this aggregate.
- Cloning, checkout, or disposable worktrees — P3-028.
- Multi-user ownership semantics beyond a single owner field.

## Runtime reachability

`POST /api/v1/workspaces` and `winch workspace add|ls|rm`; every `run create`
call after this task names one.

## Owned surfaces

`internal/application/workspace.go`,
`api/openapi/paths/workspaces.yaml`, `api/openapi/components/run.yaml` (`CreateRunRequest` only — shared with P2-059),
`internal/adapters/postgres/migrations/010_*.sql`, `cmd/winch/workspace.go`,
`web/src/features/workspaces/`.

## Demonstration

    $ winch workspace add --source /srv/code/project --name project
    → expect: an ID, the canonicalized source, and the policy defaults

    $ winch workspace add --source /srv/code/project/../../etc
    → expect: refused with a stable code naming the approved-root violation

    $ ln -s /etc /srv/code/project/escape && winch workspace add --source /srv/code/project/escape
    → expect: refused; the symlink is resolved before the check

    $ winch run create --workspace <id> --harness fake --json
    → expect: a run bound to the workspace; a run naming an unknown workspace is
      refused before any process starts

    $ winch workspace rm <id>   # while a run is active
    → expect: refused, naming the active run count and not its contents

## Verification

- Standing scenario suite passes with runs created against a registered
  workspace.
- Traversal and symlink test table, including relative, absolute, and nested
  escapes.
- Migration test proving an existing deployment's implicit root becomes one
  workspace without losing run history.
- Policy round-trip tests for each policy field.

## Acceptance criteria

- [ ] No API path accepts a free-form filesystem path for a run.
- [ ] Registration refuses any source resolving outside an approved root.
- [ ] Workspace policy is readable by the approval, authorization, and sandbox
      paths that depend on it.
- [ ] The implicit-workspace assumption recorded in P1-050 is removed from that
      brief's Deferrals when this lands.

## Deferrals

| Deferred | Owning task |
|---|---|
| Authorizing workspace access per actor and auditing it | P2-026 |
| Deleting workspace-scoped content on request | P2-027 |
| Disposable per-run copies of a workspace | P3-028 |

## Traces to

`docs/architecture.md` §6; `docs/security.md` §2 (sandbox to repository), T08;
`docs/code-structure.md` §6; `docs/roadmap.md` Phase 2
