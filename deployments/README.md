# Local deployment

A three-service development stack: the web UI, the `winchd` daemon, and
PostgreSQL.

```sh
docker compose -f deployments/compose.yml up --build
```

The UI is then served on <http://localhost:8080>. Stop the stack with `down`,
and add `-v` to discard the database volume.

## Current state

**The daemon is not wired yet.** `cmd/winchd/main.go` is still an empty
`main()`, so the `winchd` container builds, starts, and exits immediately with
status 0. Postgres and the web UI come up normally, but every `/api/v1` request
returns a proxy error until the composition root exists. The images, the network
topology, and the configuration contract below are what this stack establishes
now; they need no changes when the daemon lands.

## Services

| Service | Image | Published | Notes |
|---|---|---|---|
| `web` | nginx + built assets | `127.0.0.1:8080` | Serves the SPA and proxies `/api/v1` to the daemon |
| `winchd` | Go daemon + fake harness | internal only | Reached through the web proxy so the UI and API share one origin |
| `postgres` | `postgres:17-alpine` | internal only | Data persists in the `postgres-data` volume |

`winchd` is deliberately unpublished. The API rejects mutating requests whose
`Origin` does not match `WINCH_ALLOWED_ORIGIN`, and the session cookie is scoped
to `/api/v1` with `Secure` set, so the browser must reach both the UI and the API
through the same origin. Browsers accept `Secure` cookies over plain HTTP for
`localhost` only, which is why the published port binds to loopback.

## Configuration

| Variable | Default | Purpose |
|---|---|---|
| `WINCH_ADDR` | `:8080` | Daemon listen address |
| `WINCH_DATABASE_URL` | compose-internal | PostgreSQL connection string |
| `WINCH_ALLOWED_ORIGIN` | `http://localhost:8080` | Must equal the browser's origin |
| `WINCH_TOKEN` | development default | Session/bearer secret; minimum 32 bytes |
| `WINCH_CSRF_TOKEN` | development default | CSRF secret; minimum 32 bytes |
| `WINCH_ACTOR` | `local-user` | Actor recorded on every command |
| `WINCH_HARNESS_PROFILE` | `fake` | Harness adapter to launch |
| `WINCH_SANDBOX_PROFILE` | `local` | Sandbox driver to launch under |
| `WINCH_WEB_PORT` | `8080` | Host port for the UI |
| `POSTGRES_PASSWORD` | development default | Database password |

Override them in `deployments/.env` or `deployments/compose.override.yml`. Both
are untracked; do not commit either, and do not add a template that invites
copying secrets into the repository.

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

`WINCH_HARNESS_PROFILE=fake` launches `fake-harness`, a deterministic stand-in
for a coding-agent CLI that is built from `cmd/fake-harness` and installed on
`PATH` in the daemon image. It needs no vendor account. It reads one JSON
command per line — `{"id":"<id>","text":"<text>"}`, the format the fake harness
codec encodes — and also accepts bare text so it stays usable when attached
directly to its PTY:

| Input | Behavior |
|---|---|
| `help` | Print the command list |
| `echo <text>` | Emit `<text>` as terminal output |
| `fail` | Exit with status 1 |
| `exit`, `quit` | Exit with status 0 |
| anything else | Echoed back as terminal output |

It also emits an observation on `SIGTERM`, so a forced stop is distinguishable
from a crash in the run's event history.
