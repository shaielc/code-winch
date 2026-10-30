import json
import re
import subprocess
import threading
import time
import unittest
from http.client import HTTPConnection
from unittest.mock import patch

from control_panel import state as task_state
from control_panel.task_scheduler import TaskScheduler, TaskError
from control_panel.integrations import claude, codex, repository
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
                                  self.root / 'tracker.json', 'test-env', 1,
                                  'https://api.example/fire', 'routine-token')
        # No test may reach the real gh; the tests about the lookup replace its answer.
        patcher = patch.object(repository, 'task_pull_request', return_value=None)
        self.pull_request = patcher.start()
        self.addCleanup(patcher.stop)

    def test_sync_records_the_pull_request_into_main_until_it_merges(self):
        record = lambda: self.panel.snapshot()['state']['tasks']['P0-001']
        self.panel.sync()
        # The pass that prepares a task does not look for its pull request yet.
        self.pull_request.assert_not_called()
        self.panel.sync()
        self.assertNotIn('pull_request', record())
        self.pull_request.return_value = {'url': 'https://example/pr/5', 'state': 'OPEN'}
        self.panel.sync()
        self.assertEqual((record()['pull_request'], record()['pull_request_state']),
                         ('https://example/pr/5', 'OPEN'))
        self.pull_request.return_value = {'url': 'https://example/pr/5', 'state': 'MERGED'}
        self.panel.sync()
        # A merged pull request is final, so later passes stop asking.
        asked = self.pull_request.call_count
        self.pull_request.return_value = None
        self.panel.sync()
        self.assertEqual(self.pull_request.call_count, asked)
        self.assertEqual(record()['pull_request_state'], 'MERGED')

    def test_creating_the_pull_request_opens_one_unless_one_is_already_open(self):
        url = 'https://example/pr/12'
        with patch.object(repository, 'create_task_pull_request', return_value=url) as create:
            self.panel.sync()
            for task, status in (('P9-999', 404), ('P0-002', 409)):
                with self.assertRaises(TaskError) as refused:
                    self.panel.create_pull_request(task)
                self.assertEqual(refused.exception.status, status)
            result = self.panel.create_pull_request('P0-001')
            self.assertEqual((result['created'], result['pull_request'], result['pull_request_state']),
                             (True, url, 'OPEN'))
            self.assertEqual(create.call_args.args[1:], ('P0-001', 'Upstream'))
            # Once one is open, a second click reports it instead of opening another.
            self.pull_request.return_value = {'url': url, 'state': 'OPEN'}
            self.assertFalse(self.panel.create_pull_request('P0-001')['created'])
            create.assert_called_once()
            self.pull_request.return_value = None
            create.side_effect = subprocess.CalledProcessError(1, ['gh'])
            with self.assertRaises(TaskError) as failed:
                self.panel.create_pull_request('P0-001')
            self.assertEqual(failed.exception.status, 502)
            self.assertIn('GitHub App', str(failed.exception))

    def test_refreshing_one_task_looks_up_only_its_pull_request(self):
        self.panel.sync()
        self.pull_request.return_value = {'url': 'https://example/pr/9', 'state': 'MERGED'}
        result = self.panel.refresh_pull_request('P0-001')
        self.assertEqual((result['pull_request'], result['pull_request_state']),
                         ('https://example/pr/9', 'MERGED'))
        self.assertEqual(self.pull_request.call_args.args[1], 'P0-001')
        # An explicit refresh looks again even though a merge is already recorded.
        self.pull_request.return_value = {'url': 'https://example/pr/10', 'state': 'OPEN'}
        self.assertEqual(self.panel.refresh_pull_request('P0-001')['pull_request_state'], 'OPEN')
        for task, status in (('P9-999', 404), ('P0-002', 409)):
            with self.assertRaises(TaskError) as refused:
                self.panel.refresh_pull_request(task)
            self.assertEqual(refused.exception.status, status)
        self.pull_request.side_effect = subprocess.CalledProcessError(1, ['gh'])
        with self.assertRaises(TaskError) as failed:
            self.panel.refresh_pull_request('P0-001')
        self.assertEqual(failed.exception.status, 502)
        self.assertEqual(self.panel.snapshot()['state']['tasks']['P0-001']['pull_request_state'], 'OPEN')

    def test_a_pull_request_that_disappears_is_forgotten_and_a_failed_lookup_is_not_fatal(self):
        self.panel.sync()
        self.pull_request.return_value = {'url': 'https://example/pr/5', 'state': 'OPEN'}
        self.panel.sync()
        self.pull_request.return_value = None
        self.panel.sync()
        self.assertNotIn('pull_request', self.panel.snapshot()['state']['tasks']['P0-001'])
        self.pull_request.side_effect = subprocess.CalledProcessError(1, ['gh'])
        self.assertEqual(self.panel.sync()['prepared'], [])

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
                self.assertEqual(submitted[-1][-2], f'{stage}/P0-001-1')
                self.assertTrue(submitted[-1][-1].startswith(prefix))
                unwrapped = ' '.join(submitted[-1][-1].split())
                self.assertIn(f'`{stage}/P0-001-1`', unwrapped)
                self.assertNotIn('Include `Task:', unwrapped)
        self.assertEqual(len(submitted), 2)
        stages = self.panel.snapshot()['state']['tasks']['P0-001']['stages']
        base = git(self.origin, 'rev-parse', 'task/P0-001')
        for stage in ('refine', 'implement'):
            self.assertEqual(git(self.origin, 'rev-parse', f'{stage}/P0-001-1'), base)
            self.assertEqual(stages[stage][0]['branch'], f'{stage}/P0-001-1')
            self.assertEqual(stages[stage][0]['base'], base)
        page = render(self.panel.tracker(), self.panel.snapshot()['state'], '', False)
        # Each stage is one control per view — the table's and the tree's — and opens the
        # conversation its dispatch returned, pointed at the view that serves it.
        for number in (1, 2):
            self.assertEqual(
                page.count(f'data-state="open" href="https://chatgpt.com/remote/task_e_{number}"'), 2)
        for stage in ('refine', 'implement'):
            self.assertEqual(page.count(f'>{stage}/P0-001-1</small>'), 2)
        self.assertNotIn('clipboard', page)

    def stages(self, stage):
        return self.panel.snapshot()['state']['tasks']['P0-001'].get('stages', {}).get(stage, [])

    def stage_branch_tips(self, stage):
        refs = git(self.origin, 'for-each-ref', '--format=%(refname:short) %(objectname)',
                   f'refs/heads/{stage}')
        return dict(line.split() for line in refs.splitlines())

    def cell_states(self, stage):
        page = render(self.panel.tracker(), self.panel.snapshot()['state'], '', False)
        return re.findall(rf'data-task="P0-001" data-stage="{stage}"[^>]*>'
                          r'<a class="stage-icon" data-state="(\w+)"', page)

    def test_failed_stage_branch_push_submits_nothing_and_keeps_the_attempt_set_aside(self):
        self.panel.sync()
        with patch.object(repository, 'create_stage_branch', side_effect=OSError('rejected')), \
                patch.object(codex, 'submit') as submit:
            with self.assertRaises(TaskError) as error:
                self.panel.stage('P0-001', 'implement')
        self.assertEqual(error.exception.status, 502)
        self.assertIn('P0-001 could not create and push', str(error.exception))
        submit.assert_not_called()
        # Whether the push reached origin is unknown, so its number stays on record. It is not
        # a submission: nothing was sent, it links to no conversation, and it blocks no retry.
        (attempt,) = self.stages('implement')
        self.assertEqual((attempt['status'], attempt['failure'], attempt['branch']),
                         ('expired', 'push', 'implement/P0-001-1'))
        self.assertNotIn('task_url', attempt)
        self.assertEqual(self.cell_states('implement'), ['launch', 'launch'])

    def test_a_push_that_lands_and_then_errors_does_not_strand_the_stage(self):
        self.panel.sync()
        real = repository.github_run

        def lands_then_fails(*command, cwd, capture=True):
            output = real(*command, cwd=cwd, capture=capture)
            if command[:2] == ('git', 'push'):
                raise subprocess.CalledProcessError(128, command, stderr='connection reset')
            return output

        with patch.object(repository, 'github_run', side_effect=lands_then_fails), \
                patch.object(codex, 'submit') as submit:
            with self.assertRaises(TaskError) as error:
                self.panel.stage('P0-001', 'implement')
        self.assertEqual(error.exception.status, 502)
        submit.assert_not_called()
        base = git(self.origin, 'rev-parse', 'task/P0-001')
        # The panel was told the push failed, but the branch is on origin.
        self.assertEqual(self.stage_branch_tips('implement'), {'implement/P0-001-1': base})
        self.assertEqual([(a['status'], a['branch']) for a in self.stages('implement')],
                         [('expired', 'implement/P0-001-1')])

        with patch.object(codex, 'submit', return_value='https://chatgpt.com/codex/cloud/tasks/retry') as submit:
            result = self.panel.stage('P0-001', 'implement')
            self.assertEqual((result['status'], result['branch']), ('submitted', 'implement/P0-001-2'))
            self.assertEqual(submit.call_args.args[2], 'implement/P0-001-2')
            prompt = ' '.join(submit.call_args.args[3].split())
            self.assertIn('`implement/P0-001-2`', prompt)
            self.assertNotIn('implement/P0-001-1', prompt)
            self.assertTrue(self.panel.stage('P0-001', 'implement')['reused'])
            submit.assert_called_once()
        # The branch that landed is left as it was, and the retry is a branch of its own.
        self.assertEqual(self.stage_branch_tips('implement'),
                         {'implement/P0-001-1': base, 'implement/P0-001-2': base})
        self.assertEqual([(a['status'], a['branch']) for a in self.stages('implement')],
                         [('expired', 'implement/P0-001-1'), ('submitted', 'implement/P0-001-2')])

    def test_lost_attempt_history_does_not_make_the_panel_collide_with_origin(self):
        self.panel.sync()
        with patch.object(codex, 'submit', return_value='https://chatgpt.com/codex/cloud/tasks/one'):
            self.panel.stage('P0-001', 'implement')
        base = git(self.origin, 'rev-parse', 'task/P0-001')
        elsewhere = git(self.origin, 'rev-parse', 'main')
        git(self.origin, 'branch', 'implement/P0-001-3', elsewhere)
        state = task_state.load_state(self.panel.state_file)
        state['tasks']['P0-001']['stages'] = {}
        task_state.write_json(self.panel.state_file, state)

        # With no record, origin alone decides: past its highest number, not into the gap.
        with patch.object(codex, 'submit', return_value='https://chatgpt.com/codex/cloud/tasks/two') as submit:
            result = self.panel.stage('P0-001', 'implement')
        self.assertEqual((result['status'], result['branch']), ('submitted', 'implement/P0-001-4'))
        self.assertEqual(submit.call_args.args[2], 'implement/P0-001-4')
        self.assertEqual(self.stage_branch_tips('implement'),
                         {'implement/P0-001-1': base, 'implement/P0-001-3': elsewhere,
                          'implement/P0-001-4': base})

        # A number on record that origin no longer has (a merged branch, deleted) is not reused.
        git(self.origin, 'branch', '-D', 'implement/P0-001-1', 'implement/P0-001-3', 'implement/P0-001-4')
        state = task_state.load_state(self.panel.state_file)
        state['tasks']['P0-001']['stages']['implement'][0].update(branch='implement/P0-001-5')
        task_state.write_json(self.panel.state_file, state)
        self.panel.expire_stage('P0-001', 'implement')
        with patch.object(codex, 'submit', return_value='https://chatgpt.com/codex/cloud/tasks/three'):
            self.assertEqual(self.panel.stage('P0-001', 'implement')['branch'], 'implement/P0-001-6')

    def test_a_stage_branch_found_on_origin_moves_the_attempt_to_the_next_number(self):
        self.panel.sync()
        elsewhere = git(self.origin, 'rev-parse', 'main')
        for number in (1, 2):
            git(self.origin, 'branch', f'implement/P0-001-{number}', elsewhere)
        base = git(self.origin, 'rev-parse', 'task/P0-001')
        # A lookup that went stale, as when the branch lands between the lookup and the push.
        with patch.object(repository, 'next_stage_number', return_value=1), \
                patch.object(codex, 'submit', return_value='https://chatgpt.com/codex/cloud/tasks/one') as submit:
            result = self.panel.stage('P0-001', 'implement')
        self.assertEqual((result['status'], result['branch']), ('submitted', 'implement/P0-001-3'))
        self.assertEqual(submit.call_args.args[2], 'implement/P0-001-3')
        prompt = ' '.join(submit.call_args.args[3].split())
        self.assertIn('`implement/P0-001-3`', prompt)
        self.assertNotIn('implement/P0-001-1', prompt)
        # One attempt, moved to the number it took, not one per number it tried.
        (attempt,) = self.stages('implement')
        self.assertEqual((attempt['status'], attempt['branch'], attempt['base']),
                         ('submitted', 'implement/P0-001-3', base))
        self.assertEqual(self.stage_branch_tips('implement'),
                         {'implement/P0-001-1': elsewhere, 'implement/P0-001-2': elsewhere,
                          'implement/P0-001-3': base})

    def test_the_number_taken_after_a_collision_is_recorded_before_codex_is_called(self):
        self.panel.sync()
        git(self.origin, 'branch', 'implement/P0-001-1', git(self.origin, 'rev-parse', 'main'))
        with patch.object(repository, 'next_stage_number', return_value=1), \
                patch.object(codex, 'submit', side_effect=TimeoutError()):
            with self.assertRaises(TaskError):
                self.panel.stage('P0-001', 'implement')
        # The reply may have been lost after Codex accepted, and it was sent on the new branch.
        (attempt,) = self.stages('implement')
        self.assertEqual((attempt['status'], attempt['branch']), ('submitting', 'implement/P0-001-2'))
        self.assertEqual(self.cell_states('implement'), ['unknown', 'unknown'])

    def test_an_unreadable_stage_branch_listing_pushes_submits_and_records_nothing(self):
        self.panel.sync()
        with patch.object(repository, 'next_stage_number',
                          side_effect=subprocess.CalledProcessError(128, ['git', 'ls-remote'])), \
                patch.object(repository, 'create_stage_branch') as create, \
                patch.object(codex, 'submit') as submit:
            with self.assertRaises(TaskError) as error:
                self.panel.stage('P0-001', 'implement')
        self.assertEqual(error.exception.status, 502)
        self.assertIn('P0-001 could not read its implement stage branches', str(error.exception))
        create.assert_not_called()
        submit.assert_not_called()
        self.assertEqual(self.stages('implement'), [])

    def test_attempts_recorded_before_this_change_still_load_render_and_are_numbered_past(self):
        self.panel.sync()
        head = git(self.origin, 'rev-parse', 'task/P0-001')
        codex_url = lambda name: f'https://chatgpt.com/codex/cloud/tasks/task_e_{name}'
        state = task_state.load_state(self.panel.state_file)
        state['tasks']['P0-001']['stages'] = {
            # From before stage branches: no branch, no base.
            'refine': [{'status': 'expired', 'head': 'f' * 40, 'task_url': codex_url('old'),
                        'updated_at': '2026-10-01T00:00:00+00:00'},
                       {'status': 'submitted', 'head': head, 'task_url': codex_url('new'),
                        'updated_at': '2026-10-02T00:00:00+00:00'}],
            # From the first stage-branch release: a branch and its base, and no failure mark.
            'implement': [{'status': 'submitted', 'head': head, 'branch': 'implement/P0-001-1',
                           'base': head, 'task_url': codex_url('branch'),
                           'updated_at': '2026-10-03T00:00:00+00:00'}],
        }
        task_state.write_json(self.panel.state_file, state)

        page = render(self.panel.tracker(), self.panel.snapshot()['state'], '', False)
        for name in ('new', 'branch'):
            self.assertEqual(page.count(f'data-state="open" href="https://chatgpt.com/remote/task_e_{name}"'), 2)
        self.assertEqual(page.count('<span class="earlier" hidden><a href="https://chatgpt.com/remote/task_e_old"'), 2)
        # Only the record that has a branch names one.
        self.assertEqual(page.count('class="stage-branch"'), 2)
        self.assertEqual(page.count('>implement/P0-001-1</small>'), 2)

        with patch.object(codex, 'submit', side_effect=[codex_url('again-1'), codex_url('again-2')]) as submit:
            # Still standing at this revision, so no branch is made and nothing is sent.
            for stage in ('refine', 'implement'):
                self.assertTrue(self.panel.stage('P0-001', stage)['reused'])
            submit.assert_not_called()
            self.assertEqual(self.stage_branch_tips('refine'), {})
            # A relaunch numbers past what is on record, and a record with no branch holds no number.
            for stage, branch in (('refine', 'refine/P0-001-1'), ('implement', 'implement/P0-001-2')):
                self.panel.expire_stage('P0-001', stage)
                self.assertEqual(self.panel.stage('P0-001', stage)['branch'], branch)
        self.assertEqual([a['status'] for a in self.stages('refine')], ['expired', 'expired', 'submitted'])
        self.assertEqual([a.get('branch') for a in self.stages('refine')], [None, None, 'refine/P0-001-1'])
        self.assertEqual([a['status'] for a in self.stages('implement')], ['expired', 'submitted'])

    def test_rebase_moves_the_task_onto_main_and_reopens_its_stages(self):
        self.panel.sync()
        with patch.object(repository, 'task_pull_requests', return_value=[]):
            for task, status in (('P9-999', 404), ('P0-002', 409)):
                with self.assertRaises(TaskError) as refused:
                    self.panel.rebase(task)
                self.assertEqual(refused.exception.status, status)
            self.assertEqual(self.panel.rebase('P0-001'), {'task': 'P0-001', 'rebased': False})
            with patch.object(codex, 'submit', side_effect=['https://example/codex/1',
                                                            'https://example/codex/2']) as submit:
                self.panel.stage('P0-001', 'implement')
                seed = self.root / 'seed'
                (seed / 'upstream.txt').write_text('new\n')
                git(seed, 'add', 'upstream.txt')
                git(seed, 'commit', '-m', 'upstream')
                git(seed, 'push', 'origin', 'main')
                result = self.panel.rebase('P0-001')
                self.assertTrue(result['rebased'])
                self.assertEqual(result['main_commit'], git(seed, 'rev-parse', 'HEAD'))
                # The moved tip is a new revision, so implement is sent again.
                self.assertNotIn('reused', self.panel.stage('P0-001', 'implement'))
                self.assertEqual(submit.call_count, 2)

    def test_rebase_refuses_fork_pull_requests_and_reports_conflicts(self):
        self.panel.sync()
        fork = [{'number': 3, 'url': 'https://example/pr/3', 'headRefName': 'x',
                 'headRefOid': 'abc', 'isCrossRepository': True}]
        with patch.object(repository, 'task_pull_requests', return_value=fork), \
                patch.object(repository, 'rebase_task_branch') as rebase:
            with self.assertRaises(TaskError) as refused:
                self.panel.rebase('P0-001')
            self.assertEqual(refused.exception.status, 409)
            self.assertIn('https://example/pr/3', str(refused.exception))
            rebase.assert_not_called()
        with patch.object(repository, 'task_pull_requests', return_value=[]), \
                patch.object(repository, 'rebase_task_branch',
                             side_effect=repository.RebaseConflict('task/P0-001 conflicts')):
            with self.assertRaises(TaskError) as conflicted:
                self.panel.rebase('P0-001')
            self.assertEqual(conflicted.exception.status, 409)
            self.assertIn('Nothing was pushed', str(conflicted.exception))

    def test_expire_releases_reservation_and_retains_stage_conversations(self):
        self.panel.sync()
        urls = ['https://chatgpt.com/codex/cloud/tasks/refine',
                'https://chatgpt.com/codex/cloud/tasks/implement']
        with patch.object(codex, 'submit', side_effect=urls):
            self.panel.stage('P0-001', 'refine')
            self.panel.stage('P0-001', 'implement')
        before = self.panel.snapshot()
        page = render(before['tracker'], before['state'], '', False)
        self.assertEqual(page.count('data-action="expire"'), 2)
        # Expiry and Rebase rewrite the branch, so their buttons ask first; Prepare does not.
        self.assertEqual(page.count('data-confirm="Expire P0-001? This deletes task/P0-001 on GitHub'), 2)
        self.assertEqual(page.count('data-confirm="Rebase task/P0-001 onto main?'), 2)
        self.assertEqual(page.count('data-confirm='), 4)
        self.assertEqual(self.panel.expire('P0-001'), {'expired': 'P0-001'})
        snapshot = self.panel.snapshot()
        effective = task_state.effective_tracker(snapshot['tracker'], snapshot['state'])
        self.assertEqual(effective['tasks'][0]['status'], 'pending')
        # Every stage submission is expired with the task, and keeps its conversation.
        stages = snapshot['state']['tasks']['P0-001']['stages']
        self.assertEqual({stage: [a['status'] for a in attempts] for stage, attempts in stages.items()},
                         {'refine': ['expired'], 'implement': ['expired']})
        self.assertEqual([stages[s][0]['task_url'] for s in ('refine', 'implement')], urls)
        with self.assertRaises(TaskError):
            self.panel.stage('P0-001', 'implement')
        page = render(snapshot['tracker'], snapshot['state'], '', False)
        self.assertNotIn('data-action="expire"', page)
        for url in urls:
            self.assertNotIn(f'data-state="open" href="{url}"', page)
            self.assertEqual(page.count(f'<span class="earlier" hidden><a href="{url}"'), 2)
        # Expiry deletes the branch upstream and in the panel's clone.
        seed = self.root / 'seed'
        self.assertEqual(git(seed, 'ls-remote', '--heads', 'origin', 'task/P0-001'), '')
        self.assertEqual(git(self.panel.clone, 'branch', '--remotes', '--list', 'origin/task/P0-001'), '')
        self.assertFalse(snapshot['state']['tasks']['P0-001']['prepared'])
        # A later sync starts the branch afresh from main, and its stages submit anew.
        self.assertEqual(self.panel.sync()['prepared'], ['P0-001'])
        self.assertEqual(self.commits_ahead(), '1')
        with patch.object(codex, 'submit', return_value='https://chatgpt.com/codex/cloud/tasks/again') as submit:
            self.assertFalse(self.panel.stage('P0-001', 'refine').get('reused'))
            submit.assert_called_once()

    def test_expire_replaces_a_rewritten_branch(self):
        """The branch moved under the panel's tracking ref; expiry must still clear it."""
        self.panel.sync()
        worker = self.clone('worker')
        git(worker, 'checkout', '--quiet', '-b', 'task/P0-001', 'origin/main')
        (worker / 'unrelated.md').write_text('rewritten\n')
        git(worker, 'add', '.')
        git(worker, 'commit', '--quiet', '-m', 'rewrite the branch')
        git(worker, 'push', '--quiet', '--force', 'origin', 'HEAD:task/P0-001')
        self.panel.expire('P0-001')
        self.assertEqual(self.panel.sync()['prepared'], ['P0-001'])
        self.assertEqual(self.commits_ahead(), '1')
        self.assertNotIn('unrelated.md', git(self.panel.clone, 'ls-tree', '--name-only',
                                             'origin/task/P0-001'))

    def test_expire_keeps_the_reservation_when_the_branch_cannot_be_deleted(self):
        self.panel.sync()
        with patch.object(repository, 'delete_task_branch', side_effect=OSError('offline')):
            with self.assertRaises(TaskError) as error:
                self.panel.expire('P0-001')
        self.assertEqual(error.exception.status, 502)
        record = self.panel.snapshot()['state']['tasks']['P0-001']
        self.assertFalse(record['expired'])
        self.assertTrue(record['prepared'])

    def test_expired_and_prepared_task_is_recoverable_by_request(self):
        """The case a scheduling pass cannot serve: the slot is free but the queue is not."""
        self.panel.capacity = 2
        self.panel.sync()
        self.panel.expire('P0-001')
        # P0-002 depends on P0-001, so a scheduling pass has nothing else to offer either.
        self.assertEqual(self.panel.sync()['prepared'], ['P0-001'])
        self.panel.expire('P0-001')
        record = self.panel.snapshot()['state']['tasks']['P0-001']
        self.assertTrue(record['expired'])
        self.assertFalse(record['prepared'])
        self.assertEqual(self.panel.sync(only='P0-001')['prepared'], ['P0-001'])
        record = self.panel.snapshot()['state']['tasks']['P0-001']
        self.assertFalse(record['expired'])
        self.assertTrue(record['prepared'])
        self.assertEqual(self.commits_ahead(), '1')
        # Admission by request still refuses what the scheduling pass would refuse.
        with self.assertRaises(TaskError) as error:
            self.panel.sync(only='P0-002')
        self.assertEqual(error.exception.status, 409)
        self.assertIn('waits on P0-001', str(error.exception))
        with self.assertRaises(TaskError) as error:
            self.panel.sync(only='P0-404')
        self.assertEqual(error.exception.status, 404)

    def test_requested_preparation_needs_a_free_slot_unless_it_holds_one(self):
        # Independent on main, so only the free-slot rule can stand between them.
        seed = self.root / 'seed'
        tracker = json.loads((seed / repository.TRACKER).read_text())
        tracker['tasks'][1]['depends_on'] = []
        repository.save_tracker(seed / repository.TRACKER, tracker)
        git(seed, 'commit', '-am', 'independent tasks')
        git(seed, 'push', 'origin', 'main')
        self.panel.sync()
        self.assertEqual(self.panel.snapshot()['state']['tasks']['P0-001']['prepared'], True)
        # Capacity is one and P0-001 holds it, so P0-002 cannot be admitted by request.
        with self.assertRaises(TaskError) as error:
            self.panel.sync(only='P0-002')
        self.assertEqual(error.exception.status, 409)
        self.assertIn('slots are in flight', str(error.exception))
        # A task already holding its slot is admitted again, to retry its branch work.
        self.assertEqual(self.panel.sync(only='P0-001')['prepared'], ['P0-001'])

    def test_expiring_a_stage_allows_one_resubmission_at_the_same_revision(self):
        self.panel.sync()
        urls = ['https://chatgpt.com/codex/cloud/tasks/first',
                'https://chatgpt.com/codex/cloud/tasks/second']
        with patch.object(codex, 'submit', side_effect=urls):
            first = self.panel.stage('P0-001', 'refine')
            self.assertEqual(first['task_url'], urls[0])
            self.assertTrue(self.panel.stage('P0-001', 'refine')['reused'])
            result = self.panel.expire_stage('P0-001', 'refine')
            self.assertEqual(result['expired'], 'P0-001')
            self.assertEqual(result['stage'], 'refine')
            # The override is spent by the resubmission, not standing.
            second = self.panel.stage('P0-001', 'refine')
            self.assertEqual(second['task_url'], urls[1])
            self.assertFalse(second.get('reused'))
            self.assertTrue(self.panel.stage('P0-001', 'refine')['reused'])
        attempts = self.panel.snapshot()['state']['tasks']['P0-001']['stages']['refine']
        self.assertEqual([a['status'] for a in attempts], ['expired', 'submitted'])
        self.assertEqual([a['task_url'] for a in attempts], urls)
        self.assertEqual([a['branch'] for a in attempts], ['refine/P0-001-1', 'refine/P0-001-2'])
        self.assertEqual({a['head'] for a in attempts}, {attempts[0]['head']})
        # Both conversations survive the retirement: the icon opens the newest, and the
        # retired one stays reachable from that stage's right-click menu.
        page = render(self.panel.tracker(), self.panel.snapshot()['state'], '', False)
        self.assertEqual(page.count(f'data-state="open" href="{urls[1]}"'), 2)
        self.assertEqual(page.count(f'<span class="earlier" hidden><a href="{urls[0]}"'), 2)

    def test_expiring_a_stage_refuses_when_nothing_is_standing(self):
        self.panel.sync()
        with self.assertRaises(TaskError) as error:
            self.panel.expire_stage('P0-001', 'refine')
        self.assertEqual(error.exception.status, 409)
        self.assertIn('no recorded refine submission', str(error.exception))
        with patch.object(codex, 'submit', return_value='https://chatgpt.com/codex/tasks/one'):
            self.panel.stage('P0-001', 'refine')
        self.panel.expire_stage('P0-001', 'refine')
        with self.assertRaises(TaskError) as error:
            self.panel.expire_stage('P0-001', 'refine')
        self.assertEqual(error.exception.status, 409)
        self.assertIn('no live submission', str(error.exception))
        for task_id, stage in (('P0-404', 'refine'), ('P0-001', 'review')):
            with self.assertRaises(TaskError) as error:
                self.panel.expire_stage(task_id, stage)
            self.assertEqual(error.exception.status, 404)

    def test_a_moved_branch_revision_releases_a_stage_without_expiring_it(self):
        self.panel.sync()
        with patch.object(codex, 'submit', return_value='https://chatgpt.com/codex/tasks/one'):
            self.panel.stage('P0-001', 'refine')
        with patch.object(codex, 'submit', return_value='https://chatgpt.com/codex/tasks/two') as submit:
            self.assertTrue(self.panel.stage('P0-001', 'refine')['reused'])
            worker = self.clone('worker')
            git(worker, 'checkout', '--quiet', '-b', 'task/P0-001', 'origin/task/P0-001')
            (worker / 'refined.md').write_text('refined brief\n')
            git(worker, 'add', '.')
            git(worker, 'commit', '--quiet', '-m', 'refine the brief')
            git(worker, 'push', '--quiet', 'origin', 'HEAD:task/P0-001')
            self.assertFalse(self.panel.stage('P0-001', 'refine').get('reused'))
            submit.assert_called_once()
        attempts = self.panel.snapshot()['state']['tasks']['P0-001']['stages']['refine']
        self.assertEqual([a['status'] for a in attempts], ['submitted', 'submitted'])
        self.assertEqual(len({a['head'] for a in attempts}), 2)
        self.assertEqual([a['branch'] for a in attempts], ['refine/P0-001-1', 'refine/P0-001-2'])

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

    def pulls(self, *heads):
        return [{'number': n, 'url': f'https://github.com/owner/repo/pull/{n}', 'headRefOid': head}
                for n, head in zip((7, 8), heads)]

    def test_audit_fires_the_routine_for_the_selected_pr_head_and_reuses_it(self):
        self.panel.sync()
        sessions = ['https://claude.ai/code/session_1', 'https://claude.ai/code/session_2']
        pulls = self.pulls('head7', 'head8')
        with patch.object(repository, 'task_pull_requests', return_value=pulls), \
                patch.object(claude, 'fire', side_effect=sessions) as fire, \
                patch.object(codex, 'run') as run:
            with self.assertRaises(TaskError) as error:
                self.panel.stage('P0-001', 'audit')
            self.assertEqual(error.exception.status, 409)
            self.assertEqual(error.exception.details['pull_requests'], pulls)
            result = self.panel.stage('P0-001', 'audit', 8)
            self.assertEqual((result['status'], result['task_url']), ('submitted', sessions[0]))
            self.assertEqual((result['head'], result['pull_request']), ('head8', pulls[1]['url']))
            url, token, prompt = fire.call_args.args
            self.assertEqual((url, token), ('https://api.example/fire', 'routine-token'))
            self.assertIn(pulls[1]['url'], prompt)
            self.assertIn('head8', prompt)
            self.assertNotIn('$', prompt)
            # The same pull request head returns its session rather than firing again.
            self.assertTrue(self.panel.stage('P0-001', 'audit', 8)['reused'])
            # A new push to the pull request is a new revision to audit.
            pulls[1]['headRefOid'] = 'head8b'
            self.assertEqual(self.panel.stage('P0-001', 'audit', 8)['task_url'], sessions[1])
            self.assertEqual(fire.call_count, 2)
            run.assert_not_called()
        page = render(self.panel.tracker(), self.panel.snapshot()['state'], '', False)
        self.assertEqual(page.count(f'data-state="open" href="{sessions[1]}"'), 2)

    def test_audit_expiry_resubmits_the_newest_session(self):
        self.panel.sync()
        sessions = ['https://claude.ai/code/session_1', 'https://claude.ai/code/session_2']
        with patch.object(repository, 'task_pull_requests', return_value=self.pulls('head7')), \
                patch.object(claude, 'fire', side_effect=sessions):
            self.panel.stage('P0-001', 'audit')
            self.assertEqual(self.panel.expire_stage('P0-001', 'audit')['head'], 'head7')
            self.assertEqual(self.panel.stage('P0-001', 'audit')['task_url'], sessions[1])
        with self.assertRaises(TaskError) as error:
            self.panel.expire_stage('P0-002', 'audit')
        self.assertEqual(error.exception.status, 409)

    def test_audit_refusal_leaves_no_attempt_but_lost_reply_does(self):
        self.panel.sync()
        with patch.object(repository, 'task_pull_requests', return_value=self.pulls('head7')):
            with patch.object(claude, 'fire', side_effect=claude.RoutineError('The Claude routine is rate limited')):
                with self.assertRaises(TaskError) as error:
                    self.panel.stage('P0-001', 'audit')
            self.assertEqual(error.exception.status, 502)
            self.assertIn('rate limited', str(error.exception))
            self.assertEqual(self.panel.snapshot()['state']['tasks']['P0-001']['stages']['audit'], [])
            # A timeout may land after the routine accepted, so the attempt stands as unknown.
            with patch.object(claude, 'fire', side_effect=TimeoutError()):
                with self.assertRaises(TaskError):
                    self.panel.stage('P0-001', 'audit')
            with patch.object(claude, 'fire') as fire:
                with self.assertRaises(TaskError) as error:
                    self.panel.stage('P0-001', 'audit')
                fire.assert_not_called()
            self.assertIn('check the Claude routine', str(error.exception))
        attempts = self.panel.snapshot()['state']['tasks']['P0-001']['stages']['audit']
        self.assertEqual([a['status'] for a in attempts], ['submitting'])

    def test_audit_needs_the_routine_configured(self):
        self.panel.sync()
        self.panel.audit_routine_token = ''
        with patch.object(repository, 'task_pull_requests') as listing, patch.object(claude, 'fire') as fire:
            with self.assertRaises(TaskError) as error:
                self.panel.stage('P0-001', 'audit')
            listing.assert_not_called()
            fire.assert_not_called()
        self.assertEqual(error.exception.status, 503)
        self.assertIn('CLAUDE_AUDIT_ROUTINE_TOKEN', str(error.exception))

    def test_lock_and_tracked_edits_refuse_sync_while_untracked_files_are_tolerated(self):
        lock = task_state.acquire_lock(self.panel.state_file)
        try:
            with self.assertRaises(TaskError) as error:
                self.panel.sync()
            self.assertEqual(error.exception.status, 409)
        finally:
            lock.close()
        # A fast-forward pull cannot disturb an untracked file, so one does not hold up a
        # sync; the checkout accumulates them and they are still left exactly where they are.
        (self.panel.clone / 'untracked').write_text('keep me')
        self.assertEqual(self.panel.sync()['prepared'], ['P0-001'])
        self.assertEqual((self.panel.clone / 'untracked').read_text(), 'keep me')
        # An edit to a tracked file does refuse, because the pull would be the thing to lose it.
        tracked = self.panel.clone / repository.TRACKER
        tracked.write_text(tracked.read_text() + '\n')
        with self.assertRaises(TaskError) as error:
            self.panel.sync()
        self.assertEqual(error.exception.status, 409)
        self.assertTrue(tracked.read_text().endswith('\n\n'))

    def test_background_sync_returns_before_git_finishes_and_reports_failure(self):
        entered, release = threading.Event(), threading.Event()
        def slow_sync(event, only=None):
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
        def sync(event, only=None):
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
            # Preparing one task answers with a job envelope, exactly as a full sync does.
            status, data = request('/api/tasks/P0-001/prepare')
            self.assertEqual(status, 202)
            # A second request cannot coalesce into that job without losing its selection.
            self.assertEqual(request('/api/tasks/P0-002/prepare')[0], 409)
            deadline = time.monotonic() + 5
            while data['status'] == 'running' and time.monotonic() < deadline:
                time.sleep(0.01)
                status, data = request('/api/sync/' + data['id'], method='GET')
            self.assertEqual(data['status'], 'succeeded')
            self.assertEqual(data['result']['prepared'], ['P0-001'])
            self.assertEqual(request('/api/tasks/P0-999/prepare')[0], 404)
            # Expiring a stage is routed apart from expiring the task that owns it.
            self.assertEqual(request('/api/tasks/P0-001/refine/expire')[0], 409)
            with patch.object(codex, 'submit', return_value='https://chatgpt.com/codex/tasks/x'):
                self.assertEqual(request('/api/tasks/P0-001/refine')[0], 200)
            status, data = request('/api/tasks/P0-001/refine/expire')
            self.assertEqual((status, data['expired'], data['stage']), (200, 'P0-001', 'refine'))
            self.assertEqual(request('/api/tasks/P0-001/audit/expire')[0], 409)
            self.assertEqual(request('/api/tasks/P0-001/review/expire')[0], 404)
            # Refreshing the pull request is a route of its own, answered directly.
            self.pull_request.return_value = {'url': 'https://example/pr/9', 'state': 'OPEN'}
            status, data = request('/api/tasks/P0-001/pull-request')
            self.assertEqual((status, data['pull_request']), (200, 'https://example/pr/9'))
            self.assertEqual(request('/api/tasks/P0-002/pull-request')[0], 409)
            self.assertEqual(request('/api/tasks/P0-999/pull-request')[0], 404)
            with patch.object(repository, 'task_pull_requests', return_value=[]):
                self.assertEqual(request('/api/tasks/P0-001/rebase'),
                                 (200, {'task': 'P0-001', 'rebased': False}))
                self.assertEqual(request('/api/tasks/P0-002/rebase')[0], 409)
        finally:
            server.shutdown()
            server.server_close()
            thread.join()

