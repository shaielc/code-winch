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
    def test_stages_stay_visible_and_explain_unprepared_tasks(self):
        task = {'id': 'P0-001', 'title': 'Example', 'status': 'pending', 'depends_on': []}
        page = render({'tasks': [task]}, {'tasks': {}}, '', False)
        stages = [b for b in Buttons(page).buttons if b.get('data-task') == 'P0-001']
        self.assertEqual([b['data-action'] for b in stages], ['refine', 'implement', 'audit'] * 2)
        self.assertTrue(all('disabled' in b and 'Sync main' in b['title'] for b in stages))
        self.assertLess(page.index('id="feedback"'), page.index('<table>'))
        record = {'status': 'in_progress', 'prepared': True}
        page = render({'tasks': [task]}, {'tasks': {'P0-001': record}}, '', False)
        stages = [b for b in Buttons(page).buttons if b.get('data-action') in ('refine', 'implement', 'audit')]
        self.assertTrue(all('disabled' not in b for b in stages))

    def test_git_failure_diagnostics_do_not_echo_credentials(self):
        error = subprocess.CalledProcessError(128, ['git', 'push', 'https://SECRET@example.com'],
                                            stderr='Authentication failed for https://SECRET@example.com')
        message = failure_message(error)
        self.assertIn('git push', message)
        self.assertIn('GH_TOKEN', message)
        self.assertNotIn('SECRET', message)
