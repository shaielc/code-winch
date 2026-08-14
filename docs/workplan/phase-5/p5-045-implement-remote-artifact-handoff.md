# P5-045: Implement remote artifact handoff

**Phase:** 5 — Remote runners and hardening
**Shape:** capability
**Dependencies:** P5-042 (compile: the runner channel this transfer path extends), P2-023 (compile: the artifact store and digest contract content lands in)

## Objective

An artifact produced on a runner host arrives intact in the control plane's
store, verified by digest, without passing through the event stream.

## Scope

- Content-addressed transfer: the runner announces digest, size, and media type;
  the control plane accepts, or skips content it already holds.
- Either direct upload or a signed-storage handoff, chosen by configuration, so
  a deployment with object storage does not proxy bytes through the daemon.
- Resumable transfer with bounded concurrency and bounded memory; a partial
  transfer resumes rather than restarting.
- Digest verification on arrival; a mismatch is rejected and recorded, never
  stored.
- Cleanup of runner-side staged content on success, on failure, and after a
  crash, including deletion propagation when P2-027 deletes an artifact whose
  copy still exists on a runner.
- Size and count limits per run, refused with a stable code.

## Non-goals

- Artifact rendering — P2-023 owns the changes view.
- Cross-runner artifact sharing.

## Runtime reachability

Any remote run producing artifacts; `winch run artifacts --download` reads the
transferred content.

## Owned surfaces

`internal/adapters/transport/runnerrpc/artifact.go`,
`internal/application/artifact_transfer.go`, `test/integration/artifact/`.

## Demonstration

    $ ID=$(winch run create --harness fake --scenario writes-large-files --json | jq -r .id)
    $ winch run start $ID   # placed on a remote runner
    $ winch run artifacts $ID --download <artifact-id> | sha256sum
    → expect: matches the digest the runner announced

    $ docker network disconnect <net> <runner> && sleep 3 && docker network connect …
    → expect: the transfer resumes from where it stopped, not from zero

    # with a runner that corrupts a byte mid-transfer:
    → expect: rejected on digest mismatch, recorded, and not stored

    $ winch delete run $ID && docker compose exec runner ls <staging-dir>
    → expect: empty; deletion propagated to the runner's copy

    $ docker stats --no-stream <daemon-container>
    → expect: bounded memory during a multi-gigabyte transfer

## Verification

- Standing scenario suite passes with artifacts produced remotely.
- Resume tests interrupting at start, middle, and end of a transfer.
- Corruption and digest-mismatch tests.
- Memory-bound test under a large transfer.
- Deletion propagation test including a runner offline at deletion time.

## Acceptance criteria

- [ ] Transferred content matches its announced digest or is rejected.
- [ ] Transfers resume rather than restart.
- [ ] Bytes do not pass through the event stream.
- [ ] Deletion reaches runner-side copies, including after a delay.
- [ ] Memory stays bounded regardless of artifact size.

## Deferrals

| Deferred | Owning task |
|---|---|
| Backup and restore of transferred artifacts | P5-046 |

## Traces to

`docs/contracts.md` §4 (artifact transfer); `docs/architecture.md` §6;
`docs/security.md` T14; `docs/roadmap.md` Phase 5
