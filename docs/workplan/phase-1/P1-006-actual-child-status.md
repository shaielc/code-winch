# P1-006: Report actual child status

**Phase:** 1 — Sandbox exchange
**Shape:** capability
**Dependencies:** P1-004 (contract: this task extends the additive
`GET /api/session` `childStatus` wire values and the validated `Down`
presentation defined by P1-004)

## Objective

Launch the shipped fake harness as a child owned by `winch-sandbox` and make
the existing session endpoint and attach page truthfully report **Running**
while that exact process has not been observed to exit, then **Exited** after
its exit is observed. This status describes process ownership and observation
only; it is not a health, readiness, input-availability, or provider claim.

## Source baseline

At refinement HEAD, P1-004's dependency claim holds. `cmd/winch-sandbox/main.go`
constructs an `attach.Session` with `ChildStatus: attach.ChildDown` before
starting the HTTP server; `internal/adapters/transport/attach/server.go` stores
that value and serializes the same snapshot on every `/api/session` request.
`internal/adapters/transport/attach/session.go` defines only `ChildDown`, and
`web/src/attach/App.tsx` promotes the initial `Unknown` value only when the
one-shot response contains exact `Down`. The focused Go and React tests assert
those boundaries. Source inspection also confirms that the composition root
does not use `os/exec`, while `deployments/Dockerfile` installs
`fake-harness` in `/usr/local/bin` and starts `winch-sandbox` with that directory
on Alpine's default `PATH`.

The refinement observations were:

```text
$ go test ./test/contract/attach ./cmd/winch-sandbox
ok github.com/shaielc/code-winch/test/contract/attach
ok github.com/shaielc/code-winch/cmd/winch-sandbox
$ cd web && npm test -- --run src/attach/App.test.tsx
the focused App suite passes with 10 tests
$ curl -fsS http://127.0.0.1:7090/<refinement-branch>/sandbox/api/session
curl: (7) connection refused
```

No branch test environment was running for the pre-change request, so the curl
result is only an environment observation. P1-006 extends the verified P1-004
contract; it does not need to reconstruct missing prerequisite behavior.

## Scope

- Add a focused process owner under `cmd/winch-sandbox/`. It constructs
  `exec.Command("fake-harness", "-run-id", "sandbox")`, obtains and retains the
  child's standard-input pipe so the default fake remains alive, discards the
  currently unowned output streams without parsing them, and retains the
  `*exec.Cmd` rather than a bare PID. Keep command construction injectable in
  package tests so success, start failure, and deterministic early exit use
  test processes rather than the installed image binary.
- Give the owner a mutex-protected `Status() attach.ChildStatus` snapshot.
  Initialize it to `Down`; publish `Running` only after `Cmd.Start` succeeds;
  call `Wait` exactly once in a goroutine and publish `Exited` only after it
  returns. A failed start leaves `Down` and returns its error to `main` so the
  sandbox fails closed instead of serving a false state. Do not scan the host
  process table or infer status from output, browser timing, `/healthz`, or a
  PID without the owned handle.
- Give the owner bounded cleanup that closes the retained input and terminates
  the exact child if necessary, waits for the existing `Wait` goroutine, and
  escalates to `Process.Kill` when the shutdown context expires. Wire that
  cleanup into the composition root's existing five-second signal path before
  `main` returns. Tests must register cleanup immediately after a successful
  start so failures do not leak helper processes.
- Change `attach.New` to receive the immutable posture plus a required status
  function (or an equivalently narrow source), and construct a fresh `Session`
  for every `/api/session` request. The attach adapter must only invoke and
  serialize the source; process ownership and synchronization stay in
  `cmd/winch-sandbox`. Update all constructor call sites explicitly so neither
  a nil/default source nor a copied startup value can invent state.
- Start the owner before serving HTTP and pass its `Status` method to the attach
  server. Preserve the one-shot browser fetch: a load or reload observes one
  authoritative session response; no polling, streaming, or browser-side
  inference is added.
