# Workplan

[Phase 0](phase-0/README.md) is the active planning and issue-resolution intake.
It schedules decided phase designs across the [roadmap](../roadmap.md), then
separate derivation of specific briefs. Product implementation starts at Phase 1.
The closed plan's evidence remains in [the state report](../state.md); its
P0-001 through P0-007 IDs remain retired. Allocate new IDs beyond 007, including
the plan-wide counter when implementation tasks are derived.

## Phases to plan

| Target | Roadmap milestone | Phase 0 increment | Artifact to produce |
|---|---|---|---|
| Phase 1 — Standalone fake sandbox | Stage 0, unfinished | `plan-sandbox` | `phase-1/README.md` |
| Phase 2 — First real harness | Stage 1 | `plan-first-vendor` | `phase-2/README.md` |
| Phase 3 — Readable, rejoinable session | Stage 2 | `plan-readable-session` | `phase-3/README.md` |
| Phase 4 — Observe manual sandboxes | Stage 3 | `plan-observer` | `phase-4/README.md` |
| Phase 5 — Control-plane run lifecycle | Stage 4 | `plan-run-lifecycle` | `phase-5/README.md` |
| Phase 6 — Enforced security | Stage 5 | `plan-security` | `phase-6/README.md` |
| Phase 7 — Second provider and workflows | Stage 6 | `plan-workflows` | `phase-7/README.md` |

Target paths above are planned outputs, not existing designs. No target
implementation task exists merely because its phase is listed.

## Phase 0 graph and handoff

Derivation is **complete: all 8 increments are covered by 10 planning briefs**.
The first seven tasks assign `create-phase` for distinct target designs; their
product-stage order creates no planning completion edge. Issue tasks assign
separate `derive-task` authors and consume P0-008's decided Phase 1 artifact.
They require no Phase 1 implementation. Target designs and repair briefs remain
planned outputs; none has been created by registering these planning tasks.

| Task | Planning output | Dependencies and reasons |
|---|---|---|
| [P0-008](phase-0/P0-008-plan-standalone-fake-sandbox.md) | Phase 1 / Stage 0 decided design | None — independent target planning artifact. |
| [P0-009](phase-0/P0-009-plan-first-real-harness.md) | Phase 2 / Stage 1 decided design | None — independent target planning artifact. |
| [P0-010](phase-0/P0-010-plan-readable-rejoinable-session.md) | Phase 3 / Stage 2 decided design | None — independent target planning artifact. |
| [P0-011](phase-0/P0-011-plan-manual-sandbox-observer.md) | Phase 4 / Stage 3 decided design | None — independent target planning artifact. |
| [P0-012](phase-0/P0-012-plan-control-plane-run-lifecycle.md) | Phase 5 / Stage 4 decided design | None — independent target planning artifact. |
| [P0-013](phase-0/P0-013-plan-enforced-shared-deployment-security.md) | Phase 6 / Stage 5 decided design | None — independent target planning artifact. |
| [P0-014](phase-0/P0-014-plan-second-provider-and-workflows.md) | Phase 7 / Stage 6 decided design | None — independent target planning artifact. |
| [P0-015](phase-0/P0-015-plan-composer-refusal-focus.md) | #80 specific Phase 1 brief | P0-008 — contract: decided Phase 1 increment boundaries govern this brief. |
| [P0-016](phase-0/P0-016-plan-composer-validation-keyboard-evidence.md) | #81 specific Phase 1 brief | P0-008 — contract: decided Phase 1 increment boundaries govern this brief. |
| [P0-017](phase-0/P0-017-plan-local-message-no-ordinal-oracle.md) | #82 specific Phase 1 brief | P0-008 — contract: decided Phase 1 increment boundaries govern this brief. |

