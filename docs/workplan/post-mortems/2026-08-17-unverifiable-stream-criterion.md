# 2026-08-17 — P1-050 does too much, and its witness was split downstream

**Found:** reviewing P1-050 before its pull request merged.
**Symptom:** the branch satisfies "published events reach a subscribed
WebSocket in sequence order" — verified by hand with a throwaway WebSocket
client written for the review and then discarded. No command in the repository
reproduces that check, and no test asserts it. The criterion is true, and its
truth is unreproducible.

## What the plan did

P1-050's acceptance criteria include:

> - [ ] Published events reach a subscribed WebSocket in sequence order.

Its Demonstration block contains four `curl` invocations covering create, start,
events, input, and an unknown profile. None of them subscribes to
`/runs/{id}/events/stream`. Its Verification section lists service tests, a
PostgreSQL integration test, HTTP problem-code tests, and a delivery test. None
of them opens a socket.

The two surfaces that can produce this evidence are both downstream:

| Task | Provides | Edge |
|---|---|---|
| P1-051 | `winch run watch` — the hands-on client | depends on P1-050, "semantic: the run API must answer before a client can drive it" |
| P1-053 | `test/e2e` driving "the real HTTP and WebSocket entrypoints" | depends on P1-050, semantic |

So P1-050 cannot be judged complete without evidence that only tasks depending
on P1-050 can generate.

## Root cause

### P1-050 is not one task

Its scope carries, at minimum: five run use cases; a driver registry; profile
resolution with a migration; the HTTP backend and its five problem-code
mappings; an outbox publisher routing two topics; composition-root wiring of
runner, supervisor, outbox worker, and API handler; a startup reconciliation
sweep that discharges P1-020; three metrics; an admission seam for P2-059; and
a mechanical split of the OpenAPI document with a byte-identical output
requirement.

The plan requires every task to fit one of four shapes. This one is a seam, a
capability, a refactor, a telemetry foundation, and the completion of another
task's work, filed under a single header that reads `**Shape:** seam`. Its
owned surfaces are the composition root, the whole API document, a migration,
the application services, and the transport backend — which is to say P1-050
owns the collision set of Phase 1. Everything else waits for it, which is
exactly why it became the bottleneck that the split was reaching for relief
from.

### The split relieved that pressure in the wrong direction

Shedding the CLI made P1-050 smaller by removing the one part of it that could
show the rest of it working. The remaining scope kept every claim and lost its
only witness.

The valid direction is the reverse: **P1-051 before P1-050.** A client that can
drive create, start, watch, input, and stop is not a feature that follows the
API — it is the instrument the API is demonstrated with, and instruments come
first. Concretely, the order that works:

1. a minimal run API seam serving the all-fake profile end to end, small enough
   to be one shape;
2. the operator CLI against it — from here on, every later task has a standing
   way to be observed by hand;
3. then the substrate and capability work as separate tasks, each witnessed by
   that CLI the moment it lands: Postgres-backed use cases, runner and
   supervisor wiring, outbox delivery, the reconciliation sweep, the metrics,
   and the OpenAPI split as its own mechanical change.

Running them concurrently would not have helped either. A witness that arrives
alongside the thing it witnesses does not exist while that thing is being built
and reviewed, which is when the evidence is needed.

### The edge that permitted it is one the plan already rejects

P1-051's dependency reads "the run API must answer before a client can drive
it". That is the "it is nicer to develop against the real thing" reason, which
the plan names as the single largest source of lost width and answers with I3:
the fake profile is a supported way to run the product, so a client can be
built and demonstrated against it before any real substrate exists. Written as
a semantic edge, the fallacy passed review.

### The reviewability argument is a symptom, not a justification

"P1-050 is large, so moving the CLI out kept the pull requests reviewable" is
the defense this record originally offered, and it is the defect restated. Task
size is a statement about decomposition; when a task is too large to review,
the plan rule is to split the *task* along its shapes, and the pull-request
rule is to split the *pull requests*. Using review pressure to decide which
scope leaves a brief is how the witness ended up downstream: it was the most
detachable piece, not the piece that should have moved.

Unlike the previous record in this directory, it is not true here that no brief
is wrong on its face. P1-050 is wrong on its face — one shape header over five
shapes' work — and P1-051's dependency edge is wrong on its face. Both were
visible without running anything.

## Why no test caught it

P1-018 tests the WebSocket thoroughly and could not have caught this. Its tests
dial a real socket against an `httptest` server, but publish into the
`EventStream` directly from the test body and read runs from a stub backend.
They prove the stream handler: resume, gap repair, slow-subscriber disconnect,
origin rejection. What they cannot prove is that a real event, committed by a
real store and delivered by a real outbox worker, arrives at a subscriber —
because in those tests no store, outbox, or publisher exists.

That is the correct scope for P1-018. The end-to-end path is exactly what
P1-050 created and exactly what its criterion claims.

## Remediation

Interim, and deliberately ugly: `tmp/` holds a WebSocket client and one script
per run capability, in the shape `winch run <command>` will take. P1-051 is
amended to delete the directory, with an acceptance criterion to that effect,
so the debt has an owner rather than a hope. The directory's README states that
nothing may depend on it.

This violates the plan's own rules — an unowned surface duplicating an owned
one — and it exists because the plan left no legal way to demonstrate a
criterion the plan requires. Recorded as a violation rather than dressed up as
a tool.

P1-050's brief gains the stream in both places it was missing: a Demonstration
line a person runs, and a Verification line for the automated check.

None of that addresses the root cause. **Re-deriving P1-050 into separate tasks
along the shapes above, and bringing the CLI in front of them, is the actual
remediation**, and it is a gate decision rather than something a review of one
pull request should perform: it changes the Phase 1 graph, its critical path,
and its width. Until that happens the plan still says one task owns the
composition root, the API document, a migration, and the application services
at once.

## Prevention

- A task's witness precedes the task. If the only way to observe a criterion is
  a surface scheduled later, the schedule is wrong, not the criterion.
- Hands-on surfaces are instruments, not features. I5 puts the operator CLI in
  the first phase for this reason; scheduling an API client after the API
  inverts it. Note also that `winch dev run` existing did not satisfy I5 — it
  drives the runner in-process and reaches none of P1-050's behavior. A CLI
  existing is not the new behavior being reachable from it.
- Task size is a decomposition signal. A brief whose scope spans several shapes
  is not a big task; it is several tasks sharing a header, and the response is
  to split along the shapes rather than to detach whatever is easiest to lift
  out.
- Review effort belongs to pull requests. It is never a reason to move scope
  between briefs, in either direction.
- When splitting, list the criteria that stay and the evidence that leaves. If
  evidence for a retained criterion moves, either the criterion moves with it or
  the evidence stays.
- Compare acceptance criteria against the Demonstration block, line by line,
  before a brief is final. Demonstrations get written against the Objective
  sentence and then criteria accrete above them.
- A dependency edge justified by "the real thing must exist first" is I3's
  answer restated as a blocker. Check whether the fake profile dissolves it
  before accepting it as semantic.
