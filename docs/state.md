# System state

What exists at HEAD, what went wrong getting here, and what has no working code
behind it. This replaced the plan that was stripped from `docs/workplan/`, and
[`docs/roadmap.md`](roadmap.md) is the order the gaps below get closed in.

A plan is now in flight again: [`docs/workplan/`](workplan/README.md) decomposes
roadmap Stage 0 into phase 0. This file was the input to that derivation and is
not updated by it — the gaps below are gaps until code closes them, whatever a
brief claims it will do. Where this file and a brief disagree about what exists,
check the repository.

The first runnable slice now exists. `docker compose -f deployments/compose.yml
up --build` builds a non-root sandbox image, publishes its attach page on host
loopback, and exposes health and effective-posture reads. The image carries the
fake harness but does not start it; session records, input, and the harness
process lifecycle remain unimplemented.

## What was done

**The design baseline.** `docs/architecture.md`, `docs/code-structure.md`,
`docs/contracts.md`, `docs/security.md`, `docs/roadmap.md`, and six ADRs in
`docs/decisions/`. These describe a destination, and as of this writing almost
none of it is built. ADR-0005 and ADR-0006 are the two decisions that were made
against the current state rather than inherited.

**A controllable fake harness.** `cmd/fake-harness/` builds a deterministic
stand-in for a vendor CLI, speaking newline-delimited JSON. It is controllable in
the sense I3 requires — `-transcript` plays scripted commands, `-delay` injects
latency, `-force-failure` exits nonzero, `-malformed-line` emits an invalid
record, `-early-exit` exits after startup — and it emits a final observation on
SIGTERM so a stop is distinguishable from a crash
(`cmd/fake-harness/main.go:51-77`). `cmd/fake-harness/main_test.go` covers it.

What it does not prove: it is not a coding agent. It has no tool calls, no
approvals, no usage reporting, no login, and no terminal semantics — it reads
JSON lines from a pipe. Its dialect is invented, so an event model built only
against it would be modelling the fixture.

**The standalone sandbox slice.** `cmd/winch-sandbox` serves `GET /healthz`,
`GET /api/session`, and the built React attach page. `deployments/Dockerfile`
builds the page and all three binaries into a non-root image, while
`deployments/compose.yml` publishes it only on `127.0.0.1:8080`.
`cmd/winch status` is the maintained scriptable path to the same health and
posture contract. Contract tests live in `test/contract/attach/`, and
`test/e2e/scenario_sandbox_starts_test.go` drives the composed image through its
real entrypoint.

**The repaired quality floor.** `make check` covers formatting, vet, repository-
wide lint, Go tests (including compilation of the e2e package), and all three
binaries. The web workflow runs formatting, lint, type checking, its component
test, and the production build. `make test-cycle` supplies a Docker Go toolchain,
and `make e2e` runs the standing composed-image scenario.

**The workplan control panel.** `workplan-control-panel/` is a working HTTP API
and UI that reads `docs/workplan/tasks.json`, prepares task branches, and
dispatches refine/implement/audit stages to cloud agents. 44 tests pass:

    $ cd workplan-control-panel && python3 -m unittest discover -s tests
    → Ran 44 tests ... OK

**The planning skills.** `skills/shared/workplan-model.md` defines the four
properties, seven invariants, four task shapes, and the frozen tracker schema;
`skills/workplan/SKILL.md` and `skills/task/SKILL.md` apply it at plan and task
level. These were rewritten after the last plan failed and have not yet been
used to derive a plan.

**The task-completion scripts.** `scripts/list-available-tasks.sh` computes
availability from the tracker, and `scripts/stamp_task_completion.py` stamps
status on approval. `test/test_stamp_task_completion.py` covers the latter:

    $ python3 -m unittest discover -s test
    → Ran 9 tests ... OK

## What went wrong

**The previous plan built downward and never produced a running system.** It
started from the control plane and the domain model, with the runner wired last,
so the only way to observe a harness was through everything else first. Nothing
was ever startable end to end. This is the failure the roadmap's ordering rules
exist to prevent, and it is why ADR-0001's in-process-first sequencing was
superseded rather than retried.

**The plan was stripped without writing this file.** The close procedure
requires `docs/state.md` and the emptying of `docs/workplan/` in one commit. The
plan was removed and this file was left as a bare heading, so between then and
now the repository asserted an authoritative account of itself that did not
exist. Three documents went on citing it — `AGENTS.md`, `README.md`, and
`docs/workplan/README.md` — and all three described an active Phase 0 with tasks
that no longer existed.

