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
├── shared/
│   └── <reference>.md
└── <skill-name>/
    └── SKILL.md
```

## Available skills

| Skill | Audience | Use it when |
|---|---|---|
| [`workplan`](workplan/SKILL.md) | planning agent | Creating, extending, re-deriving at a phase gate, updating, or auditing the plan in `docs/workplan/` |
| [`task`](task/SKILL.md) | implementing agent | Implementing, finishing, reviewing, or auditing one task from `docs/workplan/` |

Audience matters. Say who a skill is for when you add one: an agent should be
able to tell from the table whether a skill is addressed to it before loading
anything.

## Shared reference material

`shared/` holds documents that more than one skill depends on. A skill links to
what it needs and does not restate it.

| Document | Defines |
|---|---|
| [`shared/workplan-model.md`](shared/workplan-model.md) | The workplan's layout, its seven invariants, the four task shapes, dependency-edge reasons, owned surfaces, brief anatomy, and the frozen tracker schema |

## Getting an agent to use them

### Any agent, no setup

Point at the path:

> Follow `skills/workplan/SKILL.md` and audit the plan for coverage gaps.

> Follow `skills/task/SKILL.md` and implement P1-054.

This always works and needs nothing configured. It is the fallback when an
agent has no skill mechanism, or when you want a skill used once without wiring
it in permanently.

### Dispatched task agents

`scripts/task-prompt.md` instructs every dispatched implementer to read its
brief, `skills/task/SKILL.md`, and all applicable `AGENTS.md` files. The root
`AGENTS.md` points here too, so an agent that lands in the repository by any
other route finds the same instructions.

Keep that wiring narrow. A dispatched implementer is pointed at the skill
addressed to it, not at every skill in this directory.

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

Add one `ln -s` line per skill. `shared/` is not symlinked — skills reference it
by repository path.

## Writing a new skill

One directory per skill, `SKILL.md` inside, directory name matching the
frontmatter `name`. Supporting files — templates, reference material — sit
alongside it and are linked from the body; material two skills both need goes in
`shared/` instead.

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

Each skill describes its own job in the affirmative. A skill that defines itself
by what a neighbouring skill covers instead is unreadable on its own and stales
the moment that neighbour changes.
