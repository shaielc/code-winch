import json
import logging
import socket
import threading
import unittest
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

from control_panel.integrations import claude

SESSION = "https://claude.ai/code/session_01ABC"


class Routine(BaseHTTPRequestHandler):
    """Answer each fire with the next scripted (status, body, headers) reply."""
    replies: list
    received: list

    def log_message(self, *args):
        return

    def do_POST(self):
        body = self.rfile.read(int(self.headers["Content-Length"]))
        headers = {name.lower(): value for name, value in self.headers.items()}
        self.received.append({"path": self.path, "headers": headers, "body": json.loads(body)})
        status, reply, headers = self.replies.pop(0)
        payload = reply if isinstance(reply, bytes) else json.dumps(reply).encode()
        self.send_response(status)
        for name, value in headers.items():
            self.send_header(name, value)
        self.send_header("Content-Length", str(len(payload)))
        self.end_headers()
        self.wfile.write(payload)


class ClaudeRoutineTests(unittest.TestCase):
    def setUp(self):
        silence, panel = logging.NullHandler(), logging.getLogger("control_panel")
        panel.addHandler(silence)
        self.addCleanup(panel.removeHandler, silence)
        self.handler = type("Handler", (Routine,), {"replies": [], "received": []})
        self.server = ThreadingHTTPServer(("127.0.0.1", 0), self.handler)
        threading.Thread(target=self.server.serve_forever, daemon=True).start()
        self.url = "http://%s:%d/v1/claude_code/routines/trig_1/fire" % self.server.server_address

    def tearDown(self):
        self.server.shutdown()
        self.server.server_close()

    def fire(self, *replies):
        self.handler.replies.extend(replies)
        return claude.fire(self.url, "sk-secret", "Audit P0-001", timeout=5)

    def test_fire_sends_text_with_bearer_and_returns_the_session(self):
        reply = {"type": "routine_fire", "claude_code_session_id": "session_01ABC",
                 "claude_code_session_url": SESSION}
        self.assertEqual(self.fire((200, reply, {})), SESSION)
        sent = self.handler.received[0]
        self.assertEqual(sent["path"], "/v1/claude_code/routines/trig_1/fire")
        self.assertEqual(sent["body"], {"text": "Audit P0-001"})
        self.assertEqual(sent["headers"]["authorization"], "Bearer sk-secret")
        self.assertEqual(sent["headers"]["anthropic-version"], claude.API_VERSION)

    def test_refusals_name_the_cause_without_echoing_the_token_or_body(self):
        cases = [
            ((401, {"error": "sk-secret is invalid"}, {}), "CLAUDE_AUDIT_ROUTINE_TOKEN"),
            ((404, {}, {}), "CLAUDE_AUDIT_ROUTINE_URL"),
            ((429, {}, {"Retry-After": "120"}), "retry after 120 seconds"),
            ((500, {"error": "boom"}, {}), "HTTP 500"),
        ]
        for reply, expected in cases:
            with self.assertRaises(claude.RoutineError) as error:
                self.fire(reply)
            self.assertIn(expected, str(error.exception))
            self.assertNotIn("sk-secret", str(error.exception))

    def test_an_accepted_call_without_a_session_is_an_unknown_outcome(self):
        # A 200 may have started a session, so it must not read as a refusal.
        for reply in ((200, {"type": "routine_fire"}, {}),
                      (200, {"claude_code_session_url": "javascript:alert(1)"}, {}),
                      (200, b"<html>proxy</html>", {})):
            with self.assertRaises(ValueError) as error:
                self.fire(reply)
            self.assertNotIsInstance(error.exception, claude.RoutineError)

    def test_a_refusal_logs_the_body_without_the_fired_token(self):
        with self.assertLogs("control_panel", "ERROR") as captured:
            with self.assertRaises(claude.RoutineError):
                self.fire((401, {"error": "sk-secret is invalid"}, {}))
        logged = captured.output[0]
        self.assertIn("HTTP 401", logged)
        self.assertIn("is invalid", logged)
        self.assertNotIn("sk-secret", logged)

    def test_a_started_session_is_logged(self):
        with self.assertLogs("control_panel", "INFO") as captured:
            self.fire((200, {"claude_code_session_url": SESSION}, {}))
        self.assertIn(SESSION, captured.output[0])

    def test_an_accepted_call_without_a_session_logs_the_reply(self):
        with self.assertLogs("control_panel", "ERROR") as captured:
            with self.assertRaises(ValueError):
                self.fire((200, {"type": "routine_fire"}, {}))
        self.assertIn("routine_fire", captured.output[0])

    def test_an_unreachable_routine_is_a_transport_failure(self):
        with socket.socket() as probe:
            probe.bind(("127.0.0.1", 0))
            closed = "http://%s:%d/fire" % probe.getsockname()
        with self.assertRaises(OSError):
            claude.fire(closed, "sk-secret", "Audit", timeout=2)