- Extend the P1-004 React validation to accept exact `Running` and `Exited`
  values while retaining exact `Down` and the `Unknown` fallback for pending,
  failed, malformed, missing, or unsupported values. Preserve the child-status
  label/value identity, posture disclosure, local composer, static serving,
  loopback-only publication, CLI status behavior, and deployment.
- Document the expanded `GET /api/session` member in `docs/contracts.md` §8:
  `Down` means no child is owned, `Running` means the owned child started and no
  exit has been observed, and `Exited` means that child's exit was observed.

The implementation may factor process ownership into focused files in
`cmd/winch-sandbox`; it must not introduce a generic harness adapter, runner
domain model, output transport, or provider-specific branch in the attach
transport.

## Required constraints

The composition root is the authority for this child because it launches,
retains, and waits for the process. Status publication must be safe under
concurrent session requests and child exit. A failed `Start` cannot transiently
or permanently produce `Running`; only successful start followed by an
unobserved exit can. `Wait` errors and nonzero exits still produce `Exited`, not
an interpretation of why the process ended. `Exited` conveys no native exit
result, and neither state conveys readiness or a writable input channel.

Use the shipped fake as a malleable test profile, not as a product protocol.
Logs, errors, and tests must not expose child output or submitted content.

## Non-goals

The manual `Exited` demonstration, a configurable executable/profile surface,
input submission or child commands, output transport/display/retention, exit
code or signal presentation, restart or retry policy, readiness, provider
semantics, PTY support, persistence, polling, and broad page restructuring are
excluded. Discarding the child's output in this increment prevents an unread
pipe from coupling lifecycle to output; it is not output processing. These
items are not required to complete this truthful owned-process status slice.

## Runtime reachability

The `winch-sandbox` composition root starts the fake harness that already ships
in the composed image. The existing `/api/session` route reads the owner's
current observation, and the existing one-shot fetch renders it at the branch
proxy attach URL. No provider account, auxiliary service, new public endpoint,
or host process scan is required.

## Write set

- `cmd/winch-sandbox/main.go`
- `cmd/winch-sandbox/child.go`
- `cmd/winch-sandbox/child_test.go`
- `internal/adapters/transport/attach/session.go`
- `internal/adapters/transport/attach/server.go`
- `test/contract/attach/session_test.go`
- `web/src/attach/App.tsx`
- `web/src/attach/App.test.tsx`
- `docs/contracts.md`

`deployments/Dockerfile` already ships the fake at the required default path;
change it only if source inspection at implementation time disproves that
observation. No fake-harness protocol, compose topology, CLI, output, input, or
persistence file is forecast.

## Contract surfaces

Extends P1-004's additive `GET /api/session` `childStatus` member from the sole
valid `Down` value to the case-sensitive set `Down | Running | Exited`, with the
owned-process meanings above. The endpoint continues to return `profile` and
`unenforcedControls`. Extends the existing accessible `Child status` region to
render each validated value and retain `Unknown` for every missing or invalid
value. The implementation must update `docs/contracts.md` in the same change.

Consumes P1-004's `Down` meaning, validation boundary, and presentation
identity. It owns the minimum fake-child launch, handle bookkeeping, exit
observation, and cleanup needed to supply the new values. It owns no readiness,
input availability, output semantics, native exit interpretation, run
lifecycle, or provider contract.

## Demonstration

After the automated gates, run the sandbox in the branch-isolated environment
described in `docs/testing.md` §Demonstrations, not the operator deployment:

```sh
make test-env
TEST_PROJECT=$(printf '%s' "$(git rev-parse --abbrev-ref HEAD)" | tr 'A-Z' 'a-z' | tr -c 'a-z0-9_-' '-')
ATTACH_URL="http://127.0.0.1:${PROXY_PORT:-7090}/${TEST_PROJECT}/sandbox"
for attempt in $(seq 1 20); do
  SESSION=$(curl -fsS "$ATTACH_URL/api/session") && break
  sleep 0.25
done
printf '%s\n' "$SESSION"
```

