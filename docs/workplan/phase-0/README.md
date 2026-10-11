# Phase 0 — A sandbox you can talk to

This phase specifies the increment partition used to derive its tasks. The
partition replaces the earlier seven-brief decomposition; the
[derivation record](#task-derivation-record) identifies completed work and
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

| Concern | Destination and selected boundary |
|---|---|
| Runnable sandbox and local interaction | Reuse `sandbox-page` and `input-local`; no repeated foundational task |
| Child ownership and safe cleanup | `harness-owned`; status makes ownership observable before output delivery |
| Raw live output | `output-live`; new output during attachment, without replay or semantic decoding |
| Retained output and input history | `history` and `input-recorded`; persistence is a new observable property |
| Harness input delivery | `input-delivered`; acceptance alone does not imply delivery |
| Slow readers and oversized output | `output-bounded`; hardening has its own observation |
| Exit records and outcome interpretation | Roadmap Stage 1, with the first real harness; excluded from Phase 0 |
| Structured decoding and no-gap reconnect | Roadmap Stages 1 and 2 respectively; invented fake JSON does not pull them forward |
| Intentional run stop, retries, escalation, orphan recovery | Roadmap Stages 4 and 5; minimum child cleanup remains in `harness-owned` |

## Increments

| Increment | Observable change and included responsibility | Prerequisites | Boundary: excludes |
|---|---|---|---|
| `sandbox-page` | A startable sandbox serves its page and posture; operator status answers | Existing deployment inputs | Harness launch and all I/O |
| `input-local` | Typing produces local page feedback | `sandbox-page` | Server submission and delivery |
| `harness-owned` | Sandbox status and page show that an actual child is running; launch, reap, and clean it up | `sandbox-page` | Live output, record envelope, detailed exit reporting, stop controls |
| `output-live` | Page and operator stream show raw output produced while attached | `harness-owned` | Startup replay, persistence, structured events, JSON-line parsing |
| `history` | Reload and operator history show previously collected output in ordinal order | `output-live` | Seamless reconnect, multiple views, ended-session history |
| `input-recorded` | Page and operator submission receive acceptance/refusal; accepted text is recorded and survives reload | `input-local`, `history` | Writing to the harness, typed modes beyond text |
| `input-delivered` | Accepted text reaches the child and its reply arrives through the existing output path; unavailable input is refused | `input-recorded`, `output-live` | Lifecycle records, recovery, invented semantic reply formats |
| `output-bounded` | A stalled reader or oversized output cannot stop healthy output or exhaust memory; failure is visible | `history` | Native structured-payload parsing, security features assigned to Stage 5 |

These boundaries are decisions. Ownership and live output are separate
increments because the existing status path can prove ownership by itself.
Live output and persistence are separate because seeing new bytes and retaining
them across reloads are different observations. Acceptance and delivery are
separate because recording a submission and seeing a harness reply are
different observations. Keep minimum operator reachability and safe resource
bounds with each increment; additional guarantees do not become prerequisite
features just to make its demonstration easier.

The fake profile must provide output observable while the live reader is
attached, and a reply when input delivery is introduced. Its implementation may
change to meet those behaviors. This phase does not select a transcript,
startup timing trick, event dialect, stream transport, or terminal outcome map.
Task derivation selects the smallest implementation and credible observation
for the decided behavior.

## Task derivation record

| Increment | Existing coverage / next derivation action |
|---|---|
| `sandbox-page` | Reuse completed P0-001 |
| `input-local` | Reuse completed P0-004 |
| `harness-owned` | Not yet derived separately; rederive pending P0-002 around ownership/status |
| `output-live` | Not yet derived separately; derive the output increment after ownership, allocating a new ID |
| `history` | Reconcile pending P0-003 with the output task and this durability boundary |
| `input-recorded` | Reconcile pending P0-005 with its prerequisite-produced history contract |
| `input-delivered` | Reconcile pending P0-006 with direct child availability and raw reply observation |
| `output-bounded` | Reconcile pending P0-007 with raw-byte delivery; remove invented-JSON malformed-record requirements |

Derive `harness-owned` next, then continue increment by increment. Each may
split further into smaller useful tasks; combining increments requires an
explicit phase revision first. Preserve completed IDs and allocate new IDs
from the plan-wide counter. Reconcile affected briefs, dependency edges, index,
and collision report before assigning those tasks under this partition.
Legacy graph availability is not evidence that a brief fits the new design.

## Phase exit

A person starts the sandbox, sees the child running, observes fresh output,
submits text, sees its reply, and reloads to see the exchange retained in order.
Healthy readers continue when another stalls, and output memory stays bounded.
Exact demonstration commands and test fixtures belong in the derived tasks.
Exit-record and stop demonstrations are not part of this phase's acceptance.
