# Sandbox deployment

Start the loopback-only sandbox with `docker compose -f deployments/compose.yml up
--build`. The image runs as an unprivileged user, includes `winch`,
`winch-sandbox`, and `fake-harness`, and serves the attach page on
`http://127.0.0.1:8080`.

The sandbox starts the `fake` harness profile itself. Its output appears live on the
attach page and through `./bin/winch stream`. The profile can replay a file with
`WINCH_HARNESS_TRANSCRIPT`, delay transcript actions with `WINCH_HARNESS_DELAY` (a Go
duration such as `250ms`), and inject failure, early exit, or malformed output with
`WINCH_HARNESS_FORCE_FAILURE=1`, `WINCH_HARNESS_EARLY_EXIT=1`, and
`WINCH_HARNESS_MALFORMED_LINE=1`.

The fake does **not** prove the behavior of a coding agent. It has no tool calls,
approvals, usage reporting, login, or terminal semantics. Its JSON-lines dialect is
invented, so an event model tested only against it would be modelling this fixture.

The `container-standard` profile is a development posture: container process and
filesystem boundaries are present, but network egress is **not enforced**. This
configuration does not prove an egress allowlist, credential isolation, or protection
from a malicious harness.

The compose project is `code-winch`. Tests never use it: `compose.test.yml` layers an
isolated project, `code-winch-test`, over the same services. The sandbox there publishes
no host port, so it cannot collide with a running dev sandbox, and it shares a
`test-network` with the `toolchain` service, which reaches it as `http://sandbox:8080`.

- `make test-cycle` is the whole path and needs only Docker: it starts the test
  environment, runs the format, vet, unit-test, and build gates plus the e2e scenarios
  inside the `toolchain` service, and always tears the environment down.
- `make test-env` starts the test environment and checks that the sandbox is healthy,
  runs as a non-root user, and that the production compose file pins the published host
  IP to `127.0.0.1`. `make e2e` runs the scenarios against it, and
  `make test-env-down` removes it. `make e2e` does not start or stop anything.

The test environment cannot show that the host's routable address refuses the port,
because nothing is published there. That check stays manual: see the Demonstration in
`docs/workplan/phase-0/P0-001-sandbox-serves-attach-page.md`.
