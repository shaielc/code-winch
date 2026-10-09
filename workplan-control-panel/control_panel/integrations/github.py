"""The one place a GitHub credential enters a command's environment.

Everything the panel runs against GitHub — `git clone`, `fetch`, `pull` and `push`, and the
`gh` pull request calls — goes through github_run, which puts a current token in GH_TOKEN
for that command alone. The panel's own environment carries no GitHub credential, so
nothing else it drives can read one: `codex cloud exec` runs agent-chosen work on the
panel's behalf and has no business holding the credential that writes to the repository.

Which credential that is belongs to the composition root, which installs one with use()
before serving. The App is the credential; a GH_TOKEN handed to the panel directly is a
temporary fallback kept only so a deployment can roll back without reverting code.
"""

import os
from pathlib import Path
from typing import Protocol

from .. import logs
from . import github_app
from .process import run


class Credentials(Protocol):
    """A source of a currently valid GitHub token."""

    def token(self) -> str:
        ...


class StaticTokenAuth:
    """A token handed to the panel whole, which it holds until it is restarted.

    Temporary: it exists so a deployment can fall back to a personal access token during
    the move to App authentication, and goes away with that fallback.
    """

    def __init__(self, token: str):
        self._token = token

    def token(self) -> str:
        return self._token


_credentials: Credentials | None = None


def use(credentials: Credentials) -> None:
    """Install the credential every GitHub command will authenticate with."""
    global _credentials
    _credentials = credentials


def credentials() -> Credentials:
    if _credentials is None:
        raise github_app.ConfigurationError(
            "The control panel has no GitHub credentials; set GITHUB_APP_ID and "
            "GITHUB_APP_PRIVATE_KEY_FILE")
    return _credentials


def from_environment() -> Credentials:
    """Read the deployment's choice of credential, preferring the App to a static token."""
    # Out of the environment whether it is used or not: a GitHub credential belongs only to
    # the commands github_run gives it to. Naming it to logs keeps redaction's reach over a
    # value redact() can no longer read back out of os.environ.
    token = os.environ.pop("GH_TOKEN", "").strip()
    logs.guard(token)
    app_id = os.environ.get("GITHUB_APP_ID", "").strip()
    private_key = os.environ.get("GITHUB_APP_PRIVATE_KEY_FILE", "").strip()
    if app_id:
        if not private_key:
            raise github_app.ConfigurationError(
                "Set GITHUB_APP_PRIVATE_KEY_FILE to the mounted GitHub App private key")
        return github_app.GitHubAppAuth(app_id, Path(private_key),
                                        os.environ.get("GITHUB_URL", ""))
    if token:
        # Temporary, with the fallback the migration keeps for rollback.
        logs.output("authenticating to GitHub with GH_TOKEN; set GITHUB_APP_ID and "
                    "GITHUB_APP_PRIVATE_KEY_FILE to authenticate as the GitHub App instead")
        return StaticTokenAuth(token)
    raise github_app.ConfigurationError(
        "Set GITHUB_APP_ID and GITHUB_APP_PRIVATE_KEY_FILE so the control panel can "
        "authenticate as its GitHub App")


def github_run(*command: str, cwd: Path, capture: bool = True) -> str:
    """Run one GitHub command with a current token in its environment and nowhere else."""
    return run(*command, cwd=cwd, capture=capture, extra_env={
        "GH_TOKEN": credentials().token(),
        # Without a credential Git would sit waiting for a username it cannot be given.
        "GIT_TERMINAL_PROMPT": "0",
    })


def setup_git(url: str, cwd: Path) -> None:
    """Point Git's credential helper at gh, which answers it from the token we pass.

    Git itself does nothing with GH_TOKEN; the helper is what turns it into a credential for
    the host. It is written to this container's Git configuration, not to the checkout, so
    it is set up once per start and holds for every later command.
    """
    github_run("gh", "auth", "setup-git", "--hostname", github_app.host(url), cwd=cwd)
