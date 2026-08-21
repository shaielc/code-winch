# System state

Written when the implementation workplan closed, at commit `ccad757`, on
2026-08-21. It replaces `docs/workplan/`, which was removed in the same commit
and remains in git history.

This is a statement about the system, not a plan. It records what runs, what
went wrong, and what has no working code behind it. The next plan is derived
from the design set (`docs/architecture.md`, `docs/code-structure.md`,
`docs/contracts.md`, `docs/security.md`, `docs/roadmap.md`, `docs/decisions/`)
together with this document. Every claim below carries the command that shows
it or a `file:line`.

## How these claims were checked

Go 1.24.6, Node 22.18, golangci-lint 2.1.6, and PostgreSQL 16.4 on the host.
The daemon was run as a native binary against a real PostgreSQL 16.4 with the
same `main.run` startup path the container uses, because Docker was unavailable
in the closing session. The compose stack in `deployments/compose.yml` was read,
not executed: nothing below claims `docker compose up` was observed.

Everything else was executed at `ccad757`: `make check` in full, the
`integration`-tagged suite against a live database, the web workspace's five
gates, and the operator CLI by hand.

## What was done

### The daemon starts, migrates, and serves

`cmd/winchd` loads configuration, connects to PostgreSQL, applies migrations,
then opens its listener.

    $ WINCH_DATABASE_URL=postgres://winch@127.0.0.1:55432/postgres?sslmode=disable \
      WINCH_TOKEN=... WINCH_CSRF_TOKEN=... WINCH_ADDR=127.0.0.1:18080 winchd
    {"msg":"schema checked","component":"database","operation":"migrate","sequence":5,"status":"applied"}
    {"msg":"startup complete","component":"daemon","operation":"start","status":"ready","duration_ms":61}
    {"msg":"listener started","component":"http","operation":"listen","status":"ready"}

    $ curl -s http://127.0.0.1:18080/api/v1/health
    {"status":"ok"}

A second start against the migrated database applies nothing:

    {"msg":"schema checked",...,"sequence":5,"status":"current"}

Startup fails closed and names fields rather than values:

    $ WINCH_TOKEN=short WINCH_CSRF_TOKEN=alsoshort winchd
    invalid configuration fields: token, csrf_token
    ERROR daemon stopped component=daemon error_code=startup_failed
    $ echo $?
    1

Configuration resolution is compiled defaults, then an optional YAML file, then
environment variables (`internal/platform/config/config.go:39`). Log records are
redacted through `telemetry.NewHandler` before reaching the JSON handler
(`cmd/winchd/main.go:50`), and the metric registry refuses to start on an
undeclared metric or label (`cmd/winchd/main.go:93`).

The deployment stack — `deployments/compose.yml`, `deployments/Dockerfile.winchd`,
`deployments/nginx.conf` — builds daemon, fake harness, and web assets into one
image and publishes it on loopback only. `deployments/README.md` describes the
stack accurately, including the run routes being unbound.

### The HTTP surface exists and enforces its security policy

Routes are mounted from the OpenAPI document at `api/openapi/code-winch.yaml`.
Authentication, CSRF, and origin checks apply before any handler:

    $ curl -s -o /dev/null -w '%{http_code}\n' -X POST http://127.0.0.1:18080/api/v1/runs -d '{}'
    401

Problem responses are RFC-9457 shaped with a stable `code` and a request ID.
The daemon injects the CSRF token into the served `index.html` and never puts it
in a cookie (`cmd/winchd/main.go:140`). Session cookie attributes, actor
binding, and content-free rejection logging are covered by
`TestAuthenticationCSRFAndObjectActor`, `TestSessionCookieSecurityAttributes`,
`TestHandlerDropsFreeFormAndSecret`, and `TestRejectionLogKeepsCorrelationIDAndErrorCode`
in `internal/adapters/transport/httpapi/server_test.go`.

