# P3-031: Enforce container network policy

**Phase:** 3 — Docker isolation
**Shape:** hardening
**Dependencies:** P3-029 (compile: the container network the policy is applied to), P3-030 (contract: the allowlist is a profile field, not a per-run flag)

## Objective

A container denied egress cannot reach the network, and one with an allowlist
can reach exactly the approved destinations and nothing else — including cloud
metadata.

## Scope

- Deny-by-default egress on a dedicated per-run network, enforced outside the
  container's own namespace so a root process inside cannot undo it.
- Domain and port allowlists from the profile, with resolved-IP checks so a
  permitted name that resolves to a denied address is still blocked.
- Link-local, loopback-to-host, and cloud-metadata address ranges blocked
  unconditionally, in every profile, including allowlisted ones.
- DNS constrained to the approved resolver, with rebinding checks between
  resolution and connection.
- Content-free audit facts for blocked and allowed connections: destination
  class, port, decision, and count — never a URL, query string, or header.
- Record the acknowledged limit: without TLS interception, enforcement is based
  on DNS, IP, and SNI, and cannot identify end-to-end TLS content.

## Non-goals

- A communication or API proxy. That is an optional adapter capability and a
  deferred decision in `docs/roadmap.md`.
- TLS interception.
- Ingress. Nothing dials into a sandbox.

## Runtime reachability

Any run on `container-standard` or `container-readonly`; blocked attempts appear
as audit facts in `winch audit ls`.

## Owned surfaces

`internal/adapters/sandbox/docker/network.go`,
`deployments/profiles/*.yaml` (network fields), `test/integration/network/`.

## Demonstration

    $ ID=$(winch run create --profile container-readonly --harness fake --scenario probes-network --json | jq -r .id)
    $ winch run start $ID && winch run watch $ID | grep -c 'connection refused\|blocked'
    → expect: every probe blocked

    $ docker exec $(docker ps -q -f label=winch.run=$ID) curl -sS --max-time 3 http://169.254.169.254/
    → expect: fails, on every profile, including one with an allowlist

    # with an allowlist of example.com:443:
    $ docker exec … curl -sS https://example.com >/dev/null && echo allowed
    $ docker exec … curl -sS https://other.example >/dev/null || echo blocked
    → expect: allowed, then blocked

    # with a name in the allowlist resolving to a denied address:
    → expect: blocked at connect time, not merely at resolution

    $ winch audit ls --run $ID
    → expect: decision facts with destination class and port; no URL or header

## Verification

- Standing scenario suite passes on both container profiles.
- Egress test matrix: denied, allowlisted, metadata, link-local, loopback,
  alternate resolver, and rebinding.
- A test asserting a root process inside the container cannot remove the policy.
- Audit field test rejecting any free-form value.

## Acceptance criteria

- [ ] Deny-by-default holds even for a root process inside the container.
- [ ] Metadata and link-local ranges are unreachable on every profile.
- [ ] An allowlisted name resolving to a denied address is blocked.
- [ ] Blocked and allowed decisions are audited without content.
- [ ] The DNS/IP/SNI limitation is documented alongside the control.

## Deferrals

| Deferred | Owning task |
|---|---|
| Communication/API proxy with credential attachment | `docs/roadmap.md` deferred decision |
| Showing network posture to the user before launch | P3-033 |

## Traces to

`docs/security.md` §6, §8, T09, LB04; `docs/roadmap.md` Phase 3
