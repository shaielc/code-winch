# P0-001: The sandbox image starts, serves its attach page, and states its posture

**Phase:** 0 — A sandbox you can talk to
**Shape:** seam
**Dependencies:** None

## Objective

`docker compose -f deployments/compose.yml up --build` produces a container bound to
loopback that serves a page naming the sandbox profile in force and what that profile
does not isolate, and every quality gate in the repository passes against the tree that
builds it.

## Scope

- `cmd/winch-sandbox/` — the first composition root. It resolves configuration before
  constructing one `http.Server`, registers the attach transport, and shuts down on
  `SIGINT`/`SIGTERM`. The transport owns `GET /healthz`, `GET /api/session`, and the
  static-page fallback; keeping route construction outside `main` makes the real HTTP
  boundary testable without inventing an application port before a second case exists.
- Configuration is compiled defaults overridden by non-empty environment values, per
  `docs/code-structure.md` §6: `WINCH_SANDBOX_ADDR` defaults to `127.0.0.1:8080`,
  `WINCH_SANDBOX_STATIC_DIR` defaults to the runtime image's `/opt/winch/web`, and
  `WINCH_SANDBOX_PROFILE` defaults to `container-standard`. Invalid addresses, unknown
  profiles, and a static directory with no `index.html` fail startup with a content-free
  diagnostic; an absent or empty override selects the default rather than erasing it.
- `GET /healthz` returns status 200, `Content-Type: application/json`, and exactly
  `{"service":"winch-sandbox","status":"ok"}`. `GET /api/session` returns status 200,
  JSON content type, and exactly
  `{"profile":"container-standard","unenforcedControls":["network-egress"]}` under
  the default profile. The array is non-empty and uses stable machine-readable names;
  it says that egress is not enforced rather than restating the aspirational profile.
  This is the control `docs/security.md` §4 requires while the container is real and
  its allowlist is not.
- The attach handler permits only `GET`/`HEAD` for static assets and API reads, returns
  405 with `Allow` for other methods, and serves files only beneath the configured
  static directory. Unknown API paths return JSON 404 rather than the SPA shell;
  unknown non-API paths fall back to `index.html` so a later attach route can reload.
- `web/` as a TypeScript/React workspace: `package.json` with `format:check`, `lint`,
  `typecheck`, `test`, `build`; `web/src/attach/` entry point and page shell rendering
  the posture from `GET /api/session`. Remove the stale `web/dist/`,
  `web/node_modules/`, and `web/*.tsbuildinfo` left by the previous plan.
- `deployments/Dockerfile` — multi-stage: node builds the page, go builds
  `winch-sandbox`, `winch`, and `fake-harness`, and a non-root runtime stage carries all
  three binaries plus the built page. The harness binary ships from this task so the
  image is already "one container holds the wrapper and the fake harness".
- `deployments/compose.yml` — a `sandbox` service that overrides the in-container
  listen address to `0.0.0.0:8080` but publishes it as `127.0.0.1:8080:8080`. The
  distinction is required: binding the process to container loopback would make the
  published port unreachable, while omitting the publish host would expose it on every
  host interface. A `toolchain` service under the `test` profile supplies Go for a host
  that has none. `deployments/README.md` documents both, which `Makefile:18` already
  cites.
- `cmd/winch` — the maintained operator CLI, one file per command. `winch status`
  accepts `--url` (default `http://127.0.0.1:8080`), applies a finite request timeout,
  checks `/healthz` and `/api/session`, and prints deterministic text naming service
  health, profile, and every unenforced control. A non-2xx response, malformed posture,
  empty `unenforcedControls`, or unreachable server produces a non-zero exit without
  printing a healthy status. Later tasks add one file and command each.
- `test/contract/attach/` pins the two JSON response shapes, methods, content types,
  SPA/API fallback split, and the posture omission failure. Package-level transport
  tests cover the configured-directory boundary and startup configuration errors.