The contract itself is gated: `make api-validate` parses the document,
`make api-compat` diffs it against `test/contract/openapi/v1.yaml` with oasdiff,
and `make api-check` proves the Go and TypeScript generated output is
deterministic. All three pass and leave the tree clean.

### A harness can be driven by hand through the local runner

`winch dev run` is the only path in the repository along which a byte moves
between Code Winch and a harness. It constructs the local PTY sandbox, the fake
harness adapter, and the in-process runner, with no daemon and no database.

    $ go build -o /tmp/winch ./cmd/winch && go build -o /tmp/fake-harness ./cmd/fake-harness
    $ printf 'echo hello\nexit\n' | PATH=/tmp:$PATH /tmp/winch dev run --harness fake --sandbox local
    [started]
    fake harness ready: run_id=00000000-0000-0000-0000-000000000001
    ...
    hello
    fake harness exiting
    [exit] successful=true code=OK

Stop escalation reaps its children. Three consecutive runs left nothing behind:

    $ for i in 1 2 3; do /tmp/winch dev run --harness fake --sandbox local --stop-after 1s </dev/null >/dev/null 2>&1;
        sleep 0.3; ps -eo pid,cmd | grep -c 'fake[-]harness'; done
    0
    0
    0

The runner assigns runner-local ordinals only and writes no events
(`internal/runner/local/runner.go`). Ordering, flush-before-exit, stale-lease
rejection, and orphan cleanup are covered by
`TestObservationsAreOrderedAndFlushPrecedesExit`,
`TestForcedStopLeavesNoDescendant`, `TestRejectionsAreStableAndContentFree`, and
`TestCleanupKillsChildAndToleratesRetries`.

### The browser app builds and is served

`web/` builds a React SPA that the daemon serves from `WINCH_STATIC_DIR`, and
its five gates pass:

    $ cd web && npm run format:check && npm run lint && npm run typecheck && npm test && npm run build
    All matched files use Prettier code style!
    Test Files  2 passed (2) / Tests  3 passed (3)
    ✓ built in 625ms

The running daemon serves it at the same origin as the API, with the CSP from
`web/index.html` intact and the CSRF token injected into the document:

    $ curl -s -o /dev/null -w '%{http_code}\n' http://127.0.0.1:18080/
    200

**Its runtime behavior was not observed.** No browser was opened against this
build, and none could usefully be: the app's only entry is a form that opens a
run by ID (`web/src/app/App.tsx:30-40`) and no run can be created, so the
snapshot fetch it issues answers 404. What is verified is the code's behavior
under jsdom with the network faked. `web/src/app/App.test.tsx:5-20,50-71`
replaces `WebSocket` with a `FakeSocket` class and stubs `fetch` outright;
against those doubles it asserts
resume with `after_sequence`, a re-delivered event not duplicating output,
input, stop, capability display, and a disabled resize control.
`web/src/renderers/ansi.test.ts` shows an OSC-8 `javascript:` link and an inline
`<script>` reduced to inert text. Three tests, none of which has ever touched a
daemon.

### Durable storage holds runs, events, input commands, an outbox, and workflows

`internal/adapters/postgres` implements the run repository, event store, input
command store, outbox store, and workflow store over five migrations plus a
`schema_migrations` ledger (`internal/adapters/postgres/migrate.go`). Against a
live database:

    $ PG_TEST_DATABASE_URL=postgres://winch@127.0.0.1:55432/postgres?sslmode=disable \
      go test -tags integration ./internal/adapters/postgres/...
    ok  github.com/shaielc/code-winch/internal/adapters/postgres  0.878s

Ten tests pass, covering gap-free concurrent sequence assignment with a secret
canary, atomic and replayable input acceptance, exclusive outbox claims,
rollback leaving no publish intent, optimistic run concurrency, workflow step
lease contention and expiry reclaim, and both migration properties —
`TestMigrationRoundTripOnCleanDatabase` and `TestMigrateUpOnMigratedDatabaseIsANoOp`.

### Domain, contracts, and adapters exist as libraries

