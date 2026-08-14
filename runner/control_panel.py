#!/usr/bin/env python3
"""Serve a local control panel for the Code Winch task scheduler."""

from __future__ import annotations

import argparse
import fcntl
import html
import json
import logging
import os
import subprocess
from collections.abc import Iterator
from contextlib import contextmanager
from http import HTTPStatus
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from typing import Any
from urllib.parse import parse_qs, urlencode, urlparse

LOGGER = logging.getLogger("control_panel")
STATUS_ORDER = ["in_progress", "blocked", "pending", "completed"]
SNAPSHOT_HINT = "run the Schedule available tasks workflow once so the runner publishes one"


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


@contextmanager
def scheduler_lock(state_file: Path) -> Iterator[bool]:
    """Hold the scheduler lock, yielding False when the scheduler already owns it."""
    path = lock_path(state_file)
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("a") as handle:
        try:
            fcntl.flock(handle, fcntl.LOCK_EX | fcntl.LOCK_NB)
        except BlockingIOError:
            yield False
            return
        try:
            yield True
        finally:
            fcntl.flock(handle, fcntl.LOCK_UN)


def refresh(state_file: Path, tracker: dict[str, Any]) -> tuple[dict[str, Any], bool]:
    """Read the state with the overrides the tracker superseded retired, and report scheduler activity.

    Mirrors the retirement the scheduler does on every run so a tracker that caught up with an
    override does not leave the override sitting in the state volume until the next dispatch. The
    entry stays behind as a record of the pull request and Codex task the work went through.
    """
    completed = {task["id"] for task in tracker["tasks"] if task["status"] == "completed"}
    with scheduler_lock(state_file) as acquired:
        state = load_state(state_file)
        superseded = sorted(
            task_id
            for task_id in completed & set(state["tasks"])
            if state["tasks"][task_id]["status"] != "completed"
        )
        for task_id in superseded:
            state["tasks"][task_id]["status"] = "completed"
            state["tasks"][task_id]["owner"] = None
        if superseded:
            LOGGER.warning(
                "tracker superseded the overrides for %s (%s)",
                ", ".join(superseded),
                "retired" if acquired else "kept; the scheduler holds the state file",
            )
        if superseded and acquired:
            save_state(state_file, state)
        return state, not acquired


def expire(state_file: Path, task_id: str) -> str:
    with scheduler_lock(state_file) as acquired:
        if not acquired:
            return "the scheduler is running right now; try again once it finishes"
        state = load_state(state_file)
        if task_id not in state["tasks"]:
            return f"{task_id} has no local override to expire"
        del state["tasks"][task_id]
        save_state(state_file, state)
        return f"expired the local override for {task_id}"


def run_scheduler(
    repo_root: Path, tracker: Path, state_file: Path, environment: str, task_id: str | None
) -> str:
    if not environment:
        return "set CODEX_ENV_ID in runner/.env so the panel can dispatch to Codex Cloud"
    if not tracker.exists():
        return f"no tracker snapshot at {tracker}; {SNAPSHOT_HINT}"

    command = [
        "python3",
        "scripts/task_scheduler.py",
        "--env",
        environment,
        "--repo-root",
        str(repo_root),
        "--tracker-file",
        str(tracker),
        "--state-file",
        str(state_file),
    ]
    if task_id:
        command += ["--task", task_id]
    result = subprocess.run(
        command, cwd=repo_root, capture_output=True, text=True, timeout=900, check=False
    )
    output = (result.stdout + result.stderr).strip() or "scheduler produced no output"
    if result.returncode != 0:
        return f"scheduler exited {result.returncode}: {output[:500]}"
    return output[:500]