- `test/e2e/` — the standing scenario suite, one file per scenario plus a shared helper
  that invokes `docker compose`, waits with a deadline for health, captures compose
  logs on failure, and always tears the topology down. The first scenario starts the
  built image through its real entrypoint, checks both endpoints, runs `winch status`,
  verifies the non-root uid, and proves loopback-only publication. Tests use an HTTP
  client with proxies disabled so a proxy-generated refusal cannot falsely satisfy the
  reachability assertion.
- Truncate the gates to the paths this tree has, leaving them green:
  `Makefile` (drop `api-generate`, `api-validate`, `api-compat`, `api-check`; limit Go
  packages to paths that exist; have `build` produce `winch-sandbox`, `winch`, and
  `fake-harness`; have `run` start `winch-sandbox`; have `e2e` run the scenario via the
  compose `toolchain`; reduce the `[docker]` group to `test-cycle`, its toolchain
  prerequisite, and teardown; drop the postgres and integration targets),
  `.github/workflows/go.yml` (`-s tests` → `-s test`,
  drop the postgres service and `make test-integration`, add the Docker e2e step),
  `.github/workflows/web.yml` (drop `npm run api:check`), `.dockerignore` (`runner/`
  and `tests/` name nothing; `test/` is what exists), `.gitignore` as needed.
- Correct the "currently broken" sections of `README.md` (`:42-47`) and `AGENTS.md`
  (`:100-120`), which describe a tree this task repairs.

## Non-goals

- The harness. Nothing starts, reads, or writes to a harness process here; the image
  carries the binary and does not run it. P0-002.
- Any record, ordinal, or store. P0-002 and P0-003.
- An OpenAPI document. `api/openapi/` is the control plane's public API, roadmap
  Stage 3; this surface is pinned by `test/contract/attach/` and the standing suite.
- A bearer token on the attach surface. `docs/roadmap.md` §3 puts authentication beyond
  a loopback binding outside Stage 0 and `docs/security.md` §11 LB11 blocks exposing
  the surface off-host until Stage 5 makes it real.
- PostgreSQL, migrations, and the daemon. Roadmap stages 3 and 4.

## Runtime reachability

`cmd/winch-sandbox` is the composition root. The `sandbox` service of
`deployments/compose.yml` runs it under the `container-standard` profile.
`docker compose -f deployments/compose.yml up --build` reaches every line of it;
`winch status` and a browser on the published loopback port reach the surface.

## Write set

- `cmd/winch-sandbox/main.go`, `cmd/winch-sandbox/config.go`
- `internal/adapters/transport/attach/server.go`, `.../session.go`, `.../static.go`
- `cmd/winch/main.go`, `cmd/winch/client.go`, `cmd/winch/status.go`
- `web/package.json`, `web/tsconfig*.json`, `web/vite.config.ts`,
  `web/eslint.config.js`, `web/index.html`, `web/src/attach/main.tsx`,
  `web/src/attach/App.tsx`, `web/src/attach/Posture.tsx`
- `deployments/Dockerfile`, `deployments/compose.yml`, `deployments/README.md`
- `test/contract/attach/session_test.go`, `test/contract/attach/static_test.go`
- `test/e2e/compose.go`, `test/e2e/scenario_sandbox_starts_test.go`
- `Makefile`, `.github/workflows/go.yml`, `.github/workflows/web.yml`,
  `.dockerignore`, `.gitignore`
- `README.md`, `AGENTS.md`

## Contract surfaces

- API: `GET /healthz`
- API: `GET /api/session` — JSON `{profile: string, unenforcedControls: string[]}`
- config keys: `WINCH_SANDBOX_ADDR`, `WINCH_SANDBOX_STATIC_DIR`, `WINCH_SANDBOX_PROFILE`
- CLI namespace: `winch`, and `winch status [--url URL]`
- compose service names `sandbox` and `toolchain`, and the `test` compose profile
- the `Makefile` target set and the two workflow job definitions

