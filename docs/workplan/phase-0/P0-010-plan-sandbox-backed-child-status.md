# P0-010: Plan Phase 1 increment 2: sandbox-backed child status

**Phase:** 0 — Plan the implementation phases
**Shape:** planning
**Dependencies:** P0-008 (semantic: this derivation consumes the independently authored child-status presentation boundary in P1-009 and preserves its accessible label/value identity)

## Objective

Produce one complete implementation brief, P1-011, for Phase 1 increment 2,
**Sandbox-backed child status**, and reconcile it with the tracker, workplan
index, and Phase 0/1 coverage. The brief must specify a runnable slice in which
the sandbox reports that it owns no child and the existing child-status
presentation consequently changes from `Unknown` to `Down`, without starting,
probing, or supervising a process.

## Scope

- Use `skills/derive-task/SKILL.md` to create only
  `docs/workplan/phase-1/P1-011-sandbox-backed-child-status.md`. P1-011 is the
  next plan-wide ID after the ten entries currently in `tasks.json`; allocate it
  only together with the complete brief and matching tracker entry.
- Base the implementation instructions on the attach server currently built in
  `internal/adapters/transport/attach/`, composed by
  `cmd/winch-sandbox/main.go`, and consumed by `web/src/attach/App.tsx`. At HEAD,
  `GET /api/session` returns only posture, the composition root passes only that
  posture, and the React application has no child-status markup. Treat the
  `Child status: Unknown` surface in pending P1-009 as a planned prerequisite,
  not observed runtime behavior.
- Make P1-011 depend on P1-009 with a `contract` reason: P1-009 first defines
  the visible and accessible child-status identity, and P1-011 changes the
  meaning and source of that same presentation. The implementation task may be
  derived while its prerequisite is pending, but it is not available until
  P1-009 is completed.
- Specify one minimum session-status transport by extending the existing
  `GET /api/session` response with `childStatus: "Down"`. The sandbox
  composition root must supply this value from its known configuration — it
  starts and owns no child — rather than the browser inferring it from a failed
  request, missing output, health, or posture. Keep the representation closed
  to `Down` in this increment; do not predesign the `Running`/`Exited` source
  needed by increment 3.
- Require the implementation brief to update `docs/contracts.md` §8 with the
  additive response field and its narrow meaning in the same change. Forecast
  the concrete server/session model, composition-root wiring, contract tests,
  React fetch/render tests, and the existing attach-page source needed to
  replace P1-009's constant. Do not introduce a domain/application port for the
  single current case or a second endpoint solely for future lifecycle states.
- Preserve `Child status` as the independently queryable label and change only
  its exact value when a valid sandbox response is received. Specify that a
  pending request remains `Unknown`, and that request failure, non-2xx,
  malformed JSON, a missing field, or any value other than exact `Down` leaves
  the status `Unknown` and exposes an accessible status-unavailable error. Such
  failures must never display `Down`, `Running`, readiness, fake availability,
  or an exit result.
- Require server contract tests for the exact additive response and method/API
  behavior, and component tests for pending, success, transport failure,
  malformed/missing/unknown status values, plus preservation of posture and the
  local composer. Keep the normal image/static-serving route, loopback host
  publication, health response, posture disclosure, and current CLI status
  behavior green.
- Reconcile the P1-011 tracker entry, dependency reason, index row, Phase 0/1
  coverage, graph metrics, and collision analysis. Preserve every existing
  status and leave increments 3–30 without IDs. Record the overlap between
  P1-009 and P1-011 as an ordered contract collision rather than a concurrent
  write warning.

The choices above are deliberately narrow. Reusing `/api/session` keeps one
session snapshot and avoids an otherwise unnecessary request and handler;
passing `Down` from the composition root makes the assertion sandbox-owned
without inventing child supervision; retaining `Unknown` on every untrusted or
failed response prevents absence of evidence from becoming a false lifecycle
claim.

## Required constraints

- Distinguish the observations at HEAD above from every predicted P1-011 result.
- Treat `Down` only as sandbox-owned knowledge that this composition owns no
  child. It is not a process probe, a health result, or evidence about a child
  outside that ownership boundary.
- Require adversarial response-validation tests and prohibit false `Down`,
  `Running`, readiness, fake availability, or exit-result claims.
- Preserve loopback-only serving, `container-standard` posture disclosure,
  local composer behavior, deployment, and existing tests. Document the changed
  `/api/session` contract in the implementation change.
- State that the fake-free `Down` hand observation proves the configured
  sandbox's status reporting and deployed browser transport, but proves neither
  process launch/absence by operating-system inspection nor supervision,
  readiness, input, output, cleanup, or exit reporting.

