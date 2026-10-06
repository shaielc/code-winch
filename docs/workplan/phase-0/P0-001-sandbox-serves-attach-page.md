# P0-001: The sandbox image starts, serves its attach page, and states its posture

**Phase:** 0 — A sandbox you can talk to
**Shape:** seam
**Dependencies:** None

## Objective

`docker compose -f deployments/compose.yml up --build` produces one non-root sandbox
container whose port is published only on host loopback, whose authenticated attach page
names the effective sandbox profile and its unenforced egress control, and whose source
tree passes every repaired repository quality gate.

## Scope

- Build `cmd/winch-sandbox` as the first composition root. It loads configuration, wires
  the attach transport, starts one `http.Server`, and shuts it down on `SIGINT` or
  `SIGTERM`. Keep routing and response construction in
  `internal/adapters/transport/attach`; the composition root supplies values and owns
  process lifetime.
- Resolve configuration from compiled defaults followed by environment overrides, as
  `docs/code-structure.md` §6 requires:
  `WINCH_SANDBOX_ADDR` (default `127.0.0.1:8080`),
  `WINCH_SANDBOX_STATIC_DIR` (default `/opt/winch/web`),
  `WINCH_SANDBOX_PROFILE` (default `container-standard`), and
  `WINCH_SANDBOX_TOKEN` (no process default). Fail before listening when the token is
  empty or an address/profile is invalid. The compose service sets the address to
  `0.0.0.0:8080`, because binding loopback *inside* a container is not reachable through
  Docker's port publishing; compose, not the process bind, restricts the published port
  to host `127.0.0.1`.
- Serve unauthenticated `GET /healthz` as JSON
  `{"service":"winch-sandbox","status":"ok"}`. Serve the static page shell at `/`,
  including history fallback, without authentication so it can collect a token. Protect
  `GET /api/session` with `Authorization: Bearer <token>` and return JSON
  `{"profile":"container-standard","unenforced_controls":["egress"]}` for the
  Stage 0 compose profile. Missing, malformed, and incorrect credentials return the
  same `401` response and never echo either credential. This is the single
  session-scoped token required by `docs/security.md` §§2–3; full account and object
  authorization remain outside this stage.
- Apply attach-origin validation at the transport boundary: browser requests carrying
  an `Origin` header are accepted only when its scheme and host equal the request's
  effective origin; requests without `Origin` remain available to the operator CLI.
  Reject disallowed origins before authorization. Bound request headers and server
  header/read/write/idle timeouts so this first surface does not establish an unbounded
  default that later input and streaming handlers inherit.
- Add focused Go tests under `internal/adapters/transport/attach/` for the exact health
  and posture documents, static fallback, authorization failures, disallowed origins,
  and content-free failure bodies. Add `test/contract/attach/` tests for the two HTTP
  operations and their security behavior; this local attach protocol, rather than the
  absent control-plane OpenAPI tree, pins the contract introduced here.
- Create `web/` as a TypeScript/React workspace with a committed lockfile and scripts
  `format:check`, `lint`, `typecheck`, `test`, and `build`. `web/src/attach/` is its only
  entry point. The page accepts the session token without sending it in a URL query
  (a `#token=...` fragment may seed it), keeps it in memory, sends it in the bearer
  header, and renders explicit loading, authorization-error, and loaded states. The
  loaded state shows both the returned profile and every unenforced control. Do not
  log or persist the token. Remove the ignored stale `web/dist/`, `web/node_modules/`,
  and `web/*.tsbuildinfo` artifacts before verifying from a clean install.
- Add `cmd/winch`, organized as one file per command, as the scriptable twin of the
  attach surface. `winch status` accepts `--address` and `--token`, with
  `http://127.0.0.1:8080` and `WINCH_SANDBOX_TOKEN` as their respective defaults, calls
  both endpoints with bounded client timeouts, and prints health, profile, and the
  unenforced controls. The CLI does not reuse `WINCH_SANDBOX_ADDR`: that value is a
  server listen address and compose deliberately sets it to an unspecified container
  address that is not a valid operator destination.
  It exits nonzero with a content-free diagnostic for an unavailable service or a
  refused credential. Later commands extend this client rather than define another
  protocol.
- Add `deployments/Dockerfile` with Node, Go, and minimal non-root runtime stages. Node
  builds the page; Go builds `winch-sandbox`, `winch`, and the existing `fake-harness`;
  the runtime image contains only the three binaries, built web assets, required CA
  material, and a non-root user. The harness ships but is not started in this task, so
  the image already satisfies the Stage 0 packaging boundary without claiming runner
  behavior.
- Add `deployments/compose.yml` with a `sandbox` service whose only published socket is
  `127.0.0.1:8080:8080`, and a `toolchain` service under the `test` profile that mounts
  the source tree and supplies Go for Docker-only checks. Compose supplies a documented
  local-development token when `WINCH_SANDBOX_TOKEN` is unset; it is a convenience for
  this loopback-only profile, is visibly labelled as such, and is not a production
  credential. Document startup, token override, browser/CLI attachment, tests, and the
  profile's unproved isolation in `deployments/README.md`.
