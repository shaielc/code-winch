# P2-024: Implement approval binding and policy evaluation

**Phase:** 2 — Structured experience and second harness
**Shape:** capability
**Dependencies:** P2-021 (semantic: approval request and resolution events must be normalized before they can be bound)

## Objective

An approval request becomes a durable, single-use decision bound to the exact
operation it describes, auto-resolved by policy where policy is decisive.

## Scope

- The approval record: id, run, operation digest, requested scope, expiry,
  state, and resolution. Single-use and immutable once resolved.
- Binding to the exact operation: the digest covers the command, working
  directory, and file or network scope the request names. A mismatch at
  resolution time is refused.
- Expiry: an unanswered approval expires, the run observes the expiry, and the
  outcome is auditable.
- A policy engine evaluating deployment, workspace, and sandbox-profile rules
  into auto-deny, require-user, or narrowly-scoped allow. Effective policy is
  the intersection, never the union.
- An approval whose operation exceeds the active sandbox profile is flagged as
  such in the record, so the UI can say so.

## Non-goals

- Submitting a decision. The API and UI round trip is P2-056; this task
  auto-resolves by policy and exposes the record.
- Workflow-level approval steps — P4-038.
- Workspace-level policy storage — P2-055 owns the workspace policy this reads.

## Runtime reachability

`GET /api/v1/runs/{id}/approvals` on the compose stack; auto-resolution is
observable in `winch run watch` as an approval resolution event.

## Owned surfaces

`internal/application/approval.go`, `internal/application/policy/`,
`api/openapi/paths/approvals.yaml`,
`internal/adapters/postgres/migrations/009_*.sql`.

## Demonstration

    $ ID=$(winch run create --harness fake --scenario asks-approval --json | jq -r .id)
    $ winch run start $ID && winch run watch $ID --json | jq -r 'select(.kind|startswith("approval"))'
    → expect: an approval request event carrying an approval ID and digest

    $ curl -fsS localhost:8080/api/v1/runs/$ID/approvals | jq '.approvals[0]'
    → expect: state pending, an expiry, and the exact operation described

    # with a deny-by-default policy configured:
    $ winch run create --harness fake --scenario asks-approval --policy deny-network
    → expect: auto-denied without user interaction; the resolution event names
      the deciding rule, not free-form text

    # after the configured expiry elapses:
    → expect: state expired; a later resolution attempt is refused

## Verification

- Standing scenario suite passes with an approval scenario added.
- Digest binding tests: a resolution against a changed operation is refused.
- Fake-clock expiry tests; double-resolution and replay tests.
- Policy intersection table test across deployment, workspace, and profile rules.

## Acceptance criteria

- [ ] An approval ID is single-use and cannot be replayed.
- [ ] A resolution against a mismatched operation digest is refused.
- [ ] Expiry is durable and survives a daemon restart.
- [ ] Policy narrows permissions; no combination widens them.
- [ ] Approval records and audit entries carry no command content beyond the
      scoped operation description the user must see.

## Deferrals

| Deferred | Owning task |
|---|---|
| Answering an approval from the API, CLI, or browser | P2-056 |
| Workspace policy as a persisted aggregate | P2-055 |
| `approval.wait` as a workflow step | P4-038 |

## Traces to

`docs/security.md` §10, T11; `docs/contracts.md` §3;
`docs/architecture.md` §4; `docs/roadmap.md` Phase 2