**The removed implementation and broken gates were repaired by the first Phase 0
slice.** The earlier tree lacked `internal/`, the sandbox and CLI composition
roots, `deployments/`, `test/contract/`, `test/e2e/`, and buildable web sources.
Those paths now exist for the standalone sandbox. Paths belonging to later
roadmap stages — the daemon, public OpenAPI, PostgreSQL, and their integration
tests — remain deliberately absent, and the Makefile no longer names them.

**The tracker did not validate and the dispatch query failed on it — since
repaired.** `docs/workplan/tasks.json` was left as `{}`, while
`tasks.schema.json` requires `schema_version`, `status_values`, and `tasks`
(`tasks.schema.json:7`), so the availability query the control panel and
contributors both depend on errored rather than reporting nothing available,
which is the one thing a surviving empty tracker was supposed to guarantee. It
was given a valid empty body before this plan was derived, and both now hold:

    $ ./scripts/list-available-tasks.sh
    → []          # exit 0

The lesson is the close procedure's, not the tracker's: the file survives a close
precisely so these two keep working, and a close that empties it to the wrong
shape has removed the guarantee it was preserving.

**The tracker schema was reported as capping phases at 5. It does not.** An
earlier revision of this file claimed `tasks.schema.json` constrained `id` to
`^P[0-5]-[0-9]{3}$`, `phase` to a maximum of 5, and `brief` to `^phase-[0-5]/`,
against `skills/shared/workplan-model.md:488-489`: "The phase number is
open-ended — a plan derives as many phases as its design set needs, and
`tasks.schema.json` must not cap them." The schema at HEAD is already
open-ended — `^P[0-9]+-[0-9]{3}$` (`:34`), `{"minimum": 0}` with no maximum
(`:36`), `^phase-[0-9]+/.+\.md$` (`:48`) — so no repair is needed and the seven
roadmap stages are not blocked. Recorded because the claim was acted on: a
reader trusting it would have changed a frozen file for no reason.

**The close procedure's own empty-tracker example is invalid.**
`skills/workplan/SKILL.md:221` gives the closed state as
`{"schema_version": 1, "tasks": []}`, which the schema rejects for missing
`status_values`. A close following it literally would produce a tracker that
fails validation — and therefore fails every pull request, since
`.github/workflows/task-status-gate.yml` reads it on each one. (An earlier
revision of this file cited `skills/shared/workplan-model.md:221`, which is a
different sentence; the defect is in `skills/workplan/SKILL.md`.)

**A code comment cites a fixture that was deleted.**
`cmd/fake-harness/main.go:29` says `streamPayload` matches
`schemas/events/v1/fixtures/raw-stream.json`. `schemas/` does not exist, so the
wire format the fixture pinned is now defined only by the struct beneath that
comment.

## What is not implemented

Most of the destination design remains absent. Stated as gaps rather than as
tasks — naming the work belongs to the plan derived from `docs/roadmap.md`.

**The sandbox does not yet own a harness process or records.** Its image,
composition root, posture endpoints, attach page, and status CLI exist. It does
not start the included fake harness, decode output, accept input, or keep an
ordered session record. Those are the remaining Stage 0 increments.

**There is no product browser application.** The sandbox-local attach page is
buildable and served, but the multi-run product surface belongs to the control
plane stages.

**No harness adapter exists.** The ports in `docs/code-structure.md` §3 —
`HarnessDriver`, `HarnessCodec` — have no implementations, and no vendor CLI is
integrated. The fake harness is a binary, not an adapter: nothing decodes its
output into records.

**No sandbox driver exists.** `local` and `docker` are both described and
neither is written. No capability descriptor, no prepare/start/attach/inspect/
stop/cleanup lifecycle, no shared contract suite.

**There is no control plane.** No daemon, no HTTP API, no OpenAPI document, no
WebSocket streaming, no run aggregate, no run state machine, no supervisor, no
leases, no event store, no outbox, no migrations, no PostgreSQL schema. Every
aggregate in `docs/architecture.md` §6 is unbuilt.

**There are no renderers.** Terminal, conversation, activity, and changes are
described in `docs/contracts.md` §6 and none exists.

**No workflow machinery exists.** No definitions, coordinator, `WorkflowRuntime`
implementation, or step types.

**Only the Stage 0 host-local controls are enforced.** Compose pins publication
to loopback, the image runs as a non-root user, the static handler refuses path
and symlink escape, and the surface states that network egress is unenforced.
Authentication, authorization, egress policy, credential references, secret
redaction, sensitivity classification, retention, and the later launch blockers
remain absent.

**The standing scenario covers only sandbox startup.** It builds the composed
image, checks both endpoints and the status CLI, verifies the non-root UID, and
proves loopback-only publication. Later Stage 0 capabilities add scenarios to
this suite.

**The operator CLI currently has only `winch status`.** Commands for streaming,
records, and input do not exist until the capabilities behind them land.
