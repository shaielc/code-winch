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

At this brief's derivation revision, P1-004 is implemented: the composition
root constructs an attach session with `ChildStatus: Down`, the session handler
returns that construction-time value, and the one-shot React fetch validates
and presents exact `Down`. The sandbox still starts no child. P1-006 therefore
extends behavior observed in source rather than supplying P1-004's predicted
contract, while its dependency remains required because both tasks own the
same wire and presentation surface.

## Scope

- Add a focused process owner under `cmd/winch-sandbox` and construct it before
  the HTTP handler. Start the sibling `fake-harness` executable with a non-empty
  fixture run ID, retain the command it starts, keep the child's standard input
  open so the default fixture remains alive, and wait for that exact child.
  Resolve the sibling from the running `winch-sandbox` executable directory:
  the image installs both in `/usr/local/bin`, while `make host-test-env` builds
  both in `bin/`; this single shipped profile needs no new configuration key.
- Publish status from the owner through a concurrency-safe read: it is `Down`
  before a successful `Start`, becomes `Running` only after `Start` succeeds,
  and becomes `Exited` only after waiting on that same command returns. A start
  error must abort sandbox startup without publishing `Running`. Do not scan
  the host process table or infer state from output, browser timing,
  `/healthz`, or a bare PID.
- Keep child output out of logs and responses because output transport is a
  later increment. On sandbox shutdown, close the retained input and signal
  the owned child, then wait for it inside the same five-second bounded
  shutdown path used by the HTTP server; cleanup must neither orphan the child
  nor wait without a bound. This increment does not parse, retain, publish, or
  display child output or accept operator input.
- Change the attach server's session dependency from a copied `Session` to a
  small current-session source, and have `/api/session` call it for every
  request. The composition root supplies the fixed posture plus the owner's
  current status. Preserve the one-shot browser fetch: a load or reload
  observes one authoritative response; no polling, streaming, or browser-side
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
unobserved exit can. `Exited` conveys no native exit interpretation or result,
and neither state conveys readiness or a writable input channel.

Use the shipped fake as a malleable test profile, not as a product protocol.
Logs, errors, and tests must not expose child output or submitted content.

## Non-goals

The manual `Exited` demonstration, input submission or child commands, output
transport/display, output retention, exit code/signal presentation, restart or
retry policy, readiness, provider semantics, PTY support, persistence, polling,
and broad page restructuring are excluded. They are not required to complete
this truthful owned-process status slice.

## Runtime reachability

The `winch-sandbox` composition root starts the fake harness that already ships
in the composed image. The existing `/api/session` route reads the owner's
current observation, and the existing one-shot fetch renders it at the branch
proxy attach URL. No provider account, auxiliary service, new public endpoint,
or host process scan is required.

At HEAD, `cmd/winch-sandbox/main.go` constructs one immutable `attach.Session`
with `ChildDown`; `attach.Server` stores that value and returns it unchanged.
`deployments/Dockerfile` and `make host-test-env` both place `fake-harness`
beside `winch-sandbox`, and the fake blocks on its open standard input unless
an exit/failure control is selected. Those observations make the sibling
executable and retained input the concrete maintained paths.

## Write set

- `cmd/winch-sandbox/main.go`
- `cmd/winch-sandbox/child.go`
- `cmd/winch-sandbox/child_test.go`
- `internal/adapters/transport/attach/session.go`
- `internal/adapters/transport/attach/server.go`
- `test/contract/attach/session_test.go`
- `test/e2e/scenario_sandbox_starts_test.go`
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

After the automated gates, use the repository's isolated host demonstration
environment (or the equivalent `make test-env` branch-proxy path when Docker is
the selected gate):

    $ make host-test-env
    $ curl -fsS http://127.0.0.1:7081/api/session
    → expect: one JSON response containing `"childStatus":"Running"`,
      `"profile":"container-standard"`, and `"unenforcedControls":["network-egress"]`
    $ python3 -m webbrowser http://127.0.0.1:7081/
    → expect: the loaded page shows **Child status: Running**, the unchanged
      posture and network-egress disclosure, and a usable local composer
    $ make host-test-env-down
    → expect: the sandbox and its owned fake process are gone

Reload the page if it was open before startup because the fetch is one-shot.
The default fake, launched without `-early-exit`, a transcript, or failure
injection, blocks reading its retained standard input after its startup records,
so `Running` remains observable without a harness command. Record the endpoint,
page, composer, and teardown observations separately.

This proves that the sandbox owns a live fake process and transports the
observation through the deployed page; it does not prove fake readiness, input
delivery, output handling, provider compatibility, or exit interpretation.

## Verification

The gates in `docs/testing.md` apply; say which option of each you ran. Checks
specific to this task:

- Process-owner tests cover successful launch reporting `Running`, a start
  failure never reporting `Running`, and an exited child eventually reporting
  `Exited`, plus bounded cleanup reaping a live child. Use a test helper process,
  not timing assumptions about the production fake, for deterministic state
  transitions.
- The attach contract tests cover exact `Down`, `Running`, and `Exited` wire
  values while preserving posture, method rejection, and unknown-API behavior.
- `go test ./cmd/winch-sandbox ./internal/adapters/transport/attach ./test/contract/attach`
  covers process state transitions, per-request status reads, cleanup, and the
  additive wire contract.
- `npm test -- --run src/attach/App.test.tsx` from `web/` covers exact
  `Down`/`Running`/`Exited` presentation and `Unknown` on pending, request
  failure, invalid JSON, missing status, and unsupported values, with posture
  and composer assertions kept green.
- The per-branch demonstration above, with endpoint, page, posture, composer,
  and teardown observations recorded separately.
- The standing `test/e2e` scenario expects the composed sandbox session to
  report `Running`, so both host `make check` and Docker `make test-cycle`
  exercise the newly real child substrate through the existing entrypoint.

The fake-child tests establish deterministic process ownership and state
transitions, not compatibility with a real vendor CLI.

## Acceptance criteria

- [ ] The maintained sandbox launches and owns the shipped fake and cleans it
      up within the bounded shutdown path; a failed start aborts startup and
      cannot produce `Running`, and child output is not logged or returned.
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
      failed launch, observed exit, bounded cleanup, live per-request wire
      values, the standing scenario, and React validation.
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
