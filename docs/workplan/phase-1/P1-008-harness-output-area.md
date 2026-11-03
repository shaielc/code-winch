# P1-008: Add the harness output area

**Phase:** 1 — Sandbox exchange
**Shape:** capability
**Dependencies:** None.

## Objective

Add an independently labelled harness-output region to the existing attach page
whose only content is the exact truthful empty-state text **Harness output is not
available in this page.**

## Source baseline

At this brief's derivation revision, `App` renders posture, child status, local
messages, and the composer. It fetches the session once, but it has no
harness-output region or output data source. The Go end-to-end scenario checks
the served static HTML for the page title; it does not render React or prove the
accessible browser DOM.

P1-006 is not implemented at this revision. Its predicted `Running` and
`Exited` states do not affect this task: the output area's identity and empty
state render unconditionally and independently of child status and the session
request.

## Scope

- In `App`, add a section named by a visible `Harness output` heading using
  `aria-labelledby`, consistent with the page's existing regions.
- Render exactly `Harness output is not available in this page.` in that region
  on the initial render and after every success or failure of the existing
  one-shot session request. The sentence describes only this page's current
  presentation capability; it does not say that a child emitted no bytes or
  make a claim about child lifecycle, readiness, input, or exit.
- Give the empty state a restrained presentation consistent with the existing
  page. Keep the region useful at the current responsive page width without
  introducing a terminal, scrolling transcript, records, or output controls.
- Add a focused React semantic test that locates the region by its accessible
  name, asserts the exact empty-state sentence, and rejects false claims about
  emitted/received/read output, readiness, running, or exit. Preserve focused
  tests for posture, child status, local messages, and composer use.
- Preserve the existing one-shot session fetch, child-status identity, posture
  disclosure, local composer behavior, relative static assets, loopback-only
  serving, deployment, and current HTTP contracts.

## Non-goals

- Generating, reading, transporting, polling, streaming, incrementally
  delivering, retaining, ordering, or interpreting harness output.
- Conditional empty-state behavior based on child state, browser timing, an API
  response, or any speculative output model.
- New HTTP operations or fields, records, fake output, provider semantics,
  terminal rendering, persistence, input changes, or broad page restructuring.

## Runtime reachability

The existing `web/src/attach/main.tsx` composition renders the region through
`App`. The production web build remains served by `winch-sandbox` through the
existing static route. A person reaches the rendered React region in the
branch-isolated Docker environment through the proxy URL printed by
`make test-env`.

## Write set

- `web/src/attach/App.tsx`
- `web/src/attach/App.test.tsx`
- `web/src/attach/style.css`

## Contract surfaces

None. This task adds presentation semantics without changing an HTTP, event,
port, persistence, configuration, profile, or registry contract.

## Demonstration

```sh
make test-env
# Open the branch proxy URL printed by make, then inspect the rendered page.
make test-env-down
```

Observe a region labelled `Harness output` containing exactly `Harness output
is not available in this page.` Confirm that child status and posture are still
visible and that submitting a non-empty message through the composer still adds
it under `Local messages`. This is a rendered-page observation; the standing Go
scenario's static HTML inspection is not evidence of the React DOM. Do not
claim that the observation proves whether a child emitted bytes or that any
bytes were read or transported.

## Verification

- Add a focused `App.test.tsx` semantic test for the accessible region, exact
  empty state, unconditional initial rendering, and absence of false output,
  readiness, and lifecycle claims; retain posture, child-status, local-message,
  and composer coverage.
- Run `cd web && npm run format:check && npm run lint && npm run typecheck && npm test && npm run build`.
- Run the repository gates in `docs/testing.md`, including the standing
  end-to-end scenario, without treating its static HTML check as a React
  assertion.

## Acceptance criteria

- [ ] The rendered attach page exposes exactly one region with the accessible
      name `Harness output` and the exact empty state `Harness output is not
      available in this page.` independently of session-fetch and child-status
      state.
- [ ] The region contains no generated or speculative output and makes no claim
      that output was emitted, received, read, transported, or absent, or that
      the child is ready, running, accepting input, or exited.
- [ ] Focused component tests prove the region semantics and preservation of
      child status, posture, local messages, and the usable composer.
- [ ] The branch-isolated deployed-page observation reaches the rendered React
      region and is reported separately from the standing scenario's static
      HTML evidence.
- [ ] No HTTP contract or output-generation, reading, transport, record,
      persistence, input, or provider behavior changes.

## Deferrals

None.

## Traces to

`docs/workplan/phase-1/README.md` increment 4; `docs/roadmap.md` Stage 0;
`docs/contracts.md` §§8–9; ADR-0005/0006.
