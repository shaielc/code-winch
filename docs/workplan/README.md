# Implementation workplan

This directory converts the architectural baseline into implementation-sized
task briefs. Each task is intended to be implemented in its own pull request
unless its brief is explicitly split further.

## How to use this plan

- Treat **Dependencies** as completion prerequisites, not suggested reading.
  Every edge names its reason — `compile`, `contract`, or `semantic` — in one
  clause. An edge you cannot justify in one clause is not a real edge.
- Tasks with satisfied dependencies may proceed in parallel; never infer a
  dependency from numeric order alone. Check **Owned surfaces** before starting
  two tasks at once: disjoint surfaces run concurrently, overlapping ones
  serialize whatever the graph says.
- **Demonstration** is a manual check, not a formality. If you cannot run those
  commands and see that result, the task is not done.
- Acceptance criteria describe observable completion. **Verification** is the
  minimum test evidence expected in the implementing pull request.
- If implementation reveals a contract change, update the relevant design
  document or add/supersede an ADR before merging the changed contract.
- Security-sensitive work requires negative and adversarial tests, not only
  happy paths.

### What every brief assumes, so no brief repeats it

- Dependencies point inward (`docs/code-structure.md` §2). Domain code imports
  no adapter; adapters do not import each other; a `cmd` composition root wires
  concrete implementations.
- Logs, traces, and metrics carry resource IDs, stable codes, and bounded
  enums — never message content, paths, URLs, or secrets
  (`docs/security.md` §5).
- New directories appear only when they contain real implementation.
- Generated output (OpenAPI clients, schemas) is regenerated from its declared
  source of truth in the same change.
- Any new command, configuration key, failure mode, or security posture gets a
  line of operator documentation.

### The standing scenario suite

`test/e2e` holds one scenario suite, introduced by **P1-053**, driving the
deployed system through its real entrypoints. Every task that makes a substrate
real re-runs it unchanged:

```
scenario: create → start → stream → input → stop

  all fake             P1-053
  + local PTY runner   P1-053   same scenarios
  + real store         P1-053   same scenarios
  + Docker sandbox     P3-029   same scenarios
  + second harness     P2-025   same scenarios
  + remote runner      P5-042   same scenarios
```

Run it with `make e2e`, optionally `make e2e PROFILE=<profile>`. It grows only
when a task introduces genuinely new observable behavior.

### The fake profile is a product configuration

The fake harness (P0-008, made controllable by **P1-052**) is a supported way to
run Code Winch, not a test double. It is documented, exercised end to end in CI,
and runnable by hand at any commit. What it does not prove is written down in
P1-052: it is not a provider, has no model latency or rate limits, no
authentication, and emits only the output shapes we chose to write.

### The hands-on surface

`cmd/winch` (**P1-049**, extended by **P1-051**) is the maintained operator CLI.
It is the guaranteed way to drive the system by hand, independent of API churn
and of whatever state the web app is in. Each phase's briefs state which
commands reach the new behavior.

## Delivery waves

The phase gates remain those in `docs/roadmap.md`. Within a phase, the
dependency graph below deliberately exposes parallel work.

### Phase 0 — Contracts and development foundation *(complete)*

| ID | Task | Depends on |
|---|---|---|
| [P0-001](phase-0/p0-001-bootstrap-go-workspace-and-quality-gates.md) | Bootstrap Go workspace and quality gates | None |
| [P0-002](phase-0/p0-002-bootstrap-web-workspace-and-generated-client-checks.md) | Bootstrap web workspace and generated-client checks | None |
| [P0-003](phase-0/p0-003-define-domain-identifiers-and-clocks.md) | Define domain identifiers and clocks | P0-001 |
| [P0-004](phase-0/p0-004-implement-run-and-attempt-state-machine.md) | Implement run and attempt state machine | P0-003 |
| [P0-005](phase-0/p0-005-define-canonical-event-schemas-and-fixtures.md) | Define canonical event schemas and fixtures | P0-003 |
| [P0-006](phase-0/p0-006-define-serializable-runner-protocol-v1.md) | Define serializable runner protocol v1 | P0-003, P0-005 |
| [P0-007](phase-0/p0-007-create-in-memory-application-ports.md) | Create in-memory application ports | P0-003, P0-005, P0-006 |
| [P0-008](phase-0/p0-008-build-fake-harness-and-adapter-contract-kits.md) | Build fake harness and adapter contract kits | P0-005, P0-006, P0-007 |
| [P0-009](phase-0/p0-009-document-threat-model-and-retention-defaults.md) | Document threat model and retention defaults | None |

