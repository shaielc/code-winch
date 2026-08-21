# AGENTS.md

Instructions for any agent working in this repository. One unit of work, one
pull request.

No implementation plan is in flight. [`docs/state.md`](docs/state.md) is the
authoritative account of what runs and what does not; read it before assuming a
capability exists.

An `AGENTS.md` deeper in the tree adds to or narrows these rules for its own
subtree.

## Start here

- **System state** — [`docs/state.md`](docs/state.md). What is reachable, what
  is built but unwired, and what has no code. Every claim in it carries the
  command or `file:line` that shows it.
- **Design baseline** — `docs/architecture.md`, `docs/code-structure.md`,
  `docs/contracts.md`, `docs/security.md`, `docs/roadmap.md`, and the ADRs in
  `docs/decisions/`.
- **How to run a task** — `skills/task/SKILL.md`. Applies once a plan exists in
  `docs/workplan/`: how to orient in a brief, what its shape must demonstrate,
  and how to judge whether it is complete. It expands on the rules below.
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
With a plan in flight that owner is a task ID that exists in
`docs/workplan/tasks.json` at the time you write it; with no plan in flight,
either finish the work or record the gap in `docs/state.md` under *what is not
implemented* and say so in the pull request. Acceptance criteria are not
satisfied by code that defers them.

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

At minimum, and always:

```sh
make check
```

That runs OpenAPI validation, compatibility, and generated-output determinism,
plus Go formatting, vet, lint, tests, and the daemon build. It is a `[host]`
target and needs go, npm, and golangci-lint installed.

Without a host toolchain, `make test-cycle` runs the Go gates and the integration
suite in Docker. It does not cover `lint` or `api-check`, so say so in the pull
request rather than claiming `make check` passed.

If your change touches storage, migrations, or process lifecycle:

```sh
make test-integration
```

If you touched `web/`:

```sh
cd web && npm run format:check && npm run lint && npm run typecheck \
  && npm run test && npm run build
```

Then run everything your brief lists under **Verification** or **Required
verification**. Those are the minimum evidence expected in the pull request, not
a ceiling.

## Pull requests

- When a plan is in flight, include `Task: <ID>` in the body and no other task
  ID, and do **not** edit status fields in `docs/workplan/tasks.json` —
  automation stamps `completed` when the pull request is approved.
- Report what you ran, what the demonstration showed, and anything you deferred
  together with its owner.

## Current state

Some foundations these rules refer to do not exist yet. The daemon has a
composition root, configuration, and telemetry, and `winch dev run` drives a
harness by hand; the run use cases are unbound, the fake profile is not
controllable, there is no standing end-to-end scenario suite, and the operator
CLI covers `dev run` only. `docs/state.md` records each of these with evidence.

The rules above bind as each foundation lands. Until then, do not add code that
makes them harder to establish.
