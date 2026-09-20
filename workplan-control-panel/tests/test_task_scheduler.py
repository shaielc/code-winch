import json
import threading
import time
import unittest
from http.client import HTTPConnection
from unittest.mock import patch

from control_panel import state as task_state
from control_panel.task_scheduler import TaskScheduler, TaskError
from control_panel.integrations import codex, repository
from control_panel.api import Handler, ThreadingHTTPServer
from control_panel.ui import render
from helpers import GitRepositoryFixture, git, TRACKER


class PanelFlowTests(GitRepositoryFixture, unittest.TestCase):
    def setUp(self):
        super().setUp()
        seed = self.root / 'seed'
        tracker = json.loads(json.dumps(TRACKER))
        tracker['tasks'][0].update(depends_on=[], brief='phase-0/first.md')
        tracker['tasks'].append({**tracker['tasks'][0], 'id': 'P0-002', 'depends_on': ['P0-001']})
        repository.save_tracker(seed / repository.TRACKER, tracker)
        git(seed, 'commit', '-am', 'task graph')
        git(seed, 'push', 'origin', 'main')
        self.panel = TaskScheduler( self.clone('panel'), self.root / 'state.json',
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
        tracker = json.loads((seed / repository.TRACKER).read_text())
        tracker['tasks'][0]['status'] = 'completed'
        repository.save_tracker(seed / repository.TRACKER, tracker)
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
        with patch.object(repository, 'push_opening_commit', side_effect=OSError('offline')):
            with self.assertRaises(TaskError):
                self.panel.sync()
        self.assertEqual(self.panel.sync()['prepared'], ['P0-001'])
        self.assertEqual(self.commits_ahead(), '1')

    def test_refine_and_implement_pass_correct_prompt_and_branch_and_reuse(self):
        self.panel.sync()
        real_run = codex.run
        submitted = []
        def run(*args, **kwargs):
            if args[0] == 'codex':
                submitted.append(args)
                return f'https://chatgpt.com/codex/cloud/tasks/task_e_{len(submitted)}'
            return real_run(*args, **kwargs)
        with patch.object(codex, 'run', side_effect=run):
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

    def test_expire_releases_reservation_and_retains_stage_conversations(self):
        self.panel.sync()
        urls = ['https://chatgpt.com/codex/tasks/refine', 'https://chatgpt.com/codex/tasks/implement']
        with patch.object(codex, 'submit', side_effect=urls):
            self.panel.stage('P0-001', 'refine')
            self.panel.stage('P0-001', 'implement')
        before = self.panel.snapshot()
        page = render(before['tracker'], before['state'], '', False)
        self.assertEqual(page.count('data-action="expire"'), 2)
        self.assertEqual(self.panel.expire('P0-001'), {'expired': 'P0-001'})
        snapshot = self.panel.snapshot()
        effective = task_state.effective_tracker(snapshot['tracker'], snapshot['state'])
        self.assertEqual(effective['tasks'][0]['status'], 'pending')
        self.assertEqual(snapshot['state']['tasks']['P0-001']['stages'], before['state']['tasks']['P0-001']['stages'])
        with self.assertRaises(TaskError):
            self.panel.stage('P0-001', 'implement')
        page = render(snapshot['tracker'], snapshot['state'], '', False)
        self.assertNotIn('data-action="expire"', page)
        for stage, url in zip(('refine', 'implement'), urls):
            self.assertEqual(page.count(f'href="{url}" data-stage="{stage}"'), 2)
        # A later sync reclaims the existing branch exactly once, retaining history.
        self.assertEqual(self.panel.sync()['prepared'], ['P0-001'])
        self.assertEqual(self.commits_ahead(), '1')
        with patch.object(codex, 'submit') as submit:
            self.assertTrue(self.panel.stage('P0-001', 'refine')['reused'])
            submit.assert_not_called()

    def test_expire_respects_lock_and_completed_tracker(self):
        self.panel.sync()
        lock = task_state.acquire_lock(self.panel.state_file)
        try:
            with self.assertRaises(TaskError) as error:
                self.panel.expire('P0-001')
            self.assertEqual(error.exception.status, 409)
        finally:
            lock.close()
        tracker = self.panel.tracker()
        tracker['tasks'][0]['status'] = 'completed'
        task_state.write_json(self.panel.tracker_path, tracker)
        with self.assertRaises(TaskError):
            self.panel.expire('P0-001')
        with self.assertRaises(TaskError):
            self.panel.expire('P0-002')

    def test_audit_formats_selected_pr_head_without_launching(self):
        self.panel.sync()
        pulls = [{'number': n, 'url': f'https://github.com/owner/repo/pull/{n}', 'headRefOid': f'head{n}'} for n in (7, 8)]
        with patch.object(repository, 'task_pull_requests', return_value=pulls), patch.object(codex, 'run') as run:
            with self.assertRaises(TaskError) as error:
                self.panel.stage('P0-001', 'audit')
            self.assertEqual(error.exception.status, 409)
            result = self.panel.stage('P0-001', 'audit', 8)
            self.assertIn(pulls[1]['url'], result['prompt'])
            self.assertIn('head8', result['prompt'])
            self.assertNotIn('$', result['prompt'])
            run.assert_not_called()

    def test_lock_and_dirty_main_refuse_operations(self):
        lock = task_state.acquire_lock(self.panel.state_file)
        try:
            with self.assertRaises(TaskError) as error:
                self.panel.sync()
            self.assertEqual(error.exception.status, 409)
        finally:
            lock.close()
        (self.panel.clone / 'untracked').write_text('keep me')
        with self.assertRaises(TaskError):
            self.panel.sync()
        self.assertEqual((self.panel.clone / 'untracked').read_text(), 'keep me')

    def test_background_sync_returns_before_git_finishes_and_reports_failure(self):
        entered, release = threading.Event(), threading.Event()
        def slow_sync(event):
            entered.set()
            release.wait(3)
            raise TaskError(502, "Main refreshed, but task preparation failed", prepared=[])
        with patch.object(self.panel, 'sync', side_effect=slow_sync):
            job = self.panel.start_sync()
            self.assertTrue(entered.wait(1))
            self.assertEqual(self.panel.sync_status(job['id'])['status'], 'running')
            release.set()
            deadline = time.monotonic() + 3
            while self.panel.sync_status(job['id'])['status'] == 'running' and time.monotonic() < deadline:
                time.sleep(0.01)
            result = self.panel.sync_status(job['id'])
            self.assertEqual(result['status'], 'failed')
            self.assertIn('preparation failed', result['error'])

    def test_new_merge_during_sync_requests_another_main_pull(self):
        entered, release = threading.Event(), threading.Event()
        def sync(event):
            entered.set()
            release.wait(3)
            return {'prepared': []}
        with patch.object(self.panel, 'sync', side_effect=sync) as sync_call:
            job = self.panel.start_sync()
            self.assertTrue(entered.wait(1))
            again = self.panel.start_sync({'action': 'closed'})
            self.assertEqual(again['id'], job['id'])
            release.set()
            deadline = time.monotonic() + 3
            while self.panel.sync_status(job['id'])['status'] == 'running' and time.monotonic() < deadline:
                time.sleep(0.01)
            self.assertEqual(self.panel.sync_status(job['id'])['status'], 'succeeded')
            self.assertEqual(sync_call.call_count, 2)

    def test_in_progress_main_task_without_local_state_gets_prepared(self):
        seed = self.root / 'seed'
        tracker = json.loads((seed / repository.TRACKER).read_text())
        tracker['tasks'][0]['status'] = 'in_progress'
        repository.save_tracker(seed / repository.TRACKER, tracker)
        git(seed, 'commit', '-am', 'existing task in progress')
        git(seed, 'push', 'origin', 'main')
        self.assertEqual(self.panel.sync()['prepared'], ['P0-001'])
        self.assertTrue(self.panel.snapshot()['state']['tasks']['P0-001']['prepared'])

    def test_http_auth_validation_and_merge_flow(self):
        handler = type('TestHandler', (Handler,), {'scheduler': self.panel, 'token': 'test-token'})
        server = ThreadingHTTPServer(('127.0.0.1', 0), handler)
        thread = threading.Thread(target=server.serve_forever, daemon=True)
        thread.start()
        def request(path, body='{}', token='test-token', method='POST'):
            connection = HTTPConnection(*server.server_address, timeout=5)
            connection.request(method, path, body if method == 'POST' else None, {'Content-Type': 'application/json', 'Authorization': 'Bearer ' + token})
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
            self.assertEqual(status, 202)
            deadline = time.monotonic() + 5
            job_id = data['id']
            while data['status'] == 'running' and time.monotonic() < deadline:
                time.sleep(0.01)
                status, data = request('/api/sync/' + job_id, method='GET')
            self.assertEqual(data['status'], 'succeeded')
            self.assertEqual(data['result']['prepared'], ['P0-001'])
            self.assertEqual(request('/api/tasks/P0-999/refine')[0], 404)
            self.assertEqual(request('/api/tasks/P0-002/refine')[0], 409)
            self.assertEqual(request('/api/tasks/P0-001/expire', token='wrong')[0], 401)
            self.assertFalse(self.panel.snapshot()['state']['tasks']['P0-001']['expired'])
            self.assertEqual(request('/api/tasks/P0-001/expire'), (200, {'expired': 'P0-001'}))
            self.assertEqual(request('/api/tasks/P0-999/expire')[0], 404)
        finally:
            server.shutdown()
            server.server_close()
            thread.join()

