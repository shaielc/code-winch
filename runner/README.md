# Code Winch task runner

The repository-scoped self-hosted GitHub runner that dispatches workplan tasks
to Codex Cloud, plus a local control panel over its scheduler state. It receives
jobs over the GitHub Actions runner connection, so it needs outbound network
access but no public webhook endpoint.

## Install

Run every command below from the repository root. Copy `runner/.env.example` to
`runner/.env` and replace these three values:

| Variable | Value |
| --- | --- |
| `GITHUB_URL` | The repository the runner registers with, replacing the `OWNER/REPOSITORY` placeholder. |
| `RUNNER_TOKEN` | A registration token from the repository's **Settings → Actions → Runners → New self-hosted runner**. It expires about an hour after it is issued, so take it shortly before the first `up`. |
| `CODEX_ENV_ID` | The Codex Cloud environment dispatched tasks run in. Empty in the example; the control panel refuses to dispatch without it. |

The remaining variables have working defaults. `RUNNER_LABELS` must keep
`code-winch` and `task-scheduler`, which is what the `runs-on` in
`.github/workflows/task-scheduler.yml` selects.

Then build:

```sh
docker compose --env-file runner/.env -f runner/compose.yml build
```

Codex authentication lives in a dedicated Docker volume. Populate it with the
interactive login before starting the services:

```sh
docker compose --env-file runner/.env -f runner/compose.yml run --rm \
  --entrypoint codex runner login
docker compose --env-file runner/.env -f runner/compose.yml up -d
```

`RUNNER_TOKEN` is only used for the initial registration, so an expired one in
`runner/.env` is harmless afterwards; take a fresh token if the runner volume is
removed and the runner has to re-register.

Set the repository Actions variable `CODEX_ENV_ID` to the same environment as
well. The workflow reads the Actions variable and the control panel reads
`runner/.env`, so both need it.

Run the **Schedule available tasks** workflow once after installing to dispatch
the dependency-free tasks and publish the first tracker snapshot.

## Deployment overrides

`compose.yml` is committed and deployment-neutral. Site-specific changes belong
in `compose.override.yml`, which is gitignored and stays out of the project.
Mappings merge key by key and sequences append, so an override adds to the
committed definition rather than replacing it.

Compose loads an override file automatically only when no `-f` flag is given, so
every command must name both files:

```sh
docker compose --env-file runner/.env \
  -f runner/compose.yml -f runner/compose.override.yml up -d
```

Set `COMPOSE_FILE` to avoid repeating them:

```sh
export COMPOSE_FILE=runner/compose.yml:runner/compose.override.yml
```

## Services

`runner` is the Actions runner itself. `control-panel` serves the scheduler view
on `127.0.0.1:8765`; override `PANEL_BIND` and `PANEL_PORT` to change the
binding, or put the service on a reverse proxy network with an override.

The two share the `scheduler-state` volume, which is their only coupling: the
panel reads and writes state there and never talks to the runner directly.

## How scheduling works

When a pull request containing exactly one known task ID merges, the Actions
workflow invokes `scripts/task_scheduler.py` with GitHub's event file. The
scheduler overlays its in-flight leases on the tracker from `origin/main` and
submits newly available tasks with `codex cloud exec`. A lock file beside the
state file prevents overlapping runs from dispatching the same task.

That run also publishes the tracker it scheduled from to
`/var/lib/code-winch/tracker.json`. The control panel reads that snapshot rather
than cloning the repository, so it holds no GitHub credentials, and it runs the
scheduler from the copy of `scripts/` baked into the image. The snapshot
refreshes on every scheduler run — the same event that changes the tracker — so
a manual **Schedule available tasks** run is what picks up a tracker edit that
did not arrive through a merged pull request. Until the first run publishes a
snapshot, the panel shows an empty table and says so.

`docs/workplan/tasks.json` on the default branch remains the sole authority for
`completed`; a local entry overrides the tracker only until the tracker records
the task as completed. At that point the entry is retired in place rather than
deleted, so the row keeps linking the pull request and the Codex task the work
went through. Expiring a lease releases the concurrency slot it holds. See
[ADR-0004](../docs/decisions/0004-task-status-authority.md).

Because the scripts are baked into the image, changing `scripts/` requires a
rebuild before the panel runs the new version.

## Operational notes

The runner configuration, work directory, Codex credentials, and scheduler state
survive container replacement in named volumes. Treat those volumes as secrets:
the Codex volume holds the CLI login, and the scheduler-state volume also
retains the GitHub runner credentials. Never mount the Docker socket into this
runner.

`.env.example` defaults to the latest runner image and Codex CLI. For repeatable
deployments, replace `RUNNER_IMAGE` and `NODE_IMAGE` with reviewed image digests
and `CODEX_VERSION` with a tested CLI version before building.
