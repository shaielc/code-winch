# P1-005: Plan Phase 1 increment 3: actual child status

**Phase:** 1 — Sandbox exchange
**Shape:** planning
**Dependencies:** P1-003 (semantic: the assigned brief must consume the independently derived sandbox session `childStatus` contract and its `Down` presentation, which P1-003 produced in P1-004, and extend rather than redefine it)

## Objective

Produce one complete, independently authored implementation brief for Phase 1
increment 3, **Actual child status**, and reconcile it with the tracker,
workplan index, and Phase 0/1 coverage. The derived capability must make the
sandbox launch and own the shipped fake, and make the attach page report
`Running` while that child runs and `Exited` once it ends, from sandbox-owned
observation.

## Scope

Use `skills/derive-task/SKILL.md` against the reviewed Phase 1 design, the
completed P1-003 planning output (the P1-004 brief), and code at HEAD. Derive
exactly increment 3 as P1-006. The source-observed baseline is that
`cmd/winch-sandbox/main.go` constructs only the attach HTTP handler and server,
`internal/adapters/transport/attach.Server` stores only posture, the React app
fetches the session once, the image contains `/usr/local/bin/fake-harness`, and
the fake blocks on stdin unless one of its existing exit controls is selected.
P1-004's `Down` response, response validation, and presentation are predictions
from its pending brief, not behavior at HEAD.

The P1-006 brief must prescribe one sandbox-owned lifecycle in the existing
`cmd/winch-sandbox` composition: start the shipped fake before advertising the
HTTP service, retain the exact process handle, wait/reap it, publish `Running`
only after a successful start and `Exited` only after that owned handle ends,
and terminate and reap it during sandbox shutdown. Keep process ownership at
the composition boundary rather than introducing a generic harness adapter
before Phase 2. Extend the P1-004 session value through the existing attach
server and one-shot React fetch; a browser reload after the controlled fake
exit is sufficient for this increment and polling or streaming belongs outside
it. The derived brief must choose and document the smallest deterministic fake
exit control and a content-free, nonzero sandbox-start failure when the fake
cannot be spawned, with focused tests that do not scan the host process table.

This task is pipelined: it derives increment 3 while P1-004 is still being
implemented. Plan against P1-004's brief as the consumed contract, label
behavior it predicts as a prediction, and record the derived brief's dependency
on P1-004's implementation as a graph edge rather than waiting for it here. Do
not combine increment 4–6 output display or transport, any input behavior, or
vendor-native exit interpretation.

Allocate P1-006 only when its complete brief and matching tracker entry exist.
Give it a `contract` dependency on P1-004 because it extends P1-004's exact
`childStatus` wire and presentation contract. Reconcile dependency reasons,
coverage, graph metrics, and collision analysis in the index and both phase
records. Then complete `skills/phase-increment-handoff/SKILL.md` in this same
change by creating P1-007, the planning brief for increment 4, **Harness output
area**; P1-007 consumes only the Phase 1 design and therefore depends on P0-001,
not on P1-005, P1-006, or another implementation task.

## Required constraints

- Distinguish source-observed behavior from behavior predicted by P1-004's
  pending implementation and by the derived brief.
- Treat `Running` and `Exited` as sandbox-owned knowledge of a child the sandbox
  launched and owns, not a browser inference from output, request timing, or
  process-table scans.
- `Exited` reports only that the owned child ended. Exit codes, signals, or
  success/failure meaning are vendor-native interpretation (Phase 2). D5 owns
  the later pipes-versus-PTY decision, not exit semantics. P1-006 may retain a
  raw `os.ProcessState` internally for cleanup but must not expose or interpret
  it in the session contract.
- Require minimum cleanup: no orphaned child after sandbox shutdown, and
  observable content-free, nonzero startup failure when spawn fails. Require
  tests for start failure, start-before-serve ordering, wait/reap, termination
  on both SIGINT and SIGTERM, repeated/concurrent status reads, and invalid wire
  values. Prohibit false health, readiness, input-availability, or successful
  child-start claims.
- Fake controls are chosen for this demonstration (for example, a way to make
  the fake exit) and may change the fake's current behavior; they do not define
  product configuration.
- Preserve loopback-only serving, posture disclosure, the local composer,
  deployment, the `Unknown` fallback and `Down` state, and existing tests. Any
  change to the `childStatus` wire contract is documented in the same change.
  Preserve the existing five-second HTTP shutdown bound; the brief must state
  and test a bounded child-stop policy rather than allowing shutdown to wait
  forever.

## Non-goals

