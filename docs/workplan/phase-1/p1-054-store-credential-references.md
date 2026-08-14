# P1-054: Store credential references

**Phase:** 1 — Local single-user vertical slice
**Shape:** seam
**Dependencies:** P1-050 (contract: run creation binds a credential reference and launch resolves it), P1-011 (contract: the credential table is a new ordered migration)

## Objective

A user registers a provider token once; the run record thereafter holds an
opaque reference, and the value reaches only the harness launch — never the
database, an event, a log, or the browser.

## Scope

- The `Credential` aggregate from `docs/architecture.md` §6: id, owner,
  provider, secret reference. Metadata is persisted; the value never is.
- A real secret-provider adapter behind the existing `SecretReferenceStore`
  port in `internal/adapters/secrets/`, so the port stops having only test
  doubles. Start with one adapter — an encrypted local file keyed from
  configuration — chosen because it works in the compose stack without external
  infrastructure.
- API: create a credential by supplying a value once, list credential metadata,
  delete a credential. The value is never returned after entry, including to the
  creating request.
- Run creation optionally names credential references; `StartRun` resolves them
  immediately before `BuildLaunch` and discards the resolved values afterwards.
- `winch credential add|list|rm`, reading the value from stdin, never a flag.
- Secret canary test: a known value registered through this path occurs zero
  times in events, artifacts, default logs, traces, and API responses.

## Non-goals

- Acquiring a credential interactively — OAuth, device flow, and interactive
  harness login are P2-058.
- Scoped injection into a container, temporary credential files, and cleanup —
  P3-032. This task passes values in memory to `BuildLaunch` only.
- Keychain and vault adapters. The port exists; adding one is a later swap.

## Runtime reachability

`POST /api/v1/credentials` and `winch credential add` against the compose
stack; resolved during `StartRun` on the `local` sandbox profile.

## Owned surfaces

`internal/adapters/secrets/`, `internal/application/credential.go`,
`api/openapi/paths/credentials.yaml`,
`internal/adapters/postgres/migrations/007_*.sql`, `cmd/winch/credential.go`,
`web/src/features/credentials/`.

## Demonstration

    $ printf 'canary-value-do-not-log' | winch credential add --provider example --name default
    → expect: an ID and provider, and no value, in the response

    $ winch credential list
    → expect: metadata only; no field of any length that could hold the value

    $ ID=$(winch run create --harness fake --credential default --json | jq -r .id)
    $ winch run start $ID && winch run watch $ID | grep -c 'canary-value-do-not-log'
    → expect: 0

    $ docker compose logs winchd | grep -c 'canary-value-do-not-log'
    → expect: 0

    $ psql "$WINCH_DATABASE_URL" -c "select count(*) from run_events where payload::text like '%canary-value%'"
    → expect: 0

## Verification

- Standing scenario suite passes with a credential-bound run in the path.
- Adapter tests: encryption at rest, wrong key fails closed, deletion removes
  the stored value and the reference.
- Canary tests across events, artifacts, logs, traces, and API responses.
- A test asserting a resolved value is not retained after launch returns.

## Acceptance criteria

- [ ] The credential value is unreadable through any API after creation.
- [ ] `ResolvedCredentials` never appears in a `RunSpec`, a persisted event, or
      a log line.
- [ ] Deleting a credential removes the stored value, not only the reference.
- [ ] A run naming an unknown credential fails before launch with a stable code.
- [ ] Canary counts are zero in every store listed in Demonstration.

## Deferrals

| Deferred | Owning task |
|---|---|
| Host-mediated OAuth/device flow and token entry UI flows | P2-058 |
| Scoping credentials to a workspace | P2-055 |
| Temporary credential files, container injection, and cleanup | P3-032 |

## Traces to

`docs/architecture.md` §6; `docs/security.md` §7, §5 (`secret` class), LB05;
`docs/code-structure.md` §1, §3; `docs/roadmap.md` Phase 1
