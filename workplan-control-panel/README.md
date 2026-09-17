# Workplan control panel

The control panel owns task orchestration and exposes it as an HTTP API. The
self-hosted GitHub runner only posts merge events. The browser UI remains at
`http://localhost:8765/`, with the existing table and dependency-tree views.

## Flow

1. A pull request merges into `main`. The GitHub workflow posts its event to
   `POST /api/events/merge`. Merges into task branches do not trigger scheduling.
2. The panel pulls its dedicated `main` checkout with `git pull --ff-only` and
   reads `docs/workplan/tasks.json` (the repository's task tracker).
3. It selects pending tasks whose dependencies are completed, up to three active
   tasks by default (`--max-concurrent`), and reserves them in durable local state.
4. It creates `task/<ID>` from that main revision and pushes an opening commit
   marking only that task `in_progress`. Repeated events reuse the branch and
   commit; interrupted preparation resumes on the next sync. Main stays checked out.
5. The UI exposes **Refine**, **Implement**, and **Audit** for prepared tasks.
   Refine and Implement submit `codex cloud exec --branch task/<ID>` using their
   respective templates in `scripts/`. Merge refinement changes into the task
   branch before choosing Implement. Each agent is instructed to target that
   branch with its pull request.
6. Audit resolves an open implementation PR into the task branch and its current
   head SHA, formats `scripts/task-audit-prompt.md`, and copies it to the clipboard.
   If several PRs exist, the UI asks which number to use. It also displays the
   prompt and a copy button, including when clipboard access is unavailable.

The tracker on main is the authority for completion. Local records reserve work
and retain cloud-task links; merge-event titles do not mark tasks completed.
Refinement and implementation PRs target the task branch; the final task PR
into main is the one the existing approval gate stamps completed.

## API

All `/api/` requests require `Authorization: Bearer <PANEL_TOKEN>`. POST bodies
must be JSON objects. The UI asks for the same token and keeps it in page memory.
The local UI and `/healthz` are readable without a token; keep the default
loopback binding or protect the entire site through your reverse proxy.

| Method | Path | Behavior |
| --- | --- | --- |
| GET | `/healthz` | Process health |
| GET | `/api/tasks` | Tracker snapshot and local task/stage records |
| POST | `/api/events/merge` | GitHub closed-PR event; only merges into main schedule |
| POST | `/api/sync` | Pull main and prepare available tasks; body `{}` |
| POST | `/api/tasks/<ID>/refine` | Submit the refinement prompt; body `{}` |
| POST | `/api/tasks/<ID>/implement` | Submit the implementation prompt; body `{}` |
| POST | `/api/tasks/<ID>/audit` | Return the audit prompt; optional `{"pr_number": 123}` |

Operations share a state-file lock. Concurrent requests return 409 and can be
retried. Repeating a successfully submitted stage at the same branch revision
returns its existing URL. A submission with an uncertain outcome stays reserved:
check Codex first, then stop the panel and remove only that entry from the task's
`stages` map in `task-state.json` before retrying. Do not remove the whole task.
This prevents a lost response from silently starting a second cloud task.

The former scheduler CLI is now an HTTP client as well:

```sh
PANEL_TOKEN=... python3 scripts/task_scheduler.py --url http://localhost:8765
```

## Install

From the repository root, copy `workplan-control-panel/.env.example` to
`workplan-control-panel/.env` and set:

| Variable | Value |
| --- | --- |
| `GITHUB_URL` | HTTPS repository URL |
| `RUNNER_TOKEN` | Current registration token for the repository's Actions runner |
| `GH_TOKEN` | Panel credential with repository contents write and pull-request read permissions |
| `PANEL_TOKEN` | Shared API secret; also set Actions secret `CONTROL_PANEL_TOKEN` to this value |
| `CODEX_ENV_ID` | Codex Cloud environment ID or label |

```sh
docker compose --env-file workplan-control-panel/.env -f workplan-control-panel/compose.yml build
./workplan-control-panel/codex_login.sh
docker compose --env-file workplan-control-panel/.env -f workplan-control-panel/compose.yml up -d
```

The panel clones main on first start. **Sync main** in the UI, or run the
**Schedule available tasks** workflow, to prepare the initial tasks. Startup
itself does not dispatch tasks. Scripts and prompts are baked into the image;
rebuild after changing them.

The workflow defaults to `http://control-panel:8765` on the Compose network. Set
Actions variable `CONTROL_PANEL_URL` if your runner reaches the panel at another
address. Use HTTPS when crossing hosts. Only the panel needs the Codex environment
and repository-write credentials; the runner receives the API secret from Actions.

`runner-config`, `runner-work`, `panel-checkout`, `scheduler-state`, and
`codex-auth` volumes persist independently. The runner has no mount of the panel
checkout, state, or Codex credentials. When upgrading the old shared-volume
installation, stop it first and preserve its `task-state.json` in the new
scheduler-state volume and its runner registration in runner-config; do not run
old and new schedulers simultaneously.

Use `PANEL_BIND` and `PANEL_PORT` for local binding changes. Site-specific Compose
changes belong in the gitignored `compose.override.yml`; `up.sh` includes it.
Pin reviewed image digests and a tested `CODEX_VERSION` for repeatable deployments.

## Verification

`python3 -m unittest discover -s tests -v` exercises the API against a real local
Git remote, including fresh main updates, dependency selection, branch/opening
commits, retries, locking, authentication, both cloud prompts, and audit PR/head
formatting. Cloud submissions and PR listing are mocked; no live provider account
is required by the tests.
