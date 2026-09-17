# What Codex Cloud tells us

Reference for what `scripts/task_scheduler.py` can learn about the tasks it
dispatches. Checked against `codex-cli` 0.144.5 and the `cloud-tasks` sources in
[openai/codex](https://github.com/openai/codex/blob/main/codex-rs/cloud-tasks/src/lib.rs);
`.env.example` pins `CODEX_VERSION=latest`, so re-check after a CLI bump.

## What dispatch returns today

`codex cloud exec` prints one line — the task URL — and exits. Its whole
implementation of the result is:

```rust
let created = CloudBackend::create_task(&*ctx.backend, &env_id, &prompt, &git_ref, false, attempts).await?;
println!("{}", util::task_url(&ctx.base_url, &created.id.0));
```

So the whole result is `https://chatgpt.com/codex/tasks/task_e_<hex>` and
nothing else. Dispatch is fire-and-forget: no completion, no diff, no pull
request, and an exit code that only reports whether the task was *submitted*.
The scheduler keeps that line as `task_url` on the lease; the task ID any
follow-up call needs is its last path segment.

## What the CLI can tell us

| Command | Result |
| --- | --- |
| `codex cloud list --json [--env E] [--limit 1-20] [--cursor C]` | The only JSON surface. Per task: `id`, `url`, `title`, `status`, `updated_at`, `environment_id`, `environment_label`, `summary{files_changed, lines_added, lines_removed}`, `is_review`, `attempt_total`, plus a paging `cursor`. |
| `codex cloud status <task_id>` | Human-readable text, but exits 1 unless the task is `ready`. |
| `codex cloud diff <task_id> [--attempt N]` | Unified diff on stdout. |
| `codex cloud apply <task_id> [--attempt N]` | Applies the diff locally, prints an outcome message, exits 1 unless it applied cleanly. |

Task status is a four-state enum: `pending`, `ready`, `applied`, `error`.

Two `exec` flags we do not pass: `--attempts N` runs best-of-N and shows up as
`attempt_total`, and `--branch` sets the base branch. `--env` accepts an
environment *label* such as `shaielc/code-winch`, not only the opaque ID, so
`CODEX_ENV_ID` does not strictly have to be the ID.

Between the `status` exit code and one `list --json` sweep, an in-flight lease
can be reconciled without a human opening the web UI. That is the piece the
scheduler is missing: leases enter `in_progress` and only ever leave it by
expiry or by a merged pull request.

## What we cannot get

`cloud-tasks-client` exposes `get_task_messages`, `get_task_text` (the creating
prompt plus assistant messages), `list_sibling_attempts`, and
`apply_task_preflight`, and it can list environments. The TUI uses all of them;
no CLI subcommand does. So there is no scriptable access to the agent's summary
text, its logs, or the environment list, and no `wait`, no `cancel`, no
webhooks, and no `--json` outside `list`.

There is no public REST API either. The `/api/codex/*` paths the CLI calls are
ChatGPT backend endpoints authenticated with the ChatGPT login in the Codex
volume, not a documented platform API, and [OpenAI's Codex Cloud
docs](https://learn.chatgpt.com/docs/cloud) describe only the web UI flow.
[openai/codex#24777](https://github.com/openai/codex/issues/24777) tracks the
request for scriptable task and environment lifecycle management.

## Consequences for the scheduler

`exec` with no `--branch` resolves the base to the *current* branch of its
working directory. `dispatch()` runs with `cwd=repo_root`, so tasks are based
on whatever branch the checkout happens to be on, while `load_tracker()` reads
the tracker from `origin/main`. Passing `--branch` closes that gap.

In the path that matters this already lands correctly: the workflow runs on a
merged pull request and checks out `github.event.pull_request.base.ref`, so the
runner sits on an up-to-date `main` and dispatched tasks branch from it — the
same revision the tracker was read from. The gap is latent rather than active,
and it opens as soon as the scheduler is run by hand from a feature branch.
`--branch` would make the intent explicit instead of incidental.

