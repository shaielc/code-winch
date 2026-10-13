# Phase 2 — First real vendor

## Outcome and starting point

At phase exit an operator can select OpenAI Codex CLI or the shipped fake behind
the standalone sandbox wrapper, submit text from the existing attach page, and
observe the selected harness's output and exit result there. The Codex adapter
uses the versioned app-server protocol over standard-I/O pipes; the fake remains
the deterministic, account-free profile and passes the same wrapper scenario.
This completes roadmap Stage 1 for one real vendor, not a general plugin system
or a shared-deployment credential service.

Phase 2 consumes the complete Stage 0 exchange planned in
[Phase 1](../phase-1/README.md): owned child launch and cleanup, selectable fake
configuration, input delivery, live and retained raw records, runner-local
ordering, reload, input refusal, and bounded output/readers. None of that is
present merely because this design names it. At this commit the sandbox root
only constructs the attach HTTP handler, the attach server rejects non-GET
requests, and the fake is a standalone binary rather than a runtime profile
([sandbox root](../../cmd/winch-sandbox/main.go),
[attach server](../../internal/adapters/transport/attach/server.go),
[fake source](../../cmd/fake-harness/main.go), [system record](../../state.md)).
There is no `internal/runner` or harness-adapter package in the tracked source.
Runtime deployment and the Go/Docker gates were not re-demonstrated while
authoring this planning artifact; those limits are recorded in the system
record. Every Phase 2 implementation brief must recheck its named Stage 0
prerequisites at dispatch and must not implement a missing prerequisite as
Stage 1 work.

