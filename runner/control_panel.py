#!/usr/bin/env python3
"""Serve a local control panel for the Code Winch task scheduler."""

from __future__ import annotations

import argparse
import fcntl
import html
import json
import os
import subprocess
from http import HTTPStatus
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from typing import Any
from urllib.parse import parse_qs, urlencode

STATUS_ORDER = ["in_progress", "blocked", "pending", "completed"]


def load_tracker(path: Path) -> dict[str, Any]:
    return json.loads(path.read_text())


def load_state(path: Path) -> dict[str, Any]:
    if not path.exists():
        return {"schema_version": 1, "tasks": {}}
    state = json.loads(path.read_text())
    if state.get("schema_version") != 1 or not isinstance(state.get("tasks"), dict):
        raise ValueError(f"unsupported scheduler state in {path}")
    return state


def save_state(path: Path, state: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    temporary = path.with_suffix(".tmp")
    temporary.write_text(json.dumps(state, indent=2, sort_keys=True) + "\n")
    temporary.replace(path)


def lock_path(state_file: Path) -> Path:
    return state_file.with_suffix(".lock")


def scheduler_running(state_file: Path) -> bool:
    path = lock_path(state_file)
    if not path.exists():
        return False
    with path.open("a") as handle:
        try:
            fcntl.flock(handle, fcntl.LOCK_EX | fcntl.LOCK_NB)
        except BlockingIOError:
            return True
        fcntl.flock(handle, fcntl.LOCK_UN)
    return False


def expire(state_file: Path, task_id: str) -> str:
    path = lock_path(state_file)
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("a") as handle:
        try:
            fcntl.flock(handle, fcntl.LOCK_EX | fcntl.LOCK_NB)
        except BlockingIOError:
            return "the scheduler is running right now; try again once it finishes"
        state = load_state(state_file)
        if task_id not in state["tasks"]:
            return f"{task_id} has no local override to expire"
        del state["tasks"][task_id]
        save_state(state_file, state)
        return f"expired the local override for {task_id}"


def git(*arguments: str, cwd: Path | None = None) -> None:
    subprocess.run(("git", *arguments), cwd=cwd, check=True, capture_output=True, text=True)


def ensure_repo(repo: Path, url: str, ref: str) -> None:
    if not (repo / ".git").exists():
        repo.parent.mkdir(parents=True, exist_ok=True)
        git("clone", "--filter=blob:none", url, str(repo))
    git("fetch", "--quiet", "origin", ref, cwd=repo)
    git("reset", "--quiet", "--hard", f"origin/{ref}", cwd=repo)


def run_scheduler(
    repo: Path, url: str, ref: str, state_file: Path, environment: str, task_id: str | None
) -> str:
    if not url:
        return "set GITHUB_URL so the panel knows which repository to schedule from"
    if not environment:
        return "set CODEX_ENV_ID in runner/.env so the panel can dispatch to Codex Cloud"
    try:
        ensure_repo(repo, url, ref)
    except subprocess.CalledProcessError as error:
        return f"could not update the panel's clone: {(error.stderr or '').strip()[:300]}"

    command = [
        "python3",
        "scripts/task_scheduler.py",
        "--env",
        environment,
        "--state-file",
        str(state_file),
    ]
    if task_id:
        command += ["--task", task_id]
    result = subprocess.run(
        command, cwd=repo, capture_output=True, text=True, timeout=900, check=False
    )
    output = (result.stdout + result.stderr).strip() or "scheduler produced no output"
    if result.returncode != 0:
        return f"scheduler exited {result.returncode}: {output[:500]}"
    return output[:500]


def rows(tracker: dict[str, Any], state: dict[str, Any]) -> list[dict[str, Any]]:
    overrides = state["tasks"]
    result = []
    for task in tracker["tasks"]:
        local = overrides.get(task["id"])
        settled = task["status"] == "completed"
        applied = local if local and not settled else None
        result.append(
            {
                "id": task["id"],
                "title": task["title"],
                "depends_on": task["depends_on"],
                "tracked": task["status"],
                "status": applied["status"] if applied else task["status"],
                "owner": (applied or task).get("owner"),
                "updated_at": (applied or {}).get("updated_at"),
                "local": bool(applied),
                "stale": bool(local and settled),
            }
        )
    return result


def available_ids(entries: list[dict[str, Any]]) -> set[str]:
    done = {entry["id"] for entry in entries if entry["status"] == "completed"}
    return {
        entry["id"]
        for entry in entries
        if entry["status"] == "pending" and set(entry["depends_on"]) <= done
    }


PAGE = """<!doctype html>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Code Winch scheduler</title>
<style>
  :root {{ color-scheme: light dark; --line: #8883; }}
  body {{ font: 15px/1.6 system-ui, sans-serif; margin: 0 auto; padding: 2.5rem 2rem; max-width: 92rem; }}
  header {{ display: flex; align-items: baseline; gap: 1rem; flex-wrap: wrap; margin-bottom: .4rem; }}
  h1 {{ font-size: 1.3rem; margin: 0; }}
  .sub {{ opacity: .7; font-size: .875rem; margin-bottom: 1.75rem; }}
  .msg {{ padding: .75rem 1rem; border: 1px solid var(--line); border-radius: 8px; margin-bottom: 1.5rem; }}
  .wrap {{ overflow-x: auto; }}
  table {{ border-collapse: collapse; width: 100%; font-size: .9rem; }}
  th, td {{ text-align: left; padding: .8rem .75rem; border-bottom: 1px solid var(--line); white-space: nowrap; }}
  th {{ font-weight: 600; opacity: .65; font-size: .78rem; text-transform: uppercase; letter-spacing: .06em; }}
  tbody tr:hover td {{ background: #8881; }}
  td.title {{ white-space: normal; min-width: 15rem; line-height: 1.45; }}
  td.deps {{ white-space: normal; min-width: 15rem; max-width: 19rem; line-height: 1.9; }}
  td.owner {{ max-width: 11rem; overflow: hidden; text-overflow: ellipsis; }}
  td.deps code {{ font-size: .78rem; padding: .1rem .4rem; border-radius: 4px; background: #8881; opacity: .55; }}
  td.deps code.unmet {{ opacity: 1; background: #f9731633; font-weight: 600; }}
  td.actions {{ text-align: right; }}
  .tag {{ font-size: .75rem; padding: .15rem .6rem; border: 1px solid var(--line); border-radius: 999px; }}
  .tag.completed {{ color: #16a34a; border-color: #16a34a66; background: #16a34a1a; font-weight: 600; }}
  .local {{ font-weight: 600; }}
  .stale td {{ background: #f9731622; }}
  .note {{ opacity: .7; font-size: .8rem; margin-top: 1.75rem; max-width: 52rem; }}
  form {{ display: inline; }}
  button {{ font: inherit; font-size: .8rem; padding: .3rem .8rem; margin-left: .4rem; cursor: pointer; border-radius: 6px; }}
</style>
<header>
  <h1>Code Winch scheduler</h1>
  <form method="post" action="/run"><button type="submit">Run scheduler</button></form>
</header>
<div class="sub">{summary}</div>
{message}
<div class="wrap">
<table>
  <thead><tr><th>Task</th><th>Title</th><th>Depends on</th><th>Status</th><th>Source</th><th>Owner</th><th>Updated</th><th></th></tr></thead>
  <tbody>
  {rows}
  </tbody>
</table>
</div>
<p class="note">Leases live in the scheduler state volume and apply only while the tracker has not
recorded the task as completed. Expiring one releases the concurrency slot it holds. Rows highlighted
in orange hold a lease the tracker has already superseded; the scheduler clears those on its next run.</p>
"""

ROW = """<tr class="{cls}">
  <td><code>{id}</code></td>
  <td class="title">{title}</td>
  <td class="deps">{deps}</td>
  <td><span class="tag {status}">{status}</span></td>
  <td>{source}</td>
  <td class="owner" title="{owner}">{owner}</td>
  <td>{updated}</td>
  <td class="actions">{actions}</td>
</tr>"""


def button(action: str, task_id: str, label: str) -> str:
    return (
        f'<form method="post" action="{action}">'
        f'<input type="hidden" name="task_id" value="{html.escape(task_id)}">'
        f'<button type="submit">{label}</button></form>'
    )


def render(tracker: dict[str, Any], state: dict[str, Any], message: str, busy: bool) -> str:
    entries = rows(tracker, state)
    runnable = available_ids(entries)
    done = {entry["id"] for entry in entries if entry["status"] == "completed"}
    counts: dict[str, int] = {name: 0 for name in STATUS_ORDER}
    for entry in entries:
        counts[entry["status"]] = counts.get(entry["status"], 0) + 1
    summary = " · ".join(f"{counts[name]} {name.replace('_', ' ')}" for name in STATUS_ORDER)
    summary += f" · {len(runnable)} available"
    if busy:
        summary += " · scheduler running"

    body = []
    for entry in entries:
        actions = ""
        if entry["id"] in runnable:
            actions += button("/run", entry["id"], "Run")
        if entry["local"] or entry["stale"]:
            actions += button("/expire", entry["id"], "Expire")
        source = "tracker"
        if entry["local"]:
            source = '<span class="local">lease</span>'
        elif entry["stale"]:
            source = "tracker (stale lease)"
        deps = " ".join(
            f'<code class="{"" if dependency in done else "unmet"}">{html.escape(dependency)}</code>'
            for dependency in entry["depends_on"]
        )
        body.append(
            ROW.format(
                cls="stale" if entry["stale"] else "",
                id=html.escape(entry["id"]),
                title=html.escape(entry["title"]),
                deps=deps or "—",
                status=html.escape(entry["status"]),
                source=source,
                owner=html.escape(entry["owner"] or "—"),
                updated=html.escape((entry["updated_at"] or "—")[:16].replace("T", " ")),
                actions=actions,
            )
        )

    banner = f'<div class="msg">{html.escape(message)}</div>' if message else ""
    return PAGE.format(summary=html.escape(summary), message=banner, rows="\n  ".join(body))


class Handler(BaseHTTPRequestHandler):
    tracker_path: Path
    state_file: Path
    repo: Path
    url: str
    ref: str
    environment: str

    def log_message(self, *args: Any) -> None:
        return

    def _send(self, status: HTTPStatus, body: bytes) -> None:
        self.send_response(status)
        self.send_header("Content-Type", "text/html; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def _redirect(self, message: str) -> None:
        self.send_response(HTTPStatus.SEE_OTHER)
        self.send_header("Location", "/?" + urlencode({"msg": message}))
        self.end_headers()

    def do_GET(self) -> None:
        path, _, query = self.path.partition("?")
        if path != "/":
            self._send(HTTPStatus.NOT_FOUND, b"not found")
            return
        message = parse_qs(query).get("msg", [""])[0]
        try:
            page = render(
                load_tracker(self.tracker_path),
                load_state(self.state_file),
                message,
                scheduler_running(self.state_file),
            )
        except (OSError, ValueError) as error:
            self._send(HTTPStatus.INTERNAL_SERVER_ERROR, str(error).encode())
            return
        self._send(HTTPStatus.OK, page.encode())

    def do_POST(self) -> None:
        if self.path not in ("/expire", "/run"):
            self._send(HTTPStatus.NOT_FOUND, b"not found")
            return
        length = int(self.headers.get("Content-Length") or 0)
        form = parse_qs(self.rfile.read(length).decode())
        task_id = (form.get("task_id") or [""])[0].upper()

        if self.path == "/run":
            self._redirect(
                run_scheduler(
                    self.repo,
                    self.url,
                    self.ref,
                    self.state_file,
                    self.environment,
                    task_id or None,
                )
            )
            return
        self._redirect(expire(self.state_file, task_id) if task_id else "no task selected")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--tracker", type=Path)
    parser.add_argument("--state-file", type=Path, required=True)
    parser.add_argument("--repo", type=Path, default=Path("/home/runner/panel-repo"))
    parser.add_argument("--github-url", default=os.environ.get("GITHUB_URL", ""))
    parser.add_argument("--ref", default="main")
    parser.add_argument("--host", default="127.0.0.1")
    parser.add_argument("--port", type=int, default=8765)
    parser.add_argument("--expire", metavar="TASK_ID")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    if args.expire:
        print(expire(args.state_file, args.expire.upper()))
        return 0

    Handler.tracker_path = args.tracker or args.repo / "docs/workplan/tasks.json"
    Handler.state_file = args.state_file
    Handler.repo = args.repo
    Handler.url = args.github_url
    Handler.ref = args.ref
    Handler.environment = os.environ.get("CODEX_ENV_ID", "")

    if args.github_url and not (args.repo / ".git").exists():
        try:
            ensure_repo(args.repo, args.github_url, args.ref)
        except subprocess.CalledProcessError as error:
            print(f"could not clone {args.github_url}: {(error.stderr or '').strip()}")
    server = ThreadingHTTPServer((args.host, args.port), Handler)
    print(f"control panel on http://{args.host}:{args.port}")
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        pass
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
