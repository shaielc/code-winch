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
└── <skill-name>/
    └── SKILL.md
```

## Available skills

| Skill | Audience | Use it when |
|---|---|---|
| [`workplan`](workplan/SKILL.md) | planning agent | Creating, extending, re-deriving at a phase gate, updating, or auditing `docs/workplan/` |

Audience matters. A planning skill is for whoever writes the plan; an
implementing agent picking up a single task usually needs only the parts of it
that constrain how a task is finished, not the whole document. Say who a skill
is for when you add one.

## Getting an agent to use them

### Any agent, no setup

Point at the path:

> Follow `skills/workplan/SKILL.md` and audit the plan for coverage gaps.

This always works and needs nothing configured. It is the fallback when an
agent has no skill mechanism, or when you want a skill used once without wiring
it in permanently.

### Dispatched task agents

`scripts/task-prompt.md` instructs every dispatched implementer to read
`docs/workplan/$brief` and **all applicable `AGENTS.md` files**. No `AGENTS.md`
exists in this repository yet, so that instruction currently resolves to
nothing.

Creating a root `AGENTS.md` that points here is what wires skills into
dispatched work. It should reference the parts an implementer needs — the
invariants a finished task must preserve, what counts as a demonstration, and
the rule that nothing is deferred without an owning task ID — rather than
loading a planning skill wholesale into an implementation prompt.

### Claude Code

Symlink the skill into `.claude/skills/`. The skill itself stays here; only a
link sits in the dot-directory, and Claude Code follows it and reads `SKILL.md`
from the target.

```sh
mkdir -p .claude/skills
ln -s ../../skills/workplan .claude/skills/workplan
echo '.claude/' >> .gitignore
```

It is then invocable as `/workplan`, and Claude loads it on its own when the
work matches its description. Edits to `SKILL.md` are picked up live through the
link; creating `.claude/skills/` for the first time needs one restart before the
directory is watched.

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
