# AGENTS.md

Instructions for any agent working in this repository. One task, one pull request.

An `AGENTS.md` deeper in the tree adds to or narrows these rules for its own subtree.

## Start here

Resolve the active workplan first. If `docs/workplan/CURRENT` exists, read the generation it names under `docs/workplan/<generation>/`; otherwise use the legacy unversioned `docs/workplan/` directory.

- **Your tracker entry and brief** — the active generation's `tasks.json` and the `brief` path named by your task entry. They are authoritative for scope, acceptance criteria, surfaces, dependencies, and required verification. Dependencies are completion prerequisites, not suggested reading.
- **Task rules** — `skills/task/SKILL.md`. Read it whenever you implement or audit one workplan task.
- **Design baseline** — `docs/architecture.md`, `docs/code-structure.md`, `docs/contracts.md`, `docs/security.md`, `docs/roadmap.md`, and the ADRs in `docs/decisions/`.
- **Planning rules** — `skills/workplan/SKILL.md`. Read it when changing or auditing the plan as a graph. Do not silently redesign the graph while implementing one task; report a plan defect instead.

## Rules that bind every task

### Respect the boundaries

Dependencies point inward: the domain depends on nothing, the application layer on the domain, adapters on what is inside them. No cross-adapter imports, and no provider-specific branches in generic application code.

Place code where `docs/code-structure.md` says it goes, and add a directory only when it will contain real implementation.

A change to a contract surface — an API path, event or protocol schema, port signature, configuration key, persisted state transition, registry namespace, or migration — requires updating the design document that describes it, or adding or superseding an ADR, in the same change.

### Leave the system runnable

Every task leaves the daemon startable and deployable. A task that introduces a seam also registers an implementation in the composition root and makes it reachable at runtime. Code that no runtime configuration reaches is not finished, however well it is tested.

### Close operator-visible capabilities in the same task

Every new operator-visible capability must be immediately reachable through the maintained hands-on surface. The maintained CLI is not a later integration layer: when a task introduces a capability, that task also adds the minimal maintained CLI/debug path needed to drive it by hand.

Later tasks may add genuinely new operator behavior, ergonomics, diagnostics, scripting guarantees, or independent hardening. They may not exist merely to make an earlier capability usable.

### Fakes are a shipped configuration

In-memory and fake implementations are a supported way to run this product, not test-only doubles. Keep them working and controllable — scripted transcripts, injectable latency, failure, malformed output, and deterministic seeding where appropriate. When you add one, state what it does not prove.

### Prove behavior, not only contracts

Contract suites are necessary and not sufficient. Beyond them:

- If a standing end-to-end scenario suite exists, your change keeps it green.
- If your task makes a substrate real — a database, process, container, or provider — run that same scenario against the real substrate as well as the fake profile.
- Your brief's demonstration is a manual check. Run it and report what you observed in the pull request.

### Defer nothing without an owner

No TODO, stub, unimplemented branch, or “handled in a later task” without a task ID that exists in the active tracker when you write it. If no such task exists, either finish the work or report a plan defect so the workplan can be corrected.

A genuine architectural open question belongs in the design set with an explicit trigger to revisit, not as an unowned implementation deferral.

### Test adversarially where it matters

Security-sensitive work requires negative and adversarial tests, not happy paths. Logs and traces carry resource IDs and exclude content and secrets by default. No live provider account is required in CI.

### Stay inside your surfaces

Other tasks may be in flight against the same tree. V2 briefs separate:

- **Write set** — concrete files or paths this task expects to modify.
- **Contract surfaces** — semantic namespaces this task may change even when files differ.

Stay inside both declarations unless repository reality makes a deviation necessary. If a change outside them is unavoidable, explain why in the pull request. For legacy briefs, treat their existing owned-surface declaration as the applicable constraint.

### Respect the authored workplan version

A task's `workplan_version` determines the planning contract its brief must satisfy. Load `skills/workplan/contracts/vN.md` for version N; do not substitute the current contract for historical work.

This does not grandfather repository behavior: current repository invariants still have to hold, and current gaps belong to explicit current-version repair tasks.

## Verification

At minimum, and always:

```sh
make check
```

That runs OpenAPI validation, compatibility, generated-output determinism, Go formatting, vet, lint, tests, and the daemon build. It is a `[host]` target and needs go, npm, and golangci-lint installed.

Without a host toolchain, `make test-cycle` runs the Go gates and integration suite in Docker. It does not cover `lint` or `api-check`, so say so rather than claiming `make check` passed.

If your change touches storage, migrations, or process lifecycle:

```sh
make test-integration
```

If you touched `web/`:

```sh
cd web && npm run format:check && npm run lint && npm run typecheck \
  && npm run test && npm run build
```

Then run everything the active brief lists under **Verification**. Those checks are the minimum evidence expected in the pull request, not a ceiling.

## Pull requests

- For a versioned generation, include `Task: <generation>/<ID>` in the body (for example `Task: v2/P1-050`) and no other task identity. Legacy unversioned plans continue to use `Task: <ID>`.
- Do **not** edit execution status fields in the active tracker while implementing a task. Automation stamps `completed` when the pull request is approved.
- Report what you ran, what the demonstration showed, surface deviations, and anything deferred together with its owning task ID.

## Planning ownership

Use `skills/workplan/SKILL.md` for plan creation, extension, clean-slate re-derivation, updates, and graph audits. Use `skills/task/SKILL.md` for one task's implementation or implementation audit.

A clean-slate re-derivation archives the old generation and derives remaining work from HEAD plus completed history only. It does not expose unfinished task identities or briefs from the previous generation to the deriving agent.

The distinction remains intentional: a task implementation may correctly satisfy its authored brief while a workplan audit still finds that the brief itself is incorrectly decomposed. The task skill reports that as a plan defect; the workplan skill owns graph repair.
