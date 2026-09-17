import unittest

from control_panel import prompts


class PromptTests(unittest.TestCase):
    def test_bundled_stage_prompts_render_all_fields(self):
        task = {"id": "P0-001", "title": "Upstream", "brief": "phase-0/P0-001-upstream.md"}
        for stage in ("refine", "implement", "audit"):
            fields = dict(task)
            if stage == "audit":
                fields.update(pr_url="https://example/pr/1", head="0123abcd")
            with self.subTest(stage=stage):
                prompt = prompts.load(stage).substitute(fields)
                for value in fields.values():
                    self.assertIn(value, prompt)
                self.assertNotIn("$", prompt)
