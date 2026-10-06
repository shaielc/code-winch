# ADR-0006: Two interaction surfaces with distinct jobs

- **Status:** Accepted
- **Date:** 2026-10-05

## Context

[ADR-0005](0005-sandbox-resident-runner.md) gives every sandbox an attach
surface a browser reaches directly. Once the control plane arrives with a
product application over many runs, there are two browser-facing surfaces, and
the obvious instinct is to treat one as temporary.

Both directions of that instinct are wrong in a predictable way. Deleting the
attach surface removes the only way to observe a sandbox when the control plane
is the thing that is broken. Growing it toward parity produces two
implementations of run listing, authorization, and renderer selection, drifting
apart at whatever rate the two get attention.

The workplan model independently requires a maintained hands-on path that does
not depend on the state of the product UI (I5). That requirement and the attach
surface are the same thing, so they should be the same artifact rather than two.

## Decision

Both surfaces are permanent, with stated and different jobs.

The **sandbox attach surface** serves one harness in one sandbox. It is the
operator and debug path, and it is the I5 hands-on path: when a task introduces
an operator-visible capability in the sandbox, that capability is drivable here.
It assumes a single user who already has access to the host or the port.

The **control-plane application** serves the product: many runs, workspaces,
history, workflows, accounts, and policy. It is where a user who is not an
operator does their work.

Feature parity is an explicit non-goal. The attach surface does not grow run
lists, search, workflow views, or multi-user access. The product surface does
not become the only way to see a harness.

Where they overlap is rendering. Record projections — terminal, conversation,
activity, changes — are shared implementations over the same record shapes, not
reimplemented per surface. Roadmap decision D2 settles whether that sharing is
one web workspace or two.

## Consequences

Debugging the control plane is possible, because the path that does not involve
it still works. A sandbox remains useful with no control plane deployed, which
is what makes the early stages of `docs/roadmap.md` shippable rather than
preparatory.

Two surfaces means two authorization stories. The attach surface's story is
deliberately thin — bound to loopback, one token, origin checked — and
`docs/security.md` §11 blocks exposing it in a shared deployment without the
control plane in front. That thinness is sustainable only because the non-goal
above holds; a surface that grew toward parity would need the full story and
would then be a second place to get it wrong.

Renderer sharing constrains the record shapes: a projection usable from both
surfaces cannot take a control-plane-only type as input. This is a real design
constraint on the record model, and the stage that introduces the second
projection is where it first bites.

## Alternatives

- **Attach surface as a scaffold**, deleted at the stage the product UI lands:
  rejected because it is removed exactly when it becomes most useful, and because
  I5 would then need a separate artifact built to replace it.
- **Attach surface grows into the product UI**, with the control plane serving
  only an API: rejected because every sandbox would then need the product's
  authorization and history model, in a process that sits inside the trust
  boundary it would be enforcing.
- **One surface, always proxied** through the control plane: rejected by
  ADR-0005, which keeps a standalone sandbox a supported configuration.
- **Terminal-only attach**, no browser surface on the sandbox: rejected because
  the product's output model is richer than a terminal from Stage 2 onward, and
  an operator path that cannot show what the product shows cannot diagnose it.

## Revisit when

The attach surface has acquired a second concern beyond one sandbox's session —
a run list, cross-session history, or a second user — which means the non-goal
has failed and the boundary needs restating or removing. Or when renderer
sharing has forced a record shape that neither surface wants.