### Phase 1 — Local single-user vertical slice

P1-010 through P1-020 are complete: they built every component a deployment
needs. Nothing connected them — `cmd/winchd/main.go` was still `func main() {}`
— so no configuration of the system had ever run. **P1-048 through P1-054 close
that gap**, and the phase gate is not met until they land.

| ID | Task | Depends on | Status |
|---|---|---|---|
| [P1-010](phase-1/p1-010-design-initial-openapi-run-contract.md) | Design initial OpenAPI run contract | P0-002, P0-004, P0-005 | complete |
| [P1-011](phase-1/p1-011-implement-postgresql-run-and-event-storage.md) | Implement PostgreSQL run and event storage | P0-004, P0-005, P0-007 | complete |
| [P1-012](phase-1/p1-012-implement-transactional-outbox-delivery.md) | Implement transactional outbox delivery | P1-011 | complete |
| [P1-013](phase-1/p1-013-implement-local-pty-sandbox-driver.md) | Implement local PTY sandbox driver | P0-006, P0-008 | complete |
| [P1-014](phase-1/p1-014-implement-first-harness-adapter.md) | Implement first harness adapter | P0-008, P1-013 | complete |
| [P1-015](phase-1/p1-015-implement-run-supervisor-and-lease-fencing.md) | Implement run supervisor and lease fencing | P0-007, P1-011, P1-012, P1-013, P1-014 | complete |
| [P1-016](phase-1/p1-016-implement-idempotent-input-command-delivery.md) | Implement idempotent input command delivery | P1-012, P1-015 | complete |
| [P1-017](phase-1/p1-017-expose-run-http-api-and-local-authentication.md) | Expose run HTTP API and local authentication | P1-010, P1-015, P1-016 | complete |
| [P1-018](phase-1/p1-018-implement-resumable-websocket-event-stream.md) | Implement resumable WebSocket event stream | P1-012, P1-017 | complete |
| [P1-019](phase-1/p1-019-build-terminal-run-web-slice.md) | Build terminal run web slice | P0-002, P1-017, P1-018 | complete |
| [P1-020](phase-1/p1-020-reconcile-runs-after-daemon-restart.md) | Reconcile runs after daemon restart | P1-015 | complete |
| [P1-048](phase-1/p1-048-boot-the-daemon.md) | Boot the daemon | P1-011, P1-017 | pending |
| [P1-049](phase-1/p1-049-implement-the-local-runner.md) | Implement the local runner | P0-006, P0-008, P1-013, P1-014 | pending |
| [P1-050](phase-1/p1-050-implement-run-use-cases-and-http-binding.md) | Implement run use cases and HTTP binding | P1-015, P1-016, P1-017, P1-048, P1-049 | pending |
| [P1-051](phase-1/p1-051-add-run-commands-to-the-operator-cli.md) | Add run commands to the operator CLI | P1-049, P1-050 | pending |
| [P1-052](phase-1/p1-052-make-the-fake-harness-controllable.md) | Make the fake harness controllable | P0-008 | pending |
| [P1-053](phase-1/p1-053-establish-the-standing-scenario-suite.md) | Establish the standing scenario suite | P1-050, P1-052 | pending |
| [P1-054](phase-1/p1-054-store-credential-references.md) | Store credential references | P1-011, P1-050 | pending |

### Phase 2 — Structured experience and second harness

