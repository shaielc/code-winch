"""What the panel writes to its log: the tools it drives, and why a call failed.

The panel is a thin shell around other people's CLIs, so its log is the only place their
output lands. That output is not ours to echo verbatim: everything here passes through
redact() first, and is kept as a clipped tail rather than a transcript. A command is named
by its arguments, with anything prompt-sized stood in for by its length.
"""

import logging
import os
import re
import traceback
from collections.abc import Iterable

LOGGER = logging.getLogger("control_panel")

# The panel's own secrets, scrubbed out of whatever a tool prints back at us.
SECRET_VARS = ("GH_TOKEN", "PANEL_TOKEN", "CLAUDE_AUDIT_ROUTINE_TOKEN", "RUNNER_TOKEN")
# Credentials the panel never held but a tool may still echo: a URL's userinfo, and the
# token shapes GitHub and Anthropic hand out.
CREDENTIAL = re.compile(r"(?<=://)[^/\s:@]+(?::[^/\s@]*)?(?=@)"
                        r"|\b(?:gh[pousr]_|github_pat_|sk-ant-)[\w\-]{8,}")
PLACEHOLDER = "[redacted]"
TAIL = 1500  # characters of a tool's output worth keeping
ARGUMENT = 80  # a longer argument is a prompt, not an argument


def redact(text: object, *also: str) -> str:
    """Replace every secret this process knows about, plus any the caller names.

    A token the panel was handed as an argument rather than read from the environment —
    the audit routine's, during a fire — has to be named, so `also` takes it.
    """
    out = str(text if text is not None else "")
    values = [os.environ.get(name, "") for name in SECRET_VARS] + list(also)
    for value in values:
        # A short value would redact half the line on a coincidental match.
        if value and len(value) >= 8:
            out = out.replace(value, PLACEHOLDER)
    return CREDENTIAL.sub(PLACEHOLDER, out)


def tail(text: object, *also: str) -> str:
    """Redact, then keep the end: a failing CLI explains itself in its last lines."""
    out = redact(text, *also).strip()
    return out if len(out) <= TAIL else "[clipped] " + out[-TAIL:]


def command(argv: Iterable[str]) -> str:
    """Name a command without pasting the prompt it carried into the log."""
    return redact(" ".join(
        arg if len(arg) <= ARGUMENT else f"<{len(arg)} chars>" for arg in argv))


def _line(label: str, text: object) -> str:
    body = tail(text)
    if not body:
        return label
    return f"{label}: {body}" if "\n" not in body else f"{label}:\n{body}"


def output(label: str, text: object = "") -> None:
    """Record what a tool produced, which is the panel's only account of what it did."""
    LOGGER.info("%s", _line(label, text))


def failure(label: str, text: object = "") -> None:
    LOGGER.error("%s", _line(label, text))


def unexpected(label: str) -> None:
    """Report a bug in the panel itself, with the traceback the operator needs."""
    LOGGER.error("%s", _line(label, traceback.format_exc()))