## Non-goals

This task changes planning artifacts only. P1-011 must not start, inspect,
signal, reap, or supervise a child; alter the fake harness; introduce
`Running`/`Exited` or native exit meaning; transport output or input; add
persistence, polling, or streaming; add provider semantics; or broadly
restructure the page. Those are excluded by later decided Phase 1 increments,
not completion work deferred from increment 2. Do not derive another brief or
reserve another implementation ID.

## Planning outputs and handoff

Create `docs/workplan/phase-1/P1-011-sandbox-backed-child-status.md` and one
matching P1-011 tracker entry, then reconcile `docs/workplan/README.md`,
`docs/workplan/phase-0/README.md`, and `docs/workplan/phase-1/README.md`. The
implementation brief must use the seam shape because it adds the first
sandbox-to-browser child-status contract, and it must forecast at least these
implementation surfaces:

- `internal/adapters/transport/attach/session.go`
- `internal/adapters/transport/attach/server.go`
- `cmd/winch-sandbox/main.go`
- `test/contract/attach/session_test.go`
- `web/src/attach/App.tsx`
- `web/src/attach/App.test.tsx`
- `docs/contracts.md`

Include `web/src/attach/style.css` or a focused child-status component/test only
if P1-009 actually introduces one and P1-011 must edit it. After derivation,
hand P1-011 to a different author for `skills/refine-task/SKILL.md`; the handoff
is workflow, not an artifact or acceptance condition of P0-010.

## Write set

- `docs/workplan/phase-1/P1-011-sandbox-backed-child-status.md`
- `docs/workplan/tasks.json` (one new entry; existing statuses unchanged)
- `docs/workplan/README.md`
- `docs/workplan/phase-0/README.md`
- `docs/workplan/phase-1/README.md`

## Contract surfaces

This planning task changes no product contract. Its P1-011 output must own the
additive `GET /api/session` `childStatus` field and the sandbox-backed `Down`
meaning, while consuming P1-009's accessible `Child status` presentation. It
must own no child lifecycle, process protocol, readiness, output, input, exit,
or provider interpretation.

## Demonstration

Review the new P1-011 brief beside its tracker, index, and Phase 0/1 coverage
records. Confirm all five records use the same real ID and P1-009 dependency,
the graph treats their shared presentation as an ordered contract collision,
and no later implementation ID is present.

Then review P1-011's predicted hand check. It must direct a person to start the
normal compose deployment, open `http://127.0.0.1:8080`, observe the exact
accessible `Child status: Down` while posture and composer remain usable, and
tear the deployment down. It must separately state the fake-free evidence
limits above; the planning demonstration does not claim that this predicted
runtime behavior exists at P0-010 HEAD.

## Verification

- Run `python3 -m unittest discover -s test` and
  `(cd workplan-control-panel && python3 -m unittest discover -s tests)`.
- Validate `docs/workplan/tasks.json` against
  `docs/workplan/tasks.schema.json`, run
  `bash scripts/list-available-tasks.sh`, and use
  `rg -n 'P0-010|P1-011' docs/workplan` to inspect record consistency.
- Run `git diff --check`, `make check`, and `make test-cycle`; report missing Go,
  lint, Docker, or frontend toolchains as limitations rather than passing
  evidence.
- Confirm P1-011 names focused Go contract and React component commands, all
  repository gates for its forecast implementation surfaces, and the exact
  compose-based manual check. Runtime results remain predictions until P1-011
  is implemented.

## Acceptance criteria

- [ ] Exactly one complete Phase 1 brief, P1-011, covers increment 2 and has an
      explicit contract dependency on P1-009.
- [ ] P1-011 defines the additive session response, sandbox-owned `Down`,
      fail-closed `Unknown` behavior, contract documentation, adversarial tests,
      deployment reachability, and evidence limits without absorbing increment
      3.
- [ ] Tracker, index, Phase 0/1 coverage, graph metrics, and collision analysis
      agree, preserve every existing status, and reserve no later ID.
- [ ] P1-011's implementation forecast and verification name every surface and
      gate needed for the described running slice.

## Deferrals

None. Child launch, ownership, cleanup, `Running`, and `Exited` are excluded by
the decided Phase 1 increment boundary, not unfinished completion work in the
sandbox-backed observation increment.

## Traces to

`docs/workplan/phase-1/README.md` increment 2; Phase 0 increment 0.1;
`docs/roadmap.md` §§1 and 3, Stage 0; `docs/contracts.md` §§8–9; ADR-0005;
ADR-0006.
