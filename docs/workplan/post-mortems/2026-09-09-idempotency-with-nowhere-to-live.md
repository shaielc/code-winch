# 2026-09-09 — An idempotency key with nowhere to live

**Found:** auditing P0-006 against pull request #49 at `044ef7a`, before merge.
**Symptom:** run creation is idempotent and correct in isolation, but the dedup
record lives inside `runs.resolved_configuration`, a column whose meaning
belongs to a different, already-designed feature. One write through that
column's own contract silently hollows the run — `workspacePath` empty,
`createdAt` `0001-01-01T00:00:00Z`, both `required` in the `Run` schema — and
destroys the dedup record, after which replaying the original `Idempotency-Key`
creates a second run.

## What the code did

`POST /api/v1/runs` requires an `Idempotency-Key` header and documents a 409
`idempotency_conflict` on reuse with a different body
(`api/openapi/code-winch.yaml`, `components.parameters.IdempotencyKey` and the
`/runs` `post` responses). P0-006 implements that contract by writing
`createActor` and `createIdempotencyKey` into the run's
`resolved_configuration` JSON and querying that path on every create, serialized
by an advisory lock (line numbers on the pull request branch at `044ef7a`):

```go
// repository.go:115
SELECT pg_advisory_xact_lock(hashtext($1),hashtext($2))
// repository.go:119
SELECT id::text FROM runs
WHERE resolved_configuration->>'createActor'=$1
  AND resolved_configuration->>'createIdempotencyKey'=$2
```

The blob is written once, at insert (`repository.go:82` and `:141`), and carries
`createdAt`, `updatedAt`, `workspacePath`, `harnessProfile`, `sandboxProfile`
alongside the two create keys (`runMetadata`, `repository.go:30-41`).

That column has a defined owner. `docs/code-structure.md` §6 *Configuration
layers*: "The fully resolved non-secret configuration is stored with a run for
reproducibility." The adapter says the same and marks the boundary explicitly —
`repository.go:46-48`, "records canonical, fully resolved non-secret JSON
*separately from* the existing RunRepository contract" — and writes it as one
canonical document, whole-value replace (`repository.go:52`), which is correct
for that meaning and fatal for anything else parked there.

## Root cause

No brief is wrong on its face. The defect is that the plan gave P0-006 an API
operation to own and no place to keep the operation's durable state.

**P0-006 declares `POST /api/v1/runs` as a contract surface**
(`phase-0/P0-006-create-and-read-runs.md`, Contract surfaces). The specification
for that operation makes `Idempotency-Key` required and documents the conflict
response, so durable deduplication is inside the surface the brief claims. Yet
the brief's Scope, Non-goals, Acceptance criteria, and Deferrals never mention
idempotency. An implementer reading only the brief builds create and get; one
who also reads the specification the brief points at must persist a dedup record
the brief never budgeted for.

**P0-006 was given no persistence surface.** Its Write set is
`cmd/winchd/main.go`, `internal/application/`, `cmd/winch/`, `test/e2e/`, and
the `Makefile`; `internal/adapters/postgres/` is absent. Its Contract surfaces
name an API pair and a partial port — no persisted aggregate, no migration slot.
The one task in Phase 0 whose whole objective is *"the run is persisted"* does
not declare the storage it persists into.

**Nothing in the current plan owns the `runs` schema.** The allocation
mechanism still points at the closed plan: `internal/adapters/postgres/migrate.go:29-31`
says "the numbered slots are pre-allocated to tasks in `docs/workplan/README.md`",
and the current README allocates none. P0-005 corrects that citation; it does not
re-establish the mechanism. So no migration slot was available to any task, and
the implementation says so in its own comment (`repository.go:106-107`): the
advisory lock exists "without changing the migration owned by the storage
foundation."

Given those three, the JSON blob was the only move left. The plan asked for
durable per-actor deduplication and offered nowhere to put it.

### What actually fits, and what does not

