# Skills

A skill is a markdown document that tells an agent how to do one kind of task
well. It is plain text with YAML frontmatter — no runtime, no installation, no
tool-specific format — so any agent can use one by reading it.

They live here rather than in a tool's dot-directory so they are versioned with
the project, reviewed in pull requests, and usable by every agent that works on
this repository regardless of which one it is.

```
skills/
├── README.md
├── <shared-reference>.md
└── <skill-name>/
    └── SKILL.md
```

A shared reference at the top level is not a skill and is never loaded on its
own. It exists when two skills act on the same thing and would otherwise drift
apart; each links to it and neither repeats it.

## Available skills

| Skill | Audience | Use it when |
|---|---|---|
| [`workplan`](workplan/SKILL.md) | planning agent | Creating, extending, re-deriving at a phase gate, updating, or auditing `docs/workplan/` as a whole |
| [`task`](task/SKILL.md) | implementing or reviewing agent | Implementing one brief, verifying it, or judging whether one task is done |

Both rest on [`workplan-model.md`](workplan-model.md): what a workplan is, the
seven invariants, the four task shapes, the tracker's shape.

Audience matters, and here it is the whole point of the split. Changing the plan
and doing one task are different jobs with different boundaries: the planning
skill is answerable for the plan's coverage, width, and graph, while the task
skill is answerable for one brief and the change that claims to satisfy it. A
single document covering both invites a task audit to report the plan's debt
against whichever brief is in flight. Say who a skill is for when you add one.

## Getting an agent to use them

### Any agent, no setup

Point at the path:

> Follow `skills/workplan/SKILL.md` and audit the plan for coverage gaps.

This always works and needs nothing configured. It is the fallback when an
agent has no skill mechanism, or when you want a skill used once without wiring
it in permanently.

### Dispatched task agents

`scripts/task-prompt.md` instructs every dispatched implementer to read
[`task/SKILL.md`](task/SKILL.md), then `docs/workplan/$brief`, then **all
applicable `AGENTS.md` files**. That is how a skill reaches dispatched work: the
prompt names it, and the root `AGENTS.md` carries the rules binding every change
in the repository.

`task/SKILL.md` is the skill for that audience — how to locate a brief from an
ID, which of its sections bind, what a single task is and is not answerable for,
and how to judge it done. Name it in a dispatch prompt rather than loading the
planning skill wholesale into an implementation.

### Claude Code

Symlink the skill into `.claude/skills/`. The skill itself stays here; only a
link sits in the dot-directory, and Claude Code follows it and reads `SKILL.md`
from the target.

```sh
mkdir -p .claude/skills
ln -s ../../skills/workplan .claude/skills/workplan
ln -s ../../skills/task .claude/skills/task
echo '.claude/' >> .gitignore
```

They are then invocable as `/workplan` and `/task`, and Claude loads one on its
own when the work matches its description. Edits to `SKILL.md` are picked up
live through the link; creating `.claude/skills/` for the first time needs one
restart before the directory is watched.

Add one `ln -s` line per skill.

## Writing a new skill

One directory per skill, `SKILL.md` inside, directory name matching the
frontmatter `name`. Supporting files — templates, reference material — sit
alongside it and are linked from the body.

```markdown
---
name: <kebab-case, matches the directory>
description: <what it does, and when an agent should reach for it>
---

# <Title>
```

The `description` is the only part an agent sees before deciding whether to
load the rest, so it carries the triggering conditions, not just a topic.

Write the body as instructions, not as an essay about the domain: what to do,
what to check, what to produce, and what "done" looks like. Prefer templates,
procedures, and checklists over explanation. If a paragraph does not change
what an agent does, cut it.
