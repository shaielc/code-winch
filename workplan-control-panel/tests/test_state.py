import json
import tempfile
import unittest
from pathlib import Path

from control_panel import state as task_state


class SchemaTests(unittest.TestCase):
    def setUp(self):
        scratch = tempfile.TemporaryDirectory()
        self.addCleanup(scratch.cleanup)
        self.path = Path(scratch.name) / "task-state.json"

    def write(self, payload):
        self.path.write_text(json.dumps(payload))
        return task_state.load_state(self.path)

    def test_version_one_stage_keys_become_one_ordered_list_per_stage(self):
        state = self.write({"schema_version": 1, "tasks": {"P0-001": {"status": "in_progress", "stages": {
            "refine:aaa": {"status": "submitted", "stage": "refine", "head": "aaa",
                           "task_url": "https://codex/one", "updated_at": "2026-09-01T00:00:00+00:00"},
            "implement:bbb": {"status": "submitted", "stage": "implement", "head": "bbb",
                              "task_url": "https://codex/three", "updated_at": "2026-09-03T00:00:00+00:00"},
            "refine:bbb": {"status": "submitted", "stage": "refine", "head": "bbb",
                           "task_url": "https://codex/two", "updated_at": "2026-09-02T00:00:00+00:00"},
        }}}})
        self.assertEqual(state["schema_version"], task_state.SCHEMA_VERSION)
        stages = state["tasks"]["P0-001"]["stages"]
        self.assertEqual(sorted(stages), ["implement", "refine"])
        # Ordered by when they were submitted, not by the key they used to carry.
        self.assertEqual([entry["head"] for entry in stages["refine"]], ["aaa", "bbb"])
        self.assertEqual([entry["task_url"] for entry in stages["refine"]],
                         ["https://codex/one", "https://codex/two"])
        self.assertNotIn("stage", stages["refine"][0])
        self.assertEqual(task_state.live_attempt(stages["refine"], "bbb")["task_url"],
                         "https://codex/two")
        self.assertIsNone(task_state.live_attempt(stages["refine"], "ccc"))

    def test_migration_leaves_records_without_stages_alone(self):
        state = self.write({"schema_version": 1, "tasks": {"P0-001": {"status": "completed"}}})
        self.assertEqual(state["tasks"]["P0-001"], {"status": "completed"})
        self.assertEqual(state["schema_version"], task_state.SCHEMA_VERSION)

    def test_a_newer_or_malformed_state_is_refused_rather_than_guessed_at(self):
        for payload in ({"schema_version": 99, "tasks": {}}, {"schema_version": 2, "tasks": []}):
            with self.assertRaises(ValueError):
                self.write(payload)
        self.assertEqual(task_state.load_state(self.path.with_name("absent.json")),
                         {"schema_version": task_state.SCHEMA_VERSION, "tasks": {}})

    def test_live_attempt_reads_the_newest_record_at_a_revision(self):
        attempts = [{"status": "submitted", "head": "aaa"}, {"status": "expired", "head": "aaa"}]
        self.assertIsNone(task_state.live_attempt(attempts, "aaa"))
        attempts.append({"status": "submitting", "head": "aaa"})
        self.assertEqual(task_state.live_attempt(attempts, "aaa")["status"], "submitting")


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
