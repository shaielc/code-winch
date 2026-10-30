# P1-008: Present the harness output area

**Phase:** 1 — Sandbox exchange
**Shape:** capability
**Dependencies:** None (the existing attach page is sufficient; this distinct
presentation does not consume the child-status contract or behavior)

## Objective

Add an independently identifiable **Harness output** region to the attach page
with the exact empty-state text **Harness output is not connected in this
view.** The region identifies where later output can be presented while making
no claim that a child exists, is ready or running, accepts input, has exited,
or has emitted any output.

## Source baseline

At derivation, `App.tsx` renders posture, an accessible Child status region,
browser-local messages, and the composer. It has no harness-output region.
The session request is one-shot and accepts only `Down` as a supported child
status at this revision; P1-006 predicts later child ownership and additional
status values but is not required by this independent presentation change.
The attach page has shared section and empty-state styling, and its component
tests already locate the Child status and Local messages regions by accessible
name. The standing Go scenario inspects served static HTML and therefore does
not establish React DOM semantics.

## Scope

- Render a `section` whose accessible name is supplied by a visible level-two
  heading with the exact text `Harness output`.
- Always render the exact empty-state sentence `Harness output is not connected
  in this view.` inside that region. Its presence must not depend on the
  session request, child status, posture, browser timing, or local messages.
- Use the existing presentation vocabulary and bounded page layout. Reuse the
  existing empty-state treatment where appropriate rather than introducing a
  terminal renderer, record shape, output item, or interactive control.
- Preserve the one-shot session fetch, Child status identity and values present
  at implementation time, posture disclosure, browser-local message behavior,
  usable composer, relative static routing, loopback-only publication,
  deployment, and existing tests.

This task changes only the rendered attach-page presentation. It adds no
request, response member, route, polling, stream, or other HTTP behavior.

## Required constraints

The empty state describes the view's missing connection, not the child or its
output. Tests and copy must not say or imply `no output`, `waiting`, `ready`,
`running`, `connected`, `listening`, `complete`, `exited`, or that bytes were
read. The region remains present and unchanged while the session request is
pending, succeeds, or fails.

Do not infer behavior promised by P1-006. If its child-status changes are
present when this task is implemented, preserve them without making the output
region conditional on them.

## Non-goals

Generating, reading, transporting, polling, streaming, buffering, decoding,
rendering, retaining, or ordering child output; output records or fake output;
child lifecycle or readiness; input delivery; persistence; provider semantics;
terminal behavior; and broad page restructuring are excluded.

## Runtime reachability

The existing React composition root renders the region on every attach-page
load. A person observes it at the branch-isolated proxy URL produced by
`make test-env`; no child behavior, provider account, API change, or additional
service is required.

## Write set

- `web/src/attach/App.tsx`
- `web/src/attach/App.test.tsx`
- `web/src/attach/style.css` only if the existing section and empty-state rules
  cannot provide the bounded presentation without a focused addition

No Go, transport, fake-harness, deployment, or design-contract file is
forecast.

## Contract surfaces

Owns the attach page's accessible `Harness output` region identity and exact
empty-state text `Harness output is not connected in this view.` It consumes
only the existing attach-page composition and shared visual treatment.

It does not own or change `GET /api/session`, child status, output transport or
records, browser-local messages, input, persistence, or provider meaning.

## Demonstration

After automated gates, run `make test-env`, obtain the branch project name as
documented in `deployments/README.md`, and open
`http://127.0.0.1:7080/<project>/sandbox/`. In the rendered page, observe an
independently labeled **Harness output** region containing exactly **Harness
output is not connected in this view.** Also observe the existing Child status,
`container-standard` posture and `network-egress` disclosure, submit a local
message through the composer, and confirm those remain usable. Run
`make test-env-down` afterward.

This rendered-page observation establishes presentation and reachability only.
It does not establish that a child emitted no bytes, or that bytes were read,
transported, or retained. Static HTML inspection by the standing Go scenario
is not a substitute for the rendered React observation.

## Verification

The gates in `docs/testing.md` apply; report which Go and web options ran.
Checks specific to this task:

- `cd web && npm test -- --run src/attach/App.test.tsx` includes a focused
  semantic assertion that the region is found by role and exact accessible name,
  contains the exact empty state, and contains no false output, connection,
  readiness, or lifecycle claim.
- The focused React tests show the same region and text while the session fetch
  is pending and after request success and failure, while retaining assertions
  for Child status, posture, and the usable local composer.
- Run the full web format, lint, typecheck, test, and build gate, the repository
  Go gate including the standing e2e scenario, and the branch-isolated deployed
  page demonstration above.

Component tests establish React semantics; the standing e2e scenario continues
to establish static serving, not a rendered DOM assertion.

## Acceptance criteria

- [ ] Every attach-page render includes a region named exactly `Harness output`
      by its visible heading and the exact sentence `Harness output is not
      connected in this view.`
- [ ] The region and empty state are independent of session fetch state, child
      status, posture, timing, and local messages, and make no output,
      connection, readiness, input, or lifecycle claim beyond the view's
      missing connection.
- [ ] Focused semantic component tests cover the region, exact copy, prohibited
      implications, pending/success/failure independence, and preservation of
      Child status, posture, and composer behavior.
- [ ] No HTTP contract, child behavior, fake output, transport, record,
      persistence, input, provider, or terminal surface changes.
- [ ] The web and repository gates remain green, and the deployed-page hand
      observation distinguishes rendered React evidence from static HTML
      inspection and records its evidence limits.

## Deferrals

None. Live output is excluded by the decided Phase 1 increment 5 boundary, not
unfinished work needed for this truthful presentation capability.

## Traces to

`docs/workplan/phase-1/README.md` increment 4; `docs/roadmap.md` §§1 and 3,
Stage 0; `docs/contracts.md` §§8–9; ADR-0005; ADR-0006.