def rows(tracker: dict[str, Any], state: dict[str, Any]) -> list[dict[str, Any]]:
    overrides = state["tasks"]
    result = []
    for task in tracker["tasks"]:
        # A retired entry no longer overrides the tracker, but it still carries the links.
        record = overrides.get(task["id"], {})
        lease = record if task["status"] != "completed" else None
        result.append(
            {
                "id": task["id"],
                "title": task["title"],
                "depends_on": task["depends_on"],
                "status": lease["status"] if lease else task["status"],
                "owner": (lease or task).get("owner"),
                "updated_at": record.get("updated_at"),
                "task_url": record.get("task_url"),
                "pull_request": record.get("pull_request"),
                "local": bool(lease),
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


def forest(entries: list[dict[str, Any]]) -> list[dict[str, Any]]:
    """Nest every task under the dependency that unlocks it last.

    Dependencies form a graph, not a tree: a task can wait on several others. Hanging it off the
    deepest of them keeps each task in exactly one place, directly below whatever gates it, and
    leaves the dependencies that lost the tie to render as chips on the node. The trailing pass
    picks up whatever the roots never reached, which a dependency cycle would leave behind.
    """
    known = {entry["id"]: entry for entry in entries}
    depth: dict[str, int] = {}

    def rank(task_id: str, walked: frozenset[str] = frozenset()) -> int:
        if task_id in depth:
            return depth[task_id]
        entry = known.get(task_id)
        if entry is None or task_id in walked:
            return 0
        depth[task_id] = 1 + max(
            (rank(other, walked | {task_id}) for other in entry["depends_on"]), default=-1
        )
        return depth[task_id]

    parents: dict[str, str | None] = {}
    children: dict[str | None, list[dict[str, Any]]] = {}
    for entry in entries:
        options = [
            dependency
            for dependency in entry["depends_on"]
            if dependency in known and dependency != entry["id"]
        ]
        parent = max(options, key=rank) if options else None
        parents[entry["id"]] = parent
        children.setdefault(parent, []).append(entry)

    placed: set[str] = set()

    def build(entry: dict[str, Any]) -> dict[str, Any]:
        placed.add(entry["id"])
        waiting = [child for child in children.get(entry["id"], []) if child["id"] not in placed]
        return {
            "entry": entry,
            "waits_on": [
                dependency
                for dependency in entry["depends_on"]
                if dependency != parents[entry["id"]]
            ],
            "children": [build(child) for child in waiting],
        }

    trees = [build(entry) for entry in children.get(None, [])]
    return trees + [build(entry) for entry in entries if entry["id"] not in placed]


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
  tbody tr {{ scroll-margin-top: 1.5rem; }}
  tbody tr:target td {{ background: #3b82f61f; }}
  td.title {{ white-space: normal; min-width: 15rem; line-height: 1.45; }}
  td.deps {{ white-space: normal; min-width: 15rem; max-width: 19rem; line-height: 1.9; }}
  td.owner {{ max-width: 11rem; overflow: hidden; text-overflow: ellipsis; }}
  td.owner a, td.pr a {{ color: inherit; text-decoration: none; }}
  .icon {{ width: 1.35rem; height: 1.35rem; vertical-align: -.4em; fill: currentColor; opacity: .6; }}
  a:hover .icon {{ opacity: 1; }}
  td.deps code, .waits code {{ font-size: .78rem; padding: .1rem .4rem; border-radius: 4px; background: #8881; opacity: .55; }}
  td.deps code.unmet, .waits code.unmet {{ opacity: 1; background: #f9731633; font-weight: 600; }}
  td.deps a, .waits a {{ color: inherit; text-decoration: none; }}
  td.deps a:hover code, .waits a:hover code {{ opacity: 1; outline: 1px solid var(--line); }}
  .tree, .tree ul {{ list-style: none; margin: 0; padding: 0; font-size: .9rem; }}
  .tree ul {{ margin-left: 1.2rem; padding-left: 1.2rem; border-left: 1px solid var(--line); }}
  .tree li {{ position: relative; }}
  .tree ul > li::before {{ content: ""; position: absolute; left: -1.2rem; top: 1.15rem; width: 1rem; border-top: 1px solid var(--line); }}
  .node {{ display: flex; align-items: center; gap: .55rem; flex-wrap: wrap; padding: .3rem 0; scroll-margin-top: 1.5rem; }}
  .node:hover {{ background: #8881; }}
  .node:target {{ background: #3b82f61f; }}
  .node .name {{ opacity: .85; }}
  .node .waits {{ font-size: .78rem; opacity: .7; }}
  td.actions {{ text-align: right; }}
  .tag {{ font-size: .75rem; padding: .15rem .6rem; border: 1px solid var(--line); border-radius: 999px; }}
  .tag.completed {{ color: #16a34a; border-color: #16a34a66; background: #16a34a1a; font-weight: 600; }}
  .local {{ font-weight: 600; }}
  .note {{ opacity: .7; font-size: .8rem; margin-top: 1.75rem; max-width: 52rem; }}
  form {{ display: inline; }}
  button {{ font: inherit; font-size: .8rem; padding: .3rem .8rem; margin-left: .4rem; cursor: pointer; border-radius: 6px; }}
</style>
<header>
  <h1>Code Winch scheduler</h1>
  <form method="post"><input type="hidden" name="action" value="run">
    <button type="submit">Run scheduler</button></form>
  <button type="button" id="swap" hidden>Tree view</button>
</header>
<div class="sub">{summary}</div>
{message}
<div class="wrap" id="table-view">
<table>
  <thead><tr><th>Task</th><th>Title</th><th>Depends on</th><th>Status</th><th>Source</th><th>Owner</th><th>Pull request</th><th>Updated</th><th></th></tr></thead>
  <tbody>
  {rows}
  </tbody>
</table>
</div>
<ul class="tree wrap" id="tree-view" hidden>
  {tree}
</ul>
<p class="note">Leases live in the scheduler state volume and apply only while the tracker has not
recorded the task as completed. Expiring one releases the concurrency slot it holds, so only a lease
that still holds one offers the option: a lease reading completed records a merged pull request, and
once the tracker agrees it is retired in place, keeping its links on the row without overriding
anything. The tree view nests each task under the dependency that unlocks it last, so a task waiting
on several others appears once, under the deepest of them, and lists the rest beside its title.</p>
<script>
  const swap = document.getElementById("swap");
  const table = document.getElementById("table-view");
  const tree = document.getElementById("tree-view");
  const show = (view) => {{
    table.hidden = view === "tree";
    tree.hidden = view !== "tree";
    swap.textContent = view === "tree" ? "Table view" : "Tree view";
    localStorage.setItem("code-winch-view", view);
  }};
  show(localStorage.getItem("code-winch-view") === "tree" ? "tree" : "table");
  swap.hidden = false;
  swap.onclick = () => show(tree.hidden ? "tree" : "table");
</script>
"""

NODE = """<li>
  <div class="node" id="tree-{id}">
    <code>{id}</code><span class="name">{title}</span>
    <span class="tag {status}">{status}</span>{waits}{actions}
  </div>
  {children}
</li>"""

ROW = """<tr id="{id}">
  <td><code>{id}</code></td>
  <td class="title">{title}</td>
  <td class="deps">{deps}</td>
  <td><span class="tag {status}">{status}</span></td>
  <td>{source}</td>
  <td class="owner" title="{hint}">{owner}</td>
  <td class="pr">{pull_request}</td>
  <td>{updated}</td>
  <td class="actions">{actions}</td>
</tr>"""


# The OpenAI mark, inlined so the panel keeps working without any outbound request.
CODEX_ICON = (
    '<svg class="icon" viewBox="0 0 24 24" role="img" aria-label="Codex conversation">'
    '<path d="M22.282 9.821a5.985 5.985 0 0 0-.516-4.911 6.046 6.046 0 0 0-6.51-2.9A6.065 6.065 0'
    " 0 0 4.981 4.182a5.985 5.985 0 0 0-3.998 2.9 6.046 6.046 0 0 0 .743 7.097 5.98 5.98 0 0 0"
    " .51 4.911 6.051 6.051 0 0 0 6.515 2.9A5.985 5.985 0 0 0 13.26 24a6.056 6.056 0 0 0"
    " 5.772-4.206 5.99 5.99 0 0 0 3.998-2.9 6.056 6.056 0 0 0-.748-7.073zm-9.022 12.608a4.476"
    " 4.476 0 0 1-2.876-1.04l.142-.08 4.778-2.759a.795.795 0 0 0 .393-.681v-6.737l2.02"
    " 1.169a.071.071 0 0 1 .038.052v5.583a4.504 4.504 0 0 1-4.495 4.494zm-9.66-4.126a4.471 4.471"
    " 0 0 1-.535-3.014l.142.086 4.783 2.758a.771.771 0 0 0 .78 0l5.843-3.368v2.332a.08.08 0 0"
    " 1-.033.062L9.74 19.95a4.499 4.499 0 0 1-6.14-1.647zM2.34 7.896a4.485 4.485 0 0 1"
    " 2.366-1.973v5.677a.766.766 0 0 0 .388.677l5.815 3.354-2.02 1.169a.076.076 0 0"
    " 1-.071 0l-4.83-2.787A4.504 4.504 0 0 1 2.34 7.872zm16.597 3.856L13.104 8.364 15.119"
    " 7.2a.076.076 0 0 1 .071 0l4.83 2.791a4.494 4.494 0 0 1-.676 8.105v-5.678a.79.79 0 0"
    " 0-.407-.666zm2.01-3.023l-.141-.085-4.774-2.782a.776.776 0 0 0-.785 0L9.41"
    " 9.23V6.897a.066.066 0 0 1 .028-.061l4.83-2.787a4.499 4.499 0 0 1 6.68 4.66zM8.307"
    " 12.863l-2.02-1.164a.08.08 0 0 1-.038-.057V6.074a4.499 4.499 0 0 1 7.376-3.454l-.142"
    " .08-4.778 2.76a.795.795 0 0 0-.393.68zm1.097-2.366l2.602-1.5 2.607 1.5v2.999l-2.597"
    ' 1.5-2.607-1.5z"/></svg>'
)

# The GitHub mark, inlined for the same reason.
GITHUB_ICON = (
    '<svg class="icon" viewBox="0 0 24 24" role="img" aria-label="Pull request">'
    '<path d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577'
    " 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633"
    " 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305"
    " 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93"
    " 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267"
    " 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24"
    " 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81"
    " 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0"
    '-6.627-5.373-12-12-12"/></svg>'
)


def button(action: str, task_id: str, label: str) -> str:
    """Render a form that posts to the page's own URL, naming the action in a field.

    Omitting the action attribute keeps every link on the page relative, so the panel works
    unchanged whether it is reached directly or through a reverse proxy that mounts it under a
    path prefix. An absolute /expire would leave that prefix behind and miss the proxy's route.
    """
    return (
        f'<form method="post">'
        f'<input type="hidden" name="action" value="{html.escape(action)}">'
        f'<input type="hidden" name="task_id" value="{html.escape(task_id)}">'
        f'<button type="submit">{label}</button></form>'
    )


def actions_for(entry: dict[str, Any], runnable: set[str]) -> str:
    actions = ""
    if entry["id"] in runnable:
        actions += button("run", entry["id"], "Run")
    if entry["local"] and entry["status"] != "completed":
        actions += button("expire", entry["id"], "Expire")
    return actions


def source_of(entry: dict[str, Any]) -> str:
    return '<span class="local">lease</span>' if entry["local"] else "tracker"


def link_target(value: Any) -> str:
    """Return the value as a link target, or an empty string when it is not one.

    The scheduler stores the raw dispatch output under task_url when no URL matched, so the value
    only becomes a link once it parses as http(s) — which also keeps a javascript: href out of the
    page if the state file is ever hand-edited.
    """
    url = str(value or "").strip()
    parsed = urlparse(url)
    return url if parsed.scheme in ("http", "https") and parsed.netloc else ""


def owner_cell(entry: dict[str, Any]) -> tuple[str, str]:
    """Render the owner cell and the tooltip that carries whatever the cell leaves out."""
    owner = entry["owner"] or "—"
    url = link_target(entry["task_url"])
    if not url:
        return html.escape(owner), owner
    link = (
        f'<a href="{html.escape(url)}" target="_blank" rel="noopener" '
        f'title="open the Codex conversation">{CODEX_ICON}</a>'
    )
    return link, f"{owner} — {url}" if entry["owner"] else url


def pull_request_cell(entry: dict[str, Any]) -> str:
    """Link the merged pull request the scheduler recorded, naming it in the tooltip."""
    url = link_target(entry["pull_request"])
    if not url:
        return "—"
    number = urlparse(url).path.rstrip("/").rpartition("/")[2]
    hint = f"open pull request #{number}" if number.isdigit() else "open the pull request"
    return (
        f'<a href="{html.escape(url)}" target="_blank" rel="noopener" '
        f'title="{hint}">{GITHUB_ICON}</a>'
    )


def chips(dependencies: list[str], done: set[str], anchor: str) -> str:
    return " ".join(
        f'<a href="#{anchor}{html.escape(dependency)}" title="jump to {html.escape(dependency)}">'
        f'<code class="{"" if dependency in done else "unmet"}">{html.escape(dependency)}</code>'
        "</a>"
        for dependency in dependencies
    )


def branches(nodes: list[dict[str, Any]], runnable: set[str], done: set[str]) -> str:
    items = []
    for node in nodes:
        entry = node["entry"]
        waits = chips(node["waits_on"], done, "tree-")
        items.append(
            NODE.format(
                id=html.escape(entry["id"]),
                title=html.escape(entry["title"]),
                status=html.escape(entry["status"]),
                waits=f'<span class="waits">also waits on {waits}</span>' if waits else "",
                actions=actions_for(entry, runnable),
                children=(
                    f'<ul>{branches(node["children"], runnable, done)}</ul>'
                    if node["children"]
                    else ""
                ),
            )
        )
    return "\n".join(items)


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
        deps = chips(entry["depends_on"], done, "")
        owner, hint = owner_cell(entry)
        body.append(
            ROW.format(
                id=html.escape(entry["id"]),
                title=html.escape(entry["title"]),
                deps=deps or "—",
                status=html.escape(entry["status"]),
                source=source_of(entry),
                owner=owner,
                hint=html.escape(hint),
                pull_request=pull_request_cell(entry),
                updated=html.escape((entry["updated_at"] or "—")[:16].replace("T", " ")),
                actions=actions_for(entry, runnable),
            )
        )

    banner = f'<div class="msg">{html.escape(message)}</div>' if message else ""
    return PAGE.format(
        summary=html.escape(summary),
        message=banner,
        rows="\n  ".join(body),
        tree=branches(forest(entries), runnable, done),
    )


class Handler(BaseHTTPRequestHandler):
    tracker_path: Path
    state_file: Path
    repo_root: Path
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
        """Redirect back to the page itself, carrying the message in the query.

        The location is a bare query so it resolves against whatever URL the browser is on,
        which keeps any reverse proxy path prefix that an absolute /? would drop.
        """
        self.send_response(HTTPStatus.SEE_OTHER)
        self.send_header("Location", "?" + urlencode({"msg": message}))
        self.end_headers()

    def do_GET(self) -> None:
        path, _, query = self.path.partition("?")
        if path != "/":
            self._send(HTTPStatus.NOT_FOUND, b"not found")
            return
        message = parse_qs(query).get("msg", [""])[0]
        try:
            tracker: dict[str, Any] = {"tasks": []}
            if self.tracker_path.exists():
                tracker = load_tracker(self.tracker_path)
            elif not message:
                message = f"no tracker snapshot at {self.tracker_path}; {SNAPSHOT_HINT}"
            state, busy = refresh(self.state_file, tracker)
            page = render(tracker, state, message, busy)
        except (OSError, ValueError) as error:
            self._send(HTTPStatus.INTERNAL_SERVER_ERROR, str(error).encode())
            return
        self._send(HTTPStatus.OK, page.encode())

    def do_POST(self) -> None:
        path, _, _ = self.path.partition("?")
        if path != "/":
            self._send(HTTPStatus.NOT_FOUND, b"not found")
            return
        length = int(self.headers.get("Content-Length") or 0)
        form = parse_qs(self.rfile.read(length).decode())
        action = (form.get("action") or [""])[0]
        task_id = (form.get("task_id") or [""])[0].upper()

        if action == "run":
            self._redirect(
                run_scheduler(
                    self.repo_root,
                    self.tracker_path,
                    self.state_file,
                    self.environment,
                    task_id or None,
                )
            )
            return
        if action == "expire":
            self._redirect(expire(self.state_file, task_id) if task_id else "no task selected")
            return
        self._send(HTTPStatus.NOT_FOUND, b"not found")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--tracker",
        type=Path,
        help="tracker snapshot; defaults to tracker.json beside the state file",
    )
    parser.add_argument("--state-file", type=Path, required=True)
    parser.add_argument(
        "--repo-root",
        type=Path,
        default=Path("/opt/code-winch"),
        help="directory holding the scripts/ the panel runs",
    )
    parser.add_argument("--host", default="127.0.0.1")
    parser.add_argument("--port", type=int, default=8765)
    parser.add_argument("--expire", metavar="TASK_ID")
    return parser.parse_args()


def main() -> int:
    logging.basicConfig(level=logging.INFO, format="%(levelname)s: %(message)s")
    args = parse_args()
    if args.expire:
        print(expire(args.state_file, args.expire.upper()))
        return 0

    Handler.tracker_path = args.tracker or args.state_file.parent / "tracker.json"
    Handler.state_file = args.state_file
    Handler.repo_root = args.repo_root
    Handler.environment = os.environ.get("CODEX_ENV_ID", "")

    server = ThreadingHTTPServer((args.host, args.port), Handler)
    print(f"control panel on http://{args.host}:{args.port}")
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        pass
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