This task writes planning artifacts only. Its target increment excludes reading
or presenting child output, input delivery, persistence, restart or retry
policy, provider semantics, and broad page restructuring. It does not derive a
second implementation brief or reserve implementation IDs for later increments.

## Planning outputs and handoff

Create `docs/workplan/phase-1/P1-006-actual-child-status.md` and its matching
tracker entry, then reconcile `docs/workplan/README.md`,
`docs/workplan/phase-0/README.md`, and `docs/workplan/phase-1/README.md`. Record
the extended `childStatus` values, the owning spawn/cleanup surface, and the
documentation/write forecast. Preserve every pre-existing status.

Handoff: create the complete planning-shaped
`docs/workplan/phase-1/P1-007-plan-harness-output-area.md` plus its pending
tracker/index/coverage records exactly as
`skills/phase-increment-handoff/SKILL.md` requires. A different refiner then
takes P1-006 with `skills/refine-task/SKILL.md`.

## Write set

- `docs/workplan/phase-1/P1-006-actual-child-status.md`
- `docs/workplan/phase-1/P1-007-plan-harness-output-area.md`
- `docs/workplan/tasks.json` (new entries only; existing statuses unchanged)
- `docs/workplan/README.md`
- `docs/workplan/phase-0/README.md`
- `docs/workplan/phase-1/README.md`

## Contract surfaces

This planning task changes no product contract. Its derived brief will extend
P1-004's `GET /api/session` `childStatus` member with `Running` and `Exited` and
own the sandbox's child launch and ownership. That is a contract collision with
P1-004, which defines `childStatus`. The derived brief therefore takes a
`contract` edge on P1-004. It must not own output, input, readiness, or provider
interpretation.

## Demonstration

Review the derived brief and matching tracker/index/coverage records. Confirm
that its predicted compose-based hand check observes `Running` after the
sandbox starts the fake, then uses the chosen deterministic fake control and a
page reload (not output, timing inference, polling, or a process scan) to
observe `Exited`. It must separately demonstrate a spawn failure and sandbox
shutdown with no owned child left running, preserve P1-004's `Down`, posture,
composer, and loopback behavior, and claim no output or input readiness. These
are predicted P1-006 checks; this planning task demonstrates only that the
brief and records require them. Confirm all records use P1-006 with the same
dependency set, no later implementation ID is reserved, and P1-007 exists with
its complete brief, pending tracker entry, and matching coverage.

## Verification

Run completion-script and control-panel tests, tracker JSON validation,
`scripts/list-available-tasks.sh`, an `rg` consistency search for both P1-005
and P1-006/P1-007, `git diff --check`, `make check`, and
`make test-cycle`. Report unavailable Docker/toolchains as limitations. The
derived brief must additionally name focused `cmd/winch-sandbox` lifecycle
tests, attach server-contract tests, React session-validation/presentation
tests, the standing e2e scenario, and the Go/web/build gates for every forecast
implementation surface. Its manual commands must use the per-branch
`make test-env` / proxy URL / `make test-env-down` path from
`deployments/README.md`; they must not replace or tear down the operator's
default `code-winch` deployment.

## Acceptance criteria

- [ ] Exactly one complete Phase 1 brief, P1-006, covers increment 3 with a
      `contract` dependency on P1-004 and labels P1-004 behavior as predicted.
- [ ] The brief defines sandbox-owned `Running`/`Exited`, spawn-failure and
      bounded cleanup/reaping behavior, deterministic fake controls, contract
      documentation, lifecycle/contract/frontend/e2e tests, per-branch
      deployment reachability, and evidence limits without absorbing
      increments 4–6.
- [ ] The brief makes start failure content-free and nonzero, never serves false
      child health, validates all three wire values while retaining `Unknown`
      for missing/invalid values, and exposes no exit result, readiness, or
      input-availability meaning.
- [ ] Tracker, index, Phase 0/1 coverage, graph metrics, and collision warnings
      agree and preserve all existing statuses.
- [ ] No speculative later implementation ID exists.
- [ ] Every `phase-increment-handoff` action holds: P1-007 is the sole complete
      derivation task for increment 4, depends only on P0-001, and has matching
      tracker, index, and Phase 0/1 coverage records.

## Deferrals

None. Output, input, and vendor-native exit interpretation are excluded by the
decided Phase 1 and roadmap boundaries, not unfinished completion work in the
actual-child-status increment.

## Traces to

`docs/workplan/phase-1/README.md` increment 3; Phase 0 increment Plan Phase 1;
`docs/roadmap.md` Stage 0 and §6 D5; `docs/contracts.md` §§8–9; ADR-0005/0006;
`skills/shared/workplan-model.md` planning shape and I1–I7.
