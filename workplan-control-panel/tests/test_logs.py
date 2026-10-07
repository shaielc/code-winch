import subprocess
import unittest
from pathlib import Path
from unittest.mock import patch

from control_panel import logs
from control_panel.integrations import codex
from control_panel.integrations.process import run


class RedactionTests(unittest.TestCase):
    def test_the_panels_own_secrets_never_reach_the_log(self):
        with patch.dict("os.environ", {"GH_TOKEN": "tokenvalue123", "PANEL_TOKEN": "panelsecret"}):
            line = logs.redact("push failed for tokenvalue123 and panelsecret")
        self.assertNotIn("tokenvalue123", line)
        self.assertNotIn("panelsecret", line)
        self.assertIn("push failed for", line)

    def test_a_short_secret_is_left_alone_rather_than_redacting_the_line(self):
        # An empty or one-character token would otherwise match everywhere.
        with patch.dict("os.environ", {"GH_TOKEN": "", "PANEL_TOKEN": "x"}):
            self.assertEqual(logs.redact("exit status 1"), "exit status 1")

    def test_credentials_the_panel_never_held_are_redacted_by_shape(self):
        cases = {
            "fatal: could not read Username for https://SECRET@github.com": "SECRET",
            "remote: bad credentials ghp_abcdefgh12345678": "ghp_abcdefgh12345678",
            "https://user:hunter2@example.com/repo.git": "hunter2",
            "x-api-key: sk-ant-api03-abcdefgh": "sk-ant-api03-abcdefgh",
        }
        for text, secret in cases.items():
            self.assertNotIn(secret, logs.redact(text))
        # A URL without userinfo is not a credential and survives intact.
        plain = "https://chatgpt.com/remote/task_e_123"
        self.assertEqual(logs.redact(plain), plain)

    def test_a_token_supplied_as_an_argument_can_be_named(self):
        self.assertNotIn("sk-secret", logs.redact('{"error": "sk-secret is invalid"}', "sk-secret"))

    def test_a_prompt_sized_argument_is_summarized_not_pasted(self):
        prompt = "Implement P0-001. " * 40
        named = logs.command(("codex", "cloud", "exec", "--branch", "task/P0-001", prompt))
        self.assertIn("codex cloud exec --branch task/P0-001", named)
        self.assertNotIn("Implement P0-001", named)
        self.assertIn(f"<{len(prompt)} chars>", named)

    def test_long_output_keeps_its_tail(self):
        clipped = logs.tail("start" + "x" * (logs.TAIL * 2) + "the actual error")
        self.assertTrue(clipped.endswith("the actual error"))
        self.assertNotIn("start", clipped)
        self.assertLess(len(clipped), logs.TAIL + 40)


class SubprocessLogTests(unittest.TestCase):
    def test_a_failed_command_logs_its_stderr_and_still_raises(self):
        with self.assertLogs("control_panel", "ERROR") as captured:
            with self.assertRaises(subprocess.CalledProcessError):
                run("git", "rev-parse", "--verify", "refs/heads/nope", cwd=Path.cwd())
        self.assertIn("git rev-parse --verify refs/heads/nope", captured.output[0])
        self.assertIn("exited", captured.output[0])

    def test_a_timeout_names_the_limit(self):
        with patch("control_panel.integrations.process.TIMEOUT", 1):
            with self.assertLogs("control_panel", "ERROR") as captured:
                with self.assertRaises(subprocess.TimeoutExpired):
                    run("sleep", "5", cwd=Path.cwd())
        self.assertIn("timed out after 1s", captured.output[0])

    def test_a_successful_command_logs_nothing(self):
        with patch.object(logs.LOGGER, "handle") as handled:
            run("git", "--version", cwd=Path.cwd())
        handled.assert_not_called()


class CodexLogTests(unittest.TestCase):
    def test_dispatch_output_is_logged_whether_or_not_it_parses(self):
        with patch("control_panel.integrations.codex.run", return_value="Not signed in."):
            with self.assertLogs("control_panel", "INFO") as captured:
                with self.assertRaises(ValueError):
                    codex.submit(Path.cwd(), "env", "task/P0-001", "prompt")
        self.assertIn("codex cloud exec on task/P0-001", captured.output[0])
        self.assertIn("Not signed in.", captured.output[0])

    def test_a_dispatched_task_logs_its_url(self):
        url = "https://chatgpt.com/codex/tasks/task_e_123"
        with patch("control_panel.integrations.codex.run", return_value=url):
            with self.assertLogs("control_panel", "INFO") as captured:
                codex.submit(Path.cwd(), "env", "task/P0-001", "prompt")
        self.assertIn(url, captured.output[0])
