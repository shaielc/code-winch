# AGENTS.md

Instructions for any agent working in this repository. One unit of work, one
pull request.

The current workplan is closed. The sandbox/page slice and local composer
exist; the harness is shipped but is not started by the sandbox yet.
[`docs/state.md`](docs/state.md) records the current source-supported capabilities,
verification limits, lessons, and unfinished work. The tracker is valid and empty.
Check code at HEAD before assuming a capability exists; the design set describes
a destination. Closure does not certify roadmap Stage 0 completion.

An `AGENTS.md` deeper in the tree adds to or narrows these rules for its own
subtree.

## Start here

- **System state** — [`docs/state.md`](docs/state.md). What is reachable, what
  is built but unwired, and what has no code. Every claim in it carries the
  command or `file:line` that shows it.
- **Delivery order** — [`docs/roadmap.md`](docs/roadmap.md). The stages, what
  becomes true in each, what is deliberately absent, and the deferred decisions
  with their triggers.
- **Workplan status** — [`docs/state.md`](docs/state.md). There is no active
  phase or assignable brief. `docs/workplan/tasks.json` is a valid empty tracker
  and `docs/workplan/tasks.schema.json` retains the format for future plans.
  `bash scripts/list-available-tasks.sh` returns `[]`. An explicitly requested
  future phase uses `create-phase`; a separate author derives tasks with
  `derive-task`. Do not infer completion from the empty tracker.
- **Design baseline** — `docs/architecture.md`, `docs/code-structure.md`,
  `docs/contracts.md`, `docs/security.md`, and the ADRs in `docs/decisions/`.
  It describes the destination, not the build order — the roadmap is the order,
  and it deliberately inverts the layering these documents are written in.
- **Operation skills** — [`skills/README.md`](skills/README.md). Phase authors
  use `create-phase`, `extend-phase`, `revise-phase`, or `audit-phase`. A separate
  task author uses `derive-task`. Dispatched tasks use `refine-task`,
  `implement-task`, and `audit-task`, with a different author for each operation.
  `task-implement-merge` merges audited implementation into its task branch after
  finding verdicts. `propagate-task` routes follow-ups through a separate PR to
  that branch; `reflect-task` examines the development methodology.
  Explicit phase closure uses `close-phase`; ending a whole plan closes all its phases.

## Rules that bind every task

### Read files with the reading tools

Read files with Read/Search tools. When attempting a read-only complex command, explain what you are trying to do before executing.

### Change files with the editing tools

Change files with Edit and Write tools. Always ask for permission when using
shell commands to change files.

### Respect the boundaries

Dependencies point inward: the domain depends on nothing, the application layer
on the domain, adapters on what is inside them. No cross-adapter imports, and no
provider-specific branches in generic application code.

Place code where `docs/code-structure.md` says it goes, and add a directory only
when it will contain real implementation.

A change to a contract surface — an API path, an event or protocol schema, a
port signature, a migration — requires updating the design document that
describes it, or adding or superseding an ADR, in the same change.

### Leave the system runnable

Every task leaves the daemon startable and deployable. A task that introduces a
seam also registers an implementation in the composition root and makes it
reachable at runtime. Code that no runtime configuration reaches is not
finished, however well it is tested.

### Fakes are a shipped configuration

In-memory and fake implementations are supported ways to run this product.
They are malleable: change their behavior and implementation to exercise the
capability being built, updating callers, tests, and documentation together.
Their current protocol and controls do not constrain product design.
State what the fake does not prove. See `docs/contracts.md` §9.

### Prove behavior, not only contracts

Contract suites are necessary and not sufficient. Beyond them:

- If a standing end-to-end scenario suite exists, your change keeps it green.
- If your task makes a substrate real — a database, a process, a container, a
  provider — run that same scenario against the real substrate as well as
  against the fake profile.
- Your brief's demonstration is a manual check. Run it and report what you
  observed in the pull request.

### Defer nothing without an owner

No TODO, stub, unimplemented branch, or "handled later" without a named owner.
That owner is a task ID that exists in `docs/workplan/tasks.json` at the time you
write it, or a deferred decision with a trigger in `docs/roadmap.md` §6.

When an active plan covers one roadmap stage, it has no IDs in later phases to
own work the roadmap schedules for one. Such work belongs in **Non-goals**, with
the stage cited. A brief may defer implementation only to an existing tracker
ID, or cite a deferred architectural decision with a trigger in roadmap §6.
Phase-local increment labels and hypothetical later IDs are not owners.
Acceptance criteria are not satisfied by code that defers them. The closed
tracker currently provides no task IDs to own new deferrals.

### Test adversarially where it matters

Security-sensitive work requires negative and adversarial tests, not happy
paths. Logs and traces carry resource IDs and exclude content and secrets by
default. No live provider account is required in CI.

### Stay inside your surfaces

Other tasks are in flight against the same tree. Your brief names the files and
contracts it owns; edits outside them cause avoidable conflicts and usually
indicate scope creep. If a change outside your surfaces is unavoidable, say why
in the pull request.

## Verification

[`docs/testing.md`](docs/testing.md) defines the repository's gates, what to run
in your environment, how CI covers the rest, and where a demonstration runs.

Then run everything your brief lists under **Verification**. That is the minimum
evidence expected in the pull request, not a ceiling.

## Pull requests

- Include `Task: <ID>` in the body and no other task ID, and do **not** edit
  status fields in `docs/workplan/tasks.json` — automation stamps `completed`
  when the pull request is approved. A pull request that implements no task needs
  the `no-task` label; without one of the two it cannot pass the status gate.
- Report what you ran, what the demonstration showed, and anything you deferred
  together with its owner.

## Current state

The first runnable slice is the sandbox image and its loopback-only attach page.
It states the `container-standard` posture and names network egress as
unenforced. The fake harness ships in the image but is not started yet.

Roadmap Stage 0 remains unfinished: no runner starts the harness, retains
output, or delivers server input. Source for the quality gates has been repaired;
`docs/state.md` records which checks ran and which could not run. Read
`docs/roadmap.md` §1 before proposing work — the delivery order deliberately
inverts the layering the design documents are written in, so "the architecture
says this layer comes first" is not a reason to build it first.

Capabilities listed under *What is not implemented* in `docs/state.md` are gaps,
not features. Their presence in the design set is not evidence that they run, and
their absence is not an invitation to add them ahead of the stage that owns them.
