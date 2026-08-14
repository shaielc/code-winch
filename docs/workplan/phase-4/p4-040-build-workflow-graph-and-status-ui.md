# P4-040: Build workflow graph and status UI

**Phase:** 4 — Top-level workflows
**Shape:** capability
**Dependencies:** P4-039 (compile: the generated workflow client and stream this renders)

## Objective

A user sees a running workflow as a graph, opens any step's run, and cancels any
active branch from the browser.

## Scope

- A graph view of the pinned definition with live step states, updated from the
  workflow stream and reconnecting the same way the run view does.
- Step detail: inputs, outputs, attempts, timing, failure reason, and a link to
  the run a `run.*` step created.
- Run lineage in the other direction: a run started by a workflow links back to
  its instance and step.
- Cancel controls for the instance and for any active branch, disabled with a
  stated reason when the state does not permit them.
- An instance list with status, definition version, and duration.
- Bounded rendering for a large graph, with the same fallback discipline the
  event renderers use.

## Non-goals

- Editing or authoring definitions in the browser.
- Live-editing a running instance.

## Runtime reachability

The workflow section of the web app served by `winchd`.

## Owned surfaces

`web/src/features/workflows/`, `web/src/renderers/graph/`,
`web/src/app/App.tsx` (routes).

## Demonstration

    # start a foreach workflow, then in the browser:
    → expect: the graph shows parallel branches advancing live, without a reload

    # click a run.start step:
    → expect: its run opens, and the run page links back to the instance

    # cancel one branch:
    → expect: that branch stops, siblings continue, and the graph reflects it
      within the stream's latency

    # drop the network for ten seconds:
    → expect: the view reconnects and shows the transitions it missed, in order

    # open a 200-step instance:
    → expect: bounded rendering with an explicit indication, not a frozen tab

## Verification

- Standing scenario suite passes; the web workspace's own tests cover the views.
- Component tests over recorded instance fixtures, including failed, cancelled,
  and compensating branches.
- Reconnect test asserting no missed or duplicated transition.
- Accessibility checks on the graph and step detail.

## Acceptance criteria

- [ ] Every active branch is inspectable and cancellable from the browser.
- [ ] Lineage links work in both directions.
- [ ] Reconnect loses no transition.
- [ ] Phase 4's exit statement in `docs/roadmap.md` is demonstrated by the
      standing scenarios plus this view.
- [ ] No provider or step-type conditional leaks into a generic component.

## Deferrals

`None.`

## Traces to

`docs/architecture.md` §4 (web application); `docs/contracts.md` §6;
`docs/roadmap.md` Phase 4 exit
