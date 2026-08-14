# P3-060: Scan dependencies and publish an SBOM

**Phase:** 3 — Docker isolation
**Shape:** hardening
**Dependencies:** None — its inputs are the images and dependency manifests that already exist, so it can be taken at any point

## Objective

A pull request that introduces a vulnerable dependency or image layer fails, and
every release publishes an SBOM — closing LB08, which `docs/security.md` maps to
a task whose scope never covered it.

## Scope

- Dependency scanning for the Go module graph and the web workspace, running on
  every pull request, failing on severities above a configured threshold.
- Image scanning for every image the repository builds — `winchd`, `web`, and
  the sandbox image from P3-029 — against pinned digests.
- SBOM generation per image and for the repository, published as a build
  artifact and attached to a release.
- An exception mechanism with teeth: a suppressed finding names an accountable
  owner, an expiry date, and a reason; an expired suppression fails the build.
- Provenance recording for pinned base image digests, so the value P3-029 stores
  with a run is verifiable against what CI built.
- A documented triage path so a scanner outage does not silently pass.

## Non-goals

- Restricting adapter or renderer loading, which `docs/code-structure.md` §4
  already settles by prohibiting in-process plugins.
- Runtime intrusion detection.

## Runtime reachability

CI on every pull request; `make sbom` locally; the SBOM artifact attached to a
release and referenced by the launch-blocker evidence.

## Owned surfaces

`.github/workflows/supply-chain.yml`, `Makefile` (`sbom`, `scan` targets),
`security/suppressions.yaml`, `docs/operations/supply-chain.md`.

## Demonstration

    $ make scan
    → expect: a report per ecosystem, exit non-zero when a finding exceeds the
      threshold

    $ make sbom && jq '.components | length' sbom/winchd.json
    → expect: a non-zero component count naming the pinned base image digest

    # on a branch that adds a known-vulnerable dependency:
    → expect: the pull request check fails and names the package and advisory

    # with a suppression whose expiry has passed:
    → expect: the build fails on the expired suppression, not on the finding

    $ grep -r "$(jq -r .baseImageDigest sbom/winchd.json)" deployments/
    → expect: a match; what CI scanned is what deployment pins

## Verification

- A fixture branch with a deliberately vulnerable dependency fails the check.
- An expired-suppression fixture fails the build.
- SBOM schema validation and a component-count floor.
- A test asserting the scanner's failure to run is a build failure, not a pass.

## Acceptance criteria

- [ ] Dependency and image scanning run on every pull request.
- [ ] An SBOM is produced for every built image and for the repository.
- [ ] Suppressions carry an owner, reason, and expiry, and expire loudly.
- [ ] Pinned digests in `deployments/` match what was scanned.
- [ ] LB08 has recordable evidence produced by this pipeline.

## Deferrals

| Deferred | Owning task |
|---|---|
| Adding the sandbox image to the scanned set | P3-029 |

## Traces to

`docs/security.md` T13, LB08, §12; `docs/roadmap.md` Phase 3
