# Phase 1 — Sandbox exchange

## Outcome and starting point

At phase exit, an operator can start the loopback-only sandbox container by
hand, watch its sandbox-owned fake harness, submit a line from the attach page,
see the fake's reply, and reload the page without losing the ordered exchange.
Harness output and input remain bounded under malformed output, an unavailable
child, or a slow reader. This completes roadmap Stage 0; it does not introduce a
run, control plane, real vendor, or seamless reconnect.

The repository already has a deployable sandbox image, loopback-published
attach server, posture and health endpoints, static attach page, status CLI,
and a browser-local composer. The image contains `fake-harness`, but
`cmd/winch-sandbox/main.go` constructs only the attach server; the server has no
stream, records, or input route, and the composer only appends component state.
There is no runner or session store. The standing end-to-end scenario checks
startup, posture, page, and status, not an exchange. `docs/state.md` records
that the Go, web, container, browser, and scenario runtime evidence remains
limited by unavailable tools in the closure environment; these source facts do
not prove that the deployed slice runs at this commit.

Sources: [roadmap Stage 0 and its two paths](../../roadmap.md), [current system
state](../../state.md), [sandbox attach and fake contracts](../../contracts.md),
[ADR-0005](../../decisions/0005-sandbox-resident-runner.md),
[ADR-0006](../../decisions/0006-two-interaction-surfaces.md), and the
[workplan model](../../../skills/shared/workplan-model.md).

## Concern boundaries

| Concern | Owner |
|---|---|
| Sandbox-owned fake child, raw live output, minimum cleanup, and maintained hand/operator path | Increment 1 — Owned live harness output |
| Runner-local record identity, gap-free ordinal allocation, and append-before-publication semantics shared by Stage 0 records | Increment 1 defines the session record seam; later record-producing increments consume it |
| One private session persistence implementation and its retained ordinal-ordered history | Increment 2 — Retained output on reload; every later durable record consumes it |
| Persisted typed input, idempotency, validation, and refusal before delivery exists | Increment 3 — Recorded input and refusal; it consumes increment 2's session persistence rather than defining a second store |
| Bounded input delivery to an available child and correlation with the fake's reply | Increment 4 — Delivered input and reply |
| Excessive or malformed harness output, resynchronization, and content-safe diagnostics | Increment 5 — Bounded malformed output |
| Per-reader buffering, slow-reader isolation, disconnect/drop observation, and bounded publication | Increment 6 — Isolated live readers |
| Existing startup, posture, page, status, loopback deployment, and repository gates | Every increment preserves them; increment 1 grows the fake-profile scenario and each later increment extends it for new behavior |
| Real-vendor selection, adapter seam, terminal semantics, and native exit interpretation | Roadmap Stage 1 / workplan Phase 2; D1 and D5 retain their triggers |
| Structured projections, caught-up/heartbeat protocol, reconnect gap repair, and gap-free resume | Roadmap Stage 2 / workplan Phase 3 |
| Runs, canonical sequence numbers, intentional stop, retry, and lifecycle | Roadmap Stage 4 / workplan Phase 5 |
| Stop escalation, orphan recovery, credentials, and enforced isolation or egress | Roadmap Stage 5 / workplan Phase 6 |

The fake is a shipped, malleable profile. Its current JSON-lines input, flags,
scanner limit, malformed-line exit, and lack of reply correlation do not decide
the product protocol or the controls used to demonstrate these increments.

## Ordered increments

