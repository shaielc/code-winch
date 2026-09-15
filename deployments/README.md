# Local deployment

A two-service development stack: the `winchd` daemon (including the built web
UI) and PostgreSQL.

```sh
docker compose -f deployments/compose.yml up --build
```

The UI is then served on <http://localhost:8080>. Stop the stack with `down`,
and add `-v` to discard the database volume.

At startup the daemon validates configuration, connects to PostgreSQL, applies
any migrations the database does not already record, and only then opens its
listener. `GET /api/v1/health` returns `{"status":"ok"}`. Shutdown signals close
live event subscribers and give HTTP requests the configured bounded drain
period.

Four run routes are bound to use cases: `POST /api/v1/runs`,
`GET /api/v1/runs/{runId}`, `POST /api/v1/runs/{runId}/start`, and
`GET /api/v1/runs/{runId}/events`. Starting a run launches the fake harness
under the local sandbox, and the events it produces are durably stored and
readable by polling. `stop`, `input`, and the event WebSocket are mounted and
still unbound: they answer `500`.

Every appended event also records publish intent in the `outbox` table, and this
daemon starts no worker to drain it, so that backlog grows with each run.

## Operator CLI

The daemon image also installs the maintained `winch` operator CLI on `PATH`.
Invoke it in the running stack without installing Go on the host:

```sh
docker compose -f deployments/compose.yml exec winchd winch --help
```

`winch run create`, `run get`, `run start`, and `run events` drive the daemon
over HTTP. They read `WINCH_API_URL` (default `http://localhost:8080`),
`WINCH_TOKEN`, and `WINCH_CSRF_TOKEN` from the environment. `run start` reads
the run first and supplies its current ETag itself, because the command is
conditional:

```sh
RUN_ID=$(winch run create --workspace /tmp/ws --harness fake --sandbox local)
winch run start "$RUN_ID"
winch run events "$RUN_ID"
```

`run events` pages through the whole durable history, printing one
`sequence  kind  sensitivity  payload` line per event. `winch run get "$RUN_ID"`
reads the run back, including the terminal state it reached.

The `dev run` command is standalone and drives the local sandbox and fake
harness directly inside the container, with no daemon and no database:

```sh
printf 'echo hello\nexit\n' | \
  docker compose -f deployments/compose.yml exec -T winchd \
  winch dev run --harness fake --sandbox local
```

Host builds place both operator and daemon binaries at `bin/winch` and
`bin/winchd` by default. Set `BUILD_DIR` to choose another output directory.

## Services

| Service | Image | Published | Notes |
|---|---|---|---|
| `winchd` | Go daemon + operator CLI + web assets + fake harness | `127.0.0.1:8080` | Serves the SPA and `/api/v1` from one origin |
| `postgres` | `postgres:17-alpine` | internal only | Data persists in the `postgres-data` volume |
| `runner` | Go toolchain, the daemon image's build stage | not published | `test` profile only; see [Running the tests](#running-the-tests) |

`winchd` is published only on loopback. The API rejects mutating requests whose
`Origin` does not match `WINCH_ALLOWED_ORIGIN`, so the browser must reach both
the UI and the API through the same origin. The daemon injects the CSRF token
into the served `index.html`; it never puts that token in a cookie or exposes a
bootstrap API.

Requests authenticate with `Authorization: Bearer $WINCH_TOKEN`. The handler
also accepts a `winch_session` cookie scoped to `/api/v1`, but nothing issues
one yet — no code establishes a browser session.

## Configuration

| Variable | Default | Purpose |
|---|---|---|
| `WINCH_ADDR` | `:8080` | Daemon listen address |
| `WINCH_DATABASE_URL` | compose-internal | PostgreSQL connection string |
| `WINCH_ALLOWED_ORIGIN` | `http://localhost:8080` | Must equal the browser's origin |
| `WINCH_TOKEN` | development default | Session/bearer secret; minimum 32 bytes |
| `WINCH_CSRF_TOKEN` | development default | CSRF secret; minimum 32 bytes |
| `WINCH_ACTOR` | `local-user` | Actor recorded on every command |
| `WINCH_WEB_PORT` | `8080` | Host port for the UI |
| `POSTGRES_PASSWORD` | development default | Database password |
| `WINCH_FAKE_HARNESS_BINARY` | resolved on `PATH` | Explicit `fake-harness` path |
| `WINCH_FAKE_HARNESS_TRANSCRIPT` | none | Transcript the harness plays on every run |
| `WINCH_FAKE_HARNESS_DELAY` | `0s` | Latency injected before each scripted action |
| `WINCH_FAKE_HARNESS_FORCE_FAILURE` | `false` | Make the harness exit unsuccessfully |
| `WINCH_FAKE_HARNESS_MALFORMED_LINE` | `false` | Make the harness emit an invalid record |

Set `WINCH_CONFIG_FILE` to load an optional YAML configuration file before
environment overrides are applied. `WINCH_STATIC_DIR` selects the built asset
directory, and `WINCH_SHUTDOWN_TIMEOUT` (default `10s`) bounds HTTP and stream
draining. Authentication secrets deliberately have no compiled default and
must contain at least 32 bytes.

The image builds the browser assets itself. Outside the image — `make run`
against a fresh clone — `web/dist` does not exist, and the daemon logs
`web assets unavailable` and serves the API alone; run `make web-build` first if
you want the UI.

Override them in `deployments/.env` or `deployments/compose.override.yml`. Both
are untracked; do not commit either, and do not add a template that invites
copying secrets into the repository.

