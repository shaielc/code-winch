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

- `cmd/winch-sandbox/` — the first composition root. HTTP server on a configured
  address, `GET /healthz`, `GET /api/session` returning the session's declared posture,
  and the built page served from a configured directory.
- Configuration from compiled defaults overridden by environment, per
  `docs/code-structure.md` §6: `WINCH_SANDBOX_ADDR` (default `127.0.0.1:8080`),
  `WINCH_SANDBOX_STATIC_DIR`, `WINCH_SANDBOX_PROFILE` (default `container-standard`).
- The posture `GET /api/session` reports: the profile name, and the controls that
  profile does not yet enforce. At this stage that is egress — the container is real
  and the allowlist is not, which `docs/security.md` §4 permits only on condition the
  surface says so.
- `web/` as a TypeScript/React workspace: `package.json` with `format:check`, `lint`,
  `typecheck`, `test`, `build`; `web/src/attach/` entry point and page shell rendering
  the posture from `GET /api/session`. Remove the stale `web/dist/`,
  `web/node_modules/`, and `web/*.tsbuildinfo` left by the previous plan.
- `deployments/Dockerfile` — multi-stage: node builds the page, go builds
  `winch-sandbox`, `winch`, and `fake-harness`, and a non-root runtime stage carries all
  three binaries plus the built page. The harness binary ships from this task so the
  image is already "one container holds the wrapper and the fake harness".
- `deployments/compose.yml` — a `sandbox` service publishing only to loopback, and a
  `toolchain` service under the `test` profile supplying Go for a host that has none.
  `deployments/README.md` documents both, which `Makefile:18` already cites.
- `cmd/winch` — the maintained operator CLI, one file per command. `winch status`
  fetches and prints the posture and health. Later tasks add one command each.
- `test/e2e/` — the standing scenario suite, one file per scenario plus a shared helper
  that brings the compose topology up and tears it down. First scenario: the sandbox
  starts, answers `/healthz`, and reports its posture.
- Truncate the gates to the paths this tree has, leaving them green:
  `Makefile` (drop `api-generate`, `api-validate`, `api-compat`, `api-check`; point
  `COMPOSE`, `build`, `run`, `e2e`, and the `[docker]` group at what exists; drop the
  postgres and integration targets), `.github/workflows/go.yml` (`-s tests` → `-s test`,
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
- `test/e2e/compose.go`, `test/e2e/scenario_sandbox_starts_test.go`
- `Makefile`, `.github/workflows/go.yml`, `.github/workflows/web.yml`,
  `.dockerignore`, `.gitignore`
- `README.md`, `AGENTS.md`

## Contract surfaces

- API: `GET /healthz`
- API: `GET /api/session` — the posture document
- config keys: `WINCH_SANDBOX_ADDR`, `WINCH_SANDBOX_STATIC_DIR`, `WINCH_SANDBOX_PROFILE`
- CLI namespace: `winch`, and `winch status`
- compose service names `sandbox` and `toolchain`, and the `test` compose profile
- the `Makefile` target set and the two workflow job definitions

## Demonstration

    $ docker compose -f deployments/compose.yml up --build -d
    $ curl -fsS http://127.0.0.1:8080/healthz
    → expect: 200 and a body naming the service

    $ ./bin/winch status            # or: docker compose exec sandbox winch status
    → expect: profile container-standard, and a line stating that egress is not
      enforced at this stage

    $ xdg-open http://127.0.0.1:8080
    → expect: the page renders the same posture, including the unenforced control

    $ curl -fsS --max-time 2 http://$(hostname -I | awk '{print $1}'):8080/healthz
    → expect: connection refused — the surface is bound to loopback only

    $ docker compose -f deployments/compose.yml down

## Verification

- Standing scenario suite passes against the all-fake profile:
  `make e2e` runs `test/e2e/scenario_sandbox_starts_test.go` against the composed image.
- `make test-cycle` succeeds on a host with no Go toolchain, using the `toolchain`
  service. This is the repository's documented way in (`Makefile:17-18`) and is
  currently broken.
- `make check` passes: `format-check`, `vet`, `lint`, `test`, `build`.
- `cd web && npm run format:check && npm run lint && npm run typecheck && npm test && npm run build`.
- `python3 -m unittest discover -s test` → 9 tests, and the control-panel suite → 44
  tests, both still pass.
- `.github/workflows/go.yml` and `web.yml` both pass on the pull request.

## Acceptance criteria

- [ ] `docker compose -f deployments/compose.yml up --build` serves `/healthz` with no
      manual step beyond that command.
- [ ] The container runs as a non-root user, and `docker compose exec sandbox id`
      shows a non-zero uid.
- [ ] `GET /api/session` and the page both name the profile **and** at least one
      control it does not enforce. A posture response that omits the unenforced
      control fails this task: `docs/security.md` §4 makes stating the gap the control
      that permits the gap, and T02 makes the omission a tracked threat.
- [ ] The published port answers on `127.0.0.1` and is refused from the host's
      routable address.
- [ ] `make check`, `make e2e`, `make test-cycle`, and both CI workflows pass. No
      Makefile target and no workflow step names a path absent from the tree —
      checkable by reading every path in `Makefile` and `.github/workflows/` against
      the tree.
- [ ] Removing `WINCH_SANDBOX_STATIC_DIR` from the environment still starts the server
      and serves the page from the compiled default, so a missing override degrades to
      a default rather than a failure to boot.
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
