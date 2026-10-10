# Phase 0 design — A sandbox you can talk to

**Review state: retrospective, with unresolved boundary findings.** This phase
was opened before the phase-design step existed. This artifact records the
baseline and proposed boundary corrections; it does not claim the existing
briefs were derived by that step. The [plan index](../README.md) and
[`tasks.json`](../tasks.json) remain the executable graph. Resolving the
findings below requires a workplan change updating the affected briefs and
graph together, before implementing their revised scope.

## Outcome and constraints

Roadmap Stage 0's hand check is the phase outcome: start one sandbox, type a
line, receive a fake-harness reply, and retain the exchange after a reload.
There is one session, one raw output view, and no control plane or run model.
Sources: [`docs/roadmap.md`](../../roadmap.md) §3 Stage 0 and §4;
[`docs/contracts.md`](../../contracts.md) §8; ADR-0005 and ADR-0006.

Ordinal ordering and the sandbox's posture are product constraints. The
existing fake's JSON dialect, event names, launch flag, and signal behavior are
provisional infrastructure, under contracts §9. End-to-end observability and a
minimal operator command apply to each chosen increment; they do not decide
which concerns must share an increment.

## Starting point

The tracker records P0-001 and P0-004 as completed. Source inspection confirms
the following baseline; runtime gates were not executed for this design review.
`docs/state.md` predates these implementations, so its missing-system claims
are not the current baseline.

| Existing behavior | Evidence | Missing change |
|---|---|---|
| Sandbox configuration and HTTP server with SIGINT/SIGTERM shutdown | `cmd/winch-sandbox/config.go:18-30`, `cmd/winch-sandbox/main.go:19-44` | Start, pump, and reap a child harness; no runner is constructed here |
| Attach page, health and posture endpoints | `internal/adapters/transport/attach/server.go:28-40`, `web/src/attach/App.tsx:8-45` | Deliver harness output to the page |
| Fake binary shipped in the image | `deployments/Dockerfile:14-20` | Launch it at runtime and pass scenario controls into the container |
| Local composer feedback | `web/src/attach/Composer.tsx:22-33`, `web/src/attach/App.tsx:41-44` | Submit to the server, persist, then deliver to the harness |
| Operator status command | `cmd/winch/status.go:21-49` | Minimal commands for each new capability |

P0-002's first scope item is **not already implemented by P0-001**. HTTP-server
shutdown and shipping an executable are not child supervision. Reuse the
composition root, configuration, image, and page; describe their missing wiring
as the delta instead of rebuilding them.

## Concerns

| Concern | Why it belongs in the phase | Existing plan coverage / boundary finding |
|---|---|---|
| Start the harness and show live output | First observable output loop | P0-002; requires child ownership, output decoding, minimal record delivery, page and CLI |
| Safe child cleanup | Owning a process must not leak it | Keep minimum cleanup with process ownership; this does not itself require a stop API or outcome taxonomy |
| Harness exit reporting and outcome mapping | §8 requires a terminal record at phase completion | Currently bundled into P0-002; separate candidate increment below |
| Durable record history and replay | Reload retains the exchange | P0-003; separate from observing live output |
| Local input, accepted submission, harness delivery | Each adds an observable property on the input path | P0-004, P0-005, P0-006 respectively |
| Malformed output and stalled readers | Preserve output under specified failures | P0-007; baseline bounds and security requirements still apply to live delivery |
| Intentional operator stop | A new control, beyond passive exit observation | No Stage 0 stop API requirement; control-plane run stop belongs to roadmap Stage 4 |

## Candidate increments and boundary decisions

The live-output increment changes a static page into one that shows a
deterministic transcript produced by a child. Its prerequisite is the existing
sandbox. It owns the child safely and provides `winch stream` plus the page.
Process launching, pumping and minimal delivery must work together to make
that observation; persistence and detailed exit outcomes are not needed to
observe a running transcript.

A distinct exit-reporting increment changes a stream that ends into a stream
that reports the child's native exit and mapped outcome. It builds on live
delivery and is observable through the same page and CLI. It does not introduce
an operator stop capability. Keep minimum child cleanup in the first increment.

**Boundary finding:** P0-002 currently includes both increments. Its assertion
that an exit without a terminal record freezes the page establishes a useful
exit guarantee, but not that it must arrive with first live output. Split that
guarantee from live output when revising this phase. Allocate any new task from
the plan-wide counter and update scope, non-goals, deferrals, dependency edges,
and collision reports together; candidate names here are not deferral owners.
Keep §8's terminal-record requirement owned within Stage 0.

History and the three input increments already have independently observable
boundaries. Preserve their existing completed work and IDs. Additional fake
controls arrive with the concern they exercise; I3 does not make failure
injection, malformed-line handling, and live output one task.

## Verification mechanics and phase exit

The current P0-002 demonstration needs these corrections during the boundary
revision and refinement:

- Select a deterministic transcript and define when the stream attaches.
  Startup output is not evidence that a late subscriber receives it; durable
  replay belongs to the history concern.
- Forward documented harness controls through `deployments/compose.yml`.
  Setting a host variable does not establish container configuration.
- Observe child exit while the attach server remains available. Signalling the
  sandbox container shuts down the server and cannot establish that result.
  The current fake's signal handler exits with status zero
  (`cmd/fake-harness/main.go:68-75`); its behavior is changeable infrastructure,
  not proof that the runner can infer an intentional stop from the wait status.

These are source-verified mechanics and predictions for new behavior, not
executed demonstrations. Each implementation must record the actual commands
and observations. Phase exit remains the Stage 0 hand check against HEAD, with
the standing suite green and all included concerns mapped to completed tasks.
