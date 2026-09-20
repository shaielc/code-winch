"""Run bounded CLI operations without a shell."""

import subprocess
from pathlib import Path


def run(*command: str, cwd: Path, capture: bool = True) -> str:
    result = subprocess.run(
        command,
        cwd=cwd,
        check=True,
        text=True,
        stdout=subprocess.PIPE if capture else None,
        stderr=subprocess.PIPE if capture else None,
        timeout=300,
    )
    return result.stdout.strip() if capture else ""


def failure_message(error: Exception) -> str:
    """Name the failed operation without exposing command arguments or credentials."""
    command = getattr(error, "cmd", [])
    operation = "Repository operation"
    if isinstance(command, (list, tuple)) and len(command) >= 2:
        if command[0] == "git" and command[1] in ("pull", "fetch", "push", "commit", "worktree", "show", "ls-remote"):
            operation = "git " + command[1]
    if isinstance(error, subprocess.TimeoutExpired):
        return f"{operation} timed out. Check remote connectivity and retry Sync main."
    detail = str(getattr(error, "stderr", "") or "").lower()
    if "identity unknown" in detail or "unable to auto-detect email" in detail:
        return f"{operation} failed: configure GIT_AUTHOR_NAME, GIT_AUTHOR_EMAIL and committer identity on the panel."
    if any(text in detail for text in ("authentication failed", "permission denied", "403", "could not read username")):
        return f"{operation} failed: check the panel's GH_TOKEN and repository write permission."
    if "non-fast-forward" in detail or "not possible to fast-forward" in detail:
        return f"{operation} failed: the branch has diverged; reconcile the checkout before retrying."
    return f"{operation} failed. Check checkout permissions, Git credentials and remote availability, then retry Sync main."
