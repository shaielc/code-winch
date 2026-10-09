"""Run bounded CLI operations without a shell."""

import os
import subprocess
from pathlib import Path

from .. import logs

TIMEOUT = 300


def run(*command: str, cwd: Path, capture: bool = True,
        extra_env: dict[str, str] | None = None) -> str:
    """Run the command, logging why it failed before the caller decides what to say.

    The caller turns the failure into a message for the operator, which deliberately
    names no command or argument; the log is where the detail behind that message goes.

    `extra_env` adds to this process's environment for this command alone, which is how a
    credential reaches the one command that needs it instead of everything the panel runs.
    """
    try:
        result = subprocess.run(
            command,
            cwd=cwd,
            check=True,
            text=True,
            stdout=subprocess.PIPE if capture else None,
            stderr=subprocess.PIPE if capture else None,
            timeout=TIMEOUT,
            env={**os.environ, **extra_env} if extra_env else None,
        )
    except subprocess.CalledProcessError as error:
        logs.failure(f"{logs.command(command)} exited {error.returncode}",
                     error.stderr or error.stdout)
        raise
    except subprocess.TimeoutExpired as error:
        logs.failure(f"{logs.command(command)} timed out after {TIMEOUT}s",
                     error.stderr or error.stdout)
        raise
    return result.stdout.strip() if capture else ""


def failure_message(error: Exception) -> str:
    """Name the failed operation without exposing command arguments or credentials."""
    command = getattr(error, "cmd", [])
    operation = "Repository operation"
    if isinstance(command, (list, tuple)) and len(command) >= 2:
        if command[0] == "git" and command[1] in ("pull", "fetch", "push", "commit", "worktree",
                                                  "show", "ls-remote", "clone"):
            operation = "git " + command[1]
    if isinstance(error, subprocess.TimeoutExpired):
        return f"{operation} timed out. Check remote connectivity and retry Sync main."
    detail = str(getattr(error, "stderr", "") or "").lower()
    if "identity unknown" in detail or "unable to auto-detect email" in detail:
        return f"{operation} failed: configure GIT_AUTHOR_NAME, GIT_AUTHOR_EMAIL and committer identity on the panel."
    if any(text in detail for text in ("authentication failed", "permission denied", "403", "could not read username")):
        return (f"{operation} failed: check the control panel's GitHub App installation and "
                "repository permissions.")
    if "non-fast-forward" in detail or "not possible to fast-forward" in detail:
        return f"{operation} failed: the branch has diverged; reconcile the checkout before retrying."
    return f"{operation} failed. Check checkout permissions, Git credentials and remote availability, then retry Sync main."
