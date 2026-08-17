package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// The whole directory is embedded rather than one variable per file: a
// migration that is added but not listed applies to nobody's database while the
// code that needs its columns ships, which is how 006 reached a release branch
// unapplied. Adding the file is now the only registration step.
//
//go:embed migrations/*.sql
var migrationFiles embed.FS

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

func migrationParts(sql string) (string, string, error) {
	up, down, found := strings.Cut(sql, "-- migrate:down")
	if !found {
		return "", "", errors.New("migration is missing its `-- migrate:down` section")
	}
	return strings.TrimPrefix(up, "-- migrate:up"), down, nil
}

// migrations reads every embedded `NNN_name.sql` in version order. Embedded
// directory entries are sorted by name, so the numeric prefix orders them.
func migrations() ([]migration, error) {
	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return nil, err
	}
	ordered := make([]migration, 0, len(entries))
	for _, entry := range entries {
		version, name, err := migrationName(entry.Name())
		if err != nil {
			return nil, fmt.Errorf("migration %s: %w", entry.Name(), err)
		}
		sql, err := fs.ReadFile(migrationFiles, "migrations/"+entry.Name())
		if err != nil {
			return nil, err
		}
		up, down, err := migrationParts(string(sql))
		if err != nil {
			return nil, fmt.Errorf("migration %s: %w", entry.Name(), err)
		}
		if len(ordered) > 0 && ordered[len(ordered)-1].version >= version {
			return nil, fmt.Errorf("migration %s: version is not greater than the migration before it", entry.Name())
		}
		ordered = append(ordered, migration{version: version, name: name, up: up, down: down})
	}
	return ordered, nil
}

func migrationName(file string) (int, string, error) {
	prefix, rest, found := strings.Cut(strings.TrimSuffix(file, ".sql"), "_")
	if !found || len(prefix) != 3 || rest == "" {
		return 0, "", errors.New("file name must be NNN_name.sql")
	}
	version, err := strconv.Atoi(prefix)
	if err != nil || version <= 0 {
		return 0, "", errors.New("file name must start with a positive three-digit version")
	}
	return version, rest, nil
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
	ordered, err := migrations()
	if err != nil {
		return result, err
	}
	err = withLedger(ctx, pool, func(conn *pgxpool.Conn, applied map[int]struct{}) error {
		for _, m := range ordered {
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
	ordered, err := migrations()
	if err != nil {
		return err
	}
	return withLedger(ctx, pool, func(conn *pgxpool.Conn, applied map[int]struct{}) error {
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