| ID | Task | Depends on |
|---|---|---|
| [P2-021](phase-2/p2-021-implement-structured-event-normalization.md) | Implement structured event normalization | P1-050, P1-052 |
| [P2-022](phase-2/p2-022-build-conversation-and-activity-renderers.md) | Build conversation and activity renderers | P2-021 |
| [P2-023](phase-2/p2-023-implement-artifact-storage-and-changes-renderer.md) | Implement artifact storage and changes renderer | P1-050, P2-021 |
| [P2-024](phase-2/p2-024-implement-approval-binding-and-policy-evaluation.md) | Implement approval binding and policy evaluation | P2-021 |
| [P2-025](phase-2/p2-025-implement-second-harness-adapter.md) | Implement second harness adapter | P1-049, P2-021 |
| [P2-026](phase-2/p2-026-add-workspace-authorization-and-audit-trail.md) | Add workspace authorization and audit trail | P1-050, P2-055 |
| [P2-027](phase-2/p2-027-implement-retention-export-and-deletion.md) | Implement retention, export, and deletion | P2-023, P2-026 |
| [P2-055](phase-2/p2-055-register-workspaces-and-approved-roots.md) | Register workspaces and approved roots | P1-050 |
| [P2-056](phase-2/p2-056-action-approvals-and-structured-answers.md) | Action approvals and structured answers | P2-022, P2-024 |
| [P2-057](phase-2/p2-057-retry-a-failed-run.md) | Retry a failed run | P1-050 |
| [P2-058](phase-2/p2-058-acquire-harness-credentials.md) | Acquire harness credentials | P1-054, P2-055 |
| [P2-059](phase-2/p2-059-bound-queued-and-concurrent-runs.md) | Bound queued and concurrent runs | P1-050 |

### Phase 3 — Docker isolation

| ID | Task | Depends on |
|---|---|---|
| [P3-028](phase-3/p3-028-implement-disposable-workspace-preparation.md) | Implement disposable workspace preparation | P2-023, P2-055 |
| [P3-029](phase-3/p3-029-implement-docker-sandbox-lifecycle.md) | Implement Docker sandbox lifecycle | P0-008, P1-049, P3-028 |
| [P3-030](phase-3/p3-030-enforce-named-sandbox-profiles.md) | Enforce named sandbox profiles | P1-050, P3-029 |
| [P3-031](phase-3/p3-031-enforce-container-network-policy.md) | Enforce container network policy | P3-029, P3-030 |
| [P3-032](phase-3/p3-032-implement-scoped-credential-injection.md) | Implement scoped credential injection | P1-054, P3-029 |
| [P3-033](phase-3/p3-033-display-and-test-effective-security-posture.md) | Display and test effective security posture | P3-030, P3-031, P3-032 |
| [P3-060](phase-3/p3-060-scan-dependencies-and-publish-an-sbom.md) | Scan dependencies and publish an SBOM | None |

### Phase 4 — Top-level workflows

| ID | Task | Depends on | Status |
|---|---|---|---|
| [P4-034](phase-4/p4-034-define-and-validate-workflow-graph-schemas.md) | Define and validate workflow graph schemas | P0-005, P1-010 | complete |
| [P4-035](phase-4/p4-035-persist-workflow-instances-and-step-leases.md) | Persist workflow instances and step leases | P1-012, P4-034 | complete |
| [P4-036](phase-4/p4-036-implement-workflow-coordinator-replay-loop.md) | Implement workflow coordinator replay loop | P4-035 | pending |
| [P4-037](phase-4/p4-037-implement-run-command-workflow-activities.md) | Implement run command workflow activities | P1-050, P4-036 | pending |
| [P4-038](phase-4/p4-038-implement-control-flow-workflow-activities.md) | Implement control-flow workflow activities | P2-023, P2-024, P4-036 | pending |
| [P4-039](phase-4/p4-039-expose-workflow-http-api.md) | Expose workflow HTTP API | P2-026, P4-037, P4-038 | pending |
| [P4-040](phase-4/p4-040-build-workflow-graph-and-status-ui.md) | Build workflow graph and status UI | P4-039 | pending |

