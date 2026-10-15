---
name: refine-guidance
description: Refine the repository's guidance — docs, skills, and agent extensions — by finding redundant or repeated instructions, outdated instructions, unused skills, gaps in the workplan model and skills, and content worth splitting, moving, or merging, or, when none are obvious, practices and agent features worth adopting; then apply only changes the maintainer approves. Use in an interactive session with a maintainer who can answer questions; not for dispatched or unattended runs.
---

# Refine guidance

Act as a guidance maintainer working **with** a person. Every classification,
plan, and edit passes through that person: ask for explanations where intent is
unclear, permission before any action beyond reading, and approval of each
planned change. Use the agent's question tool for this (in Claude Code,
`AskUserQuestion`), with concrete options and a recommended one first. A
question answered in prose by the agent's own guess is not an answer.

Read applicable `AGENTS.md` files first and follow them throughout. Read files
with the reading tools, change them with the editing tools, and explain any
complex read-only command before running it.

## Scope

The guidance surface:

- `AGENTS.md` files at every depth;
- `skills/README.md`, every `skills/*/SKILL.md`, and `skills/shared/`;
- `docs/workplan/README.md`, phase READMEs, pending briefs, and
  `tasks.schema.json`;
- agent extensions: `workplan-control-panel/control_panel/prompts/`,
  `.claude/` skill links and settings, and any other agent-specific wiring;
- design documents (`docs/*.md`, ADRs) only where they carry instructions to
  agents rather than product meaning.

Completed briefs, merged PRs, post-mortems, and `docs/state.md` are records.
Read them as evidence; do not edit them.

**Ask first:** which parts of the surface to cover in this session, and whether
there is a known pain point that prompted it. Start there.

## Find

Read the selected surface end to end. Record each finding with its category,
every `file:line` involved, and a one-sentence statement. Do not edit while
finding.

- **Redundant or repeated.** The same rule stated in more than one place, or
  near-identical boilerplate across briefs, phases, or skills. Name the
  canonical home: the model defines, skills link to it, the index states a
  plan-wide sentence once.
- **Outdated.** References to retired IDs, paths, skill names, increment labels,
  or flows; instructions that contradict the model, another skill, `AGENTS.md`,
  or the scripts and code at HEAD. Check each claim against the file it cites.
- **Unused skills.** A skill not reached from `AGENTS.md`, `skills/README.md`,
  another skill, a brief, or a control-panel prompt; a README row with no
  directory or a directory with no row; a link that does not resolve.
- **Gaps.** A step the guidance requires that no skill or role owns (a "must
  exist" with no "who creates it, when"); a term used but not defined in the
  model; a model rule no skill procedure applies; an operation in the workflow
  order with no skill.
- **Structure.** Where content lives no longer matches how it is used, so
  splitting, moving, or merging skills or documents would help. Count inbound
  references while reading. For example, not limited to:
  - many skills cite one specific section of a shared document, which could be
    split into its own file under `skills/shared/`;
  - many skills point to one skill for the same steps, which could move to
    `skills/shared/` or the model;
  - two skills or documents share an audience and overlapping procedure, and
    could merge;
  - content sits in the wrong home, such as procedure in an artifact or a
    definition in a skill instead of the model, and could move.

Repetition is not automatically redundant, and an old phrase is not
automatically outdated. A restatement can be deliberate, for example a brief
that must stand on its own.

## Improve

Run this only when Find produced no obvious issues in the selected scope, or
when the maintainer asks for it after those findings are settled. Look for
improvements rather than defects. The areas below are examples of where to
look, not a checklist:

- **Practices.** Compare the guidance with current practice for agent
  instructions and skills, for example descriptions that carry their trigger
  conditions, instructions over explanation, checklists and templates over
  prose, short files with linked reference material, or one audience per skill.
- **Agent features.** Check whether the agents this repository supports offer
  features the guidance could use, for example skill frontmatter fields, skill
  discovery and loading, subagents, hooks, permission settings, interactive
  question tools, or anything new since the guidance was written. Read the agent's
  documentation when a tool for it is available, and ask permission before any
  web search or fetch.
- **Repository conventions.** For example, gates, scripts, or automation the
  guidance could invoke instead of describing by hand, or repeated manual steps
  a script could check.

Record each one as a finding with its source — a documentation page, a release
note, or the files it would change — and how it would change what an agent
does. Pass it through Explain and Plan like any other finding. A feature is
proposed only where it removes a real cost in this repository, not because it
exists.

## Explain

Before classifying a finding as a defect, ask the maintainer about any whose
intent you cannot establish from the files: why the repetition exists, whether
a flow is retired, whether a skill is still used. Batch related findings into
one question, at most four per round. Record the answer beside the finding.
Drop findings the maintainer explains as intended.

## Plan

For each remaining finding, propose one change: the files, the canonical
location, and the replacement text or its outline. Prefer, in order:

1. deleting a duplicate and linking to the canonical statement;
2. correcting stale text in place;
3. adding the missing owner to an existing skill or the model;
4. splitting, moving, or merging skills and documents, when the reference counts
   behind a structure finding justify it;
5. a new skill, only when no existing skill's audience fits.

A split, move, or merge also lists every inbound link it changes, the
`skills/README.md` rows it adds or removes, and any agent extension (control
panel prompts, `.claude/` links) that names the old path. Retired skill names
and paths still cited by records stay readable: say how a reader of an old PR or
completed brief finds the new location.

Present the plan per finding or per small group for approval, with options to
apply, modify, or skip, and a preview of the change. Apply nothing that was not
approved; a modified plan is approved again before it is applied.

Ask permission separately before anything beyond reading and editing:
creating a branch, committing, pushing, opening a PR, running a shell command
that writes, or touching a record listed under Scope.

## Apply

Make the approved edits with the editing tools. Follow
`skills/README.md` *Writing a new skill* for any skill change. After each
group, re-read the changed text and every reference to it, and report anything
the change made stale before moving on.

## Verify

- Every relative link in changed files resolves.
- `skills/README.md` rows match the skill directories, and each frontmatter
  `name` matches its directory.
- No tracker status field changed.
- `git diff --check` and `python3 -m unittest discover -s test`; the
  control-panel tests when its prompts changed.

## Done

Every finding is applied, modified and applied, skipped by the maintainer, or
explained as intended, and the summary lists which. The change is one unit of
work in one PR with the `no-task` label, opened only with the maintainer's
permission.
