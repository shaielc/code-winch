"""Load the stage prompts bundled with the control panel."""

from importlib.resources import files
from string import Template


def load(stage: str) -> Template:
    names = {"refine": "refine.md", "implement": "implement.md", "audit": "audit.md"}
    return Template(files(__package__).joinpath(names[stage]).read_text(encoding="utf-8"))
