import unittest
from html.parser import HTMLParser

from control_panel.ui import render
from control_panel.integrations.process import failure_message
import subprocess


class Buttons(HTMLParser):
    def __init__(self, page):
        super().__init__()
        self.buttons = []
        self.feed(page)

    def handle_starttag(self, tag, attrs):
        if tag == 'button':
            self.buttons.append(dict(attrs))


class UITests(unittest.TestCase):
    def actions(self, page, *names):
        return [b for b in Buttons(page).buttons if b.get('data-action') in names]

    def test_stages_stay_visible_and_explain_unprepared_tasks(self):
        task = {'id': 'P0-001', 'title': 'Example', 'status': 'pending', 'depends_on': []}
        page = render({'tasks': [task]}, {'tasks': {}}, '', False)
        stages = [b for b in Buttons(page).buttons if b.get('data-task') == 'P0-001']
        self.assertEqual([b['data-action'] for b in stages],
                         ['refine', 'implement', 'audit', 'prepare'] * 2)
        self.assertTrue(all('disabled' in b and 'Prepare this task' in b['title']
                            for b in self.actions(page, 'refine', 'implement', 'audit')))
        self.assertTrue(all('disabled' not in b for b in self.actions(page, 'prepare')))
        self.assertLess(page.index('id="feedback"'), page.index('<table>'))
        record = {'status': 'in_progress', 'prepared': True}
        page = render({'tasks': [task]}, {'tasks': {'P0-001': record}}, '', False)
        self.assertTrue(all('disabled' not in b for b in
                            self.actions(page, 'refine', 'implement', 'audit')))
        # Prepared and leased: there is nothing left to prepare.
        self.assertEqual(self.actions(page, 'prepare'), [])

    def test_prepare_recovers_an_expired_reservation_and_hides_when_blocked(self):
        tasks = [{'id': 'P0-001', 'title': 'One', 'status': 'pending', 'depends_on': []},
                 {'id': 'P0-002', 'title': 'Two', 'status': 'pending', 'depends_on': ['P0-001']}]
        expired = {'status': 'in_progress', 'prepared': True, 'expired': True}
        page = render({'tasks': tasks}, {'tasks': {'P0-001': expired}}, '', False)
        prepare = self.actions(page, 'prepare')
        self.assertEqual([b['data-task'] for b in prepare], ['P0-001'] * 2)
        self.assertTrue(all('disabled' not in b for b in prepare))

    def test_stage_menu_offers_expiry_only_for_a_standing_submission(self):
        task = {'id': 'P0-001', 'title': 'Example', 'status': 'pending', 'depends_on': []}
        record = {'status': 'in_progress', 'prepared': True, 'stages': {
            'refine': [{'status': 'expired', 'head': 'a', 'task_url': 'https://codex/one'},
                       {'status': 'submitted', 'head': 'a', 'task_url': 'https://codex/two'}],
            'implement': [{'status': 'submitting', 'head': 'a'}]}}
        page = render({'tasks': [task]}, {'tasks': {'P0-001': record}}, '', False)
        menu = {b['data-action']: b for b in self.actions(page, 'refine/expire', 'implement/expire')}
        self.assertEqual(sorted(menu), ['implement/expire', 'refine/expire'])
        self.assertNotIn('disabled', menu['refine/expire'])
        self.assertIn('same branch revision', menu['refine/expire']['title'])
        # An unknown outcome is named in the label, not buried in a tooltip.
        self.assertIn('outcome unknown', page)
        self.assertIn('can submit it twice', menu['implement/expire']['title'])
        # The newest conversation is linked; the retired one is not lost from the record.
        self.assertIn('href="https://codex/two" data-stage="refine"', page)
        # Nothing submitted yet means no menu at all.
        plain = render({'tasks': [task]}, {'tasks': {'P0-001': {'status': 'in_progress',
                                                               'prepared': True}}}, '', False)
        self.assertEqual(self.actions(plain, 'refine/expire', 'implement/expire'), [])
        self.assertNotIn('<details class="menu">', plain)

    def test_git_failure_diagnostics_do_not_echo_credentials(self):
        error = subprocess.CalledProcessError(128, ['git', 'push', 'https://SECRET@example.com'],
                                            stderr='Authentication failed for https://SECRET@example.com')
        message = failure_message(error)
        self.assertIn('git push', message)
        self.assertIn('GH_TOKEN', message)
        self.assertNotIn('SECRET', message)
