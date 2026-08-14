# P2-023: Implement artifact storage and changes renderer

**Phase:** 2 — Structured experience and second harness
**Shape:** capability
**Dependencies:** P2-021 (semantic: file-change and artifact events must exist before they can be stored or projected), P1-050 (contract: adds artifact paths to the OpenAPI document and an ordered migration)

## Objective

A run's file changes and produced artifacts are durable, downloadable by an
authorized caller, and shown as a reviewable diff.

## Scope

- The `Artifact` aggregate: id, run, media type, digest, size, sensitivity, and
  a storage reference. Metadata in the database, bytes in blob storage behind a
  port with a local-filesystem adapter.
- Content addressing by digest, so a repeated artifact stores once.
- API: list a run's artifacts, read metadata, download content with the
  sensitivity class enforced on every read.
- Changes renderer: file-change events grouped by path, artifact-backed diffs,
  bounded rendering for very large diffs with an explicit truncation marker.
- Size limits and a rejection path for an artifact exceeding them, recorded as a
  diagnostic rather than dropped silently.

## Non-goals

- Retention and deletion sweeps — P2-027.
- Transferring artifacts from a remote runner — P5-045.
- Mutating the user's checkout. Changes come back as diffs and artifacts.

## Runtime reachability

`GET /api/v1/runs/{id}/artifacts` and the changes view on the run page;
`winch run artifacts`.

## Owned surfaces

`internal/application/artifact.go`, `internal/adapters/blob/`,
`api/openapi/paths/artifacts.yaml`,
`internal/adapters/postgres/migrations/008_*.sql`,
`web/src/renderers/changes/`.

## Demonstration

    $ ID=$(winch run create --harness fake --scenario writes-files --json | jq -r .id)
    $ winch run start $ID && winch run artifacts $ID
    → expect: one artifact per produced file, with digest, size, and sensitivity

    $ winch run artifacts $ID --download <artifact-id> | sha256sum
    → expect: matches the recorded digest

    # in the browser, the changes view for that run:
    → expect: a per-file diff; a deliberately huge file shows a truncation
      marker rather than freezing the tab

    $ winch run artifacts $ID --download <artifact-id> --token <other-user-token>
    → expect: the same problem response as a missing artifact, revealing nothing

## Verification

- Standing scenario suite passes with an artifact-producing scenario added.
- Storage tests: digest collision reuse, partial write failure, restart during
  upload.
- Diff renderer tests including binary, empty, renamed, and enormous files.
- Authorization test asserting absent and forbidden are indistinguishable.

## Acceptance criteria

- [ ] Artifact bytes are never inlined into an event payload.
- [ ] Every artifact read enforces its sensitivity class.
- [ ] A repeated artifact stores its bytes once.
- [ ] Oversized artifacts are refused with a stable code and a diagnostic event.
- [ ] The changes view renders bounded output for an unbounded diff.

## Deferrals

| Deferred | Owning task |
|---|---|
| Retention sweeps, export, and deletion of artifacts | P2-027 |
| Disposable per-run worktrees producing these changes | P3-028 |
| Content handoff from a remote runner | P5-045 |

## Traces to

`docs/architecture.md` §6; `docs/contracts.md` §6;
`docs/security.md` §5, T14; `docs/roadmap.md` Phase 2
