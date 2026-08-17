#!/usr/bin/env python3
"""Print a run's live event stream. Interim tool — deleted by P1-051.

The daemon's WebSocket endpoint has no client in this repository until
`winch run watch` exists, which left one of P1-050's acceptance criteria with
nothing to check it. See docs/workplan/post-mortems/2026-08-17-unverifiable-
stream-criterion.md.

This speaks just enough of RFC 6455 to subscribe and print: no dependencies, no
masking of outbound frames beyond the handshake, no continuation frames. It is a
verification aid, not a client library.

    ./tmp/stream.py <run-id> [seconds] [after-sequence]
"""
import base64
import json
import os
import socket
import struct
import sys
import time

BASE = os.environ.get("WINCH_ENDPOINT", "http://localhost:8080")
TOKEN = os.environ.get("WINCH_TOKEN", "local-development-session-token-0000000000")
ORIGIN = os.environ.get("WINCH_ALLOWED_ORIGIN", "http://localhost:8080")


def main(argv):
    if len(argv) < 2:
        sys.exit(__doc__)
    run_id = argv[1]
    deadline = time.time() + float(argv[2] if len(argv) > 2 else 30)
    after = argv[3] if len(argv) > 3 else "0"

    host, port = hostport(BASE)
    sock = socket.create_connection((host, port), timeout=5)
    handshake(sock, host, port, run_id, after)
    reader = frames(sock, deadline)
    for message in reader:
        report(message)


def hostport(base):
    rest = base.split("://", 1)[-1]
    host, _, port = rest.partition(":")
    return host, int(port or 80)


def handshake(sock, host, port, run_id, after):
    key = base64.b64encode(os.urandom(16)).decode()
    request = (
        f"GET /api/v1/runs/{run_id}/events/stream?after_sequence={after} HTTP/1.1\r\n"
        f"Host: {host}:{port}\r\n"
        "Upgrade: websocket\r\nConnection: Upgrade\r\n"
        f"Sec-WebSocket-Key: {key}\r\nSec-WebSocket-Version: 13\r\n"
        f"Origin: {ORIGIN}\r\n"
        f"Authorization: Bearer {TOKEN}\r\n\r\n"
    )
    sock.sendall(request.encode())
    head = b""
    while b"\r\n\r\n" not in head:
        chunk = sock.recv(4096)
        if not chunk:
            sys.exit("connection closed during handshake: " + head.decode(errors="replace"))
        head += chunk
    status = head.split(b"\r\n", 1)[0].decode()
    if "101" not in status:
        sys.exit(status + "\n" + head.decode(errors="replace"))
    print(status)
    # Anything after the header belongs to the first frames.
    return head.split(b"\r\n\r\n", 1)[1]


def frames(sock, deadline):
    """Yield decoded text messages until the deadline or a close frame."""
    buffered = b""

    def read(n):
        nonlocal buffered
        while len(buffered) < n:
            remaining = deadline - time.time()
            if remaining <= 0:
                raise TimeoutError
            sock.settimeout(remaining)
            chunk = sock.recv(65536)
            if not chunk:
                raise EOFError
            buffered += chunk
        head, buffered = buffered[:n], buffered[n:]
        return head

    while time.time() < deadline:
        try:
            first, second = read(2)
        except (EOFError, TimeoutError, socket.timeout):
            return
        opcode, length = first & 0x0F, second & 0x7F
        if length == 126:
            length = struct.unpack(">H", read(2))[0]
        elif length == 127:
            length = struct.unpack(">Q", read(8))[0]
        payload = read(length) if length else b""
        if opcode == 0x8:
            print("close")
            return
        if opcode == 0x1:
            yield json.loads(payload)


def report(message):
    kind = message.get("type")
    if kind != "event":
        print(f"{kind} lastSequence={message.get('lastSequence')}")
        return
    event = message["event"]
    print(f"{event['sequence']:>4}  {event['kind']:<16} {json.dumps(event['payload'])[:96]}")


if __name__ == "__main__":
    main(sys.argv)
