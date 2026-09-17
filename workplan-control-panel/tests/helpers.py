"""Local Git fixtures shared by repository and scheduler tests."""

import json
import os
import tempfile
from pathlib import Path
from unittest.mock import patch

from control_panel.integrations import repository
from control_panel.integrations.process import run

TASK = {"id": "P0-001", "title": "Upstream"}
TRACKER = {
    "schema_version": 1,
    "tasks": [{**TASK, "status": "pending", "owner": None, "blocked_reason": None}],
}


def git(cwd: Path, *arguments: str) -> str:
    return run("git", *arguments, cwd=cwd)


def pushes(spy) -> list[tuple[str, ...]]:
    return [call.args for call in spy.call_args_list if call.args[:2] == ("git", "push")]


class GitRepositoryFixture:
    def setUp(self):
        scratch = tempfile.TemporaryDirectory()
        self.addCleanup(scratch.cleanup)
        self.root = Path(scratch.name)
        isolated = patch.dict(
            os.environ, {"GIT_CONFIG_GLOBAL": os.devnull, "GIT_CONFIG_NOSYSTEM": "1"}
        )
        isolated.start()
        self.addCleanup(isolated.stop)

        self.origin = self.root / "origin.git"
        git(self.root, "init", "--quiet", "--bare", "--initial-branch=main", str(self.origin))
        seed = self.clone("seed")
        (seed / repository.TRACKER).parent.mkdir(parents=True)
        # Not save_tracker's layout, so only a tracker written by save_tracker matches it.
        (seed / repository.TRACKER).write_text(json.dumps(TRACKER, indent=4))
        git(seed, "add", ".")
        git(seed, "commit", "--quiet", "-m", "seed")
        git(seed, "push", "--quiet", "origin", "HEAD:main")

    def clone(self, name: str) -> Path:
        path = self.root / name
        git(self.root, "clone", "--quiet", str(self.origin), str(path))
        git(path, "config", "user.name", "Scheduler Test")
        git(path, "config", "user.email", "scheduler@example.com")
        return path

    def commits_ahead(self) -> str:
        return git(self.origin, "rev-list", "--count", "main..task/P0-001")