- `internal/domain` — typed identifiers, timestamps, and the eight-state run and
  attempt machine, with retry as a linked attempt.
- `pkg/protocol` — runner protocol v1 and the canonical event envelope, with
  version negotiation, bounded payloads, and round-trip fixtures under
  `schemas/runner/v1/` and `schemas/events/v1/`.
- `internal/application` — ports, the input service, and the outbox worker.
- `internal/supervisor` — per-run command serialization, lease fencing, and
  restart reconciliation.
- `internal/adapters/harness/{fake,codex}` and `internal/adapters/sandbox/{fake,local}` —
  two harness adapters and two sandbox drivers, all passing the shared contract
  suites in `test/contract/`.
- `internal/adapters/memory` — in-memory implementations of every port.
- `internal/workflow` — workflow graph schema validation.

Ninety Go tests cover these. None of the packages in this list is reached by any
composition root; see *What went wrong*.

### The repository's quality gates pass in full

    $ make check        # api-check, format-check, vet, lint, test, build
    ...
    0 issues.

`make check` is what `.github/workflows/go.yml` runs on every push and pull
request; `.github/workflows/web.yml` runs the five web gates plus `api:check`.

### Task-dispatch automation

`scripts/task_scheduler.py`, `scripts/stamp_task_completion.py`,
`scripts/list-available-tasks.sh`, `runner/`, and the
`task-scheduler`/`task-status-gate` workflows dispatched available tasks to
Codex Cloud and stamped completion on approval, with the tracker on the default
branch as sole authority (`docs/decisions/0004-task-status-authority.md`). They
all read `docs/workplan/tasks.json`, which no longer exists; see *What is not
implemented*.

## What went wrong

### Every post-mortem rule the plan produced

One post-mortem was written, on migration re-runnability. `MigrateUp`
concatenated five migrations into one unconditional `Exec`, so every boot after
the first failed with `relation "runs" already exists` and the daemon exited
before listening. The migrations task planned them to run once on a clean
database; the boot task made migration a per-boot operation and recorded the
no-op requirement under *Verification* rather than as scope. No single brief was
wrong on its face. Its three rules, restated:

1. **An invariant that only names a property creates no owner for it.** "Migrations
   run at startup" was stated as an invariant, and no task's acceptance criteria
   ever said they must therefore survive a second start.
2. **When a later task's verification asserts a property of an earlier task's
   surface, that is scope, not verification.** The surface belongs in the later
   task's owned surfaces and the work belongs in its brief body.
3. **Read "tested on a clean database" as a statement about which environments
   have been considered, not only about test setup.** Every migration test began
   with `DROP SCHEMA public CASCADE`, so the second-run path was unreachable
   from the suite by construction.

The remediation holds at HEAD: a `schema_migrations` ledger applying each
migration once in its own transaction under an advisory lock
(`internal/adapters/postgres/migrate.go`), proven by the second-boot log line
above and by `TestMigrateUpOnMigratedDatabaseIsANoOp`.

### Twenty-four tasks completed against a system that never ran a run

This is the central failure and it is structural, not a slip. Nineteen tasks
were marked complete while `cmd/winchd/main.go` was `func main() {}`. Every
component a deployment needs was built — storage, outbox, supervisor, HTTP
adapter, WebSocket stream, harness adapters, sandbox drivers, browser slice —
and nothing connected them. A later wave gave the daemon a composition root and
the runner a hands-on command, but binding the use cases was left to a task that
never landed.

The result at HEAD: no run can be created, so no run can be started, streamed,
input to, or stopped through the product.

    $ curl -s -o /dev/null -w '%{http_code}\n' -X POST http://127.0.0.1:18080/api/v1/runs \
        -H "Authorization: Bearer $T" -H "X-CSRF-Token: $C" -H "Origin: http://localhost:18080" \
        -d '{"workspacePath":"/tmp/ws","harnessProfile":"fake","sandboxProfile":"local"}'
    500
    $ curl -s -o /dev/null -w '%{http_code}\n' http://127.0.0.1:18080/api/v1/runs/01HZY0000000000000000000AA -H "Authorization: Bearer $T"
    404
    $ curl -s -o /dev/null -w '%{http_code}\n' http://127.0.0.1:18080/api/v1/runs/01HZY0000000000000000000AA/events -H "Authorization: Bearer $T"
    404

