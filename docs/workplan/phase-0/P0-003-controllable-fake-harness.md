# P0-003: Controllable fake harness profile

**Phase:** 0 — Foundation repair
**Shape:** capability
**Dependencies:** None

## Objective

A person can drive the fake harness with a scripted transcript and inject latency,
failure, malformed output, or early exit at runtime without editing source code.

## Scope

- Add runtime configuration for the fake harness profile (transcript file path,
  optional delay, forced failure, malformed line injection, early exit) wired
  through `cmd/fake-harness` flags and/or environment variables documented in
  `--help`.
- Teach `internal/adapters/harness/fake` to pass those controls through
  `BuildLaunch` so the profile works when the daemon eventually binds runs (the
  adapter change lands here even if the daemon does not consume it until Phase
  1).
- Resolve the bare `fake-harness` name on `PATH`: accept an explicit binary
  path in configuration or embed the resolved path in `LaunchSpec` so a host
  without a manual `go build` still works when the binary is installed by the
  deployment image.
- Document what the fake profile does **not** prove (no real provider, no network,
  no credential handling).

## Non-goals

- A second harness adapter or Codex integration.
- UI controls for transcript selection.
- Daemon composition-root wiring for runs (Phase 1).

## Runtime reachability

- **Composition root:** `cmd/winch` via `winch dev run --harness fake` (immediate);
  `cmd/winchd` once run binding lands (adapter ready, not required to demonstrate
  here).
- **Profile:** `harnessProfile=fake` with `sandboxProfile=local` or `fake`.
- **Command:** `winch dev run --harness fake --sandbox local` with transcript
  flags or env vars.

## Write set

- `cmd/fake-harness/main.go`
- `cmd/winch/main.go` (transcript and injection flags on `dev run`)
- `internal/adapters/harness/fake/fake.go`
- `internal/platform/config/` (only if a configuration key is the chosen surface)
- `deployments/README.md` (fake profile controls)
- Tests under `internal/adapters/harness/fake/` and/or `cmd/fake-harness/`

## Contract surfaces

- configuration: fake harness transcript and control keys (if added)
- driver namespace: `harnessProfile=fake` launch arguments
- `LaunchSpec` command/args for the fake harness binary

## Demonstration

Create a transcript file `/tmp/fake-transcript.txt`:

    echo hello from transcript
    fail

Run with the new controls:

    $ winch dev run --harness fake --sandbox local \
        --fake-transcript /tmp/fake-transcript.txt
    → expect: "hello from transcript" appears, then a nonzero exit with a stable
      harness exit code

Inject latency (exact flag names as implemented):

    $ winch dev run --harness fake --sandbox local \
        --fake-delay 500ms --fake-transcript /tmp/slow.txt
    → expect: observable delay before output

Malformed output mode:

    $ winch dev run --harness fake --sandbox local --fake-malformed-line
    → expect: harness or codec reports a structured error without dumping secrets

## Verification

- `make check` passes.
- New unit tests cover transcript playback and at least one injectable failure mode.
- Contract suite for `internal/adapters/harness/fake` still passes.

## Acceptance criteria

- [ ] A transcript file selected at runtime drives command/output sequence without
  recompiling.
- [ ] Injectable latency, failure, malformed output, and early exit are documented
  and covered by tests.
- [ ] `BuildLaunch` no longer depends on the operator having manually placed
  `fake-harness` on `PATH` when the binary is installed by the deployment image.
- [ ] I3 holds for the fake harness profile: controllable, documented, and honest
  about limits.

## Deferrals

| Deferred | Owning task |
|---|---|
| Daemon-backed runs using the fake profile | P0-008 |
| Controllable in-memory store profile | Phase 1 — registered in [`../phase-1/README.md`](../phase-1/README.md) |

## Traces to

- Invariant I3 — `skills/shared/workplan-model.md`
- `docs/state.md` §*Invariants that were never established* (fake profile not
  controllable; bare `PATH` launch)