### Phase 5 — Remote runners and hardening

| ID | Task | Depends on |
|---|---|---|
| [P5-041](phase-5/p5-041-implement-authenticated-runner-registration.md) | Implement authenticated runner registration | P1-049 |
| [P5-042](phase-5/p5-042-implement-remote-command-and-event-transport.md) | Implement remote command and event transport | P1-050, P5-041 |
| [P5-043](phase-5/p5-043-implement-distributed-lease-fencing.md) | Implement distributed lease fencing | P5-042 |
| [P5-044](phase-5/p5-044-implement-capability-and-capacity-scheduler.md) | Implement capability and capacity scheduler | P2-059, P5-041 |
| [P5-045](phase-5/p5-045-implement-remote-artifact-handoff.md) | Implement remote artifact handoff | P2-023, P5-042 |
| [P5-046](phase-5/p5-046-add-high-availability-and-recovery-validation.md) | Add high-availability and recovery validation | P4-036, P5-043, P5-045 |
| [P5-047](phase-5/p5-047-define-slos-dashboards-alerts-and-runbooks.md) | Define SLOs, dashboards, alerts, and runbooks | P1-048, P5-044, P5-046 |

## Parallelism

Measured over pending tasks at the 2026-08-14 re-derivation. Critical path is
the longest dependency chain inside the phase; average width is pending tasks
divided by that path.

| Phase | Pending | Critical path | Average width | Collisions |
|---|---|---|---|---|
| 1 | 7 | 3 | 2.3 | none |
| 2 | 12 | 3 | 4.0 | P2-055 / P2-059 on `api/openapi/components/run.yaml` |
| 3 | 7 | 5 | 1.4 | none |
| 4 | 5 | 4 | 1.2 | none |
| 5 | 7 | 5 | 1.4 | none |

Available immediately: **P1-048**, **P1-049**, **P1-052**, **P3-060**, and
**P4-036** — five independent roots.

Phases 3 and 5 are chains, and deliberately so. Phase 3's controls can only be
demonstrated against the container the previous task creates: profiles need a
driver, network policy needs a profile to attach to, and posture display needs
all three resolved. Phase 5 is the same shape — transport needs registration,
fencing needs two hosts that can both reach one run. Phase 4 is layered by the
workflow surface itself; only the graph UI is genuinely serial after the API.

### Keeping the spine from serializing the plan

Four shared surfaces attract every task at once. Each has a remedy, and each
remedy is owned by a task rather than left as advice:

| Surface | Remedy | Owner |
|---|---|---|
| Composition root | Drivers self-register from their own files; the root reads a registry | P1-050 |
| OpenAPI document | Split into per-resource files under `api/openapi/paths/` assembled by `$ref` | P1-050 |
| Run start path | One admission hook with a permissive default | P1-050 |
| CLI and scenarios | One file per command, one file per scenario | P1-049, P1-052 |

Ordered migration numbers are allocated in advance so two open branches cannot
claim the same one: 006 P1-050, 007 P1-054, 008 P2-023, 009 P2-024, 010 P2-055,
011 P2-026, 012 P2-027, 013 P3-030, 014 P5-041. Allocate the next free number in
this table when adding a task, not at implementation time.

The one remaining collision — the shared `Run` schema, which P2-055 and P2-059
both extend — is accepted knowingly. Splitting a single schema object further
would cost more than it saves.

## Phase exit rule

A phase is complete only after every task in that phase is accepted **and** the
corresponding exit statement in `docs/roadmap.md` is demonstrated by an
automated end-to-end or integration scenario in `test/e2e`. A happy-path UI
walkthrough does not close a gate. Later-phase tasks may begin early when all of
their explicit dependencies are complete, but that does not waive an earlier
phase gate.

## Machine-readable task tracking

[`tasks.json`](tasks.json) is the source of truth for task status, ownership,
and dependency availability. Its entries link back to the briefs above and
conform to [`tasks.schema.json`](tasks.schema.json).

Update only these tracking fields as work progresses:

