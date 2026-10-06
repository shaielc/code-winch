# P0-004: The page has an input control that echoes what you type

**Phase:** 0 — A sandbox you can talk to
**Shape:** capability
**Dependencies:** P0-001 (semantic: there is no page to put a control in until P0-001
serves one)

## Objective

Type a line into the page and press send, and the line appears in the page's own list —
nothing leaves the browser, and the composer is ready for the next line.

## Scope

- `web/src/attach/Composer.tsx` — a text input and a send action. Enter submits, Shift+Enter
  newlines, an empty or whitespace-only value is not submittable, and the control clears
  after a send.
- The sent line is appended to the page's list, visually distinguished from a record the
  harness produced, and marked as local — it has no ordinal, because the runner has not
  seen it.
- The control reports when a submission cannot be made at all rather than appearing to
  succeed, so the state it will acquire in P0-005 has somewhere to be shown.
- A web component test for each of those behaviours.

## Non-goals

- Any HTTP request, any server contract, and any change to `internal/`. This task
  declares no contract surface at all; that is what keeps it concurrent with the whole
  output path.
- Persisting the submission, giving it an ordinal, accepting or refusing it. P0-005.
- Delivering it to the harness. P0-006.
- An approval control, a structured answer, terminal bytes, resize, and interrupt. The
  typed payloads `docs/contracts.md` §3 defines beyond text arrive with the capability
  that needs them; Stage 0's hand check is a line of text.

## Runtime reachability

The composer renders inside `web/src/attach/App.tsx`, which `cmd/winch-sandbox` serves
from `WINCH_SANDBOX_STATIC_DIR`. `docker compose -f deployments/compose.yml up --build`
and a browser on the published loopback port reach it.

This is the one task in the phase whose behaviour the standing scenario suite cannot
cover: the suite drives HTTP and WebSocket, not the DOM. Its automated coverage is the
web component test, and its hands-on check is the browser. `winch` gains no command here
because there is no server capability to drive — the next task, which adds one, adds
`winch input` with it.

## Write set

- `web/src/attach/Composer.tsx`, `web/src/attach/Composer.test.tsx`
- `web/src/attach/App.tsx`
- `web/package.json` (component-test dependencies, if the P0-001 scripts need them)

## Contract surfaces

None.

## Demonstration

    $ docker compose -f deployments/compose.yml up --build -d
    $ xdg-open http://127.0.0.1:8080

    → type "hello" and press Enter
    → expect: "hello" appears in the list, marked as local and carrying no ordinal, and
      the input is empty and focused

    → press Enter on an empty input
    → expect: nothing is added and the control says why, rather than silently ignoring it

    → reload the page
    → expect: "hello" is gone. It never left the browser, which is exactly what this
      task claims and the next task changes.

## Verification

- `cd web && npm test` — component tests for submit-on-Enter, newline on Shift+Enter,
  rejection of an empty value, clearing after send, and the local marking.
- `cd web && npm run lint && npm run typecheck && npm run format:check && npm run build`.
- The standing scenario suite is unchanged and still passes; this task adds no scenario,
  for the reason given under *Runtime reachability*.
- `make check` and `make test-cycle` still pass.

## Acceptance criteria

- [ ] A typed line appears in the page's list on send, distinguishable from a harness
      record and carrying no ordinal.
- [ ] An empty or whitespace-only submission is refused visibly. The failure a person
      can inject is pressing Enter on an empty input; what they observe is a message,
      not silence and not a blank entry in the list.
- [ ] Reloading the page loses the typed line, and the brief's demonstration says so.
      A local echo that survived a reload would mean this task had persisted something,
      which is P0-005's job and would make the two indistinguishable.
- [ ] The control is operable by keyboard alone, and the send action has an accessible
      label.
- [ ] `internal/` and `cmd/` are untouched. Checkable from the diff: this task's write
      set is entirely under `web/`.
- [ ] P0-001's demonstration still passes unchanged.

## Deferrals

| Deferred | Owning task |
|---|---|
| The submission reaching the runner, being accepted or refused, and surviving a reload | P0-005 |
| The submission reaching the harness | P0-006 |

## Traces to

`docs/roadmap.md` §1 rule 4 (*An input control that echoes what you typed back into the
page is observable* — one of the four closed loops), §3 Stage 0, §4 input path;
`docs/contracts.md` §3, §8 (*Input is accepted or refused, never queued silently*);
`docs/architecture.md` §4 *Sandbox attach surface*; ADR-0006.
