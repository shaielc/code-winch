# Implementation workplan

**Phase 0 is open.** It decomposes Stage 0 of [`docs/roadmap.md`](../roadmap.md) — "a
sandbox you can talk to" — into seven tasks. The tracker records `P0-001` and `P0-004`
as completed; `P0-002` is available. The
[phase design](phase-0/README.md) records the current baseline and the boundary
findings from reviewing this already-open phase.

    $ ./scripts/list-available-tasks.sh
    → P0-002

[`skills/shared/workplan-model.md`](../../skills/shared/workplan-model.md) is the
definition this plan is written against: the four properties, the seven invariants, the
four task shapes, dependency-edge reasons, write sets, contract surfaces, brief anatomy,
and the frozen tracker schema. [`docs/state.md`](../state.md) says what existed at HEAD
when this plan was derived.

New phases are designed in `phase-N/README.md` before their briefs are drafted.
That artifact records the starting point, concerns, candidate increments,
boundary decisions, verification mechanics, and phase exit. The phase design
justifies task boundaries; this index and `tasks.json` describe the executable
graph. Retrospective findings in Phase 0's design do not change that graph.

## Scope of this plan

**Stage 0 only.** `docs/roadmap.md` decomposes one stage at a time, and the stages after
this one depend on decisions the roadmap deliberately leaves open — which vendor CLI
Stage 1 integrates (D1), how `winchd` learns a sandbox exists (D6), whether the canonical
envelope wraps or replaces the sandbox's record shape (D7). Briefs for them would be
hypotheses. Stage 1 arrives as a new derivation, or in extend mode, when this phase
closes.

One consequence is worth stating because it bears on I6. A single-stage plan has no task
IDs in later phases, so nothing in this plan can name one as an owner. The convention:

- A brief's **Deferrals** table may name only an ID in this tracker, or a
  `docs/roadmap.md` §6 deferred decision with its trigger.
- Work the roadmap schedules for a later stage goes in **Non-goals**, with the stage
  cited. That is a stage boundary, not an unowned deferral — the roadmap is the design
  set's statement of ordering, and it owns what it schedules.

A brief that puts a later stage's work in Deferrals, or that says "later" with no stage
and no ID, is a defect.

## Phase 0 — A sandbox you can talk to

| ID | Title | Shape | Depends on |
|---|---|---|---|
| [P0-001](phase-0/P0-001-sandbox-serves-attach-page.md) | The sandbox image starts, serves its attach page, and states its posture | seam | — |
| [P0-002](phase-0/P0-002-runner-owns-fake-harness.md) | The runner owns the fake harness and the page shows its output live | seam | P0-001 |
| [P0-003](phase-0/P0-003-records-persist-and-survive-reload.md) | Session records persist, and a reload shows what the harness already said | seam | P0-002 |
| [P0-004](phase-0/P0-004-page-input-control-echoes.md) | The page has an input control that echoes what you type | capability | P0-001 |
| [P0-005](phase-0/P0-005-submission-accepted-and-recorded.md) | A submission is accepted or refused, recorded with an ordinal, and survives a reload | seam | P0-003, P0-004 |
| [P0-006](phase-0/P0-006-submission-reaches-the-harness.md) | A submission reaches the harness and the harness's reply comes back | capability | P0-002, P0-005 |
| [P0-007](phase-0/P0-007-malformed-output-and-stalled-readers.md) | Malformed output and a stalled reader cannot silence the harness | hardening | P0-003 |

The two chains after P0-001 are the roadmap's two paths (§4). **Output:** P0-002 reads the
harness and shows it, P0-003 persists it and replays it on reload. **Input:** P0-004 echoes
locally, P0-005 persists what you submitted, P0-006 delivers it to the harness. They meet
at the page and at the store, which is why P0-005 waits on P0-003.

### Dependency edges

Every edge names one reason in one clause. An edge that cannot is not an edge.

| Edge | Reason | Clause |
|---|---|---|
| P0-002 → P0-001 | compile | needs the composition root, config loader, and page shell it wires the runner into |
| P0-003 → P0-002 | contract | P0-002 defines the record shape and the ordinal namespace P0-003 persists |
| P0-004 → P0-001 | semantic | there is no page to put a control in until P0-001 serves one |
| P0-005 → P0-003 | contract | the record store and ordinal allocation a submission is appended to |
| P0-005 → P0-004 | semantic | the page submission it wires has no control to submit from until P0-004 |
| P0-006 → P0-005 | contract | the accepted input command and its typed payload |
| P0-006 → P0-002 | contract | harness process I/O, which it adds the stdin side of |
| P0-007 → P0-003 | contract | the record store its diagnostic records land in, and the ordinal a disconnected reader resumes from |

