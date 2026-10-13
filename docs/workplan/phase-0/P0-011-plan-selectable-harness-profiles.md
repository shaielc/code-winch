# P0-011: Plan Phase 2 increment 1: selectable harness profiles

**Phase:** 0 — Plan the implementation phases
**Shape:** planning
**Dependencies:** P0-002 (semantic: the assigned brief must consume the reviewed Phase 2 selectable-profile boundary, descriptor ownership, and fake-parity requirements)

## Objective

Produce one complete, independently authored implementation brief for Phase 2
increment 1, **Selectable harness profiles**, and reconcile it with the tracker,
workplan index, and Phase 0/2 coverage. The derived seam task must make `fake`
and `openai-codex` selectable behind one genuine wrapper boundary, expose the
selected versioned capability descriptor through the maintained attach path,
and keep the fake exchange scenario green without starting Codex.

## Scope

Use `skills/derive-task/SKILL.md` against the independently audited Phase 2
design at commit `8a482cc42594cec434378e79199c8f49259441aa` and code at dispatch
HEAD. Derive exactly increment 1. Recheck the named Stage 0 prerequisites rather
than treating their planned Phase 1 coverage as implemented. Inspect the actual
composition root, fake launch/exchange path, attach presentation, runtime
configuration, standing scenarios, and maintained operator path before fixing
the implementation write set and contracts.

The implementation brief must form one reachable seam slice: a wrapper
interface, explicit registration and profile selection, versioned descriptors,
fake migration, attach/operator presentation, documentation, and parity evidence.
The descriptor owns the selected adapter ID and adapter version, supported CLI
version, pipe transport, login mode, and input/output capabilities decided by
Phase 2. It describes the selected Codex profile without launching it. Require
unknown/invalid profile refusal and descriptor validation tests so configuration
cannot silently fall back or advertise unsupported capability.

Allocate the next plan-wide implementation ID only when its complete brief and
matching tracker entry exist. Reconcile dependency reasons, coverage, graph
metrics, and collision analysis in the index and both phase records. Arrange
handoff to a different refiner using `skills/refine-task/SKILL.md`, then arrange
a separately authored, concrete Phase 0 planning follow-up for increment 2 while
actionable Phase 2 coverage remains. That follow-up receives its own real ID and
brief only when derived; do not reserve an implementation ID.

## Required constraints

- Distinguish source-observed Stage 0 behavior from predicted implementation
  results. If selectable launch/exchange, attach presentation, or runtime
  configuration is missing, the implementation brief must name the real
  prerequisite or return the boundary for phase revision; it must not absorb
  missing Stage 0 work into Stage 1.
- Preserve loopback-only serving, the sandbox posture disclosure, deployment,
  ordering/retention/input guarantees that exist at dispatch, and the fake as a
  documented, deterministic, account-free shipped profile.
- Keep generic application/runner and attach behavior vendor-neutral. Codex
  identity and descriptor construction belong at the adapter boundary; no
  provider-specific branch may enter generic application code.
- Require one explicit configuration choice, stable descriptor validation, and
  negative coverage for unknown profile IDs, duplicate registrations, malformed
  descriptors, and descriptor values that contradict the selected adapter.
- Require the existing standing exchange scenario through the fake profile and
  a manual attach-page check. State that fake parity proves wrapper selection,
  descriptor presentation, and existing fake exchange only—not Codex login,
  process launch, protocol compatibility, network access, or native exit.
- Treat Codex CLI `0.159.3` and stdio pipes as decided descriptor facts. Do not
  reopen D1/D5 or add PTY alternatives during task derivation.

## Non-goals

This task writes planning artifacts only. Its target increment excludes Codex
installation/login checks, Codex process launch or initialization, native
JSON-RPC exchange/decoding, native exit interpretation, a second vendor, browser
login, scoped credential injection, lifecycle control, and Stage 0 implementation.
It does not derive a second implementation brief or reserve later IDs.

## Planning outputs and handoff

