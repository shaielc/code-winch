# Architecture decision log

Decisions are recorded here before implementation makes them expensive to
reverse. Status values are Proposed, Accepted, Superseded, or Rejected.

| ID | Decision | Status |
|---|---|---|
| [ADR-0001](0001-modular-monolith-and-runner-boundary.md) | Modular monolith with a serializable runner boundary | Proposed |
| [ADR-0002](0002-canonical-events-and-renderers.md) | Canonical events separated from rendering | Proposed |
| [ADR-0003](0003-capability-based-adapters.md) | Capability-based harness and sandbox ports | Proposed |

New ADRs should contain context, decision, consequences, alternatives, and a
clear trigger for reassessment. Accepted ADRs are immutable; superseding one
adds a new ADR and updates this index.
