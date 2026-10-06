# System state

What exists at HEAD, what went wrong getting here, and what has no working code
behind it. This replaced the plan that was stripped from `docs/workplan/`, and
[`docs/roadmap.md`](roadmap.md) is the order the gaps below get closed in.

A plan is now in flight again: [`docs/workplan/`](workplan/README.md) decomposes
roadmap Stage 0 into phase 0. This file was the input to that derivation and is
not updated by it — the gaps below are gaps until code closes them, whatever a
brief claims it will do. Where this file and a brief disagree about what exists,
check the repository.

The product does not run. There is no daemon, no sandbox, no browser
application, and no way to start a harness through anything. What survives is
the design set, one test fixture, and the tooling that dispatches work.

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

**The implementation was removed with the plan.** `internal/`, `cmd/winchd`,
`cmd/winch`, `pkg/`, `api/openapi/`, `schemas/`, `migrations/`, `deployments/`,
`test/contract/`, `test/e2e/`, and `web/src/` are all absent. `web/` retains only
`node_modules/`, `dist/`, and two `tsbuildinfo` files, and has no
`package.json`.

**Every quality gate names a path that is gone.** This is an I1 and I2 failure
and the first stage of the next plan has to repair it.

- `make check` runs `api-check`, which reads `api/openapi/code-winch.yaml`
  (`Makefile:30,39`) and copies
  `internal/adapters/transport/httpapi/types.gen.go` and `web/src/api/schema.ts`
  (`Makefile:44-48`). None exists.
- `make build` builds `./cmd/winchd` and `./cmd/winch` (`Makefile:89-90`).
  Neither exists.
- `make test` runs `GO_PACKAGES`, which includes `./internal/...`, `./pkg/...`,
  and `./test/...` (`Makefile:8`). Only `./cmd/...` and `./test/...` resolve, and
  `./test/` holds Python.
- `make test-integration` runs `./internal/adapters/postgres/...`
  (`Makefile:79`), and `make e2e` runs `./test/e2e/...` (`Makefile:84`).
- **The whole `[docker]` group is broken too**, which the list above originally
  missed. `COMPOSE ?= docker compose -f deployments/compose.yml` (`Makefile:5`)
  names a directory that does not exist, so `runner-image`, `test-env`,
  `runner-verify`, `runner-integration`, `test-env-down`, `test-cycle`, and
  `runner-shell` all fail before running anything. That matters more than the
  rest of the list: `Makefile:17-18` and `AGENTS.md:116-120` both present
  `make test-cycle` as the way in for a host with no Go toolchain, which is this
  host. Docker itself is present — 27.4.1 with compose v2.32.1 — so the compose
  file is the only thing missing.
- Also unlisted and also absent: `deployments/README.md` (`Makefile:18`),
  `api/openapi/oapi-codegen.yaml` (`Makefile:30`), `cd web && npm …`
  (`Makefile:31,101-102`), `./test/contract/openapi` and its `v1.yaml` baseline
  (`Makefile:35,39`), and `./cmd/winchd` in the `run` target (`Makefile:97-98`).
  `.dockerignore:4,6` ignore `runner/` and `tests/`, neither of which exists,
  while the real `test/` is not ignored.
- `.github/workflows/go.yml` runs `python3 -m unittest discover -s tests -v`
  against a directory named `tests`; the directory is `test`. Verified:

      $ python3 -m unittest discover -s tests
      → ImportError: Start directory is not importable: 'tests'

- `.github/workflows/web.yml` runs `npm install` and six scripts in `web/`,
  which has no `package.json`.

The Go gates were not verified by running them — there is no Go toolchain on this
host (`go: command not found`, exit 127) — but each is checkable by reading the
cited line against the tree. `make check` and `make build` both fail at the
missing toolchain before they can reach a missing path.

Not every target is broken, against what `README.md:42-47` claims: `format`,
`format-check`, `vet`, and `lint` (`Makefile:51-70`) name only `./...` and would
pass on a host with Go installed.

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

Effectively the whole design set. Stated as gaps rather than as tasks — naming
the work belongs to the plan derived from `docs/roadmap.md`.

**There is no sandbox.** No container image, no compose file, no process that
owns a harness, no runner, no session record store, and no attach surface. The
contract in `docs/contracts.md` §8 has no implementation, and ADR-0005 describes
a component that does not exist. This is the whole content of roadmap Stage 0.

**There is no browser surface of any kind.** Neither the attach surface nor the
product application. `web/` cannot be built.

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

**No security control is enforced.** No authentication, authorization, egress
policy, credential references, secret redaction, traversal defenses, sensitivity
classification, or retention. Every threat in `docs/security.md` §3 is
unmitigated, and every launch blocker in §11 is open. The register states what
is intended, not what holds.

**There is no standing scenario suite.** I4 asks for one end-to-end scenario
driving the system through its real entrypoint, re-run against each substrate as
it becomes real. There is no entrypoint to drive.

**There is no operator CLI.** I5 asks for a maintained hands-on path.
`cmd/winch` was removed; the roadmap assigns its job to the sandbox's attach
surface, which is also unbuilt.