The cause is `unavailableBackend` at `cmd/winchd/main.go:159-183`, a stub
implementing every run operation as "not bound".

Five packages have no non-test importer anywhere in the tree:

    $ for p in internal/supervisor internal/adapters/memory internal/adapters/harness/codex \
        internal/adapters/sandbox/fake internal/workflow; do
        grep -rl "code-winch/$p\"" --include='*.go' cmd/ internal/ pkg/ test/ | grep -v _test.go; done
    (no output)

- `internal/supervisor` — one-writer-per-run, lease fencing, restart
  reconciliation. Nothing constructs a supervisor.
- `internal/adapters/harness/codex` — a complete adapter for the Codex CLI JSON
  protocol, registered nowhere.
- `internal/adapters/memory` and `internal/adapters/sandbox/fake` — test-only in
  practice, which is exactly the "port whose only implementations are test
  doubles" shape the plan warned against.
- `internal/workflow` — workflow graph validation, with no coordinator to use it.

Three more packages are imported for a fraction of what they hold:

- `internal/adapters/postgres` — `postgres.New` (the store) is never called;
  `cmd/winchd/main.go:63` uses the package for `MigrateUp` alone, so the run,
  event, input-command, outbox, and workflow tables are created and never
  written to. `workflow_repository.go` has no caller at all.
- `internal/application` — `cmd/winch` imports it for one event type.
  `NewOutboxWorker` and `NewInputService` are called only from tests, so nothing
  drains the outbox table the storage layer writes to.
- `internal/adapters/transport/httpapi` — mounted, but with `unavailableBackend`
  behind it.

`httpapi.EventStream` is constructed and wired into the handler
(`cmd/winchd/main.go:73`), but `EventStream.Publish` has no non-test caller, so
the live path is a socket with no producer.

### Statuses that do not hold at HEAD

One completed task's acceptance criterion is false: *"The fake harness drives a
complete run in CI."* No test anywhere executes the `cmd/fake-harness` binary.

    $ grep -rn 'fake-harness' --include='*_test.go' . | grep -v node_modules
    (no output)

The fake harness adapter is contract-tested in process, and the runner tests
drive shell scripts under a real PTY, but the deterministic executable harness
that exists to make a complete run possible without a vendor account is
exercised only by hand. There is no complete run in CI to drive, because there
is no end-to-end suite at all.

The remaining twenty-three completed tasks were each settled against the
repository rather than against a pull request: the artifacts their criteria name
exist, and the tests that prove them are present and green under `make check`
and the integration run above. Their criteria hold as component-level claims —
but nine of them describe behavior of code no runtime configuration reaches. A
criterion like "restart can rehydrate from durable state" is true of
`internal/supervisor` and false of the deployed product, because the deployed
product has no supervisor.

One judgement call is worth passing on. The boot task required that
`deployments/README.md` describe "the stack that exists, with no 'not wired yet'
caveat remaining", and that file still says the run routes are mounted and
unbound. That was read as holding: the caveat is an accurate statement of the
deployed system rather than a promissory note about the daemon itself.

### Invariants that were never established

The plan's own floor was set down and not built on.

- **The system runs from the first task.** It did not. The composition root was
  empty through the entire first component wave and was retrofitted afterwards.
  It runs now, and starting it demonstrates nothing about a run.
- **Substrate swaps prove parity against a standing scenario suite.** There is
  no suite. `test/e2e` does not exist and `make e2e`, cited throughout the plan's
  index, was never a target in the `Makefile`. Every "same scenario, more real"
  acceptance criterion in the plan was therefore unenforceable.