The actual graph has **10 tasks and 3 contract edges**. Its critical paths
are **P0-008 → P0-015**, **P0-008 → P0-016** and **P0-008 → P0-017**, each
**2 tasks** long; average width (task count / longest path) is **5.0**.
Topological level widths are **7 then 3**. The maximum possible available
frontier is **9**: complete P0-008 while the six other roots remain pending,
making all three issue-planning tasks available alongside them. With every task
pending, **P0-008–P0-014 are the 7 available tasks**;
issue planning becomes available when P0-008 is completed. No statuses or
availability facts are inferred from predicted product behavior.

**Available write collisions: 21 pairs.** Every unordered pair among
P0-008–P0-014 shares `docs/workplan/README.md` and Phase 0 artifact/handoff
coverage writes, plus `tasks.json` through approval automation's completion
stamps. Target phase README paths are distinct. These are merge/rebase warnings,
not dependency edges. Once P0-008 completes, issue-planning pairs also contend
on `phase-1/README.md`, `tasks.json` and the shared index/Phase 0 coverage;
that frontier has 36 unordered write-warning pairs, with its three issue pairs
also sharing Phase 1 coverage. They create distinct real target-brief files at
dispatch. Product test-file overlap is evaluated for the derived targets, not
assumed as a planning edge.

**Available contract collisions: none.** P0-008–P0-014 define distinct
`phase-1` through `phase-7` planning-artifact namespaces. The encoded edges
order their only shared meaning: each issue-planning task consumes P0-008's
decided `phase-1` boundary. The issue tasks define distinct
`composer-issue-80/81/82` artifact scopes; sharing a read-only phase design
does not make those scopes one contract. Index/tracker appends and coverage
references share bytes, not target-artifact meaning.

#80 owns the planned focus-refusal repair scope, including button refusal.
#81's planned specific brief covers its three remaining regression guarantees;
the duplicated button-refusal item is intake attribution, not a target completion
deferral to a planning task or a hypothetical implementation ID. #82 plans an
observable local-only message oracle against current DOM and the decided phase,
without inventing runner ordinal UI. Full-issue closure needs its own assessment
of exact current issue evidence and implemented behavior.

## Planning author checks

Planning-task completion proves its named artifact and handoff, never the
predicted product capability. Use the assigned operation and a separate
author for its phase audit or task derivation. Phase-design operations stop
before implementation brief allocation; specific-brief operations integrate
each real brief/tracker/index/coverage entry together. Leave status fields to
approval automation and keep `tasks.schema.json` unchanged.

From the repository root, validate tracker required fields/values against the
frozen schema, resolved brief paths, unique append-only IDs, referenced
dependencies, acyclicity and equality of the index/coverage graph. Read every
artifact reference and check scope/observations against current source. Run:

```sh
python3 -m unittest discover -s test
(cd workplan-control-panel && python3 -m unittest discover -s tests)
bash scripts/list-available-tasks.sh
git diff --check
make check
```

Report missing Go/Docker/web prerequisites as verification limits; do not install
them solely to produce a planning artifact. No runtime test of a target
capability is evidence required to claim that capability implemented here.

## Using the plan

All future phase planning, specific-brief derivation and issue resolution must
enter through [Phase 0's intake procedure](phase-0/README.md#future-planning-and-issue-resolution-intake).
Use `extend-phase` or `revise-phase` there for new or changed planning scope,
then a separate `derive-task` author registers the concrete planning brief.
Specific-brief planning consumes a decided target phase; a separate author
derives the implementation brief and hands it to refinement when dispatched.
Do not implement from a roadmap row, an issue, or a phase-design planning brief.

`tasks.json` is the machine-readable graph; statuses and availability come from
it and `bash scripts/list-available-tasks.sh`. Planning completion certifies
the named artifact, never its predicted runtime behavior. Status stamping stays
with approval automation and the schema is frozen. Every implementation deferral
needs an existing tracker owner or a triggered architectural decision.

## Phase exit rule

Exit follows each phase's stated observations and acceptance evidence, not an
empty tracker or a list of derived briefs. Phase 0 exits when its designs and
issue-resolution briefs are reviewed and ready for separate-author handoffs.
Adding new intake updates that exit coverage; closing a phase requires the
explicit `close-phase` operation and preserves source-supported state evidence.