- `status`: `pending`, `in_progress`, `blocked`, or `completed`;
- `owner`: an agent or contributor identifier, or `null` when unassigned; and
- `blocked_reason`: a non-empty explanation for `blocked` tasks, otherwise
  `null`.

Availability is derived rather than stored: a task is **next available** when
its status is `pending` and every ID in `depends_on` has status `completed`.
List them from the repository root with:

```sh
scripts/list-available-tasks.sh
```

Task IDs are append-only. A task inserted logically "between" two others still
takes the next free number, which is why the gap-closure wave is numbered 048
onward rather than renumbering Phase 1.

Mark a task `completed` only after verifying its acceptance criteria against
`HEAD` — not against a pull request description.

## Containerized GitHub runner

The preferred deployment is the repository-scoped self-hosted runner in
`runner/`, which dispatches available tasks to Codex Cloud when a task pull
request merges. See [`runner/README.md`](../../runner/README.md) for
installation, deployment overrides, and how scheduling and the control panel
work.

## Changelog

### 2026-08-14 — re-derivation of Phases 2–5 and a Phase 1 gap-closure wave

Prompted by [`architecture-coverage-review.md`](architecture-coverage-review.md)
and by reading the code at `HEAD` rather than the design documents.

**What the code showed that the review did not.** `cmd/winchd/main.go` was
`func main() {}`. Nineteen tasks were complete, and no configuration of the
system had ever started. The sandbox port had no way to move bytes, so no
harness output could reach a codec; `RunnerGateway` had only an in-memory
recorder; `httpapi.Backend` had no non-test implementation; nothing consumed the
outbox. There was no operator CLI and no end-to-end scenario suite. Phase 1's
exit statement was therefore not demonstrable, though every Phase 1 task was
marked complete.

**Phase 1 reopens** with seven tasks (P1-048 – P1-054) rather than a seventh
phase, because these gaps are Phase 1's own exit condition and because the
tracker schema admits phases 0–5 only. They supersede the five tasks proposed in
[`wiring-plan.md`](wiring-plan.md), whose IDs collided with existing Phase 2
numbers.

**Coverage gaps given owners.** Credential storage (P1-054) and login flows
(P2-058); the workspace aggregate (P2-055); telemetry and configuration
foundation, moved from last in the plan to first (P1-048); the approval round
trip that was rendered but not actionable (P2-056); run retry (P2-057); local
queue admission (P2-059); dependency scanning and SBOM for LB08 (P3-060);
packaging and migration bootstrap (P1-048). Renderer selection folded into
P2-022 and `cmd/winch-runner` into P5-041, both previously unnamed.

**Deferrals with triggers, not silence.** Out-of-process renderer isolation,
interactive harness login, renderer output caching, the communication/API proxy,
and the SQLite developer profile moved to `docs/roadmap.md`'s deferred-decision
table, each with a revisit trigger. P2-027 no longer claims to invalidate a
renderer cache that no task builds.

**Edges dropped after re-checking against the code.** P2-025 → P2-024: proving
the harness port does not require approvals. P5-041 → P3-030: runner
registration does not depend on Docker profile enforcement. P5-044 → P5-043:
placement can be demonstrated without distributed fencing. P1-051 → P1-050 was
kept but narrowed, and the wiring plan's strict five-task chain became three
independent roots by letting tasks depend on the controllable fake rather than
on each other.

**Edges added.** Everything in Phase 2 now depends on P1-050, because nothing in
Phase 2 was observable before the system ran. P5-047 gained P1-048: the
telemetry foundation is a dependency of the operational product, not a synonym
for it.

**Width.** Splitting the OpenAPI document reduced Phase 2's collision set from
thirty pairs to one, turning a nominal average width of 4.0 into a real one.

**Briefs rewritten.** All 25 pending briefs moved from the old template — which
repeated identical "architectural context", "deliverables", and "implementation
notes" sections across every task — to one carrying shape, runtime
reachability, owned surfaces, a runnable demonstration, and deferrals with
owning IDs. Completed tasks' briefs were left untouched as the historical
record of what was built.
