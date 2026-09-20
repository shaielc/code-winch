"""Signed browser sessions; the API token never goes into the cookie."""

import hashlib
import hmac
import re
import secrets
import time
from http.cookies import CookieError, SimpleCookie

COOKIE = "winch_panel_session"
LIFETIME = 7 * 24 * 60 * 60


def _signature(token: str, payload: str) -> str:
    return hmac.new(token.encode(), ("panel-session:" + payload).encode(), hashlib.sha256).hexdigest()


def cookie(token: str, path: str, *, clear: bool = False) -> str:
    if not isinstance(path, str) or not re.fullmatch(r"/(?:[A-Za-z0-9_./%~-]*/)?", path):
        raise ValueError("Invalid cookie path")
    payload = f"{int(time.time()) + LIFETIME}.{secrets.token_hex(16)}"
    value = "" if clear else f"{payload}.{_signature(token, payload)}"
    jar = SimpleCookie()
    jar[COOKIE] = value
    jar[COOKIE]["path"] = path
    jar[COOKIE]["max-age"] = 0 if clear else LIFETIME
    jar[COOKIE]["httponly"] = True
    jar[COOKIE]["secure"] = True
    jar[COOKIE]["samesite"] = "Strict"
    return jar[COOKIE].OutputString()


def valid(token: str, header: str) -> bool:
    if not token:
        return False
    try:
        jar = SimpleCookie()
        jar.load(header)
        value = jar[COOKIE].value
        expiry, nonce, signature = value.split(".")
        if not re.fullmatch(r"[0-9a-f]{32}", nonce) or not re.fullmatch(r"[0-9a-f]{64}", signature):
            return False
        now = int(time.time())
        return (now < int(expiry) <= now + LIFETIME and
                hmac.compare_digest(signature, _signature(token, f"{expiry}.{nonce}")))
    except (CookieError, KeyError, ValueError):
        return False
