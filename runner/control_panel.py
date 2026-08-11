#!/usr/bin/env python3
"""Serve a local control panel for the Code Winch task scheduler."""

from __future__ import annotations

import argparse
import fcntl
import html
import json
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


def rows(tracker: dict[str, Any], state: dict[str, Any]) -> list[dict[str, Any]]:
    overrides = state["tasks"]
    result = []
    for task in tracker["tasks"]:
        local = overrides.get(task["id"])
        result.append(
            {
                "id": task["id"],
                "title": task["title"],
                "phase": task["phase"],
                "tracked": task["status"],
                "status": local["status"] if local else task["status"],
                "owner": (local or task).get("owner"),
                "updated_at": (local or {}).get("updated_at"),
                "error": (local or {}).get("launch_error"),
                "local": bool(local),
                "masking": bool(local and local["status"] != task["status"]),
            }
        )
    return result


PAGE = """<!doctype html>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Code Winch scheduler</title>
<style>
  :root {{ color-scheme: light dark; --line: #8883; }}
  body {{ font: 15px/1.5 system-ui, sans-serif; margin: 0 auto; padding: 2rem 1.5rem; max-width: 70rem; }}
  h1 {{ font-size: 1.3rem; margin: 0 0 .25rem; }}
  .sub {{ opacity: .7; font-size: .875rem; margin-bottom: 1.5rem; }}
  .msg {{ padding: .6rem .8rem; border: 1px solid var(--line); border-radius: 6px; margin-bottom: 1.25rem; }}
  .wrap {{ overflow-x: auto; }}
  table {{ border-collapse: collapse; width: 100%; font-size: .9rem; }}
  th, td {{ text-align: left; padding: .5rem .6rem; border-bottom: 1px solid var(--line); white-space: nowrap; }}
  th {{ font-weight: 600; opacity: .7; font-size: .8rem; text-transform: uppercase; letter-spacing: .04em; }}
  td.title {{ white-space: normal; min-width: 20rem; }}
  .tag {{ font-size: .75rem; padding: .1rem .45rem; border: 1px solid var(--line); border-radius: 999px; }}
  .local {{ font-weight: 600; }}
  .masking td {{ background: #f9731622; }}
  .note {{ opacity: .7; font-size: .8rem; }}
  button {{ font: inherit; font-size: .8rem; padding: .2rem .6rem; cursor: pointer; }}
</style>
<h1>Code Winch scheduler</h1>
<div class="sub">{summary}</div>
{message}
<div class="wrap">
<table>
  <tr><th>Task</th><th>Title</th><th>Status</th><th>Source</th><th>Owner</th><th>Updated</th><th></th></tr>
  {rows}
</table>
</div>
<p class="note">Local overrides live in the scheduler state volume and take precedence over the tracker.
Expiring one drops the override so the task falls back to its tracked status.</p>
"""

ROW = """<tr class="{cls}">
  <td><code>{id}</code></td>
  <td class="title">{title}</td>
  <td><span class="tag">{status}</span></td>
  <td>{source}</td>
  <td>{owner}</td>
  <td>{updated}</td>
  <td>{action}</td>
</tr>"""


def render(tracker: dict[str, Any], state: dict[str, Any], message: str, busy: bool) -> str:
    entries = rows(tracker, state)
    counts = {name: 0 for name in STATUS_ORDER}
    for entry in entries:
        counts[entry["status"]] = counts.get(entry["status"], 0) + 1
    summary = " · ".join(f"{counts[name]} {name.replace('_', ' ')}" for name in STATUS_ORDER)
    if busy:
        summary += " · scheduler running"

    body = []
    for entry in entries:
        action = ""
        if entry["local"]:
            action = (
                '<form method="post" action="/expire">'
                f'<input type="hidden" name="task_id" value="{html.escape(entry["id"])}">'
                "<button type=\"submit\">Expire</button></form>"
            )
        source = "local" if entry["local"] else "tracker"
        if entry["masking"]:
            source = f"local (tracker says {html.escape(entry['tracked'])})"
        body.append(
            ROW.format(
                cls="masking" if entry["masking"] else "",
                id=html.escape(entry["id"]),
                title=html.escape(entry["title"]),
                status=html.escape(entry["status"]),
                source=f'<span class="{"local" if entry["local"] else ""}">{source}</span>',
                owner=html.escape(entry["owner"] or "—"),
                updated=html.escape((entry["updated_at"] or "—")[:19]),
                action=action,
            )
        )

    banner = f'<div class="msg">{html.escape(message)}</div>' if message else ""
    return PAGE.format(summary=html.escape(summary), message=banner, rows="\n  ".join(body))


class Handler(BaseHTTPRequestHandler):
    tracker_path: Path
    state_file: Path

    def log_message(self, fmt: str, *args: Any) -> None:
        return

    def _send(self, status: HTTPStatus, body: bytes) -> None:
        self.send_response(status)
        self.send_header("Content-Type", "text/html; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

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
        if self.path != "/expire":
            self._send(HTTPStatus.NOT_FOUND, b"not found")
            return
        length = int(self.headers.get("Content-Length") or 0)
        form = parse_qs(self.rfile.read(length).decode())
        task_id = (form.get("task_id") or [""])[0].upper()
        message = expire(self.state_file, task_id) if task_id else "no task selected"
        self.send_response(HTTPStatus.SEE_OTHER)
        self.send_header("Location", "/?" + urlencode({"msg": message}))
        self.end_headers()


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--tracker", type=Path, required=True)
    parser.add_argument("--state-file", type=Path, required=True)
    parser.add_argument("--host", default="127.0.0.1")
    parser.add_argument("--port", type=int, default=8765)
    parser.add_argument("--expire", metavar="TASK_ID")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    if args.expire:
        print(expire(args.state_file, args.expire.upper()))
        return 0

    Handler.tracker_path = args.tracker
    Handler.state_file = args.state_file
    server = ThreadingHTTPServer((args.host, args.port), Handler)
    print(f"control panel on http://{args.host}:{args.port}")
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        pass
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
