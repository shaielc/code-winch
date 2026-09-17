import unittest

from control_panel import state as task_state


class StateTests(unittest.TestCase):
    def test_effective_tracker_applies_only_local_lifecycle_fields(self):
        tracker = {
            "tasks": [
                {
                    "id": "P0-001",
                    "title": "Upstream",
                    "status": "pending",
                    "owner": None,
                    "blocked_reason": None,
                }
            ]
        }
        state = {
            "tasks": {
                "P0-001": {
                    "status": "in_progress",
                    "owner": "worker",
                    "blocked_reason": None,
                    "task_url": "ignored",
                }
            }
        }
        effective = task_state.effective_tracker(tracker, state)
        self.assertEqual(effective["tasks"][0]["title"], "Upstream")
        self.assertEqual(effective["tasks"][0]["status"], "in_progress")
        self.assertEqual(effective["tasks"][0]["owner"], "worker")
        self.assertNotIn("task_url", effective["tasks"][0])

    def test_retire_completed_keeps_the_entry_the_tracker_caught_up_with(self):
        tracker = {"tasks": [{"id": "P0-001", "status": "completed"}]}
        state = {
            "schema_version": 1,
            "tasks": {
                "P0-001": {
                    "status": "in_progress",
                    "owner": "worker",
                    "pull_request": "https://example/pr/1",
                }
            },
        }
        self.assertTrue(task_state.retire_completed(tracker, state))
        entry = state["tasks"]["P0-001"]
        self.assertEqual(entry["status"], "completed")
        self.assertIsNone(entry["owner"])
        self.assertEqual(entry["pull_request"], "https://example/pr/1")
        self.assertFalse(task_state.retire_completed(tracker, state))
