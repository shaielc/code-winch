# P1-053: Establish the standing scenario suite

**Phase:** 1 — Local single-user vertical slice
**Shape:** capability
**Dependencies:** P1-050 (semantic: the suite drives the real HTTP and WebSocket entrypoints), P1-052 (semantic: reconnect, restart, and stop scenarios need a harness that can be made to stall, stream, and refuse to exit)

## Objective

One command runs the scenarios that define Phase 1's exit condition against a
deployed daemon, and every later substrate swap re-runs them unchanged.

## Scope

- `test/e2e`: scenarios driving the deployed system through its real entrypoints
  — the HTTP API and the event WebSocket — never through internal packages.
- The standing set:
  1. **create → start → stream → input → stop** against the all-fake profile;
  2. **reconnect**: drop the socket mid-stream, resume with `after_sequence`,
     assert no gap and no duplicate against the persisted history;
  3. **restart**: kill the daemon during a live run, restart, assert the
     reported state is truthful and accepted input is delivered exactly once;
  4. **forced stop**: a harness that ignores `SIGTERM` is escalated and leaves
     no descendant process;
  5. **malformed output**: the run survives and records a diagnostic.
- A profile switch so the same suite runs against a substrate under test —
  named profiles, not conditionals inside scenarios.
- CI job running the suite on the all-fake profile for every pull request.
- `docs/workplan/README.md` gains the one-line command every later brief cites.

## Non-goals

- Browser automation. These scenarios are API-level; the web app's own tests
  stay in `web/`.
- Growing the suite for behavior that is not new. It grows only when a task
  introduces genuinely new observable behavior.

## Runtime reachability

`make e2e` against the compose stack; the same target in CI.

## Owned surfaces

`test/e2e/`, `.github/workflows/e2e.yml`, `Makefile` (`e2e` target), the
"standing scenario suite" section of `docs/workplan/README.md`.

## Demonstration

    $ docker compose -f deployments/compose.yml up --build -d
    $ make e2e
    → expect: five scenarios pass, each naming the profile it ran against

    $ make e2e PROFILE=fake-sandbox
    → expect: the same five scenarios pass against a different substrate,
      with no scenario source change

    $ docker compose -f deployments/compose.yml kill winchd   # during scenario 3
    → expect: the scenario, not the harness, decides the daemon died; the run's
      reported state after restart matches what actually happened

## Verification

- The suite itself is the verification. It must fail if the daemon is not
  running, rather than skipping.
- A deliberately broken build is run once to confirm each scenario fails for the
  right reason and reports which step failed.

## Acceptance criteria

- [ ] All three clauses of `docs/roadmap.md` Phase 1's exit statement are
      demonstrated by an automated scenario.
- [ ] Scenarios contain no conditional on which substrate is active.
- [ ] The suite runs in CI on every pull request against the all-fake profile.
- [ ] A scenario failure names the step and the run ID, and includes no event
      content in its output.

## Deferrals

| Deferred | Owning task |
|---|---|
| Scenario asserting structured events rather than raw output | P2-021 |
| Running the suite against the Docker substrate | P3-029 |
| Running the suite against a remote runner | P5-042 |

## Traces to

`docs/roadmap.md` Phase 1 exit; `docs/code-structure.md` §7;
`docs/contracts.md` §5
