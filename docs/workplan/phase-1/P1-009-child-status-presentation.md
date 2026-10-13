# P1-009: Present unknown child status

**Phase:** 1 — Sandbox exchange
**Shape:** capability
**Dependencies:** None. The existing attach page is the complete runtime prerequisite; this task does not consume another planned product change.

## Objective

Make the deployed attach page show an independently identifiable **Child status**
whose exact value is **Unknown**. This is a static, truthful presentation of the
fact that the browser has no authoritative child-state source. It must not imply
that a child exists, is down or running, accepts input, or has exited.

## Scope

- In `web/src/attach/App.tsx`, insert a sibling `<section>` between the page
  header and the posture/local-message content. Give it an accessible region
  name of exactly `Child status` with `aria-labelledby` and a visible heading,
  following the existing named-section pattern used by `Posture` and `Local
  messages`. Put a definition list in that region with a `Status` term and one
  exact `Unknown` value. This makes the region independently queryable while
  retaining explicit label/value semantics; do not use an ARIA live-status role,
  because this constant is not an asynchronously reported child transition.
- Keep the constant directly in the rendered attach-page composition. Place it
  outside the posture conditional so the section renders on the first render and
  stays rendered whether `/api/session` is pending, succeeds, or fails. Do not
  derive it from `posture`, `error`, local messages, or composer state.
- In `web/src/attach/App.test.tsx`, import `within` and add focused semantic
  assertions scoped to `getByRole("region", { name: "Child status" })`. Assert
  the `Status` term and exact `Unknown` value, plus the absence within that region
  of `Down`, `Running`, `Ready`/`readiness`, and `Exited`/`exit` language.
  Exercise the region once while the posture request remains pending and once
  after a rejected request has produced the existing posture alert; these cases
  prove the constant is independent of the fetch rather than merely present on
  its success path. Preserve the existing successful-posture and local-composer
  tests.
- Reuse the existing generic `section`, `dl`, `dt`, and `dd` rules in
  `web/src/attach/style.css`; inspection at HEAD shows they already provide the
  page's panel and label/value treatment. Do not edit CSS or extract a component
  unless implementation uncovers a concrete accessibility defect in this exact
  markup, and report such a write-set expansion before committing it.
- Preserve the normal web build, sandbox image, static-file serving, loopback
  attach URL, posture disclosure, and local composer as the reachability path.

At refinement HEAD, `App.tsx` has only a conditionally rendered `Posture`, a
named `Local messages` section, and `Composer`; there is no child-status source or
presentation. `npm test -- --run src/attach/App.test.tsx` from `web/` passes its
two existing tests, neither of which names child status. The demonstration could
not be observed because `docker --version` and the compose startup both returned
127 (`docker: command not found`); that is an environment limit, not evidence for
or against the predicted rendered behavior.

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

No Go transport, CLI, fake-harness, contract-document, container, or deployment
file is forecast. `web/src/attach/style.css` and a separate child-status component
are explicitly not forecast after inspecting the existing markup and generic
styles.

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
region containing the `Status` label and exact value **Unknown**, while the
posture panel and local composer are still present and usable. In browser
developer tools, block or fail `/api/session`, reload, and observe that the
posture-unavailable alert appears without removing or changing the child-status
region. Restore the request, reload, submit `hello`, and observe the posture panel
and local message while child status remains **Unknown**. Then tear down:

```sh
docker compose -f deployments/compose.yml down
```

This fake-free observation proves presentation and deployed reachability only.
It does not prove that a child was started or queried, that input is ready, or
that an exit status exists. The text-only Go scenario likewise does not execute
the browser UI or prove this rendered value.

## Verification

- `npm test -- --run src/attach/App.test.tsx` from `web/` proves the semantic
  child-status region, its bounded vocabulary, its pending/failed posture-fetch
  independence, and preserved page-level posture/composer behavior.
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

Keep all child-state vocabulary assertions scoped to the named child-status
region so text elsewhere cannot create a false match. Use a never-settling fetch
mock for the pending case and a rejected fetch mock followed by the visible
`Sandbox posture is unavailable.` alert for the failure case. In both cases,
assert `Unknown` synchronously from the region; no timer, polling, or eventual
state transition belongs in this task.

## Acceptance criteria

- [ ] The rendered attach page exposes one independently named child-status
      region with heading/accessible name `Child status`, a `Status` definition
      term, and exact definition value `Unknown`.
- [ ] A semantic component test fails if the region disappears, becomes
      indistinguishable from posture/local messages, or reports `Down`,
      `Running`, `Ready`/readiness, or `Exited`/exit language within the region.
- [ ] The status does not depend on `/api/session`: a never-settling request and
      a rejected request that renders the existing posture-unavailable alert
      both leave the exact `Unknown` region synchronously available.
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
