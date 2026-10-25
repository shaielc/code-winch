# What Codex offers for reading agent output and sending follow-ups

Reference for P0-009 to P0-012, which automate publishing a Codex agent's changes
and returning review reports to it. It extends
[`codex-cloud-cli.md`](codex-cloud-cli.md), which covers dispatch and status.

Checked on 2026-10-25 against the local `codex-cli 0.154.0` (`--help` of every
subcommand named below), the 0.164.0 release notes, the
[`cloud-tasks-client`](https://github.com/openai/codex/tree/main/codex-rs/cloud-tasks-client)
sources, and the OpenAI Codex documentation. **Not run:** `codex cloud list
--json`, `status`, `diff` and `apply` against a real account, so the JSON field
names below come from the upstream sources and `codex-cloud-cli.md`, not from
output seen here. Re-check after a CLI bump.

## Codex Cloud, as the panel uses it

`codex cloud` has five subcommands, unchanged since `codex-cloud-cli.md`:
`exec`, `status`, `list`, `apply`, `diff`.

| Need | Available | How |
|---|---|---|
| Dispatch on a branch | Yes | `codex cloud exec --env E --branch B "<prompt>"`; prints the task URL only |
| Completion | Yes | `list --json`: `status` is `pending`, `ready`, `applied` or `error`; also `title`, `updated_at`, `summary{files_changed,lines_added,lines_removed}`, `attempt_total`. At most 20 tasks per page; follow `cursor` before treating a task as missing |
| The changes | Yes | `diff <id> [--attempt N]` prints a unified diff; `apply <id>` runs `git apply` in the current directory |
| A commit message or PR text | No | The diff carries none; the panel writes the commit |
| Prompt and agent messages | **No CLI** | See below |
| Follow-up turn into a task | **No** | See below |
| Wait, cancel, logs, webhooks, REST API | No | [openai/codex#24777](https://github.com/openai/codex/issues/24777) asks for them; open, no maintainer reply as of 2026-10-14 |

**Why messages are unreachable.** `cloud-tasks-client` implements
`get_task_text` (prompt plus assistant messages), `get_task_messages` and
`list_sibling_attempts`. They read ChatGPT's `GET …/wham/tasks/{id}` using the
ChatGPT login in the Codex volume, and only the TUI calls them. The endpoint is
not a documented API. The panel does not depend on it.

**Why follow-ups are unreachable.** The `CloudBackend` trait has `create_task`
and no call that adds a turn to an existing task. Sending a report back to an
agent therefore means a new task on the same branch: a new conversation whose
prompt carries the report.

## Surfaces for agents the panel would run itself

These give real session history and follow-ups, but only for a Codex process the
panel starts, not for Codex Cloud tasks. They are recorded as an option, not
used by P0-009 to P0-012.

| Surface | What it gives |
|---|---|
| `codex exec --json` | JSONL events on stdout: `thread.started`, `turn.started`, `item.started`/`item.completed` (agent messages, command executions, file changes, plan updates), `turn.completed` with token usage |
| `codex exec --output-schema F -o OUT` | The final message constrained to a JSON Schema and written to a file |
| `codex exec resume <id\|--last> "<prompt>"`, `codex queue --thread T --message M` | Follow-up into an existing local session |
| `~/.codex/sessions/**/rollout-*.jsonl` | The persisted session history (`session_meta`, `response_item`, `event_msg` lines) |
| `codex app-server` | JSON-RPC over stdio, unix socket or websocket: `thread/read`, `thread/turns/list`, `thread/items/list`, `turn/start`, `turn/steer`, and `turn/diff/updated` with the aggregated unified diff. `generate-json-schema` emits the protocol for the installed version |

## GitHub integration

Commenting `@codex review` on a pull request posts a code review; any other
`@codex …` comment starts a cloud chat with the pull request as context and "can
push a fix back to the branch when it has permission". Whether it reacts to a
comment authored by the panel's GitHub App has not been tested, so the panel does
not rely on it.
