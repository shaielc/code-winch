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

The `/api/v1/runs*` routes are mounted but not yet bound to run use cases, which
is P1-050. Until then they answer `404` for reads and `500` for creation, and no
harness process is launched.

## Services

| Service | Image | Published | Notes |
|---|---|---|---|
| `winchd` | Go daemon + web assets + fake harness | `127.0.0.1:8080` | Serves the SPA and `/api/v1` from one origin |
| `postgres` | `postgres:17-alpine` | internal only | Data persists in the `postgres-data` volume |
| `runner` | Go toolchain, the daemon image's build stage | not published | `test` profile only; see [Running the tests](#running-the-tests) |

`winchd` is published only on loopback. The API rejects mutating requests whose
`Origin` does not match `WINCH_ALLOWED_ORIGIN`, so the browser must reach both
the UI and the API through the same origin. The daemon injects the CSRF token
into the served `index.html`; it never puts that token in a cookie or exposes a
bootstrap API.

Requests authenticate with `Authorization: Bearer $WINCH_TOKEN`. The handler
also accepts a `winch_session` cookie scoped to `/api/v1`, but nothing issues
one yet — browser session establishment arrives with P1-050.

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
runner down. It tears down even when a step fails, and exits with that step's
status.

The same four steps, run by hand when you want the runner to stay up between
edits:

```sh
make runner-image      # build the toolchain image (needs registry access)
make test-env          # start the runner and create the winch_test database
make test-integration  # go test -tags integration ./... inside the runner
make test-env-down     # stop the runner; the daemon keeps running
```

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
`PG_TEST_DATABASE_URL` is set; the profile sets it to a `winch_test` database
alongside `winch` on the same server. That separation matters: the test helper
runs `DROP SCHEMA public CASCADE` before each migration, so pointing this
variable at `winch` would destroy the daemon's database.

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
the daemon image, so the harness adapter can launch it by name once P1-050 binds
the runner. Nothing launches it today, but you can drive it by hand:

```sh
docker compose -f deployments/compose.yml exec winchd fake-harness
```

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
