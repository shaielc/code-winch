# P0-007: Standing end-to-end scenario suite

**Phase:** 0 — Foundation repair
**Shape:** capability
**Dependencies:** P0-001 (semantic: CI must run PostgreSQL-backed tests before this suite gates merges), P0-011 (semantic: every run operation and the live stream must exist before the standing scenario passes)

## Objective

A standing scenario suite drives the deployed system through
`create → start → stream → input → stop` on the all-fake profile, locally via
`make e2e` and in CI on every pull request.

## Scope

- Add `test/e2e/` with at least one scenario that:
  1. starts or connects to a running daemon with PostgreSQL;
  2. creates a run with `harnessProfile=fake` and `sandboxProfile=local`;
  3. starts the run;
  4. subscribes to the WebSocket event stream and observes ordered output from a
     scripted fake transcript (P0-003);
  5. sends one text input;
  6. stops the run and asserts a terminal state with no `fake-harness`
     descendant processes.
- Add a `make e2e` target documented in `deployments/README.md`.
- Wire `make e2e` into CI (extend the workflow from P0-001 or an equivalent step
  in this task's write set).
- Keep the scenario profile-specific under `test/e2e/` (one file per scenario
  per repository convention) so later swap tasks re-run the same scenario against
  more real substrates.

## Non-goals

- Browser automation or Playwright — HTTP and WebSocket clients are sufficient.
- Growing the suite beyond the standing scenario — new observable behavior adds
  scenarios in later capability tasks.
- Running the e2e suite against Codex, Docker sandbox, or in-memory storage.
- Replacing P0-004's `winch dev run` process test — that remains the CLI-path
  proof; this task owns the daemon entrypoint.

## Runtime reachability

- **Composition root:** `cmd/winchd` started by the e2e harness (native binary or
  compose stack).
- **Profile:** `harnessProfile=fake`, `sandboxProfile=local`, PostgreSQL storage.
- **Command:** `make e2e`.

## Write set

- `test/e2e/`
- `Makefile`
- `.github/workflows/go.yml` (if not fully wired by P0-001)
- `deployments/README.md`

## Contract surfaces

None. (The suite consumes existing API operations; it does not define new ones.)

## Demonstration

    $ make e2e
    → expect: scenario `create → start → stream → input → stop` passes against a
      daemon on the fake profile

On a pull request for this task:

    → expect: CI job log shows `make e2e` (or equivalent) passing

After the scenario completes:

    $ ps -eo pid,cmd | grep -c 'fake[-]harness'
    → expect: 0

## Verification

- `make check` passes without a database.
- `make e2e` passes with PostgreSQL available (same environment as
  `make test-integration`).
- CI runs the e2e suite on every push and pull request.

## Acceptance criteria

- [ ] `test/e2e/` exists and `make e2e` runs the standing scenario.
- [ ] The scenario uses the HTTP API and WebSocket stream, not `winch dev run`.
- [ ] Output is driven by a controllable fake transcript so the assertion is
  deterministic.
- [ ] CI executes the suite on every pull request.
- [ ] I4 holds: a standing end-to-end scenario suite exists against the
  all-fake profile and drives the real entrypoint.
- [ ] I1 and I2 still hold after the suite is added.

## Deferrals

| Deferred | Owning task |
|---|---|
| Additional scenarios for new observable behavior | Phase 1+ (not yet planned) |
| Running the same scenario against in-memory or Docker substrates | Phase 1+ swap tasks |

## Traces to

- Invariant I4 — `skills/shared/workplan-model.md`
- `docs/state.md` §*Invariants that were never established* (no standing scenario
  suite; `make e2e` never existed)
- `docs/code-structure.md` §1 (`test/e2e/`)
