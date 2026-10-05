# What Claude Code cloud sessions tell us

Reference for dispatching work to Claude Code on the web from the CLI, the
counterpart to [`codex-cloud-cli.md`](codex-cloud-cli.md). Checked against the
installed `claude --help` and the official docs:
[CLI reference](https://code.claude.com/docs/en/cli-reference.md) and
[Claude Code on the web](https://code.claude.com/docs/en/claude-code-on-the-web.md).
Flags change between releases, so re-check after a CLI bump.

## Dispatch

```sh
claude --cloud "<task description>"
```

`--cloud [description|session_id|url]` creates a cloud session from a
description, or addresses an existing one by session ID or `claude.ai/code`
URL.

The cloud VM clones the current directory's GitHub remote at the **current
branch**, not the local checkout, so the branch must be pushed first. There is
no `--branch` flag; to target `task/<ID>`, check it out (and push it) before
dispatching, or use `--ref` (below).

Network access, environment variables, and setup scripts come from the saved
[cloud environment](https://code.claude.com/docs/en/cloud-environments), not
from CLI flags. Onboarding creates a **Default** environment.

## Flags that apply to the cloud session

| Flag | Effect | Source |
| --- | --- | --- |
| `--environment <ccpool_...>` | Run the new session on a self-hosted environment instead of the default. | `claude --help`, CLI reference |
| `--ref <ref>` | With `--environment`, base the checkout on a named ref instead of local `HEAD`. | CLI reference only — absent from the installed `--help`, so verify before relying on it. |
| `-p "<message>" --cloud <session-id>` | Queue a follow-up into an existing session and exit without waiting for a reply. | Claude Code on the web |

## Undocumented for `--cloud`

- `--permission-mode` — the docs say a cloud session's mode is picked when the
  task is created, but only describe the web UI dropdown, not the CLI flag.
- `--model`, `--effort`, `-n/--name`, `--agent` — not documented as carrying
  over. Inside a session, `/model` and `/effort` work.
- `--add-dir`, `--settings`, `--mcp-config`, `--plugin-dir`, `--worktree` —
  local-client concerns; almost certainly ignored by the cloud session
  (inference, not stated in the docs).

A cheap probe settles the first two bullets:
`claude --cloud "say hi" --model sonnet -n test`, then check claude.ai/code.

## Related commands

| Command | Effect |
| --- | --- |
| `claude --teleport [session]` | Pull a cloud session into the local terminal. Requires claude.ai subscription auth; the session's branch must be pushed. |
| `claude --cloud <session-id>` (no `-p`) | Attach interactively. Listed in `--help`, but the docs record the error "Attaching to an existing cloud session is not enabled for your account", so it may be gated. |
| `claude --remote-control [name]` | The reverse direction: a local session drivable from claude.ai or mobile. |
| `/schedule` | Recurring or one-off scheduled cloud runs (routines). |

## The CLI cannot dispatch headlessly

Checked from the panel container, where nothing has a TTY:

```text
$ claude --cloud "Reply with OK and stop"
Error: --cloud requires an interactive terminal.

$ claude -p "Reply with OK and stop" --cloud
Error: --cloud cannot be combined with --print.
Starting a new cloud session with --cloud is interactive only ...
```

Both exit 1 and create nothing. Unlike `codex cloud exec`, `claude --cloud`
cannot be scripted, so the panel does not use the CLI.

## Routine API trigger

The documented non-interactive way to start a cloud session is a
[routine](https://code.claude.com/docs/en/routines.md) with an API trigger
([endpoint reference](https://platform.claude.com/docs/en/api/claude-code/routines-fire)):

```text
POST https://api.anthropic.com/v1/claude_code/routines/trig_.../fire
Authorization: Bearer sk-ant-oat01-...
anthropic-version: 2023-06-01
{"text": "<up to 65,536 characters>"}

200 {"type": "routine_fire",
     "claude_code_session_id": "session_01...",
     "claude_code_session_url": "https://claude.ai/code/session_01..."}
```

- The routine fixes the repository and the prompt. Each run clones the default
  branch unless the routine's prompt says otherwise.
- `text` reaches the session inside a `<routine-fire-payload>` block marked as
  untrusted. Claude acts on it only if the routine's own prompt says to.
- The trigger and its token can only be created on the web at
  claude.ai/code/routines; the CLI cannot create or revoke tokens. A token is
  shown once and is scoped to its one routine.
- Limits: 30 fires per hour per routine (shared with **Run now**) and 100 API
  fires per hour per account. Over the limit returns 429 with `Retry-After`.
- The API is marked experimental: request and response shapes, limits and token
  semantics may change.

The call returns as soon as the session exists. Nothing reports when the
session finishes, so completion is reconciled through the pull request.

## Control-panel integration

**Audit** fires the routine at `CLAUDE_AUDIT_ROUTINE_URL` with
`CLAUDE_AUDIT_ROUTINE_TOKEN`, sending the formatted `prompts/audit.md` as `text`
(`control_panel/integrations/claude.py`). The routine's own prompt accepts only
an audit request naming a task, a pull request in this repository and its head
SHA. It then checks out that head and may write nothing except a comment-only
review.

The panel records the returned session URL per pull-request head, the way
Refine and Implement record Codex task URLs per branch head. Any non-2xx answer
means no session started, so nothing is recorded. A timeout, or a 200 without a
session URL, leaves the attempt with an unknown outcome to be checked by hand.