| Order | Increment | Observable change | Includes | Excludes | Prerequisites |
|---|---|---|---|---|---|
| 1 | Owned live harness output | Starting the sandbox also starts the fake; its raw output appears live on the existing attach page and remains drivable by the minimum operator path | Sandbox-resident child ownership and reaping; raw incremental output; runner-local record identity and ordinal allocation; append-before-publication seam; bounded basic child cleanup; composition/image wiring; fake-profile scenario growth while retaining startup/posture/status | Durable reload, input, semantic decoding, native-exit interpretation, intentional stop, retries, PTY behavior | Existing sandbox page, container, fake binary, loopback posture, and gates |
| 2 | Retained output on reload | Output already observed live is present in ordinal order after a page reload in the same sandbox session | The single private session persistence implementation; output history from an ordinal; reload/bootstrap in the page and operator path; persistence/deployment wiring; continuity of the existing live-output scenario | Continuity after the session ends, seamless reconnect, heartbeat/caught-up signaling, gap repair, structured views, cross-session retention | Increment 1's record seam and emitted output; may proceed independently of increments 5 and 6 |
| 3 | Recorded input and refusal | Submitting in the existing composer produces a stable accepted input record that survives reload, while invalid, unsupported, duplicate, or currently unavailable submissions return the defined result rather than a browser-only echo or silent queue | Typed bounded request; session-scoped idempotency; validation and content-free diagnostics; input records appended through increment 2's session persistence with runner-local ordinals; page and minimum operator path; fake-profile scenario coverage | A second input-specific store, writing to child stdin, harness reply, authentication/authorization, retry, raw terminal input, run-command semantics | Increments 1 and 2; the input path consumes the already-owned session persistence contract |
| 4 | Delivered input and reply | An accepted composer submission reaches the running fake within a bound and the correlated fake reply appears through the live output path | Accepted-before-delivery ordering; child input ownership; capability/channel availability check; bounded write; correlation; accepted-but-undelivered diagnostic; refusal after child unavailability; page-to-fake-to-page scenario | Queuing for later, real-vendor semantics, intentional stop/restart, run lifecycle | Increments 1 and 3; increment 2 is inherited through increment 3, and its history path also supports the phase-exit reload observation |
| 5 | Bounded malformed output | Oversized or malformed fake output cannot grow a record without bound or prevent later valid output from being observed; the operator sees a content-safe diagnostic | Explicit output/record bounds; incremental decoding and resynchronization; preservation or classification at the raw Stage 0 depth; adversarial fake controls and scenario | Vendor semantic decoding, terminal interpretation, renderer sanitization, native-exit meaning | Increment 1; independent of retained history and the input path except for shared record-contract ownership |
| 6 | Isolated live readers | A stalled attach client neither blocks harness reads nor delays another reader, and bounded overflow has an observable content-free outcome with the last retained ordinal needed to refetch history | Per-reader bounded delivery; slow-reader disconnect/drop behavior; harness-read isolation; adversarial multi-reader scenario | Stage 2 seamless resume, heartbeat/caught-up protocol, automatic gap repair, control-plane fan-out | Increment 1 for live publication. Increment 2 supplies the history path used after a drop but is not required to establish non-backpressure |

Increments 2, 5, and 6 are deliberate branches after the live record seam, not
a serial chain. Increment 2 alone owns session persistence; increment 3 follows
it because durable input consumes that shared store and ordinal contract.
Increment 4 joins the live-output and recorded-input paths. Output hardening and
reader isolation remain independent of persistence except where increment 6's
post-drop operator recovery uses increment 2's history path. Shared files alone
are collision warnings, not extra dependency reasons.

## Task coverage

| Order | Increment | Coverage |
|---|---|---|
| 1 | Owned live harness output | [P0-008](../phase-0/P0-008-plan-owned-live-harness-output.md) assigns independent derivation; implementation brief **not yet derived** |
| 2 | Retained output on reload | **Not yet derived** |
| 3 | Recorded input and refusal | **Not yet derived** |
| 4 | Delivered input and reply | **Not yet derived** |
| 5 | Bounded malformed output | **Not yet derived** |
| 6 | Isolated live readers | **Not yet derived** |

Coverage records brief derivation, not implementation or completion. P0-008 is the complete Phase 0 planning handoff for a different task author to
derive the first increment's implementation brief; no implementation ID is
reserved yet. Subsequent uncovered increments receive separately authored
handoffs one at a time through Phase 0.

## Phase exit

With the fake profile, start the sandbox by hand, observe live harness output in
the attach page, submit input, observe the correlated reply, reload, and observe
the same gap-free runner-local ordinal history. Show invalid, duplicate,
unsupported, and unavailable-child input outcomes without silent
queueing; excessive and malformed output remains bounded and later valid output
is observable; a deliberately stalled reader does not delay the harness or a
second reader and receives the documented bounded-delivery outcome. The
startup/posture/status path and repository gates remain green.

This exit proves a single manually started, loopback-only sandbox session and a
malleable fake. It does not prove a real vendor, terminal semantics, structured
views, seamless resume, multiple sandboxes, run lifecycle, or shared-deployment
security.
