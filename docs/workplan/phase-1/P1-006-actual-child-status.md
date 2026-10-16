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

## Scope

- Add a small process owner in `cmd/winch-sandbox` that starts
  `/usr/local/bin/fake-harness -run-id sandbox`, retains the returned
  process handle, waits for that exact process, and exposes a concurrency-safe
  child-status observation. Do not scan the host process table or infer status
  from output, browser timing, `/healthz`, or a PID without the owned handle.
  Keep the executable path injectable at the process-owner constructor for
  focused helper-process and start-failure tests; do not add an operator-facing
  configuration key for a generic vendor command in this increment.
- Keep the child's standard input open for its lifetime so the shipped fake's
  normal read loop remains alive without adding an input command or protocol.
  Direct its output to a bounded sink such as `io.Discard`: this increment does
  not parse, retain, publish, or display child output, and must not allow an
  unread pipe to stall the child.
- Start the child before serving HTTP. A start error prevents the sandbox from
  advertising `Running` and fails startup with a content-free error. Once start
  succeeds, publish `Running` before the waiter can publish `Exited`; the waiter
  records `Exited` only after `Wait` returns, regardless of exit code.
- On sandbox shutdown, close the owned input, signal the child, and wait for its
  exit within the existing bounded shutdown path. Cleanup must not orphan the
  owned process, race a second `Wait`, or turn ordinary signal-driven shutdown
  into a fatal error. No restart or retry is introduced.
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
The default launch must block on its open standard input after emitting startup
text, which keeps `Running` observable without inventing a command. Do not rely
on its private JSON-lines output, and do not add a fake-only HTTP control to
force exit. Logs, errors, and tests must not expose child output or submitted
content.

## Non-goals

The manual `Exited` demonstration, input submission or child commands, output
transport/display, output retention, exit code/signal presentation, restart or
retry policy, readiness, provider semantics, PTY support, persistence, polling,
and broad page restructuring are excluded. They are not required to complete
this truthful owned-process status slice.

## Reachability

The `winch-sandbox` composition root starts `/usr/local/bin/fake-harness`, which
already ships in the composed image, before it serves the existing attach
routes. The existing `/api/session` route reads the owner's current observation,
and the existing one-shot fetch renders it at the branch proxy attach URL. No
provider account, auxiliary service, new public endpoint, or host process scan
is required.

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

After the automated gates, from the repository root, start only the branch's
isolated test environment:

```sh
make test-env
project=$(printf '%s' "$(git branch --show-current)" | tr 'A-Z' 'a-z' | tr -c 'a-z0-9_-' '-')
base="http://127.0.0.1:${PROXY_PORT:-7090}/${project}/sandbox"
until [ "$(curl -fsS "$base/api/session" | python3 -c \
  'import json,sys; print(json.load(sys.stdin).get("childStatus", ""))')" = Running ]; do
  sleep 0.2
done
printf 'session: '; curl -fsS "$base/api/session"; printf '\npage: %s/\n' "$base"
```

The default fake is launched with its standard input kept open and without
`-early-exit`, a transcript, or failure injection. After its startup records it
blocks reading that input, so the owned process remains alive long enough for
the endpoint and page observation without sending a harness command.

Open the printed trailing-slash page URL (or reload it after the loop). Observe
the independently labeled **Child status: Running**, unchanged
`container-standard` posture and `network-egress` disclosure, and a usable
local composer: enter a non-empty draft and confirm it appears only in **Local
messages**. Reloading is required if the page was loaded before the endpoint
reported `Running`, because the preserved fetch is intentionally one-shot.
Then tear down only this branch environment:

```sh
make test-env-down
```

Record the endpoint and DOM observations separately. This proves that the
sandbox owns a live fake process and transports the observation through the
deployed page. It does not prove fake readiness, input delivery, output
handling, provider compatibility, exit interpretation, or the excluded manual
`Exited` path. Never use `make main-down` or replace the operator's default
`code-winch` deployment.

## Verification

- Add focused process-owner tests using controllable helper subprocesses for:
  successful launch reports `Running`; an executable/start failure never
  reports `Running`; and an owned child that exits is eventually reported as
  `Exited`. Exercise concurrent observation and bounded cleanup so the tests
  detect status races, double waits, and an orphaned child.
- Extend `go test ./test/contract/attach -run
  'TestEndpointContracts|TestRejectsMethodsAndUnknownAPI'` to cover exact
  `Down`, `Running`, and `Exited` wire values while preserving posture,
  method rejection, and unknown-API behavior.
- Run `go test ./cmd/winch-sandbox` for configuration, successful/failed
  launch, exit observation, and cleanup. Include an integration-level
  composition test or equivalent that proves the server reads the live status
  source rather than a copied startup value.
- Run `npm test -- --run src/attach/App.test.tsx` from `web/` for exact
  `Down`/`Running`/`Exited` presentation and `Unknown` on pending, request
  failure, invalid JSON, missing status, and unsupported values. Keep posture
  and composer assertions green and prohibit readiness/input/exit-result text.
- Run `npm run format:check`, `npm run lint`, `npm run typecheck`, `npm test`,
  and `npm run build` from `web/`; run `python3 -m unittest discover -s test`
  and `(cd workplan-control-panel && python3 -m unittest discover -s tests)`.
- Run the required host `make check`, then `make test-cycle` so the unchanged
  standing e2e scenario exercises the composed image with the owned child in
  the runtime path. Report unavailable Docker/toolchains as limitations.
- Perform the per-branch demonstration above and record the endpoint, rendered
  page, posture, composer, and teardown observations separately.

The fake-child tests establish deterministic process ownership and state
transitions, not compatibility with a real vendor CLI. Race-enabled Go testing
is encouraged when available, but does not replace `make check`.

## Acceptance criteria

- [ ] The maintained sandbox starts the shipped fake before serving, retains
      ownership of that exact process, and cleans it up without an orphan or a
      second waiter; a failed start cannot produce `Running`.
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
      failed launch, observed exit, cleanup, wire values, and React validation.
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
