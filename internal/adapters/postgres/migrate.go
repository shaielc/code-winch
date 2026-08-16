package postgres

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/001_run_event_storage.sql
var migration001 string

//go:embed migrations/002_transactional_outbox.sql
var migration002 string

//go:embed migrations/003_run_supervisor.sql
var migration003 string

//go:embed migrations/004_input_delivery.sql
var migration004 string

//go:embed migrations/005_workflow_runtime.sql
var migration005 string

// ledgerDDL records which migrations a database already has. It is created
// imperatively rather than as migration 000 because the numbered slots are
// pre-allocated to tasks in docs/workplan/README.md and the ledger must precede
// 001.
const ledgerDDL = `CREATE TABLE IF NOT EXISTS schema_migrations (
  version integer PRIMARY KEY,
  name text NOT NULL,
  applied_at timestamptz NOT NULL DEFAULT now()
)`

// migrationLockID serializes migrators. Two daemons starting against one
// database would otherwise both read an empty ledger and race to create the
// same tables.
const migrationLockID = 8081

type migration struct {
	version int
	name    string
	up      string
	down    string
}

func migrationParts(migration string) (string, string) {
	parts := strings.Split(migration, "-- migrate:down")
	return strings.TrimPrefix(parts[0], "-- migrate:up"), parts[1]
}

func migrations() []migration {
	sources := []struct {
		version int
		name    string
		sql     string
	}{
		{1, "run_event_storage", migration001},
		{2, "transactional_outbox", migration002},
		{3, "run_supervisor", migration003},
		{4, "input_delivery", migration004},
		{5, "workflow_runtime", migration005},
	}
	ordered := make([]migration, 0, len(sources))
	for _, source := range sources {
		up, down := migrationParts(source.sql)
		ordered = append(ordered, migration{version: source.version, name: source.name, up: up, down: down})
	}
	return ordered
}

// MigrationResult reports what a MigrateUp call did, so a caller can log the
// difference between a start that migrated and one that found nothing to do.
type MigrationResult struct {
	Applied int // migrations this call installed
	Version int // highest version the ledger records afterwards
}

// MigrateUp installs every migration the ledger does not already record, each
// in its own transaction. It is safe on every start: an already-migrated
// database is a no-op. Production migrations are forward-only; MigrateDown
// exists to verify reversibility on clean test stores.
func MigrateUp(ctx context.Context, pool *pgxpool.Pool) (MigrationResult, error) {
	var result MigrationResult
	err := withLedger(ctx, pool, func(conn *pgxpool.Conn, applied map[int]struct{}) error {
		for _, m := range migrations() {
			result.Version = m.version
			if _, ok := applied[m.version]; ok {
				continue
			}
			if err := apply(ctx, conn, m.up, `INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`, m.version, m.name); err != nil {
				return fmt.Errorf("migration %03d %s: %w", m.version, m.name, err)
			}
			result.Applied++
		}
		return nil
	})
	return result, err
}

// MigrateDown reverses the migrations the ledger records, newest first.
func MigrateDown(ctx context.Context, pool *pgxpool.Pool) error {
	return withLedger(ctx, pool, func(conn *pgxpool.Conn, applied map[int]struct{}) error {
		ordered := migrations()
		for i := len(ordered) - 1; i >= 0; i-- {
			m := ordered[i]
			if _, ok := applied[m.version]; !ok {
				continue
			}
			if err := apply(ctx, conn, m.down, `DELETE FROM schema_migrations WHERE version = $1`, m.version); err != nil {
				return fmt.Errorf("migration %03d %s: %w", m.version, m.name, err)
			}
		}
		return nil
	})
}

// withLedger holds the migration lock on one connection, ensures the ledger
// exists, and hands the caller the versions it already records.
func withLedger(ctx context.Context, pool *pgxpool.Pool, fn func(*pgxpool.Conn, map[int]struct{}) error) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("migration connection: %w", err)
	}
	defer conn.Release()
	if _, err = conn.Exec(ctx, `SELECT pg_advisory_lock($1)`, migrationLockID); err != nil {
		return fmt.Errorf("migration lock: %w", err)
	}
	// Release on the same connection even when ctx is already done, or the lock
	// outlives the migrator and the next start blocks on it.
	defer func() {
		_, _ = conn.Exec(context.WithoutCancel(ctx), `SELECT pg_advisory_unlock($1)`, migrationLockID)
	}()
	if _, err = conn.Exec(ctx, ledgerDDL); err != nil {
		return fmt.Errorf("migration ledger: %w", err)
	}
	applied, err := appliedVersions(ctx, conn)
	if err != nil {
		return err
	}
	return fn(conn, applied)
}

func appliedVersions(ctx context.Context, conn *pgxpool.Conn) (map[int]struct{}, error) {
	rows, err := conn.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("migration ledger: %w", err)
	}
	defer rows.Close()
	applied := make(map[int]struct{})
	for rows.Next() {
		var version int
		if err = rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("migration ledger: %w", err)
		}
		applied[version] = struct{}{}
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("migration ledger: %w", err)
	}
	return applied, nil
}

// apply runs one migration direction and its ledger update together, so a
// database never records a migration it did not fully execute.
func apply(ctx context.Context, conn *pgxpool.Conn, statements, ledger string, args ...any) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err = tx.Exec(ctx, statements); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, ledger, args...); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

var errInvalidJSON = errors.New("postgres repository: invalid JSON document")
