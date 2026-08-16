# 2026-08-16 — Migrations were never planned to run twice

**Found:** reviewing P1-048 before its pull request merged.
**Symptom:** `docker compose -f deployments/compose.yml up` succeeds once. Every
later `up` that keeps the `postgres-data` volume fails with SQLSTATE 42P07,
`relation "runs" already exists`, and the daemon exits before it listens. The
only recovery was `down -v`, which discards the database.

## What the code did

`MigrateUp` concatenated all five migrations into one unconditional `Exec`. The
migrations are bare `CREATE TABLE` and `ALTER TABLE … ADD CONSTRAINT` with no
`IF NOT EXISTS` and no record of what had already been applied. P1-048 put that
call in the composition root, where it runs on every boot.

## Root cause

No single brief is wrong on its face. The defect is the seam between two.

**P1-011 owns the migrations, and planned them to run once.** Its only migration
acceptance criterion is "Migration up/down or forward-only policy is tested on a
clean database" (`phase-1/p1-011-implement-postgresql-run-and-event-storage.md`,
acceptance criteria). A clean database is the only environment the brief
contemplates, and no other brief or design document states a migration execution
model at all. The delivered code and documentation said so deliberately —
"intentionally forward-only in production" in `migrate.go`, and "call
`MigrateUp` during deployment before serving traffic" in the adapter README.
That is a run-once-per-deploy migrator, faithfully built to its brief.

**P1-048 made migration a per-boot operation** and required, under Verification,
"a second start against a migrated database is a no-op"
(`phase-1/p1-048-boot-the-daemon.md`). It recorded that as a verification line
rather than as scope: `internal/adapters/postgres/` is absent from its owned
surfaces, and its P1-011 edge is `compile`.

So the plan required a property of P1-011's code that no task ever asked
P1-011's owner to provide.

Secondary, and smaller: the implementer of P1-048 shipped without the test
rather than flagging the gap. `AGENTS.md` requires finishing the work or filing
the task and saying so in the pull request; neither happened.

## Why no test caught it

Every migration test starts clean by construction. The integration helper runs
`DROP SCHEMA public CASCADE; CREATE SCHEMA public` immediately before each
`MigrateUp`, and `TestMigrationRoundTripOnCleanDatabase` runs `MigrateDown`
first. The second-run path was unreachable from the suite.

## Contributing plan condition

P1-010 – P1-020 built every component a deployment needs while
`cmd/winchd/main.go` was still `func main() {}` (`docs/workplan/README.md`,
Phase 1), so migrations had no startup caller to be exercised from. That
violates I2 — "Migrations run at startup from the moment there is a database" —
for the whole of the original Phase 1, and is why the P1-048 – P1-054
gap-closure wave exists. Re-runnability is a specific debt that wave inherited
and did not discharge.

## Remediation

A `schema_migrations` ledger in `internal/adapters/postgres/migrate.go`: each
migration applied once in its own transaction, under an advisory lock, with
already-recorded versions skipped. The ledger is created imperatively rather
than as migration 000, because the numbered slots are pre-allocated to tasks in
`docs/workplan/README.md` and the ledger must precede 001. Delivered under
P1-048 with the second-boot test its brief already required.

## Prevention

- An invariant that only names a property creates no owner for it. I2 says
  migrations run at startup; no task's acceptance criteria ever said they must
  therefore survive a second start.
- When a later task's Verification asserts a property of an earlier task's
  surface, that is scope, not verification. The surface belongs in owned
  surfaces and the work belongs in the brief body.
- Read "tested on a clean database" as a statement about which environments the
  plan has considered, not only about test setup.
