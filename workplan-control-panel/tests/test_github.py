"""Which credential the panel uses, and which commands are allowed to see it."""

import os
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

from control_panel import logs
from control_panel.integrations import codex, github, github_app
from control_panel.integrations.process import run

TOKEN = "ghs_installationtoken0001"
URL = "https://github.com/shaielc/code-winch"
# Prints what the command it runs can read out of its own environment.
PROBE = ("python3", "-c", "import os; print(os.environ.get('GH_TOKEN', 'absent'))")


class CredentialChoiceTests(unittest.TestCase):
    def setUp(self):
        self.environment = patch.dict(os.environ)
        self.environment.start()
        self.addCleanup(self.environment.stop)
        for name in ("GH_TOKEN", "GITHUB_APP_ID", "GITHUB_APP_PRIVATE_KEY_FILE", "GITHUB_URL"):
            os.environ.pop(name, None)
        scratch = tempfile.TemporaryDirectory()
        self.addCleanup(scratch.cleanup)
        self.private_key = Path(scratch.name) / "app.pem"
        self.private_key.write_text("-----BEGIN PRIVATE KEY-----\nstub\n-----END PRIVATE KEY-----\n")

    def test_the_app_is_preferred_and_configured_from_the_repository_url(self):
        os.environ.update(GITHUB_APP_ID="123456", GITHUB_URL=URL, GH_TOKEN="ghp_oldtoken12345678",
                          GITHUB_APP_PRIVATE_KEY_FILE=str(self.private_key))
        credentials = github.from_environment()
        self.assertIsInstance(credentials, github_app.GitHubAppAuth)
        self.assertEqual((credentials.owner, credentials.repository), ("shaielc", "code-winch"))

    def test_a_static_token_is_the_fallback_when_no_app_is_configured(self):
        os.environ.update(GITHUB_URL=URL, GH_TOKEN="ghp_oldtoken12345678")
        credentials = github.from_environment()
        self.assertIsInstance(credentials, github.StaticTokenAuth)
        self.assertEqual(credentials.token(), "ghp_oldtoken12345678")

    def test_the_token_is_taken_out_of_the_panels_environment_either_way(self):
        # Whatever is chosen, nothing the panel runs can read a credential it was not given.
        for extra in ({}, {"GITHUB_APP_ID": "123456",
                           "GITHUB_APP_PRIVATE_KEY_FILE": str(self.private_key)}):
            os.environ.update(GITHUB_URL=URL, GH_TOKEN="ghp_oldtoken12345678", **extra)
            github.from_environment()
            self.assertNotIn("GH_TOKEN", os.environ)
            self.assertEqual(run(*PROBE, cwd=Path.cwd()), "absent")
            # Out of the environment is out of redact()'s reach, so it is named to logs.
            self.assertNotIn("ghp_oldtoken12345678", logs.redact("push failed for ghp_oldtoken12345678"))

    def test_an_unconfigured_panel_says_what_to_set(self):
        os.environ.update(GITHUB_URL=URL)
        for environment, named in (
                ({}, "GITHUB_APP_ID"),
                ({"GITHUB_APP_ID": "123456"}, "GITHUB_APP_PRIVATE_KEY_FILE"),
                ({"GITHUB_APP_ID": "123456", "GITHUB_APP_PRIVATE_KEY_FILE": "/absent.pem"},
                 "GITHUB_APP_PRIVATE_KEY_FILE")):
            os.environ.pop("GITHUB_APP_ID", None)
            os.environ.pop("GITHUB_APP_PRIVATE_KEY_FILE", None)
            os.environ.update(environment)
            with self.assertRaises(github_app.ConfigurationError) as error:
                github.from_environment()
            self.assertIn(named, str(error.exception))


class CredentialBoundaryTests(unittest.TestCase):
    def setUp(self):
        # The panel's own environment carries no GitHub credential; from_environment is what
        # guarantees that at startup, and these tests stand where it has already run.
        self.environment = patch.dict(os.environ)
        self.environment.start()
        self.addCleanup(self.environment.stop)
        os.environ.pop("GH_TOKEN", None)
        credentials = patch.object(github, "_credentials", github.StaticTokenAuth(TOKEN))
        credentials.start()
        self.addCleanup(credentials.stop)

    def test_a_github_command_is_given_the_token_and_a_plain_one_is_not(self):
        self.assertEqual(github.github_run(*PROBE, cwd=Path.cwd()), TOKEN)
        self.assertEqual(run(*PROBE, cwd=Path.cwd()), "absent")

    def test_a_github_command_cannot_sit_waiting_for_a_credential(self):
        prompting = ("python3", "-c", "import os; print(os.environ['GIT_TERMINAL_PROMPT'])")
        self.assertEqual(github.github_run(*prompting, cwd=Path.cwd()), "0")

    def test_codex_dispatch_carries_no_github_credential(self):
        """Codex runs agent-chosen work, so it cannot hold what writes to the repository."""
        seen = []

        def dispatch(*command, cwd, capture=True, **passed):
            seen.append(run(*PROBE, cwd=cwd, capture=capture, **passed))
            return "https://chatgpt.com/remote/task_e_1"

        with patch.object(codex, "run", side_effect=dispatch):
            codex.submit(Path.cwd(), "env", "task/P0-001", "Implement P0-001")
        self.assertEqual(seen, ["absent"])

    def test_git_credentials_are_configured_for_the_repositorys_host(self):
        with patch.object(github, "run") as runner:
            github.setup_git(URL, Path.cwd())
        self.assertEqual(runner.call_args.args,
                         ("gh", "auth", "setup-git", "--hostname", "github.com"))
        self.assertEqual(runner.call_args.kwargs["extra_env"]["GH_TOKEN"], TOKEN)

    def test_a_panel_without_credentials_runs_no_github_command(self):
        with patch.object(github, "_credentials", None):
            with self.assertRaises(github_app.ConfigurationError) as error:
                github.github_run("gh", "pr", "list", cwd=Path.cwd())
        self.assertIn("GITHUB_APP_ID", str(error.exception))