The selected first vendor is **OpenAI Codex CLI 0.159.3**, constrained to that
version until a separately reviewed compatibility change. The release tag is
[published upstream](https://github.com/openai/codex/releases/tag/rust-v0.159.3).
Its supported integration entrypoint is `codex app-server --listen stdio://`.
The version-pinned CLI declares `stdio://` as the default transport and the
`--stdio` alias, while the app-server implementation starts a single stdio
connection and processes typed JSON-RPC messages
([CLI source](https://github.com/openai/codex/blob/rust-v0.159.3/codex-rs/cli/src/main.rs#L542-L576),
[app-server source](https://github.com/openai/codex/blob/rust-v0.159.3/codex-rs/app-server/src/lib.rs#L770-L800)).
The same tag's exported protocol schemas define `initialize`, `thread/start`,
`turn/start`, `item/agentMessage/delta`, and `turn/completed` request and
notification shapes
([versioned schemas](https://github.com/openai/codex/tree/rust-v0.159.3/codex-rs/app-server-protocol/schema/json/v2)).
Together these are reproducible public source evidence for non-interactive
launch, pipe input and structured output without recording credentials,
prompts, workspace content or account identifiers.

The version-pinned CLI exposes ChatGPT and stdin-fed API-key login plus the
`codex login status` subcommand; its status implementation returns success for
a recognized stored login and failure when signed out or unreadable
([CLI login arguments](https://github.com/openai/codex/blob/rust-v0.159.3/codex-rs/cli/src/main.rs#L503-L542),
[status implementation](https://github.com/openai/codex/blob/rust-v0.159.3/codex-rs/cli/src/login.rs#L447-L505)).
Phase 2 requires one of those logins before launch, reports only the readiness
result, and never prints or retains the command's account/key-bearing detail.

Native process exit is observed independently of protocol turn status. For the
selected entrypoint, closing the sole stdio connection ends the server, a
graceful return propagates normally, and the CLI explicitly exits `0` for its
forced-shutdown result
([connection-close path](https://github.com/openai/codex/blob/rust-v0.159.3/codex-rs/app-server/src/lib.rs#L1140-L1170),
[CLI exit path](https://github.com/openai/codex/blob/rust-v0.159.3/codex-rs/cli/src/main.rs#L1240-L1278)).
The wrapper preserves the actually observed native code or signal; it maps an
expected clean exit after wrapper-initiated pipe closure separately from an
unexpected zero exit and any nonzero/signal exit. This evidence does not assign
fixed meanings to undocumented nonzero codes. A failed or interrupted turn is
decoded from protocol status and is not relabelled as a process exit.

**D1 is resolved** to Codex CLI 0.159.3. **D5 remains deferred under its roadmap
trigger:** the selected supported app-server entrypoint explicitly uses stdio
pipes, and the cited protocol requires no resize, raw mode, or TTY detection.
Phase 2 therefore plans pipes, not a PTY. A future adapter version may revisit
D5 only when evidence for the roadmap's terminal-semantics trigger exists.

Sources: [Stage 1 roadmap](../../roadmap.md),
[harness-adapter architecture](../../architecture.md#harness-adapters),
[adapter construction](../../code-structure.md#4-adding-a-harness),
[sandbox attach contract](../../contracts.md#8-sandbox-attach-contract),
[fake profile contract](../../contracts.md#9-fake-runtime-profile-contract),
[ADR-0003](../../decisions/0003-capability-based-adapters.md), and
[single-user login boundary](../../security.md#7-credentials-and-login).

## Concern boundaries

| Concern | Owner |
|---|---|
| Genuine harness wrapper seam and explicit profile selection | Increment 1; both fake and Codex implement the same boundary |
| Versioned descriptor, supported CLI version, launch mode, pipe transport and declared input/output/login capabilities | Increment 1 owns the descriptor contract; later increments consume it |
| Fake-profile parity when the wrapper seam first changes | Increment 1; account-free standing scenario plus controllable fake behavior |
| Login prerequisite and content-free readiness/refusal | Increment 2; single-user pre-login only, with no credential value crossing the attach surface |
| Codex process launch and app-server initialization | Increment 3; consumes Stage 0 child ownership and the readiness result |
| Native app-server input encoding and output decoding | Increment 4; JSON-RPC turn input and vendor output/status, incrementally decoded across arbitrary chunks |
| Native process result and harness-specific interpretation | Increment 5; visible separately from protocol turn completion |
| Stage 0 exchange, ordering, retention, refusal, bounds and reader isolation | Phase 1; prerequisites, never injected into Phase 2 |
| PTY, resize and raw terminal behavior | Roadmap D5; remains deferred because the selected protocol is pipe-native |
| Intentional stop and run lifecycle | Roadmap Stage 4 / Phase 5 |
| Stop escalation, orphan recovery, scoped short-lived credential injection and shared-deployment enforcement | Roadmap Stage 5 / Phase 6 |
| Second vendor and proof that the seam generalizes | Roadmap Stage 6 / Phase 7 |

The descriptor declares only evidence-backed behavior. It does not infer that a
successful login-status check guarantees model entitlement or network reachability.
Provider-only app-server records remain namespaced at the adapter boundary; the
generic runner and attach transport contain no Codex-specific branches.

## Ordered increments

| Order | Increment | Observable change | Includes | Excludes | Prerequisites |
|---|---|---|---|---|---|
| 1 | Selectable harness profiles | The maintained operator path selects `fake` or `openai-codex` and the attach page reports the selected adapter ID, adapter version, supported CLI version, pipe transport, login mode and input/output capabilities; the unchanged fake exchange scenario passes through the new seam | One genuine wrapper interface, explicit composition-root registration, versioned descriptors, configuration selection, fake adapter migration and parity evidence | Starting Codex, checking an account, vendor decoding, claiming the seam works for a second real vendor | Completed Stage 0 selectable fake launch/exchange, attach presentation and runtime configuration |
| 2 | Codex login readiness | With the Codex profile selected, the page/operator path reports content-free `ready` or `login required`; input is refused while login is unavailable and the fake remains account-free | Version-pinned installation/readiness validation, `codex login status` boundary, single-user prerequisite instructions and secret-free diagnostics | Browser login, credential values, scoped injection, entitlement claims, launching an unready agent | 1; Stage 0 input availability/refusal |
| 3 | Initialized Codex process | For a ready Codex profile, the page reports that the pipe-connected app server initialized; startup/protocol failure is visible without losing fake operation | Version-pinned launch construction, stdio ownership, app-server initialize handshake, provider process availability and minimum cleanup | User turns, semantic output decoding, process-exit interpretation, PTY support | 1 and 2; Stage 0 child ownership/status/cleanup and bounded raw I/O |
| 4 | Codex conversation exchange | Text submitted on the attach page starts a Codex turn and decoded agent text and terminal turn status become visible and survive reload; the fake profile still completes its equivalent standing scenario | Native JSON-RPC input encoding, incremental framing/decoding, agent-message output, completed/failed/interrupted turn interpretation, provider-extension preservation and real/fake scenario parity | Rich tool/file projections, universal prompt/reply correlation, resume by canonical event sequence, native process-exit meaning | 3; complete Stage 0 input, output, records, reload, bounds and refusal behavior |
| 5 | Interpreted harness exit | Ending or failing either harness makes the attach page show the native process result and a profile-specific interpretation, separately from the last turn result; later input is refused | Native exit code/signal observation, orderly/unexpected/nonzero interpretation for Codex and harness-specific interpretation for the fake, retained observable exit record | Intentional operator stop, retry/attempt lifecycle, escalation, orphan recovery | 3 and 4; Stage 0 exit observation, retained records and unavailable-child refusal |

The order is experience-led: selection and truthful readiness are observable
before the first real process, and initialization is observable before a turn.
Input and decoded output remain one increment because sending a turn without a
decoded visible result would not close a useful operator loop. Process exit is
separate because protocol turn completion is not native child termination.

## Task coverage

| Order | Increment | Coverage |
|---|---|---|
| 1 | Selectable harness profiles | [P0-011](../phase-0/P0-011-plan-selectable-harness-profiles.md) assigns independent derivation; implementation brief **not yet derived** |
| 2 | Codex login readiness | **Not yet derived** |
| 3 | Initialized Codex process | **Not yet derived** |
| 4 | Codex conversation exchange | **Not yet derived** |
| 5 | Interpreted harness exit | **Not yet derived** |

Coverage records brief derivation, not implementation or prerequisite
completion. The first uncovered increment is **Selectable harness profiles**.
[P0-011](../phase-0/P0-011-plan-selectable-harness-profiles.md) is the concrete
Phase 0 follow-up that assigns its independent derivation one increment at a
time; it reserves no implementation ID.

## Phase exit

With Stage 0 complete, start the standalone loopback sandbox first with the
fake profile and then with the installed, logged-in Codex CLI 0.159.3 profile.
On the existing attach page inspect the selected versioned descriptor and
readiness, submit text, observe adapter-decoded output and turn status, reload
the retained exchange, and observe native exit plus the separate
harness-specific interpretation. Repeat the standing exchange/exit scenario
with the account-free fake profile and its deterministic failure controls.

The demonstration establishes compatibility with the pinned Codex version over
pipes and preserves the fake configuration. It does not prove compatibility
with later Codex versions, a second vendor, PTY behavior, credential isolation,
multi-user safety, intentional run stop, retries, or orphan recovery.
