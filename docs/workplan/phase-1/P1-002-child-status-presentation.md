# P1-002: Present unknown child status

**Phase:** 1 — Sandbox exchange
**Shape:** capability
**Dependencies:** None. The existing attach page is the complete runtime prerequisite; this task does not consume another planned product change.

## Objective

Make the deployed attach page show an independently identifiable **Child status**
whose exact value is **Unknown**. This is a static, truthful presentation of the
fact that the browser has no authoritative child-state source. It must not imply
that a child exists, is down or running, accepts input, or has exited.

## Scope

- Add accessible label/value semantics for `Child status: Unknown` to the existing
  React attach-page composition in `web/src/attach/App.tsx`. A named region or a
  definition-list entry is suitable; the status must be distinguishable from the
  sandbox posture and local-message state by both a person and a semantic query.
- Add a focused component assertion in `web/src/attach/App.test.tsx` for the
  independently named child-status region and exact `Unknown` value. Keep the
  current posture-fetch and local-composer checks green.
- Use `web/src/attach/style.css` only if focused visual treatment is needed. A
  small component and matching test under `web/src/attach/` are permitted when
  they clarify label/value semantics, but a backend abstraction for this constant
  is not.
- Preserve the normal web build, sandbox image, static-file serving, loopback
  attach URL, posture disclosure, and local composer as the reachability path.

## Required constraints

`Unknown` is a browser presentation value, not data from `/api/session`. Do not
add a request, endpoint, polling, streaming, process probe, or child launch. The
absence of a launched child does not establish `Down`; neither `Running`, input
readiness, fake availability, nor an exit result may be displayed.

The existing Go end-to-end scenario inspects built HTML and does not execute the
React application. It protects deployment/static serving but is not evidence of
rendered child status. The semantic component test and deployed-page hand check
provide the two distinct forms of evidence required here.

## Non-goals

Child supervision, sandbox-owned status, output or input transport, persistence,
fake-harness behavior, and API or protocol changes are excluded. Do not broadly
restructure the page or change `/healthz` or `/api/session` meanings. Deeper
states belong to later Phase 1 increments and are not completion work deferred by
this presentation-only capability.

## Reachability

The changed React bundle follows the existing web build into the sandbox image
and is served by the existing attach server at `http://127.0.0.1:8080`. No fake
process, provider account, new configuration, or new composition-root wiring is
needed. The repository must remain startable and deployable through its current
compose path.

## Write set

- `web/src/attach/App.tsx`
- `web/src/attach/App.test.tsx`
- `web/src/attach/style.css` (only if focused styling is needed)
- optionally, one focused child-status component and test under `web/src/attach/`

No Go transport, CLI, fake-harness, contract-document, container, or deployment
file is forecast.

## Contract surfaces

Owns the attach page's visible and accessible `Child status: Unknown`
presentation. Consumes the existing static React page, its deployed image path,
and the permanent single-sandbox attach boundary from ADR-0005 and ADR-0006.
It owns no HTTP response shape, authoritative child-state meaning, process
lifecycle, input-readiness, or fake-harness behavior.

## Demonstration

From the repository root, after the automated gates:

```sh
docker compose -f deployments/compose.yml up --build -d
```

Open `http://127.0.0.1:8080`. Observe an independently labeled **Child status**
with the exact value **Unknown**, while the posture panel and local composer are
still present and usable. Then tear down:

```sh
docker compose -f deployments/compose.yml down
```

This fake-free observation proves presentation and deployed reachability only.
It does not prove that a child was started or queried, that input is ready, or
that an exit status exists. The text-only Go scenario likewise does not execute
the browser UI or prove this rendered value.

## Verification

- `npm test -- --run src/attach/App.test.tsx` from `web/` proves the semantic
  child-status region and preserves page-level posture/composer behavior.
- `npm run format:check`, `npm run lint`, `npm run typecheck`, and `npm run build`
  from `web/` cover the forecast frontend surfaces.
- `python3 -m unittest discover -s test` and
  `(cd workplan-control-panel && python3 -m unittest discover -s tests)` remain
  green.
- `make check` is the required host gate. `make test-cycle` exercises the normal
  composed-image scenario; report unavailable Docker/toolchains as limitations,
  never as passing evidence.
- Perform the deployed-page demonstration above. Record separately that the
  existing Go e2e scenario sees built/static HTML rather than the React DOM.

Adversarially assert that the child-status region does not report `Down`,
`Running`, readiness, or an exit result. A posture request that remains pending
or fails must not change the static `Unknown` value, which is not derived from
that response.

## Acceptance criteria

- [ ] The rendered attach page exposes one independently named child-status
      region with label `Child status` and exact value `Unknown`.
- [ ] A semantic component test fails if the region disappears, becomes
      indistinguishable from posture/local messages, or reports `Down`,
      `Running`, readiness, or an exit result.
- [ ] The status does not depend on `/api/session`; pending and failed posture
      fetches leave the truthful `Unknown` presentation available.
- [ ] Existing posture disclosure and local composer behavior remain covered and
      available, and the standard build/image/static-serving route deploys the
      changed page without a started fake.
- [ ] The hand observation records both the exact deployed value and the limits
      of its evidence; no test claims that static HTML inspection executes React.
- [ ] No child process, status API, probing, transport, persistence, or product
      contract is introduced.

## Deferrals

None. This task is complete as a truthful presentation slice; deeper child
observation and ownership are separate decided Phase 1 boundaries rather than
unfinished parts of this capability.

## Traces to

`docs/workplan/phase-1/README.md` increment 1; `docs/roadmap.md` §§1 and 3,
Stage 0; `docs/contracts.md` §§8–9; ADR-0005; ADR-0006.
