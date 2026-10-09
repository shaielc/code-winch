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
| `control_panel/integrations/github.py` | The only place a GitHub credential enters a command |
| `control_panel/integrations/github_app.py` | App JWT, installation lookup, installation tokens |
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

## GitHub identity

The panel authenticates to GitHub as its own GitHub App, using an installation access
token minted on demand from the App's private key. Nothing it does needs a long-lived
personal access token, and the identities stay separate:

| | |
| --- | --- |
| Task branch opening commit | `Code Winch`, from `GIT_AUTHOR_NAME`/`GIT_AUTHOR_EMAIL` |
| Git and `gh` authentication | the App's installation on this repository |
| Task pull request into main | `<app-name>[bot]` |
| Review and approval | a person, who is not the author |

A pull request the panel opens is therefore reviewable by the account that runs the panel,
which a pull request authored by that account's own token would not be.

`control_panel/integrations/github_app.py` signs a short App JWT with the private key,
resolves the installation covering `GITHUB_URL`, exchanges it for an installation token,
and holds that token until five minutes before it expires. Threads share one mint: the
HTTP server and the sync job cannot each ask for a token of their own. The private key is
never read into the process — `openssl` signs from the mounted file — so it cannot reach a
log, a traceback or a command's arguments.

`github_run` in `control_panel/integrations/github.py` is the only thing that puts that
token in a command's environment, for `git clone`, `fetch`, `pull` and `push` and for the
`gh pr` calls. The panel's own environment holds no GitHub credential, so `codex cloud
exec` — which runs agent-chosen work — cannot read one. Everything local, including the
opening commit itself, runs without a credential.

`task-state.json` is written at schema version 2, which records stages as one list
per stage. A version 1 file, which keyed each submission `"<stage>:<head>"` in a
single map, is migrated when the panel reads it; an unrecognised version is refused
rather than guessed at.

## Run and test

From this directory, Python 3.11+ and Git are sufficient for the tests. Running
the panel also requires GitHub CLI and an authenticated Codex CLI:

```sh
python3 -m unittest discover -s tests -v
GITHUB_URL=https://github.com/OWNER/REPOSITORY \
  GITHUB_APP_ID=... GITHUB_APP_PRIVATE_KEY_FILE=/path/to/app.pem \
  PANEL_TOKEN=... CODEX_ENV_ID=... python3 -m control_panel \
  --clone=/path/to/dedicated-main-checkout --state-file=/path/to/state/task-state.json
```

Starting the panel mints an installation token before it serves, so a GitHub App that is
misconfigured or not installed on `GITHUB_URL` is a startup failure naming what to fix
rather than a panel whose every button fails. It also configures Git's credential helper
and clones main if the checkout is absent. `openssl` is required for signing.

Tests use a local Git remote, mocked cloud submissions/PR listing, local HTTP stubs for
the Claude routine and for GitHub's App endpoints, and a throwaway RSA key. No test
reaches GitHub, and the App's JWT is verified against that key the way GitHub would
verify it.

There are no page tests; `render` is exercised through the scheduler tests that assert on
the HTML it produces.

## Deploy

### The panel's GitHub App

Create a GitHub App owned by the account that owns the repository, and install it on that
repository alone. It needs three repository permissions and no user permissions:

| Permission | Access |
| --- | --- |
| Contents | Read and write |
| Pull requests | Read and write |
| Metadata | Read |

No webhook is required; the merge notification comes from Actions, not from the App. Do
not configure user access tokens or OAuth — the panel uses an installation access token,
which is what makes `<app-name>[bot]` the author of the pull requests it opens.

Generate a private key and save the PEM where Compose can mount it, by default
`./secrets/github-app-private-key.pem` (gitignored). It is mounted read-only at
`/run/secrets/github_app_private_key` and must be readable by the image's unprivileged
`runner` user; `chmod 0444` is the simplest way. Never put the PEM in `.env`.

The repository's `main` ruleset must leave **dismiss stale pull request approvals when new
commits are pushed** disabled, as [ADR-0004](../docs/decisions/0004-task-status-authority.md)
requires: the completion stamp is pushed after approval, and dismissal would invalidate the
approval that triggered it.

### Configuration

From this directory, copy `.env.example` to `.env` and set `GITHUB_URL`, `GITHUB_APP_ID`,
`GITHUB_APP_PRIVATE_KEY_PATH`, a current runner registration `RUNNER_TOKEN`,
`CODEX_ENV_ID`, `PANEL_TOKEN`, and the audit routine's API trigger as
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

`GH_TOKEN` is a temporary fallback for rolling back to a personal access token without
reverting code: it is used only while `GITHUB_APP_ID` is empty, and the panel logs a line
saying so. It, `StaticTokenAuth`, and this paragraph go away once the App path has been
stable. A token set alongside a configured App is ignored, and removed from the panel's
environment either way.

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
own secrets, a URL's embedded credentials, GitHub and Anthropic token shapes — `ghs_`
installation tokens among them — a signed JWT, and a PEM private key are replaced, and a
prompt-sized argument is logged as its length rather than its text. A credential the panel
holds but its environment no longer carries is named to `logs.guard` so redaction keeps
its reach. Treat the log as sensitive anyway — it carries branch names, task IDs and tool
output. One line per mint records which installation the token came from and when it
expires, and never the token.

Use `PANEL_BIND` and `PANEL_PORT` to override `127.0.0.1:8765`. Site-specific
Compose changes belong in the gitignored `compose.override.yml`; `up.sh` includes
it. Pin reviewed image digests and a tested Codex CLI version for repeatable builds.
