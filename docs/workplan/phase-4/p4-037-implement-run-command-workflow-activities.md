# P4-037: Implement run command workflow activities

**Phase:** 4 — Top-level workflows
**Shape:** capability
**Dependencies:** P4-036 (compile: the activity registry and step lease this registers into), P1-050 (compile: the run use cases these activities call)

## Objective

A workflow starts a run, sends it a message, waits for an event, and stops it —
issuing ordinary application commands rather than touching a process.

## Scope

- `run.start`, `run.send`, `run.stop`, and `event.wait` with a typed predicate
  and deadline, as `docs/contracts.md` §7 specifies.
- Every command carries the coordinator's deterministic idempotency key, so a
  replayed step reuses the original run and input rather than creating a second.
- Run lineage: a run started by a workflow records the instance and step that
  started it, and the run detail shows that link.
- Instances pin harness and sandbox profiles at start; a step cannot select a
  different profile mid-instance.
- A workflow cannot broaden a run's permissions: the effective policy is the
  intersection of deployment, workspace, workflow, and profile.
- Retry with backoff at the step level, distinct from a user-initiated run retry.

## Non-goals

- Branching, parallelism, approvals, and artifacts — P4-038.
- Direct process or supervisor access. Activities call the same use cases the
  API calls.

## Runtime reachability

`winch workflow start` with a definition containing `run.*` steps, on the
compose stack with the fake harness.

## Owned surfaces

`internal/workflow/activities/run.go`, `internal/workflow/activities/wait.go`,
`schemas/workflow/v1/` (step payload schemas),
`test/scenarios/workflow-run-*.json`.

## Demonstration

    $ winch workflow start --definition test/scenarios/workflow-run-echo.json --json | jq -r .id
    $ winch workflow get $WID --json | jq -r '.steps[] | select(.type=="run.start") | .output.runId'
    $ winch run get $RID --json | jq -r '.lineage.workflowId'
    → expect: the run exists and points back at the instance and step

    $ winch run watch $RID
    → expect: the message the `run.send` step supplied, and the harness reply

    $ docker compose kill winchd && docker compose up -d winchd
    $ winch run ls --workflow $WID | wc -l
    → expect: still one run; replay reused the idempotency key

    # a definition whose step requests a profile the workspace prohibits:
    → expect: refused at start, naming the policy, before any run is created

## Verification

- Standing scenario suite passes with a workflow-driven run in the path.
- Crash-replay test asserting one run and one input per step attempt.
- `event.wait` tests: predicate match, deadline expiry, and an event arriving
  before the wait begins.
- Policy intersection test proving a workflow cannot widen permissions.

## Acceptance criteria

- [ ] Replay creates no duplicate run or input.
- [ ] Run lineage is recorded and visible.
- [ ] `event.wait` honours its deadline and cannot hang an instance forever.
- [ ] A workflow never obtains permissions its workspace denies.

## Deferrals

| Deferred | Owning task |
|---|---|
| Waiting on an approval | P4-038 |
| Exposing these steps over HTTP | P4-039 |

## Traces to

`docs/contracts.md` §7; `docs/security.md` §10 (workflows cannot broaden
permissions); `docs/architecture.md` §4; `docs/roadmap.md` Phase 4
