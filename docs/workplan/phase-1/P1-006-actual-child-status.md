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

- Launch the shipped fake harness as a child owned by `winch-sandbox`, retain
  the process handle, and observe that exact process's exit through a
  concurrency-safe status observation. Do not scan the host process table or
  infer status from output, browser timing, `/healthz`, or a PID without the
  owned handle.
- A failed start cannot produce `Running`. Clean up the owned child on sandbox
  shutdown within the existing bounded shutdown path. This increment does not
  parse, retain, publish, or display child output or accept input.
- Extend P1-004's session status source so each request reads the current owned
  process observation rather than a construction-time string. Preserve the
  one-shot browser fetch: a load or reload observes one authoritative session
  response; no polling, streaming, or browser-side inference is added.
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

## Write set

- `cmd/winch-sandbox/main.go`
- one focused process-owner implementation and test under `cmd/winch-sandbox/`
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

After the automated gates, run the sandbox in an isolated environment as
described in `docs/testing.md` §Demonstrations. Once the session reports
`Running`, load the attach page; reload it if it was loaded earlier, because the
fetch is one-shot. Observe **Child status: Running**, the unchanged
`container-standard` posture and `network-egress` disclosure, and a usable local
composer. The default fake, launched without `-early-exit`, a transcript, or
failure injection, blocks reading its open standard input after its startup
records, so the owned process stays alive for the observation without any
harness command.

This proves that the sandbox owns a live fake process and transports the
observation through the deployed page; it does not prove fake readiness, input
delivery, output handling, provider compatibility, or exit interpretation.

## Verification

The gates in `docs/testing.md` apply; say which option of each you ran. Checks
specific to this task:

- Process-owner tests cover successful launch reporting `Running`, a start
  failure never reporting `Running`, and an exited child eventually reporting
  `Exited`.
- The attach contract tests cover exact `Down`, `Running`, and `Exited` wire
  values while preserving posture, method rejection, and unknown-API behavior.
- `go test ./cmd/winch-sandbox` proves the server reads the live status source
  rather than a copied startup value.
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
      up; a failed start cannot produce `Running`.
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
      failed launch, observed exit, wire values, and React validation.
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
