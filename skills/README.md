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
| [`create-phase`](create-phase/SKILL.md) | phase author | Opening a phase with ordered increments and explicit boundaries |
| [`extend-phase`](extend-phase/SKILL.md) | phase author | Adding coverage to an existing phase |
| [`revise-phase`](revise-phase/SKILL.md) | phase author | Changing increment boundaries or order |
| [`audit-phase`](audit-phase/SKILL.md) | independent phase auditor | Judging phase design, task coverage, graph, and completion |
| [`derive-task`](derive-task/SKILL.md) | task author, separate from phase author | Deriving one brief or rederiving a pending brief and its graph records |
| [`refine-task`](refine-task/SKILL.md) | task refiner | Making a derived brief concrete against dispatch-time code |
| [`implement-task`](implement-task/SKILL.md) | task implementer | Building, verifying, and opening the task PR |
| [`audit-task`](audit-task/SKILL.md) | independent task auditor | Judging the implementation against its accepted brief |
| [`task-implement-merge`](task-implement-merge/SKILL.md) | implementation merge operator | Checking the current audit, obtaining finding verdicts, and merging implementation into the task branch |
| [`close-phase`](close-phase/SKILL.md) | phase closer | Explicitly ending a phase and preserving its system record |
| [`propagate-task`](propagate-task/SKILL.md) | task agent or maintainer | Routing audit findings to tasks/issues and opening a new-branch propagation PR into the task branch |
| [`reflect-task`](reflect-task/SKILL.md) | task agent or maintainer | Finding flaws in the development methodology and improving its planning, handoffs, and decisions |

Use different authors for phase design and task derivation. Derivation,
refinement, implementation, and audit of a task are separate assignments.
The flow is architecture → phase → task → optional phase revision → next task;
dispatched task delivery is refine → implement → audit → task-implement-merge.
That merge goes into the task integration branch; its PR to the default branch
remains a separate operation.

Audience matters. Say who a skill is for when you add one: an agent should be
able to tell from the table whether a skill is addressed to it before loading
anything.

## Shared reference material

`shared/` holds documents that more than one skill depends on. A skill links to
what it needs and does not restate it.

| Document | Defines |
|---|---|
| [`shared/workplan-model.md`](shared/workplan-model.md) | Shared terms, artifact layout, Phase 0 meta tasks, per-phase task IDs, seven runtime invariants, planning and implementation task shapes, dependency-edge reasons, write sets and contract surfaces, brief anatomy, and the frozen tracker schema |

The [phase format](shared/phase-design.md) is shared by the phase design
operations. Procedure lives in each operation skill; the model holds definitions.

## Getting an agent to use them

### Any agent, no setup

Point at the path:

> Follow `skills/audit-phase/SKILL.md` and audit Phase 0 for coverage gaps.

> Follow `skills/implement-task/SKILL.md` and implement `<task ID>`.

This always works and needs nothing configured. It is the fallback when an
agent has no skill mechanism, or when you want a skill used once without wiring
it in permanently.

### Dispatched task agents

The control panel's bundled prompts in
`workplan-control-panel/control_panel/prompts/` point each dispatched stage at
its own skill: `refine.md` uses `refine-task`, `implement.md` uses
`implement-task`, and `audit.md` uses `audit-task`. The refiner commits only the
brief before implementation; the implementer builds the accepted brief; the
auditor posts its report on the implementation PR. The root `AGENTS.md` routes
agents to the same operations.

`task-implement-merge` handles the audited implementation PR explicitly; it is
not a new automatically dispatched control-panel stage. `propagate-task`
reconstructs the branch stack and submits follow-up changes through a separate
PR into the task branch. `reflect-task` examines the development method rather
than deciding findings or merging changes.

Keep that wiring narrow. A dispatched agent is pointed at the skill addressed to
it, not at every skill in this directory.

### Claude Code

Symlink the skill into `.claude/skills/`. The skill itself stays here; only a
link sits in the dot-directory, and Claude Code follows it and reads `SKILL.md`
from the target.

```sh
mkdir -p .claude/skills
ln -s ../../skills/create-phase .claude/skills/create-phase
ln -s ../../skills/derive-task .claude/skills/derive-task
ln -s ../../skills/implement-task .claude/skills/implement-task
echo '.claude/' >> .gitignore
```

They are then invocable by their directory names, and Claude loads one on its
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