| Occupant of `resolved_configuration` | Belongs there |
|---|---|
| `harnessProfile`, `sandboxProfile`, `workspacePath` | Plausibly yes. §6 lists "named harness/sandbox profile" as a resolution layer; these are resolved non-secret configuration. |
| `createdAt`, `updatedAt` | No. Aggregate lifecycle, not configuration. `updatedAt` is mutable, which an insert-only document cannot express — `Save`'s update branch touches only `version` (`repository.go:85`), as do every supervisor writer, so it freezes at creation and will be wrong the moment P0-008 transitions a run. |
| `createActor`, `createIdempotencyKey` | No, and worst. Request-scoped bookkeeping used as a **uniqueness index** in a container that cannot express uniqueness. |

The distinction matters for the remedy. Move the last two rows out and the
column has one owner again, writing one document at two moments; whole-value
replace becomes correct rather than destructive. Leave them in and uniqueness
stays a convention: there is no constraint, only an advisory lock that every
writer must remember to take — and `RunRepository` still advertises a second
creation path that does not (`ports.go:52-56` on the branch, "A zero expected
version creates a record"), which writes no identity and takes no lock. The
lookup is also a `Seq Scan on runs` inside that lock, once per create.

## Why no test caught it

- The unit tests run against `memory.RunRepository`, which keeps a separate
  `creates` map. Idempotency is correct there by construction, in a structure
  that shares nothing with the JSON path it must match in PostgreSQL.
- There is no store contract suite. `test/contract/` covers harness and sandbox
  and openapi; a repository has no shared suite, so memory/PostgreSQL parity for
  this contract is asserted nowhere.
- The integration test exercises replay and conflict against a table only it
  writes. It never simulates the column's other owner.
- The destructive interaction needs a `SaveResolvedConfiguration` call, and that
  method has no runtime caller yet — the collision is unreachable from any
  current code path and only appears by hand or in the future task that
  implements configuration layering.

## Consequences if left as is

- Silent data loss and a broken idempotency guarantee the first time
  configuration layering writes the column. No error, no log.
- `updatedAt` wrong from the first state transition (P0-008).
- Uniqueness enforced by convention rather than by a constraint.
- A sequential scan per create, inside a held lock, growing with the table.
- P0-013 inherits the shape when it revises create/read for the memory store
  profile.

## Remediation

Not yet applied; it needs an owner, because it needs a migration slot.

1. `run_create_requests(actor, idempotency_key, run_id)` with
   `UNIQUE (actor, idempotency_key)` and a foreign key to `runs`. That retires
   the advisory lock, the sequential scan, and the overwrite hazard together.
2. Columns for `created_at` and `updated_at` (and `workspace_path`);
   `harness_driver` and `sandbox_driver` already exist from migration 003.
3. Leave the profile names in `resolved_configuration` — they belong there — and
   let create write the initial document that layering later rewrites.
4. One insert path: drop `Save(_, 0)` as a creation route or send it through
   `Create`.

If a migration slot cannot be allocated first, the interim step is to namespace
the blob (`{"run": …, "resolved": …}`) so the two writers stop clobbering each
other. That removes the data loss and leaves uniqueness a convention, so it is a
stopgap, not the fix.

Either way the work belongs to a task: P0-006's brief extended to declare the
persisted aggregate and a migration slot, or a task carrying a `revision` edge
to P0-006.

## Prevention

- Declaring an API operation as a contract surface means owning everything its
  specification requires of that operation — required headers and documented
  error responses included, not only the happy-path body. A brief that claims
  the operation and describes less has under-declared its own scope.
- A task that persists a new aggregate needs a persistence surface among its
  declarations. A write set that excludes the adapter where the task's data must
  live is a derivation error, visible while the brief is being written.
- When a plan closes, migration-slot allocation closes with it. Re-establishing
  who may add a migration belongs to the first task of the next plan, not to
  whoever first needs one — otherwise the need is met with a workaround inside
  whatever column happens to be reachable.
- A column with a documented owner is a contract surface. Storing something else
  in it is a contract change subject to the same rule as an API path or a port
  signature: it takes its design-document update in the same change, or it takes
  a dependency edge.
