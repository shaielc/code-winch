# Phase 1 — Sandbox exchange

## Outcome and starting point

At phase exit an operator can start the loopback-only sandbox by hand, observe
its fake child state and live raw output on the attach page, submit text and see
subsequent fake-produced output, then reload the ordered session exchange.
Input availability and failure are represented truthfully; output and input are
bounded, and a slow live reader cannot backpressure the child. This completes
roadmap Stage 0, not a generic real-vendor adapter or seamless reconnect.

The repository already serves a loopback attach page and health/posture routes,
has a status-only CLI and browser-local composer, and ships a fake binary in the
image. The sandbox composition root does not launch that binary; the page has
no live child stream, session history or real input delivery. These are source
observations rather than a demonstrated live deployment at this commit; see
[the recorded limits](../../state.md).

Sources: [Stage 0 roadmap](../../roadmap.md), [current state](../../state.md),
[sandbox attach and fake contracts](../../contracts.md),
[ADR-0005](../../decisions/0005-sandbox-resident-runner.md),
[ADR-0006](../../decisions/0006-two-interaction-surfaces.md),
[workplan model](../../../skills/shared/workplan-model.md).

## Derivation principle

Order increments by progressively enabling the experience on the **existing
attach page**, not by backend layer. A truthful shallow state (for example
`Unknown`) is a legitimate observable step before a sandbox-reported state
(`Down`) and then actual child state (`Running` / `Exited`). An increment
includes the minimum composition, supported fake-profile and maintained
operator path required for its manual observation; it does not silently claim
the deeper behavior of a later increment. Manual demonstration is distinct
from automated standing-scenario verification.

Numbers give a preferred derivation order, **not** dependency edges between
neighbors. The prerequisites column names actual consumed behavior; independent
input and output branches may proceed concurrently. The process can pipeline
implementation of already-derived work with serial, separately authored
planning handoffs. Shallow, truthful UI states are not prohibited because the
backend does not yet exist.

## Concern boundaries

| Concern | Owner |
|---|---|
| Attach-page child state and actual child ownership | Increments 1–3; increment 3 owns launch and minimum cleanup |
| Raw live-output presentation and transport | Increments 4–6; 5 first reaches child bytes, 6 makes delivery incremental |
| Submission presentation, sandbox receipt, child write and fake exchange | Increments 7–12; no provider-independent reply semantics |
| Runner-local record identity and a single session record order | Increment 13 first defines output ordinals; 14 extends the same boundary to input |
| One private retained session store, reload and history/live join | Increment 15 owns persistence, 16–18 consume it |
| Fake-specific input capability, channel status and refusal | Increments 19–25; unavailable/unsupported behavior is not prerequisite for first live output |
| Output bounds, malformed output and reader isolation | Increments 26–30; basic safety and existing gates remain inherited at every step |
| Existing startup, posture, status, loopback deploy and maintained operator checks | Preserved by every increment; new operator-visible behavior gets its minimum maintained hand path in the same task |
| Real vendor adapter, PTY/native exit interpretation | Roadmap Stage 1 / Phase 2 (D1, D5) |
| Structured projections and seamless resume | Roadmap Stage 2 / Phase 3 |
| Runs/control plane, lifecycle, enforced security and workflows | Roadmap Stages 3–6 / Phases 4–7 |

Ordinals describe the order in which the runner **records** observations, not an
unknown harness's internal order or semantic message boundaries. Inputs and
outputs observed through different channels have no inferred causal relation.

Stage 0 treats the fake as a shipped, malleable runtime profile. Increments 10–17 describe operator-visible exchange and history guarantees, not a fake-specific data model or protocol. The fake only supplies a controllable demonstration profile. Current
JSON-lines dialect, reply behavior and flags do not establish universal
real-harness input readiness, message correlation, response completion, PTY
semantics or vendor decoding. Raw process observations and fake-specific
behavior must be described as such.

## Ordered increments