- **Every operator-visible capability is immediately reachable by hand.** Partly.
  `winch dev run` reaches the harness pump; nothing reaches a run. The CLI is
  also not built by `make build` (`Makefile:75-78` builds `cmd/winchd` only) and
  is not installed into the daemon image (`deployments/Dockerfile.winchd:39,42`
  installs `winchd` and `fake-harness`), so the maintained hands-on surface is
  absent from the deployment it is meant to drive.
- **The fake configuration is a first-class controllable runtime profile.**
  Partly. The fake harness runs and is honest about what it is, but it is not
  controllable — no scripted transcript, injectable latency, malformed output, or
  early exit. Its command set is fixed at `cmd/fake-harness/main.go:42-47`. It is
  also launched by bare name from `PATH`
  (`internal/adapters/harness/fake/fake.go:28`), so on a host it works only if
  the operator has built and placed the binary themselves.

### Gates that do not cover what they appear to

- **CI never runs the integration suite.** `.github/workflows/go.yml` runs
  `make check`, whose test leg is `go test ./...`; the storage tests are behind the
  `integration` build tag and additionally skip unless `PG_TEST_DATABASE_URL` is
  set (`internal/adapters/postgres/repository_integration_test.go:148-151`).
  Every durability, concurrency, and lease guarantee in the storage layer is
  proven only when someone runs `make test-integration` locally.
- **`make test` walks `node_modules`.** `go test ./...` from a tree where
  `npm install` has run descends into `web/node_modules/flatted/golang/pkg/flatted`.
  It passes today; the gate's scope is nonetheless not what it appears to be.

### Documentation that outran the code

`README.md:22-24` at `ccad757` said the repository has "an intentionally empty
daemon composition root" and that "domain, adapter, and product behavior have
not been implemented yet". Both were false long before this close. The closing
commit replaced that section with a pointer here.

## What is not implemented

Design-set promises with no working code behind them. Each is a gap in the
system as it stands.

### The run round trip

`docs/architecture.md` §4 names `CreateRun`, `StartRun`, `SendInput`,
`StopRun`, and `ResumeSubscription` as application services. None exists. The
HTTP operations, the storage, the supervisor, the runner, and the browser page
that would consume them all exist separately; the use-case layer that joins them
does not. Concretely missing: a driver registry the composition root reads, a
backend binding for `httpapi.Backend`, an outbox worker started by the daemon and
publishing into `EventStream`, and browser session establishment (no code issues
the `winch_session` cookie the handler accepts).

`docs/roadmap.md` Phase 1's exit statement — a browser reconnecting to a live
run without losing ordered history, a truthful terminal state after daemon
restart, forced stop leaving no child processes — is not demonstrable. Only the
third clause can be shown, and only through `winch dev run`.

### The standing end-to-end scenario suite

No `test/e2e` directory exists and no `make` target runs one. Nothing anywhere
drives the deployed system through `create → start → stream → input → stop`.

### Credentials, workspaces, and login

`docs/architecture.md` §1 names browser-based login as a goal and §6 names
`Credential` and `Workspace` as aggregates. Neither aggregate exists in the
domain, and the five migrations create runs, run attempts, run events, the
outbox, input commands, and the workflow tables — no credential or workspace
table among them. `application.SecretReferenceStore`
(`internal/application/ports.go:104`) has exactly one implementation,
`internal/adapters/memory/memory.go:299`, which nothing reaches: a port whose
only implementation is a test double. A run today names an arbitrary `workspacePath` string
(`api/openapi/code-winch.yaml:258`) with no registration, ownership, or policy
behind it, and no harness receives any credential.

### Structured events, renderers, and the second provider

`docs/contracts.md` §2 lists the canonical event families, and
`schemas/events/v1/` carries the envelope schema and twelve payload fixtures.
Only raw stream output is produced or rendered. There is no normalization from harness output into typed message,
tool-call, approval, file-change, artifact, or usage events; no conversation,
activity, or diff renderer; and no renderer selection. The Codex adapter exists
but is registered nowhere, so `docs/roadmap.md` Phase 2's exit statement — the
same views over two providers — has never been tested against a second provider.