Observe exact `childStatus: "Running"` alongside `profile:
"container-standard"` and `network-egress`. Open `$ATTACH_URL/`; reload it if
it was loaded earlier because the fetch is one-shot. Observe the independently
labeled **Child status: Running**, unchanged posture disclosure, and a usable
local composer. Then tear down only this branch's deployment even if an earlier
observation failed:

```sh
make test-env-down
```

The default fake, launched without `-early-exit`, a transcript, or failure
injection, blocks reading its retained open standard input after its startup
records, so the owned process stays alive for the observation without any
harness command. The bounded retry avoids treating proxy startup timing as
child state; it must not accept a response whose exact status is anything but
`Running`.

This proves that the sandbox owns a live fake process and transports the
observation through the deployed page; it does not prove fake readiness, input
delivery, output handling, provider compatibility, or exit interpretation.

## Verification

The gates in `docs/testing.md` apply; say which option of each you ran. Checks
specific to this task:

- `go test -race ./cmd/winch-sandbox` covers successful launch reporting
  `Running`, a start failure remaining `Down`, an early and a nonzero exit
  eventually reporting `Exited`, concurrent status reads during exit, and
  bounded cleanup/reaping without a leaked helper process.
- The attach contract tests cover exact `Down`, `Running`, and `Exited` wire
  values and a mutable test source observed across two requests while
  preserving posture, method rejection, and unknown-API behavior. Run
  `go test ./test/contract/attach -run
  'TestEndpointContracts|TestSessionReadsCurrentChildStatus|TestRejectsMethodsAndUnknownAPI'`.
- A focused `cmd/winch-sandbox` server-construction test proves the composition
  passes the live owner source rather than copying its startup value.
- `npm test -- --run src/attach/App.test.tsx` from `web/` covers exact
  `Down`/`Running`/`Exited` presentation and `Unknown` on pending, request
  failure, invalid JSON, missing status, and unsupported values, with posture
  and composer assertions kept green.
- The per-branch demonstration above, with endpoint, page, posture, composer,
  and teardown observations recorded separately.

The fake-child tests establish deterministic process ownership and state
transitions, not compatibility with a real vendor CLI.

## Acceptance criteria

- [ ] The maintained sandbox launches and owns the shipped fake and cleans it
      up within the bounded shutdown path; start failure leaves the owner
      `Down`, fails sandbox startup, and cannot produce `Running`.
- [ ] A concurrency-safe observation reports `Running` only after successful
      start and before observed exit, then `Exited` after `Wait` returns,
      without host process scans, browser inference, output inference, or
      readiness/input claims.
- [ ] `GET /api/session`, `docs/contracts.md`, and focused server tests agree on
      exact `Down | Running | Exited` meanings and preserve posture and route
      behavior.
- [ ] The existing accessible child-status region renders each validated value
      and retains `Unknown` for pending, failed, malformed, missing, or unknown
      data while preserving posture and the local composer.
- [ ] The default fake configuration keeps `Running` observable without input
      commands or output handling, and automated tests cover successful launch,
      failed launch, concurrent observation, graceful/forced cleanup, early and
      nonzero observed exit, live wire values, and React validation.
- [ ] The branch-isolated hand check observes `Running` at the endpoint and on
      the loaded/reloaded page, retains posture/composer behavior, tears down
      its own environment, and states the limits of fake-process evidence.

## Deferrals

None. The manual `Exited` demonstration and the later output/input capabilities
are excluded by the decided Phase 1 increment boundaries, not unfinished work
required for this owned-process status capability.

## Traces to

`docs/workplan/phase-1/README.md` increment 3; P1-004;
`docs/roadmap.md` §§1 and 3, Stage 0; `docs/contracts.md` §§8–9; ADR-0005;
ADR-0006.
