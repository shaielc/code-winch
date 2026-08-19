# Skills

A skill is a markdown document that tells an agent how to do one kind of task well. It is plain text with YAML frontmatter, versioned with the repository, and usable by any agent that can read it.

```text
skills/
├── README.md
├── task/
│   └── SKILL.md
└── workplan/
    ├── SKILL.md
    └── tasks.schema.json
```

## Available skills

| Skill | Audience | Use it when |
|---|---|---|
| [`workplan`](workplan/SKILL.md) | planning agent | Creating, extending, re-deriving, updating, or auditing the plan as a graph |
| [`task`](task/SKILL.md) | implementation/review agent | Implementing or auditing exactly one active workplan task against its authored brief |

The ownership boundary is intentional: `workplan` decides whether task boundaries, dependencies, invariant coverage, and dispositions are correct; `task` decides whether one implementation satisfies one authored brief.

## Active workplan generations

When `docs/workplan/CURRENT` exists, it names the active generation under `docs/workplan/<generation>/`. Agents and automation must read that generation's `tasks.json`, `tasks.schema.json`, README, and briefs. Repositories without `CURRENT` use the legacy unversioned `docs/workplan/` layout.

`skills/workplan/tasks.schema.json` is the normative V2 tracker schema copied into newly seeded V2 generations.

## Getting an agent to use them

Point directly at the relevant skill when the agent has no native skill mechanism:

> Follow `skills/workplan/SKILL.md` and audit the plan.

> Follow `skills/task/SKILL.md` and implement P1-050.

Dispatched task agents are instructed by `scripts/task-prompt.md` to read the active brief, `skills/task/SKILL.md`, and applicable `AGENTS.md` files.

For agents that support skill directories, link each skill directory into the tool-specific location rather than duplicating the source. The canonical instructions remain under `skills/`.

## Writing a new skill

One directory per skill, `SKILL.md` inside, directory name matching the frontmatter `name`. Supporting templates, schemas, and reference material sit beside it.

```markdown
---
name: <kebab-case, matches the directory>
description: <what it does and when to load it>
---

# <Title>
```

Write instructions rather than commentary: what to read, what to do, what to verify, and what counts as done. If a paragraph does not change agent behavior, remove it.