- Add the standing suite in `test/e2e/`: a shared helper brings up the composed image,
  waits with a deadline, and always tears it down; the first scenario proves health,
  authorized posture, rejected unauthorized posture, page asset delivery, the
  non-root UID, and the host-loopback port mapping. Commands must select a collision-free
  compose project name and report container logs when startup fails.
- Repair gates to match this tree: in `Makefile`, remove the absent OpenAPI, PostgreSQL,
  daemon, and integration targets; make `build`, `run`, `e2e`, and the Docker group use
  `winch-sandbox`, `sandbox`, and `toolchain`; include attach contract tests through the
  ordinary Go packages. In `.github/workflows/go.yml`, use `test` (not `tests`), remove
  PostgreSQL/integration setup, and run Docker e2e after `make check`. In
  `.github/workflows/web.yml`, use `npm ci` and remove the absent `api:check` script.
  Correct `.dockerignore` to name `test/`, not absent `runner/` or `tests/`, and extend
  `.gitignore` only as the new workspace requires.
- Update the now-stale gate and run instructions in `README.md` and `AGENTS.md`: once
  this implementation lands, they must describe the repaired targets, the Docker-only
  route, and the commands that actually work rather than retaining the current
  “product does not run” warning.

## Non-goals

- Starting, reading, or writing a harness process. The image carries the existing fake
  binary but this task does not claim that the runner owns it. Roadmap Stage 0's next
  output increment owns that behavior.
- Any record, ordinal, stream, input operation, or store. Those are later closed loops
  within roadmap Stage 0; this task defines only health and posture.
- The control-plane OpenAPI document. `api/openapi/` belongs to roadmap Stage 3. The
  standalone sandbox contract is pinned by `test/contract/attach/` meanwhile.
- Accounts, cross-session/object authorization, TLS termination, or exposure beyond
  loopback. Roadmap Stage 0 deliberately has one host operator and one session; Stage 5
  must satisfy `docs/security.md` §11 before shared exposure. The one attach token and
  same-origin check required by the existing design are not deferred.
- Enforcing egress policy, filesystem isolation beyond the container baseline,
  credentials, or hostile multi-tenant isolation. Roadmap Stage 5 owns enforcement;
  this task must accurately disclose its absence.
- PostgreSQL, migrations, `winchd`, runs, and run lifecycle. Roadmap Stages 3 and 4.

## Runtime reachability

`cmd/winch-sandbox` is the composition root. The `sandbox` compose service runs it with
the `container-standard` profile, a session token, and a container-wide listener whose
published host socket is loopback-only. `docker compose -f deployments/compose.yml up
--build` reaches all production code introduced here; a browser and `winch status`
reach the attach transport through that published socket.

## Write set

- `cmd/winch-sandbox/main.go`, `cmd/winch-sandbox/config.go`,
  `cmd/winch-sandbox/config_test.go`
- `internal/adapters/transport/attach/server.go`, `session.go`, `static.go`,
  `server_test.go`
- `cmd/winch/main.go`, `cmd/winch/client.go`, `cmd/winch/status.go`,
  `cmd/winch/status_test.go`
- `web/package.json`, `web/package-lock.json`, `web/tsconfig*.json`,
  `web/vite.config.ts`, `web/eslint.config.js`, `web/index.html`,
  `web/src/attach/main.tsx`, `web/src/attach/App.tsx`,
  `web/src/attach/Posture.tsx`, `web/src/attach/App.test.tsx`
- `deployments/Dockerfile`, `deployments/compose.yml`, `deployments/README.md`
- `test/contract/attach/attach_test.go`, `test/e2e/compose.go`,
  `test/e2e/scenario_sandbox_starts_test.go`
- `Makefile`, `.github/workflows/go.yml`, `.github/workflows/web.yml`,
  `.dockerignore`, `.gitignore`
- `README.md`, `AGENTS.md`

## Contract surfaces

- API: unauthenticated `GET /healthz` and its service/status document
- API: bearer-protected `GET /api/session` and its profile/unenforced-controls document
- attach authorization: `Authorization: Bearer`, same-origin browser requests, and
  indistinguishable missing/invalid-token failures
- config keys: `WINCH_SANDBOX_ADDR`, `WINCH_SANDBOX_STATIC_DIR`,
  `WINCH_SANDBOX_PROFILE`, `WINCH_SANDBOX_TOKEN`
- CLI namespace: `winch`, `winch status`, and its `--address`/`--token` options
- compose service names `sandbox` and `toolchain`, and the `test` compose profile
- the `Makefile` target set and the Go/web workflow job definitions

