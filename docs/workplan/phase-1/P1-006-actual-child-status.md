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

Refinement confirmed P1-004's completed dependency at HEAD. In
`cmd/winch-sandbox/main.go:21-24` the composition root supplies
`attach.ChildDown`; `internal/adapters/transport/attach/server.go:9-19,31-32`
copies that session into the server and returns the copy on every request; and
`web/src/attach/App.tsx:10-23` performs one fetch and accepts only exact `Down`.
The matching Go and React tests are present at
`test/contract/attach/session_test.go:28-52` and
`web/src/attach/App.test.tsx:45-109`. The dependency therefore provides the
wire value, validation boundary, and presentation identity this task extends.

The sandbox still starts no child (`cmd/winch-sandbox/main.go:15-40`). The
image does ship `fake-harness` beside `winch-sandbox`
(`deployments/Dockerfile:14-24`), and the fake requires `-run-id`
(`cmd/fake-harness/main.go:51-76`) before it emits startup records and reads
standard input (`cmd/fake-harness/main.go:88-119`). A nil or immediately closed
input would make the fake observe EOF and exit, so the owned launch must keep
its input pipe open even though this task does not write commands. Refinement
ran the existing deployment demonstration entry command, `make test-env`; it
stopped at `docker: No such file or directory`, so no runtime behavior is
claimed at HEAD.

## Scope

- Add a focused, package-local process owner under `cmd/winch-sandbox`. It
  starts the shipped sibling `fake-harness` executable with a non-secret fixed
  run ID, retains the resulting `*os.Process`/`*exec.Cmd` ownership, and waits
  for that exact child. Resolving the sibling rather than assuming the current
  working directory keeps both the composed `/usr/local/bin` installation and
  the existing host `bin/` build usable. Keep the fake's standard input open so
  its default mode remains alive; discard its output for this increment rather
  than logging, decoding, buffering, or publishing it.
- Publish `Running` only after `Start` succeeds, and publish `Exited` only
  after the single owning `Wait` returns. Synchronize reads and transitions so
  concurrent session requests and exit observation are race-free. Treat a
  start failure as sandbox startup failure before HTTP serving begins; it must
  never expose `Running`. Tests may inject a helper executable/arguments into
  the package-local owner, but production always selects the shipped sibling.
- Connect child teardown to the composition root's existing signal-driven,
  five-second shutdown window. Close the held input, wait for the owned child,
  and forcibly terminate it if graceful cleanup does not finish before that
  same bound; do not use process-table scans or PID-only signaling. Complete
  child cleanup as well as `http.Server.Shutdown` before `main` returns.
- Change the attach server's construction contract from a copied `Session` to
  the posture plus a current-child-status source owned by the composition root.
  Build a fresh `Session` for every `/api/session` request, so a request before
  and after `Wait` can observe different values without making posture mutable.
  Keep `Down` available for the no-child contract fixture and accept only the
  source supplied at construction; do not infer state from output, browser
  timing, `/healthz`, or a process-table lookup.
- Preserve the one-shot browser fetch: a load or reload observes one
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

Keep process ownership package-local to `cmd/winch-sandbox`; this first owned
process does not justify a generic harness adapter or runner domain model. The
attach transport consumes only the status source and must not know the fake's
path, flags, pipes, or process type.

## Required constraints

The composition root is the authority for this child because it launches,
retains, waits for, and cleans up the process. There must be exactly one `Wait`
owner. Status publication must be safe under concurrent session requests and
child exit. A failed `Start` cannot transiently or permanently produce
`Running`; only successful start followed by an unobserved exit can. `Exited`
conveys no native exit interpretation or result, and neither state conveys
readiness or a writable input channel.

Use the shipped fake as a malleable test profile, not as a product protocol.
Logs, errors, and tests must not expose child output or submitted content.

## Non-goals

The manual `Exited` demonstration, input submission or child commands, child
output logging/transport/display/retention, exit code/signal presentation,
restart or retry policy, readiness, provider semantics, PTY support,
persistence, polling, configurable harness selection, and broad page
restructuring are excluded. The held input pipe is lifecycle ownership, not an
input-delivery contract. These exclusions are not required to complete this
truthful owned-process status slice.

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

`deployments/Dockerfile` already installs both executables in the same directory
and is not in the write set. No configuration, fake-harness protocol, compose
topology, CLI, output, input, or persistence file is forecast.

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

After the automated gates, use the branch-isolated Docker environment described
in `docs/testing.md` §Demonstrations:

```sh
make test-env
curl -fsS http://127.0.0.1:7090/api/session
# Open http://127.0.0.1:7090, then reload if it was already open.
make test-env-down
```

The endpoint response must contain exact `"childStatus":"Running"`,
`"profile":"container-standard"`, and `"network-egress"`. On the page,
observe **Child status: Running**, the unchanged `container-standard` posture
and `network-egress` disclosure, and submit a local composer message to confirm
the composer remains usable and makes no server-delivery claim. The default
fake, launched without `-early-exit`, a transcript, or failure injection, reads
its owner-held standard input after its startup records, so it stays alive for
the observation without any harness command. Always run `make test-env-down`,
including after a failed observation, and record that the branch's containers
were removed.

This proves that the sandbox owns a live fake process and transports the
observation through the deployed page; it does not prove fake readiness, input
delivery, output handling, provider compatibility, or exit interpretation.

## Verification

The gates in `docs/testing.md` apply; say which option of each you ran. Checks
specific to this task:

- `go test -race ./cmd/winch-sandbox` covers successful launch reporting
  `Running`, a start failure never reporting `Running`, observed exit reporting
  `Exited`, concurrent status reads during exit, and bounded cleanup of a child
  that does not stop gracefully.
- `go test ./test/contract/attach` covers exact `Down`, `Running`, and `Exited`
  wire values, proves successive requests read the live source rather than a
  copied startup value, and preserves posture, method rejection, and
  unknown-API behavior.
- `npm test -- --run src/attach/App.test.tsx` from `web/` covers exact
  `Down`/`Running`/`Exited` presentation and `Unknown` on pending, request
  failure, invalid JSON, missing status, and unsupported values, with posture
  and composer assertions kept green.
- The per-branch demonstration above, with endpoint, page, posture, composer,
  and teardown observations recorded separately.

The fake-child tests establish deterministic process ownership and state
transitions, not compatibility with a real vendor CLI.

## Acceptance criteria

- [ ] The maintained composed and host-built sandbox launches the shipped
      sibling fake with an open input pipe, owns its sole wait path, and cleans
      it up within the existing shutdown bound; a failed start aborts sandbox
      startup and cannot produce `Running`.
- [ ] A concurrency-safe observation reports `Running` only after successful
      start and before observed exit, then `Exited` after `Wait` returns,
      without host process scans, browser inference, output inference, or
      readiness/input claims.
- [ ] Each `GET /api/session` request reads the owner's current observation;
      the endpoint, `docs/contracts.md`, and focused server tests agree on exact
      `Down | Running | Exited` meanings and preserve posture and route
      behavior.
- [ ] The existing accessible child-status region renders each validated value
      and retains `Unknown` for pending, failed, malformed, missing, or unknown
      data while preserving posture and the local composer.
- [ ] The default fake configuration keeps `Running` observable without input
      commands or output handling, and automated tests cover successful launch,
      failed launch, concurrent observation through exit, bounded cleanup, live
      per-request wire values, and React validation.
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