### Approvals, retry, and queue admission

`docs/contracts.md` §3 defines approval and structured-answer input payloads.
Nothing binds an approval to an operation or resolves one. `domain.Run`'s
`Failed → Queued` retry transition (`internal/domain/run.go:113-124`) is
implemented and tested and reachable through no API. `RunStateQueued` exists and
means nothing: no admission control bounds concurrent or queued runs, so nothing
limits starting runs in a loop.

### Artifacts, retention, authorization, and audit

`docs/architecture.md` §6 names `Artifact` as an aggregate and
`docs/security.md` §5 defines retention, export, telemetry, and deletion
defaults per sensitivity class. No artifact storage, no export, no deletion, no
retention schedule, and no audit trail exist. Authorization is a single shared
bearer token plus a configured actor string
(`internal/platform/config/config.go:35`); holding a resource ID is sufficient
to act on it.

### Container isolation

`docs/roadmap.md` Phase 3 in full. There is no Docker sandbox driver, no named
sandbox profiles, no network deny/allowlist, no scoped credential injection, no
disposable workspace preparation, and no security-posture display. The only
sandbox is `local`, which reports `unisolated` honestly
(`deployments/README.md`, "Security posture"). `docs/security.md`'s LB08 —
dependency scanning and an SBOM — has no implementation and no CI step.

### Workflows

`docs/architecture.md` §4's workflow coordinator does not exist. Graph schema
validation (`internal/workflow`) and instance/step-lease persistence
(`internal/adapters/postgres/workflow_repository.go`) are implemented and
tested, and nothing advances an instance, executes an activity, exposes a
workflow API, or renders a graph. `docs/roadmap.md` Phase 4's exit statement is
untestable.

### Remote runners

`docs/architecture.md` §5 Phase B in full. The runner protocol is serializable
and versioned (`pkg/protocol/runner.go`) and has never crossed a process
boundary: no `cmd/winch-runner` binary, no registration, no heartbeats, no
distributed lease fencing, no capability scheduler, no artifact handoff, no HA
validation, and no SLOs, dashboards, alerts, or runbooks.

### Inert code left in the tree

- `unavailableBackend`, `cmd/winchd/main.go:159-183` — six run operations
  returning "not bound". Reachable, and the reason every run route fails.
- The five unreachable packages and the unreached halves of three more, listed
  under *What went wrong*. They compile, test, and lint clean; no runtime
  configuration reaches them.
- `config.ParseInt`, `internal/platform/config/config.go:102` — exported and
  called by nothing.

There are no `TODO`, `FIXME`, or `panic("not implemented")` markers in the tree.

### Automation left without its tracker

Removing `docs/workplan/` leaves the task-dispatch machinery reading a file that
does not exist. It will not fail quietly:

- `.github/workflows/task-status-gate.yml:27` runs
  `scripts/stamp_task_completion.py --check-only` on **every** pull request, and
  that script opens `docs/workplan/tasks.json` unconditionally
  (`scripts/stamp_task_completion.py:15,21`). The gate now fails on every pull
  request until a new tracker exists or the workflow is disabled.
- `.github/workflows/task-scheduler.yml` and `scripts/task_scheduler.py:94`
  (`git show <remote>/<branch>:docs/workplan/tasks.json`) dispatch nothing.
- `scripts/list-available-tasks.sh:6` and the `runner/` control panel have no
  input.
- `AGENTS.md` §"Start here" and its "Current state" section, `README.md`'s
  design-document list, `scripts/task-prompt.md`, `runner/README.md:96`,
  `docs/decisions/0004-task-status-authority.md`, and the migration-slot comment
  at `internal/adapters/postgres/migrate.go:30` all refer to plan files that are
  gone. `README.md` and `AGENTS.md` were updated in the closing commit; the
  others were left as they are.
