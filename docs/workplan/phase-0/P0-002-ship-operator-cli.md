# P0-002: Ship operator CLI in build and deployment

**Phase:** 0 — Foundation repair
**Shape:** hardening
**Dependencies:** None

## Objective

The maintained operator CLI (`winch`) is produced by `make build` and installed on
`PATH` in the deployment image alongside `winchd` and `fake-harness`.

## Scope

- Extend `make build` (or add a clearly documented sibling target that `build`
  depends on) to compile `./cmd/winch` to a predictable output name.
- Install `winch` into the runtime image in `deployments/Dockerfile.winchd`.
- Update `deployments/README.md` to describe how operators invoke `winch` inside
  the stack (for example via `docker compose exec`).

## Non-goals

- New CLI subcommands beyond what already exists (`dev run`).
- Publishing `winch` as a separate image or package.
- Wiring run use cases through the CLI — HTTP binding is Phase 1.

## Runtime reachability

- **Composition root:** `cmd/winch` (standalone; no daemon required for
  `dev run`).
- **Profile:** host build tree and `deployments/compose.yml` runtime image.
- **Command:** `winch dev run --harness fake --sandbox local`.

## Write set

- `Makefile`
- `deployments/Dockerfile.winchd`
- `deployments/README.md`

## Contract surfaces

None.

## Demonstration

    $ make build
    → expect: `winch` and `winchd` binaries are produced (document the output paths in the PR)

Build and inspect the image:

    $ docker build -f deployments/Dockerfile.winchd -t winchd:test .
    $ docker run --rm winchd:test which winch
    → expect: /usr/local/bin/winch

Inside a running compose stack (after `docker compose up`):

    $ docker compose -f deployments/compose.yml exec winchd winch --help
    → expect: usage text for the operator CLI

## Verification

- `make check` passes.
- `deployments/README.md` describes the operator path without a "not wired yet"
  caveat for the CLI itself.

## Acceptance criteria

- [ ] `make build` compiles `winch` in addition to `winchd`.
- [ ] The deployment image installs `winch` on `PATH`.
- [ ] `winch dev run --harness fake --sandbox local` works from a container that
  includes the built binaries (stdin script with `echo` / `exit` is sufficient).
- [ ] I1 and I2 still hold: the daemon starts from the same image and serves
  `/api/v1/health`.

## Deferrals

| Deferred | Owning task |
|---|---|
| CLI commands for daemon-backed runs (`create`, `start`, …) | Phase 1 (not yet planned) |

## Traces to

- Invariant I5 — `skills/shared/workplan-model.md`
- `docs/state.md` §*Invariants that were never established* (operator CLI absent
  from `make build` and deployment image)
