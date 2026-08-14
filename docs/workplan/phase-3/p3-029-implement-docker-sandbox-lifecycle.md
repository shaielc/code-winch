# P3-029: Implement Docker sandbox lifecycle

**Phase:** 3 — Docker isolation
**Shape:** swap
**Dependencies:** P3-028 (semantic: there must be a disposable preparation to mount), P1-049 (contract: the sandbox port including attached I/O is what this driver implements), P0-008 (contract: the shared sandbox suite it must pass)

## Objective

The standing scenarios pass unchanged with runs executing inside containers
instead of host processes.

## Scope

- A `docker` sandbox driver implementing the full port — prepare, start, attach,
  inspect, resize, stop, cleanup — and declaring honest capabilities including
  `container` isolation, network policy, and resource limits.
- The `docs/security.md` §6 baseline where the engine supports it: non-root
  user, pinned image digest with recorded provenance, dropped capabilities,
  `no-new-privileges`, default seccomp, read-only root with explicit writable
  mounts, and CPU, memory, process-count, disk, and wall-clock limits.
- Prohibited outright: privileged mode, host PID or IPC namespaces, device
  passthrough, and any mount of the Docker socket. These are refused at
  construction, not merely omitted.
- Deterministic labels enabling reconciliation, plus an orphan sweep that
  removes containers this deployment owns after a crash.
- Attached I/O over the container's TTY, satisfying the same contract cases the
  local driver passes.
- Opt-in integration tests, skipped with a clear message when no engine is
  available, and run in CI where one is.

## Non-goals

- Profiles as a policy surface — P3-030 turns these controls into named,
  administrator-approved profiles.
- Network allowlists — P3-031. This task denies or permits wholesale.
- Rootless Podman, microVMs, or Kubernetes; the port makes them later swaps.

## Runtime reachability

`winch run create --sandbox docker` on the compose stack, and `make e2e
PROFILE=docker`.

## Owned surfaces

`internal/adapters/sandbox/docker/`, `deployments/Dockerfile.sandbox`,
`test/integration/docker/`, `.github/workflows/docker.yml`.

## Demonstration

    $ make e2e PROFILE=docker
    → expect: every standing scenario passes, unchanged
    $ make e2e PROFILE=fake
    → expect: still passes; the fake profile is not regressed

    $ ID=$(winch run create --sandbox docker --harness fake --json | jq -r .id)
    $ winch run start $ID && docker inspect $(docker ps -q -f label=winch.run=$ID) \
        --format '{{.HostConfig.Privileged}} {{.Config.User}} {{.HostConfig.ReadonlyRootfs}}'
    → expect: false, a non-root user, true

    $ winch run create --sandbox docker --privileged
    → expect: refused; there is no request path that reaches privileged mode

    $ docker compose kill winchd   # mid-run, then restart
    → expect: the sweep removes the labelled container; `docker ps` is clean

## Verification

- Standing scenario suite passes against the Docker substrate and the fake one.
- Shared sandbox contract suite passes for the Docker driver with no exemption.
- Resource-exhaustion tests: fork bomb, memory bomb, disk fill — each bounded by
  the container, not the host.
- Forced-stop and orphan-cleanup tests after a killed daemon.

## Acceptance criteria

- [ ] The same scenarios pass against local, fake, and Docker substrates.
- [ ] Escape-prone options cannot be reached through any request path.
- [ ] Image digests are pinned and provenance is recorded with the run.
- [ ] Cleanup survives daemon failure and leaves no owned container or volume.
- [ ] Declared capabilities match what the driver actually enforces.

## Deferrals

| Deferred | Owning task |
|---|---|
| Named administrator-approved profiles over these controls | P3-030 |
| Egress deny-by-default and allowlisting | P3-031 |
| Scanning this image and adding it to the SBOM | P3-060 |

## Traces to

`docs/security.md` §6, T06, T07, LB03; `docs/architecture.md` §4;
`docs/decisions/0003-capability-based-adapters.md`; `docs/roadmap.md` Phase 3
