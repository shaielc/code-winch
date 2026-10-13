# P0-009: Plan Phase 2 increment 1: harness profile presentation

**Phase:** 0 — Plan the implementation phases
**Shape:** planning
**Dependencies:** P0-002

## Objective

Produce one complete, independently authored implementation brief for Phase 2
increment 1, **Harness profile presentation**, and reconcile that brief with the
tracker, index, Phase 0 coverage and Phase 2 coverage.

## Scope

Use `skills/derive-task/SKILL.md` against the independently audited design at
`docs/workplan/phase-2/README.md` (reviewed commit
`0d8e2f85b77bff9952ca4bce47aad1f0e7c2ed42`; no blocking audit findings). Derive
the single smallest complete vertical slice that lets a person select the fake
profile through the maintained operator path and see its identity and versioned
capabilities on the existing loopback attach page before exchange. The brief
must include the genuine wrapper-owned adapter interface and descriptor, the
fake implementation, composition-root/runtime selection wiring, page
presentation, and parity evidence against the standing Stage 0 exchange.

Treat completed Stage 0 exchange as a runtime prerequisite to recheck at
dispatch, not as behavior present at this planning commit. If current source or
the eventual Stage 0 contract invalidates the decided boundary, return the
finding through Phase 0 to the phase author rather than silently redesigning
the increment.

## Required constraints

The selected fake remains a supported shipped runtime profile and the wrapper
seam must be reached by a runnable configuration; an interface, descriptor, or
page mock with no runtime wiring is incomplete. The descriptor is versioned and
truthfully reports only fake-profile identity, adapter version, I/O and login
modes, and capabilities that the implementation actually supports. The same
Stage 0 scenario must remain green through the selected fake profile, and the
brief must state what fake parity does not prove.

The implementation brief owns no Claude Code descriptor, executable/version
preflight, login readiness, vendor launch or decoding, or exit semantics. It
must not claim that one fake implementation proves a seam suitable for a second
vendor. Preserve the phase's decided ordinary-pipe transport and deferred D5
trigger without introducing PTY support.

## Non-goals

No Phase 2 product implementation occurs in this planning task. Do not derive a
second Phase 2 implementation brief here. Vendor readiness, Claude Code
exchange, vendor output interpretation, and native exit visibility retain their
not-yet-derived coverage; second-adapter proof remains roadmap Stage 6, and
credential/egress enforcement remains Stage 5.

## Planning outputs and handoff

- A different task author derives
  `docs/workplan/phase-2/<allocated-ID>-harness-profile-presentation.md` and its
  concrete tracker entry together. No implementation ID is reserved in advance.
- Reconcile `docs/workplan/README.md`, `docs/workplan/phase-0/README.md`,
  `docs/workplan/phase-2/README.md`, tracker dependencies and coverage, graph
  metrics, and write/contract collisions.
- Hand the finished implementation brief to a separate refiner using
  `skills/refine-task/SKILL.md` against current HEAD.
- **Create the next concrete Phase 0 planning task** for increment 2, Vendor
  readiness, with a separately authored complete brief and matching
  tracker/index/coverage records. This is a required tracked handoff, not a
  recommendation; do not allocate the increment 2 implementation ID.

## Write set

Forecast for the assigned operation:
`docs/workplan/phase-2/<allocated-ID>-harness-profile-presentation.md`,
`docs/workplan/tasks.json` (new entries only; existing statuses unchanged),
`docs/workplan/README.md`, `docs/workplan/phase-0/README.md`,
`docs/workplan/phase-2/README.md`, and the new concrete Phase 0 planning brief
for Vendor readiness. No later implementation IDs are reserved.

## Contract surfaces

This Phase 0 planning task changes no product contract. The derived
implementation brief must declare the versioned capability-descriptor contract,
runtime profile-selection surface, wrapper-owned adapter port, and attach-page
presentation it defines or changes. It must identify the Stage 0 exchange and
composition-root surfaces it consumes and resolve every overlap with another
available implementation task through a justified contract edge, not a write
collision warning.

## Demonstration

Review the resulting implementation brief and confirm its manual path starts the
deployed sandbox in its fake profile, uses the maintained operator selection
path, opens the loopback attach page, observes the fake identity and truthful
versioned capabilities before exchange, and completes the standing Stage 0
input/output/reload scenario through that same selected seam. Confirm its
negative checks reject unknown profiles and false descriptor claims without
making the runtime unstartable. This is artifact review, not a claim that the
profile selector or adapter seam already exists.

## Verification

Validate tracker schema, brief existence, dependency references, unchanged
preexisting status fields, links, graph/collision reporting, and matching Phase
0/2 coverage. Inspect current Stage 0 source and gates when deriving the brief;
label proposed implementation commands and behavior as predictions. Run
applicable planning/completion tests and record any unavailable gates rather
than claiming runtime evidence that did not run.

## Acceptance criteria

- [ ] One complete Phase 2 implementation brief covers exactly Harness profile
      presentation and has one matching real tracker ID.
- [ ] Its scope forms a runnable vertical slice through selection, the genuine
      wrapper port, versioned descriptor, fake implementation, composition root,
      maintained operator path, attach page, and standing Stage 0 scenario.
- [ ] The brief separates current source truth from its Stage 0 prerequisite and
      specifies truthful, adversarial observations for unknown selection and
      unsupported descriptor claims.
- [ ] Concrete write sets, owned and consumed contracts, dependencies, manual
      demonstration, automated evidence, fake limitations, and inherited gates
      are explicit and credible.
- [ ] Tracker, index, Phase 0/2 coverage, graph and collision warnings agree,
      with every preexisting status value preserved.
- [ ] The implementation brief goes to a separate refiner, and a separately
      authored concrete Phase 0 planning brief for Vendor readiness exists in
      tracker and coverage records without a speculative implementation ID.

## Deferrals

The remaining Phase 2 increments own vendor readiness, real exchange, decoding,
and native-exit behavior in the decided artifact. Those coverage boundaries are
not IDs that permit the derived implementation brief to leave an incomplete
fake-profile slice or an unwired seam. D5 remains deferred under its roadmap
trigger; new terminal evidence must return through Phase 0.

## Traces to

`docs/workplan/phase-2/README.md` increment 1; Phase 0 increment 0.2;
`docs/roadmap.md` Stage 1 and D5; `docs/contracts.md` §§8–9;
ADR-0003/0005; `skills/shared/workplan-model.md` invariants I1–I7.
