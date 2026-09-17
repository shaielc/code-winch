import json
import sys
import threading
import unittest
from http.client import HTTPConnection
from pathlib import Path
from unittest.mock import patch

import test_task_scheduler as helpers
from test_task_scheduler import git, TRACKER

ROOT = Path(__file__).parents[1]
sys.path.insert(0, str(ROOT / 'workplan-control-panel'))
from service import ControlPanel, Problem, scheduler
from control_panel import Handler, ThreadingHTTPServer, render


class PanelFlowTests(unittest.TestCase):
    clone = helpers.TaskBranchTests.clone
    commits_ahead = helpers.TaskBranchTests.commits_ahead
    def setUp(self):
        helpers.TaskBranchTests.setUp(self)
        seed = self.root / 'seed'
        tracker = json.loads(json.dumps(TRACKER))
        tracker['tasks'][0].update(depends_on=[], brief='phase-0/first.md')
        tracker['tasks'].append({**tracker['tasks'][0], 'id': 'P0-002', 'depends_on': ['P0-001']})
        scheduler.save_tracker(seed / scheduler.TRACKER, tracker)
        git(seed, 'commit', '-am', 'task graph')
        git(seed, 'push', 'origin', 'main')
        self.panel = ControlPanel(ROOT, self.clone('panel'), self.root / 'state.json',
                                  self.root / 'tracker.json', 'test-env', 1)

    def test_merge_pulls_main_prepares_once_and_does_not_launch_codex(self):
        event = {'action': 'closed', 'pull_request': {'merged_at': 'today', 'base': {'ref': 'main'}}}
        self.assertEqual(self.panel.sync(event)['prepared'], ['P0-001'])
        self.assertEqual(self.panel.sync(event)['prepared'], [])
        self.assertEqual(self.commits_ahead(), '1')
        self.assertEqual(git(self.panel.clone, 'branch', '--show-current'), 'main')
        self.assertEqual(git(self.panel.clone, 'status', '--porcelain'), '')
        state = self.panel.snapshot()['state']['tasks']
        self.assertTrue(state['P0-001']['prepared'])
        self.assertNotIn('task_url', state['P0-001'])
        self.assertNotIn('P0-002', state)
        # Only main's tracker completion unlocks a dependency.
        seed = self.root / 'seed'
        tracker = json.loads((seed / scheduler.TRACKER).read_text())
        tracker['tasks'][0]['status'] = 'completed'
        scheduler.save_tracker(seed / scheduler.TRACKER, tracker)
        git(seed, 'commit', '-am', 'complete first task')
        git(seed, 'push', 'origin', 'main')
        self.assertEqual(self.panel.sync(event)['prepared'], ['P0-002'])
        self.assertEqual(self.panel.snapshot()['state']['tasks']['P0-001']['status'], 'completed')

    def test_non_main_merge_is_ignored(self):
        for event in ({'action': 'opened'}, {'action': 'closed', 'pull_request': {
                'merged_at': 'today', 'base': {'ref': 'task/P0-001'}}}):
            self.assertTrue(self.panel.sync(event)['ignored'])
        self.assertFalse(self.panel.state_file.exists())

    def test_preparation_recovers_after_push_failure(self):
        with patch.object(scheduler, 'push_opening_commit', side_effect=OSError('offline')):
            with self.assertRaises(OSError):
                self.panel.sync()
        self.assertEqual(self.panel.sync()['prepared'], ['P0-001'])
        self.assertEqual(self.commits_ahead(), '1')

    def test_refine_and_implement_pass_correct_prompt_and_branch_and_reuse(self):
        self.panel.sync()
        real_run = scheduler.run
        submitted = []
        def run(*args, **kwargs):
            if args[0] == 'codex':
                submitted.append(args)
                return f'https://chatgpt.com/codex/cloud/tasks/task_e_{len(submitted)}'
            return real_run(*args, **kwargs)
        with patch.object(scheduler, 'run', side_effect=run):
            for stage, prefix in [('refine', 'Refine the brief'), ('implement', 'Implement')]:
                result = self.panel.stage('P0-001', stage)
                self.assertEqual(result['status'], 'submitted')
                self.assertTrue(self.panel.stage('P0-001', stage)['reused'])
                self.assertIn('--branch', submitted[-1])
                self.assertEqual(submitted[-1][-2], 'task/P0-001')
                self.assertTrue(submitted[-1][-1].startswith(prefix))
                self.assertIn('pull request on `task/P0-001`', submitted[-1][-1])
        self.assertEqual(len(submitted), 2)
        page = render(**{'tracker': self.panel.tracker(), 'state': self.panel.snapshot()['state'], 'message': '', 'busy': False})
        for label in ('Refine', 'Implement', 'Audit'):
            self.assertIn('>' + label + '</button>', page)
        self.assertIn('navigator.clipboard.writeText', page)

    def test_audit_formats_selected_pr_head_without_launching(self):
        self.panel.sync()
        pulls = [{'number': n, 'url': f'https://github.com/owner/repo/pull/{n}', 'headRefOid': f'head{n}'} for n in (7, 8)]
        with patch.object(scheduler, 'task_pull_requests', return_value=pulls), patch.object(scheduler, 'run') as run:
            with self.assertRaises(Problem) as error:
                self.panel.stage('P0-001', 'audit')
            self.assertEqual(error.exception.status, 409)
            result = self.panel.stage('P0-001', 'audit', 8)
            self.assertIn(pulls[1]['url'], result['prompt'])
            self.assertIn('head8', result['prompt'])
            self.assertNotIn('$', result['prompt'])
            run.assert_not_called()

    def test_lock_and_dirty_main_refuse_operations(self):
        lock = scheduler.acquire_lock(self.panel.state_file)
        try:
            with self.assertRaises(Problem) as error:
                self.panel.sync()
            self.assertEqual(error.exception.status, 409)
        finally:
            lock.close()
        (self.panel.clone / 'untracked').write_text('keep me')
        with self.assertRaises(Problem):
            self.panel.sync()
        self.assertEqual((self.panel.clone / 'untracked').read_text(), 'keep me')

    def test_http_auth_validation_and_merge_flow(self):
        handler = type('TestHandler', (Handler,), {'service': self.panel, 'token': 'test-token'})
        server = ThreadingHTTPServer(('127.0.0.1', 0), handler)
        thread = threading.Thread(target=server.serve_forever, daemon=True)
        thread.start()
        def request(path, body='{}', token='test-token'):
            connection = HTTPConnection(*server.server_address, timeout=5)
            connection.request('POST', path, body, {'Content-Type': 'application/json', 'Authorization': 'Bearer ' + token})
            response = connection.getresponse()
            result = response.status, json.loads(response.read())
            connection.close()
            return result
        try:
            self.assertEqual(request('/api/sync', token='wrong')[0], 401)
            self.assertEqual(request('/api/sync', body='[]')[0], 400)
            self.assertEqual(request('/api/sync', body='{')[0], 400)
            self.assertFalse(self.panel.state_file.exists())
            status, data = request('/api/events/merge', json.dumps({'action': 'closed', 'pull_request': {'merged_at': 'today', 'base': {'ref': 'main'}}}))
            self.assertEqual(status, 200)
            self.assertEqual(data['prepared'], ['P0-001'])
            self.assertEqual(request('/api/tasks/P0-999/refine')[0], 404)
            self.assertEqual(request('/api/tasks/P0-002/refine')[0], 409)
        finally:
            server.shutdown()
            server.server_close()
            thread.join()

