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

- `internal/adapters/transport/attach/` owns an `http.Handler`; keep route registration,
  JSON encoding, and static-file fallback there. `cmd/winch-sandbox/` is the first
  composition root: resolve configuration, construct that handler, start one
  `http.Server`, and shut it down on `SIGINT` or `SIGTERM`. There is no application or
  domain layer yet because this task has no session behaviour to orchestrate.
- Pin the initial wire responses in `test/contract/attach/`: `GET /healthz` returns
  `200` and `{"service":"winch-sandbox","status":"ok"}`; `GET /api/session`
  returns `200` and
  `{"profile":"container-standard","unenforcedControls":["egress"]}`. Use JSON
  content types and reject non-`GET` methods rather than allowing the static-page
  fallback to answer them.
- Configuration from compiled defaults overridden by environment, per
  `docs/code-structure.md` §6: `WINCH_SANDBOX_ADDR` (default `127.0.0.1:8080`),
  `WINCH_SANDBOX_STATIC_DIR` (default `/srv/winch/web`), and
  `WINCH_SANDBOX_PROFILE` (default `container-standard`). Empty environment values are
  treated as unset. Fail startup with a field-naming error when an explicit static
  directory is missing or the listen address is invalid.
- The posture `GET /api/session` reports: the profile name, and the controls that
  profile does not yet enforce. At this stage that is egress — the container is real
  and the allowlist is not, which `docs/security.md` §4 permits only on condition the
  surface says so.
- `web/` as a reproducibly installed TypeScript/React workspace: commit
  `package-lock.json`; provide `format:check`, `lint`, `typecheck`, `test`, and `build`
  scripts; and put the entry point, page shell, posture fetch, loading state, and
  explicit fetch-failure state under `web/src/attach/`. The page renders the profile
  and every value in `unenforcedControls`; it must not replace a failed or malformed
  posture response with reassuring defaults. Remove the stale `web/dist/`,
  `web/node_modules/`, and `web/*.tsbuildinfo` left by the previous plan and keep them
  ignored.
- `deployments/Dockerfile` — multi-stage: node builds the page, go builds
  `winch-sandbox`, `winch`, and `fake-harness`, and a non-root runtime stage carries all
  three binaries plus the built page. The harness binary ships from this task so the
  image is already "one container holds the wrapper and the fake harness".
- `deployments/compose.yml` — a `sandbox` service that overrides the in-container
  listener to `0.0.0.0:8080` but publishes it as `127.0.0.1:8080:8080`; binding the
  process itself to loopback would make Docker's published port unreachable. Add a
  `toolchain` service under the `test` profile with the source tree mounted at a fixed
  working directory and the Docker CLI/socket available, so `make test-cycle` can run
  the Go and compose-backed checks without host Go. `deployments/README.md` documents
  both, which `Makefile:18` already cites.
- `cmd/winch` — the maintained operator CLI, one file per command. `winch status`
  fetches and prints the posture and health. Later tasks add one command each.
- `test/e2e/` — the standing scenario suite, one file per scenario plus a shared helper
  that runs Compose with a unique project name, waits with a bounded retry for
  `/healthz`, captures service logs on failure, and always tears down with volumes and
  orphan removal. The first scenario starts the built image, checks both JSON
  responses, asserts the runtime uid is non-zero, and verifies the loopback-only port.