## Running the tests

Every Make target is marked `[host]` or `[docker]`. `[host]` targets run on this
machine and need `go`, `node`, or `golangci-lint` installed — CI uses those.
The Go workflow supplies PostgreSQL and runs both `make check` and the host
`make test-integration` target on every push and pull request.
`[docker]` targets need Docker only: the `runner` container supplies the Go
toolchain and `postgres` supplies the database. With no Go toolchain installed,
the `[docker]` group is the way in, and no target asks you to type a
`docker compose` command yourself.

The whole cycle is one command:

```sh
make test-cycle
```

That builds the image, starts the runner and its database, formats/vets/unit-tests
and compiles inside the container, runs the integration suite, and tears the
runner and its database down. It tears down even when a step fails, and exits
with that step's status.

The same four steps, run by hand when you want the runner to stay up between
edits:

```sh
make runner-image      # build the toolchain image (needs registry access)
make test-env          # start the runner and create the winch_test database
make runner-integration # go test -tags integration ./... inside the runner
make test-env-down     # stop the runner and drop winch_test; the daemon keeps running
```

Teardown drops the test database rather than leaving it on the server between
cycles. Nothing is lost: the integration helper drops and recreates `public`
before each migration anyway, so the database carries no state worth keeping,
and `test-env` recreates it on demand. Teardown refuses outright if
`TEST_DATABASE` has been pointed at `winch`, and leaves the database alone
rather than failing when postgres is already down.

`make test-env` builds the image on demand if it is missing, so `runner-image` is
only needed to pick up a Dockerfile change. Keeping it separate means repeated
test runs never touch the registry.

Two more `[docker]` targets: `make runner-verify` runs gofmt, `go vet`,
`go test ./...`, and `go build ./...` in the container — `check` minus `lint` and
`api-check`, which need golangci-lint and npm that the image does not carry.
`make runner-shell` opens a shell in it.

The runner is the daemon image's build stage with the repository bind-mounted
over its copy of the source, so it compiles what is in the working tree without
a rebuild. Module and build caches persist in the `go-mod` and `go-build`
volumes. That stage also marks `/src` a git safe directory: the mount arrives
owned by your host user while the container runs as root, and without it git
refuses to report VCS status and `go build` fails.

Integration tests are behind the `integration` build tag and skip unless
`PG_TEST_DATABASE_URL` is set. With a host PostgreSQL, run them directly with:

```sh
PG_TEST_DATABASE_URL='postgres://winch@127.0.0.1:55432/winch_test?sslmode=disable' \
  make test-integration
```

The Docker profile sets the same variable to a `winch_test` database alongside
`winch` on the same server. That separation matters: the test helper runs
`DROP SCHEMA public CASCADE` before each migration, so pointing this variable
at `winch` would destroy the daemon's database.

## Security posture

This stack is **local-trusted**, as defined in `docs/security.md`. The `local`
sandbox driver reports `unisolated`: harness processes run as the container's
user with no filesystem or network restriction, sharing the daemon's container.
That is a process-lifecycle boundary, not a security boundary. Container
isolation arrives with the Docker sandbox driver in phase 3.

The committed secrets are development defaults chosen so the stack starts
without setup. They are not credentials and must be replaced before this is
reachable from anything but loopback.

## The fake harness

`fake-harness` is a deterministic stand-in for a coding-agent CLI that needs no
vendor account. It is built from `cmd/fake-harness` and installed on `PATH` in
the daemon image. The adapter resolves that installed name to an absolute path
before launch; `winch dev run --fake-binary /path/to/fake-harness` can select an
explicit development build. The operator CLI drives the profile — see *Operator
CLI* above — and the fixture is also runnable on its own:

```sh
docker compose -f deployments/compose.yml exec winchd fake-harness
```

The daemon configures the same profile through the `WINCH_FAKE_HARNESS_*`
variables above, which are the transcript, latency, and injection controls
resolved through the ordinary configuration layering rather than compiled in. A
daemon-started run always runs the harness in early-exit mode, because such a
run has no way to answer an interactive prompt yet and a harness left reading
its terminal would never reach a terminal state. Early exit takes effect only
after a transcript has played out without ending the harness, so a transcript
ending in `exit` or `fail` still decides the run's outcome.

`winch dev run --help` documents the runtime controls. `--fake-transcript FILE`
plays the file's commands before interactive input, and `--fake-delay 500ms`
delays each scripted command. `--fake-force-failure`, `--fake-malformed-line`,
and `--fake-early-exit` inject a nonzero exit, an invalid JSON-lines record
(followed by a nonzero exit), or an exit before interactive input. An
unsuccessful harness exit becomes the CLI's own exit status. The corresponding
`fake-harness` flags omit the `fake-` prefix when invoking the fixture directly.

It reads one JSON command per line — `{"id":"<id>","text":"<text>"}`, the format
the fake harness codec encodes — and also accepts bare text so it stays usable
when attached directly to its PTY:

| Input | Behavior |
|---|---|
| `help` | Print the command list |
| `echo <text>` | Emit `<text>` as terminal output |
| `fail` | Exit with status 1 |
| `exit`, `quit` | Exit with status 0 |
| anything else | Echoed back as terminal output |

It also emits an observation on `SIGTERM`, so a forced stop is distinguishable
from a crash in the run's event history.

This profile proves deterministic runner, codec, and process-lifecycle behavior.
It does **not** exercise a real provider, make network requests, or prove
credential discovery, transport, or handling.
