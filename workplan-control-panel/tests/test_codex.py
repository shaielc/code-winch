import unittest

from control_panel.integrations import codex


class CodexTests(unittest.TestCase):
    def test_task_url_survives_noise_on_dispatch_output(self):
        cloud = "https://chatgpt.com/codex/cloud/tasks/task_e_123"
        for path in ("codex/tasks", "codex/cloud/tasks"):
            url = f"https://chatgpt.com/{path}/task_e_123"
            self.assertEqual(codex.task_url_from(f"warning: stale login\n{url}\n"), cloud)
        self.assertIsNone(codex.task_url_from("Not signed in. Please run 'codex login'."))
