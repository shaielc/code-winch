# System state

The current workplan was explicitly closed on 2026-10-11. Its sole phase covered
roadmap Stage 0, which remains unfinished: the sandbox serves a page and local
composer, but does not start a harness, stream output, or retain an exchange.
Closure retires the plan; it does not certify its outcome or open the next stage.
`docs/workplan/tasks.json` is a supported empty tracker and
`docs/workplan/tasks.schema.json` remains unchanged. Delivery order and deferred
architectural decisions remain in [the roadmap](roadmap.md).

Source inspection covered every tracked product source, every one of the seven
briefs, the phase/index, the design set, and the prior state report. Retired
artifacts can be read at the [closure baseline](https://github.com/shaielc/code-winch/tree/6582a894efe59ca1d84be6da556d36f7b22cbf9d/docs/workplan).
The tables below record all task dispositions rather than inheriting completion
claims. Surviving `file:line` references describe this closure tree; runtime
observations are limited to the commands actually run below.

## What was done

### Sandbox page, posture, and operator status

The first composition root loads configuration, constructs the attach transport,
serves HTTP, and handles SIGINT/SIGTERM with a five-second shutdown deadline
(`cmd/winch-sandbox/main.go:15-36`). Only `container-standard` is accepted;
compiled address/static defaults survive empty environment overrides, and an
invalid address, unknown profile, or missing `index.html` prevents startup
(`cmd/winch-sandbox/config.go:18-36`). This is the implemented configuration
subset, not the policy/configuration hierarchy in `docs/code-structure.md` §6.

The registered HTTP surface is `GET /healthz`, `GET /api/session`, and static
assets with SPA fallback. Other methods receive 405 and `Allow: GET, HEAD`;
unknown API paths receive JSON 404 (`internal/adapters/transport/attach/server.go:22-40`).
Default posture reports `container-standard` and `network-egress` as unenforced
(`internal/adapters/transport/attach/session.go:9-10`). Static serving checks
resolved paths against the static root
(`internal/adapters/transport/attach/static.go:11-44`); existing contract tests
cover methods, API/SPA separation, path traversal, escaping symlinks, missing
index, and default posture (`test/contract/attach/session_test.go:28-82`,
`test/contract/attach/static_test.go:14-57`). These controls are narrower than
workspace traversal policy and the shared-deployment security model.

`winch status [--url URL]` is the only CLI command (`cmd/winch/main.go:8-17`).
It verifies both health and posture before printing success, rejects empty
controls and malformed/non-2xx responses, and uses a three-second client timeout
(`cmd/winch/status.go:21-76`, `cmd/winch/client.go:8`). The maintained deployment
command is:

```sh
docker compose -f deployments/compose.yml up --build -d
docker compose -f deployments/compose.yml exec sandbox winch status
# Source-defined output; not observed at closure because Docker is absent:
# service: winch-sandbox (ok)
# profile: container-standard
# unenforced control: network-egress
docker compose -f deployments/compose.yml down
```

The image builds the React page and three Go binaries, runs uid 10001, and starts
only `winch-sandbox` (`deployments/Dockerfile:1-24`). Compose listens on container
`0.0.0.0:8080` while publishing host `127.0.0.1:8080`
(`deployments/compose.yml:3-14`). These are source-supported deployment properties;
image build, non-root execution, shutdown, loopback reachability, and routable-host
refusal were not reverified in this environment. An off-host reachability scan
for shared deployment is also unperformed (`docs/security.md` §11 LB11).

### Browser-local composition

The Vite entrypoint renders `App`, which fetches posture and mounts `Composer`
(`web/src/attach/main.tsx:1-11`, `App.tsx:4-45`). Non-empty drafts are echoed exactly
as typed in an append-only `Local messages` list, without ordinals or server
submission. Reload creates a fresh empty state (`App.tsx:7-16,26-44`). Enter
submits, Shift+Enter keeps the browser's newline behavior, success clears and
focuses the textarea, and whitespace-only input preserves the draft and displays
an associated alert (`Composer.tsx:13-65`). CSS uses `white-space: pre-wrap` for
local text (`web/src/attach/style.css:111-114`).

The maintained hand path is the deployed page at `http://127.0.0.1:8080`:
type `hello`, press Enter, observe `Local`, then reload and observe it disappear.
Component tests cover keyboard/button submission, exact text, whitespace
refusal, alert recovery, focus, local marking, and posture
(`Composer.test.tsx:7-75`, `App.test.tsx:6-32`). Their Shift+Enter test supplies
the newline with `fireEvent.change`; it does not prove a real browser's default
keyboard behavior (`Composer.test.tsx:33-44`). The DOM hand check and web tests
were not executable at closure; no real-browser observation is claimed.

### Shipped standalone fake binary

The image carries `fake-harness`, but the sandbox composition root never starts
it (`deployments/Dockerfile:14-24`, `cmd/winch-sandbox/main.go:15-36`). It is usable
as a separate deterministic binary, with required `-run-id`, transcript,
delay, force-failure, malformed-line, and early-exit flags
(`cmd/fake-harness/main.go:51-76,88-119`). A reproducible direct command on a host
with Go is:

```sh
printf '%s\n' '{"id":"manual","text":"echo ping"}' | go run ./cmd/fake-harness -run-id manual
# Source-defined output includes stream.raw payload data "ping\r\n".
```

It reads JSON lines or bare text, emits raw JSON-line records, and produces a
termination observation on SIGINT/SIGTERM (`main.go:67-75,122-142,173-206`). The
fake's 1 MiB input scanner limit is a fixture input limit, not runner output
hardening (`main.go:122-124`). It is not a coding agent: no tool calls, approvals,
usage, provider login, or terminal/PTY integration is implemented. Its private
dialect is malleable and proves no vendor compatibility (`docs/contracts.md` §9).
No fake run was observed at closure because Go is absent.

### Dispatch tooling and the closed tracker

The control panel reads the tracker, maintains task branches and local
reservations, and prepares refine/implement/audit submissions
(`workplan-control-panel/control_panel/task_scheduler.py:95-111,201-250,308-362`).
This is contributor orchestration, not the product control plane. Completion
still comes from the tracked status; local state cannot override `completed`
(`control_panel/state.py:78-105`). UI rows and availability are computed only
from tracker tasks (`control_panel/ui.py:11-44`). Local tests use mocked cloud
services and Git/HTTP fixtures; they prove no live account or deployed panel.

The retained closed body is exactly:

```json
{
  "schema_version": 1,
  "status_values": ["pending", "in_progress", "blocked", "completed"],
  "tasks": []
}
```

Its required keys, root property types, constants, absence of extra properties,
and empty-array constraints were checked directly against the retained schema;
there are no task items or dependency edges to validate. No general JSON Schema
validator is installed. The availability query returned `[]` with exit 0.
The status gate's supported `no-task` path returned exit 0; an unlabelled PR was
correctly refused because no task exists. The panel's tracker/UI/effective-state
paths returned empty tasks and a locally mocked sync returned `prepared: []`
without task preparation or cloud submission. These checks used the actual
closed tracker, not a replacement sample. A future task plan is authored with
`create-phase` and a separate `derive-task` author; none was created here.

### Closure verification and limits

| Command/check | Observed result |
|---|---|
| `python3 -m unittest discover -s test` | 9 tests, OK |
| `cd workplan-control-panel && python3 -m unittest discover -s tests` | 90 tests, OK |
| `bash scripts/list-available-tasks.sh` after closure | `[]`, exit 0 |
| Direct schema checks below; `git diff --exit-code -- docs/workplan/tasks.schema.json` | Passed; schema bytes unchanged |
| `python3 scripts/stamp_task_completion.py --event-file <local-event> --check-only`, label `no-task` | `pull request is labelled no-task; skipping the task gate`, exit 0 |
| Same gate without the label | Expected known-task error, exit 1; no status write |
| Local panel snapshot/UI/effective-state and sync with remote reads mocked | Empty rows/availability/forest; `prepared: []`; preparation/submission hooks not called |
| `git diff --check` and local Markdown path sweep | Passed; no dangling local Markdown links outside generic skill examples |
| `make check` | Exit 2: `gofmt: not found`, then `go: No such file or directory`; no Go gate passed |
| `make test-cycle` | Exit 2: `docker: No such file or directory` at toolchain image build |
| `make e2e` | Exit 2: Docker absent; standing scenario did not run |
| `docker compose -f deployments/compose.yml up --build -d` and `... config` | Exit 127: `docker: command not found`; system start and rendered compose unverified |
| `cd web && npm run format:check && npm run lint && npm run typecheck && npm test && npm run build` | Exit 127 at first gate: `prettier: not found`; later gates not reached |
| Offline dependency-cache check | `npm ci --offline --ignore-scripts --no-audit --no-fund` failed `ENOTCACHED`; dependencies remain unavailable |

The empty body can be rechecked without installing a schema validator:

```sh
python3 - <<'PYTHON'
import json
from pathlib import Path
schema = json.loads(Path("docs/workplan/tasks.schema.json").read_text())
tracker = json.loads(Path("docs/workplan/tasks.json").read_text())
assert schema["type"] == "object" and isinstance(tracker, dict)
assert set(schema["required"]) <= tracker.keys()
assert schema["additionalProperties"] is False
assert set(tracker) <= set(schema["properties"])
for key, constraints in schema["properties"].items():
    if "const" in constraints:
        assert tracker[key] == constraints["const"]
assert schema["properties"]["tasks"] == {"type": "array", "items": {"$ref": "#/$defs/task"}}
assert tracker["tasks"] == []
print("valid empty tracker; graph has no edges")
PYTHON
```

For the check-only status gate, `<local-event>` contained
`{"pull_request":{"title":"Close current workplan","body":"Preserve closure evidence.","head":{"ref":"closure"},"labels":[{"name":"no-task"}]}}`.
Removing its label yields the expected refusal. The panel check used the actual
closed body with `TaskScheduler.snapshot()`/`sync()`, `state.effective_tracker()`,
and `ui.rows()`/`available_ids()`/`forest()`; only the repository refresh was
mocked to avoid remote access, and preparation/submission functions were spied
for zero calls (`task_scheduler.py:95-111,201-250`, `state.py:78-87`, `ui.py:11-44,47-104`).

The checkout's materialized script mode prevents `./scripts/list-available-tasks.sh`
from executing directly here; invoking it with Bash runs the unchanged query.
This local mode is not evidence of the published repository's script mode.
Go, gofmt, golangci-lint, and Docker are unavailable; browser interaction and CI
were not verified. Those limitations do not prove a product failure or a green
gate. The source also exposes a gate weakness: the shell assignment in
`Makefile:21` can hide a missing `gofmt` because its output is empty; the observed
`make check` subsequently failed at vet. Formatting was not validated.

## What went wrong

### Every retired task was settled against source

The [historical tracker](https://github.com/shaielc/code-winch/blob/6582a894efe59ca1d84be6da556d36f7b22cbf9d/docs/workplan/tasks.json)
and [briefs](https://github.com/shaielc/code-winch/tree/6582a894efe59ca1d84be6da556d36f7b22cbf9d/docs/workplan/phase-0)
are evidence for recorded statuses and acceptance criteria. IDs below are
historical citations, not live owners or assignments.

| Retired task | Recorded status | Acceptance at closure HEAD |
|---|---|---|
| P0-001 — sandbox page/posture | completed | Composition root, configuration, transport, CLI, container definition, and gates target existing sources. Runtime deployment, non-root/port/shutdown demonstration and green Go/web/Docker/CI gates remain unverified; full acceptance is not certified (`cmd/winch-sandbox/main.go:15-36`, `config.go:18-36`, attach `server.go:22-40`, `static.go:11-44`, CLI `status.go:21-76`, `deployments/Dockerfile:1-24`, `compose.yml:3-14`, `Makefile:6-52`, contract tests cited above). |
| P0-002 — harness ownership/live output | pending | Not implemented: no runner, record/ordinal allocation, stream handler, decoder, `winch stream`, fake launch configuration, or live-output scenario (`git ls-files internal cmd test/e2e`; `cmd/winch-sandbox/main.go:21-25`, attach `server.go:28-40`, CLI `main.go:8-17`). Fake controls exist only on the separate binary; no integrated profile acceptance holds. |
| P0-003 — retained records/history | pending | Not implemented: no store, append-before-publish, records API, ordinal query/validation, persistent volume, restart continuity, history CLI, or replay scenario (same source inventory; `deployments/compose.yml:7-14,36-39`, `App.tsx:7-16`, attach `server.go:28-40`). |
| P0-004 — local composer | completed | Source and component test cases support local labeling/no ordinals, exact text, Enter/button, whitespace alert, clearing/focus, accessible names, and transient state. Runtime/browser, Shift+Enter default, gate preservation, and original-task write-set acceptance are not independently certified; closure performs no backend behavior change (`App.tsx:7-44`, `Composer.tsx:13-65`, web test files cited above). |
| P0-005 — recorded input | pending | Not implemented: POST is rejected globally; no submission record/idempotency/refusal API, payload bound, ordinal, CLI, or canary/scenario tests. Local append and positional React keys remain (`server.go:22-25`, `App.tsx:32-43`, CLI `main.go:8-17`). |
| P0-006 — harness delivery/reply | pending | Not implemented: no harness stdin ownership, accepted-before-delivery path, write deadline, unavailable-child refusal, diagnostic/correlation record, or delivery scenario. The fake parses but discards command IDs and its output does not echo them (`cmd/fake-harness/main.go:173-186`); there is no product delivery path to test. |
| P0-007 — malformed output/stalled readers | pending | Not implemented: no runner decoder, output-size/resynchronization bound, subscriber queues/disconnect reasons, confidential raw diagnostic record, or stalled-reader/canary scenarios. Fixture `-malformed-line` writes a bad line then exits 1 (`main.go:103-106`); it does not establish product recovery. |

The two completed statuses have source-supported capabilities, but their full
runtime acceptance cannot be recovered from status alone. None of the five
pending outcomes was silently promoted to complete. The source inventory is
recheckable with `git ls-files cmd internal test/e2e web/src`; it contains no
`internal/runner/`, stream/records/input handler, or corresponding CLI/scenario.

### The phase revision and graph were not reconciled

The [retired phase design](https://github.com/shaielc/code-winch/blob/6582a894efe59ca1d84be6da556d36f7b22cbf9d/docs/workplan/phase-0/README.md#task-coverage)
specified eight increments while the tracker retained seven older briefs.
Pending briefs were explicitly unready for assignment, and output-size limits
had no separately derived coverage. Live output must include launch and minimum
child cleanup/bounds; history, recorded input, delivery, reader isolation, and
additional output-size enforcement remain distinct observable concerns.
Structured decoding/native exit reporting belong to roadmap Stage 1; multiple
views/no-gap resume to Stage 2; intentional stop/retry to Stage 4; escalation
and orphan recovery to Stage 5. Rule: revise a phase before combining increments,
then reconcile task coverage and graph before dispatching old briefs.

The [retired index](https://github.com/shaielc/code-winch/blob/6582a894efe59ca1d84be6da556d36f7b22cbf9d/docs/workplan/README.md#width)
identified an unencoded collision: delivery-failure diagnostics and malformed
output diagnostics shared a contract, but their briefs could become concurrent.
It also retained a history dependency for combined reader hardening even though
the revised reader/size concerns depended only on live output. No implementation
had reached these seams. Rule: declare every emitted contract, assign its
meaning one owner, encode contract collisions as justified edges, and keep
write overlaps as merge warnings rather than artificial dependencies. The
legacy graph had seven tasks, critical path five, average width 1.4, and maximum
two; it was not evidence of the revised phase's width.

### Invariants at closure

| Invariant | Source/observed disposition |
|---|---|
| I1 — runs | HTTP composition root exists and is wired; start not observed without Go/Docker (`cmd/winch-sandbox/main.go:15-36`). No evidence of a newly lost running slice. |
| I2 — deploys | Dockerfile/compose/configuration exist; actual image build, startup, user and ports unverified (`deployments/Dockerfile:1-24`, `compose.yml:3-14`). |
| I3 — shipped controllable fake profile | Partial: binary and controls exist, but no integrated product configuration launches it; no all-fake product hand check (`main.go` and fake source cited above). |
| I4 — standing scenarios/parity | Only startup/posture/page/status scenario exists (`test/e2e/scenario_sandbox_starts_test.go:10-30`). It neither drives harness I/O nor browser DOM; parity and composed execution unverified. Test topology, user/publication checks and teardown live in `Makefile:39-52`, not in the scenario helper (`test/e2e/sandbox.go:18-42`). |
| I5 — hand reachability | Status CLI and mounted local composer have maintained paths; no implemented backend capability is awaiting a separate operator wrapper. Runtime hand checks remain unverified; stream/records/input commands do not exist (`cmd/winch/main.go:8-17`, `App.tsx:41-45`). |
| I6 — owned deferrals | All brief deferrals were read and carried into the gaps below: retention/replay, malformed output and readers, input recording/refusal, stable persisted UI identity, delivery and after-exit refusal. No retired ID remains a live owner. Roadmap D1–D7 retain their explicit triggers (`docs/roadmap.md` §6). |
| I7 — width | Legacy graph/revised increments disagreed; collision above was known and uncorrected. After explicit closure the graph is empty; no width or outcome completion is claimed. |

### Lessons retained from the earlier closure

There were no actual post-mortem records in this plan: its directory held only
[the format README](https://github.com/shaielc/code-winch/blob/6582a894efe59ca1d84be6da556d36f7b22cbf9d/docs/workplan/post-mortems/README.md).
There was no propagation directory or ledger, hence no recorded unresolved
accepted finding or task/issue destination to carry. This is not evidence that
external PR reviews or issues contained none; those sources were not inspected.
The [prior state report](https://github.com/shaielc/code-winch/blob/6582a894efe59ca1d84be6da556d36f7b22cbf9d/docs/state.md#what-went-wrong)
recorded older failures and rules that remain part of the system record:

- Building downward from control-plane/domain abstractions never yielded a
  running system. Rule: start and deploy the smallest observable slice first;
  introduce abstractions when a second runtime case forces them (roadmap §1,
  ADR-0005). This plan restored source for that first slice, not the whole stage.
- Removing the old plan without its state report left an authoritative empty
  heading and stale contributor links. Rule: report and retirement are one
  commit; scan contributors, workflows, prompts, scripts, ADRs and source for
  retired links and deferral owners (`skills/close-phase/SKILL.md`).
- Product implementation was removed with the earlier plan, leaving gates aimed
  at absent files. Rule: retire planning artifacts, preserve working code,
  reconcile every gate against the sources that exist. Current Makefile and
  workflows target existing sources, but closure cannot certify they pass here
  (`Makefile:6-52`, `.github/workflows/go.yml`, `web.yml`).
- The earlier tracker was `{}`, and an older closure example omitted required
  `status_values`. Rule: keep the unchanged schema and its smallest accepted
  tracker, validate required fields/constants, and exercise availability plus
  the `no-task` gate (`tasks.schema.json:5-19`, completion script `:155-186`).
- An earlier report falsely claimed the schema capped phases at five. Rule:
  inspect source before changing a frozen format. It is already open-ended
  (`tasks.schema.json:34-36,48`); no schema change is justified.
- The fake's comment cited a deleted raw-stream fixture and an absent codec.
  Rule: an inherited wire-format comment is not executable compatibility
  evidence; document the actual fixture dialect and create product contract
  evidence only when the product capability exists. Closure corrects these
  comments; product record/golden tests remain absent (`cmd/fake-harness/main.go:21-43`,
  source inventory above).

The old statements that there was no sandbox/page/CLI/scenario and that all
security controls were absent are superseded by the narrow capabilities above.
The earlier gate-path failures describe history, not the current source tree.
The prior snapshot's claimed Docker availability does not describe this host.

## What is not implemented

These are system gaps, not assignments. The roadmap retains ordering and D1–D7
retain decision triggers; closure assigns no new task IDs.

**A sandbox that watches and talks to its harness.** The HTTP root registers no
runner or child process. No raw output stream, runner-local record shape or
gap-free ordinal, timestamp/sensitivity mapping, incremental decoder, child
reaping/cleanup, session store, durable volume, snapshot pagination, replay, or
restart continuity exists (`cmd/winch-sandbox/main.go:21-25`, attach
`server.go:28-40`; `git ls-files 'internal/runner/*' 'cmd/winch/stream.go'
'cmd/winch/records.go'` returns nothing). `docs/contracts.md` §8's live and history
promises remain gaps; reporting posture alone does not implement them.

**Recorded and delivered input.** The browser never submits to a server; POST is
rejected. No accepted/refused typed input, persisted submission, session
idempotency, bounded body, correlation, unavailable-child refusal, delivery
write deadline, or accepted-but-undelivered diagnostic exists (`App.tsx:41-44`,
attach `server.go:22-25`, CLI `main.go:8-17`). Local entries retain positional
React keys (`App.tsx:32-35`); persistent records have no stable UI identity yet.
The fake accepts an ID but drops it before emitting a reply
(`cmd/fake-harness/main.go:173-186`). A fake echo alone cannot satisfy Stage 0's
page-to-harness-to-page exchange and reload check.

**Output hardening and resumable sessions.** No product output-size limits,
bounded fan-out, stalled-reader isolation/disconnection with last ordinal,
content-free drop diagnostic, malformed-output preservation/classification,
resynchronization, heartbeat/caught-up marker, gap repair or no-gap reconnect
exists (attach `server.go:28-40`; no runner/stream sources). The malformed-line
fixture exits and therefore needs suitable behavior to exercise future recovery
(`cmd/fake-harness/main.go:103-106`). No real browser or integrated all-fake
scenario covers these promises (`test/e2e/scenario_sandbox_starts_test.go:10-30`).
The local browser list itself is unbounded (`App.tsx:7,41-44`), and posture JSON
is type-asserted rather than validated (`App.tsx:9-15`, `Posture.tsx:15-19`).

**Real harness integration and sandbox drivers.** No vendor CLI, descriptor,
capability negotiation, `HarnessDriver`/`HarnessCodec`, native exit interpretation,
provider extension schema, provider login, or PTY implementation exists. No
`local`/`docker` driver lifecycle or shared adapter contract suite exists. A
Compose container definition is not such a driver (`git ls-files
'internal/adapters/harness/*' 'internal/adapters/sandbox/*'` returns nothing;
`docs/code-structure.md` §3–5 and `docs/contracts.md` §4,9 describe the destination).

**The product control plane and canonical durability.** No `winchd`, domain or
application aggregates, run lifecycle/attempts, supervisor, leases/fencing,
registration/heartbeats/remote commands/artifact transfer, authenticated runner
protocol/version negotiation, OpenAPI/client generation, PostgreSQL/SQLite
store/migrations, canonical event/sequence assignment, transactional outbox,
scheduling/reconciliation or product UI over multiple sandboxes exists
(`git ls-files cmd internal web/src` shows only sandbox/status/fake, attach
transport, and local attach UI; `docs/architecture.md` §4–8,
`docs/contracts.md` §1–5). The contributor control panel does not fill this gap.

**Renderers, artifacts and workflows.** No terminal/conversation/activity/changes
projection, bounded/shared renderer model/cache, untrusted-renderer isolation,
artifact object storage/export, workflow definitions/runtime/coordinator,
step leases, retry/compensation or cross-run workflow exists (`git ls-files
'web/src/renderers/*' 'internal/workflow/*' 'internal/domain/*'
'internal/application/*'` returns nothing; `docs/contracts.md` §6–7,
`docs/architecture.md` §6–7). Search and cross-session history are also absent.

**Shared-deployment security and data lifecycle.** Existing posture reporting,
loopback publication source, non-root image source, static traversal/symlink
checks, React text rendering, and content-free startup/CLI errors do not establish
LB01–LB11. No product identities/object authorization, CSRF/WebSocket origin or
reauthorization, bearer token, rate/payload policy, credential references/provider
flows/scoped injection, redaction/canaries, approved disposable workspaces,
resource/stop/orphan guarantees, deny-egress/allowlist/metadata controls,
renderer sanitization/CSP/fuzzing, encryption, security audit, retention/export/
deletion/cache/backup policy, provenance/digest pinning/SBOM/scans, or incident/
revocation/recovery exercise is implemented (`docs/security.md` §2–12;
`deployments/Dockerfile:1,8,18-24`, `compose.yml:3-14`; tracked product source
inventory above). Image tags are not digest pins, and network egress is explicitly
unenforced. Off-host LB11 evidence is absent even though the declared host port
is loopback. No shared-deployment blocker is certified closed.

**Verification still required to establish the existing slice.** Go and web
quality gates, the deployed image and standing scenario, the real browser
composer demonstration, non-root/runtime port/shutdown checks, and CI results
remain uncertain for the reasons in the verification table. The format-check
shell failure propagation issue remains (`Makefile:21`). No runtime regression
was inferred from missing host dependencies, and no unobserved gate was marked
successful.
