# P2-058: Acquire harness credentials

**Phase:** 2 — Structured experience and second harness
**Shape:** capability
**Dependencies:** P1-054 (compile: the credential aggregate and secret-provider port the acquired token is stored through), P2-055 (semantic: an acquired credential is scoped to a workspace)

## Objective

A user logs in to a provider from the browser and the resulting token lands in
the secret provider as a reference — the goal `docs/architecture.md` §1 names
and no task previously owned.

## Scope

- Login mode declaration: each harness descriptor states which of the
  `docs/security.md` §7 patterns it supports, and the UI offers only those.
- **Host-mediated OAuth or device flow:** the control plane displays the
  provider URL and user code, polls for completion, stores the resulting token
  through the secret-provider port, and records only metadata.
- **User-supplied token entry:** accepted over TLS, stored immediately, then
  referenced by ID and never returned.
- Credential scope: provider plus workspace, so a credential cannot be used by a
  run in another workspace.
- Expiry and refresh where the provider supports it, with a visible expired
  state rather than a failing launch.
- Audit: acquisition, use, and revocation are recorded as facts without values.

## Non-goals

- Interactive harness login through the terminal. It is deferred as an
  architectural decision in `docs/roadmap.md` with an explicit trigger, and
  remains disabled by default per `docs/security.md` §7.
- Injecting credentials into a container — P3-032.
- Provider-side account management.

## Runtime reachability

`POST /api/v1/credentials/login` and the credential page in the web app;
`winch credential login`.

## Owned surfaces

`internal/application/login.go`, `internal/adapters/harness/*/login.go`,
`api/openapi/paths/credential-login.yaml`,
`web/src/features/credentials/LoginFlow.tsx`, `cmd/winch/credential_login.go`.

## Demonstration

    $ winch credential login --provider <provider> --workspace <id>
    → expect: a URL and a user code printed, then a stored credential ID once
      the flow completes — never the token

    # in the browser, the same flow:
    → expect: the code and a completion state; reloading mid-flow does not
      restart or leak it

    $ winch credential list --json | jq '.credentials[0]'
    → expect: provider, workspace, expiry, and reference metadata only

    $ winch run create --workspace <other> --credential <id>
    → expect: refused; a credential does not cross workspaces

    $ docker compose logs winchd | grep -Ec '<token-canary>'
    → expect: 0

## Verification

- Standing scenario suite passes with a login-acquired credential in the path.
- Flow tests against a stubbed provider: success, user denial, timeout,
  interrupted poll, and replayed callback.
- Cross-workspace use test.
- Canary tests across logs, traces, events, and API responses.

## Acceptance criteria

- [ ] At least the two non-interactive login patterns work end to end for one
      provider.
- [ ] A token value is never returned by any API after acquisition.
- [ ] A credential is unusable outside its workspace.
- [ ] An expired credential produces a clear state before launch, not a launch
      failure.
- [ ] Acquisition and use are audited without values.

## Deferrals

| Deferred | Owning task |
|---|---|
| Interactive harness login proxying | `docs/roadmap.md` deferred decision |
| Scoped short-lived injection into a sandbox | P3-032 |

## Traces to

`docs/security.md` §7, T10, LB05; `docs/architecture.md` §1, §6;
`docs/code-structure.md` §4; `docs/roadmap.md` Phase 2
