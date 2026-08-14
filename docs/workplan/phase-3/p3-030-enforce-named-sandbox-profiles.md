# P3-030: Enforce named sandbox profiles

**Phase:** 3 — Docker isolation
**Shape:** hardening
**Dependencies:** P3-029 (semantic: there must be a container driver whose controls a profile can constrain), P1-050 (contract: profile resolution lives in the driver registry this extends)

## Objective

A run selects an administrator-approved profile by name, and no request path can
assemble a weaker configuration out of low-level flags.

## Scope

- The three profiles in `docs/security.md` §4 — `local-trusted`,
  `container-standard`, `container-readonly` — defined as data, validated at
  startup, with the deployment refusing to start on an unknown or malformed
  profile.
- Profile resolution order per `docs/code-structure.md` §6: compiled defaults,
  system config, workspace policy, named profile, then only those per-run
  overrides the policy marks overridable.
- Overrides can only narrow. An override that would widen filesystem, network,
  credential, or resource permissions is refused, not clamped silently.
- Workspace policy can prohibit `local-trusted` entirely, and a shared
  deployment can prohibit it globally.
- The resolved effective profile is stored with the run and returned by the API,
  so the UI and the audit trail read the same value.

## Non-goals

- Displaying posture to a user — P3-033 renders what this resolves.
- Network allowlist semantics — P3-031 defines the fields; this task carries
  them through resolution.

## Runtime reachability

`winch profiles` and `winch run create --profile`; the resolved profile appears
in `GET /api/v1/runs/{id}`.

## Owned surfaces

`internal/application/profile/`, `deployments/profiles/*.yaml`,
`api/openapi/components/run.yaml` (resolved-profile fields),
`internal/adapters/postgres/migrations/013_*.sql`.

## Demonstration

    $ winch profiles --json | jq -r '.sandbox[].name'
    → expect: local-trusted, container-standard, container-readonly

    $ winch run create --profile container-readonly --harness fake --json | jq -r '.resolvedProfile.filesystem'
    → expect: read-only, with a scratch mount, matching the documented table

    $ winch run create --profile container-standard --override network=host
    → expect: refused with a stable code; widening is not expressible

    $ winch run create --profile container-standard --override memory=256m
    → expect: accepted; narrowing is allowed and recorded

    $ WINCH_ALLOW_LOCAL_TRUSTED=false winch run create --profile local-trusted
    → expect: refused, naming the policy that prohibits it

    $ winch run create --profile does-not-exist
    → expect: refused before launch

## Verification

- Standing scenario suite passes with profiles enforced for every substrate.
- Resolution table test across the five configuration layers.
- Negative test per widening dimension: filesystem, network, credentials, CPU,
  memory, process count, wall clock.
- Startup validation test rejecting a malformed profile definition.

## Acceptance criteria

- [ ] Every run records a resolved, named profile.
- [ ] No per-run override can widen any permission.
- [ ] `local-trusted` is prohibitable by workspace and by deployment policy.
- [ ] An unknown profile fails before any process or container starts.
- [ ] The stored resolved profile is the value the API returns.

## Deferrals

| Deferred | Owning task |
|---|---|
| Rendering the effective posture to the user | P3-033 |
| Enforcing the network fields these profiles declare | P3-031 |

## Traces to

`docs/security.md` §4, T02, LB02; `docs/code-structure.md` §6;
`docs/roadmap.md` Phase 3 exit
