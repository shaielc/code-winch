"""Fire a Claude Code routine and return the cloud session it starts.

The CLI cannot start a cloud session without a terminal, so the panel calls the routine's
API trigger instead. The routine owns the repository and its own prompt; the panel only
supplies the per-call text, which reaches the session as an untrusted payload.
"""

import json
import urllib.error
import urllib.request
from urllib.parse import urlparse

from .. import logs

API_VERSION = "2023-06-01"
BODY = 4096  # enough of a reply to explain a refusal


class RoutineError(ValueError):
    """The routine refused the call, so no session started; the message is safe to show.

    An accepted call whose reply cannot be read raises a plain ValueError instead: a
    session may exist, so the caller must treat the outcome as unknown.
    """


def fire(url: str, token: str, text: str, timeout: float = 30) -> str:
    request = urllib.request.Request(
        url,
        data=json.dumps({"text": text}).encode(),
        method="POST",
        headers={
            "Authorization": "Bearer " + token,
            "anthropic-version": API_VERSION,
            "Content-Type": "application/json",
        },
    )
    try:
        with urllib.request.urlopen(request, timeout=timeout) as response:
            payload = response.read()
    except urllib.error.HTTPError as error:
        logs.failure(f"claude routine refused the fire with HTTP {error.code}",
                     logs.redact(_body(error), token))
        if error.code in (401, 403):
            raise RoutineError("The Claude routine refused the token; check CLAUDE_AUDIT_ROUTINE_TOKEN") from None
        if error.code == 404:
            raise RoutineError("The Claude routine was not found; check CLAUDE_AUDIT_ROUTINE_URL") from None
        if error.code == 429:
            wait = error.headers.get("Retry-After")
            raise RoutineError("The Claude routine is rate limited"
                               + (f"; retry after {wait} seconds" if wait else "")) from None
        raise RoutineError(f"The Claude routine returned HTTP {error.code}") from None
    try:
        body = json.loads(payload)
    except json.JSONDecodeError:
        logs.failure("claude routine returned an unreadable body",
                     logs.redact(payload[:BODY].decode(errors="replace"), token))
        raise ValueError("The Claude routine returned an unreadable response") from None
    session = body.get("claude_code_session_url") if isinstance(body, dict) else None
    parsed = urlparse(str(session or ""))
    if parsed.scheme != "https" or not parsed.netloc:
        logs.failure("claude routine accepted the fire without a session URL",
                     logs.redact(payload[:BODY].decode(errors="replace"), token))
        raise ValueError("The Claude routine returned no session URL")
    logs.output("claude routine started", session)
    return session


def _body(error: urllib.error.HTTPError) -> str:
    try:
        return error.read(BODY).decode(errors="replace")
    except OSError:
        return ""
