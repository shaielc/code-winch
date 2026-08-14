# P3-033: Display and test effective security posture

**Phase:** 3 — Docker isolation
**Shape:** capability
**Dependencies:** P3-030 (semantic: there must be a resolved profile to show), P3-031 (semantic: the network decision is part of the posture), P3-032 (semantic: the credential binding is part of the posture)

## Objective

Before starting a run, a user sees the isolation, filesystem, network, and
credential posture that will actually apply — and `local-trusted` says plainly
that it is not a sandbox.

## Scope

- A posture summary computed from the resolved profile and the driver's declared
  capabilities, not from the requested configuration: isolation class,
  filesystem access, network policy, bound credentials by name, and resource
  limits.
- Shown before launch and on the run detail afterwards, with the run's stored
  posture used for a finished run so history stays truthful.
- A prominent, non-dismissible warning wherever `local-trusted` is enabled,
  stating it is a process-lifecycle boundary rather than a security boundary.
- A driver whose declared capability and enforced behavior disagree fails a test
  rather than being displayed optimistically.
- Adversarial integration tests: the escape, exhaustion, egress, and credential
  attempts from `docs/security.md` §12, each attempted and visibly refused, with
  the refusal appearing in the audit trail.

## Non-goals

- Changing enforcement. This task displays and tests what P3-029 through P3-032
  enforce; a disagreement is fixed in the owning task.
- Security dashboards and alerting — P5-047.

## Runtime reachability

The run creation form and run detail in the web app; `winch run posture <id>`
and `winch profiles --posture`.

## Owned surfaces

`web/src/features/runs/PosturePanel.tsx`, `cmd/winch/posture.go`,
`api/openapi/components/run.yaml` (posture fields), `test/security/`.

## Demonstration

    $ winch profiles --posture
    → expect: local-trusted reported as unisolated, in the same words the UI uses

    # in the browser, creating a run on local-trusted:
    → expect: a non-dismissible warning that this is not a sandbox

    # creating a run on container-readonly:
    → expect: read-only filesystem, egress denied, no credentials — before start

    $ winch run posture $ID   # after the run finished
    → expect: the posture that applied, from the stored resolved profile

    $ make security-tests
    → expect: every adversarial attempt refused, each with an audit entry

## Verification

- Standing scenario suite passes with the posture surface present.
- A test asserting the displayed posture is derived from resolved configuration
  and declared capabilities, never from the request.
- Capability-honesty test per driver: declared equals enforced.
- The adversarial suite from `docs/security.md` §12 runs in CI where an engine
  is available.

## Acceptance criteria

- [ ] Posture is visible before launch and immutable in history afterwards.
- [ ] `local-trusted` is labelled unisolated everywhere it appears.
- [ ] A driver cannot advertise a capability it does not enforce and still pass.
- [ ] Every adversarial attempt is refused and audited.
- [ ] Phase 3's exit statement in `docs/roadmap.md` is demonstrated by this suite
      together with the standing scenarios.

## Deferrals

| Deferred | Owning task |
|---|---|
| Alerting on posture regressions | P5-047 |

## Traces to

`docs/security.md` §4, §12, T02, T12, LB02, LB03;
`docs/architecture.md` §4 (web application); `docs/roadmap.md` Phase 3 exit
