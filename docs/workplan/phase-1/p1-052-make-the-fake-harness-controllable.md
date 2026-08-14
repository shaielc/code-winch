# P1-052: Make the fake harness controllable

**Phase:** 1 — Local single-user vertical slice
**Shape:** capability
**Dependencies:** P0-008 (compile: the fake harness binary and codec this extends)

## Objective

A person can make the fake harness produce a chosen transcript, stall, emit
malformed output, or die mid-stream, by naming a scenario file — turning the
fake from a smoke test into a driveable runtime profile.

## Scope

- A scenario file format: an ordered list of steps, each emitting output,
  waiting for input matching a pattern, sleeping, exiting, or emitting a
  deliberately malformed record. Documented with a schema and fixtures.
- `fake-harness --scenario <file>` plus fault flags: injected latency per step,
  failure on the Nth input, truncated or oversized records, early exit without
  flush, and refusal to respond to `SIGTERM` for a bounded interval.
- Deterministic seeding by default with an explicit way to turn it off, so a
  scenario is reproducible when a test wants that and varied when exploration
  does.
- Ship the scenarios later tasks rely on: a plain echo session, a long-running
  streaming session, a session that exits non-zero, a session that ignores
  graceful stop, and a session emitting a malformed record mid-stream.
- Document what this profile does **not** prove: it is not a provider, it has
  no model latency distribution, no rate limits, no authentication, and its
  output shapes are the ones we chose to write.

## Non-goals

- Recording or replaying a real provider transcript — sanitized recordings are
  the second adapter's concern (P2-025).
- Structured tool-call and approval payloads, which do not exist as canonical
  events until P2-021.

## Runtime reachability

`fake-harness --scenario` by hand; `winch dev run --scenario` (P1-049);
`WINCH_HARNESS_SCENARIO` on the compose stack once P1-050 passes launch
configuration through.

## Owned surfaces

`cmd/fake-harness/`, `internal/adapters/harness/fake/`,
`schemas/scenarios/v1/`, `test/scenarios/*.json`.

## Demonstration

    $ fake-harness --scenario test/scenarios/echo.json
    > hello
    → expect: the scripted reply, not a generic echo

    $ fake-harness --scenario test/scenarios/malformed.json | head -5
    → expect: one record that is deliberately not valid, framed so a codec can
      detect it rather than hang

    $ fake-harness --scenario test/scenarios/ignores-sigterm.json &
    $ kill -TERM %1; sleep 1; pgrep -f fake-harness
    → expect: still alive, so stop escalation has something real to escalate against

    $ fake-harness --scenario test/scenarios/echo.json --seed 7 > a
    $ fake-harness --scenario test/scenarios/echo.json --seed 7 > b && diff a b
    → expect: no difference

## Verification

- Scenario schema validation tests, including rejection of an unknown step kind.
- Codec tests asserting the malformed record degrades to a diagnostic event and
  does not terminate parsing.
- A test asserting `--seed` reproducibility and that seeding off produces
  variation.

## Acceptance criteria

- [ ] Every fault listed in Scope is reachable from a scenario file or a flag.
- [ ] Scenarios are data files, so adding one does not change Go code.
- [ ] The fake's documented limits are written down in the same change.
- [ ] Existing bare-text behavior still works when no scenario is given.

## Deferrals

| Deferred | Owning task |
|---|---|
| Structured message, tool-call, and approval scenario steps | P2-021 |
| Scenarios exercising two adapters against one suite | P2-025 |

## Traces to

`docs/code-structure.md` §4, §7; `docs/roadmap.md` Phase 0 exit;
`docs/decisions/0003-capability-based-adapters.md`
