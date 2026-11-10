import os
import unittest
from pathlib import Path
from unittest.mock import patch

from control_panel.integrations import codex


class CodexTests(unittest.TestCase):
    @patch("control_panel.integrations.codex.run")
    def test_find_task_follows_pages_and_codex_never_receives_gh_token(self, runner):
        runner.side_effect = [
            '{"tasks":[{"id":"other"}],"cursor":"next"}',
            '{"tasks":[{"id":"wanted","status":"ready"}]}',
        ]
        with patch.dict(os.environ, {"GH_TOKEN": "secret"}):
            found = codex.find_task(Path("/clone"), "env", "wanted")
        self.assertEqual(found["status"], "ready")
        self.assertEqual(runner.call_count, 2)
        self.assertIn("--cursor", runner.call_args_list[1].args)
        for call in runner.call_args_list:
            self.assertEqual(call.kwargs["remove_env"], ("GH_TOKEN",))

    @patch("control_panel.integrations.codex.run", return_value="diff --git a/x b/x\n ")
    def test_diff_preserves_trailing_content(self, runner):
        self.assertEqual(codex.diff(Path("/clone"), "wanted"),
                         "diff --git a/x b/x\n ")
        self.assertFalse(runner.call_args.kwargs["strip"])

    def test_task_url_survives_noise_on_dispatch_output(self):
        remote = "https://chatgpt.com/remote/task_e_123"
        for path in ("codex/tasks", "codex/cloud/tasks", "remote"):
            url = f"https://chatgpt.com/{path}/task_e_123"
            self.assertEqual(codex.task_url_from(f"warning: stale login\n{url}\n"), remote)
        self.assertIsNone(codex.task_url_from("Not signed in. Please run 'codex login'."))

    def test_canonical_url_rewrites_retired_forms_and_keeps_anything_else(self):
        for path in ("codex/tasks", "codex/cloud/tasks", "remote"):
            self.assertEqual(codex.canonical_task_url(f"https://chatgpt.com/{path}/task_e_123"),
                             "https://chatgpt.com/remote/task_e_123")
        for other in ("", "Not signed in.", "https://chatgpt.com/", "https://example.com/remote/x"):
            self.assertEqual(codex.canonical_task_url(other), other)
