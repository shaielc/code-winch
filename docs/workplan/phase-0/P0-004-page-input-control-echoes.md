# P0-004: The page has an input control that echoes what you type

**Phase:** 0 — A sandbox you can talk to
**Shape:** capability
**Dependencies:** P0-001 (semantic: there is no page to put a control in until P0-001
serves one)

## Objective

Type a line into the page and press send, and the line appears in the page's own list —
nothing leaves the browser, and the composer is ready for the next line.

## Scope

- Add `web/src/attach/Composer.tsx` as a controlled `<textarea>` with a labelled send
  button. The textarea is required because Shift+Enter must insert a newline; Enter
  without Shift submits through the same handler as the button. After a successful
  submission, clear the value and restore focus to the textarea so consecutive lines
  need no pointer interaction.
- Keep the draft and submission validation in `Composer`. Trim only to decide whether a
  value is empty: a non-empty submission is echoed exactly as typed, including leading,
  trailing, and embedded whitespace. On an empty or whitespace-only attempt, keep the
  draft, add no item, and show `Enter a message before sending.` in an assertive status
  associated with the textarea. Clear that status as soon as the user supplies a
  non-whitespace character or a later submission succeeds. This makes refusal visible
  without inventing the runner refusal contract owned by P0-005.
- In `web/src/attach/App.tsx`, own an in-memory array of local submissions and pass an
  append callback to `Composer`. Render that array in a `Local messages` region beside,
  not as a synthetic member of, any runner-record list introduced by the concurrent
  output path. Each item renders the exact submitted text and the visible label `Local`;
  it has no ordinal field or placeholder. A local array is deliberate: reload must erase
  it, and P0-005 can replace the callback with runner submission rather than unwind
  browser persistence.
- Extend `web/src/attach/style.css` only enough to make the composer, validation status,
  and local items legible and visibly distinct from runner output. Do not encode record
  kinds or ordinal semantics in CSS class names.
- Add `web/src/attach/Composer.test.tsx` for the component-level keyboard, button,
  validation, exact-payload, clearing, and focus behavior. Extend `App.test.tsx` to prove
  a successful submission appears with the `Local` label and no ordinal, while retaining
  the existing posture check. The dependencies needed for these tests are already in
  `web/package.json`; do not add another interaction library solely for this task.
  `@testing-library/user-event` is not among them, so drive keys with `fireEvent`, and
  have the Enter handler call `preventDefault()` so Enter never also inserts a newline.

## Non-goals

- Any HTTP request, any server contract, and any change to `internal/`. This task
  declares no contract surface at all; that is what keeps it concurrent with the whole
  output path.
- Browser persistence, generated identifiers, timestamps, record-shaped local entries,
  and optimistic reconciliation. The local echo is component state only.
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
because there is no server capability to drive — P0-005, which adds one, adds `winch
input` with it.

At the dependency revision, `App.tsx` is the page shell and fetches `/api/session` to
render `Posture`; there is no record list or composer yet. The local list therefore lands
as its own region and does not depend on the concurrent output-path tasks. The existing
Vite entry point already renders `App`, and the sandbox already copies `web/dist`, so no
composition-root or container change is needed to make the control reachable.

## Write set

- `web/src/attach/Composer.tsx`, `web/src/attach/Composer.test.tsx`
- `web/src/attach/App.tsx`, `web/src/attach/App.test.tsx`
- `web/src/attach/style.css`

## Contract surfaces

None.

## Demonstration

    $ docker compose -f deployments/compose.yml up --build -d
    $ xdg-open http://127.0.0.1:8080

    → type "hello" and press Enter
    → expect: "hello" appears in the list, marked as local and carrying no ordinal, and
      the input is empty and focused

    → type "  keep me  ", then select the Send button without pressing Enter
    → expect: the list shows the leading and trailing spaces; the input is empty and
      focused again

    → type two lines by pressing Shift+Enter between them, then press Enter
    → expect: one local item contains both lines; Shift+Enter did not submit early

    → press Enter on an empty input
    → expect: nothing is added, focus stays in the input, and the associated alert says
      "Enter a message before sending."

    → reload the page
    → expect: "hello" is gone. It never left the browser, which is exactly what this
      task claims and the next task changes.

## Verification

- `cd web && npm test` — component tests for submit-on-Enter, newline on Shift+Enter,
  equivalent button and keyboard submission, exact preservation of non-empty payloads,
  visible rejection of empty and whitespace-only values, error recovery, clearing and
  refocusing after send, and the local marking without an ordinal.
- `cd web && npm run lint && npm run typecheck && npm run format:check && npm run build`.
- The standing scenario suite is unchanged and still passes; this task adds no scenario,
  for the reason given under *Runtime reachability*.
- `make check` and `make test-cycle` still pass.

## Acceptance criteria

- [ ] A typed line appears in the page's list on send, distinguishable from a harness
      record by the visible `Local` label and carrying no rendered ordinal or ordinal
      placeholder. The component test covers both the Enter and button paths.
- [ ] An empty or whitespace-only submission is refused visibly. The failure a person
      can inject is pressing Enter on an empty input; what they observe is the associated
      `Enter a message before sending.` alert, not silence and not a blank entry in the
      list. Typing non-whitespace content clears the alert.
- [ ] Shift+Enter inserts a newline without submitting. A later Enter submits one item
      whose text, including leading/trailing whitespace and the embedded newline, is
      unchanged; only all-whitespace detection trims.
- [ ] After a successful send, the composer is empty and focused. A refused send does
      not clear the draft or move focus, so correction is possible by keyboard alone.
- [ ] Reloading the page loses the typed line, and the brief's demonstration says so.
      A local echo that survived a reload would mean this task had persisted something,
      which is P0-005's job and would make the two indistinguishable.
- [ ] The textarea has an accessible name, the send button has the accessible name
      `Send message`, and the validation alert is programmatically associated with the
      textarea.
- [ ] `internal/` and `cmd/` are untouched. Checkable from the diff: this task's write
      set is entirely under `web/`.
- [ ] P0-001's demonstration still passes unchanged.

## Deferrals

| Deferred | Owning task |
|---|---|
| The submission reaching the runner, being accepted or refused, and surviving a reload | P0-005 |
| Replace the local append-only list's positional React keys with stable runner-backed identity when submissions become persistent records. Positional keys are acceptable only for P0-004's transient append-only local echo. | P0-005 |
| The submission reaching the harness | P0-006 |

## Traces to

`docs/roadmap.md` §1 rule 4 (*An input control that echoes what you typed back into the
page is observable* — one of the four closed loops), §3 Stage 0, §4 input path;
`docs/contracts.md` §3, §8 (*Input is accepted or refused, never queued silently*);
`docs/architecture.md` §4 *Sandbox attach surface*; ADR-0006.
