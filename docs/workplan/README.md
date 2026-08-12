# Implementation workplan

This directory converts the architectural baseline into implementation-sized task briefs. Each task is intended to be implemented in its own pull request unless its brief is explicitly split further.

## How to use this plan

- Treat **Dependencies** as completion prerequisites, not merely suggested reading.
- Tasks with satisfied dependencies may proceed in parallel; never infer a dependency from numeric order alone.
- Each task must preserve the inward dependency rule and add only directories that contain real implementation.
- Acceptance criteria describe observable completion. Required verification is the minimum test evidence expected in the implementing pull request.
- If implementation reveals a contract change, update the relevant design document or add/supersede an ADR before merging the changed contract.
- Security-sensitive work requires negative/adversarial tests, not only happy paths.

## Delivery waves

The phase gates remain those in `docs/roadmap.md`. Within a phase, the dependency graph below deliberately exposes parallel work.

### Phase 0 — Contracts and development foundation

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

| ID | Task | Depends on |
|---|---|---|
| [P1-010](phase-1/p1-010-design-initial-openapi-run-contract.md) | Design initial OpenAPI run contract | P0-002, P0-004, P0-005 |
| [P1-011](phase-1/p1-011-implement-postgresql-run-and-event-storage.md) | Implement PostgreSQL run and event storage | P0-004, P0-005, P0-007 |
| [P1-012](phase-1/p1-012-implement-transactional-outbox-delivery.md) | Implement transactional outbox delivery | P1-011 |
| [P1-013](phase-1/p1-013-implement-local-pty-sandbox-driver.md) | Implement local PTY sandbox driver | P0-006, P0-008 |
| [P1-014](phase-1/p1-014-implement-first-harness-adapter.md) | Implement first harness adapter | P0-008, P1-013 |
| [P1-015](phase-1/p1-015-implement-run-supervisor-and-lease-fencing.md) | Implement run supervisor and lease fencing | P0-007, P1-011, P1-012, P1-013, P1-014 |
| [P1-016](phase-1/p1-016-implement-idempotent-input-command-delivery.md) | Implement idempotent input command delivery | P1-012, P1-015 |
| [P1-017](phase-1/p1-017-expose-run-http-api-and-local-authentication.md) | Expose run HTTP API and local authentication | P1-010, P1-015, P1-016 |
| [P1-018](phase-1/p1-018-implement-resumable-websocket-event-stream.md) | Implement resumable WebSocket event stream | P1-012, P1-017 |
| [P1-019](phase-1/p1-019-build-terminal-run-web-slice.md) | Build terminal run web slice | P0-002, P1-017, P1-018 |
| [P1-020](phase-1/p1-020-reconcile-runs-after-daemon-restart.md) | Reconcile runs after daemon restart | P1-015 |

### Phase 2 — Structured experience and second harness

| ID | Task | Depends on |
|---|---|---|
| [P2-021](phase-2/p2-021-implement-structured-event-normalization.md) | Implement structured event normalization | P0-005, P1-014 |
| [P2-022](phase-2/p2-022-build-conversation-and-activity-renderers.md) | Build conversation and activity renderers | P1-019, P2-021 |
| [P2-023](phase-2/p2-023-implement-artifact-storage-and-changes-renderer.md) | Implement artifact storage and changes renderer | P1-011, P1-017, P2-021 |
| [P2-024](phase-2/p2-024-implement-approval-binding-and-policy-evaluation.md) | Implement approval binding and policy evaluation | P1-016, P2-021 |
| [P2-025](phase-2/p2-025-implement-second-harness-adapter.md) | Implement second harness adapter | P2-021, P2-024 |
| [P2-026](phase-2/p2-026-add-workspace-authorization-and-audit-trail.md) | Add workspace authorization and audit trail | P1-017, P1-018, P2-023 |
| [P2-027](phase-2/p2-027-implement-retention-export-and-deletion.md) | Implement retention, export, and deletion | P2-023, P2-026 |

### Phase 3 — Docker isolation

