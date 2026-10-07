import unittest

from control_panel.integrations import codex


class CodexTests(unittest.TestCase):
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