## Demonstration

    $ export WINCH_SANDBOX_TOKEN='p0-local-demo-token'
    $ docker compose -f deployments/compose.yml up --build -d
    $ curl -fsS http://127.0.0.1:8080/healthz
    → expect: {"service":"winch-sandbox","status":"ok"}

    $ curl -sS -o /tmp/session-without-token -w '%{http_code}\n' \
        http://127.0.0.1:8080/api/session
    → expect: 401, with no token or profile in the response body

    $ ./bin/winch status --address http://127.0.0.1:8080 \
        --token "$WINCH_SANDBOX_TOKEN"
    → expect: healthy winch-sandbox; profile container-standard; egress is not enforced

    Browser: navigate to http://127.0.0.1:8080/#token=p0-local-demo-token
    → expect: the page renders the same profile and unenforced egress control; the token
      is absent from HTTP request targets and browser persistent storage

    $ docker compose -f deployments/compose.yml port sandbox 8080
    → expect: 127.0.0.1:8080 (or the IPv6 loopback equivalent), never 0.0.0.0

    $ docker compose -f deployments/compose.yml exec -T sandbox id -u
    → expect: a non-zero UID

    $ docker compose -f deployments/compose.yml down --remove-orphans

## Verification

- `make check` passes the repaired host gates: `format-check`, `vet`, `lint`, `test`,
  and `build`, with no target reading a path absent from the tree.
- `make e2e` passes `test/e2e/scenario_sandbox_starts_test.go` against a freshly built
  composed image, including the authorization, non-root, page, posture, and published
  socket assertions.
- `make test-cycle` passes with Docker as its only host prerequisite by using the
  `toolchain` service, and always tears the test topology down.
- `go test ./internal/adapters/transport/attach ./cmd/winch-sandbox ./cmd/winch
  ./test/contract/attach` passes, including invalid-token and hostile-origin cases.
- `cd web && npm ci && npm run format:check && npm run lint && npm run typecheck && npm
  test && npm run build` passes from an absent `node_modules` and `dist`.
- `python3 -m unittest discover -s test` and
  `(cd workplan-control-panel && python3 -m unittest discover -s tests)` still pass.
- `.github/workflows/go.yml` and `.github/workflows/web.yml` pass on the pull request.
- Run the Demonstration exactly as written and record the observed output in the
  implementation pull request.

## Acceptance criteria

- [ ] `docker compose -f deployments/compose.yml up --build -d` reaches healthy without
      a source checkout mounted into `sandbox`; a startup test also proves an invalid
      address, empty token, or missing built-static directory fails clearly before a
      misleading healthy response can be served.
- [ ] `/healthz` returns the exact documented healthy response, while a stopped server
      makes both `curl` and `winch status` fail within their configured deadlines rather
      than hang or report stale health.
- [ ] `GET /api/session`, the attach page, and `winch status` all name
      `container-standard` and `egress` as unenforced. A contract test fails if either
      field is absent, empty, renamed, or if the UI hides the unenforced control.
- [ ] Missing, malformed, and incorrect bearer credentials receive indistinguishable
      `401` responses with no credential or posture disclosure; a hostile browser
      `Origin` is rejected, while the no-`Origin` CLI request with the correct token
      succeeds.
- [ ] The browser can seed its token from the URL fragment, removes it from the visible
      URL after reading, keeps it out of query strings and persistent storage, and shows
      an explicit authorization error instead of falsely presenting an isolated profile.
- [ ] The runtime container contains all three required binaries and the page, runs with
      a non-zero UID, and does not start `fake-harness` in this task.
- [ ] Docker reports the sandbox's published port on loopback only. The process listens
      on the compose-configured container address, so loopback restriction does not
      accidentally make the page unreachable through port publishing.
- [ ] With `WINCH_SANDBOX_STATIC_DIR` omitted from the process environment, the image
      serves the page from `/opt/winch/web`; overriding it to an absent directory fails
      at startup with a content-free diagnostic.
- [ ] `make check`, `make e2e`, `make test-cycle`, the clean-install web command, both
      Python suites, and both CI workflows pass. Every path named by the Makefile and
      the two workflows exists when its command runs.
- [ ] `README.md` and `AGENTS.md` no longer call the repaired gates or product startup
      broken, and their documented host and Docker commands match the working targets.
- [ ] I1 and I2 hold at HEAD: this change leaves a manually startable, deployable,
      authenticated, loopback-only vertical slice, and its standing scenario exercises
      the deployed slice rather than an in-process substitute.

## Deferrals

None.

## Traces to

`docs/roadmap.md` §3 Stage 0 (including the floor repair) and §5;
`docs/architecture.md` §5 Topology A; `docs/code-structure.md` §1, §6;
`docs/contracts.md` §8 (*The surface states its posture*);
`docs/security.md` §2 (browser → attach surface), §3 T02 and T15, §4, §6, §11 LB11;
ADR-0005, ADR-0006; `docs/state.md` *Every quality gate names a path that is gone*.
