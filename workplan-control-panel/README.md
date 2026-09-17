# Workplan control panel

A self-contained HTTP API and UI for scheduling workplan tasks. GitHub Actions
only posts merge notifications; the panel owns its checkout and the task flow.

## Layout

All application code, prompts, tests, and deployment files are in this directory.
There are no imports from the repository's `scripts/` or tests directories.

| Path | Responsibility |
| --- | --- |
| `control_panel/api.py` | HTTP routes, authentication, server startup |
| `control_panel/ui.py` | Browser page and its API calls |
| `control_panel/task_scheduler.py` | Task selection, branch preparation, stage orchestration |
| `control_panel/state.py` | Reservations, locking, tracker reconciliation |
| `control_panel/integrations/` | Git/GitHub and Codex CLI operations |
| `control_panel/prompts/` | Bundled Refine, Implement, and Audit templates |
| `tests/` | Panel tests, including local Git and HTTP scenarios |

## Flow and API

On a merge into main, the runner posts to `/api/events/merge`. The panel pulls
its dedicated main checkout, reads `docs/workplan/tasks.json`, and selects pending
tasks whose dependencies are completed. It creates `task/<ID>` with an opening
in-progress commit, up to three active tasks by default (`--max-concurrent`).
Retries reuse prepared branches. The main tracker remains authoritative for completion.

The UI at `/` retains table and dependency-tree views. **Refine** and **Implement**
submit the corresponding prompt to Codex Cloud on the task branch. Merge the
refinement PR into that branch before implementing. **Audit** copies a prompt
naming an open implementation PR and its current head; when several PRs exist,
the UI asks which to use. A text field is available if clipboard access fails.

All `/api/` routes require `Authorization: Bearer <PANEL_TOKEN>`; POST bodies must
be JSON objects. Enter the token in the UI to use its buttons. The UI and
`/healthz` are readable without authentication, so retain the loopback binding or
protect the entire site behind a reverse proxy.

| Method | Path | Body / result |
| --- | --- | --- |
| GET | `/healthz` | Process health |
| GET | `/api/tasks` | Tracker and local task/stage records |
| POST | `/api/events/merge` | GitHub closed-PR event; ignores non-main merges |
| POST | `/api/sync` | `{}`; pull main and prepare available tasks |
| POST | `/api/tasks/<ID>/refine` | `{}`; submit refinement |
| POST | `/api/tasks/<ID>/implement` | `{}`; submit implementation |
| POST | `/api/tasks/<ID>/audit` | Optional `{"pr_number": 123}`; return formatted prompt |

Concurrent operations return 409. Repeating a submitted stage at the same branch
revision returns its existing URL. If a submission has an uncertain outcome,
check Codex before stopping the panel and clearing that entry from the task's
`stages` map in `task-state.json`; retries otherwise remain blocked.

## Run and test

From this directory, Python 3.11+ and Git are sufficient for the tests. Running
the panel also requires GitHub CLI and an authenticated Codex CLI:

```sh
python3 -m unittest discover -s tests -v
PANEL_TOKEN=... CODEX_ENV_ID=... python3 -m control_panel \
  --clone=/path/to/dedicated-main-checkout --state-file=/path/to/state/task-state.json
```

Tests use a local Git remote and mocked cloud submissions/PR listing.

## Deploy

From this directory, copy `.env.example` to `.env` and set `GITHUB_URL`, a current
runner registration `RUNNER_TOKEN`, `GH_TOKEN` (contents write and pull-request
read), `CODEX_ENV_ID`, and `PANEL_TOKEN`. Set the repository Actions secret
`CONTROL_PANEL_TOKEN` to that same panel token.

```sh
docker compose --env-file .env -f compose.yml build
./codex_login.sh
docker compose --env-file .env -f compose.yml up -d
```

The Docker build context is this directory alone. The panel clones main on first
start; use **Sync main** or the **Schedule available tasks** workflow to prepare
the initial tasks. Rebuild the image after code or prompt changes.

The workflow uses `curl` directly, defaulting to `http://control-panel:8765` on the
Compose network. Set Actions variable `CONTROL_PANEL_URL` for another address;
use HTTPS across hosts. No scheduler script runs on the GitHub runner.

The runner's registration/work volumes are separate from the panel's checkout,
state, and Codex login. When upgrading the old shared-volume installation, stop
it and preserve its `task-state.json` in scheduler-state and its registration in
runner-config. Do not run old and new schedulers together.

Use `PANEL_BIND` and `PANEL_PORT` to override `127.0.0.1:8765`. Site-specific
Compose changes belong in the gitignored `compose.override.yml`; `up.sh` includes
it. Pin reviewed image digests and a tested Codex CLI version for repeatable builds.
