"""Run bounded CLI operations without a shell."""

import subprocess
from pathlib import Path


def run(*command: str, cwd: Path, capture: bool = True) -> str:
    result = subprocess.run(
        command,
        cwd=cwd,
        check=True,
        text=True,
        stdout=subprocess.PIPE if capture else None,
        stderr=subprocess.PIPE if capture else None,
        timeout=300,
    )
    return result.stdout.strip() if capture else ""
