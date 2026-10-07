import unittest
from html.parser import HTMLParser

from control_panel.ui import CLAUDE_ICON, CODEX_ICON, render
from control_panel.integrations.process import failure_message
import subprocess

TASK = {'id': 'P0-001', 'title': 'Example', 'status': 'pending', 'depends_on': []}


class Elements(HTMLParser):
    """Collect the buttons, stage cells and stage icons in document order."""

    def __init__(self, page):
        super().__init__()
        self.buttons, self.cells, self.icons = [], [], []
        self.feed(page)

    def handle_starttag(self, tag, attrs):
        attributes = dict(attrs)
        if tag == 'button':
            self.buttons.append(attributes)
        elif attributes.get('class') == 'stage-cell':
            self.cells.append(attributes)
        elif attributes.get('class') == 'stage-icon':
            self.icons.append(attributes)


class UITests(unittest.TestCase):
    def actions(self, page, *names):
        return [b for b in Elements(page).buttons if b.get('data-action') in names]

    def stages(self, page):
        """Each stage control as (cell, icon). The table renders one set and the tree another."""
        parsed = Elements(page)
        self.assertEqual(len(parsed.cells), len(parsed.icons))
        return list(zip(parsed.cells, parsed.icons))

    def by_stage(self, page):
        return {cell['data-stage']: (cell, icon) for cell, icon in self.stages(page)}

    def test_each_stage_is_one_control_marked_with_the_agent_that_runs_it(self):
        page = render({'tasks': [TASK]}, {'tasks': {}}, '', False)
        controls = self.stages(page)
        self.assertEqual([cell['data-stage'] for cell, _ in controls],
                         ['refine', 'implement', 'audit'] * 2)
        self.assertEqual([cell['data-service'] for cell, _ in controls],
                         ['Codex', 'Codex', 'Claude'] * 2)
        # Codex runs two of the three stages, so its mark appears twice per view.
        self.assertEqual(page.count(CODEX_ICON), 4)
        self.assertEqual(page.count(CLAUDE_ICON), 2)
        # Stage launches are not buttons any more; only whole-task actions are.
        self.assertEqual({b['data-action'] for b in Elements(page).buttons if 'data-task' in b},
                         {'prepare'})

    def test_an_unprepared_stage_carries_its_refusal_instead_of_launching(self):
        page = render({'tasks': [TASK]}, {'tasks': {}}, '', False)
        for cell, icon in self.stages(page):
            self.assertEqual(icon['data-state'], 'launch')
            self.assertNotIn('href', icon)
            self.assertIn('Prepare this task', cell['data-reason'])
            self.assertIn('is waiting to be expired', cell['data-expire-reason'])
        self.assertTrue(all('disabled' not in b for b in self.actions(page, 'prepare')))
        self.assertLess(page.index('id="feedback"'), page.index('<table>'))

    def test_a_prepared_stage_is_launchable_and_says_what_a_click_will_do(self):
        record = {'status': 'in_progress', 'prepared': True}
        page = render({'tasks': [TASK]}, {'tasks': {'P0-001': record}}, '', False)
        for cell, icon in self.stages(page):
            self.assertEqual(cell['data-reason'], '')
            self.assertEqual(icon['data-state'], 'launch')
            self.assertIn('Launch', cell['data-launch'])
            self.assertIn(cell['data-service'], cell['data-launch'])
        # Prepared and leased: there is nothing left to prepare.
        self.assertEqual(self.actions(page, 'prepare'), [])

    def test_prepare_recovers_an_expired_reservation_and_hides_when_blocked(self):
        tasks = [TASK, {'id': 'P0-002', 'title': 'Two', 'status': 'pending',
                        'depends_on': ['P0-001']}]
        expired = {'status': 'in_progress', 'prepared': True, 'expired': True}
        page = render({'tasks': tasks}, {'tasks': {'P0-001': expired}}, '', False)
        prepare = self.actions(page, 'prepare')
        self.assertEqual([b['data-task'] for b in prepare], ['P0-001'] * 2)
        self.assertTrue(all('disabled' not in b for b in prepare))

    def test_a_submitted_stage_opens_its_conversation_and_keeps_the_earlier_ones(self):
        record = {'status': 'in_progress', 'prepared': True, 'stages': {
            'refine': [{'status': 'expired', 'head': 'a',
                        'task_url': 'https://chatgpt.com/codex/tasks/task_e_one'},
                       {'status': 'submitted', 'head': 'a',
                        'task_url': 'https://chatgpt.com/codex/cloud/tasks/task_e_two'}],
            'audit': [{'status': 'submitted', 'head': 'b',
                       'task_url': 'https://claude.ai/code/cse_01'}]}}
        page = render({'tasks': [TASK]}, {'tasks': {'P0-001': record}}, '', False)
        cells = self.by_stage(page)
        refine, refine_icon = cells['refine']
        # The live attempt is what the icon opens, pointed at the current conversation view.
        self.assertEqual(refine_icon['data-state'], 'open')
        self.assertEqual(refine_icon['href'], 'https://chatgpt.com/remote/task_e_two')
        self.assertEqual(refine['data-expire-reason'], '')
        self.assertIn('same branch revision', refine['data-expire-hint'])
        # Expiring must not lose the conversation, so the retired link stays on the page.
        self.assertIn('href="https://chatgpt.com/remote/task_e_one"', page)
        # An audit is scoped to the pull request it audited, not to the branch.
        self.assertIn('pull request head', cells['audit'][0]['data-expire-hint'])
        self.assertEqual(cells['audit'][1]['href'], 'https://claude.ai/code/cse_01')
        # A stage with no attempt has nothing to open and nothing to expire.
        implement, implement_icon = cells['implement']
        self.assertEqual(implement_icon['data-state'], 'launch')
        self.assertIn('No implement submission', implement['data-expire-reason'])

    def test_an_uncertain_submission_reads_as_unknown_rather_than_launchable(self):
        record = {'status': 'in_progress', 'prepared': True,
                  'stages': {'implement': [{'status': 'submitting', 'head': 'a'}]}}
        page = render({'tasks': [TASK]}, {'tasks': {'P0-001': record}}, '', False)
        implement, icon = self.by_stage(page)['implement']
        self.assertEqual(icon['data-state'], 'unknown')
        self.assertNotIn('href', icon)
        self.assertIn('lost the reply', icon['title'])
        # Expiring is offered, with the risk of a double submission named.
        self.assertEqual(implement['data-expire-reason'], '')
        self.assertIn('can submit it twice', implement['data-expire-hint'])

    def test_git_failure_diagnostics_do_not_echo_credentials(self):
        error = subprocess.CalledProcessError(128, ['git', 'push', 'https://SECRET@example.com'],
                                            stderr='Authentication failed for https://SECRET@example.com')
        message = failure_message(error)
        self.assertIn('git push', message)
        self.assertIn('GH_TOKEN', message)
        self.assertNotIn('SECRET', message)
