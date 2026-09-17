"""HTTP API routes, authentication, and server startup."""

import argparse
import hmac
import json
import logging
import os
import re
import subprocess
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from typing import Any
from urllib.parse import urlparse

from .task_scheduler import TaskScheduler, TaskError
from .ui import render

LOGGER = logging.getLogger("control_panel")


class Handler(BaseHTTPRequestHandler):
    scheduler: TaskScheduler
    token: str

    def log_message(self, *args: Any) -> None:
        return

    def send(self, status: int, body: bytes, content_type="application/json") -> None:
        self.send_response(status)
        self.send_header("Content-Type", content_type)
        self.send_header("Content-Length", str(len(body)))
        self.send_header("Cache-Control", "no-store")
        self.send_header("X-Content-Type-Options", "nosniff")
        self.end_headers()
        self.wfile.write(body)

    def respond(self, status: int, data: dict) -> None:
        self.send(status, json.dumps(data).encode())

    def authorized(self) -> bool:
        expected = ("Bearer " + self.token).encode()
        supplied = self.headers.get("Authorization", "").encode()
        if not self.token or not hmac.compare_digest(supplied, expected):
            self.respond(401, {"error": "A valid control-panel bearer token is required"})
            return False
        return True

    def do_GET(self) -> None:
        path = urlparse(self.path).path
        if path == "/healthz":
            self.respond(200, {"status": "ok"})
            return
        if path not in ("/", "/api/tasks"):
            self.respond(404, {"error": "Not found"})
            return
        if path == "/api/tasks" and not self.authorized():
            return
        try:
            snapshot = self.scheduler.snapshot()
            if path == "/api/tasks":
                self.respond(200, snapshot)
            else:
                message = "" if self.scheduler.tracker_path.exists() else "Sync main to load the task tracker."
                page = render(snapshot["tracker"], snapshot["state"], message, False)
                self.send(200, page.encode(), "text/html; charset=utf-8")
        except (OSError, ValueError):
            self.respond(500, {"error": "Unable to read control-panel state"})

    def do_POST(self) -> None:
        if not self.authorized():
            return
        if self.headers.get_content_type() != "application/json":
            self.respond(415, {"error": "Content-Type must be application/json"})
            return
        try:
            length = int(self.headers.get("Content-Length", "0"))
            if not 0 < length <= 2_000_000:
                raise TaskError(413, "Request body must be between 1 and 2000000 bytes")
            data = json.loads(self.rfile.read(length))
            if not isinstance(data, dict):
                raise ValueError("Expected a JSON object")
            path = urlparse(self.path).path
            match = re.fullmatch(r"/api/tasks/(P\d+-\d{3})/(refine|implement|audit)", path)
            if path == "/api/events/merge":
                if not isinstance(data.get("pull_request"), dict):
                    raise ValueError("Expected a pull_request event")
                result = self.scheduler.sync(data)
            elif path == "/api/sync":
                result = self.scheduler.sync()
            elif match:
                number = data.get("pr_number")
                if number is not None and (type(number) is not int or number <= 0):
                    raise ValueError("pr_number must be a positive integer")
                result = self.scheduler.stage(*match.groups(), pr_number=number)
            else:
                raise TaskError(404, "Not found")
            self.respond(200, result)
        except TaskError as error:
            self.respond(error.status, {"error": str(error), **error.details})
        except (ValueError, UnicodeError):
            self.respond(400, {"error": "Invalid JSON request"})
        except (OSError, subprocess.SubprocessError):
            # Subprocess messages can contain credential-bearing URLs or prompt content.
            LOGGER.error("Control-panel operation failed")
            self.respond(502, {"error": "Repository operation failed; check checkout, credentials and remote availability"})


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--tracker", type=Path)
    parser.add_argument("--state-file", type=Path, required=True)
    parser.add_argument("--clone", type=Path, required=True, help="dedicated main checkout")
    parser.add_argument("--max-concurrent", type=int, default=3)
    parser.add_argument("--host", default="127.0.0.1")
    parser.add_argument("--port", type=int, default=8765)
    return parser.parse_args()


def main() -> int:
    logging.basicConfig(level=logging.INFO, format="%(levelname)s: %(message)s")
    args = parse_args()
    Handler.token = os.environ.get("PANEL_TOKEN", "")
    if not Handler.token:
        raise SystemExit("Set PANEL_TOKEN before starting the control panel")
    Handler.scheduler = TaskScheduler(
        args.clone, args.state_file,
        args.tracker or args.state_file.parent / "tracker.json",
        os.environ.get("CODEX_ENV_ID", ""), args.max_concurrent,
    )
    with ThreadingHTTPServer((args.host, args.port), Handler) as server:
        print(f"control panel on http://{args.host}:{args.port}")
        try:
            server.serve_forever()
        except KeyboardInterrupt:
            pass
    return 0
