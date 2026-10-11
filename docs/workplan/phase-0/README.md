# Phase 0 — A sandbox you can talk to

This phase specifies the increment partition used to derive its tasks. The
partition replaces the earlier seven-brief decomposition; the
[task coverage](#task-coverage) identifies completed work and
pending legacy briefs to reconcile. The [tracker](../tasks.json) still records
those existing task IDs and statuses; it has not been rewritten by this design.

## Outcome and starting point

Start one sandbox, type a line, see the harness answer, and reload to see the
exchange retained. One session, one raw-output view, no control plane.
Sources: [`roadmap`](../../roadmap.md) Stage 0 and §4, [contracts](../../contracts.md)
§8–9, ADR-0005, and ADR-0006.

The source baseline already includes the sandbox server and configuration
(`cmd/winch-sandbox/main.go:19-44`, `config.go:18-30`), its page and posture
(`internal/adapters/transport/attach/server.go:28-40`), a shipped fake binary
(`deployments/Dockerfile:14-20`), local composer feedback
(`web/src/attach/App.tsx:41-44`), and operator status
(`cmd/winch/status.go:21-49`). The sandbox does not start a harness yet.
P0-001 and P0-004 are recorded completed; these are reusable capabilities, not
new scope. This baseline was inspected in source; runtime gates were not run
for this phase design. `docs/state.md` is the earlier derivation's snapshot.

## Concern boundaries

| Concern | Destination |
|---|---|
| Sandbox page and posture | Reuse `sandbox-page` |
| Local composer feedback | Reuse `input-local` |
| Raw live output | `output-live` |
| Harness launch | Required wiring for `output-live` |
| Child cleanup with sandbox shutdown | Required wiring for `output-live` |
| Retained output history | `history` |
| Submission acceptance/refusal | `input-recorded` |
| Retained submission history | `input-recorded` |
| Harness input delivery | `input-delivered` |
| Slow-reader isolation | `reader-isolation` |
| Output-size limits | `output-limits` |
| Native exit reporting | Roadmap Stage 1 |
| Structured harness decoding | Roadmap Stage 1 |
| Structured views | Roadmap Stage 2 |
| Seamless reconnect | Roadmap Stage 2 |
| Intentional run stop | Roadmap Stage 4 |
| Retries | Roadmap Stage 4 |
| Stop escalation | Roadmap Stage 5 |
| Orphan recovery | Roadmap Stage 5 |

## Ordered increments

| Order | Increment | Observable change | Includes | Excludes | Prerequisites |
|---|---|---|---|---|---|
| 1 | `sandbox-page` | A startable sandbox serves its page and posture; operator status answers | Sandbox entrypoint and deployment | Harness I/O | Existing deployment inputs |
| 2 | `input-local` | Typing produces local page feedback | Local composer | Server submission | `sandbox-page` |
| 3 | `output-live` | Page and operator stream show newly produced raw output | Fake launch and output delivery; minimum shutdown cleanup and memory bounds | Startup replay, history, structured decoding, exit reporting | `sandbox-page` |
| 4 | `history` | Reload and operator history show earlier output in ordinal order | Retained output history | Seamless reconnect, multiple views, ended-session history | `output-live` |
| 5 | `input-recorded` | A submission is accepted or refused; accepted text survives reload | Server acceptance and retained input | Harness delivery, input modes beyond text | `input-local`, `history` |
| 6 | `input-delivered` | Submitted text reaches the child and its reply arrives in the output view | Input delivery and refusal when unavailable | Lifecycle records, recovery, semantic reply formats | `input-recorded`, `output-live` |
| 7 | `reader-isolation` | A stalled reader cannot stop healthy readers | Slow-reader disconnection and visible failure | Additional output-size guarantees | `output-live` |
| 8 | `output-limits` | Oversized output is visibly limited without exhausting memory | Output-size enforcement | Structured-payload parsing, Stage 5 security | `output-live` |

`input-local` and `output-live` are independent after `sandbox-page`.
`reader-isolation` and `output-limits` can be derived after live output without
waiting for history or input.

## Task coverage

| Increment | Coverage |
|---|---|
| `sandbox-page` | Reuse completed P0-001 |
| `input-local` | Reuse completed P0-004 |
| `output-live` | Next: rederive pending P0-002 for live raw output |
| `history` | Pending P0-003 needs reconciliation |
| `input-recorded` | Pending P0-005 needs reconciliation |
| `input-delivered` | Pending P0-006 needs reconciliation |
| `reader-isolation` | Pending P0-007 covers a combined legacy hardening scope; needs rederivation |
| `output-limits` | Not yet derived separately |

The tracker and index still describe the legacy graph. Pending briefs require
rederivation before assignment under this partition.

## Phase exit

A person starts the sandbox, sees fresh output, submits text, sees its reply,
and reloads to see the exchange retained in order. A stalled reader cannot stop
healthy readers, and oversized output is limited without exhausting memory.
