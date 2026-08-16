# Local deployment

A two-service development stack: the `winchd` daemon (including the built web
UI) and PostgreSQL.

```sh
docker compose -f deployments/compose.yml up --build
```

The UI is then served on <http://localhost:8080>. Stop the stack with `down`,
and add `-v` to discard the database volume.

At startup the daemon validates configuration, connects to PostgreSQL, applies
all five migrations, and only then opens its listener. `GET /api/v1/health`
returns `{"status":"ok"}`. Shutdown signals close live event subscribers and
give HTTP requests the configured bounded drain period.

## Services

| Service | Image | Published | Notes |
|---|---|---|---|
| `winchd` | Go daemon + web assets + fake harness | `127.0.0.1:8080` | Serves the SPA and `/api/v1` from one origin |
| `postgres` | `postgres:17-alpine` | internal only | Data persists in the `postgres-data` volume |

`winchd` is published only on loopback. The API rejects mutating requests whose
`Origin` does not match `WINCH_ALLOWED_ORIGIN`, and the session cookie is scoped
to `/api/v1` with `Secure` set, so the browser must reach both the UI and the API
through the same origin. The daemon injects the CSRF token into the served
`index.html`; it never puts that token in a cookie or exposes a bootstrap API.

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

Set `WINCH_CONFIG_FILE` to load an optional YAML configuration file before
environment overrides are applied. `WINCH_STATIC_DIR` selects the built asset
directory, and `WINCH_SHUTDOWN_TIMEOUT` (default `10s`) bounds HTTP and stream
draining. Authentication secrets deliberately have no compiled default and
must contain at least 32 bytes.

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
