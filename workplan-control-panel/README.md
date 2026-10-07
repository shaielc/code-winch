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
| `control_panel/integrations/` | Git/GitHub and Codex CLI operations; Claude routine trigger |
| `control_panel/prompts/` | Bundled Refine, Implement, and Audit templates |
| `tests/` | Panel tests, including local Git and HTTP scenarios |

## Flow and API

On a merge into main, the runner posts to `/api/events/merge`. The panel pulls
its dedicated main checkout, reads `docs/workplan/tasks.json`, and selects pending
tasks whose dependencies are completed. It creates `task/<ID>` with an opening
in-progress commit, up to three active tasks by default (`--max-concurrent`).
Sync returns HTTP 202 with a job ID; the UI and runner poll its status.
This keeps Git operations from holding a proxy connection open. Preparation
failures name the task and operation, and retries reuse prepared branches. The main tracker remains authoritative for completion.

A scheduling pass fills free slots in tracker order, which cannot serve a task you
want prepared next. `/api/tasks/<ID>/prepare` admits one named task instead, applying
the same rules — the tracker's status, completed dependencies, and a free slot, unless
the task already holds one. It refreshes main and answers with a job like a full sync,
and refuses while another sync is running rather than coalescing and losing the
selection. This is how a released reservation is taken back.

The UI at `/` retains table and dependency-tree views. Stage buttons are always
visible, with reasons when disabled; errors appear above the task list. **Refine** and **Implement**
submit the corresponding prompt to Codex Cloud on the task branch. Merge the
refinement PR into that branch before implementing. **Audit** fires the Claude
routine at `CLAUDE_AUDIT_ROUTINE_URL` with a prompt naming an open implementation
PR and its current head, starting a Claude cloud session; when several PRs exist,
the UI asks which to use. The routine is configured on claude.ai/code/routines
(see `claude-cloud-cli.md`): the panel supplies only the per-call text, because
`claude --cloud` cannot start a session without a terminal. Both views include
separate Refine, Implement, and Audit conversation links. **Prepare**
appears on a task the panel could admit now — available, mid-retry, or previously
released — and claims that one task. **Expire** releases the local reservation while
keeping those links, and the row's **⋯** menu expires a single stage record instead.
Neither cancels a cloud task.

API clients use `Authorization: Bearer <PANEL_TOKEN>`; POST bodies must be JSON
objects. In the UI, enter the token and **Sign in** to remember this browser for
seven days. The browser stores a signed `HttpOnly; Secure; SameSite=Strict`
session cookie scoped to the proxy mount path; it never stores the API token.
Persistent sign-in requires HTTPS (or localhost); plain HTTP can still use a
token temporarily in page memory. **Sign out** deletes this browser's cookie;
rotating `PANEL_TOKEN` invalidates all sessions. Cookie-authenticated POSTs require
`X-Panel-Request: 1`; the panel does not allow cross-origin requests. The UI and
`/healthz` are readable without authentication, so retain the loopback binding or
protect the entire site behind a reverse proxy.

| Method | Path | Body / result |
| --- | --- | --- |
| GET | `/healthz` | Process health |
| GET | `/api/tasks` | Tracker and local task/stage records |
| GET | `/api/session` | Browser authentication status |
| POST | `/api/session` | Bearer token plus `{"path": "/panel/"}`; create browser session |
| POST | `/api/session/logout` | `{"path": "/panel/"}`; clear browser session |
| POST | `/api/events/merge` | GitHub closed-PR event; returns 202 and a sync job ID |
| POST | `/api/sync` | `{}`; returns 202 and a sync job ID |
| GET | `/api/sync/<ID>` | `running`, `succeeded` with result, or `failed` with error |
| POST | `/api/tasks/<ID>/refine` | `{}`; submit refinement |
| POST | `/api/tasks/<ID>/implement` | `{}`; submit implementation |
| POST | `/api/tasks/<ID>/expire` | `{}`; release the local reservation, retaining conversation links |
| POST | `/api/tasks/<ID>/prepare` | `{}`; admit one named task; returns 202 and a sync job ID |
| POST | `/api/tasks/<ID>/refine\|implement\|audit/expire` | `{}`; retire that stage's current submission so it can be sent again |
| POST | `/api/tasks/<ID>/audit` | Optional `{"pr_number": 123}`; start a Claude session for that PR's head and return its `task_url` |

Concurrent sync requests share a job and request a fresh pass over main. Job
status is retained until restart; task reservations remain on disk. Other
concurrent operations return 409.

Repeating a submitted stage at the same branch revision returns its existing URL
rather than launching a second cloud task. The branch revision is what scopes that:
once a stage's pull request merges into the task branch, the tip moves and the stage
is open again. Audit is scoped the same way to the pull request's head, so a new
push to the PR can be audited again. Expiring a stage overrides it while the tip
has not moved; expiring Audit retires its newest session. A routine that refuses
the call (bad token, rate limit) records nothing, since no session started. Expiring
retires the submission rather than deleting it, so its conversation stays linked, and
a submission whose outcome is uncertain says so — Codex may have accepted it before
the reply was lost, so check there first, because expiring can run it twice.

`task-state.json` is written at schema version 2, which records stages as one list
per stage. A version 1 file, which keyed each submission `"<stage>:<head>"` in a
single map, is migrated when the panel reads it; an unrecognised version is refused
rather than guessed at.

## Run and test

From this directory, Python 3.11+ and Git are sufficient for the tests. Running
the panel also requires GitHub CLI and an authenticated Codex CLI:

```sh
python3 -m unittest discover -s tests -v
node tests/test_ui.cjs
PANEL_TOKEN=... CODEX_ENV_ID=... python3 -m control_panel \
  --clone=/path/to/dedicated-main-checkout --state-file=/path/to/state/task-state.json
```

Tests use a local Git remote, mocked cloud submissions/PR listing, and a local
HTTP stub for the Claude routine.

The page tests (`tests/test_ui.py` and `tests/test_ui.cjs`) are paused while the
interface settles and skip unless `RUN_UI_TESTS=1` is set.

## Deploy

From this directory, copy `.env.example` to `.env` and set `GITHUB_URL`, a current
runner registration `RUNNER_TOKEN`, `GH_TOKEN` (contents write and pull-request
read), `CODEX_ENV_ID`, `PANEL_TOKEN`, and the audit routine's API trigger as
`CLAUDE_AUDIT_ROUTINE_URL` and `CLAUDE_AUDIT_ROUTINE_TOKEN`. Set the repository Actions secret
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

## Logs

`docker compose logs -f control-panel` is the account of what the panel ran. At the
default `PANEL_LOG_LEVEL=INFO` it writes one line per scheduling pass, one per cloud
dispatch carrying the CLI's own output, and one per started audit session; failures add
the subprocess's exit status and the tail of its stderr, or the routine's HTTP status and
reply. `WARNING` keeps only the failures.

Operator-facing messages in the panel deliberately name no command or argument, so the
log is where that detail goes. Every line passes through `logs.redact` first: the panel's
own secrets, a URL's embedded credentials, and GitHub and Anthropic token shapes are
replaced, and a prompt-sized argument is logged as its length rather than its text.
Treat the log as sensitive anyway — it carries branch names, task IDs and tool output.

Use `PANEL_BIND` and `PANEL_PORT` to override `127.0.0.1:8765`. Site-specific
Compose changes belong in the gitignored `compose.override.yml`; `up.sh` includes
it. Pin reviewed image digests and a tested Codex CLI version for repeatable builds.
