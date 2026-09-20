import json
import threading
import unittest
from http.client import HTTPConnection
from http.cookies import SimpleCookie
from unittest.mock import Mock, patch

from control_panel import sessions
from control_panel.api import Handler, ThreadingHTTPServer


class SessionTests(unittest.TestCase):
    def test_cookie_is_protected_and_expires_and_token_rotation_invalidates_it(self):
        with patch.object(sessions.time, 'time', return_value=100):
            header = sessions.cookie('secret-token', '/tools/panel/')
            self.assertNotIn('secret-token', header)
            jar = SimpleCookie(header)
            cookie = jar[sessions.COOKIE]
            self.assertTrue(cookie['httponly'])
            self.assertTrue(cookie['secure'])
            self.assertEqual(cookie['samesite'], 'Strict')
            self.assertEqual(cookie['path'], '/tools/panel/')
            self.assertTrue(sessions.valid('secret-token', header))
            self.assertFalse(sessions.valid('rotated-token', header))
            self.assertFalse(sessions.valid('secret-token', header.replace(cookie.value, cookie.value + '0')))
        with patch.object(sessions.time, 'time', return_value=100 + sessions.LIFETIME):
            self.assertFalse(sessions.valid('secret-token', header))
        self.assertFalse(sessions.valid('secret-token', 'malformed'))
        with self.assertRaises(ValueError):
            sessions.cookie('secret-token', '/; Domain=example.com')

    def test_login_cookie_access_csrf_guard_and_logout(self):
        scheduler = Mock()
        scheduler.snapshot.return_value = {'tracker': {'tasks': []}, 'state': {'tasks': {}}}
        handler = type('TestHandler', (Handler,), {'scheduler': scheduler, 'token': 'test-token'})
        server = ThreadingHTTPServer(('127.0.0.1', 0), handler)
        thread = threading.Thread(target=server.serve_forever, daemon=True)
        thread.start()
        def request(method, path, headers=None, data=None):
            connection = HTTPConnection(*server.server_address, timeout=3)
            connection.request(method, path, json.dumps(data or {}) if method == 'POST' else None,
                               {'Content-Type': 'application/json', **(headers or {})})
            response = connection.getresponse()
            result = response.status, json.loads(response.read()), response.getheader('Set-Cookie')
            connection.close()
            return result
        try:
            self.assertEqual(request('POST', '/api/session')[0], 401)
            status, data, cookie = request('POST', '/api/session', {'Authorization': 'Bearer test-token'}, {'path': '/panel/'})
            self.assertEqual(status, 200)
            self.assertTrue(data['authenticated'])
            self.assertNotIn('test-token', cookie)
            headers = {'Cookie': cookie.split(';')[0]}
            self.assertEqual(request('GET', '/api/tasks', headers)[0], 200)
            self.assertEqual(request('POST', '/api/sync', headers)[0], 403)
            scheduler.start_sync.assert_not_called()
            headers['X-Panel-Request'] = '1'
            scheduler.start_sync.return_value = {'id': 'job', 'status': 'running'}
            self.assertEqual(request('POST', '/api/sync', headers)[0], 202)
            self.assertEqual(request('GET', '/api/tasks', {**headers, 'Authorization': 'Bearer wrong'})[0], 401)
            status, data, cleared = request('POST', '/api/session/logout', headers, {'path': '/panel/'})
            self.assertEqual(status, 200)
            self.assertIn('Max-Age=0', cleared)
            self.assertIn('Path=/panel/', cleared)
        finally:
            server.shutdown()
            server.server_close()
            thread.join()
