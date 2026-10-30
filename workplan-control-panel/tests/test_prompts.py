import unittest

from control_panel import prompts


class PromptTests(unittest.TestCase):
    def test_bundled_stage_prompts_render_all_fields(self):
        task = {"id": "P0-001", "title": "Upstream", "brief": "phase-0/P0-001-upstream.md"}
        for stage in ("refine", "implement", "audit"):
            fields = dict(task)
            if stage == "audit":
                fields.update(pr_url="https://example/pr/1", head="0123abcd")
            else:
                fields.update(branch=f"{stage}/P0-001-1")
            with self.subTest(stage=stage):
                prompt = prompts.load(stage).substitute(fields)
                for value in fields.values():
                    self.assertIn(value, prompt)
                self.assertNotIn("$", prompt)

    def test_codex_prompts_leave_branch_and_pull_request_ownership_to_panel(self):
        task = {"id": "P0-001", "title": "Upstream",
                "brief": "phase-0/P0-001-upstream.md"}
        refine = prompts.load("refine").substitute(task, branch="refine/P0-001-1")
        implement = prompts.load("implement").substitute(task, branch="implement/P0-001-1")
        self.assertIn("commit\nthe refinement", refine)
        for prompt in (refine, implement):
            self.assertIn("do not create a branch", prompt.lower())
            self.assertIn("open a pull request", prompt.lower())
