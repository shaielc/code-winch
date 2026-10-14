# P1-011: Report sandbox-backed child status

**Phase:** 1 — Sandbox exchange
**Shape:** capability
**Dependencies:** P1-009 (semantic: this task replaces the independently
identifiable `Child status: Unknown` browser constant while preserving its
accessible label/value identity)

## Objective

Make the deployed attach page report **Child status: Down** from sandbox-owned
session data when the sandbox owns no child. The browser must render the
sandbox's observation rather than infer `Down` from a failed request, missing
output, or the absence of a fake process.

## Scope

- Extend `GET /api/session` with `"childStatus":"Down"`. For this increment,
  `Down` is the only valid wire value: it means this sandbox composition owns no
  child. Preserve the existing `profile` and `unenforcedControls` fields and
  document the additive response in `docs/contracts.md` §8.
- Define the child-status value in the attach transport and have
  `cmd/winch-sandbox` explicitly supply `Down` when it constructs the server.
  This is sandbox-owned configuration, not a browser default or process probe.
- Replace P1-009's static `Unknown` value only after a successful response with
  the exact supported value. Preserve the independently named `Child status`
  region and use `Unknown` while the request is pending, fails, contains invalid
  JSON, or contains a missing/unknown `childStatus`.
- Extend focused Go contract and React component tests for the response and its
  semantic presentation. Preserve posture disclosure, the local composer,
  static serving, and the current loopback-only deployment.

The implementation may keep the status in the existing session fetch; it must
not add polling, streaming, a second endpoint, or a speculative lifecycle
abstraction. Increment 3 may change how the composition root obtains the value
when it first owns a child.

## Required constraints

`Down` is affirmative sandbox knowledge that its composition owns no child. A
request error or malformed/unsupported status leaves the presentation
`Unknown`; it must never manufacture `Down`, `Running`, readiness, fake
availability, or an exit result. JSON decoding must validate `childStatus`
rather than trusting a TypeScript assertion, while continuing to render the
existing posture response as before.

Do not start, inspect, signal, wait for, or supervise a process. Do not infer
status from `/healthz`, output, input availability, a PID search, or the shipped
fake binary. Keep logs and errors content-free, retain the loopback host
publication, and leave the image and local composer behavior intact.

## Non-goals

Child launch, ownership beyond the explicit no-child composition, cleanup,
`Running`/`Exited`, exit results, output or input transport, readiness,
persistence, provider semantics, polling, and broad page restructuring are
excluded. These are later decided Phase 1 increments, not unfinished work in
this closed `Down` observation.

## Reachability

`cmd/winch-sandbox` supplies the no-child status to the existing attach server;
the existing `/api/session` route carries it through the React bundle already
served by the image at `http://127.0.0.1:8080`. The maintained compose path uses
this composition root without starting `fake-harness`, so no new configuration,
provider account, or auxiliary service is required.

## Write set

- `internal/adapters/transport/attach/session.go`
- `internal/adapters/transport/attach/server.go` (only if construction needs the
  explicit session value separated from the current posture field)
- `cmd/winch-sandbox/main.go`
- `test/contract/attach/session_test.go`
- `web/src/attach/App.tsx`
- `web/src/attach/App.test.tsx`
- `web/src/attach/Posture.tsx` (only if response validation is factored beside
  the existing session type)
- `docs/contracts.md`

No fake-harness, process runner, deployment, CLI, output, input, or persistence
file is forecast.

## Contract surfaces

Owns the additive `childStatus` member of `GET /api/session` for the standalone
sandbox attach surface. Its complete wire contract in this increment is the
case-sensitive string `Down`; missing values and every other value are invalid
and must not be presented as authoritative. Owns replacing the value in
P1-009's accessible `Child status` region after a valid response, while
preserving that region's label and `Unknown` fallback.

Consumes the existing `/api/session` posture fields and P1-009 presentation
identity. It owns no child lifecycle, readiness, output/input state, exit
interpretation, control-plane run state, or provider meaning.

## Demonstration

After the automated gates, from the repository root:

```sh
docker compose -f deployments/compose.yml up --build -d
curl -fsS http://127.0.0.1:8080/api/session
```

Observe `childStatus` exactly `Down` alongside the unchanged
`container-standard` posture and `network-egress` disclosure. Open
`http://127.0.0.1:8080`; observe the independently labeled **Child status** with
value **Down**, the posture panel, and the usable local composer. Confirm the
compose service list contains no started fake/child process, then tear down:

```sh
docker compose -f deployments/compose.yml ps
docker compose -f deployments/compose.yml down
```

This fake-free check proves that the deployed no-child sandbox owns and
transports its `Down` observation and that the page renders it. It does not
prove child startup, supervision, readiness, output/input behavior, exit
reporting, provider compatibility, or shared-deployment security. The `curl`
response alone does not prove the React DOM; the page observation does not prove
a child was inspected.

## Verification

- `go test ./test/contract/attach -run 'TestEndpointContracts|TestRejectsMethodsAndUnknownAPI'`
  covers the exact additive session response and preserves request-method/API
  behavior. Add a focused constructor/default test if needed to prove that the
  composition root must explicitly supply the no-child status.
- `npm test -- --run src/attach/App.test.tsx` from `web/` covers the named region,
  valid `Down`, pending/error/invalid/missing status fallback to `Unknown`, and
  the prohibition on false `Running`, readiness, fake availability, and exit
  claims. It must also keep posture and composer assertions green.
- Run `npm run format:check`, `npm run lint`, `npm run typecheck`,
  `npm test`, and `npm run build` from `web/` for the forecast frontend surface.
- Run `python3 -m unittest discover -s test` and
  `(cd workplan-control-panel && python3 -m unittest discover -s tests)`.
- Run the required host `make check`, then `make test-cycle` for the composed
  image and standing scenarios. Report unavailable Docker/toolchains as
  limitations rather than passing evidence.
- Perform the compose, API, and deployed-page demonstration above and record
  each observation separately.

Negative tests must show that request rejection, invalid JSON, a missing
`childStatus`, and values including `Running` and `Exited` cannot produce an
authoritative status or readiness/exit claim. Server tests must show methods
remain constrained and posture fields are unchanged.

## Acceptance criteria

- [ ] The deployed no-child composition reports the exact additive
      `childStatus: "Down"` value from `GET /api/session`, with the posture
      response and contract documentation updated together.
- [ ] The existing accessible `Child status` identity renders exact `Down` only
      from the validated sandbox response and retains `Unknown` before or
      without a valid response.
- [ ] Negative tests prove request failure and malformed, missing, or unknown
      status cannot imply `Down`, `Running`, readiness, fake availability, or an
      exit result.
- [ ] The status is explicitly supplied by the no-child sandbox composition;
      no process is launched, probed, supervised, or treated as ready.
- [ ] Existing posture, local composer, loopback serving, image deployment,
      contract tests, and standing scenarios remain covered and reachable.
- [ ] The hand observation records exact API and DOM values and states what the
      fake-free `Down` result does and does not prove.

## Deferrals

None. Child launch, supervision, cleanup, and `Running`/`Exited` behavior are
excluded by the decided increment 3 boundary, not deferred completion work for
this sandbox-owned no-child status.

## Traces to

`docs/workplan/phase-1/README.md` increment 2; P1-009;
`docs/roadmap.md` §§1 and 3, Stage 0; `docs/contracts.md` §§8–9; ADR-0005;
ADR-0006.
