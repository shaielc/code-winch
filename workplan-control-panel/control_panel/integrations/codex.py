"""Submit a prompt to Codex Cloud and extract its task URL."""

import re
from pathlib import Path

from .process import run

TASK_URL = re.compile(r"(https?://\S*?)/(?:remote|codex/(?:cloud/)?tasks)/(task_[\w-]+)")


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
    output = run("codex", "cloud", "exec", "--env", environment,
                 "--branch", branch, prompt, cwd=clone)
    url = task_url_from(output)
    if not url:
        raise ValueError("Codex returned no task URL")
    return url
