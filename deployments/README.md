# Sandbox deployment

Start the loopback-only sandbox with `docker compose -f deployments/compose.yml up
--build`. The image runs as an unprivileged user, includes `winch`,
`winch-sandbox`, and `fake-harness`, and serves the attach page on
`http://127.0.0.1:8080`.

The `container-standard` profile is a development posture: container process and
filesystem boundaries are present, but network egress is **not enforced**. This
configuration does not prove an egress allowlist, credential isolation, or protection
from a malicious harness.

For a machine without Go, `make test-cycle` uses the `toolchain` service to run the
format, vet, unit-test, and build gates. `make e2e` requires both Go and Docker and
drives the deployed image through its real entrypoint.
