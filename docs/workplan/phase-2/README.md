# Phase 2 — First real vendor

## Outcome and starting point

At exit, a person using the existing loopback-only sandbox attach page can select
either the shipped fake profile or Claude Code, exchange input and output through
the same wrapper, and see both the process's native exit result and the adapter's
interpretation. This is roadmap Stage 1: one real agent proves the first genuine
harness-adapter seam while the fake remains a supported profile. It does not prove
that the seam accommodates a second vendor; Stage 6 owns that evidence.

Phase 2 consumes the completed Stage 0 exchange rather than rebuilding it. That
prerequisite is planned in [Phase 1](../phase-1/README.md), but it is not present at
this design commit: `cmd/winch-sandbox/main.go` constructs only the attach HTTP
handler, `internal/runner/` does not exist, and the only attach API routes are
health and session metadata. The image contains `fake-harness`, whose standalone
stdin/stdout behavior and failure controls are exercised by unit tests, but the
sandbox starts no child and offers no integrated fake profile. The browser's
composer remains local-only. These source facts are recorded in
[`docs/state.md`](../../state.md); no Stage 0 runtime behavior is claimed here.

The selected first vendor is **Claude Code 2.1.296**, pinned and rejected on a
different reported version until a later phase revision updates the supported
version. Anthropic's immutable
[2.1.296 release](https://github.com/anthropics/claude-code/releases/tag/v2.1.296)
supplies the reproducible version reference. Its
[CLI reference](https://code.claude.com/docs/en/cli-usage) documents `-p` as
non-interactive print mode, `stream-json` input and output, stdin user-message
replay, and error exits such as a reached turn limit. Its
[setup guide](https://code.claude.com/docs/en/getting-started) documents the
supported Linux packages, installation/version checks, network requirement, and
the prerequisite interactive account or provider authentication. These public
sources contain no Code Winch credentials, prompts, workspace content, or account
identifiers. A task must verify the pinned binary's reported version and fixture
behavior without recording any of those values.

The selected transport is **ordinary stdin/stdout pipes**, using Claude Code
print mode with streaming JSON in both directions. The cited CLI contract is
explicitly programmatic and pipe-oriented; it provides no evidence that this
mode requires resize events, raw mode, or TTY detection. Therefore D5 remains
deferred under its roadmap trigger: revisit only if execution of the pinned CLI
demonstrates a terminal requirement. A task author may not choose PTY as an
alternative without routing that new evidence through Phase 0.

Runtime verification is limited to source inspection at this commit. No Claude
Code binary or credential is checked into the repository, no live vendor account
is required in CI, and this planning artifact did not run an authenticated agent.
Stage 1 requires a single-user login prepared before sandbox launch and reports
its readiness truthfully; scoped, short-lived credential injection and egress
enforcement remain Stage 5 work.

Sources: [roadmap Stage 1 and D1/D5](../../roadmap.md),
[sandbox contract](../../contracts.md), [code structure](../../code-structure.md),
[security posture](../../security.md), [capability adapter ADR](../../decisions/0003-capability-based-adapters.md),
and [sandbox-resident runner ADR](../../decisions/0005-sandbox-resident-runner.md).

## Concern boundaries

| Concern | Owner |
|---|---|
| Harness profile selection and truthful page presentation | Harness profile presentation |
| Versioned capability-descriptor contract, fake descriptor, and first genuine adapter seam | Harness profile presentation |
| Claude Code descriptor values for stable identity, adapter version, I/O and login modes, and capabilities | Vendor readiness |
| Fake profile parity at the changed wrapper seam | Harness profile presentation |
| Claude Code 2.1.296 version check and launch construction | Claude Code exchange |
| Non-interactive stdin/stdout transport and input/output ownership | Claude Code exchange |
| Single-user vendor login prerequisite and readiness/refusal | Vendor readiness |
| Bounded incremental Claude stream-JSON decoding into the existing sandbox record path, including adversarial output | Vendor output interpretation |
| Native process exit result | Native exit visibility |
| Harness-specific interpretation of the native exit | Native exit visibility |
| PTY support | Deferred decision D5; pipes are decided for the selected CLI and D5 reopens only on evidence of resize, raw-mode, or TTY-detection behavior |
| Stage 0 child ownership, ordered exchange, retained history, and server input delivery | Phase 1; required behavior consumed by every increment here |
| Second adapter and proof that the seam generalizes | Roadmap Stage 6 |
| Intentional run stop and lifecycle | Roadmap Stage 4 |
| Stop escalation, orphan recovery, scoped credential injection, and egress enforcement | Roadmap Stage 5 |

## Ordered increments

| Order | Increment | Observable change | Includes | Excludes | Prerequisites |
|---|---|---|---|---|---|
| 1 | Harness profile presentation | The attach page names the selected fake profile and displays its versioned capabilities before an exchange; the maintained operator path can select it and the Stage 0 fake exchange remains green | A genuine wrapper-owned adapter interface, versioned capability descriptor, runtime profile selection, fake implementation, and parity evidence against the standing Stage 0 exchange | Claude launch, vendor login, vendor decoding, claims about a second adapter | Completed Stage 0 child ownership, input/output, retained exchange, and attach-page path from Phase 1 |
| 2 | Vendor readiness | Selecting Claude Code shows its declared capabilities and whether the pinned executable and required single-user login are ready; absence, wrong version, and unauthenticated state refuse launch without exposing credential material | Claude Code descriptor values (stable ID, adapter version, I/O modes, login modes, and capabilities), executable/version preflight, login boundary and truthful page/operator status | Launching a conversation, credential provisioning or injection, egress enforcement | Harness profile presentation; Stage 0 page and record path |
| 3 | Claude Code exchange | An authenticated operator selects Claude Code and the same attach-page input/output demonstration reaches the real agent through the wrapper | Adapter-owned launch construction, non-interactive print mode, stream-JSON stdin/stdout pipes, input encoding, raw output capture, cleanup, and the same maintained exchange scenario used by the fake | Semantic vendor decoding beyond safely retaining/showing the raw stream; native-exit presentation | Vendor readiness; Stage 0 delivery, output, history, and reload behavior |
| 4 | Vendor output interpretation | The attach page distinguishes decoded Claude response and operational records while preserving the underlying ordered exchange and confidentiality defaults; malformed, oversized, schema-invalid, and arbitrarily chunked fixtures produce a bounded safe diagnostic rather than content leakage, an unbounded allocation, or a lost decoder boundary | Incremental vendor-specific stream-JSON framing across arbitrary byte chunks, explicit size and schema bounds, safe unknown/malformed-message handling, adversarial and fuzz evidence, mapping into the existing sandbox record path, and visible decoded output/diagnostics | Multiple projections, canonical control-plane events, second-vendor abstractions, exit meaning | Claude Code exchange; Stage 0 ordered record and replay path |
| 5 | Native exit visibility | When either profile ends, the attach page and maintained operator path show the native process result separately from the adapter's interpretation; fake success/failure and sanitized Claude success/failure fixtures exercise both | Native exit code/signal result, adapter-specific success/failure interpretation, retained visible terminal observation, and fake parity for deterministic exit cases | Operator stop, retry, run lifecycle, escalation, orphan recovery | Harness profile presentation; Claude Code exchange; Stage 0 retained record path (vendor decoding is independently useful and need not precede process-exit observation) |

The readiness increment is separated from launch because a truthful refusal is
useful before an account is available and prevents an interactive login prompt
from becoming hidden process I/O. Raw real-vendor exchange precedes semantic
decoding because it is independently observable through the existing terminal
path. Native exit visibility consumes the launch seam but not decoded vendor
messages, so its task dependency must name that consumed surface rather than
infer an edge merely from the displayed order.

## Task coverage

| Increment | Coverage |
|---|---|
| Harness profile presentation | [P0-009](../phase-0/P0-009-plan-harness-profile-presentation.md) assigns independent derivation; implementation brief **not yet derived** |
| Vendor readiness | **Not yet derived** |
| Claude Code exchange | **Not yet derived** |
| Vendor output interpretation | **Not yet derived** |
| Native exit visibility | **Not yet derived** |

The next uncovered increment is **Harness profile presentation**. P0-009 assigns
a different task author to derive one implementation brief for it; no
implementation task ID has been allocated. Remaining actionable coverage stays
explicitly underived and is handed off one concrete brief at a time.

## Phase exit

From the loopback attach page, an authenticated single user selects the pinned
Claude Code profile, sees its concrete descriptor, submits input, observes
decoded output survive reload, and sees the native exit result and Claude-specific
interpretation. Malformed, oversized, schema-invalid, and arbitrarily chunked
vendor fixtures yield bounded operational diagnostics without exposing their
content, exhausting the decoder, or corrupting the next valid record. Repeating the
same exchange and exit cases with the selectable fake profile passes, including
its deterministic failure controls. Wrong-version, missing-binary, and missing-
login cases visibly refuse vendor launch without leaking credentials or content.

This demonstrates one pinned Claude Code instance through pipe transport and
fake parity. It does not demonstrate compatibility with later Claude versions,
a second vendor, generalized terminal behavior, credential isolation, denied
egress, multi-user safety, intentional stop, or orphan recovery.
