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
isolated project over the same services. The Make targets name it with `-p`, as
`code-winch-test-<current branch>` (the short SHA on a detached HEAD), so each branch or
worktree gets its own environment and they can run side by side. Override the name with
`TEST_NAME=<name>`, and pass the same value to every target that touches that
environment (`test-env`, `docker-e2e`, `test-env-down`, `test-cycle`). Switching branches
with an environment up changes the default name, so stop it first or pass the old name.
A test sandbox publishes no host port, so it cannot collide with a running dev sandbox. It
shares a `test-network` with the `toolchain` service, which reaches it as
`http://sandbox:8080`, and it joins the `winch-proxy` network so the proxy below can reach it.

- `make test-cycle` is the whole path and needs only Docker: it starts the test
  environment, runs the format, vet, unit-test, and build gates plus the e2e scenarios
  inside the `toolchain` service, and always tears the environment down.
- `make test-env` starts the test environment and checks that the sandbox is healthy,
  runs as a non-root user, and that the production compose file pins the published host
  IP to `127.0.0.1`. It does not start the proxy; it prints the proxy URL for the
  environment, which works only while the proxy is up. `make e2e` runs the scenarios from
  the host against it (`WINCH_E2E_URL=<that URL> make e2e`, needs the proxy),
  `make docker-e2e` runs them in the `toolchain` service on the test network (no proxy
  needed), and `make test-env-down` removes the environment. Neither e2e target starts or
  stops anything.

### Proxy

For the browser and host tools, `make proxy-up` starts one nginx container (project
`code-winch-proxy`) that publishes `127.0.0.1:8088` (`PROXY_PORT=<port>` changes it) and
routes by name. It resolves upstreams through Docker's DNS on every request, so
environments come and go without reconfiguring it:

- `/test/<name>/app` reaches the `code-winch-test-<name>` sandbox. `winch status|stream
  --url http://127.0.0.1:8088/test/<name>` and
  `WINCH_E2E_URL=http://127.0.0.1:8088/test/<name> make e2e` use the same base.
- `/main/app` reaches the dev sandbox when it is started with
  `docker compose -f deployments/compose.yml -f deployments/compose.routed.yml up -d`.

Only names matching `[a-z0-9_-]+` are routed, and an environment that is not running
answers 502. `make proxy-down` removes the proxy. The `winch-proxy` network is declared by
both the proxy and the sandbox overlay, so Compose creates it for whichever starts first;
starting the proxy second prints a harmless "network exists but was not created for
project" warning.

The test environment cannot show that the host's routable address refuses the port,
because only the proxy publishes a port, on loopback. That check stays manual: see the Demonstration in
`docs/workplan/phase-0/P0-001-sandbox-serves-attach-page.md`.
