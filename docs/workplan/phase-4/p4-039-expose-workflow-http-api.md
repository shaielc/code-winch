# P4-039: Expose workflow HTTP API

**Phase:** 4 — Top-level workflows
**Shape:** capability
**Dependencies:** P4-037 (compile: the run activities being exposed), P4-038 (compile: the control-flow activities being exposed), P2-026 (semantic: workflow operations authorize against a workspace)

## Objective

A client registers a definition, starts an instance, watches it advance, and
cancels a branch — over the same authenticated API the run surface uses.

## Scope

- Definition registration with version pinning: a registered version is
  immutable, and an instance records the version it pinned.
- Instance operations: start, read, list, cancel instance, cancel branch, and a
  step-level retry where the definition permits it.
- A workflow event stream reusing the resumable `after_sequence` semantics of
  the run stream, so a client reconnects the same way.
- Authorization on every operation against the workspace, and audit entries for
  start, cancel, and definition registration.
- Generated client regeneration for the web workspace, keeping the Go and
  TypeScript boundary checked.
- `winch workflow` commands for every operation.

## Non-goals

- The graph UI — P4-040.
- Definition authoring tools or a visual editor.

## Runtime reachability

`/api/v1/workflows*` on the compose stack; `winch workflow`.

## Owned surfaces

`api/openapi/paths/workflows.yaml`,
`internal/adapters/transport/httpapi/workflow.go`,
`cmd/winch/workflow.go` (extended), `web/src/api/schema.ts` (generated).

## Demonstration

    $ winch workflow register --definition test/scenarios/workflow-run-echo.json --json | jq -r '.id, .version'
    $ winch workflow register --definition test/scenarios/workflow-run-echo.json  # same id, changed body
    → expect: refused; a registered version is immutable

    $ WID=$(winch workflow start --definition-id $DID --version 1 --json | jq -r .id)
    $ winch workflow watch $WID --after-sequence 0
    → expect: ordered step transitions, resumable after a dropped connection

    $ winch workflow cancel $WID --branch <step-id>
    → expect: that branch cancelled, siblings unaffected, and an audit entry

    $ winch workflow get $WID --token $OTHER_USER
    → expect: the same response as a workflow that does not exist

    $ make api-check
    → expect: the generated Go and TypeScript clients match the document

## Verification

- Standing scenario suite passes with a workflow started over HTTP.
- OpenAPI compatibility check against the committed baseline.
- Authorization tests per operation, including the negative case.
- Stream resume test with a forced disconnect mid-instance.

## Acceptance criteria

- [ ] A registered definition version is immutable and pinned by its instances.
- [ ] Every operation authorizes against the workspace and is audited.
- [ ] The workflow stream resumes with the same semantics as the run stream.
- [ ] Generated clients stay in sync and the compatibility check passes.

## Deferrals

| Deferred | Owning task |
|---|---|
| Graph and status rendering | P4-040 |

## Traces to

`docs/contracts.md` §5, §7; `docs/security.md` §9;
`docs/architecture.md` §8; `docs/roadmap.md` Phase 4
