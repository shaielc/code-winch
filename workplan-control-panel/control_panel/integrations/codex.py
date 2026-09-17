"""Submit a prompt to Codex Cloud and extract its task URL."""

import re
from pathlib import Path

from .process import run

TASK_URL = re.compile(r"https?://\S+/codex/(?:cloud/)?tasks/\S+")


def task_url_from(output: str) -> str | None:
    match = TASK_URL.search(output)
    return match.group(0) if match else None


def submit(clone: Path, environment: str, branch: str, prompt: str) -> str:
    output = run("codex", "cloud", "exec", "--env", environment,
                 "--branch", branch, prompt, cwd=clone)
    url = task_url_from(output)
    if not url:
        raise ValueError("Codex returned no task URL")
    return url
