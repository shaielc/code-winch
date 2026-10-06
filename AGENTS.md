# AGENTS.md

Instructions for any agent working in this repository. One unit of work, one
pull request.

No implementation plan is in flight, and the product does not run.
[`docs/state.md`](docs/state.md) is the authoritative account of what exists and
what does not; read it before assuming a capability exists, because the design
set describes a destination and almost none of it is built.

An `AGENTS.md` deeper in the tree adds to or narrows these rules for its own
subtree.

## Start here

- **System state** — [`docs/state.md`](docs/state.md). What is reachable, what
  is built but unwired, and what has no code. Every claim in it carries the
  command or `file:line` that shows it.
- **Delivery order** — [`docs/roadmap.md`](docs/roadmap.md). The stages, what
  becomes true in each, what is deliberately absent, and the deferred decisions
  with their triggers.
- **Implementation plan** — [`docs/workplan/README.md`](docs/workplan/README.md).
  Phase 0 is open: seven tasks decomposing roadmap Stage 0. It carries the
  dependency graph, the width report, and the rule for what a brief may defer.
  `./scripts/list-available-tasks.sh` says which tasks you can pick up now.
- **Design baseline** — `docs/architecture.md`, `docs/code-structure.md`,
  `docs/contracts.md`, `docs/security.md`, and the ADRs in `docs/decisions/`.
  It describes the destination, not the build order — the roadmap is the order,
  and it deliberately inverts the layering these documents are written in.
- **How to run a task** — `skills/task/SKILL.md`. How to orient in an active
  brief, what its shape must demonstrate, and how to judge whether it is
  complete. It expands on the rules below.
- **Planning rules** — `skills/workplan/SKILL.md`. Read it when deriving the next
  plan from the design set and `docs/state.md`, or when changing a plan in
  flight.

## Rules that bind every task

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

In-memory and fake implementations are a supported way to run this product, not
test-only doubles. Keep them working and keep them controllable — scripted
transcripts, injectable latency, failure, malformed output. When you add one,
state in its documentation what it does not prove.

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

The plan in flight covers one roadmap stage, so it has no IDs in later phases to
own work the roadmap schedules for one. That work is a **non-goal** of your task,
with the stage cited — not a deferral. `docs/workplan/README.md` §*Scope of this
plan* states the convention. Acceptance criteria are not satisfied by code that
defers them.

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

**The Makefile and both CI workflows are currently broken**, every one of them
against a path that was removed with the last plan — `docs/state.md` lists each
with the line that names it. `make check`, `make build`, `make test`,
`make test-integration`, `make e2e`, and `make test-cycle` all fail before
reaching anything you wrote. Repairing them is the first stage's work, so until
that lands, do not report `make check` as passing and do not treat its failure as
caused by your change.

What does run today:

```sh
python3 -m unittest discover -s test                        # completion scripts
cd workplan-control-panel && python3 -m unittest discover -s tests
./scripts/list-available-tasks.sh                           # expect []
```

Once the gates are repaired, `make check` is the minimum and always: OpenAPI
validation, compatibility, and generated-output determinism, plus Go formatting,
vet, lint, tests, and the build. It is a `[host]` target needing go, npm, and
golangci-lint. `make test-cycle` is the Docker path and covers neither `lint` nor
`api-check`, so say which gates you ran rather than claiming `make check`.

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

Nothing of the product runs. What exists is the design set, `cmd/fake-harness`,
the workplan control panel, the planning skills, and the completion scripts —
`docs/state.md` has the detail, with a command or a `file:line` behind every
claim. No plan is in flight, so there are no task IDs to cite.

The next thing built is roadmap Stage 0: one container holding a runner and the
fake harness, started by hand, serving a page you can type at. That stage also
repairs the quality gates, because a stage whose first invariant is that the
system starts and deploys cannot stand on a build that fails. Read
`docs/roadmap.md` §1 before proposing work — the delivery order deliberately
inverts the layering the design documents are written in, so "the architecture
says this layer comes first" is not a reason to build it first.

Capabilities listed under *What is not implemented* in `docs/state.md` are gaps,
not features. Their presence in the design set is not evidence that they run, and
their absence is not an invitation to add them ahead of the stage that owns them.
