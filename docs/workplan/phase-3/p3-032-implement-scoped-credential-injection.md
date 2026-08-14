# P3-032: Implement scoped credential injection

**Phase:** 3 — Docker isolation
**Shape:** hardening
**Dependencies:** P1-054 (compile: the resolved credential values being injected), P3-029 (semantic: injection into a container is what this task makes safe)

## Objective

A container receives exactly the credential it needs, for exactly as long as it
runs, through a mechanism that does not leave the value in the process
environment or on disk afterwards.

## Scope

- Injection as a temporary mounted file with a bounded lifetime, preferred over
  environment variables wherever the harness supports it; the environment path
  remains available only for harnesses that offer nothing safer, and says so in
  its descriptor.
- Scoping: only credentials the profile and workspace policy bind to this run
  are resolved, and only for the launch call.
- Cleanup on every exit path — success, failure, forced stop, daemon crash — with
  a restart sweep removing temporary material and reporting failures as security
  signals rather than warnings.
- Redaction of known credential values before persistence, on top of the
  supervisor's existing redaction step.
- Prohibited mounts enforced at construction: the user's home directory, SSH
  directory, cloud config, and agent sockets.
- Secret canaries in CI covering events, artifacts, exports, default logs, and
  traces.

## Non-goals

- Acquiring credentials — P1-054 and P2-058.
- A credential broker or proxy issuing short-lived upstream tokens; that is the
  communication-proxy deferred decision.

## Runtime reachability

Any container run whose profile binds a credential; the injected path appears in
the run's resolved configuration as a path, never a value.

## Owned surfaces

`internal/application/credential_injection.go`,
`internal/adapters/sandbox/docker/credentials.go`,
`test/integration/credentials/`.

## Demonstration

    $ printf 'canary-value-do-not-log' | winch credential add --provider example --name default
    $ ID=$(winch run create --profile container-standard --credential default --harness fake --json | jq -r .id)
    $ winch run start $ID
    $ docker exec $(docker ps -q -f label=winch.run=$ID) env | grep -c canary-value
    → expect: 0

    $ docker exec … cat /run/winch/credentials/default | head -c 6
    → expect: the value is present inside the container while it runs

    $ winch run stop $ID && ls /run/winch/credentials 2>/dev/null
    → expect: nothing left on the host

    $ docker compose kill winchd   # mid-run, then restart
    → expect: the sweep removes temporary material and records a security signal
      if it cannot

    $ for src in events artifacts exports logs traces; do winch canary-scan $src; done
    → expect: 0 occurrences in each

    $ winch run create --profile container-standard --mount $HOME/.ssh
    → expect: refused at construction

## Verification

- Standing scenario suite passes with an injected credential in the path.
- Cleanup tests on each exit path, including a killed daemon.
- Canary tests across every store listed above.
- Negative mount tests for each prohibited path.

## Acceptance criteria

- [ ] Credential values never appear in the container environment when a file
      mechanism is available.
- [ ] Temporary material is removed on every exit path, including a crash.
- [ ] A cleanup failure raises a content-free security signal, not a warning.
- [ ] Prohibited host paths cannot be mounted through any request path.
- [ ] Canary counts are zero in events, artifacts, exports, logs, and traces.

## Deferrals

| Deferred | Owning task |
|---|---|
| Showing which credential a run will receive, before launch | P3-033 |
| Short-lived broker-issued upstream tokens | `docs/roadmap.md` deferred decision |

## Traces to

`docs/security.md` §7, §5 (`secret` class), T10, LB05;
`docs/code-structure.md` §3; `docs/roadmap.md` Phase 3
