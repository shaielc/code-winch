# P2-026: Add workspace authorization and audit trail

**Phase:** 2 — Structured experience and second harness
**Shape:** hardening
**Dependencies:** P2-055 (compile: authorization is scoped to the workspace aggregate), P1-050 (compile: the HTTP binding whose every operation is being authorized)

## Objective

Holding a run, artifact, credential, or workspace ID stops being sufficient to
act on it, and every security-relevant action leaves a content-free audit
record.

## Scope

- Identity beyond the single deployment actor: users, workspace membership, and
  a role granting read or control.
- Authorization on every object operation — workspace, run, artifact, input,
  credential, approval — checked against the object, not the request path.
- Absent and forbidden produce the same problem response, so existence is not
  disclosed.
- Stream authorization: the WebSocket authorizes at connect and reauthorizes
  periodically, dropping a subscriber whose access was revoked mid-stream.
- An append-only audit trail recording policy changes, credential use (never
  value), run launch and stop, approvals, and exports: actor, object ID,
  operation, outcome, and time — no content, path, or free-form field.
- Audit records are readable by an authorized administrator through the API.

## Non-goals

- Federated identity or SSO. Local accounts only; the port is what matters.
- Retention and deletion of audit records — P2-027 applies the class rules.

## Runtime reachability

Every `/api/v1` operation; `winch audit ls` for an administrator.

## Owned surfaces

`internal/application/authz/`, `internal/application/audit.go`,
`api/openapi/paths/audit.yaml`, `api/openapi/paths/identity.yaml`,
`internal/adapters/postgres/migrations/011_*.sql`, `cmd/winch/audit.go`.

## Demonstration

    $ winch run get $ID --token $OTHER_USER
    → expect: the same 404 body as a run that does not exist

    $ winch run watch $ID --token $MEMBER &
    $ winch workspace member rm <workspace> $MEMBER
    → expect: the stream closes within the reauthorization interval, with a
      resumable last-sequence indication and no further events

    $ winch audit ls --workspace <id>
    → expect: launch, stop, approval, and credential-use entries; no command,
      path, or payload text in any field

    $ winch audit ls --token $MEMBER
    → expect: refused

## Verification

- Standing scenario suite passes with authorization active.
- Authorization table test across every object type and role, including the
  negative case for each.
- Mid-stream revocation test with a fake clock.
- Audit field test asserting the record schema admits no free-form string.

## Acceptance criteria

- [ ] Every object operation authorizes the object; no operation trusts the ID.
- [ ] Absent and forbidden are indistinguishable to an unauthorized caller.
- [ ] A revoked subscriber is dropped from a live stream.
- [ ] Audit entries contain identifiers, enums, and times only.
- [ ] Audit records cannot be updated or deleted through any API path.

## Deferrals

| Deferred | Owning task |
|---|---|
| Applying retention and deletion rules to audit records | P2-027 |
| Runner machine identity, which is authenticated separately | P5-041 |

## Traces to

`docs/security.md` §9, §5 (telemetry and audit fields), T01, T04, LB01;
`docs/architecture.md` §4; `docs/roadmap.md` Phase 2
