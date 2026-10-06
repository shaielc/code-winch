# System state

What exists at HEAD, what went wrong getting here, and what has no working code
behind it. This replaces the plan that was stripped from `docs/workplan/`.
[`docs/roadmap.md`](roadmap.md) is the order the gaps below get closed in; no
implementation plan is in flight.

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
- `.github/workflows/go.yml` runs `python3 -m unittest discover -s tests -v`
  against a directory named `tests`; the directory is `test`. Verified:

      $ python3 -m unittest discover -s tests
      → ImportError: Start directory is not importable: 'tests'

- `.github/workflows/web.yml` runs `npm install` and six scripts in `web/`,
  which has no `package.json`.

These were not verified by running the Go gates — there is no Go toolchain on
this host — but each is checkable by reading the cited line against the tree.

**The tracker does not validate and the dispatch query fails on it.**
`docs/workplan/tasks.json` is `{}`, while `tasks.schema.json` requires
`schema_version`, `status_values`, and `tasks`. The availability query the
control panel and contributors both depend on errors rather than reporting
nothing available, which is the one thing a surviving empty tracker was supposed
to guarantee:

    $ ./scripts/list-available-tasks.sh
    → jq: error ... Cannot iterate over null (null); exit 5

**The tracker schema caps what the model says it must not.**
`tasks.schema.json` constrains `id` to `^P[0-5]-[0-9]{3}$`, `phase` to a maximum
of 5, and `brief` to `^phase-[0-5]/`, against
`skills/shared/workplan-model.md:489-490`: "The phase number is open-ended — a
plan derives as many phases as its design set needs, and `tasks.schema.json`
must not cap them." The roadmap has seven stages, so the cap bites immediately.

**The model's own empty-tracker example is invalid.**
`skills/shared/workplan-model.md:221` gives the closed state as
`{"schema_version": 1, "tasks": []}`, which the schema rejects for missing
`status_values`. A close following the model literally would produce a tracker
that fails validation.

**A code comment cites a fixture that was deleted.**
`cmd/fake-harness/main.go:31` says `streamPayload` matches
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
