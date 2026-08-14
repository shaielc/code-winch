# P2-022: Build conversation and activity renderers

**Phase:** 2 — Structured experience and second harness
**Shape:** capability
**Dependencies:** P2-021 (semantic: there is nothing to project until typed events exist)

## Objective

A run page shows a conversation and an activity timeline built from canonical
events, and the viewer chooses which projection to see.

## Scope

- Conversation renderer: messages and deltas coalesced by message ID, Markdown
  sanitized, links constrained.
- Activity renderer: tool calls, results, lifecycle, approvals, and usage as
  cards, each showing its stable kind and time.
- Renderer selection by event kind and user preference, persisted per user —
  the `docs/architecture.md` §4 requirement that no task previously owned.
- A declared contract per renderer as `docs/contracts.md` §6 requires: supported
  kinds and schema versions, view-model schema, renderer version, incremental or
  windowed, and fallback behavior.
- Fallback: an unknown kind renders as a bounded, escaped diagnostic card rather
  than disappearing.

## Non-goals

- Answering an approval. The card displays; the round trip is P2-056.
- File changes and diffs — P2-023.
- Caching rendered output. See `docs/roadmap.md` deferred decisions.
- Server-side or out-of-process renderers; these are in-browser projections.

## Runtime reachability

The run page in the web app served by `winchd`, with a view selector.

## Owned surfaces

`web/src/renderers/conversation/`, `web/src/renderers/activity/`,
`web/src/renderers/registry.ts`, `web/src/features/runs/RunPage.tsx`.

## Demonstration

    $ docker compose -f deployments/compose.yml up -d --build
    # create and start a run with the structured-tools scenario, then in the browser:
    → expect: a conversation view with coalesced assistant text, and an activity
      view listing the tool call and its result
    → expect: switching views and reloading keeps the chosen view
    → expect: a scenario emitting an unknown kind shows a diagnostic card, not a
      blank timeline

    # with the malicious-payload scenario selected:
    → expect: no script executes; the CSP report is empty; a crafted link is
      inert; a 5 MB single message does not freeze the tab

## Verification

- Standing scenario suite passes unchanged.
- Renderer unit tests over recorded event fixtures, including out-of-order and
  duplicated deltas.
- Malicious-payload tests: script tags, `javascript:` URLs, ANSI escape abuse,
  oversized and deeply nested payloads.
- Accessibility checks on both views.

## Acceptance criteria

- [ ] Both renderers declare their contract, and the declaration is asserted by
      a test rather than only documented.
- [ ] An unknown kind or schema version degrades visibly and safely.
- [ ] No renderer path writes provider output as raw DOM HTML.
- [ ] Renderer selection persists across reload and is per user, not global.
- [ ] Renderer failure cannot affect run execution.

## Deferrals

| Deferred | Owning task |
|---|---|
| Submitting an approval or structured answer from a card | P2-056 |
| Diff and file-change projection | P2-023 |

## Traces to

`docs/contracts.md` §6; `docs/architecture.md` §4;
`docs/security.md` T12, LB06; `docs/decisions/0002-canonical-events-and-renderers.md`