## Demonstration

    $ docker compose -f deployments/compose.yml up --build -d
    $ curl --noproxy '*' -i http://127.0.0.1:8080/healthz
    → expect: HTTP 200, JSON content type, and
      {"service":"winch-sandbox","status":"ok"}

    $ ./bin/winch status            # or: docker compose exec sandbox winch status
    → expect exactly:
      service: winch-sandbox (ok)
      profile: container-standard
      unenforced control: network-egress

    $ curl --noproxy '*' -fsS http://127.0.0.1:8080/
    → expect: HTML whose visible shell names the attach page; opening the same URL in
      a browser shows profile container-standard and network egress as not enforced

    $ host_ip=$(hostname -I | awk '{print $1}')
    $ curl --noproxy '*' -fsS --max-time 2 "http://${host_ip}:8080/healthz"
    → expect: curl exit 7 (connection refused), while the loopback request still passes

    $ docker compose -f deployments/compose.yml down

## Verification

- Unit and attach-contract tests pass inside the supplied toolchain:
  `docker compose -f deployments/compose.yml --profile test run --rm toolchain go test ./cmd/... ./internal/... ./test/contract/...`.
- Standing scenario suite passes against the `container-standard` profile:
  `make e2e` runs `test/e2e/scenario_sandbox_starts_test.go` against the composed image.
  The harness is present but deliberately not started, so this is not yet the all-fake
  harness profile introduced by P0-002.
- `make test-cycle` succeeds on a host with no Go toolchain, using the `toolchain`
  service. This is the repository's documented way in (`Makefile:17-18`) and is
  currently broken.
- `make check` passes: `format-check`, `vet`, `lint`, `test`, `build`.
- `cd web && npm run format:check && npm run lint && npm run typecheck && npm test && npm run build`.
- `python3 -m unittest discover -s test` → 9 tests, and the control-panel suite → 44
  tests, both still pass.
- `docker compose -f deployments/compose.yml config` succeeds and its rendered
  `sandbox` port has host IP `127.0.0.1`.
- `.github/workflows/go.yml` and `web.yml` both pass on the pull request.

## Acceptance criteria

- [ ] `docker compose -f deployments/compose.yml up --build` serves `/healthz` with no
      manual step beyond that command; its status, content type, and exact body match
      the contract in Scope, and termination stops cleanly.
- [ ] The container runs as a non-root user, and `docker compose exec sandbox id`
      shows a non-zero uid.
- [ ] `GET /api/session` and the page both name the profile **and** at least one
      control it does not enforce. A posture response that omits the unenforced
      control fails this task: `docs/security.md` §4 makes stating the gap the control
      that permits the gap, and T02 makes the omission a tracked threat.
- [ ] The published port answers on `127.0.0.1` and a proxy-disabled request to the
      host's routable address is refused while loopback remains healthy. The rendered
      compose configuration pins the published host IP to `127.0.0.1` rather than
      relying only on the process's in-container bind address.
- [ ] `make check`, `make e2e`, `make test-cycle`, and both CI workflows pass. No
      Makefile target and no workflow step names a path absent from the tree —
      checkable by reading every path in `Makefile` and `.github/workflows/` against
      the tree.
- [ ] Removing or emptying `WINCH_SANDBOX_STATIC_DIR` still starts the image and serves
      the page from `/opt/winch/web`; setting it to a directory without `index.html`,
      setting an invalid address, or selecting an unknown profile fails startup.
- [ ] `winch status` exits zero only for a healthy, valid posture and prints the exact
      three observations in the Demonstration; endpoint failure, malformed JSON, or an
      empty `unenforcedControls` array exits non-zero.
- [ ] The attach contract tests reject non-read methods, traversal outside the static
      root, an SPA response for an unknown `/api/` path, and a default posture that
      omits `network-egress`.
- [ ] `README.md` and `AGENTS.md` no longer describe the gates as broken, and the
      commands they list as working still work.
- [ ] I1 and I2 hold at HEAD: the system starts and deploys from this task forward.

## Deferrals

None.

## Traces to

`docs/roadmap.md` §3 Stage 0 (including the floor repair at `:86-90`) and §5;
`docs/architecture.md` §5 Topology A; `docs/code-structure.md` §1, §6;
`docs/contracts.md` §8 (*The surface states its posture*);
`docs/security.md` §2 (browser → attach surface), §4, §11 LB11, T02, T15;
ADR-0005, ADR-0006; `docs/state.md` *Every quality gate names a path that is gone*.
