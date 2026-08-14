# P1-051: Add run commands to the operator CLI

**Phase:** 1 — Local single-user vertical slice
**Shape:** seam
**Dependencies:** P1-050 (semantic: the run API must answer before a client can drive it), P1-049 (compile: the `cmd/winch` shell this extends)

## Objective

An operator drives a complete run — create, start, watch, input, stop — against
a deployed daemon without a browser. This is the plan's maintained hands-on
surface: not a substitute for the web app, but what guarantees the system stays
drivable while the web app churns.

## Scope

- `winch health`, `winch run create|start|stop|input|get`, `winch run watch`
  (resumable `after_sequence` stream to stdout), `winch profiles` listing the
  registered harness and sandbox drivers with their capabilities.
- One file per command, so later phases append rather than edit.
- Configuration from flags and environment, sharing `internal/platform/config`;
  the API token is read from the environment or a file, never a flag.
- `--json` output on every command so the CLI is scriptable by tests.

## Non-goals

- Workflow, credential, and approval commands — added by P4-039, P1-054, and
  P2-056 respectively when those surfaces exist.
- Any behavior not reachable through the public HTTP API. The CLI is a client.

## Runtime reachability

`cmd/winch` against the compose stack or any deployment reachable over HTTP.

## Owned surfaces

`cmd/winch/` (command files, excluding `dev.go` from P1-049),
`docs/operations/cli.md`.

## Demonstration

    $ export WINCH_ENDPOINT=http://localhost:8080 WINCH_TOKEN=…
    $ winch health
    → expect: ok, with the daemon version and applied migration version
    $ ID=$(winch run create --workspace /workspace --harness fake --sandbox local --json | jq -r .id)
    $ winch run start $ID && winch run watch $ID &
    $ winch run input $ID --text 'echo hello'
    → expect: "hello" appears in the watch stream within a second
    $ winch run stop $ID
    → expect: watch prints a terminal lifecycle event and exits 0

    $ winch run watch $ID --after-sequence 0
    → expect: the full history replays in the same order, from persisted events

## Verification

- CLI tests against a stubbed HTTP server covering exit codes, `--json` shape,
  and resume after a dropped stream.
- A test asserting the token never appears in process arguments or usage output.

## Acceptance criteria

- [ ] Every Phase 1 run operation is reachable from the CLI.
- [ ] `watch` resumes from a supplied sequence without gaps or duplicates.
- [ ] Non-zero exit codes distinguish usage errors, API problems, and transport
      failures.
- [ ] Adding a command adds a file.

## Deferrals

| Deferred | Owning task |
|---|---|
| `winch credential` commands | P1-054 |
| `winch approve` commands | P2-056 |
| `winch workflow` commands | P4-039 |

## Traces to

`docs/architecture.md` §4; `docs/contracts.md` §5; `docs/roadmap.md` Phase 1