Create `docs/workplan/phase-2/<allocated-ID>-selectable-harness-profiles.md` and
its matching tracker entry, then reconcile `docs/workplan/README.md`,
`docs/workplan/phase-0/README.md`, and `docs/workplan/phase-2/README.md`. Record
the concrete wrapper/descriptor/configuration contract surfaces, composition
root and attach/operator wiring, scenario updates, documentation, and exact
write forecast discovered at dispatch. Preserve every pre-existing status.

Hand the finished brief to a different refiner. If increment 2 remains
actionable, arrange another independently authored tracked Phase 0 planning
brief that assigns its derivation; repeat this concrete one-increment handoff
until Phase 2 coverage is complete. Unproduced briefs remain **not yet derived**
and carry no speculative IDs.

## Write set

- `docs/workplan/phase-2/<allocated-ID>-selectable-harness-profiles.md`
- `docs/workplan/tasks.json` (one new implementation entry; existing statuses unchanged)
- `docs/workplan/README.md`
- `docs/workplan/phase-0/README.md`
- `docs/workplan/phase-2/README.md`
- `docs/workplan/phase-0/<next-allocated-ID>-<next-increment-slug>.md` and its
  pending tracker record only when the required next concrete handoff is derived

## Contract surfaces

This planning task changes no product contract. Its derived brief will own the
genuine harness wrapper interface, adapter registry/profile-selection
configuration, versioned capability descriptor, and selected-descriptor attach
presentation. It consumes Stage 0's actual child exchange and configuration
surfaces found at dispatch and must record collisions with their owners. It does
not own login readiness, Codex launch/protocol messages, vendor decoding, native
exit meaning, lifecycle, credentials, or a second-adapter generalization proof.

## Demonstration

Review the derived implementation brief and matching tracker/index/coverage
records. Confirm its predicted composed manual path selects `fake`, shows the
versioned fake descriptor on the loopback attach page, and completes the
standing fake exchange; then selects `openai-codex`, shows only its truthful
versioned descriptor, and does not launch it or claim readiness. Confirm an
unknown profile fails closed, all records use one real implementation ID and
dependency set, and the separately tracked next planning handoff—if actionable
coverage remains—allocates no later implementation ID.

## Verification

Run completion-script and control-panel tests, tracker JSON-schema validation,
dependency/brief resolution, `scripts/list-available-tasks.sh`, an `rg`
consistency search for P0-011 and the allocated implementation ID,
`git diff --check`, `make check`, and `make test-cycle`. Report unavailable
Docker/toolchains as limitations. The derived brief must additionally name
focused descriptor/configuration/registry tests, fake scenario parity, attach
presentation checks, and the gates for every forecast implementation surface.

## Acceptance criteria

- [ ] Exactly one complete Phase 2 implementation brief covers increment 1
      with one real ID and dependencies on only behavior it actually consumes.
- [ ] The brief owns one genuine reachable wrapper seam, explicit profile
      registration/selection, a versioned evidence-backed descriptor, attach
      presentation, fake migration, and fake scenario parity in the same slice.
- [ ] Invalid/unknown selection and malformed or contradictory descriptors fail
      closed; generic code remains free of Codex-specific branching.
- [ ] Source-observed Stage 0 behavior and predicted results are separate, and
      missing Stage 0 prerequisites are not silently implemented as Stage 1.
- [ ] Tracker, index, Phase 0/2 coverage, graph metrics, and collision warnings
      agree and preserve all existing statuses.
- [ ] A different refiner receives the finished implementation brief, and a
      separately authored real planning follow-up is arranged for increment 2
      when it remains actionable; no speculative implementation ID exists.

## Deferrals

None. Login readiness, initialized Codex launch, native conversation exchange,
and interpreted harness exit are excluded by decided Phase 2 increments 2–5,
not unfinished work in the selectable-profile seam.

## Traces to

`docs/workplan/phase-2/README.md` increment 1; Phase 0 increment 0.2;
`docs/roadmap.md` Stage 1 and D1/D5; `docs/contracts.md` §§8–9;
`docs/code-structure.md` §4; ADR-0003.
