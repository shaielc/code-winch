# ADR-0005: Sandbox-resident runner with its own attach surface

- **Status:** Accepted
- **Date:** 2026-10-05
- **Supersedes:** [ADR-0001](0001-modular-monolith-and-runner-boundary.md)

## Context

ADR-0001 put the runner inside the control-plane daemon and had it reached
through direct in-process transport, with the serializable protocol kept as
discipline for a later split. The first implementation attempt followed that
order — control plane, domain model, supervisor, event store — and never reached
a configuration anybody could start. The runner was the last thing wired, so the
only way to observe a harness was through everything else first.

The protocol discipline ADR-0001 asked for also went untested in the way that
mattered. A boundary exercised only by in-process calls is a boundary whose
serialization, versioning, and failure modes are hypothetical.

Meanwhile the smallest useful product is not the control plane. It is one agent,
in one container, that a person can watch and type at.

## Decision

The first deployable unit is a **sandbox container** holding three things: the
harness process, the runner that owns it, and an HTTP/WebSocket **attach
surface** a browser connects to directly. It also holds its own private,
ordinal-ordered record store, so a sandbox is self-sufficient and needs no
external database to be useful.

The control plane is added later, and never becomes the only path to a harness.
A sandbox started by hand with no control plane present is a supported
configuration, not a development shortcut.

Consequently the runner boundary is a network protocol from the moment the
control plane exists, because the runner is already in another process and
another container. There is no in-process transport stage.

## Consequences

The first stage of work is small and demonstrable: build the container, serve a
page, read the harness, write to it. Deployment is an invariant from the first
task rather than a later integration, because the unit of deployment is the unit
of work.

The browser gains a second surface, with the trust boundary that implies — it
now talks directly to a container running an untrusted agent. `docs/security.md`
§2 carries that boundary and §11 carries the blocker that keeps it off a shared
deployment.

The runner protocol cannot be deferred into a convenient in-process call, which
is the cost ADR-0001 was trying to avoid paying early. We pay it instead at the
stage the control plane arrives, against a runner that already works.

The sandbox's store is not the canonical event store. It holds runner-local
ordinals, as `docs/code-structure.md` §3 already required of a runner, and the
control plane promotes them to canonical sequence numbers when it begins
persisting. Whether that promotion wraps or replaces the sandbox's record shape
is roadmap decision D7.

Two stores means two retention stories and a window where a record exists in one
and not the other. That is accepted: the alternative is a sandbox that cannot run
without the control plane, which is the property this ADR exists to avoid.

## Alternatives

- **Keep ADR-0001's order** — runner inside the daemon, in-process first:
  rejected because it is what produced a system that never ran, and because it
  leaves the runner boundary untested in the only way that counts.
- **Runner as a separate process but with no surface of its own** — the browser
  always reaches it through the control plane: rejected because it makes the
  control plane a precondition for observing anything, which puts first light
  behind the whole middle of the system.
- **Sandbox surface as a temporary scaffold** to be deleted once the control
  plane serves the real UI: rejected in favor of [ADR-0006](0006-two-interaction-surfaces.md),
  which keeps it as the operator path. A scaffold would be removed exactly when
  debugging the control plane starts to need it.
- **No store in the sandbox** — stream only, and let the control plane persist:
  rejected because it makes a standalone sandbox forget everything on reload,
  and reload is the cheapest observation of persistence there is.

## Revisit when

The attach surface duplicates enough control-plane behavior — authorization,
multi-session history, renderer selection — that maintaining both costs more
than routing everything through the control plane would. Or when a sandbox
backend appears that cannot host a process of ours at all, such as a remote
execution service that only accepts a command and returns output.
