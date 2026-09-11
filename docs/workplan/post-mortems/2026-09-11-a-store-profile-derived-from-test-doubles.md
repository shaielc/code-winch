# 2026-09-11 — A store profile derived from test doubles

**Found:** auditing P0-012 against pull request #50 at `b161fa9`, before merge.
**Symptom:** the implementation matched its brief, passed every gate, and
reproduced every demonstration — and still could not satisfy its own acceptance
criteria. Following the brief exactly produced a composition root in which the
**postgres** profile is wired with an in-memory adapter, and a memory profile
missing two of the ports the run path is about to consume.

**Outcome:** P0-012 through P0-018 withdrawn from phase 0. The work is wanted and
is re-registered as a deferral into phase 1, to be derived from the ports rather
than from the adapters.

## What the brief asked for

P0-012's Scope named the set the composition root was to wire:

> Introduce a composition-root helper that constructs the memory repositories
> (`RunRepository`, `EventStore`, `OutboxPublisher`, `SupervisorStore`, clock,
> and ID source) as one wired set, parallel to the existing postgres path.

The implementation did exactly that (`cmd/winchd/main.go:104-142` on the branch),
including a doc comment promising what the list could not deliver:

```go
// storeSet keeps the complete run-path storage seam together so selecting a
// profile cannot accidentally leave one of its ports backed by another one.
```

and then, in the **postgres** branch:

```go
return storeSet{Runs: store, Events: store, Outbox: &memory.OutboxPublisher{},
    Supervisor: store, ...}, pool.Close, nil
```

That is not a slip. `postgres.Store` has no `Publish` method, and a compile-time
assertion says so: `*Store does not implement application.OutboxPublisher
(missing method Publish)`. Having been told to put `OutboxPublisher` in the set,
there was nothing else to put there.

## Root cause

No brief is wrong on its face. The defect is that P0-012's port list was read off
the **contents of `internal/adapters/memory`** instead of off the ports the run
path consumes, and those two sets are not the same set.

Interface assertions against both adapters at HEAD:

| Port | `memory` | `postgres.Store` |
|---|---|---|
| `CreateRunRepository` | yes | yes |
| `RunRepository` | yes | yes |
| `EventStore` | yes | yes |
| `SupervisorStore` | yes | yes |
| `OutboxPublisher` | yes | **no** |
| `OutboxStore` | **no** | yes |
| `InputCommandStore` | **no** | yes |
| `WorkspaceRepository` | yes | **no** |

`internal/adapters/memory` was never a parallel implementation of the store. It
is a collection of test doubles whose membership was decided, one at a time, by
whatever a use-case test needed next — which is why it holds a
`WorkspaceRepository` for an aggregate that does not exist in the domain, and no
`OutboxStore` for one that does. Listing its contents and calling them "the
durable ports" imported that history into the plan.

The two outbox ports are where it surfaces, because their names are close and
their roles are not:

- `OutboxPublisher` (`internal/application/ports.go:83`) is `Publish` — the
  delivery side, a transport concern. A store does not implement it. Memory has
  it because a test needed to observe published messages.
- `OutboxStore` (`:97`) is `ClaimOutbox` / `CompleteOutbox` / `RetryOutbox` /
  `PoisonOutbox` / `OutboxBacklog` — the durable side, which is what a store
  profile owes. Postgres has it; memory does not.

A store set must declare the second and never the first. P0-012's declared the
first and never the second, so the postgres branch had to borrow a memory object
for a port it should never have been asked for, and the memory branch silently
lacked the port it actually owed.

### The surface already had an owner

`OutboxPublisher` is not an unclaimed port. **P0-009 declares it as a contract
surface of its own**
(`phase-0/P0-009-send-input-and-drain-outbox.md`, Contract surfaces):

```
- API: POST /api/v1/runs/{runId}/input
- port: OutboxPublisher (non-streaming implementation)
```

So P0-012 and P0-009 shared a contract surface. By the model's own rule that is a
dependency edge, with the task that *defines* the surface first — and the plan
recorded no edge between them at all. P0-012 was listed as available at plan
open, concurrent with the task that owns the port it was told to wire.

This is the mechanical form of the defect, and it is the one a rule can catch:
had the two declarations been compared, the collision would have been visible
while the phase was being derived, and the port list would have had to be
justified against the ports rather than the package.

### Why the repair tasks inherited it

P0-013 through P0-018 each revise one seam onto the memory profile. Every one of
them assumes P0-012 produced a store set the run path can be moved onto. Because
the set was wrong at its root, each revision task was specified against a
foundation that does not hold:

- P0-015 was to *revise* input and outbox for memory. Memory has neither
  `InputCommandStore` nor `OutboxStore`, so P0-015 was not a revision of anything
  — it was the first implementation of two adapters, filed under a `revision`
  edge.
- P0-013 through P0-017 collectively own every file in `test/e2e/`, while P0-012,
  a `capability`, was given no scenario file. A capability's demonstration is a
  new scenario in the standing suite, so P0-012 was specified unable to
  demonstrate its own shape.

Seven tasks, each individually plausible, resting on a port list that was never
checked against the application layer.

## Why no test caught it

- `stores.Events`, `stores.Outbox`, and `stores.Supervisor` are assigned and
  never read; only `Runs`, `Clock`, and `IDs` reach `NewRunService`
  (`cmd/winchd/main.go:65`). Exported fields of an unexported struct, so neither
  `go vet` nor the linter objects, and the wrong wiring is inert rather than
  broken.
- The task's own test asserts the six fields are non-nil
  (`cmd/winchd/main_test.go:140`). Non-nil is satisfied by a memory publisher in
  the postgres profile, and says nothing about the two ports absent from the
  struct.
- `make check`, `make test-integration`, and `make e2e` all pass on the branch.
  The defect is in which ports were chosen, and no gate reads the plan.
- The e2e suite could not have caught it either: every file in `test/e2e/`
  requires `PG_TEST_DATABASE_URL` and skips without it, so the postgres profile
  is the only profile it has ever exercised.

## Consequences if left as is

- The postgres profile carries a memory adapter. Inert today, live the moment
  P0-009 starts the outbox worker and something reads `stores.Outbox`.
- P0-015 is mis-shaped: a `revision` edge over adapters that do not exist.
- I3 stays unmet for storage while seven tasks report progress toward it, which
  is the status-truthfulness failure the model warns about — `completed` meaning
  "the brief was followed" rather than "the criteria hold".

## Remediation

P0-012 through P0-018 are removed from `docs/workplan/tasks.json` and their
briefs deleted. Phase 0 now states plainly that it exits with I3 unmet for
storage and I4's first rung unbuilt, and phase 1's *Deferrals in* register carries
all three gaps.

Removing IDs is against the tracker's append-only rule, and this record is what
keeps the history the rule protects: P0-012 through P0-018 existed, were derived
on 2026-09-09, were never dispatched or merged, and were withdrawn here. No
merged pull request cites them. When the work is re-derived in phase 1 it takes
new IDs; the old ones are not reissued.

What the re-derivation must start from, so the same mistake is not made twice:

1. **Enumerate the ports, not the package.** The set is what the run path
   consumes at the moment the task lands — read off `internal/application`, with
   `OutboxStore` and `InputCommandStore` in it and `OutboxPublisher` out.
2. **Compare against P0-009's declarations** before declaring anything. The
   publisher is its surface.
3. **Expect to write adapters.** `memory.OutboxStore` and
   `memory.InputCommandStore` do not exist. "The adapters already exist" is true
   of four ports and false of two, and the difference is the size of the task.
4. **Give the task a scenario it can run.** Every current e2e file skips without
   a database; a fake profile that no scenario exercises is the test-only double
   the profile exists to stop being.

## Prevention

- **A port list in a brief is a claim about the application layer, and it is
  checkable.** `var _ application.Port = (*adapter.Type)(nil)` compiles or it
  does not. Any brief naming the ports a composition root wires can be checked
  this way while it is being written, in about a minute. P0-012's list would not
  have compiled.
- **Deriving a profile from an adapter package inverts the dependency rule.**
  `docs/code-structure.md` §2: application packages define ports, outer adapters
  implement them. A brief that lists what an adapter happens to contain and calls
  that the port set has let the adapter define the boundary. Read the ports; then
  ask which adapter satisfies each.
- **"Parallel to the existing X path" is an assertion that needs checking.** Two
  adapter sets are parallel only if they implement the same ports. Where they do
  not, the asymmetry is the finding, and it belongs in the brief that discovered
  it rather than being absorbed by whoever implements it.
- **Ports whose names differ by one word deserve the roles spelled out.**
  `OutboxPublisher` and `OutboxStore` differ by delivery versus durability, and
  nothing in either name says so. A brief naming one should say which side it
  means.
- **A test double is not a profile, and the gap between them is work.** A package
  built to satisfy tests has the shape its tests asked for. Promoting it to a
  supported way to run the product means deriving what the product needs and
  filling the difference — not wiring up what is already there. This is what I3
  asks for, and the shortest way to fail it is to assume the adapters are done.
- **A `capability` with no scenario file is mis-shaped.** The phase's own e2e
  table allocates one file per task. A task whose shape demands a scenario, in a
  phase that allocates it none, is a derivation error visible in the phase
  README.