No `revision` edges. A revision edge in a freshly derived phase is a defect in the
derivation — it records something learned after work started, and is added in extend mode.

### Width

Critical path **5** — P0-001 → P0-002 → P0-003 → P0-005 → P0-006. Seven tasks. Average
width **1.4**.

Five pairs can be pending and available at the same moment, which is what the collision
tables below are computed over: `{002, 004}`, `{003, 004}`, `{004, 007}`, `{005, 007}`,
`{006, 007}`.

The single root is forced rather than chosen: I1 and I2 require one task that makes the
system start and deploy, and every other task's demonstration assumes it. After it, both
roadmap paths open at once. The ceiling is two because the paths meet at the store, so the
input path's persistence step cannot precede the output path's. Stage 0 is one vertical
slice through one process; width is a property of the stages that add substrates.

**Write-collision set** — a cost, not an edge. Both tasks proceed and the second rebases.

| Pair | Overlap |
|---|---|
| P0-002, P0-004 | `web/src/attach/App.tsx` |
| P0-003, P0-004 | `web/src/attach/App.tsx` |
| P0-004, P0-007 | none — disjoint write sets |
| P0-005, P0-007 | `internal/runner/record.go`, `internal/runner/session.go` |
| P0-006, P0-007 | `internal/runner/harness.go`, `internal/runner/session.go` |

Scenario files are one per scenario and CLI commands one per file, so `test/e2e/` and
`cmd/winch/` do not collide beyond the shared compose helper P0-001 creates.

**Contract-collision set: empty.** P0-004 declares no server surface at all. P0-002
*defines* the record-kind namespace; P0-005 and P0-007 each only *add* a kind to it, which
is appending to a registry rather than redefining it. P0-006's input-frame encoding does
not touch P0-007's diagnostic kind or stream-disconnect semantics. Two concurrently
available tasks sharing a contract surface would be a defect to fix before the phase
opens, not a cost to report.

### Phase exit

Phase 0 is finished when all seven tasks are `completed`, verified against HEAD rather
than against the pull requests that closed them, and the Stage 0 hand check in
`docs/roadmap.md` §2 passes: start the sandbox, type a line, the fake harness echoes, and
the exchange survives a reload.

## How the invariants are owned

| Invariant | Owner in this phase |
|---|---|
| I1 — the system runs from the first task | P0-001, and every later task keeps it running |
| I2 — the system deploys from the first task | P0-001: `deployments/Dockerfile` and `deployments/compose.yml`. Deployment is never its own task |
| I3 — the fake is a controllable runtime profile | P0-002, which also documents what the fake does not prove |
| I4 — swaps prove parity against a standing scenario suite | P0-001 creates `test/e2e/`; every later task adds one scenario file and re-runs the rest |
| I5 — every operator-visible capability is reachable by hand | `cmd/winch`, built in P0-001 and gaining one command per capability: `status`, `stream`, `records`, `input` |
| I6 — nothing deferred without an owning task | the Deferrals convention above |
| I7 — the plan stays wide | the width report above |

**On I5 and the two surfaces.** ADR-0006 makes the sandbox attach surface the operator and
debug path, and that is what the page is. `cmd/winch` is its scriptable twin: a thin client
of the same `docs/contracts.md` §8 contract, not a second contract and not a second
surface. It exists because I5 says raw HTTP calls are fine as verification but are not the
maintained hands-on path, and because the host this plan was derived on has no Go toolchain
and no headless browser — a demonstration has to be a command somebody can run.

**On Docker.** Every demonstration in this phase goes through
`docker compose -f deployments/compose.yml`. That is not a convenience: the unit of
deployment is the unit of work (ADR-0005), and the host has no Go toolchain, so the
compose `toolchain` service is also how `make test-cycle` runs the Go gates.

## Working the plan

- **`tasks.json` is the graph.** The tables above restate it; drift between them is a
  defect. A task is available when its status is `pending` and every ID in `depends_on` is
  `completed`. `./scripts/list-available-tasks.sh` computes it.
- **One task, one pull request.** Include `Task: <ID>` in the body.
  `.github/workflows/task-status-gate.yml` requires every pull request to resolve to
  exactly one known task ID, and stamps `completed` on approval. Do not edit status fields
  by hand. Pull requests that implement no task need the `no-task` label.
- **A write set is a forecast, not a boundary.** A change that writes outside it lists
  those files so the write set and the collision table above can be updated. Scope is
  judged against Scope and Non-goals; meaning against Contract surfaces.
- **Plan failures get post-mortems.** When a defect's root cause is a seam between briefs
  rather than an implementation, it goes in [`post-mortems/`](post-mortems/). Ordinary
  bugs do not.
