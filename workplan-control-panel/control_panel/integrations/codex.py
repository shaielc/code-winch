"""Submit and inspect Codex Cloud tasks without exposing GitHub credentials."""

import json
import re
from pathlib import Path

from .. import logs
from .process import run

TASK_URL = re.compile(r"(https?://\S*?)/(?:remote|codex/(?:cloud/)?tasks)/(task_[\w-]+)")


def _run(*args: str, clone: Path, strip: bool = True) -> str:
    """Run Codex with GitHub's credential removed from its environment."""
    return run("codex", "cloud", *args, cwd=clone, strip=strip,
               remove_env=("GH_TOKEN",))


def find_task(clone: Path, environment: str, task_id: str) -> dict | None:
    """Find a task across every status-list page."""
    cursor = None
    while True:
        args = ["list", "--json", "--env", environment, "--limit", "20"]
        if cursor:
            args.extend(("--cursor", cursor))
        page = json.loads(_run(*args, clone=clone))
        tasks = page if isinstance(page, list) else page.get("tasks", [])
        found = next((task for task in tasks if task.get("id") == task_id), None)
        if found:
            return found
        cursor = page.get("cursor") if isinstance(page, dict) else None
        if not cursor:
            return None


def diff(clone: Path, task_id: str) -> str:
    """Return the unified diff byte-for-byte as decoded by subprocess."""
    return _run("diff", task_id, clone=clone, strip=False)


def canonical_task_url(url: str) -> str:
    """Point a task URL at the view the conversation opens in today, which is /remote/<task_id>.

    Dispatch has printed three shapes so far — /codex/tasks/, /codex/cloud/tasks/ and now
    /remote/ — and stored records keep whichever was current when they were written. Anything
    that is not a recognisable task URL is returned untouched, since the scheduler also stores
    raw dispatch output under the same key.
    """
    match = TASK_URL.search(url)
    return f"{match[1]}/remote/{match[2]}" if match else url


def task_url_from(output: str) -> str | None:
    match = TASK_URL.search(output)
    return f"{match[1]}/remote/{match[2]}" if match else None


def submit(clone: Path, environment: str, branch: str, prompt: str) -> str:
    output = _run("exec", "--env", environment, "--branch", branch, prompt, clone=clone)
    logs.output(f"codex cloud exec on {branch}", output)
    url = task_url_from(output)
    if not url:
        raise ValueError("Codex returned no task URL")
    return url
