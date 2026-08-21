# Workplan coverage review

A point-in-time review of how completely the 47 task briefs in this directory
cover the design documents in `docs/` (`architecture.md`, `code-structure.md`,
`contracts.md`, `security.md`, `roadmap.md`, and the ADRs). It records gaps and
recommendations; it is not itself a task and does not change any contract.

**Reviewed:** 2026-08-14, against 47 tasks (19 completed, 28 pending).

**Resolved:** 2026-08-14 by the plan rewrite recorded in
[`README.md`](README.md#changelog). Every gap below now has an owner. The
review stands as written — it is a point-in-time record, not a living document —
and this table is the only thing added to it.

| Gap | Owner |
|---|---|
| 1. Credential and login lifecycle | P1-054 (reference storage, secret-provider adapter), P2-058 (OAuth/device flow, token entry) |
| 2. Workspace aggregate assumed, never built | P2-055; P1-050 records the implicit-workspace assumption until then |
| 3. Approvals never actioned from the browser | P2-056 |
| 4. Out-of-process renderer isolation | Deferred decision in `docs/roadmap.md` with a revisit trigger |
| 5. Telemetry foundation arrives last | P1-048 (foundation, first available task); P5-047 keeps SLOs, dashboards, runbooks |
| 6. Packaging and deployment unscoped | P1-048 (composition root, config, migration bootstrap, compose); P5-041 names `cmd/winch-runner` |
| 7. Supply chain (T13 / LB08) | P3-060 |
| 8. Run retry modelled but unreachable | P2-057 |
| 9a. Renderer selection by kind and preference | P2-022 |
| 9b. Renderer output caching | Deferred decision; P2-027 no longer claims to invalidate it |
| 9c. Communication/API proxy | Deferred decision |
| 9d. Local queue admission and concurrency limits | P2-059 |
| 9e. SQLite developer profile | Deferred decision |

The two tight edges the review flagged were both dropped: P5-041 no longer
depends on P3-030, and P2-025 no longer depends on P2-024. P1-014 → P1-013 was
left as-is because both tasks are complete.

The review's own blind spot is worth recording: it audited the plan against the
design documents and found 85–90% coverage, but it did not check whether the
composition root existed. `cmd/winchd/main.go` was `func main() {}` at the time
it was written. Coverage of a design is not evidence that anything runs.

## Summary

Coverage of the architecture is high — roughly **85–90%** of documented
behavior has an owning task. The execution spine is covered end to end and in
the right order: canonical events, run state machine, runner protocol, run
supervisor and leases, storage and outbox, streaming, sandbox drivers, harness
adapters, workflows, and remote runners. Every task traces back to a documented
requirement; there is no scope inflation, and the dependency graph has no
unknown references, forward references, or cycles.

The gaps cluster in three places rather than being spread evenly:

1. **Identity and ownership objects** — credentials/login and the Workspace
   aggregate are assumed by many tasks but built by none.
2. **The operational edge** — telemetry foundation, packaging/deployment, and
   supply-chain checks arrive last or not at all.
3. **A few round trips that stop halfway** — approvals are modelled and
   rendered but never actioned from the browser; retry is modelled but never
   exposed.

## Well covered

| Architecture element | Owning tasks |
|---|---|
| Canonical event envelope, families, compatibility fixtures | P0-005, P2-021 |
| Run/attempt state machine and terminal immutability | P0-004 |
| Serializable runner protocol and negotiation (ADR-0001) | P0-006, P5-041, P5-042 |
| One-writer supervisor, lease fencing, redaction, sequencing | P1-015, P1-020, P5-043 |
| Storage, gap-free sequences, transactional outbox | P1-011, P1-012 |
| Resumable streaming with `after_sequence`, gap repair, slow-consumer policy | P1-018, P1-019 |
| Capability-based harness and sandbox ports (ADR-0003) | P0-008, P1-013, P1-014, P2-025, P3-029 |
| Renderers as disposable projections (ADR-0002) | P1-019, P2-022, P2-023 |
| Sandbox hardening, profiles, egress policy, visible posture | P3-028 – P3-033 |
| Workflow definitions, persistence, replay, activities, API, UI | P4-034 – P4-040 |
| Retention, export, deletion by sensitivity class | P2-027 |
| HA, backup/restore, SLOs, runbooks | P5-046, P5-047 |

Security threats T01–T14 all have at least one mapped task, and launch blockers
LB01–LB07, LB09, LB10 have owners in the plan.

## Gaps

### 1. Credential and login lifecycle — largest gap

`architecture.md` §1 lists login as a first-class goal, `security.md` §7 defines
three login patterns, §6 of `architecture.md` defines a `Credential` aggregate,
and `code-structure.md` §1 reserves `internal/adapters/secrets/` for
keychain/file/vault adapters. The string "login" appears in **zero** of the 47
briefs.

No task covers: acquiring a credential (host-mediated OAuth/device flow, or
user-supplied token entry), the credential API and UI, a real secret-provider
adapter behind the port, or interactive harness login as a sensitive structured
exchange. P3-032 covers only *injection at launch*, and it sits in Phase 3 —
while P1-014 (Phase 1) needs `ResolvedCredentials` to launch a real provider,
and `roadmap.md` Phase 1 promises "minimal local authentication and
credential-reference storage", which maps to no task at all.

*Recommendation:* add a Phase 1 task for credential-reference storage plus one
real secret-provider adapter, and a Phase 2/3 task for the login flows.
Re-check T10's mapping (P1-017, P2-026, P3-032) afterwards — none of those
three scopes credential entry today.

### 2. Workspace aggregate is assumed, never built

Workspace is the root policy boundary in the data model, and it is a
precondition for P2-024 (policy intersection), P2-026 (authorization), P2-027
(deletion scope), P3-028 (approved roots), and P3-032 (credential binding).
But no task registers, validates, or persists a workspace: P1-011's scope is
"runs, attempts, commands, and events", and P1-010's contract is runs only.
The code confirms it — `WorkspaceID` exists as an identifier type from P0-003,
and `api/openapi/code-winch.yaml` exposes no workspace resource.

*Recommendation:* add a workspace registration/policy task before P2-026, or
state explicitly that Phase 1 runs against a single implicit workspace and name
the task that removes that assumption.

### 3. Approvals are never actioned from the browser

`contracts.md` §3 lists `structured answer` and `approval` as input payload
kinds, and P1-016 deliberately ships only `text`, `interrupt`,
`terminal_bytes`, and `resize`. No later task picks the other two up. P2-024
covers binding, expiry, and policy evaluation; P2-022 renders approval cards.
The first place a user can actually submit a decision is P4-040, the *workflow*
UI — so a Phase 2 run-level approval is displayed but cannot be answered.

*Recommendation:* name the approval/structured-answer input path explicitly in
P2-024's scope (API + UI), or add a small companion task.

### 4. Out-of-process renderer isolation has no owner

`architecture.md` §4 states that experimental or untrusted server-side
renderers "run out of process with bounded input, time, memory, and no
credentials", and §5 Phase C gives renderer workers a separate security
profile. No task implements this, and it is not listed in `roadmap.md`'s
deferred-decisions table either. Today's built-in renderers are in-browser, so
the practical risk is low — but the claim is currently unbacked.

*Recommendation:* add a deferral row with a revisit trigger, or a task.

### 5. Telemetry foundation arrives as the final task

All 47 briefs carry the acceptance criterion "logs/traces include resource IDs
but exclude content and secrets by default", and `security.md` §5 defines
bounded-label rules in detail. Yet `internal/platform` telemetry, the metric
list in `architecture.md` §7 (queue time, startup time, active runs, dropped
subscribers, parser failures, workflow retries, cleanup failures), and content
exclusion enforcement are only owned by **P5-047, the last task in the plan**.
P1-012 is the sole earlier task with metrics in scope.

*Recommendation:* split a small Phase 0/1 foundation task — structured logger,
redaction helper, metric registry with bounded labels — and leave SLOs,
dashboards, alerts, and runbooks at P5-047. Retrofitting 46 tasks' worth of
ad-hoc logging at the end is the expensive path.

### 6. Packaging and deployment are unscoped

`code-structure.md` §1 reserves `deployments/` for local compose and production
examples, and `architecture.md` §5 Phase A describes a deployable daemon. No
task covers a container image for `winchd`, a compose profile, migration
execution on startup, or configuration file loading (`internal/platform`
config). "compose" appears in zero briefs. P5-047's runbooks presume something
operable exists. `cmd/winch-runner`, the standalone runner composition root, is
also never named in P5-041/P5-042's scopes.

### 7. Supply chain (T13 / LB08) has no real owner

`security.md` maps T13 to "P3-029; shared-deployment launch blocker LB08", but
P3-029's scope stops at pinned image digests and recorded provenance.
Dependency scanning, image scanning, and SBOM production appear in **zero**
briefs ("SBOM" matches nothing).

*Recommendation:* extend P3-029's scope or add a CI-side task adjacent to
P0-001.

### 8. Run retry is modelled but unreachable

P0-004 encodes retry as a linked attempt and P1-011 persists attempts, but no
task exposes retry as an application use case, API operation, or UI control —
"retry" in later briefs always means outbox or workflow retry. The documented
`Failed → Queued` transition cannot currently be triggered by a user.

### 9. Smaller unowned items

- Renderer *selection* by event kind and user preference (`architecture.md` §4)
  — implied by P2-022, never scoped.
- Renderer output caching keyed by `run/event-range + renderer-version`
  (`contracts.md` §6) — P2-027 is required to invalidate a cache no task builds.
- Communication/API proxy (`security.md` §8) — documented as an optional
  adapter capability; neither scoped nor deferred.
- Local queue admission and concurrency limits — the `Queued` state exists from
  Phase 0, but nothing bounds concurrent runs until P5-044 in Phase 5.
- SQLite single-process developer profile (`architecture.md` §5) — offered
  conditionally in the doc, unscoped in the plan. Reasonable to leave.

## Dependency-graph observations

The graph is internally sound (verified: no unknown references, no forward
references, no cycles). Two edges look tighter than the architecture requires:

- **P1-014 → P1-013.** The first harness adapter depends on the local PTY
  sandbox driver, but ADR-0003 makes harness and sandbox independent ports and
  P0-008 already supplies sandbox test doubles. Dropping this edge would let the
  two largest Phase 1 adapter tasks proceed in parallel.
- **P5-041 → P3-030.** The entire remote-runner phase is gated on Docker
  profile enforcement. Defensible for security sequencing, worth confirming it
  is deliberate rather than incidental.

`P2-025 → P2-024` (second adapter after approvals) is also stronger than the
port-validation goal strictly needs, though it is justified if the second
provider's approvals are the point.

## Suggested additions, in priority order

| Priority | Proposed task | Suggested phase |
|---|---|---|
| 1 | Credential reference storage and secret-provider adapter | 1 |
| 2 | Workspace registration, policy, and API | 1–2 |
| 3 | Telemetry and configuration foundation | 0–1 |
| 4 | Approval/structured-answer input path (API + UI) | 2 |
| 5 | Harness login flows (OAuth/device, token entry, interactive) | 2–3 |
| 6 | Packaging, deployment examples, and migration bootstrap | 1–3 |
| 7 | Dependency/image scanning and SBOM (LB08) | 3 |
| 8 | Run retry use case and control | 2 |
| 9 | Out-of-process renderer isolation — task or explicit deferral | 4–5 |