| Order | Increment | Observable change / manual demonstration | Includes | Excludes | Prerequisites |
|---|---|---|---|---|---|
| 1 | Child status presentation | The attach page shows an honest `Unknown` child state | A status region and unknown state in the existing page | Child spawning or inferred readiness | Existing attach page |
| 2 | Sandbox-backed child status | The page reports `Down` for a sandbox that has no child, rather than a browser constant | Sandbox-owned status observation and UI transport; maintainable hand observation | Child startup, output transport | 1; existing attach server |
| 3 | Actual child status | The page changes to `Running` for the owned fake and `Exited` when it ends | Child spawn/ownership, basic cleanup and supervision, status wiring and observable fake controls | Output processing, delivery, vendor-native exit interpretation | 2; shipped fake binary and deployable sandbox |
| 4 | Harness output area | The page identifies the harness-output area with a truthful empty state | Presentation and empty-state behavior in the existing attach page | Generating or streaming output | Existing attach page; independent of 1–3 |
| 5 | First live child output | Raw output emitted by the owned fake appears in the page | Child byte reading, minimum live transport and browser display, maintained operator observation | Durable replay, input, semantic decoding, full malformed-output hardening | 3 and 4 |
| 6 | Incremental output | The first emitted piece appears before the second is emitted | Incremental transport/rendering, delayed controllable fake demonstration | History replay, content interpretation, advanced fan-out guarantees | 5 |
| 7 | Draft versus submitted | The existing composer distinguishes unsent text from locally submitted text without claiming delivery | Only missing truthful composer-state presentation | Backend receipt, child delivery, input history | Existing browser-local composer; independent of output |
| 8 | Sandbox receipt | A composer submission obtains a sandbox-origin receipt distinct from local UI acknowledgment | Minimum browser-to-sandbox request/result, maintained operator path | Claiming child processing, durable input, idempotency policy | 7; existing attach server |
| 9 | Child input write | A submitted line is written to the fake's supported input channel and the UI reports the write result | Minimum child-stdin ownership/write path and truthful result; fake can be adapted | Semantic readiness, durable record, native vendor modes, exactly-once delivery | 3 and 8; independent of 5–6 |
| 10 | Output after submission | After submitting text, a person can observe new harness output in the existing output area | End-to-end interaction through the current supported fake profile, with no inferred semantic reply contract | Universal input/reply correlation, response completion and vendor interpretation | 5 and 9 |
| 11 | Honest submission outcome | The composer distinguishes locally displayed input from sandbox receipt and a reported channel-write result | User-visible outcome presentation from the already reachable submission path | Claims of harness comprehension or acceptance | 8 and 9 |
| 12 | Readable exchange | A person can distinguish submitted text and later output while continuing to interact on the same page | Combined presentation of existing input and output observations | Attributing a particular output to a particular input; retained history | 10 and 11 |
| 13 | Runner-observed output order | The operator can inspect consecutive raw output observations in the order the runner received them | Runner-local ordinal allocation on observed output chunks and publication in that order | Claims about harness emission order, semantic message boundaries, timestamps as ordering keys, or retained history | 5; independent of input |
| 14 | Runner-recorded exchange order | Submitted input receipts and observed output are shown in the order the runner recorded them | Extend the single runner-local ordinal allocation to input receipts and observed output | Causal ordering inside the harness, prompt/reply matching and canonical run events | 8 and 13 |
| 15 | Submission survives reload | A person reloads the page and still sees their earlier submitted text | The single private session persistence implementation, plus the minimum reload path | Cross-session retention, seamless rejoin | 8 and 14 |
| 16 | Output survives reload | A person reloads the page and still sees earlier harness output | Consume the same private history path for output; no second persistence implementation | Semantic output projections, cross-session retention | 13 and 15 |
| 17 | Recorded order survives reload | After submitting input and observing output, reloading reproduces the runner-recorded order | Reuse stored runner-local ordinals for input receipts and raw output observations | Claims about child-internal ordering, automatic reconnect or gap repair | 14, 15 and 16 |
| 18 | Live/history continuity | New output arriving during history load appears without duplicated visible records | Bounded bootstrap/live handoff over existing records | Heartbeat/caught-up protocol and Stage 2 gap-free resume | 6, 16 and 17 |
| 19 | Input availability presentation | The page has an independently identifiable `Unknown` input-availability state | Truthful presentation distinct from child-running status | Pretending semantic readiness is known | Existing attach page; may be derived independently and later |
| 20 | Fake input capability report | The UI reports which input kind the active fake profile declares supported | Fake-specific capability description through the harness boundary | General real-vendor capability semantics, readiness guarantee | 3 and 19 |
| 21 | Observed channel availability | The displayed availability changes with the fake's observable transport/channel state | Harness-specific channel observation and reporting; unknown when not knowable | Claim that writable means the harness understands input | 9 and 20 |
| 22 | Unavailable-input refusal | Input submitted to a known unavailable channel is refused, not silently queued | Truthful refusal and no silent acceptance after a known channel failure | Full adapter-specific readiness inference, retries | 9 and 21 |
| 23 | Invalid or unsupported input refusal | Malformed requests and unsupported fake input kinds receive distinct defined results | Validation and fake-declared kind checks with content-safe diagnostics | Provider-independent input modes or vendor policies | 8 and 20 |
| 24 | Idempotent submission | Repeating the same session-scoped key gives the original logical submission | Stable input identity, duplicate lookup in shared record path | Cross-run idempotency or delivery exactly once | 15 and 23 |
| 25 | Interrupted-delivery observation | Failure between receipt and child write has a truthful undelivered result | Accepted-versus-written distinction, bounded write/failure reporting | Queue for later, run retry lifecycle | 9 and 11 |
| 26 | Excessive-output bound | An oversized fake output does not grow a published record without bound | Explicit raw output/record bound and observable content-safe outcome | Vendor decoding and terminal interpretation | 5 |
| 27 | Malformed-output recovery | An adversarial framing/output condition does not conceal later valid bytes | Resynchronization for chosen raw transport boundary and safe diagnostics | Treating arbitrary CLI bytes as invalid JSON or semantic error | 6 and 26 |
| 28 | Slow-reader isolation | Stalling one live reader does not block child reads or another reader | Per-reader bounded publication and fake multi-reader observation | Automatic history recovery | 6 |
| 29 | Reader overflow outcome | A saturated reader receives a bounded, observable drop or disconnect outcome | Reader-buffer bounds, drop/disconnect signal without content leakage | Automatic resume, heartbeat protocol | 28 |
| 30 | Manual refetch after overflow | An operator can fetch missing retained records from the last retained ordinal after a drop | Explicit history refetch using existing private session history | Seamless reconnect and automatic gap repair | 16 and 29 |

