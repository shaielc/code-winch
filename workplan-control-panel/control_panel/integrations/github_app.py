"""Mint the installation access token the panel's Git and gh calls authenticate with.

The panel holds no long-lived GitHub credential. It holds its App's private key, signs a
short-lived App JWT with it, asks GitHub which installation covers the repository named by
GITHUB_URL, and exchanges that for an installation access token. Those last about an hour,
so a token is minted when one is wanted and close to expiring, never once at startup.

Deriving the installation from GITHUB_URL rather than configuring an installation id keeps
one less value in the deployment and makes the first call prove the App is installed where
the panel thinks it is.

The private key is never read into this process: openssl signs from the file, so the PEM
stays in its mount and cannot reach a log, a traceback, or a command's arguments. Nothing
raised from here carries a key, a JWT or a token either, because these messages are shown
to the operator.
"""

import base64
import json
import subprocess
import threading
import time
import urllib.error
import urllib.request
from datetime import UTC, datetime, timedelta
from pathlib import Path
from urllib.parse import urlparse

from .. import logs

ACCEPT = "application/vnd.github+json"
API_VERSION = "2022-11-28"
AGENT = "code-winch-control-panel"
JWT_LIFETIME = 540  # GitHub refuses an App JWT claiming more than ten minutes
CLOCK_SKEW = 60  # backdate `iat`, which GitHub rejects if our clock runs fast
MARGIN = timedelta(minutes=5)  # mint again this long before GitHub's expiry
UNKNOWN_EXPIRY = timedelta(minutes=10)  # how long to trust a reply we cannot read an expiry from
TIMEOUT = 15
SIGN_TIMEOUT = 15
REPLY = 1 << 20  # an installation carries its whole account, so a reply is not small
BODY = 4096  # enough of a refusal's body to explain it


class ConfigurationError(Exception):
    """The panel cannot authenticate as its App, and the message says what to fix.

    It is shown to the operator, so it names configuration and never a credential.
    """


class _Missing(ConfigurationError):
    """GitHub has no such installation: the App is not installed, or no longer is."""


def now() -> datetime:
    return datetime.now(UTC)


def repository_path(url: str) -> tuple[str, str]:
    """The owner and repository GITHUB_URL names, which is what the installation follows."""
    parsed = urlparse(url)
    if parsed.scheme not in ("http", "https") or not parsed.netloc:
        raise ConfigurationError(
            "Set GITHUB_URL to the repository's HTTPS URL, as https://host/owner/repository")
    path = parsed.path.strip("/")
    path = path[:-4] if path.endswith(".git") else path
    owner, _, repository = path.partition("/")
    if not owner or not repository or "/" in repository:
        raise ConfigurationError(
            "Set GITHUB_URL to one repository, as https://host/owner/repository")
    return owner, repository


def host(url: str) -> str:
    """The host GITHUB_URL names, which Git's credential helper is configured for."""
    repository_path(url)  # a URL naming no repository is a configuration error, not a host
    return urlparse(url).netloc


def api_base(url: str) -> str:
    """Where the REST API lives for the host GITHUB_URL names."""
    host = urlparse(url).netloc
    return "https://api.github.com" if host in ("github.com", "www.github.com") \
        else f"https://{host}/api/v3"


def _segment(claims: dict) -> bytes:
    """One base64url JWT segment, which carries no padding."""
    packed = json.dumps(claims, separators=(",", ":"), sort_keys=True).encode()
    return base64.urlsafe_b64encode(packed).rstrip(b"=")


