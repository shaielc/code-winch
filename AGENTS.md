# AGENTS.md

Instructions for any agent working in this repository. One task, one pull
request.

An `AGENTS.md` deeper in the tree adds to or narrows these rules for its own
subtree.

## Start here

- **Your brief** — `docs/workplan/<phase>/<id>-*.md`. Authoritative for scope,
  acceptance criteria, and required verification. Its dependencies are
  completion prerequisites, not suggested reading.
- **Design baseline** — `docs/architecture.md`, `docs/code-structure.md`,
  `docs/contracts.md`, `docs/security.md`, `docs/roadmap.md`, and the ADRs in
  `docs/decisions/`.
- **Planning rules** — `skills/workplan/SKILL.md`. Read it when you are changing
  the plan itself; you do not need it to implement a task.

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

No TODO, stub, unimplemented branch, or "handled in a later task" without a task
ID that exists in `docs/workplan/tasks.json` at the time you write it. If no
such task exists, either finish the work or add the task and say so in the pull
request. A brief's acceptance criteria are not satisfied by code that defers
them.

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

- Include `Task: <ID>` in the body, and no other task ID.
- Do **not** edit status fields in `docs/workplan/tasks.json`. Automation stamps
  `completed` when the pull request is approved.
- Report what you ran, what the demonstration showed, and anything you deferred
  together with its owning task ID.

## Current state

Some of the foundations these rules refer to are still being introduced. Each
now has an owning task:

| Foundation | Owner |
|---|---|
| Composition root, configuration, telemetry | P1-048 |
| Local runner and `winch dev run` | P1-049 |
| Run use cases reachable over HTTP | P1-050 |
| Maintained operator CLI | P1-049, P1-051 |
| Controllable fake runtime profile | P1-052 |
| Standing end-to-end scenario suite | P1-053 |

The rules above bind as each lands; until then, do not add code that makes them
harder to establish. `docs/workplan/wiring-plan.md` has the original diagnosis
of what was missing and why.
