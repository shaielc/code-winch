# P2-025: Implement second harness adapter

**Phase:** 2 — Structured experience and second harness
**Shape:** swap
**Dependencies:** P2-021 (contract: both adapters map onto the same normalization helpers), P1-049 (compile: the runner and sandbox attach the adapter is launched through)

## Objective

The same web views and the same standing scenarios drive a second provider, with
no provider conditional in a generic component. If a generic component needs a
provider name to behave correctly, this task has failed even if the run works.

## Scope

- A second harness package: descriptor with stable ID, adapter version,
  input/output modes, login modes, and capabilities; launch construction;
  incremental codec; exit mapping; namespaced extension schema for
  provider-only data.
- Registration through P1-050's append-only registry — no edit to a central
  switch.
- A recorded, sanitized transcript fixture so CI exercises the codec without a
  vendor account, plus the shared harness contract suite.
- Capability-driven UI: controls the second provider does not support are
  disabled with the reason, rather than hidden or silently inert.
- A capability matrix in the docs comparing both adapters honestly.

## Non-goals

- A lowest-common-denominator UI. Provider-only data stays available through its
  extension namespace.
- Third-party or plugin adapters — deferred in `docs/roadmap.md`.
- Login flows for the second provider — P2-058.

## Runtime reachability

`winch run create --harness <second>` on the compose stack; the same run page
and the same standing scenarios.

## Owned surfaces

`internal/adapters/harness/<second>/`, its `testdata/`,
`schemas/events/v1/extensions/<second>/`, `docs/harness-capabilities.md`.

## Demonstration

    $ winch profiles
    → expect: two harnesses listed with differing capability sets

    $ make e2e HARNESS=<second>
    → expect: the same standing scenarios pass, unchanged, against the second
      adapter

    # in the browser, one run per provider, side by side:
    → expect: identical conversation and activity components; a control the
      second provider lacks is disabled and states why

    $ grep -rn '<second>' web/src/renderers web/src/features/runs
    → expect: no match outside a capability lookup

## Verification

- Standing scenario suite passes against the second adapter and still passes
  against the first.
- Shared harness contract suite passes for both adapters.
- Codec tests over the sanitized transcript with adversarial chunk boundaries.
- A test asserting the transcript fixture contains no credential-shaped value.

## Acceptance criteria

- [ ] Both adapters pass one contract suite with no adapter-specific exemption.
- [ ] Generic web components contain no provider conditional.
- [ ] Unsupported controls are disabled with a stated reason.
- [ ] Provider-only data survives in its extension namespace and is ignored by
      generic consumers.
- [ ] CI needs no vendor account.

## Deferrals

| Deferred | Owning task |
|---|---|
| Login for this provider | P2-058 |
| Running this adapter under Docker isolation | P3-029 |

## Traces to

`docs/code-structure.md` §4; `docs/architecture.md` §4 (harness adapters);
`docs/decisions/0003-capability-based-adapters.md`; `docs/roadmap.md` Phase 2 exit