class GitHubAppAuth:
    """A current installation access token, minted on demand and shared between threads.

    The panel serves HTTP on a thread per request and syncs on a background thread, so
    `token()` holds its lock across the mint: callers that arrive during one wait for it and
    then reuse its result, rather than each asking GitHub for a token of its own.
    """

    def __init__(self, app_id: str, private_key: Path, repository_url: str, api: str = ""):
        if not app_id:
            raise ConfigurationError("Set GITHUB_APP_ID to the panel's GitHub App id")
        self.app_id = app_id
        self.private_key = private_key
        self.owner, self.repository = repository_path(repository_url)
        self.api = (api or api_base(repository_url)).rstrip("/")
        self._lock = threading.Lock()
        self._token = ""
        self._mint_again_at = now()
        self._installation: int | None = None
        if not private_key.is_file():
            raise ConfigurationError(
                "GITHUB_APP_PRIVATE_KEY_FILE does not name a readable file; mount the "
                "GitHub App's PEM private key there")

    def token(self) -> str:
        """The installation token, minted unless the one in hand is good for a while yet."""
        with self._lock:
            if self._token and now() < self._mint_again_at:
                return self._token
            self._token, self._mint_again_at = self._mint()
            return self._token

    def _mint(self) -> tuple[str, datetime]:
        jwt = self._jwt()
        if self._installation is None:
            self._installation = self._resolve_installation(jwt)
        try:
            minted = self._create_token(jwt)
        except _Missing:
            # A reinstalled App has a new installation, so look again before giving up; if it
            # is gone altogether, resolving says so, which is the more useful message.
            self._installation = self._resolve_installation(jwt)
            minted = self._create_token(jwt)
        token = minted.get("token")
        if not isinstance(token, str) or not token:
            raise ConfigurationError(
                "GitHub returned no installation access token for the control panel's App")
        expires_at = self._expiry(minted.get("expires_at"))
        logs.output(f"github app token for installation {self._installation} "
                    f"expires at {expires_at.isoformat(timespec='seconds')}")
        return token, expires_at - MARGIN

    def _expiry(self, expires_at: object) -> datetime:
        try:
            parsed = datetime.fromisoformat(str(expires_at))
        except (TypeError, ValueError):
            # Hold an unreadable reply briefly rather than trusting it for GitHub's full hour.
            logs.failure("github app token came back without a readable expiry")
            return now() + UNKNOWN_EXPIRY
        return parsed if parsed.tzinfo else parsed.replace(tzinfo=UTC)

    def _resolve_installation(self, jwt: str) -> int:
        installation = self._call(
            f"/repos/{self.owner}/{self.repository}/installation", jwt,
            missing=f"The control panel's GitHub App is not installed on "
                    f"{self.owner}/{self.repository}; install it there with Contents and "
                    f"Pull requests write permission")
        identifier = installation.get("id")
        if not isinstance(identifier, int):
            raise ConfigurationError(
                f"GitHub named no installation of the control panel's App on "
                f"{self.owner}/{self.repository}")
        return identifier

    def _create_token(self, jwt: str) -> dict:
        return self._call(f"/app/installations/{self._installation}/access_tokens", jwt,
                          method="POST",
                          missing="The control panel's GitHub App installation no longer exists")

    def _call(self, path: str, jwt: str, method: str = "GET", missing: str = "") -> dict:
        request = urllib.request.Request(
            self.api + path,
            method=method,
            headers={
                "Authorization": "Bearer " + jwt,
                "Accept": ACCEPT,
                "X-GitHub-Api-Version": API_VERSION,
                "User-Agent": AGENT,
            },
        )
        try:
            with urllib.request.urlopen(request, timeout=TIMEOUT) as response:
                payload = response.read(REPLY)
        except urllib.error.HTTPError as error:
            logs.failure(f"github refused {method} {path} with HTTP {error.code}", _body(error))
            if error.code == 404:
                raise _Missing(missing or "GitHub did not find the App installation") from None
            if error.code in (401, 403):
                raise ConfigurationError(
                    "GitHub rejected the control panel's App credentials; check GITHUB_APP_ID, "
                    "the mounted private key, and that the installation still holds its "
                    "Contents and Pull requests write permission") from None
            raise ConfigurationError(
                f"GitHub answered the control panel's App with HTTP {error.code}") from None
        try:
            body = json.loads(payload)
        except json.JSONDecodeError:
            logs.failure(f"github answered {method} {path} unreadably",
                         payload[:BODY].decode(errors="replace"))
            raise ConfigurationError("GitHub's reply to the control panel's App was unreadable") from None
        if not isinstance(body, dict):
            raise ConfigurationError("GitHub's reply to the control panel's App was unreadable")
        return body

    def _jwt(self) -> str:
        issued = int(time.time())
        header = _segment({"alg": "RS256", "typ": "JWT"})
        claims = _segment({"iat": issued - CLOCK_SKEW, "exp": issued + JWT_LIFETIME,
                           "iss": self.app_id})
        signing_input = header + b"." + claims
        signature = base64.urlsafe_b64encode(self._sign(signing_input)).rstrip(b"=")
        return (signing_input + b"." + signature).decode()

    def _sign(self, signing_input: bytes) -> bytes:
        """RS256 over the JWT's signing input, which is what `openssl dgst -sign` produces.

        openssl reads the key from its mount, so the PEM never enters this process, and the
        signing input arrives on stdin rather than as an argument.
        """
        try:
            signed = subprocess.run(
                ("openssl", "dgst", "-sha256", "-sign", str(self.private_key)),
                input=signing_input,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                check=True,
                timeout=SIGN_TIMEOUT,
            )
        except (OSError, subprocess.SubprocessError) as error:
            # openssl reports the key's path and what is wrong with it, never its contents.
            detail = getattr(error, "stderr", b"") or b""
            logs.failure("signing the GitHub App token request failed",
                         detail.decode(errors="replace") if isinstance(detail, bytes) else detail)
            raise ConfigurationError(
                "The control panel could not sign with GITHUB_APP_PRIVATE_KEY_FILE; check that "
                "the mounted file is the App's PEM private key and is readable") from None
        return signed.stdout


def _body(error: urllib.error.HTTPError) -> str:
    try:
        return error.read(BODY).decode(errors="replace")
    except OSError:
        return ""