- Truncate the gates to the paths this tree has, leaving them green. Preserve the
  public target names `build`, `run`, `e2e`, `check`, and `test-cycle`, because they
  are repository/operator contracts, but replace their deleted daemon, PostgreSQL,
  OpenAPI, and `tests/` inputs:
  `Makefile` (drop `api-generate`, `api-validate`, `api-compat`, `api-check`; point
  `COMPOSE`, `build`, `run`, `e2e`, and the `[docker]` group at what exists; drop the
  postgres and integration targets), `.github/workflows/go.yml` (`-s tests` → `-s test`,
  drop the postgres service and `make test-integration`, add `make e2e` after the host
  checks), `.github/workflows/web.yml` (use `npm ci` and drop `npm run api:check`),
  `.dockerignore` (`runner/`
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
- `web/package.json`, `web/package-lock.json`, `web/tsconfig*.json`,
  `web/vite.config.ts`, `web/vitest.config.ts`, `web/eslint.config.js`,
  `web/.prettierrc.json`, `web/index.html`, `web/src/attach/main.tsx`,
  `web/src/attach/App.tsx`, `web/src/attach/App.test.tsx`,
  `web/src/attach/Posture.tsx`
- `deployments/Dockerfile`, `deployments/compose.yml`, `deployments/README.md`
- `test/e2e/compose.go`, `test/e2e/scenario_sandbox_starts_test.go`
- `test/contract/attach/session_test.go`
- `Makefile`, `.github/workflows/go.yml`, `.github/workflows/web.yml`,
  `.dockerignore`, `.gitignore`
- `README.md`, `AGENTS.md`

## Contract surfaces

- API: `GET /healthz`
- API: `GET /api/session` —
  `{"profile": string, "unenforcedControls": string[]}`
- config keys: `WINCH_SANDBOX_ADDR`, `WINCH_SANDBOX_STATIC_DIR`, `WINCH_SANDBOX_PROFILE`
- CLI namespace: `winch`, and `winch status`
- compose service names `sandbox` and `toolchain`, and the `test` compose profile
- the `Makefile` target set and the two workflow job definitions

## Demonstration

    $ docker compose -f deployments/compose.yml up --build -d
    $ curl -fsS http://127.0.0.1:8080/healthz | jq -e \
        '.service == "winch-sandbox" and .status == "ok"'
    → expect: jq exits 0

    $ ./bin/winch status
    → expect: `profile: container-standard` and `unenforced control: egress`

    $ curl -fsS http://127.0.0.1:8080 | grep -F '<div id="root"></div>'
    $ xdg-open http://127.0.0.1:8080
    → expect: curl finds the React mount point; in a browser the page renders the same
      profile and unenforced egress control returned by `/api/session`

    $ curl -fsS --max-time 2 http://$(hostname -I | awk '{print $1}'):8080/healthz
    → expect: curl exits non-zero because Compose publishes the surface on host
      loopback only

    $ docker compose -f deployments/compose.yml exec -T sandbox id -u
    → expect: a non-zero uid

    $ docker compose -f deployments/compose.yml down

## Verification

- Standing scenario suite passes against the all-fake profile:
  `make e2e` runs `test/e2e/scenario_sandbox_starts_test.go` against the composed image.
- `go test ./test/contract/attach/...` pins the health and posture responses, rejects
  non-GET requests, and covers a missing or malformed static directory.
- Web component tests cover loading, the declared posture, and an API error; the error
  case must not render `container-standard` or claim egress is enforced.
- `make test-cycle` succeeds on a host with no Go toolchain, using the `toolchain`
  service. This is the repository's documented way in (`Makefile:17-18`) and is
  currently broken.
- `make check` passes: `format-check`, `vet`, `lint`, `test`, `build`.
- `cd web && npm run format:check && npm run lint && npm run typecheck && npm test && npm run build`.
- `python3 -m unittest discover -s test` and
  `cd workplan-control-panel && python3 -m unittest discover -s tests` still pass; do
  not pin counts, because these suites may grow independently of this task.
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
- [ ] If `/api/session` fails, returns malformed JSON, or omits either required field,
      the page displays an unavailable/error state and does not invent a profile or
      imply that every control is enforced.
- [ ] The published port answers on `127.0.0.1` and is refused from the host's
      routable address.
- [ ] `make check`, `make e2e`, `make test-cycle`, and both CI workflows pass. No
      Makefile target and no workflow step names a path absent from the tree —
      checkable by reading every path in `Makefile` and `.github/workflows/` against
      the tree.
- [ ] Removing `WINCH_SANDBOX_STATIC_DIR` from the Compose environment still starts
      the server and serves the page from the compiled `/srv/winch/web` default. An
      explicitly configured nonexistent directory instead fails startup with the key
      named in the error; operator mistakes must not degrade into an API-only server.
- [ ] `SIGTERM` stops the composed service within Compose's stop timeout and leaves no
      listening process; the e2e helper asserts teardown even after a failed check.
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