## Task coverage

| Order | Increment | Coverage |
|---|---|---|
| 1 | Child status presentation | [P1-009](P1-009-child-status-presentation.md), independently derived by [P0-008](../phase-0/P0-008-plan-owned-live-harness-output.md), owns the accessible static `Child status: Unknown` presentation |
| 2 | Sandbox-backed child status | [P1-011](P1-011-sandbox-backed-child-status.md), independently derived by [P0-010](../phase-0/P0-010-plan-sandbox-backed-child-status.md), consumes P1-009 and owns the additive sandbox session `childStatus: "Down"` transport |
| 3 | Actual child status | **Not yet derived** |
| 4 | Harness output area | **Not yet derived** |
| 5 | First live child output | **Not yet derived** |
| 6 | Incremental output | **Not yet derived** |
| 7 | Draft versus submitted | **Not yet derived** |
| 8 | Sandbox receipt | **Not yet derived** |
| 9 | Child input write | **Not yet derived** |
| 10 | Output following input | **Not yet derived** |
| 11 | Composer delivery feedback | **Not yet derived** |
| 12 | Combined exchange presentation | **Not yet derived** |
| 13 | Observable output ordering | **Not yet derived** |
| 14 | Shared input/output ordering | **Not yet derived** |
| 15 | Retained submitted input | **Not yet derived** |
| 16 | Retained child output | **Not yet derived** |
| 17 | Ordered reload exchange | **Not yet derived** |
| 18 | Live/history continuity | **Not yet derived** |
| 19 | Input availability presentation | **Not yet derived** |
| 20 | Fake input capability report | **Not yet derived** |
| 21 | Observed channel availability | **Not yet derived** |
| 22 | Unavailable-input refusal | **Not yet derived** |
| 23 | Invalid or unsupported input refusal | **Not yet derived** |
| 24 | Idempotent submission | **Not yet derived** |
| 25 | Interrupted-delivery observation | **Not yet derived** |
| 26 | Excessive-output bound | **Not yet derived** |
| 27 | Malformed-output recovery | **Not yet derived** |
| 28 | Slow-reader isolation | **Not yet derived** |
| 29 | Reader overflow outcome | **Not yet derived** |
| 30 | Manual refetch after overflow | **Not yet derived** |

Coverage records brief derivation, not implementation. P1-009 and P1-011 are
derived implementation briefs ready for different authors to refine with
`skills/refine-task/SKILL.md`; their existence
does not claim that the UI or API has changed. P1-011 depends on P1-009 because
it consumes that presentation identity. Remaining increments receive
independently authored concrete Phase 0 handoffs **one at a time** as planning
proceeds; no implementation ID for increment 3 or later is reserved.
The increment labels are not substitutes for task IDs or owners of completion
deferrals. An implementation brief must include its own inherited safety and
operability obligations; later increments never excuse a broken current slice.

## Phase exit

With the shipped fake profile, manually start the loopback-only sandbox and
observe the child status, raw output arriving incrementally, submissions,
child-write outcomes and later fake-produced output on the attach page. Reload
and observe the same session's gap-free runner-local ordinal history. Show
validation, duplicate-key, unsupported and unavailable-input results without
silently queuing; cause excessive or malformed output and still observe
subsequent valid output; stall one reader without stalling harness reads or a
second reader, observe bounded overflow, and refetch retained history manually.

The demonstration establishes only fake-specific input behavior and a single
manually started sandbox session. It does not establish universal real-harness
readiness, semantic prompt/reply correlation, native exit interpretation,
structured views, automatic resume, run lifecycle or shared-deployment
security. Preserve repository gates and existing health, posture, page and
status behavior across every implementation increment.
