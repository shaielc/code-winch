# P2-027: Implement retention, export, and deletion

**Phase:** 2 — Structured experience and second harness
**Shape:** capability
**Dependencies:** P2-023 (semantic: artifacts and blobs are the content being retained and deleted), P2-026 (semantic: export and deletion require an authorization and audit boundary)

## Objective

Content ages out on its sensitivity class's schedule, an authorized user can
export a scope with a manifest, and a deletion request provably removes every
copy the system controls.

## Scope

- A retention worker applying the per-class durations in `docs/security.md` §5
  to events, artifacts, blobs, outbox payloads, projections, exports, and audit
  records, with the security-audit exception at 365 days.
- Export: fresh authorization for the workspace and every included resource, a
  sensitivity summary, explicit confirmation for `confidential` data, and a
  manifest of IDs, schema versions, sizes, and digests. Bundles inherit the
  highest included class, expire after 24 hours, and are deleted after a
  successful download.
- Deletion: idempotent, deny-by-default on ambiguous ownership, driven by a
  retryable ledger across every store, leaving only a content-free tombstone
  recording resource ID, request and completion time, policy version, and
  partial failures.
- Partial deletion is retried and visible to an operator; success is not claimed
  until every mutable store acknowledges and immutable-copy expiry is recorded.
- Credential deletion means provider revocation plus reference removal.

## Non-goals

- Backup and restore procedures — P5-046.
- A renderer output cache. None exists; caching is a deferred decision in
  `docs/roadmap.md`, so there is no cache to invalidate here.
- Configurable retention longer than a class maximum.

## Runtime reachability

`winch export create|download`, `winch delete run|workspace`, and the retention
worker running inside `winchd`.

## Owned surfaces

`internal/application/retention.go`, `internal/application/export.go`,
`api/openapi/paths/exports.yaml`,
`internal/adapters/postgres/migrations/012_*.sql`, `cmd/winch/export.go`.

## Demonstration

    $ winch export create --run $ID --include user-content
    → expect: a manifest with digests and a sensitivity summary; confidential
      data requires an explicit second confirmation

    $ winch export download <export-id> && winch export download <export-id>
    → expect: the first succeeds, the second reports the bundle is gone

    $ winch delete run $ID && winch run get $ID
    → expect: a tombstone with IDs and times, and no content anywhere

    $ psql "$WINCH_DATABASE_URL" -c "select count(*) from run_events where run_id='$ID'"
    $ ls "$WINCH_BLOB_ROOT" | grep <artifact-digest>
    → expect: 0 and no match

    # with a blob store injected to fail:
    $ winch delete run $ID2
    → expect: the request is recorded as partially failed, retried, and visible
      to an operator; success is never claimed

## Verification

- Standing scenario suite passes with the retention worker running.
- Fake-clock retention tests for every sensitivity class and the audit exception.
- Crash-replay deletion test: a kill between stores resumes and completes.
- Export authorization tests, including a resource the actor may not read.
- Canary test: an exported bundle contains no `secret`-class value.

## Acceptance criteria

- [ ] Every sensitivity class has a tested retention duration and deletion
      deadline.
- [ ] Deletion is idempotent and survives a crash mid-sweep.
- [ ] An export bundle's manifest matches its contents by digest.
- [ ] A partial deletion failure is retried and never reported as success.
- [ ] Tombstones contain no content.

## Deferrals

| Deferred | Owning task |
|---|---|
| Backup expiry verification and restore drills | P5-046 |
| Deleting content held on a remote runner | P5-045 |

## Traces to

`docs/security.md` §5, T14, LB07; `docs/contracts.md` §2 (sensitivity);
`docs/roadmap.md` Phase 2
