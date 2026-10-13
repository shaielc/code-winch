# P0-008: Plan Phase 1 increment 1: child status presentation

**Phase:** 0 — Plan the implementation phases
**Shape:** planning
**Dependencies:** P0-001

## Objective

Produce one complete, independently authored implementation brief for Phase 1
increment 1, **Child status presentation**, and reconcile that brief with the
tracker, workplan index, Phase 0 coverage, and Phase 1 coverage. The derived
task must make the existing deployed attach page identify the child state as
`Unknown` without claiming that a child exists, is running, or accepts input.

## Scope

Use `skills/derive-task/SKILL.md` against the reviewed
`docs/workplan/phase-1/README.md` and the code at HEAD. Allocate the next
plan-wide ID only when the implementation brief and its tracker record are
written. Derive exactly increment 1; do not combine its presentation-only loop
with increment 2's sandbox-owned observation or increment 3's child ownership.

The task author must ground the implementation brief in these current facts:

- `web/src/attach/App.tsx` is the existing attach-page composition point. It
  renders posture, local messages, and the composer, but no child-status region.
  `web/src/attach/App.test.tsx` is the existing page-level component suite.
- `web/src/attach/style.css` owns the attach page's current visual treatment.
  A derived write forecast may add a focused presentation component and test
  under `web/src/attach/` if that makes the label/value semantics clearer, but
  it must not create a backend abstraction for a browser constant.
- `internal/adapters/transport/attach/server.go` currently exposes only health,
  session posture, and static content. There is no child-status endpoint, and
  `cmd/winch-sandbox/main.go` does not start the shipped fake harness.
- `test/e2e/scenario_sandbox_starts_test.go` verifies the served shell by
  inspecting built HTML; it does not execute the React application or inspect
  its DOM. The derived brief must therefore require a component-level semantic
  assertion for an independently identifiable child-status region and the
  exact `Unknown` value, plus the deployed-page hand check. It must not pretend
  the existing Go scenario proves rendered child status.

The implementation brief should instruct the implementer to render static,
truthful presentation in the React attach page, preserve the existing posture
and local-composer behavior, and keep the normal image/build path as the route
by which the changed page reaches the sandbox. Require accessible label/value
semantics (for example, a named region or definition-list entry) so tests and a
person can distinguish the child state from posture and local-message status.
The exact component split and CSS selectors remain implementation choices.

If inspection no longer finds the existing attach page, or finds an
authoritative child-status source already wired into it, stop and return the
boundary to the phase author under `skills/revise-phase/SKILL.md`; do not make
this planning task absorb sandbox status transport or child supervision.

## Required constraints

- `Unknown` is a presentation value, not a response obtained from
  `/api/session` and not a new API contract. It conveys only that authoritative
  child state is unavailable.
- Do not infer `Down` from the absence of a launched process, and do not display
  `Running`, input readiness, fake availability, or an exit result.
- Preserve the loopback attach surface, posture disclosure, SPA/static serving,
  local composer, container build, and existing tests. No live provider account
  or started fake process is required.
- The implementation brief must separate observed HEAD behavior from predicted
  post-implementation results and state that a fake-free `Unknown` display
  proves presentation and reachability only.

## Non-goals

This planning task does not implement Phase 1 product behavior. Its derived
increment does not add a child supervisor, process probing, a status API,
polling or streaming, output transport, input delivery, persistence,
harness-specific capability semantics, vendor interpretation, or broad page
restructuring. It does not change the meanings of `/healthz` or `/api/session`.

Do not derive a second Phase 1 implementation brief in this task. Increment 2,
**Sandbox-backed child status**, receives a new, separately authored Phase 0
planning brief and tracker entry as part of this task's planning handoff; that
planning entry must not reserve or invent increment 2's implementation ID.

## Planning outputs and handoff

Perform the assigned operation in this order so the records cannot describe an
artifact that does not exist:

1. Derive `docs/workplan/phase-1/<allocated-ID>-child-status-presentation.md`
   with a matching tracker entry. Its shape should reflect a small
   operator-visible presentation capability, and its dependencies must name
   only behavior it actually consumes; a neighboring increment number is not
   automatically a dependency.
2. Reconcile `docs/workplan/README.md`, `docs/workplan/phase-0/README.md`, and
   `docs/workplan/phase-1/README.md` with the real ID, dependency reason,
   coverage state, graph metrics, and concurrent write/contract collision
   analysis. Preserve every pre-existing tracker status.
3. Create the next concrete Phase 0 planning task for increment 2,
   **Sandbox-backed child status**, including its independently authored brief,
   tracker entry, index/dependency reason, and Phase 0/1 coverage records. Its
   target operation is future implementation-brief derivation, not product
   implementation, and it allocates no implementation ID.
4. Hand the finished increment-1 implementation brief to a different refiner
   using `skills/refine-task/SKILL.md`. The author of this planning task must
   not also perform that refinement.

## Write set

Forecast for the assigned planning operation:

- `docs/workplan/phase-1/<allocated-ID>-child-status-presentation.md`;
- `docs/workplan/phase-0/<next-planning-ID>-plan-sandbox-backed-child-status.md`;
- `docs/workplan/tasks.json` (the two matching new entries, with all existing
  statuses unchanged);
- `docs/workplan/README.md`;
- `docs/workplan/phase-0/README.md`; and
- `docs/workplan/phase-1/README.md`.

The derived implementation brief should forecast the existing React attach
surface and focused tests (`web/src/attach/App.tsx`,
`web/src/attach/App.test.tsx`, and, if needed, `web/src/attach/style.css` or a
focused component/test pair). It should not forecast Go transport, fake-harness,
CLI, or contract-document changes unless a source-supported boundary finding is
first returned for phase revision. No later implementation IDs are reserved.

## Contract surfaces

This Phase 0 task changes no product contract. The derived implementation brief
owns only the attach page's visible and accessible **Child status: Unknown**
presentation. It consumes the existing static React page, its deployed image
path, and the permanent single-sandbox attach-surface boundary from ADR-0005
and ADR-0006. It does not own HTTP response shapes, child-state semantics,
process lifecycle, input readiness, or fake-harness behavior.

Record shared writes as collision warnings rather than inventing dependency
edges. If another available task would redefine the same child-status
presentation, report a contract collision and reconcile the graph before the
new implementation task is assignable.

## Demonstration

Review the planning artifacts rather than claiming that the UI has already
changed:

1. Open the derived implementation brief and confirm its predicted hand check
   uses `docker compose -f deployments/compose.yml up --build -d`, then opens
   `http://127.0.0.1:8080` and observes an independently labeled child status
   whose exact value is `Unknown` while the posture panel and local composer
   remain available. Teardown uses
   `docker compose -f deployments/compose.yml down`.
2. Confirm the brief explicitly says this observation does not prove a child
   was started, queried, or ready for input, and does not claim that the current
   text-only Go e2e check executes the browser UI.
3. Confirm the implementation brief, its tracker entry, the workplan index, and
   Phase 0/1 coverage all use the same allocated ID and dependency set.
4. Confirm the separately authored Phase 0 increment-2 planning brief and its
   tracker/coverage records exist, while no increment-2 implementation ID does.

## Verification

The planning author must run and report:

```sh
python3 -m unittest discover -s test
(cd workplan-control-panel && python3 -m unittest discover -s tests)
python3 -m json.tool docs/workplan/tasks.json >/dev/null
bash scripts/list-available-tasks.sh
rg -n '<allocated-ID>|<next-planning-ID>' docs/workplan
git diff --check
make check
make test-cycle
```

Replace the angle-bracket placeholders in the search with the allocated IDs.
Inspect the search results to prove brief existence, dependency references, and
index/coverage agreement; compare the tracker diff to prove existing status
values are unchanged. `make check` is the minimum host gate. Run
`make test-cycle` for the composed-image scenario and report an unavailable
Docker/toolchain as a verification limitation rather than as passing evidence.
The target implementation brief must additionally require the focused web test
and web format/lint/typecheck/build gates for the files it forecasts.

## Acceptance criteria

- [ ] Exactly one complete Phase 1 implementation brief covers increment 1 and
      has one matching real tracker ID; its objective, failures, tests, and hand
      observation all require an accessible `Child status` presentation with
      exact value `Unknown` on the deployed attach page.
- [ ] The brief cites the current React composition/test surfaces, predicts the
      concrete write set, preserves posture/composer/deployment behavior, and
      requires semantic component coverage rather than treating static HTML
      inspection as rendered-DOM evidence.
- [ ] The brief explicitly rejects a status endpoint, child probing/spawning,
      inferred `Down` or `Running`, and input-readiness claims; its demonstration
      states what the static, fake-free value does not prove.
- [ ] Owned and consumed contract surfaces, dependency reasons, expected
      observation, inherited gates, automated/adversarial checks, and all
      verification limitations are explicit and source-credible.
- [ ] Tracker, index, Phase 0/1 coverage, graph metrics, and collision warnings
      agree, and the diff proves all pre-existing tracker statuses are unchanged.
- [ ] A different refiner receives the increment-1 brief, and a separately
      authored, fully registered Phase 0 planning brief assigns increment 2
      without allocating a speculative implementation ID.

## Deferrals

None. Deeper child observation and ownership are excluded product boundaries in
the reviewed Phase 1 design, not completion work deferred by this task or by the
implementation brief it produces. This task creates the required tracked
planning handoff for increment 2; it must not cite increment labels or
hypothetical IDs as owners.

## Traces to

`docs/workplan/phase-1/README.md` increment 1 and concern boundaries; Phase 0
increment 0.1; `docs/roadmap.md` §§1 and 3, Stage 0;
`docs/contracts.md` §§8–9; ADR-0005/0006; and
`skills/shared/workplan-model.md` invariants I1–I7.