| ID | Task | Depends on |
|---|---|---|
| [P3-028](phase-3/p3-028-implement-disposable-workspace-preparation.md) | Implement disposable workspace preparation | P2-023 |
| [P3-029](phase-3/p3-029-implement-docker-sandbox-lifecycle.md) | Implement Docker sandbox lifecycle | P0-008, P3-028 |
| [P3-030](phase-3/p3-030-enforce-named-sandbox-profiles.md) | Enforce named sandbox profiles | P1-015, P3-029 |
| [P3-031](phase-3/p3-031-enforce-container-network-policy.md) | Enforce container network policy | P3-029, P3-030 |
| [P3-032](phase-3/p3-032-implement-scoped-credential-injection.md) | Implement scoped credential injection | P2-026, P3-029, P3-030 |
| [P3-033](phase-3/p3-033-display-and-test-effective-security-posture.md) | Display and test effective security posture | P3-030, P3-031, P3-032 |

### Phase 4 — Top-level workflows

| ID | Task | Depends on |
|---|---|---|
| [P4-034](phase-4/p4-034-define-and-validate-workflow-graph-schemas.md) | Define and validate workflow graph schemas | P0-005, P1-010 |
| [P4-035](phase-4/p4-035-persist-workflow-instances-and-step-leases.md) | Persist workflow instances and step leases | P1-012, P4-034 |
| [P4-036](phase-4/p4-036-implement-workflow-coordinator-replay-loop.md) | Implement workflow coordinator replay loop | P4-035 |
| [P4-037](phase-4/p4-037-implement-run-command-workflow-activities.md) | Implement run command workflow activities | P2-024, P4-036 |
| [P4-038](phase-4/p4-038-implement-control-flow-workflow-activities.md) | Implement control-flow workflow activities | P2-023, P4-036 |
| [P4-039](phase-4/p4-039-expose-workflow-http-api.md) | Expose workflow HTTP API | P2-026, P4-037, P4-038 |
| [P4-040](phase-4/p4-040-build-workflow-graph-and-status-ui.md) | Build workflow graph and status UI | P4-039 |

### Phase 5 — Remote runners and hardening

| ID | Task | Depends on |
|---|---|---|
| [P5-041](phase-5/p5-041-implement-authenticated-runner-registration.md) | Implement authenticated runner registration | P0-006, P3-030 |
| [P5-042](phase-5/p5-042-implement-remote-command-and-event-transport.md) | Implement remote command and event transport | P1-015, P5-041 |
| [P5-043](phase-5/p5-043-implement-distributed-lease-fencing.md) | Implement distributed lease fencing | P5-042 |
| [P5-044](phase-5/p5-044-implement-capability-and-capacity-scheduler.md) | Implement capability and capacity scheduler | P5-041, P5-043 |
| [P5-045](phase-5/p5-045-implement-remote-artifact-handoff.md) | Implement remote artifact handoff | P2-023, P5-042 |
| [P5-046](phase-5/p5-046-add-high-availability-and-recovery-validation.md) | Add high-availability and recovery validation | P4-036, P5-043, P5-045 |
| [P5-047](phase-5/p5-047-define-slos-dashboards-alerts-and-runbooks.md) | Define SLOs, dashboards, alerts, and runbooks | P5-044, P5-046 |

## Phase exit rule

A phase is complete only after every task in that phase is accepted **and** the corresponding exit statement in `docs/roadmap.md` is demonstrated by an automated end-to-end or integration scenario. Later-phase tasks may begin early when all of their explicit dependencies are complete, but that does not waive an earlier phase gate.

## Machine-readable task tracking

[`tasks.json`](tasks.json) is the source of truth for task status, ownership, and dependency availability. Its entries link back to the detailed briefs above and conform to [`tasks.schema.json`](tasks.schema.json).

Update only these tracking fields as work progresses:

- `status`: `pending`, `in_progress`, `blocked`, or `completed`;
- `owner`: an agent or contributor identifier, or `null` when unassigned; and
- `blocked_reason`: a non-empty explanation for `blocked` tasks, otherwise `null`.

Availability is derived rather than stored: a task is **next available** when its status is `pending` and every ID in `depends_on` has status `completed`. This avoids a stale duplicated `ready` flag. From the repository root, list all currently available tasks with:

```sh
scripts/list-available-tasks.sh
```

The initial result contains `P0-001`, `P0-002`, and `P0-009`. Once a task is marked `completed`, rerun the query to reveal newly unblocked work. A blocked or in-progress task is never returned even if all its dependencies are complete.

## Containerized GitHub runner

The preferred deployment is the repository-scoped self-hosted runner in
`runner/`, which dispatches available tasks to Codex Cloud when a task pull
request merges. See [`runner/README.md`](../../runner/README.md) for
installation, deployment overrides, and how scheduling and the control panel
work.
