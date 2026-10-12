# Sandbox deployment

Start the loopback-only sandbox with `docker compose -f deployments/compose.yml up
--build`. The image runs as an unprivileged user, includes `winch`,
`winch-sandbox`, and `fake-harness`, and serves the attach page on
`http://127.0.0.1:8080`.

The `container-standard` profile is a development posture: container process and
filesystem boundaries are present, but network egress is **not enforced**. This
configuration does not prove an egress allowlist, credential isolation, or protection
from a malicious harness.

The compose project is `code-winch`. Tests never use it: `compose.test.yml` layers an
isolated project over the same services, named with `-p` after the current git branch
(lowercased, other characters replaced by `-`; override with `TEST_PROJECT=...`). Worktrees
on different branches therefore run test environments side by side. The sandbox there
publishes no host port, so it cannot collide with a running dev sandbox, and it shares a
`test-network` with the `toolchain` service, which reaches it as `http://sandbox:8080`.

One shared proxy, `compose.proxy.yml` (project `code-winch-proxy`, nginx), fronts every
test project. It publishes only on `127.0.0.1:${PROXY_PORT:-7080}` and routes
`/<project>/sandbox/...` to that project's sandbox over the `code-winch-proxy` network;
`/<project>/sandbox` redirects to the trailing-slash form. The attach page builds with
relative asset URLs and derives its API base from `location`, so it works at the root or
under any `/{proxy/}+sandbox/` prefix.

- `make test-cycle` is the whole path and needs only Docker: it starts the test
  environment, runs the format, vet, unit-test, and build gates plus the e2e scenarios
  inside the `toolchain` service, and always tears the environment down.
- `make test-env` starts the proxy if needed, then the test environment, and checks that
  the sandbox is healthy, runs as a non-root user, and that the production and proxy
  compose files pin the published host IP to `127.0.0.1`. `make test-env-down` removes
  this branch's environment and leaves the shared proxy; `make proxy-down` removes it.
- `make e2e` runs the scenarios with the Go wherever make runs. On the host it targets
  `http://127.0.0.1:$PROXY_PORT/<project>/sandbox`; override with `WINCH_E2E_URL`.
  `make docker-e2e` runs them in the `toolchain` service on the test network. Neither
  starts or stops anything.

The test environment cannot show that the host's routable address refuses the port,
because nothing is published there. That check stays manual against the development
project: start it, confirm a proxy-disabled loopback request succeeds, then confirm
a proxy-disabled request to the host's routable IP is refused while loopback remains
healthy. For example:

```sh
docker compose -f deployments/compose.yml up --build -d
curl --noproxy '*' -fsS http://127.0.0.1:8080/healthz
host_ip=$(hostname -I | awk '{print $1}')
curl --noproxy '*' -fsS --max-time 2 "http://${host_ip}:8080/healthz"
# Expect connection refusal (curl exit 7); loopback must still answer.
curl --noproxy '*' -fsS http://127.0.0.1:8080/healthz
docker compose -f deployments/compose.yml down
```

This host-local check does not replace the off-host scan required by LB11.
[`docs/state.md`](../docs/state.md) records the closure's source evidence and runtime
verification limits.
